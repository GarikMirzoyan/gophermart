package user

import "context"

type Repository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
}
