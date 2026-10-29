package repository

import (
	"context"
	"errors"

	"ksp/internal/repository/models"
)

var (
	ErrNotFound = errors.New("threat not found")
	ErrConflict = errors.New("threat already exists")
)

type Repository interface {
	GetByMD5(ctx context.Context, md5 string) (*models.Threat, error)
	Create(ctx context.Context, threat *models.Threat) error
	Update(ctx context.Context, md5 string, threat *models.Threat) error
}
