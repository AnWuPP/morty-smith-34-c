package usecase

import (
	"context"
	"errors"
	"morty-smith-34-c/internal/app/entity"
	"morty-smith-34-c/internal/app/repository"
	"time"
)

type UserUseCase struct {
	UserRepo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		UserRepo: repo,
	}
}

// RegisterUser добавляет пользователя в базу данных
func (u *UserUseCase) RegisterUser(ctx context.Context, user *entity.User) error {
	return u.UserRepo.Create(ctx, user)
}

func (u *UserUseCase) CheckRole(ctx context.Context, telegramID int64, allowedRoles []string) error {
	user, err := u.UserRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return errors.New("user not found")
	}

	for _, role := range allowedRoles {
		if user.Role == role {
			return nil
		}
	}

	return errors.New("permission denied")
}

func (u *UserUseCase) UpdateRole(ctx context.Context, telegramID int64, role string) error {
	return u.UserRepo.UpdateRole(ctx, telegramID, role)
}

func (u *UserUseCase) SaveNickname(ctx context.Context, telegramID int64, nickname string) error {
	user := &entity.User{
		TelegramID: telegramID,
		SchoolName: nickname,
		Role:       "user",
		CreatedAt:  time.Now(),
	}

	return u.UserRepo.Create(ctx, user)
}

func (u *UserUseCase) Exists(ctx context.Context, telegramID int64) (bool, error) {
	return u.UserRepo.Exists(ctx, telegramID)
}
