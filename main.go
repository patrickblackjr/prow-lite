package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"github.com/rs/zerolog/log"
	sloggin "github.com/samber/slog-gin"
)

var logLevel = new(slog.LevelVar)

func initGithubApp() *github.Client {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 269804, 32477892, "prow-dev.pem")
	if err != nil {
		log.Error().Err(err).Msg("failed to create github app client")
	}
	client := github.NewClient(&http.Client{Transport: itr})
	return client
}

func main() {
	logLevel.Set(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(sloggin.New(logger))
	r.Use(gin.Recovery())

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	client := initGithubApp()
	githubapi.RegisterEventHandlers(r, client, logger, githubapi.ProcessComment)
	githubapi.EnsureLabels("patrickblackjr", []string{"lgtm"}, client, logger)

	logger.Info("server is running", slog.String("port", "8080"))
	r.Run(":8080")
}
