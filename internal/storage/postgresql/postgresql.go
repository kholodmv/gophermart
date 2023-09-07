package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kholodmv/gophermart/internal/models/order"
	"github.com/kholodmv/gophermart/internal/models/user"
	"github.com/kholodmv/gophermart/internal/models/withdraw"
	"golang.org/x/exp/slog"
)

type Storage struct {
	db  *sql.DB
	log *slog.Logger
}

const tableUser = `
    CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
        login VARCHAR(256) UNIQUE NOT NULL,
        pass_hash VARCHAR(256) NOT NULL);`

const tableOrder = `
	CREATE TABLE IF NOT EXISTS orders(
	    id SERIAL PRIMARY KEY,
		number VARCHAR(256) UNIQUE NOT NULL,
		user_login VARCHAR(256) NOT NULL,
		status VARCHAR(256) NOT NULL,
		accrual DOUBLE PRECISION,
		uploaded_at TIMESTAMP NOT NULL);`

const tableWithdrawals = `
	CREATE TABLE IF NOT EXISTS withdrawals(
	    id SERIAL PRIMARY KEY,
	    order_number VARCHAR(256) UNIQUE,
	    user_login VARCHAR(256) NOT NULL,
		sum DOUBLE PRECISION NOT NULL,
		processed_at TIMESTAMP NOT NULL);`

var (
	ErrorNotFound       = errors.New(`can not get order by number`)
	ErrorOrderAdded     = errors.New(`order number added yet by this user`)
	ErrorOrderExist     = errors.New(`order number added yet by another user`)
	ErrorNotEnoughFunds = errors.New(`there are not enough funds on the account`)
	ErrorAddWithdrawal  = errors.New(`error add withdrawal`)
)

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"

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

func (s *Storage) AddUser(ctx context.Context, u user.User) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO users (login, pass_hash) VALUES ($1, $2)", u.Login, u.HashPassword)
	if err != nil {
		return errors.New(`can not add user to db`)
	}
	return nil
}

func (s *Storage) GetUser(ctx context.Context, login string) (*user.User, error) {
	u := new(user.User)
	row := s.db.QueryRowContext(ctx,
		"SELECT login, pass_hash FROM users WHERE login = $1", login)

	if err := row.Scan(&u.Login, &u.HashPassword); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Storage) AddOrder(ctx context.Context, o order.Order) error {
	stmt, err := s.db.Prepare("INSERT INTO orders(number, user_login, status, accrual, uploaded_at) values($1,$2,$3,$4,$5)")
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(
		ctx,
		o.Number,
		o.UserLogin,
		o.Status,
		o.Accrual,
		o.UploadedAt,
	)
	if err != nil {
		existOrder, err := s.GetOrder(ctx, o.Number)
		if err != nil {
			return ErrorNotFound
		}
		if existOrder.UserLogin == o.UserLogin {
			return ErrorOrderAdded
		} else {
			return ErrorOrderExist
		}
	}
	return nil
}

func (s *Storage) GetOrder(ctx context.Context, number order.Number) (*order.Order, error) {
	stmt, err := s.db.Prepare("SELECT number, user_login, status, accrual, uploaded_at FROM orders WHERE number=$1")
	if err != nil {
		return nil, err
	}
	o := &order.Order{}
	row := stmt.QueryRowContext(ctx, number)
	err = row.Scan(&o.Number, &o.UserLogin, &o.Status, &o.Accrual, &o.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("%s: %w", errors.New("can't get order"), err)
	}
	return o, nil
}

func (s *Storage) GetOrders(ctx context.Context, login string) ([]*order.Order, error) {
	stmt, err := s.db.Prepare("SELECT number, user_login, status, accrual, uploaded_at FROM orders WHERE user_login = $1 ORDER BY uploaded_at DESC")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*order.Order, 0)

	for rows.Next() {
		o := &order.Order{}
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

func (s *Storage) GetOrderWithStatuses(ctx context.Context, processing order.Status, new order.Status) ([]order.Number, error) {
	stmt, err := s.db.Prepare("SELECT number FROM orders WHERE status=$1 OR status=$2 ORDER BY uploaded_at")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, &processing, &new)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]order.Number, 0)

	for rows.Next() {
		var o order.Order
		err = rows.Scan(&o.Number)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o.Number)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *Storage) UpdateOrder(ctx context.Context, o order.Order) error {
	stmt, err := s.db.Prepare("UPDATE orders SET status=$1, accrual=$2 WHERE number=$3")
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx,
		o.Status,
		o.Accrual,
		o.Number,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", errors.New("can't update order"), err)
	}
	return nil
}

func (s *Storage) GetAccruals(ctx context.Context, login string) (float32, error) {
	var accrual float32
	row := s.db.QueryRowContext(ctx,
		"SELECT sum(accrual) FROM orders WHERE user_login = $1", login)

	if err := row.Scan(&accrual); err != nil {
		return 0, err
	}
	return accrual, nil
}

func (s *Storage) GetWithdrawn(ctx context.Context, login string) (float32, error) {
	var withdrawn float32
	row := s.db.QueryRowContext(ctx,
		"SELECT sum(sum) FROM withdrawals WHERE user_login = $1", login)

	if err := row.Scan(&withdrawn); err != nil {
		return 0, err
	}
	return withdrawn, nil
}

func (s *Storage) AddWithdrawal(ctx context.Context, wd withdraw.Withdraw, login string) (*withdraw.Withdraw, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	var accrual float32
	row := tx.QueryRowContext(ctx,
		"SELECT sum(accrual) FROM orders WHERE user_login = $1", login)
	if err = row.Scan(&accrual); err != nil {
		s.log.Error("error get current balance")
		tx.Rollback()
		return nil, err
	}

	var withdrawn float32
	row = tx.QueryRowContext(ctx,
		"SELECT coalesce(SUM(sum), 0.00) FROM withdrawals WHERE user_login = $1", login)
	if err = row.Scan(&withdrawn); err != nil {
		s.log.Error("error get withdrawn")
		tx.Rollback()
		return nil, err
	}

	if wd.Sum > accrual-withdrawn {
		s.log.Error("there are not enough funds on the account")
		return nil, ErrorNotEnoughFunds
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO withdrawals (order_number, user_login, sum, processed_at) VALUES ($1, $2, $3, $4)",
		wd.Order, wd.User, wd.Sum, wd.ProcessedAt)
	if err != nil {
		s.log.Error("error add withdrawal")
		tx.Rollback()
		return nil, ErrorAddWithdrawal
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &wd, nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, login string) ([]*withdraw.Withdraw, error) {
	stmt, err := s.db.Prepare("SELECT order_number, user_login, sum, processed_at FROM withdrawals WHERE user_login=$1 ORDER BY processed_at DESC")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]*withdraw.Withdraw, 0)

	for rows.Next() {
		w := &withdraw.Withdraw{}
		err = rows.Scan(&w.Order, &w.User, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
