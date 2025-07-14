package auth_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/GarikMirzoyan/gophermart/internal/domain/user"
	mock_user "github.com/GarikMirzoyan/gophermart/internal/domain/user/mocks"
	auth "github.com/GarikMirzoyan/gophermart/internal/usecase/auth"
)

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mock_user.NewMockRepository(ctrl)
	service := auth.New(mockRepo)

	t.Run("login already taken", func(t *testing.T) {
		mockRepo.EXPECT().GetByLogin(ctx, "user1").Return(&user.User{ID: 1, Login: "user1"}, nil)

		_, err := service.Register(ctx, "user1", "pass")
		if !errors.Is(err, auth.ErrLoginTaken) {
			t.Errorf("expected ErrLoginTaken, got %v", err)
		}
	})

	t.Run("successful registration", func(t *testing.T) {
		mockRepo.EXPECT().GetByLogin(ctx, "user2").Return(nil, nil)
		mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, u *user.User) (*user.User, error) {
				if u.Login != "user2" || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("pass123")) != nil {
					t.Errorf("password hash mismatch or login incorrect")
				}
				return u, nil
			})

		_, err := service.Register(ctx, "user2", "pass123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestService_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mock_user.NewMockRepository(ctrl)
	service := auth.New(mockRepo)

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByLogin(ctx, "nouser").Return(nil, nil)

		_, err := service.Authenticate(ctx, "nouser", "pass")
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		mockRepo.EXPECT().GetByLogin(ctx, "user").Return(&user.User{Login: "user", Password: string(hashed)}, nil)

		_, err := service.Authenticate(ctx, "user", "wrong")
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("successful auth", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		expectedUser := &user.User{ID: 1, Login: "user", Password: string(hashed)}

		mockRepo.EXPECT().GetByLogin(ctx, "user").Return(expectedUser, nil)

		u, err := service.Authenticate(ctx, "user", "correct")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if u.ID != expectedUser.ID || u.Login != expectedUser.Login {
			t.Errorf("unexpected user returned: %+v", u)
		}
	})
}
