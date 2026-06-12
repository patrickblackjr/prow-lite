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
	Action      string
	Description string
}

// LabelDef documents a GitHub label used by the event plugin.
type LabelDef struct {
	Name    string
	Meaning string
}

// CommandPlugin groups related slash commands into a named plugin for documentation.
type CommandPlugin struct {
	Name        string
	Description string
	// Trigger describes which webhook events activate this plugin.
	Trigger  string
	Commands []CommandDef
	PREvents []PREventDef
	Labels   []LabelDef
	// Notes are appended as a freeform section at the end of the generated doc.
	Notes []string
}

// commandHandler pairs a CommandDef with its runtime implementation.
type commandHandler struct {
	def     CommandDef
	handler func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger)
}

func buildDispatchTable(minApprovals int, categories map[string][]string) []commandHandler {
	handlers := []commandHandler{
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
				Triggers:    []string{"/unlabel", "/remove-label"},
				Description: "Removes a label from the issue or PR. Supports `category/name`, `category:name`, or `category name` syntax.",
				Usage:       "/unlabel <label-name>",
			},
			handler: func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				unlabel(command, event, client, logger)
			},
		},
	}

	for category, labels := range categories {
		cat := category
		lbls := labels
		handlers = append(handlers, commandHandler{
			def: CommandDef{
				Triggers:    []string{"/" + cat},
				Description: "Adds a `" + cat + "/<label>` label to the issue or PR.",
				Usage:       "/" + cat + " <label>",
			},
			handler: func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				applyCategoryLabel(command, cat, lbls, event, client, logger)
			},
		})
		handlers = append(handlers, commandHandler{
			def: CommandDef{
				Triggers:    []string{"/un" + cat},
				Description: "Removes a `" + cat + "/<label>` label from the issue or PR.",
				Usage:       "/un" + cat + " <label>",
			},
			handler: func(command string, event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
				removeCategoryLabel(command, cat, lbls, event, client, logger)
			},
		})
	}
	return handlers
}

// LGTMPlugin documents the lgtm/approve command group and its PR lifecycle behavior.
var LGTMPlugin = CommandPlugin{
	Name:        "lgtm",
	Description: "Manages pull request approvals via slash commands. Tracks approval count against a configurable minimum and maintains a GitHub check run.",
	Trigger:     "GitHub `issue_comment` and `pull_request` webhook events.",
	Commands: []CommandDef{
		{
			Triggers:    []string{"/lgtm", "/approve"},
			Description: "Approves the pull request. Adds the `lgtm` label, removes `do-not-merge`, and marks the LGTM check run as passed.",
			Usage:       "/lgtm",
		},
		{
			Triggers:    []string{"/remove-lgtm", "/remove-approve", "/remove-approval", "/unapprove", "/unlgtm"},
			Description: "Revokes approval of the pull request. Removes the `lgtm` label, adds `do-not-merge`, and resets the LGTM check run to neutral.",
			Usage:       "/remove-lgtm",
		},
	},
	PREvents: []PREventDef{
		{
			Action:      "opened",
			Description: "Adds `do-not-merge` label; creates LGTM check run with `neutral` conclusion.",
		},
		{
			Action:      "reopened",
			Description: "Adds `do-not-merge`, removes `lgtm`, posts an approval-reset comment, and creates an LGTM check run with `neutral` conclusion.",
		},
	},
	Labels: []LabelDef{
		{Name: "lgtm", Meaning: "PR has been approved."},
		{Name: "do-not-merge", Meaning: "PR is not ready to merge."},
	},
}

// AssignPlugin documents the assign/unassign command group.
var AssignPlugin = CommandPlugin{
	Name:        "assign",
	Description: "Assigns and unassigns users on issues and pull requests via slash commands.",
	Trigger:     "GitHub `issue_comment` webhook event.",
	Commands: []CommandDef{
		{
			Triggers:    []string{"/assign"},
			Description: "Assigns up to 3 users to the issue or pull request. The `@` prefix on usernames is optional.",
			Usage:       "/assign @user1 [@user2] [@user3]",
		},
		{
			Triggers:    []string{"/unassign"},
			Description: "Removes up to 3 users from issue or pull request assignees. The `@` prefix on usernames is optional.",
			Usage:       "/unassign @user1 [@user2] [@user3]",
		},
	},
}

// LabelPlugin documents the label/unlabel command group.
var LabelPlugin = CommandPlugin{
	Name:        "label",
	Description: "Adds and removes labels on issues and pull requests via slash commands.",
	Trigger:     "GitHub `issue_comment` webhook event.",
	Commands: []CommandDef{
		{
			Triggers:    []string{"/label"},
			Description: "Adds a label to the issue or PR. Supports `category/name`, `category:name`, or `category name` syntax.",
			Usage:       "/label <label-name>",
		},
		{
			Triggers:    []string{"/unlabel", "/remove-label"},
			Description: "Removes a label from the issue or PR. Supports `category/name`, `category:name`, or `category name` syntax.",
			Usage:       "/unlabel <label-name>",
		},
	},
	Notes: []string{
		"When `labels.yml` is configured, category-specific commands are generated automatically: `/<category> <label>` adds `<category>/<label>` and `/un<category> <label>` removes it. For example, `/kind bug` adds the `kind/bug` label and `/unkind bug` removes it.",
	},
}

// NewProcessComment returns a ProcessComment function configured with the given minimum approvals
// and optional label categories loaded from labels.yml. Pass nil for categories to skip dynamic commands.
func NewProcessComment(minApprovals int, categories map[string][]string) func(*github.IssueCommentEvent, *github.Client, *slog.Logger) {
	table := buildDispatchTable(minApprovals, categories)
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

// ProcessComment dispatches slash commands with a default of 1 required approval and no category commands.
func ProcessComment(event *github.IssueCommentEvent, client *github.Client, logger *slog.Logger) {
	NewProcessComment(1, nil)(event, client, logger)
}
