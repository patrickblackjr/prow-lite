package logging

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

var logLevel = new(slog.LevelVar)

func SetupLogging() *slog.Logger {
	logLevel.Set(slog.LevelDebug)

	tintOptions := &tint.Options{Level: logLevel, TimeFormat: "2006-01-02 15:04:05"}
	logger := slog.New(tint.NewHandler(os.Stdout, tintOptions))

	return logger
}
