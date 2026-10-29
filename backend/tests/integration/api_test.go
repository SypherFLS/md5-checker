package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ksp/internal/api/handlers"
	"ksp/internal/api/router"
	"ksp/internal/config"
	"ksp/internal/dto"
	"ksp/internal/repository"
	"ksp/internal/repository/models"
	"ksp/internal/services"

	"github.com/gin-gonic/gin"
)

type memRepo struct {
	data map[string]*models.Threat
}

func (m *memRepo) GetByMD5(_ context.Context, md5 string) (*models.Threat, error) {
	threat, ok := m.data[md5]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *threat
	return &cp, nil
}

func (m *memRepo) Create(_ context.Context, threat *models.Threat) error {
	if _, ok := m.data[threat.MD5]; ok {
		return repository.ErrConflict
	}
	cp := *threat
	m.data[threat.MD5] = &cp
	return nil
}

func (m *memRepo) Update(_ context.Context, md5 string, threat *models.Threat) error {
	existing, ok := m.data[md5]
	if !ok {
		return repository.ErrNotFound
	}
	existing.Name = threat.Name
	existing.Severity = threat.Severity
	existing.Description = threat.Description
	existing.Category = threat.Category
	return nil
}

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := &memRepo{data: map[string]*models.Threat{}}
	svc := services.NewService(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handlers.NewHandler(svc, logger)
	cfg := config.Config{Server: config.ServerConfig{
		Timeout:        5,
		RateLimitRPS:   1000,
		RateLimitBurst: 1000,
	}}
	return router.NewRouter(h, cfg, logger)
}

func TestAPI_CRUD(t *testing.T) {
	r := setupRouter(t)
	md5 := "d41d8cd98f00b204e9800998ecf8427e"

	createBody, _ := json.Marshal(dto.ThreatCreateRequest{
		MD5:      md5,
		Name:     "Trojan.Win32.Generic",
		Severity: "high",
		Category: "trojan",
	})
	req := httptest.NewRequest(http.MethodPost, "/info", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Trace-Id") == "" {
		t.Fatal("expected X-Trace-Id header")
	}

	req = httptest.NewRequest(http.MethodGet, "/info/"+md5, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", w.Code, w.Body.String())
	}

	var got dto.ThreatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.Name != "Trojan.Win32.Generic" {
		t.Fatalf("unexpected name: %s", got.Name)
	}

	updateBody, _ := json.Marshal(dto.ThreatUpdateRequest{
		Name:     "Trojan.Win32.Updated",
		Severity: "critical",
	})
	req = httptest.NewRequest(http.MethodPut, "/info/"+md5, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/info", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_ValidationAndNotFound(t *testing.T) {
	r := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/info/not-md5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad md5 status=%d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/info/d41d8cd98f00b204e9800998ecf8427e", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d", w.Code)
	}
}
