package config

import (
	"context"
	"crypto/ecdsa"
	"os"
	"strconv"

	"github.com/graytonio/warframe-wishlist/pkg/logger"
	"github.com/lestrrat-go/jwx/jwk"
)

type Config struct {
	ServerPort           string
	MongoURI             string
	MongoDatabase        string
	SupabaseURL          string
	SupabaseJWTPublicKey *ecdsa.PublicKey
	AllowedOrigins       string
	LogLevel             string
}

func Load() *Config {
	return &Config{
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		MongoURI:             getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:        getEnv("MONGO_DATABASE", "warframe"),
		SupabaseURL:          getEnv("SUPABASE_URL", ""),
		SupabaseJWTPublicKey: parseJWTPublicKey(getEnv("SUPABASE_JWT_PUBLIC_KEY", "")),
		AllowedOrigins:       getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
	}
}

func parseJWTPublicKey(publicKey string) *ecdsa.PublicKey {
	key, err := jwk.ParseKey([]byte(publicKey))
	if err != nil {
		logger.Error(context.Background(), "failed to parse JWT public key", "error", err)
		panic(err)
	}

	var raw interface{}
	err = key.Raw(&raw)
	if err != nil {
		logger.Error(context.Background(), "failed to get raw key: %v", err)
		panic(err)
	}

	public, ok := raw.(*ecdsa.PublicKey)
	if !ok {
		logger.Error(context.Background(), "failed to cast raw key to *ecdsa.PublicKey")
		panic("failed to cast raw key to *ecdsa.PublicKey")
	}

	return public
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
