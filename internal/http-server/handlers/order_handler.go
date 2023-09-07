package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kholodmv/gophermart/internal/models/order"
	"github.com/kholodmv/gophermart/internal/storage/postgresql"
	"github.com/kholodmv/gophermart/internal/utils"
	"golang.org/x/exp/slog"
	"net/http"
	"regexp"
	"strconv"
)

func (mh *Handler) PostOrderNumber(res http.ResponseWriter, req *http.Request) {
	const op = "order_handler.PostOrderNumber"
	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	var number int64
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&number)
	if err != nil {
		mh.log.Error("Invalid request format")
		http.Error(res, "Invalid request format", http.StatusBadRequest)
		return
	}

	validNumberPattern := regexp.MustCompile("^[0-9]+$")
	if !validNumberPattern.MatchString(strconv.FormatInt(number, 10)) {
		mh.log.Error("Invalid order number format")
		http.Error(res, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	if !utils.IsValidLuhnNumber(strconv.FormatInt(number, 10)) {
		mh.log.Error("Invalid order number format")
		http.Error(res, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	login := utils.GetLogin(req.Context())

	var orderNew order.Order
	fullOrder := order.NewOrder(orderNew, login, number)
	err = mh.db.AddOrder(req.Context(), fullOrder)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrorOrderAdded):
			res.WriteHeader(http.StatusOK)
			mh.log.Error("The order number has already been added by this user - ", res)
			return
		case errors.Is(err, postgresql.ErrorOrderExist):
			res.WriteHeader(http.StatusConflict)
			mh.log.Error("The order number has already been added by another user - ", res)
			return
		default:
			res.WriteHeader(http.StatusInternalServerError)
			mh.log.Error("Error adding order number - ", res)
			return
		}
	}
	res.WriteHeader(http.StatusAccepted)
	mh.log.Info("New order number accepted for processing")
}

func (mh *Handler) GetOrderNumbers(res http.ResponseWriter, req *http.Request) {
	const op = "order_handler.GetOrderNumbers"
	mh.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(req.Context())),
	)

	login := utils.GetLogin(req.Context())

	orders, err := mh.db.GetOrders(req.Context(), login)
	if err != nil {
		mh.log.Error("New order number accepted for processing")
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
