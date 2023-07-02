package api

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
	"github.com/stretchr/testify/assert"
)

func TestGetPullRequests(t *testing.T) {
	expectedTitles := []string{
		"PR number 1",
		"PR number 3",
	}
	closedPullRequestTitle := "PR number 2"
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposPullsByOwnerByRepo,
			[]github.PullRequest{
				{
					State: github.String("open"),
					Title: &expectedTitles[0],
				},
				{
					State: github.String("closed"),
					Title: &closedPullRequestTitle,
				},
				{
					State: github.String("open"),
					Title: &expectedTitles[1],
				},
			},
		),
	)
	client := github.NewClient(mockedHTTPClient)
	config.Config.GitHubClient = client

	gin.SetMode(gin.TestMode)
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	ctx.Params = []gin.Param{
		{Key: "owner", Value: "octocat"},
		{Key: "repo", Value: "hello-world"},
	}

	GetPullRequests(ctx)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		println(err)
	}

	assert.Equal(t, 200, res.Code)
	assert.Contains(t, string(body), expectedTitles[0])
	assert.NotContains(t, string(body), closedPullRequestTitle[1])
	assert.Contains(t, string(body), expectedTitles[1])
}
