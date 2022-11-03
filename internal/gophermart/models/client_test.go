package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vukit/gomac/internal/gophermart/models"
)

func TestClient(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		want     error
	}{
		{
			name:     "case 1",
			login:    "mark",
			password: "secret",
			want:     nil,
		},
		{
			name:     "case 2",
			login:    "mark",
			password: "",
			want:     models.ErrLoginPasswordEmpity,
		},
		{
			name:     "case 3",
			login:    "",
			password: "secret",
			want:     models.ErrLoginPasswordEmpity,
		},
		{
			name:     "case 4",
			login:    "",
			password: "",
			want:     models.ErrLoginPasswordEmpity,
		},
		{
			name:     "case 5",
			login:    "longusernamelongusernamelongusernamelongusernamelongusernamelongusername",
			password: "secret",
			want:     models.ErrLongLogin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := models.Client{Login: tt.login, Password: tt.password}
			assert.Equal(t, tt.want, client.Validate())
		})
	}

}
