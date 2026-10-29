package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	data   map[string]*models.Threat
	getErr error
}

func (m *memRepo) GetByMD5(ctx context.Context, md5 string) (*models.Threat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.getErr != nil {
		return nil, m.getErr
	}
	threat, ok := m.data[md5]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *threat
	return &cp, nil
}

func (m *memRepo) Create(ctx context.Context, threat *models.Threat) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := m.data[threat.MD5]; ok {
		return repository.ErrConflict
	}
	cp := *threat
	m.data[threat.MD5] = &cp
	return nil
}

func (m *memRepo) Update(ctx context.Context, md5 string, threat *models.Threat) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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

func setupRouterWithRepo(t *testing.T, repo repository.Repository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

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

func setupRouter(t *testing.T) (*gin.Engine, *memRepo) {
	t.Helper()
	repo := &memRepo{data: map[string]*models.Threat{}}
	return setupRouterWithRepo(t, repo), repo
}

func TestAPI_CRUD(t *testing.T) {
	r, _ := setupRouter(t)
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

	var created dto.ThreatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.Severity != "high" || created.Category != "trojan" {
		t.Fatalf("unexpected create response: %+v", created)
	}

	req = httptest.NewRequest(http.MethodGet, "/info/"+md5, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", w.Code, w.Body.String())
	}

	updateBody, _ := json.Marshal(map[string]any{"name": "Trojan.Win32.Updated"})
	req = httptest.NewRequest(http.MethodPut, "/info/"+md5, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", w.Code, w.Body.String())
	}

	var updated dto.ThreatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Name != "Trojan.Win32.Updated" || updated.Severity != "high" || updated.Category != "trojan" {
		t.Fatalf("put merge failed: %+v", updated)
	}
}

func TestAPI_InvalidJSON(t *testing.T) {
	r, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_EmptyName(t *testing.T) {
	r, _ := setupRouter(t)

	body := `{"md5":"d41d8cd98f00b204e9800998ecf8427e","name":"   "}`
	req := httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_Defaults(t *testing.T) {
	r, _ := setupRouter(t)

	body := `{"md5":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","name":"Plain"}`
	req := httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var got dto.ThreatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Severity != "medium" || got.Category != "other" {
		t.Fatalf("defaults not applied: %+v", got)
	}
}

func TestAPI_NotFound(t *testing.T) {
	r, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/info/d41d8cd98f00b204e9800998ecf8427e", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestAPI_Conflict(t *testing.T) {
	r, _ := setupRouter(t)
	md5 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	body := `{"md5":"` + md5 + `","name":"one"}`

	req := httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first create status=%d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_InternalError(t *testing.T) {
	repo := &memRepo{
		data:   map[string]*models.Threat{},
		getErr: errors.New("db exploded"),
	}
	r := setupRouterWithRepo(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/info/d41d8cd98f00b204e9800998ecf8427e", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_MethodNotAllowed(t *testing.T) {
	r, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/info/d41d8cd98f00b204e9800998ecf8427e", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
