package storage

import (
	"context"
	"github.com/kholodmv/gophermart/internal/models/order"
	"github.com/kholodmv/gophermart/internal/models/user"
	"github.com/kholodmv/gophermart/internal/models/withdraw"
)

type Storage interface {
	AddUser(ctx context.Context, user user.User) error
	GetUser(ctx context.Context, login string) (*user.User, error)

	AddOrder(ctx context.Context, o order.Order) error
	GetOrders(ctx context.Context, login string) ([]*order.Order, error)
	GetOrder(ctx context.Context, number order.Number) (*order.Order, error)
	GetOrderWithStatuses(ctx context.Context, processing order.Status, new order.Status) ([]order.Number, error)

	UpdateOrder(ctx context.Context, o order.Order) error

	GetAccruals(ctx context.Context, login string) (float32, error)

	GetWithdrawn(ctx context.Context, login string) (float32, error)
	GetWithdrawals(ctx context.Context, login string) ([]*withdraw.Withdraw, error)
	AddWithdrawal(ctx context.Context, wd withdraw.Withdraw, login string) (*withdraw.Withdraw, error)
}
