package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v68/github"
	"github.com/patrickblackjr/prow-lite/plugins"
	"github.com/rs/zerolog/log"
	sloggin "github.com/samber/slog-gin"
)

func initGithubApp() *github.Client {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 269804, 32477892, "prow-lite-qa.2025-01-13.private-key.pem")
	if err != nil {
		log.Error().Err(err).Msg("failed to create github app client")
	}
	client := github.NewClient(&http.Client{Transport: itr})
	return client
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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
	registerEventHandlers(r, client, logger, plugins.ProcessComment)

	logger.Info("server is running", slog.String("port", "8080"))
	r.Run(":8080")
}
