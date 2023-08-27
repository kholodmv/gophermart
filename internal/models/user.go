package models

import (
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Login        string `json:"login"`
	Password     string `json:"password"`
	HashPassword string `json:"-"`
}

func (u *User) GenerateHashPassword() (*User, error) {
	if u.HashPassword == "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		u.HashPassword = hex.EncodeToString(hash)
	}
	return u, nil
}
