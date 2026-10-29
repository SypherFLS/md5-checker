package services_test

import (
	"context"
	"errors"
	"testing"

	"ksp/internal/dto"
	"ksp/internal/repository"
	"ksp/internal/repository/models"
	"ksp/internal/services"
)

type mockRepo struct {
	byMD5   map[string]*models.Threat
	create  func(ctx context.Context, threat *models.Threat) error
	update  func(ctx context.Context, md5 string, threat *models.Threat) error
	getErr  error
	created []*models.Threat
}

func (m *mockRepo) GetByMD5(_ context.Context, md5 string) (*models.Threat, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	threat, ok := m.byMD5[md5]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *threat
	return &cp, nil
}

func (m *mockRepo) Create(ctx context.Context, threat *models.Threat) error {
	if m.create != nil {
		return m.create(ctx, threat)
	}
	if _, ok := m.byMD5[threat.MD5]; ok {
		return repository.ErrConflict
	}
	cp := *threat
	cp.ID = len(m.byMD5) + 1
	m.byMD5[threat.MD5] = &cp
	m.created = append(m.created, &cp)
	return nil
}

func (m *mockRepo) Update(ctx context.Context, md5 string, threat *models.Threat) error {
	if m.update != nil {
		return m.update(ctx, md5, threat)
	}
	existing, ok := m.byMD5[md5]
	if !ok {
		return repository.ErrNotFound
	}
	existing.Name = threat.Name
	existing.Severity = threat.Severity
	existing.Description = threat.Description
	existing.Category = threat.Category
	return nil
}

func TestServiceCreateAndGet(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{byMD5: map[string]*models.Threat{}}
	svc := services.NewService(repo)

	created, err := svc.Create(context.Background(), dto.ThreatCreateRequest{
		MD5:         "D41D8CD98F00B204E9800998ECF8427E",
		Name:        "Trojan.Win32.Generic",
		Severity:    "high",
		Description: "demo",
		Category:    "trojan",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.MD5 != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Fatalf("md5 not normalized: %s", created.MD5)
	}

	got, err := svc.GetByMD5(context.Background(), created.MD5)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Trojan.Win32.Generic" || got.Severity != "high" {
		t.Fatalf("unexpected threat: %+v", got)
	}
}

func TestServiceCreateConflict(t *testing.T) {
	t.Parallel()

	md5 := "d41d8cd98f00b204e9800998ecf8427e"
	repo := &mockRepo{byMD5: map[string]*models.Threat{
		md5: {MD5: md5, Name: "exists"},
	}}
	svc := services.NewService(repo)

	_, err := svc.Create(context.Background(), dto.ThreatCreateRequest{
		MD5:  md5,
		Name: "other",
	})
	if !errors.Is(err, repository.ErrConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
}

func TestServiceUpdateNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{byMD5: map[string]*models.Threat{}}
	svc := services.NewService(repo)

	_, err := svc.Update(context.Background(), "d41d8cd98f00b204e9800998ecf8427e", dto.ThreatUpdateRequest{
		Name: "x",
	})
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()

	md5 := "d41d8cd98f00b204e9800998ecf8427e"
	repo := &mockRepo{byMD5: map[string]*models.Threat{
		md5: {MD5: md5, Name: "old", Severity: "low"},
	}}
	svc := services.NewService(repo)

	got, err := svc.Update(context.Background(), md5, dto.ThreatUpdateRequest{
		Name:     "new",
		Severity: "critical",
		Category: "malware",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Name != "new" || got.Severity != "critical" || got.Category != "malware" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
