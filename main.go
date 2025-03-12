package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"github.com/lmittmann/tint"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	sloggin "github.com/samber/slog-gin"
)

var logLevel = new(slog.LevelVar)

type ProwLiteGitHubClient struct {
	client *github.Client
}

func setupRouter(client *github.Client, logger *slog.Logger) *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(sloggin.New(logger))
	r.Use(gin.Recovery())

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	githubapi.RegisterEventHandlers(r, client, logger, githubapi.ProcessComment)
	githubapi.EnsureLabels("patrickblackjr", []string{"lgtm"}, client, logger)

	return r
}

func main() {

	logLevel.Set(slog.LevelDebug)
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: logLevel}))

	client, err := githubapi.NewGithubClient(logger)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(2)
	}
	r := setupRouter(client.GetClient(logger), logger)

	logger.Info("server is running", slog.String("port", "8080"))
	r.Run(":8080")
}
