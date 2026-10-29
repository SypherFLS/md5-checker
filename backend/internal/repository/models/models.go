package models

import "time"

type Threat struct {
	ID          int       `gorm:"primaryKey"`
	MD5         string    `gorm:"type:varchar(32);uniqueIndex;not null"`
	Name        string    `gorm:"type:text;not null"`
	Severity    string    `gorm:"column:severity;type:varchar(255)"`
	Description string    `gorm:"type:text"`
	Category    string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
