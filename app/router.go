package app

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/gateways"
	apiV1 "github.com/rainbowmga/timetravel/handler/v1"
	apiV2 "github.com/rainbowmga/timetravel/handler/v2"
	"github.com/rainbowmga/timetravel/observability"
)

func BuildRouter(dbPath string, runMigrations bool) (*mux.Router, error) {
	
	sqlPath := "script/create_v2_tables.sql"

    // Skip SQL file if dbPath is in-memory (for tests)
    if dbPath == ":memory:" || dbPath == "file:testdb?mode=memory&cache=shared" {
        sqlPath = ""
    }

    db := gateways.ConnectDB(dbPath, sqlPath)
    
    if runMigrations && sqlPath != "" {
        migrationsPath := "script/migrations"
        if err := gateways.RunMigrations(db, migrationsPath); err != nil {
            return nil, err
        }
    }

	metricsRepo, err := gateways.NewMetricsRepository(dbPath)
	if err != nil {
    	return nil, err
	}
	observability.InitMetricsRepository(metricsRepo)

	router := mux.NewRouter()
	router.Handle("/metrics", observability.MetricsHandler()).Methods("GET")

	// v1
	v1Route := router.PathPrefix("/api/v1").Subrouter()
	v1Route.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}).Methods("POST")
	
	v1Service := controller.NewInMemoryRecordService()
	v1Handler := apiV1.NewAPI(&v1Service)
	v1Handler.CreateRoutes(v1Route)

	// v2
	v2Route := router.PathPrefix("/api/v2").Subrouter()

	// 1️⃣ v2 health check and metrics(no user context required)
	v2Route.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}).Methods("POST") 
	// 2️⃣ middleware-protected routes
	v2Route.Use(observability.RequireUserContext)
	v2Route.Use(observability.LoggingAndMetrics)

	// Metrics endpoint under v2
	v2Controller, err := controller.NewSQLiteRecordController(dbPath)
	if err != nil {
		return nil, err
	}

	flagService, err := controller.NewFeatureFlagController(dbPath)
	if err != nil {
		return nil, err
	}

	v2Handler := apiV2.NewAPI(v2Controller, flagService)
	v2Handler.CreateRoutes(v2Route)

	return router, nil
}
