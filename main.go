package main

import (
	"fmt"
	"net/http"

	"github.com/ariesmaulana/payroll/app/timeclock"
	"github.com/ariesmaulana/payroll/app/user"
	"github.com/ariesmaulana/payroll/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/ariesmaulana/payroll/internal/jwtutil"
	"github.com/ariesmaulana/payroll/lib/database"
	"github.com/ariesmaulana/payroll/lib/logger"
	customMiddleware "github.com/ariesmaulana/payroll/lib/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	logCfg := logger.LogConfig{
		Debug:    true,   // set to false in production
		FilePath: "logs", // logs will be stored in ./logs directory
		MaxSize:  100,    // rotate logs when file reaches 100MB
	}

	if err := logger.Init(logCfg); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	jwtutil.SetSecret(cfg.JWTSecret)

	// Initialize database connection pool
	pool, err := database.NewPostgresPool(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create database pool")
	}
	defer pool.Close()

	// Initialize user components
	userStorage := user.NewStorage(pool)
	userService := user.NewService(userStorage)
	userHandler := user.NewHandler(userService)

	//Initialize timeclock component
	// Setup order (tanpa storage, dummy service aja)
	timeClockStorage := timeclock.NewStorage(pool)
	timeClockService := timeclock.NewService(timeClockStorage, userService)
	timeClockHandler := timeclock.NewHandler(timeClockService)

	// Setup router with middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.TraceMiddleware) // Our custom trace middleware

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Register routes
	user.RegisterRoutes(r, userHandler)
	timeclock.RegisterRoutes(r, timeClockHandler)

	// Start the server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Info().Msgf("Server is running on port %s", cfg.ServerPort)
	log.Fatal().Err(http.ListenAndServe(addr, r)).Msg("Server failed")
}
