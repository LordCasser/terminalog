package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"terminalog/internal/config"
	"terminalog/internal/service"
)

func TestAuthService_Validate(t *testing.T) {
	// Create test config with hashed passwords
	hashedPassword1, err := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.DefaultCost)
	require.NoError(t, err)

	hashedPassword2, err := bcrypt.GenerateFromPassword([]byte("password2"), bcrypt.DefaultCost)
	require.NoError(t, err)

	cfg := &config.Config{
		Auth: config.AuthConfig{
			Users: []config.UserConfig{
				{Username: "user1", Password: string(hashedPassword1)},
				{Username: "user2", Password: string(hashedPassword2)},
			},
		},
	}

	authSvc := service.NewAuthService(cfg)

	tests := []struct {
		name     string
		username string
		password string
		want     bool
	}{
		{
			name:     "valid user1",
			username: "user1",
			password: "password1",
			want:     true,
		},
		{
			name:     "valid user2",
			username: "user2",
			password: "password2",
			want:     true,
		},
		{
			name:     "invalid password",
			username: "user1",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "invalid username",
			username: "user3",
			password: "password1",
			want:     false,
		},
		{
			name:     "empty credentials",
			username: "",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := authSvc.Validate(tt.username, tt.password)
			require.NoError(t, err)
			assert.Equal(t, tt.want, valid)
		})
	}
}

func TestAuthService_GenerateDefaultUser(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Users: []config.UserConfig{},
		},
	}

	authSvc := service.NewAuthService(cfg)

	user, password, err := authSvc.GenerateDefaultUser()
	require.NoError(t, err)

	assert.Equal(t, "admin", user.Username)
	assert.NotEmpty(t, password)
	assert.GreaterOrEqual(t, len(password), 16) // Password should be reasonably long

	// Add the generated user to the service for validation
	authSvc.UpdateUsers([]config.UserConfig{*user})

	// Verify the password works with the hash in user.Password
	valid, err := authSvc.Validate(user.Username, password)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestAuthService_HashPassword(t *testing.T) {
	cfg := config.Default()
	authSvc := service.NewAuthService(cfg)

	hashed, err := authSvc.HashPassword("testpassword")
	require.NoError(t, err)

	assert.NotEqual(t, "testpassword", hashed)
	assert.True(t, authSvc.VerifyPassword(hashed, "testpassword"))
	assert.False(t, authSvc.VerifyPassword(hashed, "wrongpassword"))
}

func TestAuthService_PlainTextPassword(t *testing.T) {
	// Test plain text password (for initial setup)
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Users: []config.UserConfig{
				{Username: "admin", Password: "plainpassword"},
			},
		},
	}

	authSvc := service.NewAuthService(cfg)

	valid, err := authSvc.Validate("admin", "plainpassword")
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = authSvc.Validate("admin", "wrongpassword")
	require.NoError(t, err)
	assert.False(t, valid)
}
