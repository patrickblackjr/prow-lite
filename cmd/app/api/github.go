package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v45/github"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
)

func GetIssues(c *gin.Context) {
	if issues, resp, err := config.Config.GitHubClient.Issues.List(c, false, &github.IssueListOptions{}); err != nil {
		log.Println(err)
		c.AbortWithStatus(resp.StatusCode)
	} else {
		var issueTitles []string
		for _, iss := range issues {
			issueTitles = append(issueTitles, *iss.Title, *iss.Body)
		}
		c.JSON(http.StatusOK, gin.H{
			"issues": issueTitles,
		})
	}
}

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
