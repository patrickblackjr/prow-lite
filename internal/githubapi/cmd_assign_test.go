package githubapi

import (
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

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

func TestProcessComment_AssignCommands(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesAssigneesByOwnerByRepoByIssueNumber, github.Issue{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesAssigneesByOwnerByRepoByIssueNumber, github.Issue{}),
	))
	event := makeIssueCommentEvent("owner", "repo", 1,
		"/assign alice\n/unassign bob\nnot-a-command")
	ProcessComment(event, c, discardLogger())
}
