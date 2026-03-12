package models

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:100;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserSession struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"user_id" gorm:"index;not null"`
	SessionID    string     `json:"session_id" gorm:"size:64;uniqueIndex;not null"`
	TokenHash    string     `json:"-" gorm:"size:64;not null"`
	LastActivity time.Time  `json:"last_activity" gorm:"index;not null"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"index;not null"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
	SessionTTL string    `json:"session_ttl"`
}
