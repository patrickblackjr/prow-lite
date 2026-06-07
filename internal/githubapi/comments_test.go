package githubapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func makeIssueCommentEvent(owner, repo string, prNumber int, commentBody string) *github.IssueCommentEvent {
	return &github.IssueCommentEvent{
		Repo: &github.Repository{
			Name:  github.Ptr(repo),
			Owner: &github.User{Login: github.Ptr(owner)},
		},
		Issue:   &github.Issue{Number: github.Ptr(prNumber)},
		Comment: &github.IssueComment{Body: github.Ptr(commentBody), User: &github.User{Login: github.Ptr("commenter")}},
	}
}

func TestAssignUsers_TooFewArgs(t *testing.T) {
	assignUsers("/assign", makeIssueCommentEvent("owner", "repo", 1, "/assign"), nil, discardLogger())
}

func TestAssignUsers_TooManyArgs(t *testing.T) {
	assignUsers("/assign a b c d e", makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestAssignUsers_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesAssigneesByOwnerByRepoByIssueNumber,
			github.Issue{Number: github.Ptr(1)}),
	))
	assignUsers("/assign @alice", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestAssignUsers_Failure(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesAssigneesByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	assignUsers("/assign alice", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestUnassignUsers_TooFewArgs(t *testing.T) {
	unassignUsers("/unassign", makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestUnassignUsers_TooManyArgs(t *testing.T) {
	unassignUsers("/unassign a b c d e", makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestUnassignUsers_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesAssigneesByOwnerByRepoByIssueNumber,
			github.Issue{Number: github.Ptr(1)}),
	))
	unassignUsers("/unassign @bob", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestUnassignUsers_Failure(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesAssigneesByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	unassignUsers("/unassign bob", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestProcessComment_AllCommands(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesAssigneesByOwnerByRepoByIssueNumber, github.Issue{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesAssigneesByOwnerByRepoByIssueNumber, github.Issue{}),
	))
	event := makeIssueCommentEvent("owner", "repo", 1,
		"/assign alice\n/unassign bob\nnot-a-command")
	ProcessComment(event, c, discardLogger())
}

func TestProcessComment_Lgtm(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	ProcessComment(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger())
}

func TestProcessComment_Approve(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	ProcessComment(makeIssueCommentEvent("owner", "repo", 1, "/approve"), c, discardLogger())
}

func TestProcessComment_Unlgtm(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	ProcessComment(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger())
}

func TestProcessComment_RemoveApprove(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	ProcessComment(makeIssueCommentEvent("owner", "repo", 1, "/remove-approve"), c, discardLogger())
}

func TestProcessComment_RemoveApproval(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	ProcessComment(makeIssueCommentEvent("owner", "repo", 1, "/remove-approval"), c, discardLogger())
}

func TestProcessComment_Unapprove(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	ProcessComment(makeIssueCommentEvent("owner", "repo", 1, "/unapprove"), c, discardLogger())
}

func TestProcessComment_RemoveLgtm(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	ProcessComment(makeIssueCommentEvent("owner", "repo", 1, "/remove-lgtm"), c, discardLogger())
}

func TestLgtm_AddLabelsFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger())
}

func TestLgtm_RemoveLabelFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger())
}

func TestLgtm_UpdateCheckRunFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger())
}

func TestUnlgtm_RemoveLabelFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger())
}

func TestUnlgtm_AddLabelFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger())
}

func TestUnlgtm_UpdateCheckRunFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger())
}

func TestUpdateApprovalCheckRun_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("abc123")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(42))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId,
			github.CheckRun{ID: github.Ptr(int64(42))}),
	))
	err := updateApprovalCheckRun(context.Background(), c, "owner", "repo", 1, "success", "Approved", "The PR is approved.")
	assert.NoError(t, err)
}

func TestUpdateApprovalCheckRun_GetPRFailure(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "internal error")
			}),
		),
	))
	err := updateApprovalCheckRun(context.Background(), c, "owner", "repo", 1, "success", "Approved", "The PR is approved.")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get pull request")
}

func TestUpdateApprovalCheckRun_CreateCheckRunFailure(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("abc123")}}),
		mock.WithRequestMatchHandler(
			mock.PostReposCheckRunsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "internal error")
			}),
		),
	))
	err := updateApprovalCheckRun(context.Background(), c, "owner", "repo", 1, "success", "Approved", "The PR is approved.")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create check run")
}

func TestUpdateApprovalCheckRun_UpdateCheckRunFailure(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("abc123")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo,
			github.CheckRun{ID: github.Ptr(int64(42))}),
		mock.WithRequestMatchHandler(
			mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "internal error")
			}),
		),
	))
	err := updateApprovalCheckRun(context.Background(), c, "owner", "repo", 1, "success", "Approved", "The PR is approved.")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update check run")
}
