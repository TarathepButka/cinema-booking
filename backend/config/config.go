package config

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// loadDotEnv reads a local .env file and sets environment variables if they are not already set.
func loadDotEnv() {
	file, err := os.Open(".env")
	if err != nil {
		// .env is optional, skip if it doesn't exist
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip quotes if wrapped
		if len(val) >= 2 && ((strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) || (strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'"))) {
			val = val[1 : len(val)-1]
		}

		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("warning: error scanning .env file: %v", err)
	}
}

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port    string
	GinMode string

	// MongoDB
	MongoURI string
	MongoDB  string

	// Redis
	RedisAddr     string
	RedisPassword string

	// JWT
	JWTSecret      string
	JWTExpireHours int

	// Google OAuth 2.0
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
	AppOrigins         []string
	CookieSecure       bool

	// Resend Email
	ResendAPIKey    string
	ResendFromEmail string

	// Admin (used for admin login fallback)
	AdminEmail    string
	AdminPassword string
}

// Load reads configuration from environment variables.
// It panics on missing required fields to fail fast on startup.
func Load() *Config {
	loadDotEnv()
	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		MongoURI: loadMongoURI(),
		MongoDB:  getEnv("MONGODB_DB", "cinemadb"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		JWTSecret:      requireEnv("JWT_SECRET"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		AppOrigins:         getEnvList("APP_ORIGINS", []string{"http://localhost:5173"}),
		CookieSecure:       getEnvBool("COOKIE_SECURE", false),

		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		ResendFromEmail: getEnv("RESEND_FROM_EMAIL", "noreply@cinema.com"),

		AdminEmail:    requireEnv("ADMIN_EMAIL"),
		AdminPassword: requireEnv("ADMIN_PASSWORD"),
	}

	if cfg.GinMode == "release" {
		if len(cfg.JWTSecret) < 32 {
			panic("JWT_SECRET must be at least 32 characters in release mode")
		}
		if len(cfg.AdminPassword) < 12 {
			panic("ADMIN_PASSWORD must contain at least 12 characters in release mode")
		}
		if len(cfg.AppOrigins) == 0 {
			panic("APP_ORIGINS must contain at least one trusted origin in release mode")
		}
	}

	return cfg
}

func loadMongoURI() string {
	if uri := strings.TrimSpace(os.Getenv("MONGODB_URI")); uri != "" {
		return addLocalMongoOptions(uri)
	}

	user := requireEnv("MONGO_USER")
	password := requireEnv("MONGO_PASSWORD")
	host := getEnv("MONGO_HOST", "localhost:27017")
	database := getEnv("MONGODB_DB", "cinemadb")
	replicaSet := getEnv("MONGO_REPLICA_SET", "rs0")
	credentials := url.UserPassword(user, password).String()
	query := url.Values{
		"authSource": {getEnv("MONGO_AUTH_SOURCE", "admin")},
		"replicaSet": {replicaSet},
	}
	if getEnvBool("MONGO_DIRECT_CONNECTION", false) {
		query.Set("directConnection", "true")
	}
	return fmt.Sprintf("mongodb://%s@%s/%s?%s", credentials, host, database, query.Encode())
}

func addLocalMongoOptions(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	host := strings.ToLower(parsed.Hostname())
	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return uri
	}

	query := parsed.Query()
	if query.Get("replicaSet") == "" {
		query.Set("replicaSet", getEnv("MONGO_REPLICA_SET", "rs0"))
	}
	if _, exists := query["directConnection"]; !exists {
		query.Set("directConnection", strconv.FormatBool(getEnvBool("MONGO_DIRECT_CONNECTION", true)))
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// getEnv returns the env variable or the fallback default.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// requireEnv returns the env variable or panics if not set.
func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return val
}

// getEnvInt returns the env variable as int or the fallback default.
func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvList(key string, fallback []string) []string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}

	var result []string
	for _, item := range strings.Split(val, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, strings.TrimRight(trimmed, "/"))
		}
	}
	return result
}
