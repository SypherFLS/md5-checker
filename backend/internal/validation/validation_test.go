package validation_test

import (
	"testing"

	"ksp/internal/dto"
	"ksp/internal/validation"
)

func TestIsValidMD5(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want bool
	}{
		{"d41d8cd98f00b204e9800998ecf8427e", true},
		{"D41D8CD98F00B204E9800998ECF8427E", true},
		{"d41d8cd98f00b204e9800998ecf8427", false},
		{"d41d8cd98f00b204e9800998ecf8427eg", false},
		{"", false},
		{"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", false},
	}

	for _, tt := range tests {
		if got := validation.IsValidMD5(tt.in); got != tt.want {
			t.Fatalf("IsValidMD5(%q)=%v want %v", tt.in, got, tt.want)
		}
	}
}

func TestValidateCreateRequest(t *testing.T) {
	t.Parallel()

	ok := dto.ThreatCreateRequest{
		MD5:      "d41d8cd98f00b204e9800998ecf8427e",
		Name:     "Trojan.Win32.Generic",
		Severity: "high",
	}
	if err := validation.Validate(ok); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	bad := dto.ThreatCreateRequest{
		MD5:      "not-an-md5",
		Name:     "x",
		Severity: "ultra",
	}
	if err := validation.Validate(bad); err == nil {
		t.Fatal("expected validation error")
	}
}
