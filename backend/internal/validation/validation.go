package validation

import (
	"regexp"
	"strings"
	"sync"

	"ksp/internal/dto"

	"github.com/go-playground/validator/v10"
)

const (
	DefaultSeverity = "medium"
	DefaultCategory = "other"
)

var (
	validate *validator.Validate
	once     sync.Once
	md5Re    = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
)

func getValidate() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
		_ = validate.RegisterValidation("md5hash", func(fl validator.FieldLevel) bool {
			return md5Re.MatchString(fl.Field().String())
		})
	})
	return validate
}

func Validate(v any) error {
	return getValidate().Struct(v)
}

func IsValidMD5(md5 string) bool {
	return md5Re.MatchString(md5)
}

func SanitizeCreate(req *dto.ThreatCreateRequest) {
	req.MD5 = strings.TrimSpace(req.MD5)
	req.Name = strings.TrimSpace(req.Name)
	req.Severity = strings.TrimSpace(req.Severity)
	req.Description = strings.TrimSpace(req.Description)
	req.Category = strings.TrimSpace(req.Category)
}

func SanitizeUpdate(req *dto.ThreatUpdateRequest) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Severity != nil {
		v := strings.TrimSpace(*req.Severity)
		if v == "" {
			v = DefaultSeverity
		}
		req.Severity = &v
	}
	if req.Description != nil {
		v := strings.TrimSpace(*req.Description)
		req.Description = &v
	}
	if req.Category != nil {
		v := strings.TrimSpace(*req.Category)
		if v == "" {
			v = DefaultCategory
		}
		req.Category = &v
	}
}

func ApplyCreateDefaults(req *dto.ThreatCreateRequest) {
	if req.Severity == "" {
		req.Severity = DefaultSeverity
	}
	if req.Category == "" {
		req.Category = DefaultCategory
	}
}
