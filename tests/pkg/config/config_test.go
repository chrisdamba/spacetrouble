package config_test

import (
	"github.com/chrisdamba/spacetrouble/pkg/config"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	os.Clearenv()

	cfg, err := config.NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ":5000", cfg.Server.Address)
	assert.Equal(t, 15*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "space", cfg.Database.Name)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "", cfg.Database.Password)
	assert.Equal(t, 99, cfg.Database.MaxPoolConns)
	assert.Equal(t, "https://api.spacexdata.com/v4", cfg.SpaceX.BaseURL)
}

func TestNewConfigWithEnvVars(t *testing.T) {
	os.Clearenv()

	envVars := map[string]string{
		"SERVER_ADDRESS":       ":8080",
		"SERVER_WRITE_TIMEOUT": "30s",
		"SERVER_READ_TIMEOUT":  "30s",
		"SERVER_IDLE_TIMEOUT":  "60s",
		"POSTGRES_HOST":        "db.example.com",
		"POSTGRES_PORT":        "5433",
		"POSTGRES_DB":          "testdb",
		"POSTGRES_USER":        "testuser",
		"POSTGRES_PASSWORD":    "testpass",
		"MAX_CONNS":            "50",
		"SPACEX_URL":           "https://api.spacex.com/v5",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}

	cfg, err := config.NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test environment values
	assert.Equal(t, ":8080", cfg.Server.Address)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "5433", cfg.Database.Port)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, 50, cfg.Database.MaxPoolConns)
	assert.Equal(t, "https://api.spacex.com/v5", cfg.SpaceX.BaseURL)
}

func TestDatabaseDSN(t *testing.T) {
	dbConfig := config.DatabaseConfig{
		Host:         "localhost",
		Port:         "5432",
		Name:         "testdb",
		User:         "testuser",
		Password:     "testpass",
		MaxPoolConns: 50,
	}

	expected := "host=localhost port=5432 dbname=testdb user=testuser password=testpass pool_max_conns=50"
	assert.Equal(t, expected, dbConfig.DSN())
}

func TestInvalidConfigurations(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "Invalid write timeout",
			envVars: map[string]string{
				"SERVER_WRITE_TIMEOUT": "invalid",
			},
		},
		{
			name: "Invalid read timeout",
			envVars: map[string]string{
				"SERVER_READ_TIMEOUT": "invalid",
			},
		},
		{
			name: "Invalid idle timeout",
			envVars: map[string]string{
				"SERVER_IDLE_TIMEOUT": "invalid",
			},
		},
		{
			name: "Invalid max connections",
			envVars: map[string]string{
				"MAX_CONNS": "invalid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			_, err := config.NewConfig()
			assert.Error(t, err)
		})
	}
}
