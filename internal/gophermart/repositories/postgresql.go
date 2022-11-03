package repositories

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/vukit/gomac/internal/gophermart/models"

	// Register pgx stdlib
	_ "github.com/jackc/pgx/v4/stdlib"
)

type RepoPostgreSQL struct {
	db *sql.DB
}

func NewRepositoryPostgreSQL(dsn string) (repo RepoPostgreSQL, err error) {
	db, err := sql.Open("pgx", dsn)

	repo = RepoPostgreSQL{db: db}

	if err != nil {
		return repo, err
	}

	return repo, err
}

func (repo RepoPostgreSQL) SaveClient(ctx context.Context, client models.Client) (clientID int, err error) {
	if repo.db == nil {
		return 0, ErrNoDBConn
	}

	passwordHash := sha256.Sum256([]byte(client.Password))

	err = repo.db.QueryRowContext(ctx,
		`INSERT INTO clients (login, password) VALUES($1, $2) RETURNING client_id`,
		client.Login, hex.EncodeToString(passwordHash[:])).Scan(&clientID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrLoginIsAlreadyTaken
		}

		return 0, err
	}

	return clientID, err
}

func (repo RepoPostgreSQL) FindClient(ctx context.Context, client models.Client) (clientID int, err error) {
	if repo.db == nil {
		return 0, ErrNoDBConn
	}

	passwordHash := sha256.Sum256([]byte(client.Password))

	err = repo.db.QueryRowContext(ctx,
		`SELECT client_id FROM clients WHERE login = $1 and password = $2`,
		client.Login, hex.EncodeToString(passwordHash[:])).Scan(&clientID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrInvalidLoginPasswordPair
		default:
			return 0, err
		}
	}

	return clientID, err
}

func (repo RepoPostgreSQL) SaveOrder(ctx context.Context, order *models.Order) (err error) {
	var dbClientID int

	if repo.db == nil {
		return ErrNoDBConn
	}

	err = repo.db.QueryRowContext(ctx,
		`SELECT client_id FROM orders WHERE order_number = $1`,
		order.Number).Scan(&dbClientID)

	if err == nil {
		if order.ClientID == dbClientID {
			return ErrOrderNumberUploadedThisClient
		}

		return ErrOrderNumberUploadedAnotherClient
	}

	_, err = repo.db.ExecContext(ctx,
		`INSERT INTO orders (client_id, order_number, status, uploaded_at) VALUES($1, $2, $3, now())`,
		order.ClientID, order.Number, "NEW")
	if err != nil {
		return ErrOrderNumberUploadedAnotherClient
	}

	return err
}

func (repo RepoPostgreSQL) FindOrders(ctx context.Context, client models.Client) (orders []models.Order, err error) {
	if repo.db == nil {
		return nil, ErrNoDBConn
	}

	rows, err := repo.db.QueryContext(ctx,
		"SELECT order_number, accrual, status, uploaded_at FROM orders WHERE client_id = $1 ORDER BY uploaded_at",
		client.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders = make([]models.Order, 0)

	for rows.Next() {
		order := models.Order{}

		err = rows.Scan(&order.Number, &order.Accrual, &order.Status, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, err
}

func (repo RepoPostgreSQL) SaveWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) (err error) {
	if repo.db == nil {
		return ErrNoDBConn
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil && tx != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("save withdrawal: tx err %w: roll back err %v", err, rbErr)
			}
		}
	}()

	newBalance := 0

	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE((SELECT sum(accrual) FROM orders WHERE status = 'PROCESSED' AND client_id = $1), 0) - 
				COALESCE((SELECT sum(sum) FROM withdrawals WHERE client_id = $1), 0) - $2 as balance`,
		withdrawal.ClientID, withdrawal.Sum).Scan(&newBalance)
	if err != nil {
		return err
	}

	if newBalance < 0 {
		return ErrThereAreNotEnoughAccrual
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO withdrawals (client_id, order_number, sum, processed_at) VALUES($1, $2, $3, now())`,
		withdrawal.ClientID, withdrawal.Order, withdrawal.Sum)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (repo RepoPostgreSQL) FindWithdrawals(ctx context.Context, client models.Client) (withdrawals []models.Withdrawal, err error) {
	if repo.db == nil {
		return nil, ErrNoDBConn
	}

	rows, err := repo.db.QueryContext(ctx,
		"SELECT order_number, sum, processed_at FROM withdrawals WHERE client_id = $1 ORDER BY processed_at",
		client.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	withdrawals = make([]models.Withdrawal, 0)

	for rows.Next() {
		withdrawal := models.Withdrawal{}

		err = rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, err
}

func (repo RepoPostgreSQL) FindBalance(ctx context.Context, client models.Client) (balance *models.Balace, err error) {
	if repo.db == nil {
		return nil, ErrNoDBConn
	}

	balance = &models.Balace{}

	accurals := float64(0)

	err = repo.db.QueryRowContext(ctx,
		`SELECT COALESCE((SELECT sum(accrual) FROM orders WHERE status = 'PROCESSED' AND client_id = $1), 0) as accruals,
				COALESCE((SELECT sum(sum) FROM withdrawals WHERE client_id = $1), 0) as withdrawn`,
		client.ID).Scan(&accurals, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}

	balance.Current = accurals - balance.Withdrawn

	return balance, err
}

func (repo RepoPostgreSQL) SaveTask(ctx context.Context, task models.Task) (err error) {
	if repo.db == nil {
		return ErrNoDBConn
	}

	_, err = repo.db.ExecContext(ctx,
		`UPDATE orders SET accrual = $1, status = $2 WHERE order_id = $3`,
		task.Accrual, task.Status, task.OrderID)
	if err != nil {
		return err
	}

	return err
}

func (repo RepoPostgreSQL) FindTasks(ctx context.Context, statuses ...string) (tasks []models.Task, err error) {
	if repo.db == nil {
		return nil, ErrNoDBConn
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil && tx != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("find tasks: tx err %w: roll back err %v", err, rbErr)
			}
		}
	}()

	query := `SELECT order_id, order_number FROM orders WHERE status IN ('` + strings.Join(statuses, "','") + `') FOR UPDATE`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	if errRows := rows.Err(); errRows != nil {
		return nil, errRows
	}

	defer rows.Close()

	tasks = make([]models.Task, 0)

	for rows.Next() {
		task := models.Task{}

		err = rows.Scan(&task.OrderID, &task.OrderNumber)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE orders SET status = 'PROCESSING' WHERE status = 'NEW'`)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (repo RepoPostgreSQL) Close() error {
	if repo.db == nil {
		return ErrNoDBConn
	}

	return repo.db.Close()
}

func (repo RepoPostgreSQL) Ping(ctx context.Context) error {
	if repo.db == nil {
		return ErrNoDBConn
	}

	return repo.db.PingContext(ctx)
}
