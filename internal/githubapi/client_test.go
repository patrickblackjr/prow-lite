package githubapi

import (
	"context"
	"log/slog"
	"reflect"
	"testing"

	"github.com/google/go-github/v71/github"
)

func TestNewGithubClient(t *testing.T) {
	type args struct {
		logger *slog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *ProwLiteGitHubClient
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGithubClient(tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGithubClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGithubClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProwLiteGitHubClient_GetClient(t *testing.T) {
	type args struct {
		logger *slog.Logger
	}
	tests := []struct {
		name string
		g    *ProwLiteGitHubClient
		args args
		want *github.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.GetClient(tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProwLiteGitHubClient.GetClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProwLiteGitHubClient_CreateCheckRun(t *testing.T) {
	type args struct {
		ctx   context.Context
		owner string
		repo  string
		opt   github.CreateCheckRunOptions
	}
	tests := []struct {
		name    string
		g       *ProwLiteGitHubClient
		args    args
		want    *github.CheckRun
		want1   *github.Response
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.g.CreateCheckRun(tt.args.ctx, tt.args.owner, tt.args.repo, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProwLiteGitHubClient.CreateCheckRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProwLiteGitHubClient.CreateCheckRun() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ProwLiteGitHubClient.CreateCheckRun() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestProwLiteGitHubClient_GetPRSHA(t *testing.T) {
	type args struct {
		owner    string
		repo     string
		prNumber int
		logger   *slog.Logger
	}
	tests := []struct {
		name string
		g    *ProwLiteGitHubClient
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.GetPRSHA(tt.args.owner, tt.args.repo, tt.args.prNumber, tt.args.logger); got != tt.want {
				t.Errorf("ProwLiteGitHubClient.GetPRSHA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProwLiteGitHubClient_RemoveLabel(t *testing.T) {
	type args struct {
		owner    string
		repo     string
		prNumber int
		label    string
		logger   *slog.Logger
	}
	tests := []struct {
		name string
		g    *ProwLiteGitHubClient
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.g.RemoveLabel(tt.args.owner, tt.args.repo, tt.args.prNumber, tt.args.label, tt.args.logger)
		})
	}
}

func TestProwLiteGitHubClient_AddComment(t *testing.T) {
	type args struct {
		owner       string
		repo        string
		prNumber    int
		commentText string
		logger      *slog.Logger
	}
	tests := []struct {
		name string
		g    *ProwLiteGitHubClient
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.g.AddComment(tt.args.owner, tt.args.repo, tt.args.prNumber, tt.args.commentText, tt.args.logger)
		})
	}
}

func TestProwLiteGitHubClient_CreateLabel(t *testing.T) {
	type args struct {
		owner       string
		repo        string
		name        string
		color       string
		description string
		logger      *slog.Logger
	}
	tests := []struct {
		name string
		g    *ProwLiteGitHubClient
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.g.CreateLabel(tt.args.owner, tt.args.repo, tt.args.name, tt.args.color, tt.args.description, tt.args.logger)
		})
	}
}

func TestProwLiteGitHubClient_AddLabelsToIssue(t *testing.T) {
	type args struct {
		owner    string
		repo     string
		prNumber int
		labels   []string
		logger   *slog.Logger
	}
	tests := []struct {
		name string
		g    *ProwLiteGitHubClient
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.g.AddLabelsToIssue(tt.args.owner, tt.args.repo, tt.args.prNumber, tt.args.labels, tt.args.logger)
		})
	}
}
