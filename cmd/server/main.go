package main

import (
	"log"
	"time"

	"github.com/cherdsak-kh/hospital-middleware-api/config"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/client"
	httpDelivery "github.com/cherdsak-kh/hospital-middleware-api/internal/delivery/http"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/repository"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/usecase"
)

// @title           Hospital Middleware API
// @version         1.0
// @description     Hospital Middleware API for searching and managing patient information from Hospital Information Systems (HIS) with hospital-level data isolation.
// @contact.name    Cherdsak Kh.
// @contact.email   cherd8524@gmail.com
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Enter "Bearer " followed by your JWT token

func main() {
	// 1. Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Fatal: failed to load configuration: %v\n", err)
	}

	// 2. Initialize PostgreSQL database connection and auto-migrations
	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize database: %v\n", err)
	}

	// 3. Initialize Repositories (Data Access Layer)
	hospitalRepo := repository.NewHospitalRepository(db)
	staffRepo := repository.NewStaffRepository(db)
	patientRepo := repository.NewPatientRepository(db)

	// 4. Initialize External Clients
	hisClient := client.NewHISClient(cfg.HospitalABaseURL, 5*time.Second)

	// 5. Initialize UseCases (Business Logic Layer)
	staffUseCase := usecase.NewStaffUseCase(staffRepo, hospitalRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	patientUseCase := usecase.NewPatientUseCase(patientRepo, hisClient)

	// 6. Initialize HTTP Handlers (Delivery Layer)
	staffHandler := httpDelivery.NewStaffHandler(staffUseCase)
	patientHandler := httpDelivery.NewPatientHandler(patientUseCase)

	// 7. Initialize Router
	router := httpDelivery.SetupRouter(cfg, staffHandler, patientHandler)

	// 8. Start HTTP Server
	log.Printf("hospital-middleware-api server running on port %s in %s mode\n", cfg.Port, cfg.AppEnv)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Fatal: server terminated unexpectedly: %v\n", err)
	}
}
