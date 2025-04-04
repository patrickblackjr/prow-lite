package githubapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
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
