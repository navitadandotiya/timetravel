package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/gateways"
	apiV1 "github.com/rainbowmga/timetravel/handler/v1"
	apiV2 "github.com/rainbowmga/timetravel/handler/v2"
	"github.com/rainbowmga/timetravel/observability"
	"github.com/rainbowmga/timetravel/conf"
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func main() {
	 // Load config
	 cfg := conf.LoadConfig("conf/config.yaml")

	// --- Step 1: Initialize SQLite DB for durability ---
	dbPath := cfg.Database.Path
	db := gateways.ConnectDB(dbPath)
	defer db.Close()
	observability.DefaultLogger.Info("SQLite DB connected", "path", dbPath)

	// --- Run migrations only if enabled ---
    if cfg.Database.Migrations.RunOnStartup {
        if err := gateways.RunMigrations(db); err != nil {
            log.Fatalf("migration failed: %v", err)
        }
        log.Println("migrations completed!")
    } else {
        log.Println("migrations skipped (run_on_startup=false)")
    }

	// --- Initialize router ---
	router := mux.NewRouter()

	// Observability: metrics endpoint (no logging middleware to avoid noise)
	router.Handle("/metrics", observability.MetricsHandler()).Methods("GET")

	// ----------------------
	// v1: In-memory service
	// ----------------------
	v1Route := router.PathPrefix("/api/v1").Subrouter()

	// v1 health check
	v1Route.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	}).Methods("POST")

	// v1 in-memory service
	v1Service := controller.NewInMemoryRecordService()
	v1Handler := apiV1.NewAPI(&v1Service)
	v1Handler.CreateRoutes(v1Route)

	// ----------------------
	// v2: SQLite persistent service
	// ----------------------
	v2Route := router.PathPrefix("/api/v2").Subrouter()

	// v2 health check
	v2Route.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	}).Methods("POST")

	// v2 SQLite-backed controller
	v2Controller, err := controller.NewSQLiteRecordController(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize v2 controller: %v", err)
	}
	v2Handler := apiV2.NewAPI(v2Controller)
	v2Handler.CreateRoutes(v2Route)

	// ----------------------
	// Start HTTP server
	// ----------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	address := "127.0.0.1:" + port
	// Wrap with logging and metrics middleware
	handler := observability.LoggingAndMetrics(router)

	srv := &http.Server{
		Handler:      handler,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	observability.DefaultLogger.Info("server listening", "address", address)
	log.Fatal(srv.ListenAndServe())
}
