package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"cyto-viewer/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	config      *config.AuthConfig
	activeSessions map[string]*Session
	mu          sync.RWMutex
}

type Session struct {
	Username  string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewManager(cfg *config.AuthConfig) *Manager {
	return &Manager{
		config:         cfg,
		activeSessions: make(map[string]*Session),
	}
}

func (m *Manager) Authenticate(username, password string) (string, error) {
	// In production, validate against database
	// For now, using environment-based validation
	if !m.validateCredentials(username, password) {
		return "", fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := m.generateToken(username)
	if err != nil {
		return "", err
	}

	// Store session
	m.mu.Lock()
	m.activeSessions[token] = &Session{
		Username:  username,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.config.TokenExpiry),
	}
	m.mu.Unlock()

	return token, nil
}

func (m *Manager) ValidateToken(tokenString string) bool {
	m.mu.RLock()
	session, exists := m.activeSessions[tokenString]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		m.mu.Lock()
		delete(m.activeSessions, tokenString)
		m.mu.Unlock()
		return false
	}

	// Verify JWT signature
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		return false
	}

	return token.Valid
}

func (m *Manager) RevokeToken(tokenString string) {
	m.mu.Lock()
	delete(m.activeSessions, tokenString)
	m.mu.Unlock()
}

func (m *Manager) generateToken(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "cyto-viewer",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.JWTSecret))
}

func (m *Manager) validateCredentials(username, password string) bool {
	// In production, this would check against a secure database
	// For now, using environment variables
	expectedHash := m.config.PasswordHash
	if expectedHash == "" {
		// Development mode - accept any password
		return true
	}

	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(password))
	return err == nil
}

func (m *Manager) CleanupExpiredSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for token, session := range m.activeSessions {
		if now.After(session.ExpiresAt) {
			delete(m.activeSessions, token)
		}
	}
}

func (m *Manager) StartCleanupWorker() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			m.CleanupExpiredSessions()
		}
	}()
}

// HashPassword creates a bcrypt hash of a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// GenerateSecureToken creates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
