package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists        = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidToken      = errors.New("invalid or expired token")
)

type Service struct {
	db         *sqlx.DB
	jwtSecret  []byte
	jwtExpiry  time.Duration
}

type User struct {
	ID              string     `db:"id" json:"id"`
	Username        string     `db:"username" json:"username"`
	Email           string     `db:"email" json:"email"`
	PasswordHash    string     `db:"password_hash" json:"-"`
	ELO             int        `db:"elo" json:"elo"`
	IsPremium       bool       `db:"is_premium" json:"is_premium"`
	PremiumUntil    *time.Time `db:"premium_until" json:"premium_until,omitempty"`
	StripeCustomerID *string    `db:"stripe_customer_id" json:"stripe_customer_id,omitempty"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewService(db *sqlx.DB, jwtSecret []byte, jwtExpiry time.Duration) *Service {
	return &Service{
		db:         db,
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	// Check if user exists
	var exists bool
	err := s.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE email = $1 OR username = $2
		)
	`, input.Email, input.Username)
	if err != nil {
		return nil, fmt.Errorf("check user exists: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hash),
		ELO:         1200, // Starting ELO
	}

	query := `
		INSERT INTO users (username, email, password_hash, elo)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err = s.db.GetContext(ctx, user, query,
		user.Username, user.Email, user.PasswordHash, user.ELO,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	user := &User{}
	err := s.db.GetContext(ctx, user, `
		SELECT * FROM users WHERE email = $1
	`, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	return s.generateTokenPair(user)
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Parse and validate refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Get user
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	user := &User{}
	err = s.db.GetContext(ctx, user, `
		SELECT * FROM users WHERE id = $1
	`, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Generate new token pair
	return s.generateTokenPair(user)
}

func (s *Service) generateTokenPair(user *User) (*TokenPair, error) {
	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"is_premium": user.IsPremium,
		"exp":        time.Now().Add(s.jwtExpiry).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Generate refresh token (valid for 30 days)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	user := &User{}
	err = s.db.Get(user, `SELECT * FROM users WHERE id = $1`, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}
