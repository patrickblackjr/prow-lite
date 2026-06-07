package githubapi

import (
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestLabel_MissingName(t *testing.T) {
	label("/label", makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestLabel_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
	))
	label("/label kind/bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestLabel_Failure(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	label("/label kind/bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestRemoveLabel_MissingName(t *testing.T) {
	removeLabel("/remove-label", makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestRemoveLabel_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	removeLabel("/remove-label kind/bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestRemoveLabel_Failure(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(w, http.StatusInternalServerError, "boom")
			}),
		),
	))
	removeLabel("/remove-label kind/bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestProcessComment_LabelCommands(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	event := makeIssueCommentEvent("owner", "repo", 1, "/label priority/urgent\n/remove-label kind/bug")
	ProcessComment(event, c, discardLogger())
}
