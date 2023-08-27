package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/kholodmv/gophermart/internal/config"
	"github.com/kholodmv/gophermart/internal/http-server/handlers"
	"github.com/kholodmv/gophermart/internal/logger"
	"github.com/kholodmv/gophermart/internal/logger/sl"
	"github.com/kholodmv/gophermart/internal/storage/postgreSQL"
	_ "github.com/lib/pq"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var secretKey []byte

func main() {
	cfg := config.UseServerStartParams()

	log := logger.SetupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	db, err := postgreSQL.New(cfg.DatabaseUri)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
	}

	router := chi.NewRouter()

	handler := handlers.NewHandler(router, log, db)
	handler.RegisterRoutes(router)

	log.Info("initializing server", slog.String("address", cfg.RunAddress))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
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

	<-done
	log.Info("stopping server")
}
