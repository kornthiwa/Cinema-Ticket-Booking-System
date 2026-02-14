package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort     int
	MongoURI       string
	RedisAddr      string
	JWTSecret      string
	LockTTLSeconds int
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	lockTTL, _ := strconv.Atoi(getEnv("LOCK_TTL_SECONDS", "300"))
	return &Config{
		ServerPort:     port,
		MongoURI:       getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret"),
		LockTTLSeconds: lockTTL,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
