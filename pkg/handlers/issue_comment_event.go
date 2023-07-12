package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/sirupsen/logrus"
)

type IssueCommentHandler struct {
	githubapp.ClientCreator
}

func (h *IssueCommentHandler) Handles() []string {
	return []string{"issue_comment"}
}

func (h *IssueCommentHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssueCommentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		logrus.Errorf("failed to parse issue comment event payload: %s", err)
		return err
	}

	if event.GetIssue().IsPullRequest() {
		logrus.Debug("issue comment is on a pull request")
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
