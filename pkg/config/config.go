package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	SpaceX   SpaceXConfig
}

type ServerConfig struct {
	Address      string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         string
	Name         string
	User         string
	Password     string
	MaxPoolConns int
}

type SpaceXConfig struct {
	BaseURL string
}

func (dc *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s pool_max_conns=%d",
		dc.Host,
		dc.Port,
		dc.Name,
		dc.User,
		dc.Password,
		dc.MaxPoolConns,
	)
}

func NewConfig() (*Config, error) {
	serverCfg, err := newServerConfig()
	if err != nil {
		return nil, fmt.Errorf("server config error: %w", err)
	}

	dbCfg, err := newDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("database config error: %w", err)
	}

	spaceXCfg := newSpaceXConfig()

	return &Config{
		Server:   serverCfg,
		Database: dbCfg,
		SpaceX:   spaceXCfg,
	}, nil
}

func newServerConfig() (ServerConfig, error) {
	writeTimeout, err := getDurationFromEnv("SERVER_WRITE_TIMEOUT", "15s")
	if err != nil {
		return ServerConfig{}, fmt.Errorf("write timeout parse error: %w", err)
	}

	readTimeout, err := getDurationFromEnv("SERVER_READ_TIMEOUT", "15s")
	if err != nil {
		return ServerConfig{}, fmt.Errorf("read timeout parse error: %w", err)
	}

	idleTimeout, err := getDurationFromEnv("SERVER_IDLE_TIMEOUT", "30s")
	if err != nil {
		return ServerConfig{}, fmt.Errorf("idle timeout parse error: %w", err)
	}

	return ServerConfig{
		Address:      getEnvOrDefault("SERVER_ADDRESS", ":5000"),
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}, nil
}

func newDatabaseConfig() (DatabaseConfig, error) {
	maxConns, err := strconv.Atoi(getEnvOrDefault("MAX_CONNS", "99"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("max connections parse error: %w", err)
	}

	return DatabaseConfig{
		Host:         getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:         getEnvOrDefault("POSTGRES_PORT", "5432"),
		Name:         getEnvOrDefault("POSTGRES_DB", "space"),
		User:         getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:     getEnvOrDefault("POSTGRES_PASSWORD", ""),
		MaxPoolConns: maxConns,
	}, nil
}

func newSpaceXConfig() SpaceXConfig {
	return SpaceXConfig{
		BaseURL: getEnvOrDefault("SPACEX_URL", "https://api.spacexdata.com/v4"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationFromEnv(key, defaultValue string) (time.Duration, error) {
	return time.ParseDuration(getEnvOrDefault(key, defaultValue))
}
