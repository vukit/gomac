package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vukit/gomac/internal/gophermart/models"
)

func TestOrder(t *testing.T) {
	tests := []struct {
		name     string
		clientID int
		number   string
		want     error
	}{
		{
			name:     "case 1",
			clientID: 1,
			number:   "12345678903",
			want:     nil,
		},
		{
			name:     "case 2",
			clientID: 1,
			number:   "12345678904",
			want:     models.ErrInvalidOrderNumberFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := models.Order{ClientID: tt.clientID, Number: tt.number}
			assert.Equal(t, tt.want, order.Validate())
		})
	}

}
