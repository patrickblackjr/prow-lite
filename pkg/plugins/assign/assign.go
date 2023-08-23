package assign

import (
	"context"
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/sirupsen/logrus"
)

const pluginName = "assign"

var (
	// AssignRegexp ...
	AssignRegexp = `(?mi)^/(un)?assign(( @?[-\w]+?)*)\s*$`
	// CCRegexp ...
	CCRegexp = `(?mi)^/(un)?cc(( +@?[-/\w]+?)*)\s*$`
)

// Users assigns users provided a comment and GH client
func Users(c *github.Client, event *github.IssueCommentEvent) {
	ctx := context.Background()
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	issueNumber := *event.Issue.Number
	users := parseLogins(*event.Issue.Body)

	c.Issues.AddAssignees(ctx, owner, repo, issueNumber, users)

	logrus.Printf("%s assigned in %s/%s #%d", users, owner, repo, issueNumber)
}

func parseLogins(text string) []string {
	var parts []string
	for _, p := range strings.Split(text, " ") {
		t := strings.Trim(p, "@ ")
		if t == "" {
			continue
		}
		parts = append(parts, t)
	}
	return parts
}
