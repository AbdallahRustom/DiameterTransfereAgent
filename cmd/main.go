package main

import (
	"diametertransfereagent/internal/app"
	"diametertransfereagent/pkg/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	application := app.NewApp(cfg)
	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
