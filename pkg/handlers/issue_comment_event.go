package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/patrickblackjr/prow-lite/pkg/plugins/assign"
	"github.com/sirupsen/logrus"
)

// IssueCommentHandler ...
type IssueCommentHandler struct {
	githubapp.ClientCreator
}

// Handles tells go-githubapp what types of events this handles
func (h *IssueCommentHandler) Handles() []string {
	return []string{"issue_comment"}
}

// Handle handles
func (h *IssueCommentHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssueCommentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		logrus.Errorf("failed to parse issue comment event payload: %s", err)
		return err
	}

	if event.GetIssue().IsPullRequest() == false {
		logrus.Debug("issue comment is issue")
	}

	repo := event.GetRepo()
	issueNumber := event.Issue.Number
	installationID := githubapp.GetInstallationIDFromEvent(&event)

	// Ignore edits and deletions
	if event.GetAction() != "created" {
		return nil
	}

	client, err := h.NewInstallationClient(installationID)
	if err != nil {
		return err
	}

	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	author := event.GetComment().GetUser().GetLogin()
	body := event.GetComment().GetBody()

	if strings.HasSuffix(author, "[bot]") {
		logrus.Debug("issue comment was created by a bot and therefore ignored")
		return nil
	}

	assignMatch, err := regexp.MatchString(assign.AssignRegexp, body)
	ccMatch, err := regexp.MatchString(assign.CCRegexp, body)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	logrus.Info(assignMatch, ccMatch)
	if assignMatch || ccMatch {
		assign.Users(client, &event)
	}

	logrus.Debugf("echoing comment on %s/%s#%d by %s", repoOwner, repoName, *issueNumber, author)
	msg := fmt.Sprintf("**@%s said:**\n```\n%s\n```\n", author, body)
	issueComment := github.IssueComment{
		Body: &msg,
	}

	if _, _, err := client.Issues.CreateComment(ctx, repoOwner, repoName, *issueNumber, &issueComment); err != nil {
		logrus.Errorf("failed to comment on issue: %s", err)
	}

	return nil
}
