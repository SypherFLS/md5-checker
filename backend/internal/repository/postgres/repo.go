package postgres

import (
	"context"
	"errors"
	"strings"

	"ksp/internal/repository"
	"ksp/internal/repository/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func (r *Repo) GetByMD5(ctx context.Context, md5 string) (*models.Threat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var threat models.Threat
	err := r.db.WithContext(ctx).Where("md5 = ?", strings.ToLower(md5)).First(&threat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &threat, nil
}

func (r *Repo) Create(ctx context.Context, threat *models.Threat) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	threat.MD5 = strings.ToLower(threat.MD5)
	err := r.db.WithContext(ctx).Create(threat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return repository.ErrConflict
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrConflict
		}
		return err
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, md5 string, threat *models.Threat) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	result := r.db.WithContext(ctx).
		Model(&models.Threat{}).
		Where("md5 = ?", strings.ToLower(md5)).
		Updates(map[string]any{
			"name":        threat.Name,
			"severity":    threat.Severity,
			"description": threat.Description,
			"category":    threat.Category,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
