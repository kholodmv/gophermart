package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kholodmv/gophermart/internal/logger/sl"
	"github.com/kholodmv/gophermart/internal/models"
	"golang.org/x/exp/slog"
	"net/http"
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

	res.WriteHeader(http.StatusOK)
	fmt.Fprintln(res, "User successfully registered and authenticated")
}

func (mh *Handler) Login(res http.ResponseWriter, req *http.Request) {

}

func (mh *Handler) PostOrderNumber(res http.ResponseWriter, req *http.Request) {

}

func (mh *Handler) GetOrderNumbers(res http.ResponseWriter, req *http.Request) {

}

func (mh *Handler) GetBalance(res http.ResponseWriter, req *http.Request) {

}

func (mh *Handler) PostWithdrawFromBalance(res http.ResponseWriter, req *http.Request) {

}

func (mh *Handler) GetWithdrawals(res http.ResponseWriter, req *http.Request) {

}
