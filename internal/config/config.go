package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	SSHHost       string
	SSHPort       int
	SSHUser       string
	SSHPassword   string
	SSHPrivateKey string
	SSHTimeout    time.Duration
	StockfishPath string
	AnalysisDepth int
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		SSHHost:       getEnv("SSH_HOST", ""),
		SSHPort:       getEnvInt("SSH_PORT", 22),
		SSHUser:       getEnv("SSH_USER", ""),
		SSHPassword:   getEnv("SSH_PASSWORD", ""),
		SSHPrivateKey: getEnv("SSH_PRIVATE_KEY", ""),
		SSHTimeout:    getEnvDuration("SSH_TIMEOUT", 5*time.Second),
		StockfishPath: getEnv("STOCKFISH_PATH", "stockfish"),
		AnalysisDepth: getEnvInt("ANALYSIS_DEPTH", 12),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
