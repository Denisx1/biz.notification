package domain

import "time"

type SMSVerification struct {
	ID        int64
	CodeHash  string
	ExpiresAt time.Time
}
