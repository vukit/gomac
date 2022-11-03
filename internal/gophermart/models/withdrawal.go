package models

import (
	"strconv"

	"github.com/vukit/gomac/internal/gophermart/utils"
)

type Withdrawal struct {
	ID          int     `json:"-"`
	ClientID    int     `json:"-"`
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (r *Withdrawal) Validate() error {
	number, err := strconv.Atoi(r.Order)
	if err != nil || !utils.IsValidLuhnNumber(number) {
		return ErrInvalidOrderNumberFormat
	}

	if r.Sum <= 0 {
		return ErrWrongWithdrawalSum
	}

	return nil
}
