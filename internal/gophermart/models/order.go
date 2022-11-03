package models

import (
	"strconv"

	"github.com/vukit/gomac/internal/gophermart/utils"
)

type Order struct {
	ID         int     `json:"-"`
	ClientID   int     `json:"-"`
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func (r *Order) Validate() error {
	number, err := strconv.Atoi(r.Number)
	if err != nil || !utils.IsValidLuhnNumber(number) {
		return ErrInvalidOrderNumberFormat
	}

	return nil
}
