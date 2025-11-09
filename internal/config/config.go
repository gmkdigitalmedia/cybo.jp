package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	GPU       GPUConfig
	Scanner   ScannerConfig
	Auth      AuthConfig
	Storage   StorageConfig
	StartTime time.Time
}

type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type GPUConfig struct {
	DeviceID        int
	CacheSize       int64 // in bytes
	ColorCorrection bool
	BatchSize       int
}

type ScannerConfig struct {
	Protocol string // "tcp" or "serial"
	Address  string // IP:Port or serial device path
	Timeout  time.Duration
}

type AuthConfig struct {
	JWTSecret     string
	TokenExpiry   time.Duration
	AllowedUsers  []string
	PasswordHash  string
}

type StorageConfig struct {
	BasePath       string
	TempPath       string
	MaxSlideSize   int64
	RetentionDays  int
}

func Load() (*Config, error) {
	cfg := &Config{
		StartTime: time.Now(),
		Server: ServerConfig{
			Port:            getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:     time.Duration(getEnvInt("READ_TIMEOUT", 30)) * time.Second,
			WriteTimeout:    time.Duration(getEnvInt("WRITE_TIMEOUT", 30)) * time.Second,
			ShutdownTimeout: time.Duration(getEnvInt("SHUTDOWN_TIMEOUT", 10)) * time.Second,
		},
		GPU: GPUConfig{
			DeviceID:        getEnvInt("GPU_DEVICE_ID", 0),
			CacheSize:       int64(getEnvInt("GPU_CACHE_SIZE", 8192)) * 1024 * 1024, // MB to bytes
			ColorCorrection: getEnvBool("GPU_COLOR_CORRECTION", true),
			BatchSize:       getEnvInt("GPU_BATCH_SIZE", 16),
		},
		Scanner: ScannerConfig{
			Protocol: getEnv("SCANNER_PROTOCOL", "tcp"),
			Address:  getEnv("SCANNER_ADDRESS", "localhost:9090"),
			Timeout:  time.Duration(getEnvInt("SCANNER_TIMEOUT", 30)) * time.Second,
		},
		Auth: AuthConfig{
			JWTSecret:   getEnv("JWT_SECRET", generateRandomSecret()),
			TokenExpiry: time.Duration(getEnvInt("TOKEN_EXPIRY", 24)) * time.Hour,
		},
		Storage: StorageConfig{
			BasePath:      getEnv("STORAGE_PATH", "./data/slides"),
			TempPath:      getEnv("TEMP_PATH", "./data/temp"),
			MaxSlideSize:  int64(getEnvInt("MAX_SLIDE_SIZE", 50)) * 1024 * 1024 * 1024, // GB
			RetentionDays: getEnvInt("RETENTION_DAYS", 365),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.GPU.DeviceID < 0 {
		return fmt.Errorf("invalid GPU device ID: %d", c.GPU.DeviceID)
	}

	if c.GPU.CacheSize < 1024*1024 {
		return fmt.Errorf("GPU cache size too small: %d bytes", c.GPU.CacheSize)
	}

	if c.Scanner.Protocol != "tcp" && c.Scanner.Protocol != "serial" {
		return fmt.Errorf("invalid scanner protocol: %s", c.Scanner.Protocol)
	}

	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret cannot be empty")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func generateRandomSecret() string {
	// In production, this should be loaded from secure storage
	return "CHANGE_THIS_IN_PRODUCTION_" + strconv.FormatInt(time.Now().Unix(), 10)
}
