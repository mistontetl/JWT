package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"portal_autofacturacion/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const inactivityTTL = 3 * time.Minute

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,100}$`)

type Service struct {
	db        *gorm.DB
	jwtSecret []byte
	now       func() time.Time
}

type Claims struct {
	SessionID string `json:"sid"`
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewService(db *gorm.DB, secret string) *Service {
	return &Service{db: db, jwtSecret: []byte(secret), now: time.Now}
}

func (s *Service) Register(username, password string) error {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("username inválido")
	}
	if len(password) < 8 {
		return fmt.Errorf("password debe tener al menos 8 caracteres")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{Username: username, PasswordHash: string(hash)}
	if err := s.db.Create(&user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return fmt.Errorf("usuario ya existe")
		}
		return err
	}
	return nil
}

func (s *Service) Login(username, password string) (*models.AuthResponse, error) {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("credenciales inválidas")
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	now := s.now().UTC()
	var existing models.UserSession
	err := s.db.Where("user_id = ? AND revoked_at IS NULL", user.ID).First(&existing).Error
	if err == nil {
		if existing.LastActivity.Add(inactivityTTL).After(now) {
			return nil, fmt.Errorf("el usuario ya tiene una sesión activa")
		}
		revokedAt := now
		s.db.Model(&existing).Update("revoked_at", &revokedAt)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	sessionID := uuid.NewString()
	expiresAt := now.Add(inactivityTTL)
	claims := Claims{SessionID: sessionID, Subject: user.Username, IssuedAt: now.Unix(), ExpiresAt: expiresAt.Unix()}

	tokenString, err := s.signToken(claims)
	if err != nil {
		return nil, err
	}

	session := models.UserSession{
		UserID:       user.ID,
		SessionID:    sessionID,
		TokenHash:    hashToken(tokenString),
		LastActivity: now,
		ExpiresAt:    expiresAt,
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: tokenString, ExpiresAt: expiresAt, SessionTTL: inactivityTTL.String()}, nil
}

func (s *Service) Logout(tokenString string) error {
	sessionID, err := s.sessionIDFromToken(tokenString)
	if err != nil {
		return err
	}
	now := s.now().UTC()
	return s.db.Model(&models.UserSession{}).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{"revoked_at": &now, "expires_at": now}).Error
}

func (s *Service) ValidateAndTouch(tokenString string) (*Claims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	var session models.UserSession
	if err := s.db.Where("session_id = ? AND revoked_at IS NULL", claims.SessionID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("sesión inválida")
	}
	if session.TokenHash != hashToken(tokenString) {
		return nil, fmt.Errorf("sesión inválida")
	}
	if session.LastActivity.Add(inactivityTTL).Before(now) || session.ExpiresAt.Before(now) || claims.ExpiresAt < now.Unix() {
		revokedAt := now
		s.db.Model(&session).Update("revoked_at", &revokedAt)
		return nil, fmt.Errorf("sesión expirada por inactividad")
	}

	newExpiry := now.Add(inactivityTTL)
	if err := s.db.Model(&session).Updates(map[string]any{"last_activity": now, "expires_at": newExpiry}).Error; err != nil {
		return nil, err
	}
	return claims, nil
}

func (s *Service) signToken(claims Claims) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEnc := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadJSON)
	input := headerEnc + "." + payloadEnc

	mac := hmac.New(sha256.New, s.jwtSecret)
	if _, err := mac.Write([]byte(input)); err != nil {
		return "", err
	}
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return input + "." + sig, nil
}

func (s *Service) parseToken(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token inválido")
	}
	input := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.jwtSecret)
	if _, err := mac.Write([]byte(input)); err != nil {
		return nil, fmt.Errorf("token inválido")
	}
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, fmt.Errorf("token inválido")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("token inválido")
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("token inválido")
	}
	if strings.TrimSpace(claims.SessionID) == "" || strings.TrimSpace(claims.Subject) == "" {
		return nil, fmt.Errorf("token inválido")
	}
	return &claims, nil
}

func (s *Service) sessionIDFromToken(tokenString string) (string, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.SessionID, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
