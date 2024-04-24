// package app предоставляет функционал для создания и запуска веб-приложения.
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/eeboAvitoLovers/eal-backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App представляет собой веб-приложение.
type App struct {
	router http.Handler
	pgpool *pgxpool.Pool
}

// NewApp создает новый экземпляр веб-приложения.
func NewApp(ctx context.Context, connString string) *App {
	// Инициализация пула подключений к базе данных PostgreSQL.
	pgpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(fmt.Fprintf(os.Stderr, "unable to create connections:: %v\n", err))
	}

	a := &App{
		pgpool: pgpool,
	}
	a.newRoutes() // Загрузка маршрутов
	return a
}

// Start запускает веб-сервер.
func (a *App) Start(ctx context.Context, c config.Config) error {
	// Формирование адреса сервера.
	addrStr := c.Server.Hostname + ":" + strconv.Itoa(c.Server.Port)
	server := &http.Server{
		Addr:    addrStr,
		Handler: a.router,

		ReadTimeout:  time.Duration(c.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.Server.WriteTimeout) * time.Second,
	}

	// Проверка подключения к базе данных.
	err := a.pgpool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer a.pgpool.Close()

	log.Println("Starting server")

	ch := make(chan error, 1)

	// Запуск сервера в отдельной горутине.
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server %w", err)
		}
		close(ch)
	}()

	// Ожидание завершения работы сервера или отмены контекста.
	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		log.Println("Shutting down server")
		return server.Shutdown(timeout)
	}
}
