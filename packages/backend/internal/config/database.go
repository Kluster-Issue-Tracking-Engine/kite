package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Returns the database configuration using ENV variables. Uses defaults if ENV variables are not found.
func GetDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnvOrDefault("KITE_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("KITE_DB_PORT", "5432"),
		User:     getEnvOrDefault("KITE_DB_USER", "postgres"),
		Password: getEnvOrDefault("KITE_DB_PASSWORD", "postgres"),
		Name:     getEnvOrDefault("KITE_DB_NAME", "issuesdb"),
		SSLMode:  getEnvOrDefault("KITE_DB_SSL_MODE", "disable"),
	}
}

// Initializes the database.
func InitDatabase() (*gorm.DB, error) {
	config := GetDatabaseConfig()

	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.Host, config.User, config.Password, config.Name, config.Port, config.SSLMode)

	var gormLogger logger.Interface
	if os.Getenv("KITE_PROJECT_ENV") == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// DB connection timeout settings
	maxRetries := GetEnvIntOrDefault("KITE_DB_MAX_RETRIES", 10)
	delay := GetEnvDurationOrDefault("KITE_DB_RETRY_DELAY", 5*time.Second)

	db, err := connectWithRetries(connectionString, gormLogger, maxRetries, delay)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	// Keep x idle connections open
	sqlDB.SetMaxIdleConns(GetEnvIntOrDefault("KITE_DB_MAX_IDLE_CONNS", 10))
	// Max number of DB connections allowed to be open at the same time
	sqlDB.SetMaxOpenConns(GetEnvIntOrDefault("KITE_DB_MAX_OPEN_CONNS", 100))
	// Refresh the connection periodically
	sqlDB.SetConnMaxLifetime(GetEnvDurationOrDefault("KITE_DB_CONN_MAX_LIFETIME", 1*time.Hour))

	log.Println("Database connection established successfully")
	return db, nil
}

// Connects to the specified database a specific number of times (maxRetries) with a delay for each retry.
//
// The delay strategy uses a linear backoff (delay × attempt number).
// This helps reduce pressure on the DB and gives it time to recover on each retry.
func connectWithRetries(connectionString string, gormLogger logger.Interface, maxRetries int, delay time.Duration) (*gorm.DB, error) {
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
			Logger: gormLogger,
		})
		if err == nil {
			sqlDB, err := db.DB()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Ping the DB with timeout to test connection
			if err == nil && sqlDB.PingContext(ctx) == nil {
				return db, nil
			}
		}

		log.Printf("Database connection attempt %d failed: %v", i+1, err)
		// Lets avoid hammering the DB and use a linear backoff
		backoff := delay * time.Duration(i+1)
		time.Sleep(backoff)
	}
	return nil, fmt.Errorf("could not connect to database after %d attempts: %w", maxRetries, err)
}

// Gets an ENV variable, returns a defaultValue if not found.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Structured database health details
type DatabaseHealthDetails struct {
	ConnectionStatus string  `json:"connection_status"`
	ResponseTime     float64 `json:"response_time_seconds"`
	OpenConnections  int     `json:"open_connections"`
	IdleConnections  int     `json:"idle_connections"`
	MaxOpenConns     int     `json:"max_open_connections"`
}

// Performs database health checks and returns detailed stats
func CheckDatabaseHealth(db *gorm.DB) (*DatabaseHealthDetails, error) {
	// Start timer
	start := time.Now()

	// Grab DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve underlying database: %w", err)
	}

	// Ping DB with timeout to test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return &DatabaseHealthDetails{
			ConnectionStatus: "Unhealthy",
		}, fmt.Errorf("database ping failed: %w", err)
	}

	// Check response time
	responseTime := time.Since(start)

	// Get connection pool stats
	stats := sqlDB.Stats()

	return &DatabaseHealthDetails{
		ConnectionStatus: "Healthy",
		ResponseTime:     responseTime.Seconds(),
		OpenConnections:  stats.OpenConnections,
		IdleConnections:  stats.Idle,
		MaxOpenConns:     stats.MaxOpenConnections,
	}, nil
}
