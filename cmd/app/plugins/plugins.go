package plugins

import (
	"context"
	"regexp"
	"strings"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
	log "github.com/sirupsen/logrus"
)

type GenericCommentHandler func(GenericCommentEvent) error

func RegisterGenericCommentHandler(name string, fn GenericCommentHandler) {
	genericCommentHandlers[name] = fn
}

var (
	genericCommentHandlers = map[string]GenericCommentHandler{}
)

// NewResponder adds a comment on an issue
func NewResponder(msg string) githubevents.IssueCommentEventHandleFunc {
	return func(deliveryID, eventName string, event *github.IssueCommentEvent) error {
		ctx := context.Background()
		owner := *event.Repo.Owner.Login
		repo := *event.Repo.Name
		issueNumber := *event.Issue.Number
		body := event.Comment.GetBody()

		if strings.Contains(body, "/close") {
			closeIssue := &github.IssueRequest{
				State:       github.String("Closed"),
				StateReason: github.String("completed"),
			}
			config.Config.GitHubClient.Issues.Edit(ctx, owner, repo, issueNumber, closeIssue)
		}

		if strings.Contains(body, "/assign") {
			re := regexp.MustCompile(`(?m)[@][A-Za-z][A-Za-z0-9]+(?:[.|_|-][A-Za-z0-9]+)*`)
			assignees := []string{}
			for _, match := range re.FindAllString(body, -1) {
				str := strings.Replace(match, "@", "", -1)
				assignees = append(assignees, str)
			}
			log.Debugf("Assignees: %s", assignees)
			config.Config.GitHubClient.Reactions.CreateIssueCommentReaction(ctx, owner, repo, *event.Comment.ID, "+1")
			_, res, err := config.Config.GitHubClient.Issues.AddAssignees(ctx, owner, repo, issueNumber, assignees)
			if err != nil {
				log.Error(err.Error())
			}
			defer res.Body.Close()
		}
		return nil
	}
}
