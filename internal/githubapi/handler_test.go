package githubapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v70/github"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &github.Client{} // Mock client if needed
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	r := gin.New()
	RegisterEventHandlers(r, client, logger, func(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
		// Mock processComment function
	})

	payload := []byte(`{"action": "created", "comment": {"body": "test comment"}}`)
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payload))
	req.Header.Set("X-GitHub-Event", "issue_comment")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "event_type")
}

// Add more tests for other event handlers

func TestRegisterEventHandlers(t *testing.T) {
	type args struct {
		r              *gin.Engine
		client         *github.Client
		logger         *slog.Logger
		processComment func(*github.IssueCommentEvent, *github.Client, *slog.Logger)
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterEventHandlers(tt.args.r, tt.args.client, tt.args.logger, tt.args.processComment)
		})
	}
}

func Test_handleIssueCommentEvent(t *testing.T) {
	type args struct {
		request        []byte
		client         *github.Client
		logger         *slog.Logger
		processComment func(*github.IssueCommentEvent, *github.Client, *slog.Logger)
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleIssueCommentEvent(tt.args.request, tt.args.client, tt.args.logger, tt.args.processComment)
		})
	}
}

func Test_handlePullRequestEvent(t *testing.T) {
	type args struct {
		request []byte
		client  *github.Client
		logger  *slog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlePullRequestEvent(tt.args.request, tt.args.client, tt.args.logger)
		})
	}
}
