package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kholodmv/gophermart/internal/http-server/middleware/auth"
	"github.com/kholodmv/gophermart/internal/http-server/middleware/gzip"
	mwLogger "github.com/kholodmv/gophermart/internal/http-server/middleware/logger"
	"github.com/kholodmv/gophermart/internal/storage"
	"golang.org/x/exp/slog"
)

type Handler struct {
	router chi.Router
	log    *slog.Logger
	db     storage.Storage
}

func NewHandler(router chi.Router, log *slog.Logger, db storage.Storage) *Handler {
	h := &Handler{
		router: router,
		log:    log,
		db:     db,
	}

	return h
}

func (mh *Handler) RegisterRoutes() {
	mh.router.Use(middleware.RequestID)
	mh.router.Use(mwLogger.New(mh.log))
	mh.router.Use(middleware.Recoverer)
	mh.router.Use(middleware.URLFormat)
	mh.router.Use(gzip.GzipHandler)

	mh.router.Post("/api/user/register", mh.Register)
	mh.router.Post("/api/user/login", mh.Login)

	mh.router.Group(func(r chi.Router) {
		r.Use(auth.AuthenticationMiddleware)

		r.Post("/api/user/orders", mh.PostOrderNumber)
		r.Get("/api/user/orders", mh.GetOrderNumbers)
		r.Get("/api/user/balance", mh.GetBalance)
		r.Post("/api/user/balance/withdraw", mh.PostWithdrawFromBalance)
		r.Get("/api/user/withdrawals", mh.GetWithdrawals)
	})
}
