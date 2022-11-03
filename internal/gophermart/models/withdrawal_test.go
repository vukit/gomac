package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vukit/gomac/internal/gophermart/models"
)

func TestWithdrawal(t *testing.T) {
	tests := []struct {
		name  string
		order string
		sum   float64
		want  error
	}{
		{
			name:  "case 1",
			order: "12345678903",
			sum:   120,
			want:  nil,
		},
		{
			name:  "case 2",
			order: "12345678904",
			sum:   100,
			want:  models.ErrInvalidOrderNumberFormat,
		},
		{
			name:  "case 3",
			order: "12345678903",
			sum:   -10,
			want:  models.ErrWrongWithdrawalSum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withdrawal := models.Withdrawal{Order: tt.order, Sum: tt.sum}
			assert.Equal(t, tt.want, withdrawal.Validate())
		})
	}

}
