package integration_test

import (
	"bytes"
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
	pgrepo "ksp/internal/repository/postgres"
	"ksp/internal/services"

	"github.com/gin-gonic/gin"
)

func TestPostgresAPI_EndToEnd(t *testing.T) {
	db := openTestDB(t)
	repo := pgrepo.NewRepo(db)
	svc := services.NewService(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handlers.NewHandler(svc, logger)
	cfg := config.Config{Server: config.ServerConfig{
		Timeout:        5,
		RateLimitRPS:   1000,
		RateLimitBurst: 1000,
	}}

	gin.SetMode(gin.TestMode)
	r := router.NewRouter(h, cfg, logger)

	md5 := "test" + "dddddddddddddddddddddddddddd"
	_ = db.Exec("DELETE FROM threats WHERE md5 = ?", md5)

	createBody := `{"md5":"` + md5 + `","name":"  Postgres Threat  "}`
	req := httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", w.Code, w.Body.String())
	}

	var created dto.ThreatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.Name != "Postgres Threat" {
		t.Fatalf("name not trimmed: %q", created.Name)
	}
	if created.Severity != "medium" || created.Category != "other" {
		t.Fatalf("defaults: %+v", created)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("timestamps missing: %+v", created)
	}

	req = httptest.NewRequest(http.MethodGet, "/info/"+md5, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("conflict status=%d body=%s", w.Code, w.Body.String())
	}

	updateBody := `{"name":"Renamed","severity":"high"}`
	req = httptest.NewRequest(http.MethodPut, "/info/"+md5, bytes.NewBufferString(updateBody))
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
	if updated.Name != "Renamed" || updated.Severity != "high" || updated.Category != "other" {
		t.Fatalf("unexpected update: %+v", updated)
	}
}
