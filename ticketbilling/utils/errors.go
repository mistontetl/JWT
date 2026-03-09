package utils

import (
	"errors"
	"strings"
)

func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	temporaryPatterns := []string{
		"timeout",
		"deadline exceeded",
		"connection refused",
		"connection reset",
		"no such host",          // Error de DNS temporal
		"service unavailable",   // Error 503
		"internal server error", // Error 500
		"bad gateway",           // Error 502
		"too many requests",     // Error 429 (Rate limiting)
		"network is unreachable",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}

	return false
}

// ERROR
var ErrServerError = errors.New("SERVER_ERROR")
var ErrTicketNotFound = errors.New("TICKET_NOT_FOUND")
