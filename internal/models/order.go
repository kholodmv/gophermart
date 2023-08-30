package models

import (
	"strconv"
	"time"
)

type Order struct {
	UserLogin  string     `json:"-"`
	Number     string     `json:"number"`
	Status     Status     `json:"status"`
	Accrual    *int64     `json:"accrual,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at"`
}

type Status string

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

func NewOrder(order *Order, login string, number int64) *Order {
	createTime := time.Now()

	order.Number = strconv.FormatInt(number, 2)
	order.UserLogin = login
	order.Status = StatusNew
	order.Accrual = nil
	order.UploadedAt = &createTime

	return order
}

func IsValidLuhnNumber(number string) bool {
	var sum int
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
