package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/eeboAvitoLovers/eal-backend/internal/app"
	"github.com/eeboAvitoLovers/eal-backend/internal/config"
)

const configFilename = "./internal/config/config.yaml"


func main() {
	// Загружаем конфиг из файла config.yaml
	config, err := config.LoadConfig(configFilename)
	if err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	// Создаем контекст для изящного завершения работы, ожидающего сигналы SIGNAL INTERRUPT 
	// и сигнал SIGNAL TERMINATE
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Создаем новый инстанс приложения 
	a := app.NewApp(ctx, config.CreateConnString())

	err = a.Start(ctx, config)
	if err != nil {
		log.Fatal("failed to start new app:", err)
	}
}
