package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kholodmv/gophermart/internal/models/withdraw"
	"github.com/kholodmv/gophermart/internal/storage/postgresql"
	"github.com/kholodmv/gophermart/internal/utils"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

func (mh *Handler) PostWithdrawFromBalance(res http.ResponseWriter, req *http.Request) {
	const op = "withdrawal_handler.PostWithdrawFromBalance"
	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	var wd withdraw.Withdraw
	if err := json.NewDecoder(req.Body).Decode(&wd); err != nil {
		mh.log.Error("Invalid request format")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	login := utils.GetLogin(req.Context())

	wd.User = login
	createdTime := time.Now()
	wd.ProcessedAt = &createdTime

	w, err := mh.db.AddWithdrawal(req.Context(), wd, login)
	switch err {
	case postgresql.ErrorNotEnoughFunds:
		http.Error(res, "there are not enough funds on the account", http.StatusPaymentRequired)
		return
	case postgresql.ErrorAddWithdrawal:
		http.Error(res, "statusConflict", http.StatusConflict)
		return
	}

	if !utils.IsValidLuhnNumber(w.Order) {
		mh.log.Error("invalid order number format")
		http.Error(res, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	mh.log.Info("post withdraw from balance successful")
	res.WriteHeader(http.StatusOK)
}

func (mh *Handler) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	const op = "withdrawal_handler.GetWithdrawals"
	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	login := utils.GetLogin(req.Context())

	withdrawals, err := mh.db.GetWithdrawals(req.Context(), login)
	if err != nil {
		mh.log.Error("error get withdrawals")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if len(withdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(withdrawals)
	}
}

func (mh *Handler) GetBalance(res http.ResponseWriter, req *http.Request) {
	const op = "withdrawal_handler.GetBalance"
	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	login := utils.GetLogin(req.Context())

	currentBalance, err := mh.db.GetAccruals(req.Context(), login)
	if err != nil {
		mh.log.Error("error get accruals")
	}
	withdrawnPoints, err := mh.db.GetWithdrawn(req.Context(), login)
	if err != nil {
		mh.log.Error("error get withdrawn")
	}

	balance := withdraw.Balance{
		Current:   currentBalance - withdrawnPoints,
		Withdrawn: withdrawnPoints,
	}

	responseJSON, err := json.Marshal(balance)
	if err != nil {
		mh.log.Error("error responseJSON")
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	mh.log.Info("get balance successful")
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(responseJSON)
}
