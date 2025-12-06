package utils

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func MsgForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	}
	return ""
}

func ValidateRequestBody(err error, ve validator.ValidationErrors) map[string]string {
	
	out := make(map[string]string)
	for _, fe := range ve {
		out[strings.ToLower(fe.Field())] = MsgForTag(fe.Tag())
	}

	return out

}
