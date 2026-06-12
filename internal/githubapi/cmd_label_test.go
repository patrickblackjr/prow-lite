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

func TestUnlabel_MissingName(t *testing.T) {
	unlabel("/unlabel", makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestUnlabel_SlashSyntax(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	unlabel("/unlabel kind/bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestUnlabel_ColonSyntax(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	unlabel("/unlabel kind:bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestUnlabel_SpaceSyntax(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	unlabel("/unlabel kind bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestRemoveCategoryLabel_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	removeCategoryLabel("/unkind bug", "kind", []string{"bug", "feature"}, makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestRemoveCategoryLabel_InvalidLabel(t *testing.T) {
	removeCategoryLabel("/unkind nope", "kind", []string{"bug"}, makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestRemoveCategoryLabel_MissingLabel(t *testing.T) {
	removeCategoryLabel("/unkind", "kind", []string{"bug"}, makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestProcessComment_UnlabelCommand(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	event := makeIssueCommentEvent("owner", "repo", 1, "/unkind bug")
	categories := map[string][]string{"kind": {"bug", "feature"}}
	NewProcessComment(1, categories)(event, c, discardLogger())
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

func TestLabel_CategoryColonSyntax(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
	))
	label("/label kind:bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestLabel_CategorySpaceSyntax(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
	))
	label("/label kind bug", makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestApplyCategoryLabel_Success(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
	))
	applyCategoryLabel("/kind bug", "kind", []string{"bug", "feature"}, makeIssueCommentEvent("owner", "repo", 1, ""), c, discardLogger())
}

func TestApplyCategoryLabel_InvalidLabel(t *testing.T) {
	// Should warn and not call the API (nil client would panic if called)
	applyCategoryLabel("/kind nope", "kind", []string{"bug", "feature"}, makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestApplyCategoryLabel_MissingLabel(t *testing.T) {
	// Should warn and not call the API
	applyCategoryLabel("/kind", "kind", []string{"bug"}, makeIssueCommentEvent("owner", "repo", 1, ""), nil, discardLogger())
}

func TestProcessComment_CategoryCommand(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
	))
	event := makeIssueCommentEvent("owner", "repo", 1, "/kind bug")
	categories := map[string][]string{"kind": {"bug", "feature"}}
	NewProcessComment(1, categories)(event, c, discardLogger())
}

func TestProcessComment_LabelCommands(t *testing.T) {
	c := github.NewClient(mock.NewMockedHTTPClient(
		mock.WithRequestMatch(mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber, []github.Label{}),
		mock.WithRequestMatch(mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName, nil),
	))
	event := makeIssueCommentEvent("owner", "repo", 1, "/label priority/urgent\n/remove-label kind/bug")
	ProcessComment(event, c, discardLogger())
}
