package utils

import "time"

//Tyes

func DateTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

const (
	PENDING     = 1
	VALIDATING  = 2
	STAMPING    = 3
	SUCCESS     = 4
	MAIL_FAILED = 5
	ERROR       = 6
)
