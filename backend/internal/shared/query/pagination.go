// Package query provides helpers for parsing HTTP query parameters.
package query

import (
	"net/url"
	"strconv"
)

// Pagination parses positive page and limit values and caps the limit.
func Pagination(values url.Values, defaultLimit, maxLimit int) (page, limit int) {
	page = PositiveInt(values.Get("page"), 1)
	limit = PositiveInt(values.Get("limit"), defaultLimit)
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

// PositiveInt parses a positive integer or returns fallback.
func PositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
