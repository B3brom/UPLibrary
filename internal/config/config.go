package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBPath    string
	UploadDir string
	JWTSecret string
	JWTTTL    time.Duration
	LogLevel  string
}

func Load() *Config {
	_ = godotenv.Load()
	dataDir := defaultDataDir()
	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		ttl = 24 * time.Hour
	}
	return &Config{
		Port:      getEnv("SERVER_PORT", "8080"),
		DBPath:    getEnv("DB_PATH", filepath.Join(dataDir, "store.db")),
		UploadDir: getEnv("UPLOAD_DIR", filepath.Join(dataDir, "uploads")),
		JWTSecret: getEnv("JWT_SECRET", "secret"),
		JWTTTL:    ttl,
		LogLevel:  getEnv("LOG_LEVEL", "info"),
	}
}

func defaultDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "ClothingStore")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
