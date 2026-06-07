package githubapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/checkrun"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/pullrequest"
)

func RegisterEventHandlers(r *gin.Engine, client *github.Client, logger *slog.Logger, processComment func(*github.IssueCommentEvent, *github.Client, *slog.Logger)) func(string, []byte) {
	eventHandlers := map[string]func([]byte, *github.Client, *slog.Logger){
		"issue_comment": func(request []byte, client *github.Client, logger *slog.Logger) {
			handleIssueCommentEvent(request, client, logger, processComment)
		},
		"pull_request": func(request []byte, client *github.Client, logger *slog.Logger) {
			handlePullRequestEvent(request, client, logger)
		},
	}

	if r != nil {
		r.POST("/webhook", func(c *gin.Context) {
			eventType := c.Request.Header.Get("X-GitHub-Event")
			logger.Info("received webhook", slog.String("event_type", eventType))

			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"event_type": eventType})

			if handler, ok := eventHandlers[eventType]; ok {
				handler(body, client, logger)
			} else {
				logger.Warn("unhandled event type", slog.String("event_type", eventType))
			}
		})
	}

	return func(eventType string, payload []byte) {
		if handler, ok := eventHandlers[eventType]; ok {
			handler(payload, client, logger)
		} else {
			logger.Warn("unhandled event type", slog.String("event_type", eventType))
		}
	}
}

func handleIssueCommentEvent(request []byte, client *github.Client, logger *slog.Logger, processComment func(*github.IssueCommentEvent, *github.Client, *slog.Logger)) {
	var event github.IssueCommentEvent
	if err := json.Unmarshal(request, &event); err != nil {
		logger.Error("failed to unmarshal issue_comment payload", slog.String("error", err.Error()))
		return
	}
	logger.Info("handling issue comment event", slog.String("comment", event.GetComment().GetBody()))
	processComment(&event, client, logger)
}

func handlePullRequestEvent(request []byte, client *github.Client, logger *slog.Logger) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(request, &event); err != nil {
		logger.Error("failed to unmarshal pull_request payload", slog.String("error", err.Error()))
		return
	}

	action := event.GetAction()
	logger.Info("handling pull request event", slog.String("action", action))

	if action != "opened" && action != "reopened" {
		return
	}

	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetPullRequest().GetNumber()

	if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{"do-not-merge"}); err != nil {
		logger.Error("failed to add do-not-merge label", slog.String("error", err.Error()))
		return
	}
	logger.Info("added do-not-merge label")

	if action == "reopened" {
		if err := pullrequest.RemoveLabel(owner, repo, prNumber, "lgtm", client, logger); err != nil {
			logger.Error("failed to remove lgtm label", slog.String("error", err.Error()))
		}
		if err := pullrequest.AddComment(owner, repo, prNumber, "Approval has been reset since this PR was reopened.", client, logger); err != nil {
			logger.Error("failed to add comment", slog.String("error", err.Error()))
		}
	}

	sha, err := pullrequest.GetPRSHA(owner, repo, prNumber, client, logger)
	if err != nil {
		logger.Error("failed to get PR SHA", slog.String("error", err.Error()))
		return
	}

	if _, err := checkrun.CreateCheckRun(owner, repo, sha, "neutral", "Approval needed", client, logger); err != nil {
		logger.Error("failed to create check run", slog.String("error", err.Error()))
	}
}
