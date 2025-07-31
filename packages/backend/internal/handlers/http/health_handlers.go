package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	kiteConf "github.com/konflux-ci/kite/internal/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type HealthStatus struct {
	Status     string                     `json:"status"`
	Message    string                     `json:"message"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

type ComponentHealth struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func NewHealthHandler(db *gorm.DB, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Initialize response structure
		health := HealthStatus{
			Timestamp:  time.Now().UTC(),
			Components: make(map[string]ComponentHealth),
		}

		// Track overall health
		overallHealthy := true

		// Check database health
		dbHealth := checkDatabaseHealth(db, logger)
		health.Components["database"] = dbHealth
		if dbHealth.Status != "UP" {
			overallHealthy = false
		}

		// Check API health
		apiHealth := checkAPIHealth()
		health.Components["api"] = apiHealth

		// Add response time
		responseTime := time.Since(startTime)
		health.Components["response_time"] = ComponentHealth{
			Status:  "UP",
			Message: "Response time measurement",
			Details: map[string]interface{}{
				"duration_seconds": responseTime.Seconds(),
			},
		}

		if overallHealthy {
			health.Status = "UP"
			health.Message = "All systems operational"
			c.JSON(http.StatusOK, health)
		} else {
			health.Status = "DOWN"
			health.Message = "One or more components are unhealthy"
			c.JSON(http.StatusServiceUnavailable, health)
		}
	}
}

// checkDatabaseHealth performs a real-time database health check
func checkDatabaseHealth(db *gorm.DB, logger *logrus.Logger) ComponentHealth {
	start := time.Now()
	dbHealth, err := kiteConf.CheckDatabaseHealth(db)
	duration := time.Since(start)
	if err != nil {
		logger.WithError(err).Error("Database health check failed")
		return ComponentHealth{
			Status:  "DOWN",
			Message: err.Error(),
			Details: map[string]interface{}{
				"check_duration_seconds": duration.Seconds(),
				"cause_of_failure":       fmt.Sprintf("Database ping failed: %v", err),
			},
		}
	}

	return ComponentHealth{
		Status:  "UP",
		Message: "Database connection successful",
		Details: dbHealth,
	}
}

func checkAPIHealth() ComponentHealth {
	return ComponentHealth{
		Status:  "UP",
		Message: "API server is responding",
		Details: map[string]interface{}{
			"version": kiteConf.GetEnvOrDefault("KITE_VERSION", "0.0.1"),
		},
	}
}
