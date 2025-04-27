package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v70/github"
	"github.com/patrickblackjr/prow-lite/cmd/labelsync"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"github.com/patrickblackjr/prow-lite/internal/logging"
	sloggin "github.com/samber/slog-gin"
	"github.com/urfave/cli/v3"
)

var (
	mode   string
	plugin string
)

func main() {
	logger := logging.SetupLogging()

	cmd := &cli.Command{
		Name:    "prow",
		Usage:   "Prow Lite",
		Version: "v0.1.0",
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run prow lite",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "mode",
						Aliases:     []string{"m"},
						Usage:       "Provide the Prow mode. One of: standalone, ci",
						Destination: &mode,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					client, err := githubapi.NewGithubClient(logger)
					if err != nil {
						slog.Error(err.Error())
						os.Exit(2)
					}

					// default to standalone
					if mode == "" && plugin == "" || mode == "standalone" {
						r := setupRouter(client.GetClient(logger), logger)

						logger.Info("server is running", slog.String("port", "8080"))
						if err := r.Run(":8080"); err != nil {
							logger.Error("failed to run server", slog.Any("error", err.Error()))
						}
					}

					if mode == "ci" {
						if plugin == "event" {
							logger.Error("Prow Lite CI mode not implemented yet")
							os.Exit(1)
						}
						if plugin == "labelsync" || plugin == "label-sync" || plugin == "" {
							labelsync.LabelSync()
							logger.Info("label sync completed")
						}
					}
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logger.Error(err.Error())
	}
}

func setupRouter(client *github.Client, logger *slog.Logger) *gin.Engine {
	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		logger.Error("failed to set trusted proxies", slog.Any("error", err.Error()))
	}
	r.Use(sloggin.New(logger))
	r.Use(gin.Recovery())

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	githubapi.RegisterEventHandlers(r, client, logger, githubapi.ProcessComment)

	return r
}
