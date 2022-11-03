package router

import (
	"context"
	"crypto/rand"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/vukit/gomac/internal/gophermart/handlers"
	"github.com/vukit/gomac/internal/gophermart/logger"
	"github.com/vukit/gomac/internal/gophermart/repositories"
)

func NewRouter(ctx context.Context, repo repositories.Repo, mLogger *logger.Logger) (r chi.Router, err error) {
	secret, err := generateSecret(32)
	if err != nil {
		return nil, err
	}

	tokenAuth := jwtauth.New("HS256", secret, nil)

	r = chi.NewRouter()

	r.Use(middleware.Compress(5))

	h := handlers.NewHandler(tokenAuth, repo, mLogger)

	r.Get("/", h.Index)

	r.Post("/api/user/register", h.Register(ctx))

	r.Post("/api/user/login", h.Login(ctx))

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Post("/api/user/orders", h.Order(ctx))
		r.Get("/api/user/orders", h.Orders(ctx))
		r.Get("/api/user/balance", h.Balance(ctx))
		r.Post("/api/user/balance/withdraw", h.Withdraw(ctx))
		r.Get("/api/user/balance/withdrawals", h.Withdrawals(ctx))
	})

	return r, nil
}

func generateSecret(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
