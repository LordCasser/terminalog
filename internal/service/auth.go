// Package service provides business logic services for the application.
package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"terminalog/internal/config"
	"terminalog/internal/model"
)

// AuthService provides authentication operations.
type AuthService struct {
	users []config.UserConfig
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		users: cfg.GetUsers(),
	}
}

// Validate checks if the username and password are valid.
func (s *AuthService) Validate(username, password string) (bool, error) {
	for _, user := range s.users {
		if user.Username != username {
			continue
		}

		// Check if password is bcrypt hash or plain text
		if isBcryptHash(user.Password) {
			// Verify bcrypt hash
			return s.VerifyPassword(user.Password, password), nil
		} else {
			// Compare plain text (for initial setup)
			return user.Password == password, nil
		}
	}

	return false, nil
}

// GetUsers returns the list of users.
func (s *AuthService) GetUsers() []model.User {
	users := make([]model.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u.ToModelUser())
	}
	return users
}

// GenerateDefaultUser generates a default admin user with a random password.
func (s *AuthService) GenerateDefaultUser() (*config.UserConfig, string, error) {
	// Generate random password
	password := generateRandomPassword(16)

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	user := &config.UserConfig{
		Username: "admin",
		Password: hashedPassword,
	}

	return user, password, nil
}

// HashPassword hashes a password using bcrypt.
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a bcrypt hash.
func (s *AuthService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// UpdateUsers updates the user list.
func (s *AuthService) UpdateUsers(users []config.UserConfig) {
	s.users = users
}

// Helper functions

// isBcryptHash checks if a string is a bcrypt hash.
func isBcryptHash(s string) bool {
	return len(s) >= 7 && (s[:4] == "$2a$" || s[:4] == "$2b$" || s[:4] == "$2y$")
}

// generateRandomPassword generates a random password of the given length.
func generateRandomPassword(length int) string {
	bytes := make([]byte, length/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}
