package githubapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

// oneApprovalComments returns a mock comment list representing one approval from a commenter.
func oneApprovalComments() []*github.IssueComment {
	return []*github.IssueComment{
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("commenter")}},
	}
}

// noApprovalComments returns a mock comment list with no approvals.
func noApprovalComments() []*github.IssueComment {
	return []*github.IssueComment{
		{Body: github.Ptr("/unlgtm"), User: &github.User{Login: github.Ptr("commenter")}},
	}
}

func TestProcessComment_Lgtm(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, oneApprovalComments()),
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
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, oneApprovalComments()),
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
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
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
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
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
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
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
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
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
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
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
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, oneApprovalComments()),
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger(), 1)
}

func TestLgtm_RemoveLabelFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, oneApprovalComments()),
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
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger(), 1)
}

func TestLgtm_UpdateCheckRunFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, oneApprovalComments()),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger(), 1)
}

func TestLgtm_CountApprovalsFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger(), 1)
}

func TestUnlgtm_RemoveLabelFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
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
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger(), 1)
}

func TestUnlgtm_AddLabelFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
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
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger(), 1)
}

func TestUnlgtm_UpdateCheckRunFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, noApprovalComments()),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger(), 1)
}

func TestUnlgtm_CountApprovalsFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger(), 1)
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

// TestLgtm_ZeroMinApprovals verifies that min_approvals=0 auto-approves without counting comments.
func TestLgtm_ZeroMinApprovals(t *testing.T) {
	// No comment-list mock — countApprovals must not be called when minApprovals=0.
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger(), 0)
}

// TestLgtm_UpdateCheckRunFails_BelowMinApprovals verifies the error path when updateApprovalCheckRun
// fails after recording an approval that does not yet meet the required count.
func TestLgtm_UpdateCheckRunFails_BelowMinApprovals(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, oneApprovalComments()),
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger(), 2)
}

// TestLgtm_MinApprovalsNotMet verifies that when minApprovals > current count, the check run
// stays neutral and labels are not added.
func TestLgtm_MinApprovalsNotMet(t *testing.T) {
	// Only 1 approval, but 2 required — expect neutral check run, no label changes
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, oneApprovalComments()),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	lgtm(makeIssueCommentEvent("owner", "repo", 1, "/lgtm"), c, discardLogger(), 2)
}

// TestUnlgtm_UpdateCheckRunFails_StillApproved verifies the error path when updateApprovalCheckRun
// fails after a revocation that still leaves enough approvals.
func TestUnlgtm_UpdateCheckRunFails_StillApproved(t *testing.T) {
	twoApprovalComments := []*github.IssueComment{
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("alice")}},
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("bob")}},
		{Body: github.Ptr("/unlgtm"), User: &github.User{Login: github.Ptr("alice")}},
	}
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, twoApprovalComments),
		mock.WithRequestMatchHandler(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger(), 1)
}

// TestUnlgtm_StillApprovedByOthers verifies that when enough approvals remain after one is revoked,
// the check run stays successful and labels are not changed.
func TestUnlgtm_StillApprovedByOthers(t *testing.T) {
	// Two users approved and one revokes. Make sure count drops to 1, which still meets minApprovals=1
	twoApprovalComments := []*github.IssueComment{
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("alice")}},
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("bob")}},
		{Body: github.Ptr("/unlgtm"), User: &github.User{Login: github.Ptr("alice")}},
	}
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, twoApprovalComments),
		mock.WithRequestMatch(mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{Head: &github.PullRequestBranch{SHA: github.Ptr("sha1")}}),
		mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, github.CheckRun{ID: github.Ptr(int64(1))}),
		mock.WithRequestMatch(mock.PatchReposCheckRunsByOwnerByRepoByCheckRunId, github.CheckRun{ID: github.Ptr(int64(1))}),
	))
	unlgtm(makeIssueCommentEvent("owner", "repo", 1, "/unlgtm"), c, discardLogger(), 1)
}

// TestCountApprovals verifies approval counting logic across multiple users and unlgtm.
func TestCountApprovals(t *testing.T) {
	comments := []*github.IssueComment{
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("alice")}},
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("bob")}},
		{Body: github.Ptr("/unlgtm"), User: &github.User{Login: github.Ptr("alice")}},
		{Body: github.Ptr("/approve"), User: &github.User{Login: github.Ptr("charlie")}},
	}
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, comments),
	))
	count, err := countApprovals(context.Background(), c, "owner", "repo", 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, count) // bob and charlie; alice revoked
}

// TestCountApprovals_ResetOnReopen verifies that approvals posted before a reopen reset are not counted.
func TestCountApprovals_ResetOnReopen(t *testing.T) {
	comments := []*github.IssueComment{
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("alice")}},
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("bob")}},
		{Body: github.Ptr(ApprovalResetComment), User: &github.User{Login: github.Ptr("bot")}},
		{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("charlie")}},
	}
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber, comments),
	))
	count, err := countApprovals(context.Background(), c, "owner", "repo", 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, count) // alice and bob cleared by reset; only charlie counts
}

// TestCountApprovals_Pagination verifies that multi-page comment responses are fully consumed.
func TestCountApprovals_Pagination(t *testing.T) {
	page := 0
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				page++
				w.Header().Set("Content-Type", "application/json")
				if page == 1 {
					w.Header().Set("Link", `<https://api.github.com/repos/owner/repo/issues/1/comments?page=2>; rel="next"`)
				}
				var comments []*github.IssueComment
				if page == 1 {
					comments = []*github.IssueComment{
						{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("alice")}},
					}
				} else {
					comments = []*github.IssueComment{
						{Body: github.Ptr("/lgtm"), User: &github.User{Login: github.Ptr("bob")}},
					}
				}
				_ = json.NewEncoder(w).Encode(comments)
			}),
		),
	))
	count, err := countApprovals(context.Background(), c, "owner", "repo", 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestCountApprovals_ListFails verifies that API errors are shown.
func TestCountApprovals_ListFails(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	_, err := countApprovals(context.Background(), c, "owner", "repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list comments")
}
