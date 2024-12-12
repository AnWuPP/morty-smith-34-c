package entity

import "time"

type Chat struct {
	ID         int64     `gorm:"primaryKey"`
	ChatID     int64     `gorm:"uniqueIndex;not null"`
	CampusName string    `gorm:"not null"`
	ThreadID   int       `gorm:"default:-1"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}
