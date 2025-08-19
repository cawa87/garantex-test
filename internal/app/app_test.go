package app

import (
	"testing"
	"time"

	"github.com/cawa87/garantex-test/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			GRPCPort:    50051,
			HTTPPort:    8080,
			MetricsPort: 9090,
			Timeout:     30 * time.Second,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			DBName:   "garantex_test",
			SSLMode:  "disable",
		},
		Exchange: config.ExchangeConfig{
			BaseURL: "https://grinex.io",
			Timeout: 10 * time.Second,
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	app, err := New(cfg)

	// Note: This test will fail if PostgreSQL is not running
	// In a real test environment, you would use a test database or mock
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
	}

	assert.NotNil(t, app)
	assert.NotNil(t, app.config)
	assert.NotNil(t, app.logger)
}

func TestApp_Shutdown(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			GRPCPort:    50051,
			HTTPPort:    8080,
			MetricsPort: 9090,
			Timeout:     30 * time.Second,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			DBName:   "garantex_test",
			SSLMode:  "disable",
		},
		Exchange: config.ExchangeConfig{
			BaseURL: "https://grinex.io",
			Timeout: 10 * time.Second,
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	app, err := New(cfg)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
	}

	// Test shutdown
	err = app.Shutdown()
	assert.NoError(t, err)
}

func TestConfig_GetDSN(t *testing.T) {
	dbConfig := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	dsn := dbConfig.GetDSN()
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)
}
