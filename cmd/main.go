package main

import (
	"context"
	"log/slog"

	"github.com/njayp/hermes/pkg/manager"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Error(manager.Run(context.Background()).Error())
}
