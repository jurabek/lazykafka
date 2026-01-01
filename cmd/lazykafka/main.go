package main

import (
	"log"
	"os"

	"github.com/jurabek/lazykafka/internal/tui"
)

func main() {
	app, err := tui.NewApp()
	if err != nil {
		log.Printf("failed to create app: %v", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		log.Printf("failed to run app: %v", err)
		os.Exit(1)
	}
}
