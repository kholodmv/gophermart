package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mwLogger "github.com/kholodmv/gophermart/internal/http-server/middleware/logger"
	"golang.org/x/exp/slog"
)

type Handler struct {
	router chi.Router
	log    *slog.Logger
}

func NewHandler(router chi.Router, log *slog.Logger) *Handler {
	h := &Handler{
		router: router,
		log:    log,
	}

	return h
}

func (mh *Handler) RegisterRoutes(router *chi.Mux) {
	mh.router.Use(middleware.RequestID)
	mh.router.Use(mwLogger.New(mh.log))
	mh.router.Use(middleware.Recoverer)
	mh.router.Use(middleware.URLFormat)

	router.Post("/api/user/register", mh.Register)
	router.Post("/api/user/login", mh.Login)
	router.Post("/api/user/orders", mh.PostOrderNumber)
	router.Get("/api/user/orders", mh.GetOrderNumbers)
	router.Get("/api/user/balance", mh.GetBalance)
	router.Post("/api/user/balance/withdraw", mh.PostWithdrawFromBalance)
	router.Get("/api/user/withdrawals", mh.GetWithdrawals)
}
