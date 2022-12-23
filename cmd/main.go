package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/patrickblackjr/prow-lite/cmd/app/api"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
	"github.com/patrickblackjr/prow-lite/cmd/app/utils"
	"github.com/patrickblackjr/prow-lite/cmd/app/webhooks"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// load application configurations
	if err := config.LoadConfig("../config"); err != nil {
		panic(fmt.Errorf("invalid application configuration: %s", err))
	}

	// Creates a router without any middleware by default
	r := gin.New()

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		v1.POST("/github/payload", webhooks.ConsumeEvent)
		v1.GET("/github/pullrequests/:owner/:repo", api.GetPullRequests)
		v1.GET("/github/pullrequests/:owner/:repo/:page", api.GetPullRequestsPaginated)
	}

	utils.InitGitHubClient()

	r.Run(fmt.Sprintf(":%v", config.Config.ServerPort))
}
