package server

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/ddddami/laivan/internal/validator"
)

func readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

func readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddFieldError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func readBool(qs url.Values, key string, defaultValue *bool, v *validator.Validator) *bool {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	switch strings.ToLower(s) {
	case "true", "1":
		result := true
		return &result
	case "false", "0":
		result := false
		return &result
	default:
		v.AddFieldError(key, "must be a boolean value")
		return defaultValue
	}
}
