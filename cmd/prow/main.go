package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
	"github.com/patrickblackjr/prow-lite/cmd/labelsync"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"github.com/patrickblackjr/prow-lite/internal/logging"
	sloggin "github.com/samber/slog-gin"
	"github.com/urfave/cli/v3"
)

var (
	mode   string
	plugin string
	event  string
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
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "plugin",
						Aliases:     []string{"p"},
						Usage:       "Plugin to run. One of: event, labelsync",
						Destination: &plugin,
						Required:    false,
						DefaultText: "event",
					},
					&cli.StringFlag{
						Name:        "event",
						Aliases:     []string{"e"},
						Usage:       "Content of the event to be handled. --plugin flag must be event to use this.",
						Destination: &event,
						Required:    false,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					client, err := githubapi.NewGithubClient(logger)
					if err != nil {
						slog.Error(err.Error())
						os.Exit(2)
					}

					if mode == "standalone" {
						r := setupRouter(client.GetClient(logger), logger)

						logger.Info("server is running", slog.String("port", "8080"))
						if err := r.Run(":8080"); err != nil {
							logger.Error("failed to run server", slog.Any("error", err.Error()))
						}
					}

					if mode == "ci" {
						if plugin == "event" {
							if event == "" {
								logger.Error("event flag is required in CI mode with plugin 'event'")
								os.Exit(1)
							}

							// Validate and parse the JSON payload
							var eventPayload map[string]interface{}
							if err := json.Unmarshal([]byte(event), &eventPayload); err != nil {
								logger.Error("failed to parse event payload", slog.String("error", err.Error()))
								os.Exit(1)
							}

							// Extract the event type (e.g., "issue_comment")
							eventType, ok := eventPayload["action"].(string)
							if !ok || eventType == "" {
								logger.Error("failed to determine event type from payload")
								os.Exit(1)
							}

							handleEvent := githubapi.RegisterEventHandlers(nil, client.GetClient(logger), logger, githubapi.ProcessComment)
							handleEvent(eventType, []byte(event))
						}
						if plugin == "labelsync" || plugin == "label-sync" {
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
			"status": "ok",
		})
	})

	githubapi.RegisterEventHandlers(r, client, logger, githubapi.ProcessComment)

	return r
}
