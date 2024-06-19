package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v29/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type gitImpl struct {
	owner       string
	repo        string
	branch      string
	accessToken string
	client      *github.Client
	gqlClient   *githubv4.Client
}

func (g *gitImpl) gQLClient() *githubv4.Client {
	if g.gqlClient == nil {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: g.accessToken})
		tc := oauth2.NewClient(context.Background(), ts)
		g.gqlClient = githubv4.NewClient(tc)
	}
	return g.gqlClient
}

func (g *gitImpl) DeleteBranch(branchName string) error {
	client := g.gQLClient()

	// Define GraphQL query to find the Ref ID
	var q struct {
		Repository struct {
			Ref struct {
				ID githubv4.ID
			} `graphql:"ref(qualifiedName: $ref)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"name":  githubv4.String(g.repo),
		"owner": githubv4.String(g.owner),
		"ref":   githubv4.String("refs/heads/" + branchName),
	}

	if err := client.Query(context.Background(), &q, variables); err != nil {
		return err
	}

	// Define GraphQL mutation to delete branch
	var mutation struct {
		DeleteRef struct {
			ClientMutationId githubv4.String
		} `graphql:"deleteRef(input: $input)"`
	}

	input := githubv4.DeleteRefInput{
		RefID: q.Repository.Ref.ID,
	}

	if err := client.Mutate(context.Background(), &mutation, input, nil); err != nil {
		return fmt.Errorf("error deleting branch: %w", err)
	}
	return nil
}
func main() {
	cl := gitImpl{
		owner:       "jitinchekka2",
		repo:        "git-go",
		branch:      "branch-1",
		accessToken: os.Getenv("GITHUB_TOKEN"),
	}
	if err := cl.DeleteBranch("branch-1"); err != nil {
		fmt.Println(err)
	}
}

//
