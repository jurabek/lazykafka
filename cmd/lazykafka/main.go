package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/jurabek/lazykafka/internal/tui"
)

func main() {
	logFile, err := os.OpenFile("lazykafka.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, nil))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := tui.NewApp(ctx)
	if err != nil {
		log.Printf("failed to create app: %v", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		log.Printf("failed to run app: %v", err)
		os.Exit(1)
	}
}
