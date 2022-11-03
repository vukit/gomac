package models

import (
	"strings"
)

const maxLoginLegth = 64

type Client struct {
	ID       int `json:"-"`
	Login    string
	Password string
}

func (r *Client) Validate() error {
	if strings.TrimSpace(r.Login) == "" || strings.TrimSpace(r.Password) == "" {
		return ErrLoginPasswordEmpity
	}

	if len(r.Login) > maxLoginLegth {
		return ErrLongLogin
	}

	return nil
}
