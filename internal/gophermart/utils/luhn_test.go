package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vukit/gomac/internal/gophermart/utils"
)

func TestLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number int
		want   bool
	}{
		{
			name:   "valid number",
			number: 12345678903,
			want:   true,
		},
		{
			name:   "wrong number",
			number: 12345678904,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, utils.IsValidLuhnNumber(tt.number))
		})
	}
}
