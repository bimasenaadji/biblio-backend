package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all environment configurations for the Biblio application.
type Config struct {
	// Application Settings
	AppName string `mapstructure:"APP_NAME"`
	AppEnv  string `mapstructure:"APP_ENV"`
	AppPort string `mapstructure:"APP_PORT"`

	// PostgreSQL Database Configuration
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE"`
	DBTimezone string `mapstructure:"DB_TIMEZONE"`

	// MinIO S3 Object Storage Configuration (Phase 4)
	MinioEnabled      bool   `mapstructure:"MINIO_ENABLED"`
	MinioAPIPort      string `mapstructure:"MINIO_API_PORT"`
	MinioRootUser     string `mapstructure:"MINIO_ROOT_USER"`
	MinioRootPassword string `mapstructure:"MINIO_ROOT_PASSWORD"`
	MinioBucket       string `mapstructure:"MINIO_BUCKET"`

	// OpenTelemetry & Jaeger Observability (Phase 5)
	OtelEnabled    bool   `mapstructure:"OTEL_ENABLED"`
	JaegerOtlpPort string `mapstructure:"JAEGER_OTLP_GRPC_PORT"`
}

// LoadConfig reads configuration from .env file and environment variables using Viper.
func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Set fallback sensible defaults
	viper.SetDefault("APP_NAME", "Biblio")
	viper.SetDefault("APP_PORT", "3000")
	viper.SetDefault("DB_PORT", "5433")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_TIMEZONE", "Asia/Jakarta")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read .env file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config into struct: %w", err)
	}

	return &cfg, nil
}
