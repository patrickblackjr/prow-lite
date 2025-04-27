package checkrun_test

import (
	"net/http"
	"os"
	"reflect"
	"testing"

	"log/slog"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/patrickblackjr/prow-lite/internal/githubapi/checkrun"
)

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestCreateCheckRun(t *testing.T) {
	type args struct {
		owner      string
		repo       string
		sha        string
		conclusion string
		name       string
		client     *github.Client
		logger     *slog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *github.CheckRun
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				owner:      "owner",
				repo:       "repo",
				sha:        "sha",
				conclusion: "success",
				name:       "Test",
				client: github.NewClient(mock.NewMockedHTTPClient(
					mock.WithRequestMatch(
						mock.PostReposCheckRunsByOwnerByRepo,
						github.CheckRun{
							ID: github.Ptr(int64(1)),
						},
					),
				)),
				logger: setupLogger(),
			},
			want:    &github.CheckRun{ID: github.Ptr(int64(1))},
			wantErr: false,
		},
		{
			name: "Failure",
			args: args{
				owner:      "owner",
				repo:       "repo",
				sha:        "sha",
				conclusion: "failure",
				name:       "Test",
				client: github.NewClient(mock.NewMockedHTTPClient(
					mock.WithRequestMatchHandler(
						mock.PostReposCheckRunsByOwnerByRepo,
						http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							mock.WriteError(w, http.StatusInternalServerError, "failed to create check run")
						}),
					),
				)),
				logger: setupLogger(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkrun.CreateCheckRun(tt.args.owner, tt.args.repo, tt.args.sha, tt.args.conclusion, tt.args.name, tt.args.client, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCheckRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateCheckRun() = %v, want %v", got, tt.want)
			}
		})
	}
}
