package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kholodmv/gophermart/internal/auth"
	"github.com/kholodmv/gophermart/internal/logger/sl"
	"github.com/kholodmv/gophermart/internal/models"
	"golang.org/x/exp/slog"
	"net/http"
	"regexp"
	"time"
)

func (mh *Handler) Register(res http.ResponseWriter, req *http.Request) {
	const op = "handler.Register"

	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	var newUser models.User
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newUser)
	if err != nil {
		http.Error(res, "Invalid request format", http.StatusBadRequest)
		mh.log.Error("Invalid request format", sl.Err(err))
		return
	}

	hashUser, err := newUser.GenerateHashPassword()
	err = mh.db.AddUser(req.Context(), hashUser)
	if err != nil {
		mh.log.Error("New user has not been register", sl.Err(err))
		res.WriteHeader(http.StatusConflict)
		return
	}
	mh.log.Info("User successfully registered")

	tokenString, err := auth.GenerateToken(newUser.Login)
	if err != nil {
		http.Error(res, "Error creating token", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", tokenString)
	res.WriteHeader(http.StatusOK)
	fmt.Fprintln(res, "User successfully registered and authenticated")
}

func (mh *Handler) Login(res http.ResponseWriter, req *http.Request) {
	var credentials models.User
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&credentials)
	if err != nil {
		http.Error(res, "Invalid request format", http.StatusBadRequest)
		return
	}

	_, err = mh.db.GetUser(req.Context(), credentials.Login)
	if err != nil {
		http.Error(res, "Invalid username/password pair", http.StatusUnauthorized)
		return
	}

	tokenString, err := auth.GenerateToken(credentials.Login)
	if err != nil {
		http.Error(res, "Error creating token", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", tokenString)

	res.WriteHeader(http.StatusOK)
	fmt.Fprintln(res, "User successfully authenticated")
}

func (mh *Handler) PostOrderNumber(res http.ResponseWriter, req *http.Request) {
	var order models.Order
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&order)
	if err != nil {
		http.Error(res, "Invalid request format", http.StatusBadRequest)
		return
	}

	validNumberPattern := regexp.MustCompile("^[0-9]+$")
	if !validNumberPattern.MatchString(order.Number) {
		http.Error(res, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	if !models.IsValidLuhnNumber(order.Number) {
		http.Error(res, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}
	login := auth.GetLogin(req)
	fullOrder, _ := models.NewOrder(&order, login)
	err = mh.db.AddOrder(req.Context(), fullOrder)
	if err != nil {

	}
	res.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(res, "New order number accepted for processing")
}

func (mh *Handler) GetOrderNumbers(res http.ResponseWriter, req *http.Request) {
	login := auth.GetLogin(req)
	orders, err := mh.db.GetOrders(req.Context(), login)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) > 0 {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(orders)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (mh *Handler) GetBalance(res http.ResponseWriter, req *http.Request) {
	login := auth.GetLogin(req)

	currentBalance, err := mh.db.GetAccruals(req.Context(), login)
	withdrawnPoints, err := mh.db.GetWithdrawn(req.Context(), login)

	balance := models.Balance{
		Current:   currentBalance - withdrawnPoints,
		Withdrawn: withdrawnPoints,
	}

	responseJSON, err := json.Marshal(balance)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(responseJSON)
}

func (mh *Handler) PostWithdrawFromBalance(res http.ResponseWriter, req *http.Request) {
	var wd models.Withdraw
	if err := json.NewDecoder(req.Body).Decode(&wd); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	login := auth.GetLogin(req)
	wd.User = login
	createdTime := time.Now()
	wd.ProcessedAt = &createdTime

	currentBalance, _ := mh.db.GetAccruals(req.Context(), login)
	withdrawnPoints, _ := mh.db.GetWithdrawn(req.Context(), login)

	balance := models.Balance{
		Current:   currentBalance - withdrawnPoints,
		Withdrawn: withdrawnPoints,
	}
	if wd.Sum > balance.Current {
		http.Error(res, "there are not enough funds on the account", http.StatusPaymentRequired)
		return
	}

	err := mh.db.AddWithdrawal(req.Context(), &wd)
	if err != nil {
		http.Error(res, "statusConflict", http.StatusConflict)
	}

	if !models.IsValidLuhnNumber(wd.Order) {
		http.Error(res, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (mh *Handler) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	login := auth.GetLogin(req)

	withdrawals, err := mh.db.GetWithdrawals(req.Context(), login)
	if err != nil {
		//http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	/*if len(withdrawals) == 0 {
		//http.Error(res, "StatusNoContent", http.StatusNoContent)
		res.WriteHeader(http.StatusNoContent)
		return
	}*/
	//return c.JSON(http.StatusOK, withdrawals)

	res.Header().Set("Content-Type", "application/json")
	if len(withdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(withdrawals)
	}
}
