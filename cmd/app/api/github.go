package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
)

func GetPullRequests(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	if pullRequests, res, err := config.Config.GitHubClient.PullRequests.List(c, owner, repo, &github.PullRequestListOptions{
		State: "open",
	}); err != nil {
		log.Println(err)
		c.AbortWithStatus(res.StatusCode)
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
