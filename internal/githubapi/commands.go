package githubapi

import (
	"log/slog"
	"strings"

	"github.com/google/go-github/v71/github"
)

// CommandDef describes a slash command handled by the event plugin.
type CommandDef struct {
	// Triggers are the slash-command prefixes that activate this command.
	Triggers    []string
	Description string
	// Usage shows the full invocation syntax.
	Usage string
}

// PREventDef documents a pull_request webhook action handled by the event plugin.
type PREventDef struct {
	Action   string
	Behavior string
}

// LabelDef documents a GitHub label used by the event plugin.
type LabelDef struct {
	Name    string
	Meaning string
}

// commandHandler pairs a CommandDef with its runtime implementation.
type commandHandler struct {
	def     CommandDef
	handler func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger)
}

func buildDispatchTable(minApprovals int) []commandHandler {
	return []commandHandler{
		{
			def: CommandDef{
				Triggers:    []string{"/lgtm", "/approve"},
				Description: "Approves the pull request. Adds the `lgtm` label, removes `do-not-merge`, and marks the LGTM check run as passed.",
				Usage:       "/lgtm",
			},
			handler: func(_ string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				lgtm(event, client, logger, minApprovals)
			},
		},
		{
			def: CommandDef{
				Triggers:    []string{"/remove-lgtm", "/remove-approve", "/remove-approval", "/unapprove", "/unlgtm"},
				Description: "Revokes approval of the pull request. Removes the `lgtm` label, adds `do-not-merge`, and resets the LGTM check run to neutral.",
				Usage:       "/remove-lgtm",
			},
			handler: func(_ string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				unlgtm(event, client, logger, minApprovals)
			},
		},
		{
			def: CommandDef{
				Triggers:    []string{"/assign"},
				Description: "Assigns up to 3 users to the pull request. The `@` prefix on usernames is optional.",
				Usage:       "/assign @user1 [@user2] [@user3]",
			},
			handler: func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				assignUsers(command, event, client, logger)
			},
		},
		{
			def: CommandDef{
				Triggers:    []string{"/unassign"},
				Description: "Removes up to 3 users from pull request assignees. The `@` prefix on usernames is optional.",
				Usage:       "/unassign @user1 [@user2] [@user3]",
			},
			handler: func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				unassignUsers(command, event, client, logger)
			},
		},
		{
			def: CommandDef{
				Triggers:    []string{"/label"},
				Description: "Adds an arbitrary label to the pull request.",
				Usage:       "/label <label-name>",
			},
			handler: func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				label(command, event, client, logger)
			},
		},
		{
			def: CommandDef{
				Triggers:    []string{"/remove-label"},
				Description: "Removes an arbitrary label from the pull request.",
				Usage:       "/remove-label <label-name>",
			},
			handler: func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				removeLabel(command, event, client, logger)
			},
		},
	}
}

// Commands exposes the command definitions for doc generation.
var Commands = func() []CommandDef {
	table := buildDispatchTable(1)
	defs := make([]CommandDef, len(table))
	for i, ch := range table {
		defs[i] = ch.def
	}
	return defs
}()

// PREvents documents the pull_request webhook actions handled by the event plugin.
var PREvents = []PREventDef{
	{
		Action:   "opened",
		Behavior: "Adds `do-not-merge` label; creates LGTM check run with `neutral` conclusion.",
	},
	{
		Action:   "reopened",
		Behavior: "Adds `do-not-merge`, removes `lgtm`, posts an approval-reset comment, and creates an LGTM check run with `neutral` conclusion.",
	},
}

// EventPluginLabels documents the labels used by the event plugin.
var EventPluginLabels = []LabelDef{
	{Name: "lgtm", Meaning: "PR has been approved."},
	{Name: "do-not-merge", Meaning: "PR is not ready to merge."},
}

// NewProcessComment returns a ProcessComment function configured with the given minimum approvals.
func NewProcessComment(minApprovals int) func(*github.IssueCommentEvent, *github.Client, *slog.Logger) {
	table := buildDispatchTable(minApprovals)
	return func(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
		lines := strings.Split(event.GetComment().GetBody(), "\n")
		for _, line := range lines {
			for _, ch := range table {
				for _, trigger := range ch.def.Triggers {
					if strings.HasPrefix(line, trigger) {
						ch.handler(line, event, client, logger)
						break
					}
				}
			}
		}
	}
}

// ProcessComment dispatches slash commands with a default of 1 required approval.
func ProcessComment(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	NewProcessComment(1)(event, client, logger)
}
