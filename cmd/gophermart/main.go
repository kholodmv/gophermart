package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/kholodmv/gophermart/internal/client"
	"github.com/kholodmv/gophermart/internal/config"
	"github.com/kholodmv/gophermart/internal/http-server/handlers"
	"github.com/kholodmv/gophermart/internal/logger"
	"github.com/kholodmv/gophermart/internal/logger/sl"
	"github.com/kholodmv/gophermart/internal/storage/postgresql"
	_ "github.com/lib/pq"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg := config.UseServerStartParams()

	log := logger.SetupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	db, err := postgresql.New(cfg.DatabaseURI)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
	}

	router := chi.NewRouter()

	handler := handlers.NewHandler(router, log, db)
	handler.RegisterRoutes()

	log.Info("initializing server", slog.String("address", cfg.RunAddress))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()
	log.Info("server started")

	var wg sync.WaitGroup
	done := make(chan struct{})
	c := client.New(cfg.AccrualSystemAddress, db, cfg.IntervalAccrualSystem, log)
	go func() {
		wg.Add(1)
		c.ReportOrders(done, &wg)
	}()
	go func() {
		wg.Wait()
		close(done)
	}()
	for range done {
	}

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("error", err)
	}

	log.Info("stopping server")
}
