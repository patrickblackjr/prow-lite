package githubapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/checkrun"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/pullrequest"
)

func RegisterEventHandlers(r *gin.Engine, client *github.Client, logger *slog.Logger, processComment func(*github.IssueCommentEvent, *github.Client, *slog.Logger)) {
	eventHandlers := map[string]func([]byte, *github.Client, *slog.Logger){
		"issue_comment": func(request []byte, client *github.Client, logger *slog.Logger) {
			handleIssueCommentEvent(request, client, logger, processComment)
		},
		"pull_request": func(request []byte, client *github.Client, logger *slog.Logger) {
			handlePullRequestEvent(request, client, logger)
		},
		// Add more event handlers here
	}

	r.POST("/webhook", func(c *gin.Context) {
		ghEventHeader := c.Request.Header.Get("X-GitHub-Event")
		logger.Info(string(ghEventHeader))
		request, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"event_type": ghEventHeader})

		if handler, ok := eventHandlers[ghEventHeader]; ok {
			handler(request, client, logger)
		} else {
			logger.Warn("unhandled event type", slog.String("event_type", ghEventHeader))
		}
	})
}

func handleIssueCommentEvent(request []byte, client *github.Client, logger *slog.Logger, processComment func(*github.IssueCommentEvent, *github.Client, *slog.Logger)) {
	var event github.IssueCommentEvent
	if err := json.Unmarshal(request, &event); err != nil {
		logger.Error("failed to unmarshal event payload", slog.String("error", err.Error()))
		return
	}
	// Handle the issue comment event
	comment := *event.Comment.Body
	logger.Info("handling issue comment event", slog.String("comment", comment))
	processComment(&event, client, logger)
}

func handlePullRequestEvent(request []byte, client *github.Client, logger *slog.Logger) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(request, &event); err != nil {
		logger.Error("failed to unmarshal event payload", slog.String("error", err.Error()))
		return
	}
	// Handle the pull request event
	logger.Info("handling pull request event", slog.String("action", *event.Action))

	if *event.Action == "opened" || *event.Action == "reopened" {
		ctx := context.Background()
		owner := event.GetRepo().GetOwner().GetLogin()
		repo := event.GetRepo().GetName()
		prNumber := *event.PullRequest.Number

		// Ensure the "do-not-merge" label exists
		_, _, err := client.Issues.GetLabel(ctx, owner, repo, "do-not-merge")
		if err != nil {
			_, _, err = client.Issues.CreateLabel(ctx, owner, repo, &github.Label{
				Name:        github.Ptr("do-not-merge"),
				Color:       github.Ptr("b60205"),
				Description: github.Ptr("Do not merge this PR"),
			})
			if err != nil {
				logger.Error("failed to create do-not-merge label", slog.String("error", err.Error()))
				return
			}
			logger.Info("created do-not-merge label")
		}

		// Add the "do-not-merge" label to the pull request
		_, _, err = client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{"do-not-merge"})
		if err != nil {
			logger.Error("failed to add do-not-merge label", slog.String("error", err.Error()))
			return
		}
		logger.Info("added do-not-merge label")

		if *event.Action == "reopened" {
			pullrequest.RemoveLabel(owner, repo, prNumber, "lgtm", client, logger)
			pullrequest.AddComment(owner, repo, prNumber, "Approval has been reset since this PR was reopened.", client, logger)
		}

		checkrun.CreateCheckRun(owner, repo, pullrequest.GetPRSHA(owner, repo, prNumber, client, logger), "neutral", "Approval needed", client, logger)
	}
}
