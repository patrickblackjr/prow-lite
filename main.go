package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	fmt.Println(os.Environ())
	actions := githubactions.New()
	v := actions.GetInput("var2")
	if v == "" {
		actions.Fatalf("value for var2 not provided")
	}

	token, _ := actions.GetIDToken(ctx, "prow-lite")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	repoName := os.Getenv("GITHUB_ACTION_REPOSITORY")

	if repoName == "" {
		repoName = "prow-lite"
	}

	IssueListByRepoOptions := github.IssueListByRepoOptions{}
	issues, _, err := client.Issues.ListByRepo(ctx, owner, repoName, &IssueListByRepoOptions)
	if err != nil {
		githubactions.Debugf("No issues matched the filters. No acts to take.")
	}

	for i := 0; i < len(issues); i++ {
		opt := &github.IssueListCommentsOptions{}
		comments, _, err := client.Issues.ListComments(ctx, owner, repoName, *issues[i].Number, opt)
		if err != nil {
			githubactions.Warningf("%v", err)
		} else if len(comments) > 0 {
			log.Println(*comments[0].Body)
		} else {
			log.Println("no comment for this issue")
		}
	}

}
