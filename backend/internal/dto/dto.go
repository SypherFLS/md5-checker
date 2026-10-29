package dto

import "time"

type ThreatCreateRequest struct {
	MD5         string `json:"md5" validate:"required,md5hash"`
	Name        string `json:"name" validate:"required,min=1"`
	Severity    string `json:"severity" validate:"omitempty,oneof=low medium high critical"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type ThreatUpdateRequest struct {
	Name        string  `json:"name" validate:"required,min=1"`
	Severity    *string `json:"severity" validate:"omitempty,oneof=low medium high critical"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
}

type ThreatResponse struct {
	MD5         string    `json:"md5"`
	Name        string    `json:"name"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
