package repository

import (
	"context"
	"morty-smith-34-c/internal/app/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error)
	Exists(ctx context.Context, telegramID int64) (bool, error)
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
