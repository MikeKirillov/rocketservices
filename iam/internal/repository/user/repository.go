package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/repository/converter"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		pool: pool,
	}
}

func (r *repository) InsertUser(ctx context.Context, user model.User) error {
	rec := converter.UserToRecord(user)

	query := `INSERT INTO users(uuid, login, password_hash, created_at) VALUES ($1, $2, $3, $4)`

	tag, err := r.pool.Exec(ctx, query, rec.Uuid, rec.Login, rec.PasswordHash, rec.CreatedAt)
	if err != nil {
		// При Register два параллельных запроса с одинаковым логином приведут к unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.ErrUserAlreadyExists
		}
		return fmt.Errorf("создание записи пользователя %s: %w", rec.Login, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("не удалось создать пользователя %s", rec.Login)
	}

	return nil
}

func (r *repository) GetUser() error {
	return nil
}
