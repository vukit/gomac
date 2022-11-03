package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/vukit/gomac/internal/gophermart/logger"
	"github.com/vukit/gomac/internal/gophermart/models"
	"github.com/vukit/gomac/internal/gophermart/repositories"
)

type Handler struct {
	tokenAuth  *jwtauth.JWTAuth
	repository repositories.Repo
	mLogger    *logger.Logger
}

var ErrNotFindClientID = errors.New("not find client id")

func NewHandler(tokenAuth *jwtauth.JWTAuth, repo repositories.Repo, mLogger *logger.Logger) Handler {
	return Handler{
		tokenAuth:  tokenAuth,
		repository: repo,
		mLogger:    mLogger,
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	result := strings.Builder{}
	result.WriteString(`
	<!doctype html>
	<html lang="en">
	<head>
	  <meta charset="utf-8">
	  <title>Gophermart Index Page</title>
	</head>
	<body>
	<h1>Gophermart Index Page</h1>
	</body>
	</html>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintln(w, result.String())
}

func (h *Handler) Register(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		client, err := getClientFromBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		if err = client.Validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		clientID, err := h.repository.SaveClient(ctx, client)
		if err != nil {
			switch {
			case errors.Is(err, repositories.ErrLoginIsAlreadyTaken):
				w.WriteHeader(http.StatusConflict)
			default:
				w.WriteHeader(http.StatusUnauthorized)
			}

			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		if err = h.setJWToken(w, clientID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		fmt.Fprintf(w, "{}")
	}
}

func (h *Handler) Login(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		client, err := getClientFromBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		if err = client.Validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		clientID, err := h.repository.FindClient(ctx, client)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		if err = h.setJWToken(w, clientID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		fmt.Fprintf(w, "{}")
	}
}

func (h *Handler) Order(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		clientID, err := getClientID(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		order := models.Order{ClientID: clientID, Number: string(data)}

		if err = order.Validate(); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		err = h.repository.SaveOrder(ctx, &order)
		if err != nil {
			switch {
			case errors.Is(err, repositories.ErrOrderNumberUploadedThisClient):
				w.WriteHeader(http.StatusOK)
			case errors.Is(err, repositories.ErrOrderNumberUploadedAnotherClient):
				w.WriteHeader(http.StatusConflict)
			default:
				w.WriteHeader(http.StatusNotAcceptable)
			}

			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		w.WriteHeader(http.StatusAccepted)

		fmt.Fprintf(w, "{}")
	}
}

func (h *Handler) Orders(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		clientID, err := getClientID(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		orders, err := h.repository.FindOrders(ctx, models.Client{ID: clientID})
		if err != nil || len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		body := &bytes.Buffer{}
		encoder := json.NewEncoder(body)

		err = encoder.Encode(orders)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		_, err = w.Write(body.Bytes())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}
	}
}

func (h *Handler) Withdraw(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		clientID, err := getClientID(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		withdrawal := models.Withdrawal{ClientID: clientID}

		decoder := json.NewDecoder(r.Body)

		err = decoder.Decode(&withdrawal)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		if err = withdrawal.Validate(); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		err = h.repository.SaveWithdrawal(ctx, &withdrawal)
		if err != nil {
			switch {
			case errors.Is(err, repositories.ErrThereAreNotEnoughAccrual):
				w.WriteHeader(http.StatusPaymentRequired)
			default:
				w.WriteHeader(http.StatusNotAcceptable)
			}

			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		fmt.Fprintf(w, "{}")
	}
}

func (h *Handler) Withdrawals(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		clientID, err := getClientID(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		withdrawals, err := h.repository.FindWithdrawals(ctx, models.Client{ID: clientID})
		if err != nil || len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		body := &bytes.Buffer{}
		encoder := json.NewEncoder(body)

		err = encoder.Encode(withdrawals)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		_, err = w.Write(body.Bytes())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}
	}
}

func (h *Handler) Balance(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		clientID, err := getClientID(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		balance, err := h.repository.FindBalance(ctx, models.Client{ID: clientID})
		if err != nil {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		body := &bytes.Buffer{}
		encoder := json.NewEncoder(body)

		err = encoder.Encode(balance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}

		_, err = w.Write(body.Bytes())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}
	}
}

func getClientFromBody(r *http.Request) (client models.Client, err error) {
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&client)

	return
}

func (h *Handler) setJWToken(w http.ResponseWriter, clientID int) (err error) {
	claims := map[string]interface{}{"client_id": strconv.Itoa(clientID)}
	jwtauth.SetExpiry(claims, time.Now().Add(time.Hour))

	_, tokenString, err := h.tokenAuth.Encode(claims)
	if err == nil {
		http.SetCookie(w, &http.Cookie{Name: "jwt", Value: tokenString, Path: "/"})
	}

	return
}

func getClientID(r *http.Request) (id int, err error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return 0, err
	}

	clientID, ok := claims["client_id"]
	if !ok {
		return 0, ErrNotFindClientID
	}

	id, err = strconv.Atoi(clientID.(string))
	if err != nil {
		return 0, ErrNotFindClientID
	}

	return
}
