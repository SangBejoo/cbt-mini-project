package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"

	"cbt-test-mini-project/init/config"
)

func Load(cfgMain config.Main) {
	cfg := cfgMain.Log
	var defaultWriter io.Writer
	defaultWriter = os.Stdout
	fileName := cfg.Directory
	if len(fileName) > 0 {
		defaultWriter = &lumberjack.Logger{
			Filename: fileName,
		}
	}

	logLevel := slog.Level(cfg.Level)

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	jsonLogger := slog.NewJSONHandler(defaultWriter, opts)
	logger := slog.New(jsonLogger)

	slog.SetDefault(logger)
}
