package entity

import "time"

type User struct {
	ID         int64     `gorm:"primaryKey"`
	TelegramID int64     `gorm:"uniqueIndex;not null"`
	SchoolName string    `gorm:"uniqueIndex;not null"`
	Role       string    `gorm:"not null;default:user"` // user, moderator, admin, superadmin
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}
