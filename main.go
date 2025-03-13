package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"github.com/lmittmann/tint"
	"github.com/patrickblackjr/prow-lite/cmd/labelsync"
	"github.com/patrickblackjr/prow-lite/internal/githubapi"
	sloggin "github.com/samber/slog-gin"
)

var logLevel = new(slog.LevelVar)

var (
	runMode string
	module  string
)

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

func main() {
	logLevel.Set(slog.LevelDebug)
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: logLevel}))

	client, err := githubapi.NewGithubClient(logger)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(2)
	}

	flag.StringVar(&runMode, "run-mode", "", "run-mode: standalone or ci")
	flag.StringVar(&module, "module", "", "module: prow or labelsync")
	flag.Parse()

	// default to standalone
	if runMode == "" && module == "" || runMode == "standalone" {
		r := setupRouter(client.GetClient(logger), logger)

		logger.Info("server is running", slog.String("port", "8080"))
		if err := r.Run(":8080"); err != nil {
			logger.Error("failed to run server", slog.Any("error", err.Error()))
		}
	}

	if runMode == "ci" {
		if module == "prow" {
			logger.Error("Prow CI mode not implemented yet")
			os.Exit(1)
		}
		if module == "labelsync" || module == "label-sync" || module == "" {
			labelsync.LabelSync()
			logger.Info("label sync completed")
		}

	}
}
