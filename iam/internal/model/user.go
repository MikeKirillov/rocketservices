package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
)

type User struct {
	UUID         uuid.UUID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatetAt    *time.Time
}

func NewUser(login, password string) (User, error) {
	if login == "" {
		return User{}, errs.ErrInvalidLogin
	}
	if len(password) < 8 {
		return User{}, errs.ErrWeakPassword
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("генерация hash пароля: %w", err)
	}
	return User{
		UUID:         uuid.New(),
		Login:        login,
		PasswordHash: string(passHash),
		CreatedAt:    time.Now(),
	}, nil
}
