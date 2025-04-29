package githubapi

import (
	"log/slog"
	"testing"

	"github.com/google/go-github/v71/github"
)

func TestProcessComment(t *testing.T) {
	type args struct {
		event  *github.IssueCommentEvent
		client *github.Client
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ProcessComment(tt.args.event, tt.args.client, tt.args.logger)
		})
	}
}

func Test_assignUsers(t *testing.T) {
	type args struct {
		command string
		event   *github.IssueCommentEvent
		client  *github.Client
		logger  *slog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignUsers(tt.args.command, tt.args.event, tt.args.client, tt.args.logger)
		})
	}
}

func Test_unassignUsers(t *testing.T) {
	type args struct {
		command string
		event   *github.IssueCommentEvent
		client  *github.Client
		logger  *slog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unassignUsers(tt.args.command, tt.args.event, tt.args.client, tt.args.logger)
		})
	}
}

func Test_lgtm(t *testing.T) {
	type args struct {
		event  *github.IssueCommentEvent
		client *github.Client
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lgtm(tt.args.event, tt.args.client, tt.args.logger)
		})
	}
}

func Test_unlgtm(t *testing.T) {
	type args struct {
		event  *github.IssueCommentEvent
		client *github.Client
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unlgtm(tt.args.event, tt.args.client, tt.args.logger)
		})
	}
}
