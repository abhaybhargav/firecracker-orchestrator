package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Host string
	Port int

	// Database configuration
	DatabasePath   string
	DatabaseDriver string // "sqlite3" (CGO) or "sqlite" (pure Go)

	// Firecracker configuration
	FirecrackerBinary string
	KernelPath        string
	RootfsPath        string
	SocketDir         string

	// Networking configuration
	BridgeName    string
	TAPDeviceBase string

	// VM defaults
	DefaultMemoryMB int64
	DefaultCPUs     int
	DefaultDiskGB   int64

	// Logging
	LogLevel string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	config := &Config{
		Host:              getEnv("HOST", "0.0.0.0"),
		Port:              getEnvAsInt("PORT", 8080),
		DatabasePath:      getEnv("DATABASE_PATH", "./orchestrator.db"),
		DatabaseDriver:    getEnv("DATABASE_DRIVER", "sqlite"), // Default to pure Go
		FirecrackerBinary: getEnv("FIRECRACKER_BINARY", "/usr/bin/firecracker"),
		KernelPath:        getEnv("KERNEL_PATH", "./vm-images/vmlinux.bin"),
		RootfsPath:        getEnv("ROOTFS_PATH", "./vm-images/rootfs.ext4"),
		SocketDir:         getEnv("SOCKET_DIR", "/tmp/firecracker"),
		BridgeName:        getEnv("BRIDGE_NAME", "fc-br0"),
		TAPDeviceBase:     getEnv("TAP_DEVICE_BASE", "fc-tap"),
		DefaultMemoryMB:   getEnvAsInt64("DEFAULT_MEMORY_MB", 512),
		DefaultCPUs:       getEnvAsInt("DEFAULT_CPUS", 1),
		DefaultDiskGB:     getEnvAsInt64("DEFAULT_DISK_GB", 2),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}

	return config
}

// Address returns the server address
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 gets an environment variable as an int64 or returns a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
