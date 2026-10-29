package services

import (
	"context"
	"fmt"
	"strings"

	"ksp/internal/dto"
	"ksp/internal/repository"
	"ksp/internal/repository/models"
	"ksp/internal/validation"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetByMD5(ctx context.Context, md5 string) (*dto.ThreatResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	threat, err := s.repo.GetByMD5(ctx, md5)
	if err != nil {
		return nil, err
	}
	return toResponse(threat), nil
}

func (s *Service) Create(ctx context.Context, req dto.ThreatCreateRequest) (*dto.ThreatResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	threat := &models.Threat{
		MD5:         strings.ToLower(req.MD5),
		Name:        req.Name,
		Severity:    req.Severity,
		Description: req.Description,
		Category:    req.Category,
	}

	if err := s.repo.Create(ctx, threat); err != nil {
		return nil, err
	}

	created, err := s.repo.GetByMD5(ctx, threat.MD5)
	if err != nil {
		return nil, fmt.Errorf("fetch created threat: %w", err)
	}

	return toResponse(created), nil
}

func (s *Service) Update(ctx context.Context, md5 string, req dto.ThreatUpdateRequest) (*dto.ThreatResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByMD5(ctx, md5)
	if err != nil {
		return nil, err
	}

	existing.Name = req.Name

	if req.Severity != nil {
		sev := *req.Severity
		if sev == "" {
			sev = validation.DefaultSeverity
		}
		existing.Severity = sev
	}

	if req.Description != nil {
		existing.Description = *req.Description
	}

	if req.Category != nil {
		cat := *req.Category
		if cat == "" {
			cat = validation.DefaultCategory
		}
		existing.Category = cat
	}

	if err := s.repo.Update(ctx, md5, existing); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetByMD5(ctx, md5)
	if err != nil {
		return nil, fmt.Errorf("fetch updated threat: %w", err)
	}

	return toResponse(updated), nil
}

func toResponse(threat *models.Threat) *dto.ThreatResponse {
	return &dto.ThreatResponse{
		MD5:         threat.MD5,
		Name:        threat.Name,
		Severity:    threat.Severity,
		Description: threat.Description,
		Category:    threat.Category,
		CreatedAt:   threat.CreatedAt,
		UpdatedAt:   threat.UpdatedAt,
	}
}
