package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v45/github"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
)

// Manual Test: curl http://localhost:8080/api/v1/github/pullrequests/MartinHeinz/python-project-blueprint
// Result `{"pull_requests":["Some Instructions","Add newline to match dev.Dockerfile"]}`
func GetPullRequests(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	if pullRequests, resp, err := config.Config.GitHubClient.PullRequests.List(c, owner, repo, &github.PullRequestListOptions{
		State: "open",
	}); err != nil {
		log.Println(err)
		c.AbortWithStatus(resp.StatusCode)
	} else {
		var pullRequestTitles []string
		for _, pr := range pullRequests {
			pullRequestTitles = append(pullRequestTitles, *pr.Title)
		}
		c.JSON(http.StatusOK, gin.H{
			"pull_requests": pullRequestTitles,
		})
	}
}

func GetPullRequestsPaginated(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	pageSize, _ := strconv.ParseInt(c.Param("page"), 10, 32)

	options := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: int(pageSize)},
	}
	var allPullRequests []*github.PullRequest
	for {
		pullRequests, resp, err := config.Config.GitHubClient.PullRequests.List(c, owner, repo, options)
		if err != nil {
			c.AbortWithError(resp.StatusCode, err)
		}
		allPullRequests = append(allPullRequests, pullRequests...)
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}
	var pullRequestTitles []string
	for _, pr := range allPullRequests {
		pullRequestTitles = append(pullRequestTitles, *pr.Title)
	}
	c.JSON(http.StatusOK, gin.H{
		"pull_requests": pullRequestTitles,
	})
}
