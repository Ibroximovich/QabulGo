package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	Port              string
	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	JWTSecret         string
	JWTIssuer         string
	JWTAudience       string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	CookieSecure      bool
	CookieDomain      string
	FrontendURL       string
	TermsVersion      string
	EmailMock         bool
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPassword      string
	SMTPFrom          string
	RateLimitRPS              float64
	RateLimitBurst            int
	TelegramClinicBotToken    string
	TelegramClinicBotUsername string
	TelegramAdminBotToken     string
	TelegramAdminBotUsername  string
	TelegramAdminAllowedIDs   []int64
	AppPublicURL              string
}

// LoadConfig loads application configuration from environment variables or .env file
func LoadConfig() (*Config, error) {
	// Try loading .env if it exists, ignore if not found
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		AppEnv:                    getEnv("APP_ENV", "development"),
		Port:                      getEnv("PORT", "8080"),
		DatabaseURL:               getEnv("DATABASE_URL", "postgres://qabulgo:qabulgo_secret_2026@localhost:5433/qabulgo?sslmode=disable"),
		DBMaxOpenConns:            getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:            getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime:         getEnvAsDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		JWTSecret:                 getEnv("JWT_SECRET", "super_secret_jwt_signing_key_qabulgo_2026_minimum_32_chars!"),
		JWTIssuer:                 getEnv("JWT_ISSUER", "qabulgo-api"),
		JWTAudience:               getEnv("JWT_AUDIENCE", "qabulgo-web"),
		AccessTokenTTL:            getEnvAsDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:           getEnvAsDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		CookieSecure:              getEnvAsBool("COOKIE_SECURE", false),
		CookieDomain:              getEnv("COOKIE_DOMAIN", "localhost"),
		FrontendURL:               getEnv("FRONTEND_URL", "http://localhost:3000"),
		TermsVersion:              getEnv("TERMS_VERSION", "v1.0"),
		EmailMock:                 getEnvAsBool("EMAIL_MOCK", true),
		SMTPHost:                  getEnv("SMTP_HOST", ""),
		SMTPPort:                  getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:                  getEnv("SMTP_USER", ""),
		SMTPPassword:              getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                  getEnv("SMTP_FROM", "QabulGo <no-reply@qabulgo.uz>"),
		RateLimitRPS:              getEnvAsFloat("RATE_LIMIT_RPS", 10.0),
		RateLimitBurst:            getEnvAsInt("RATE_LIMIT_BURST", 20),
		TelegramClinicBotToken:    getEnv("TELEGRAM_CLINIC_BOT_TOKEN", ""),
		TelegramClinicBotUsername: strings.TrimPrefix(getEnv("TELEGRAM_CLINIC_BOT_USERNAME", "QabulGoClinicBot"), "@"),
		TelegramAdminBotToken:     getEnv("TELEGRAM_ADMIN_BOT_TOKEN", ""),
		TelegramAdminBotUsername:  strings.TrimPrefix(getEnv("TELEGRAM_ADMIN_BOT_USERNAME", "QabulGoAdminBot"), "@"),
		TelegramAdminAllowedIDs:   getEnvAsInt64Slice("TELEGRAM_ADMIN_ALLOWED_IDS"),
		AppPublicURL:              getEnv("APP_PUBLIC_URL", getEnv("FRONTEND_URL", "http://localhost:3000")),
	}

	if len(cfg.JWTSecret) < 32 && cfg.AppEnv == "production" {
		log.Println("WARNING: JWT_SECRET should be at least 32 characters in production!")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseFloat(valStr, 64); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsInt64Slice(key string) []int64 {
	valStr := getEnv(key, "")
	if valStr == "" {
		return nil
	}
	parts := strings.Split(valStr, ",")
	res := make([]int64, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil && id != 0 {
			res = append(res, id)
		}
	}
	return res
}
