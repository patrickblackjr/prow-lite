package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
	eventplugin "github.com/patrickblackjr/prow-lite/cmd/event"
	"github.com/patrickblackjr/prow-lite/cmd/labelsync"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	"github.com/patrickblackjr/prow-lite/internal/logging"
	sloggin "github.com/samber/slog-gin"
	"github.com/urfave/cli/v3"
)

var osExit = os.Exit
var runServer = func(r *gin.Engine, addr string) error { return r.Run(addr) }

func main() {
	logger := logging.SetupLogging()

	var (
		mode   string
		plugin string
		event  string
	)

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
						Value:       "event",
					},
					&cli.StringFlag{
						Name:        "event",
						Aliases:     []string{"e"},
						Usage:       "Content of the event to be handled. --plugin flag must be event to use this.",
						Destination: &event,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					client, err := githubapi.NewGithubClient(logger)
					if err != nil {
						logger.Error("failed to create GitHub client", slog.String("error", err.Error()))
						osExit(2)
						return nil
					}
					runAction(ctx, mode, plugin, event, client.GetClient(), logger)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logger.Error(err.Error())
	}
}

func runAction(ctx context.Context, mode, plugin, event string, client *github.Client, logger *slog.Logger) {
	processComment := eventplugin.NewProcessComment(logger)
	handlePR := eventplugin.NewPREventHandler(logger)

	switch mode {
	case "standalone":
		r := setupRouter(client, logger, processComment, handlePR)
		logger.Info("server is running", slog.String("port", "8080"))
		if err := runServer(r, ":8080"); err != nil {
			logger.Error("failed to run server", slog.String("error", err.Error()))
		}

	case "ci":
		switch plugin {
		case "event":
			if event == "" {
				logger.Error("--event flag is required when --plugin=event")
				osExit(1)
				return
			}
			var payload map[string]any
			if err := json.Unmarshal([]byte(event), &payload); err != nil {
				logger.Error("failed to parse event payload", slog.String("error", err.Error()))
				osExit(1)
				return
			}
			eventType, ok := payload["action"].(string)
			if !ok || eventType == "" {
				logger.Error("failed to determine event type from payload")
				osExit(1)
				return
			}
			handleEvent := githubapi.RegisterEventHandlers(nil, client, logger, processComment, handlePR)
			handleEvent(eventType, []byte(event))

		case "labelsync", "label-sync":
			lsc, err := labelsync.GetLabelSyncConfig(logger)
			if err != nil {
				logger.Error("failed to load label sync config", slog.String("error", err.Error()))
				osExit(1)
				return
			}
			logger.Info("starting label sync",
				slog.Int("categories", len(lsc.Categories)),
				slog.Int("extra_labels", len(lsc.ExtraLabels)),
			)
			syncer := labelsync.NewSyncer(lsc, client, logger)
			if err := syncer.Run(ctx); err != nil {
				logger.Error("label sync failed", slog.String("error", err.Error()))
				osExit(1)
				return
			}
			logger.Info("label sync complete")

		default:
			logger.Error("unknown plugin", slog.String("plugin", plugin))
			osExit(1)
		}
	}
}

func setupRouter(client *github.Client, logger *slog.Logger, processComment func(*github.IssueCommentEvent, *github.Client, *slog.Logger), handlePR func(*github.PullRequestEvent, *github.Client, *slog.Logger)) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(sloggin.New(logger))
	r.Use(gin.Recovery())

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	githubapi.RegisterEventHandlers(r, client, logger, processComment, handlePR)

	return r
}
