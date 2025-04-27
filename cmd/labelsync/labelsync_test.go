package labelsync

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/google/go-github/v71/github"
)

func TestGetLabelSyncConfig(t *testing.T) {
	type args struct {
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *LabelSyncConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabelSyncConfig(tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabelSyncConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syncLabels(t *testing.T) {
	type args struct {
		owner     string
		labels    []github.Label
		overwrite bool
		dryRun    bool
		client    *github.Client
		logger    *slog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncLabels(tt.args.owner, tt.args.labels, tt.args.overwrite, tt.args.dryRun, tt.args.client, tt.args.logger)
		})
	}
}

func Test_createLabels(t *testing.T) {
	type args struct {
		owner  string
		config *LabelSyncConfig
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
			createLabels(tt.args.owner, tt.args.config, tt.args.client, tt.args.logger)
		})
	}
}

func Test_createExtraLabels(t *testing.T) {
	type args struct {
		owner  string
		config *LabelSyncConfig
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
			createExtraLabels(tt.args.owner, tt.args.config, tt.args.client, tt.args.logger)
		})
	}
}

func TestLabelSync(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LabelSync()
		})
	}
}
