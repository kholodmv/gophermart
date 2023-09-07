package order

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type want struct {
	order *Order
}

type testCase struct {
	name   string
	order  Order
	login  string
	number int64
	want   want
}

var hashTests = []testCase{
	{
		name:   "Return full order with status 'New'",
		order:  Order{},
		login:  "login",
		number: 123456,
		want: want{
			order: &Order{
				Number:    "123456",
				UserLogin: "login",
				Status:    StatusNew,
			},
		},
	},
}

func TestNewOrder(t *testing.T) {
	for _, test := range hashTests {
		t.Run(test.name, func(t *testing.T) {
			order := NewOrder(test.order, test.login, test.number)
			assert.Equal(t, order.Number, test.want.order.Number)
			assert.Equal(t, order.UserLogin, test.want.order.UserLogin)
			assert.Equal(t, order.Status, test.want.order.Status)
		})
	}
}
