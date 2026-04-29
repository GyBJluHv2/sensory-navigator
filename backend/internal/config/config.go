package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config — типизированные настройки, считываемые из переменных окружения.
type Config struct {
	Port            string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	JWTSecret       string
	JWTAccessTTL    time.Duration
	AllowedOrigins  []string
	BCryptCost      int
	NearbyMaxRadius int
}

func Load() *Config {
	cfg := &Config{
		Port:            getenv("APP_PORT", "8080"),
		DBHost:          getenv("DB_HOST", "localhost"),
		DBPort:          getenv("DB_PORT", "5432"),
		DBUser:          getenv("DB_USER", "navigator"),
		DBPassword:      getenv("DB_PASSWORD", "navigator"),
		DBName:          getenv("DB_NAME", "sensory_navigator"),
		DBSSLMode:       getenv("DB_SSLMODE", "disable"),
		JWTSecret:       getenv("JWT_SECRET", "change-me-in-production"),
		JWTAccessTTL:    parseDuration(getenv("JWT_TTL", "168h")),
		AllowedOrigins:  []string{"*"},
		BCryptCost:      atoi(getenv("BCRYPT_COST", "10")),
		NearbyMaxRadius: atoi(getenv("NEARBY_MAX_RADIUS", "50000")),
	}
	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Europe/Moscow",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 168 * time.Hour
	}
	return d
}