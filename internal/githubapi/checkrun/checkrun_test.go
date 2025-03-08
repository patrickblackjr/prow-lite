package checkrun

import (
	"context"
	"errors"
	"os"
	"testing"

	"log/slog"

	"github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/mock"
)

// MockGitHubClient is a mock of the GitHub client
type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) CreateCheckRun(ctx context.Context, owner, repo string, opt github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	args := m.Called(ctx, owner, repo, opt)
	checkRun, _ := args.Get(0).(*github.CheckRun)
	response, _ := args.Get(1).(*github.Response)
	return checkRun, response, args.Error(2)
}

func TestCreateCheckRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tests := []struct {
		name       string
		owner      string
		repo       string
		sha        string
		conclusion string
		nameParam  string
		setupMock  func(m *MockGitHubClient)
		expectLog  string
	}{
		{
			name:       "Empty SHA",
			owner:      "owner",
			repo:       "repo",
			sha:        "",
			conclusion: "success",
			nameParam:  "Test",
			setupMock:  func(m *MockGitHubClient) {},
			expectLog:  "SHA is empty, skipping check run creation",
		},
		{
			name:       "Create Check Run Success",
			owner:      "owner",
			repo:       "repo",
			sha:        "sha",
			conclusion: "success",
			nameParam:  "Test",
			setupMock: func(m *MockGitHubClient) {
				m.On("CreateCheckRun", mock.Anything, "owner", "repo", mock.Anything).Return(&github.CheckRun{ID: github.Int64(1)}, &github.Response{}, nil)
			},
			expectLog: "Created check run",
		},
		{
			name:       "Create Check Run Failure",
			owner:      "owner",
			repo:       "repo",
			sha:        "sha",
			conclusion: "failure",
			nameParam:  "Test",
			setupMock: func(m *MockGitHubClient) {
				m.On("CreateCheckRun", mock.Anything, "owner", "repo", mock.Anything).Return(nil, nil, errors.New("failed to create check run"))
			},
			expectLog: "Failed to create check run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockGitHubClient)
			tt.setupMock(mockClient)

			CreateCheckRun(tt.owner, tt.repo, tt.sha, tt.conclusion, tt.nameParam, mockClient, logger)

			mockClient.AssertExpectations(t)
		})
	}
}
