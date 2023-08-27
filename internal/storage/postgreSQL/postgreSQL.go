package postgreSQL

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/kholodmv/gophermart/internal/models"
)

type Storage struct {
	db *sql.DB
}

const tableUser = `
    CREATE TABLE IF NOT EXISTS users(
        login VARCHAR(256) PRIMARY KEY,
        pass_hash VARCHAR(256) NOT NULL);`

const tableOrder = `
	CREATE TABLE IF NOT EXISTS orders(
		number VARCHAR(256) PRIMARY KEY,
		user_login VARCHAR(256) NOT NULL,
		status VARCHAR(256) NOT NULL,
		accrual INT,
		uploaded_at TIMESTAMP NOT NULL);`

const tableWithdrawals = `
	CREATE TABLE IF NOT EXISTS withdrawals(
	    order_number VARCHAR(256) PRIMARY KEY,
	    user_login VARCHAR(256) NOT NULL,
		sum INT NOT NULL,
		processed_at TIMESTAMP NOT NULL);`

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgreSQL.NewStorage"

	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err = db.Exec(tableUser); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if _, err = db.Exec(tableOrder); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if _, err = db.Exec(tableWithdrawals); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) AddUser(ctx context.Context, u *models.User) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO users (login, pass_hash) VALUES ($1, $2)", u.Login, u.HashPassword)
	if err != nil {
		return errors.New(`user is exist`)
	}
	return nil
}
