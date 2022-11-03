package repositories

import (
	"context"
	"errors"

	"github.com/vukit/gomac/internal/gophermart/models"
)

var (
	ErrNoDBConn                         = errors.New("no database connection")
	ErrLoginIsAlreadyTaken              = errors.New("login is already taken")
	ErrInvalidLoginPasswordPair         = errors.New("invalid login/password pair")
	ErrOrderNumberUploadedThisClient    = errors.New("order number has been uploaded by this client")
	ErrOrderNumberUploadedAnotherClient = errors.New("order number has been uploaded by another client")
	ErrThereAreNotEnoughAccrual         = errors.New("there are not enough accrual")
)

type Repo interface {
	SaveClient(context.Context, models.Client) (id int, err error)
	FindClient(context.Context, models.Client) (id int, err error)

	SaveOrder(context.Context, *models.Order) (err error)
	FindOrders(context.Context, models.Client) (orders []models.Order, err error)

	SaveWithdrawal(context.Context, *models.Withdrawal) (err error)
	FindWithdrawals(context.Context, models.Client) (withdrawals []models.Withdrawal, err error)

	FindBalance(context.Context, models.Client) (balance *models.Balace, err error)

	SaveTask(context.Context, models.Task) (err error)
	FindTasks(context.Context, ...string) (tasks []models.Task, err error)

	Ping(context.Context) error
	Close() error
}
