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

func (s *Storage) GetUser(ctx context.Context, login string) (*models.User, error) {
	u := new(models.User)
	row := s.db.QueryRowContext(ctx,
		"SELECT login, pass_hash FROM users WHERE login = $1", login)

	if err := row.Scan(&u.Login, &u.HashPassword); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Storage) AddOrder(ctx context.Context, o *models.Order) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO orders (number, user_login, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)",
		o.Number, o.UserLogin, o.Status, o.Accrual, o.UploadedAt)
	if err != nil {
		return errors.New(`order not added`)
	}
	return nil
}

func (s *Storage) GetOrders(ctx context.Context, login string) ([]*models.Order, error) {
	stmt, err := s.db.Prepare("SELECT number, user_login, status, accrual, uploaded_at FROM orders WHERE user_login = $1 ORDER BY uploaded_at DESC")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	orders := make([]*models.Order, 0, len(columns))

	for rows.Next() {
		o := &models.Order{}
		err = rows.Scan(&o.Number, &o.UserLogin, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *Storage) GetAccruals(ctx context.Context, login string) (int64, error) {
	var accrual sql.NullInt64
	row := s.db.QueryRowContext(ctx,
		"SELECT sum(accrual) FROM orders WHERE login = $1", login)

	if err := row.Scan(&accrual); err != nil {
		return 0, err
	}
	return accrual.Int64, nil
}

func (s *Storage) GetWithdrawn(ctx context.Context, login string) (int64, error) {
	var withdrawn sql.NullInt64
	row := s.db.QueryRowContext(ctx,
		"SELECT sum(sum) FROM withdrawals WHERE user_login = $1", login)

	if err := row.Scan(&withdrawn); err != nil {
		return 0, err
	}
	return withdrawn.Int64, nil
}
