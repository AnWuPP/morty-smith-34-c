package repository

import (
	"context"
	"errors"
	"fmt"
	"morty-smith-34-c/internal/app/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error)
	UpdateRole(ctx context.Context, telegramID int64, role string) error
	Exists(ctx context.Context, telegramID int64) (bool, error)
	UpdateSchoolNick(ctx context.Context, telegramID int64, nick string) (bool, error)
}

type PostgresUserRepository struct {
	DB *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{DB: db}
}

// Create добавляет нового пользователя в таблицу users
func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *PostgresUserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error) {
	var user entity.User
	if err := r.DB.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) UpdateRole(ctx context.Context, telegramID int64, role string) error {
	return r.DB.WithContext(ctx).
		Model(&entity.User{}).
		Where("telegram_id = ?", telegramID).
		Update("role", role).Error
}

func (r *PostgresUserRepository) Exists(ctx context.Context, telegramID int64) (bool, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&entity.User{}).
		Where("telegram_id = ?", telegramID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresUserRepository) UpdateSchoolNick(ctx context.Context, telegramID int64, nick string) (bool, error) {
	var user struct {
		SchoolNick string
	}

	if err := r.DB.WithContext(ctx).Model(&entity.User{}).Select("school_nick").Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("user with telegram_id %d not found", telegramID)
		}
		return false, err
	}

	if user.SchoolNick == nick {
		return false, nil
	}

	result := r.DB.WithContext(ctx).Model(&entity.User{}).Where("telegram_id = ?", telegramID).Update("school_nick", nick)
	if result.Error != nil {
		return false, result.Error
	}

	if result.RowsAffected == 0 {
		return false, fmt.Errorf("failed to update school_nick for telegram_id %d", telegramID)
	}

	return true, nil
}
