package auth

import (
	"context"
	"errors"

	"github.com/GarikMirzoyan/gophermart/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrLoginTaken         = errors.New("login already taken")
)

type Service struct {
	repo user.Repository
}

func New(repo user.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, login, password string) (*user.User, error) {
	existing, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, ErrLoginTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.repo.CreateUser(ctx, &user.User{
		Login:    login,
		Password: string(hashed),
	})
}

func (s *Service) Authenticate(ctx context.Context, login, password string) (*user.User, error) {
	u, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u, nil
}
