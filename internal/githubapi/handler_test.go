package githubapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func noop(_ *github.IssueCommentEvent, _ *github.Client, _ *slog.Logger) {}
func noopPR(_ *github.PullRequestEvent, _ *github.Client, _ *slog.Logger) {}

func TestRegisterEventHandlers_NilGin_KnownEvent(t *testing.T) {
	payload, _ := json.Marshal(github.IssueCommentEvent{
		Comment: &github.IssueComment{Body: github.Ptr("hello")},
	})
	handle := RegisterEventHandlers(nil, github.NewClient(nil), discardLogger(), noop, noopPR)
	handle("issue_comment", payload)
}

func TestRegisterEventHandlers_NilGin_UnknownEvent(t *testing.T) {
	handle := RegisterEventHandlers(nil, github.NewClient(nil), discardLogger(), noop, noopPR)
	handle("unknown_event", []byte(`{}`))
}

func TestRegisterEventHandlers_WithGin_KnownEvent(t *testing.T) {
	r := gin.New()
	RegisterEventHandlers(r, github.NewClient(mock.NewMockedHTTPClient()), discardLogger(), noop, noopPR)

	payload, _ := json.Marshal(github.IssueCommentEvent{
		Comment: &github.IssueComment{Body: github.Ptr("hello")},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "issue_comment")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterEventHandlers_WithGin_UnknownEvent(t *testing.T) {
	r := gin.New()
	RegisterEventHandlers(r, github.NewClient(nil), discardLogger(), noop, noopPR)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "unknown_event")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// errorReader always fails on Read.
type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

func TestRegisterEventHandlers_WithGin_BodyReadError(t *testing.T) {
	r := gin.New()
	RegisterEventHandlers(r, github.NewClient(nil), discardLogger(), noop, noopPR)

	req := httptest.NewRequest(http.MethodPost, "/webhook", errorReader{})
	req.Header.Set("X-GitHub-Event", "issue_comment")
	req.ContentLength = -1
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterEventHandlers_WithGin_PullRequestEvent(t *testing.T) {
	r := gin.New()
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	RegisterEventHandlers(r, c, discardLogger(), noop, NewPREventHandler(1))

	payload, _ := json.Marshal(github.PullRequestEvent{
		Action: github.Ptr("opened"),
		Repo:   &github.Repository{Name: github.Ptr("repo"), Owner: &github.User{Login: github.Ptr("owner")}},
		PullRequest: &github.PullRequest{Number: github.Ptr(1)},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleIssueCommentEvent_InvalidJSON(t *testing.T) {
	handleIssueCommentEvent([]byte("not-json"), nil, discardLogger(), noop)
}

func TestDispatchPullRequestEvent_InvalidJSON(t *testing.T) {
	dispatchPullRequestEvent([]byte("not-json"), nil, discardLogger(), noopPR)
}

func prEvent(action string) *github.PullRequestEvent {
	return &github.PullRequestEvent{
		Action: github.Ptr(action),
		Repo: &github.Repository{
			Name:  github.Ptr("repo"),
			Owner: &github.User{Login: github.Ptr("owner")},
		},
		PullRequest: &github.PullRequest{Number: github.Ptr(1)},
	}
}

func TestHandlePullRequestEvent_NonOpenedAction(t *testing.T) {
	handlePullRequestEvent(prEvent("closed"), nil, discardLogger(), 1)
}

func TestHandlePullRequestEvent_Opened_AddLabelsFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	handlePullRequestEvent(prEvent("opened"), c, discardLogger(), 1)
}

func TestHandlePullRequestEvent_Opened_GetSHAFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	handlePullRequestEvent(prEvent("opened"), c, discardLogger(), 1)
}

func TestHandlePullRequestEvent_Opened_CreateCheckRunFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatchHandler(
			mock.PostReposCheckRunsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	handlePullRequestEvent(prEvent("opened"), c, discardLogger(), 1)
}

func TestHandlePullRequestEvent_Opened_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	handlePullRequestEvent(prEvent("opened"), c, discardLogger(), 1)
}

func TestHandlePullRequestEvent_Reopened_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			github.IssueComment{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	handlePullRequestEvent(prEvent("reopened"), c, discardLogger(), 1)
}

func TestHandlePullRequestEvent_Reopened_RemoveLabelFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
		mock.WithRequestMatch(mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			github.IssueComment{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	handlePullRequestEvent(prEvent("reopened"), c, discardLogger(), 1)
}

func TestHandlePullRequestEvent_Reopened_AddCommentFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	handlePullRequestEvent(prEvent("reopened"), c, discardLogger(), 1)
}

func TestHandlePullRequestEvent_ZeroMinApprovals_Opened(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	handlePullRequestEvent(prEvent("opened"), c, discardLogger(), 0)
}

func TestHandlePullRequestEvent_ZeroMinApprovals_RemoveLabelFails(t *testing.T) {
	// WithRequestMatch can't queue two calls for the same endpoint — use a stateful handler.
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
			}),
		),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		// do-not-merge remove fails — warns but continues
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	handlePullRequestEvent(prEvent("opened"), c, discardLogger(), 0)
}

func TestHandlePullRequestEvent_ZeroMinApprovals_CreateCheckRunFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
			}),
		),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatchHandler(
			mock.PostReposCheckRunsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	handlePullRequestEvent(prEvent("opened"), c, discardLogger(), 0)
}


func TestHandlePullRequestEvent_ZeroMinApprovals_Reopened(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			github.IssueComment{ID: github.Ptr(int64(1))}),
		// second label add for "lgtm" after reset
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	handlePullRequestEvent(prEvent("reopened"), c, discardLogger(), 0)
}
