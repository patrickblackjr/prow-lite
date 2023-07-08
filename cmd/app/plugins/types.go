// Package plugins rips the majority of its contents
// from https://github.com/kubernetes/test-infra/blob/master/prow/github/types.go
package plugins

// GenericCommentEventAction coerces multiple actions into its generic equivalent.
type GenericCommentEventAction string

// Comments indicate values that are coerced to the specified value.
const (
	// GenericCommentActionCreated means something was created/opened/submitted
	GenericCommentActionCreated GenericCommentEventAction = "created" // "opened", "submitted"
	// GenericCommentActionEdited means something was edited.
	GenericCommentActionEdited GenericCommentEventAction = "edited"
	// GenericCommentActionDeleted means something was deleted/dismissed.
	GenericCommentActionDeleted GenericCommentEventAction = "deleted" // "dismissed"
)

type Repo struct {
	Owner         User   `json:"owner"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HTMLURL       string `json:"html_url"`
	Fork          bool   `json:"fork"`
	DefaultBranch string `json:"default_branch"`
	Archived      bool   `json:"archived"`
	Private       bool   `json:"private"`
	Description   string `json:"description"`
	Homepage      string `json:"homepage"`
	HasIssues     bool   `json:"has_issues"`
	HasProjects   bool   `json:"has_projects"`
	HasWiki       bool   `json:"has_wiki"`
	NodeID        string `json:"node_id"`
	// Permissions reflect the permission level for the requester, so
	// on a repository GET call this will be for the user whose token
	// is being used, if listing a team's repos this will be for the
	// team's privilege level in the repo
	Permissions RepoPermissions `json:"permissions"`
	Parent      ParentRepo      `json:"parent"`
}

// RepoPermissions describes which permission level an entity has in a
// repo. At most one of the booleans here should be true.
type RepoPermissions struct {
	// Pull is equivalent to "Read" permissions in the web UI
	Pull   bool `json:"pull"`
	Triage bool `json:"triage"`
	// Push is equivalent to "Edit" permissions in the web UI
	Push     bool `json:"push"`
	Maintain bool `json:"maintain"`
	Admin    bool `json:"admin"`
}

// User is a GitHub user account.
type User struct {
	Login       string          `json:"login"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	ID          int             `json:"id"`
	HTMLURL     string          `json:"html_url"`
	Permissions RepoPermissions `json:"permissions"`
	Type        string          `json:"type"`
}

type GenericCommentEvent struct {
	ID           int    `json:"id"`
	NodeID       string `json:"node_id"`
	CommentID    *int
	IsPR         bool
	Action       GenericCommentEventAction
	Body         string
	HTMLUrl      string
	Number       int
	Repo         Repo
	User         User
	IssueAuthor  User
	Assignees    []User
	IssueState   string
	IssueTitle   string
	IssueBody    string
	IssueHTMLURL string
	GUID         string
}

// ParentRepo contains a small subsection of general repository information: it
// just includes the information needed to confirm that a parent repo exists
// and what the name of that repo is.
type ParentRepo struct {
	Owner    User   `json:"owner"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}
