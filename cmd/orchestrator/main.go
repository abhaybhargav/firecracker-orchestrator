package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/abhaybhargav/firecracker-orchestrator/internal/config"
	"github.com/abhaybhargav/firecracker-orchestrator/internal/database"
	"github.com/abhaybhargav/firecracker-orchestrator/pkg/api"
	"github.com/abhaybhargav/firecracker-orchestrator/pkg/firecracker"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.Info("Starting Firecracker Orchestrator")

	// Initialize database
	var db *database.Database

	if cfg.DatabaseDriver == "sqlite3" {
		// Use CGO-based SQLite driver (faster but requires CGO)
		db, err = database.NewDatabase(cfg.DatabasePath)
	} else {
		// Use pure Go SQLite driver (slower but no CGO required)
		db, err = database.NewPureGoDatabase(cfg.DatabasePath)
	}

	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	logger.Info("Database initialized successfully")

	// Initialize Firecracker manager
	vmManager := firecracker.NewManager(cfg, db, logger)
	logger.Info("Firecracker manager initialized")

	// Setup Gin router
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware for API requests
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize API server
	apiServer := api.NewServer(vmManager, db, logger)
	apiServer.SetupRoutes(r)

	logger.Infof("Server starting on %s", cfg.Address())

	// Start server in a goroutine
	go func() {
		if err := r.Run(cfg.Address()); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	logger.Info("Shutting down server...")

	// TODO: Implement graceful shutdown
	// - Stop all running VMs
	// - Close database connections
	// - Clean up resources

	logger.Info("Server stopped")
}
