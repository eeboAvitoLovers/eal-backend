package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/eeboAvitoLovers/eal-backend/internal/app"
	"github.com/eeboAvitoLovers/eal-backend/internal/config"
)

const configFilename = "/home/slave/Documents/eal-backend/internal/config/config.yaml"

func main() {
	config, err := config.LoadConfig(configFilename)
	if err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := app.NewApp(ctx, config.CreateConnString())

	err = a.Start(ctx, config)
	if err != nil {
		log.Fatal("failed to start new app:", err)
	}
}
