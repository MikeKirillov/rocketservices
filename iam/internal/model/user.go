package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
)

type User struct {
	UUID      uuid.UUID
	Login     string
	Password  Password
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func NewUser(login, password string) (User, error) {
	if login == "" {
		return User{}, errs.ErrInvalidLogin
	}
	if len(password) < 8 {
		return User{}, errs.ErrWeakPassword
	}
	pass, err := NewPassword(password)
	if err != nil {
		return User{}, fmt.Errorf("генерация hash пароля: %w", err)
	}
	return User{
		UUID:      uuid.New(),
		Login:     login,
		Password:  pass,
		CreatedAt: time.Now(),
	}, nil
}

type Password struct {
	hash string
}

func NewPassword(password string) (Password, error) {
	if len(password) < 8 {
		return Password{}, errs.ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, err
	}
	return Password{hash: string(hash)}, nil
}

// RestorePassword - восстанавливает пароль из БД готовый хэш.
func RestorePassword(hash string) Password {
	return Password{hash: hash}
}

func (p *Password) Hash() string {
	return p.hash
}

func (p *Password) Verify(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(p.Hash()), []byte(password))
}

// Маскировка пароля на уровне логов.

func (p *Password) String() string               { return "***" }
func (p *Password) MarshalJSON() ([]byte, error) { return []byte("***"), nil }
