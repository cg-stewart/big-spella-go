package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	
	// Server
	Port            int           `mapstructure:"PORT"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`
	
	// Database
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	
	// Redis
	RedisURL string `mapstructure:"REDIS_URL"`
	
	// JWT
	JWTSecret     string        `mapstructure:"JWT_SECRET"`
	JWTExpiration time.Duration `mapstructure:"JWT_EXPIRATION"`
	
	// AWS
	AWSRegion          string `mapstructure:"AWS_REGION"`
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	
	// Chime
	ChimeAppARN string `mapstructure:"CHIME_APP_ARN"`
	
	// GetStream
	GetStreamAPIKey    string `mapstructure:"GETSTREAM_API_KEY"`
	GetStreamAPISecret string `mapstructure:"GETSTREAM_API_SECRET"`
	
	// OpenAI
	OpenAIAPIKey string `mapstructure:"OPENAI_API_KEY"`
	
	// Stripe
	StripeSecretKey      string `mapstructure:"STRIPE_SECRET_KEY"`
	StripeWebhookSecret  string `mapstructure:"STRIPE_WEBHOOK_SECRET"`
	StripePremiumPriceID string `mapstructure:"STRIPE_PREMIUM_PRICE_ID"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	
	// Environment variables take precedence
	viper.AutomaticEnv()
	
	// Set defaults
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("SHUTDOWN_TIMEOUT", time.Second*30)
	viper.SetDefault("JWT_EXPIRATION", time.Hour*24*7)
	
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK if we're using env vars
	}
	
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	// Validate required fields
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if config.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}
	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	
	return config, nil
}
