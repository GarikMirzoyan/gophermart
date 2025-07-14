package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GarikMirzoyan/gophermart/internal/domain/user"
)

type UserPG struct {
	DB *sql.DB
}

func NewUserPG(db *sql.DB) *UserPG {
	return &UserPG{DB: db}
}

func (r *UserPG) CreateUser(ctx context.Context, u *user.User) (*user.User, error) {
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO users (login, password)
		VALUES ($1, $2)
		RETURNING id
	`, u.Login, u.Password).Scan(&u.ID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserPG) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, login, password FROM users WHERE login = $1
	`, login)

	var u user.User
	if err := row.Scan(&u.ID, &u.Login, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
