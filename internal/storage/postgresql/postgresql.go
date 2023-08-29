package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
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

var (
	ErrorNotFound   = errors.New(`can not get order by number`)
	ErrorOrderAdded = errors.New(`order number added yet by this user`)
	ErrorOrderExist = errors.New(`order number added yet by another user`)
)

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.NewStorage"

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

func (s *Storage) GetOrder(ctx context.Context, number string) (*models.Order, error) {
	stmt, err := s.db.Prepare("SELECT number, user_login, status, accrual, uploaded_at FROM orders WHERE number=$1")
	if err != nil {
		return nil, err
	}
	o := &models.Order{}
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

func (s *Storage) AddWithdrawal(ctx context.Context, wd *models.Withdraw) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO withdrawals (order_number, user_login, sum, processed_at) VALUES ($1, $2, $3, $4)",
		wd.Order, wd.User, wd.Sum, wd.ProcessedAt)
	if err != nil {
		return errors.New(`order not added`)
	}
	return nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, login string) ([]*models.Withdraw, error) {
	stmt, err := s.db.Prepare("SELECT order_number, user_login, sum, processed_at FROM withdrawals WHERE user_login=$1 ORDER BY processed_at DESC")
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

	withdrawals := make([]*models.Withdraw, 0, len(columns))

	for rows.Next() {
		w := &models.Withdraw{}
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