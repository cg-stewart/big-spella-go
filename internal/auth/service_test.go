package auth

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgres://postgres:postgres@localhost:5432/bigspella_test?sslmode=disable")
	require.NoError(t, err)

	// Clear users table
	_, err = db.Exec("TRUNCATE users CASCADE")
	require.NoError(t, err)

	return db
}

func TestRegister(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db, []byte("test-secret"), time.Hour)

	t.Run("successful registration", func(t *testing.T) {
		input := RegisterInput{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		user, err := service.Register(context.Background(), input)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, input.Username, user.Username)
		assert.Equal(t, input.Email, user.Email)
		assert.Equal(t, 1200, user.ELO)
	})

	t.Run("duplicate user", func(t *testing.T) {
		input := RegisterInput{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		_, err := service.Register(context.Background(), input)
		assert.ErrorIs(t, err, ErrUserExists)
	})
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db, []byte("test-secret"), time.Hour)

	// Register a user first
	_, err := service.Register(context.Background(), RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	t.Run("successful login", func(t *testing.T) {
		input := LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}

		tokens, err := service.Login(context.Background(), input)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)

		// Verify token
		user, err := service.ValidateToken(tokens.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		input := LoginInput{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		_, err := service.Login(context.Background(), input)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})
}

func TestRefreshToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db, []byte("test-secret"), time.Hour)

	// Register and login a user first
	_, err := service.Register(context.Background(), RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	tokens, err := service.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	t.Run("successful refresh", func(t *testing.T) {
		// Wait a moment to ensure tokens will be different
		time.Sleep(time.Second)
		
		newTokens, err := service.RefreshToken(context.Background(), tokens.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)

		// Verify both tokens are valid but different
		user1, err := service.ValidateToken(tokens.AccessToken)
		require.NoError(t, err)
		user2, err := service.ValidateToken(newTokens.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, user1.ID, user2.ID)
		assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
		assert.NotEqual(t, tokens.RefreshToken, newTokens.RefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		_, err := service.RefreshToken(context.Background(), "invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
