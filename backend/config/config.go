package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	AI          AIConfig
	Storage     StorageConfig
	Qdrant      QdrantConfig
	Email       EmailConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type AIConfig struct {
	OpenAI    OpenAIConfig
	Gemini    GeminiConfig
	Claude    ClaudeConfig
	DefaultProvider string
}

type OpenAIConfig struct {
	APIKey string
	Model  string
}

type GeminiConfig struct {
	APIKey string
	Model  string
}

type ClaudeConfig struct {
	APIKey string
	Model  string
}

type StorageConfig struct {
	Type     string
	Endpoint string
	AccessKey string
	SecretKey string
	Bucket   string
}

type QdrantConfig struct {
	Host string
	Port string
}

type EmailConfig struct {
	Username string
	Password string
}

func LoadConfig() (*Config, error) {
	// Load .env file if exists
	godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "herlin"),
			Password: getEnv("DB_PASSWORD", "herlin123"),
			DBName:   getEnv("DB_NAME", "herlin_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration: time.Hour * 24 * 7, // 7 days
		},
		AI: AIConfig{
			OpenAI: OpenAIConfig{
				APIKey: getEnv("OPENAI_API_KEY", ""),
				Model:  getEnv("OPENAI_MODEL", "gpt-4"),
			},
			Gemini: GeminiConfig{
				APIKey: getEnv("GEMINI_API_KEY", ""),
				Model:  getEnv("GEMINI_MODEL", "gemini-pro"),
			},
			Claude: ClaudeConfig{
				APIKey: getEnv("CLAUDE_API_KEY", ""),
				Model:  getEnv("CLAUDE_MODEL", "claude-3-opus-20240229"),
			},
			DefaultProvider: getEnv("AI_DEFAULT_PROVIDER", "openai"),
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "minio"),
			Endpoint:  getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("STORAGE_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("STORAGE_BUCKET", "herlin-files"),
		},
		Qdrant: QdrantConfig{
			Host: getEnv("QDRANT_HOST", "localhost"),
			Port: getEnv("QDRANT_PORT", "6333"),
		},
		Email: EmailConfig{
			Username: getEnv("EMAIL_USERNAME", ""),
			Password: getEnv("EMAIL_PASSWORD", ""),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
