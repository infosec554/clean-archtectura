package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	AppName     string
	AppPort     string
	Environment string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisTTL      time.Duration

	JWTSecretKey      string
	AccessExpireTime  time.Duration
	RefreshExpireTime time.Duration

	BrevoAPIKey     string
	BrevoSenderEmail string
	BrevoSenderName  string
}

func Load() Config {
	_ = godotenv.Load(".env")

	cfg := Config{}

	cfg.AppName = cast.ToString(getOrDefault("APP_NAME", "clean-archtectura"))
	cfg.AppPort = cast.ToString(getOrDefault("APP_PORT", ":8080"))
	cfg.Environment = cast.ToString(getOrDefault("ENVIRONMENT", "development"))

	cfg.PostgresHost = cast.ToString(getOrDefault("DB_HOST", "localhost"))
	cfg.PostgresPort = cast.ToString(getOrDefault("DB_PORT", "5432"))
	cfg.PostgresUser = cast.ToString(getOrDefault("DB_USER", "postgres"))
	cfg.PostgresPassword = cast.ToString(getOrDefault("DB_PASSWORD", "1234"))
	cfg.PostgresDB = cast.ToString(getOrDefault("DB_NAME", "mydb"))

	cfg.RedisHost = cast.ToString(getOrDefault("REDIS_HOST", "localhost"))
	cfg.RedisPort = cast.ToString(getOrDefault("REDIS_PORT", "6379"))
	cfg.RedisPassword = cast.ToString(getOrDefault("REDIS_PASSWORD", ""))
	cfg.RedisDB = cast.ToInt(getOrDefault("REDIS_DB", 0))
	cfg.RedisTTL = cast.ToDuration(getOrDefault("REDIS_TTL", "10m"))

	cfg.JWTSecretKey = cast.ToString(getOrDefault("JWT_SECRET_KEY", "supersecretkey"))
	cfg.AccessExpireTime = cast.ToDuration(getOrDefault("ACCESS_TOKEN_TTL", "24h"))
	cfg.RefreshExpireTime = cast.ToDuration(getOrDefault("REFRESH_TOKEN_TTL", "168h"))

	cfg.BrevoAPIKey      = cast.ToString(getOrDefault("BREVO_API_KEY", ""))
	cfg.BrevoSenderEmail = cast.ToString(getOrDefault("BREVO_SENDER_EMAIL", "noreply@example.com"))
	cfg.BrevoSenderName  = cast.ToString(getOrDefault("BREVO_SENDER_NAME", "MyApp"))

	return cfg
}

func getOrDefault(key string, defaultValue any) any {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
