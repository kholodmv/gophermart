package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kholodmv/gophermart/internal/auth"
	"github.com/kholodmv/gophermart/internal/logger/sl"
	"github.com/kholodmv/gophermart/internal/models/user"
	"github.com/kholodmv/gophermart/internal/utils"
	"golang.org/x/exp/slog"
	"net/http"
)

func (mh *Handler) Register(res http.ResponseWriter, req *http.Request) {
	const op = "user_handler.Register"
	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	var newUser user.User
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newUser)
	if err != nil {
		mh.log.Error("Invalid request format", sl.Err(err))
		http.Error(res, "Invalid request format", http.StatusBadRequest)
		return
	}

	var hashPass string
	if newUser.HashPassword == "" {
		hashPass, err = utils.GenerateHashPassword(newUser.Password)
		newUser.HashPassword = hashPass
	}

	if err != nil {
		mh.log.Error("Error generate hash password")
	}
	err = mh.db.AddUser(req.Context(), newUser)
	if err != nil {
		mh.log.Error("New user has not been register", sl.Err(err))
		res.WriteHeader(http.StatusConflict)
		return
	}
	mh.log.Info("User successfully registered")

	tokenString, err := auth.GenerateToken(newUser.Login)
	if err != nil {
		mh.log.Error("Error creating token")
		http.Error(res, "Error creating token", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", "Bearer "+tokenString)
	mh.log.Info("tokenString is: " + tokenString)
	res.WriteHeader(http.StatusOK)
	mh.log.Info("User successfully registered and authenticated")
}

func (mh *Handler) Login(res http.ResponseWriter, req *http.Request) {
	const op = "user_handler.Login"
	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	var credentials user.User
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&credentials)
	if err != nil {
		mh.log.Error("Invalid request format")
		http.Error(res, "Invalid request format", http.StatusBadRequest)
		return
	}

	user1, err := mh.db.GetUser(req.Context(), credentials.Login)
	if err != nil {
		mh.log.Error("Invalid username/password pair")
		http.Error(res, "Invalid username/password pair", http.StatusUnauthorized)
		return
	}

	err = utils.CompareHashAndPassword(user1.HashPassword, credentials.Password)
	if err != nil {
		mh.log.Error("Invalid username/password pair")
		http.Error(res, "Invalid username/password pair", http.StatusUnauthorized)
		return
	}

	tokenString, err := auth.GenerateToken(credentials.Login)
	if err != nil {
		mh.log.Error("Error creating token")
		http.Error(res, "Error creating token", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", "Bearer "+tokenString)

	res.WriteHeader(http.StatusOK)
	mh.log.Info("User successfully authenticated")
}
