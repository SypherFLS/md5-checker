package validation

import (
	"regexp"
	"sync"

	"github.com/go-playground/validator/v10"
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
