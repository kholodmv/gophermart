package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/kholodmv/gophermart/internal/models/order"
	"github.com/kholodmv/gophermart/internal/storage"
	"golang.org/x/exp/slog"
	"net/http"
	"sync"
	"time"
)

const APIGetAccrual = `/api/orders/{number}`

var ErrorOrderNotRegistered = errors.New(`order isn't registered in system`)

const (
	StatusRegistered = `REGISTERED`
	StatusInvalid    = `INVALID`
	StatusProcessing = `PROCESSING`
	StatusProcessed  = `PROCESSED`
)

type Accrual struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Client struct {
	client   *resty.Client
	address  string
	db       storage.Storage
	interval int
	log      *slog.Logger
}

func New(address string, db storage.Storage, interval int, log *slog.Logger) *Client {
	return &Client{
		client:   resty.New().SetDebug(true),
		address:  address,
		db:       db,
		interval: interval,
		log:      log,
	}
}

func (c *Client) ReportOrders() {
	orderNumbers := make(chan order.Number)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t := time.NewTicker(time.Duration(c.interval) * time.Second)
		for {
			select {
			case <-t.C:
				processingOrders, err := c.db.GetOrderStatus(context.Background(), order.StatusProcessing)
				if err != nil {
					c.log.Error("there are no orders with status PROCESSING in the database", err)
					continue
				}
				newOrders, err := c.db.GetOrderStatus(context.Background(), order.StatusNew)
				if err != nil {
					c.log.Error("there are no orders with status NEW in the database", err)
					continue
				}
				allOrders := append(processingOrders, newOrders...)
				for _, number := range allOrders {
					orderNumbers <- number
				}
			}
		}
	}()
	for {
		select {
		case n := <-orderNumbers:
			o := &order.Order{
				Number: n,
			}
			a, err := c.GetStatusOrderFromAccrualSystem(n)
			switch err {
			case nil:
				o.Status = accrualToOrderStatus(a.Status)
				o.Accrual = a.Accrual
			case ErrorOrderNotRegistered:
				o.Status = order.StatusInvalid
			default:
				c.log.Error("default error - ", err)
				continue
			}
			err = c.db.UpdateOrder(context.Background(), o)
			if err != nil {
				c.log.Error("can not update order in database", err)
			}
		}
	}
}

func (c *Client) GetStatusOrderFromAccrualSystem(number order.Number) (*order.Order, error) {
	endpoint := fmt.Sprintf("%s%s", c.address, APIGetAccrual)
	//a := &Accrual{}
	o := &order.Order{}
	resp, err := c.client.R().
		SetPathParam("number", string(number)).
		SetResult(o).
		Get(endpoint)
	if err != nil {
		c.log.Error("error client response - ", err)
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return o, nil
	case http.StatusTooManyRequests:
		c.log.Info("too many requests")
		time.Sleep(60 * time.Second)
	case http.StatusNoContent:
		c.log.Info("no content")
		return nil, ErrorOrderNotRegistered
	}
	return nil, errors.New("invalid status code")
}

func accrualToOrderStatus(status order.Status) order.Status {
	switch status {
	case StatusRegistered:
		return order.StatusNew
	case StatusProcessing:
		return order.StatusProcessing
	case StatusInvalid:
		return order.StatusInvalid
	case StatusProcessed:
		return order.StatusProcessed
	}
	return order.StatusInvalid
}
