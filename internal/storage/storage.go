package storage

import (
	"context"
	"github.com/kholodmv/gophermart/internal/models"
)

type Storage interface {
	AddUser(ctx context.Context, user *models.User) error
}
