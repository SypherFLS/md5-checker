package httperr_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"ksp/internal/api/httperr"
	"ksp/internal/repository"
)

func TestFromService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		md5    string
		status int
		msg    string
	}{
		{
			name:   "not found",
			err:    repository.ErrNotFound,
			md5:    "D41D8CD98F00B204E9800998ECF8427E",
			status: http.StatusNotFound,
			msg:    "Threat not found",
		},
		{
			name:   "conflict",
			err:    repository.ErrConflict,
			md5:    "d41d8cd98f00b204e9800998ecf8427e",
			status: http.StatusConflict,
			msg:    "Threat already exists",
		},
		{
			name:   "timeout",
			err:    context.DeadlineExceeded,
			md5:    "d41d8cd98f00b204e9800998ecf8427e",
			status: http.StatusGatewayTimeout,
			msg:    "Request timeout",
		},
		{
			name:   "wrapped not found",
			err:    errors.Join(errors.New("wrap"), repository.ErrNotFound),
			md5:    "d41d8cd98f00b204e9800998ecf8427e",
			status: http.StatusNotFound,
			msg:    "Threat not found",
		},
		{
			name:   "internal",
			err:    errors.New("db down"),
			md5:    "d41d8cd98f00b204e9800998ecf8427e",
			status: http.StatusInternalServerError,
			msg:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, body := httperr.FromService(tt.err, tt.md5)
			if status != tt.status {
				t.Fatalf("status: got %d want %d", status, tt.status)
			}
			if body.Message != tt.msg {
				t.Fatalf("message: got %q want %q", body.Message, tt.msg)
			}
			if body.Details == "" {
				t.Fatal("details must not be empty")
			}
		})
	}
}

func TestBadRequest(t *testing.T) {
	t.Parallel()

	status, body := httperr.BadRequest("Invalid MD5", "bad hash")
	if status != http.StatusBadRequest {
		t.Fatalf("status: got %d", status)
	}
	if body.Message != "Invalid MD5" || body.Details != "bad hash" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	t.Parallel()

	status, body := httperr.MethodNotAllowed()
	if status != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d", status)
	}
	if body.Message == "" {
		t.Fatal("message must not be empty")
	}
}
