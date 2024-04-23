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

type App struct {
	router http.Handler
	pgpool *pgxpool.Pool
}

func NewApp(ctx context.Context, connString string) *App {
	pgpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(fmt.Fprintf(os.Stderr, "unable to create connections:: %v\n", err))
	}

	a := &App{
		pgpool: pgpool,
	}
	a.newRoutes()
	return a
}

func (a *App) Start(ctx context.Context,c config.Config ) error {
	addrStr := c.Server.Hostname + ":" + strconv.Itoa(c.Server.Port)
	server := &http.Server{
		Addr: addrStr,
		Handler: a.router,

		ReadTimeout: time.Duration(c.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.Server.WriteTimeout) * time.Second,
	}

	err := a.pgpool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer a.pgpool.Close()

	log.Println("Starting server")

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server %w", err)
		}
		close(ch)
	}()

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
