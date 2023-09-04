package order

import (
	"strconv"
	"time"
)

type Order struct {
	UserLogin  string    `json:"-"`
	Number     Number    `json:"number"`
	Status     Status    `json:"status"`
	Accrual    float32   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Status string
type Number string

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

func NewOrder(order *Order, login string, number int64) *Order {
	createTime := time.Now()

	order.Number = Number(strconv.FormatInt(number, 10))
	order.UserLogin = login
	order.Status = StatusNew
	order.UploadedAt = createTime

	return order
}
