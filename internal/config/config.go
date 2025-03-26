package config

import (
	"encoding/hex"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort     string
	DBDSN          string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	CookieHashKey  []byte
	CookieBlockKey []byte
	BcryptCost     int
	TokenExpiry    time.Duration
}

func Load() (*Config, error) {
	// Set defaults
	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		DBDSN:         getEnv("DB_DSN", "heartrate.db"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		BcryptCost:    getEnvAsInt("BCRYPT_COST", 10),
		TokenExpiry:   24 * time.Hour,
	}

	// Load cookie keys
	hashKey, err := decodeHexKey(getEnv("COOKIE_HASH_KEY", ""), 64)
	if err != nil {
		return nil, err
	}
	cfg.CookieHashKey = hashKey

	blockKey, err := decodeHexKey(getEnv("COOKIE_BLOCK_KEY", ""), 32)
	if err != nil {
		return nil, err
	}
	cfg.CookieBlockKey = blockKey

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

func decodeHexKey(hexKey string, expectedBytes int) ([]byte, error) {
	if hexKey == "" {
		return nil, nil
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	if len(key) != expectedBytes {
		return nil, nil
	}
	return key, nil
}
