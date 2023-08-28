package storage

import (
	"context"
	"github.com/kholodmv/gophermart/internal/models"
)

type Storage interface {
	AddUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, login string) (*models.User, error)

	AddOrder(ctx context.Context, o *models.Order) error
	GetOrders(ctx context.Context, login string) ([]*models.Order, error)

	GetAccruals(ctx context.Context, login string) (int64, error)

	GetWithdrawn(ctx context.Context, login string) (int64, error)
}
