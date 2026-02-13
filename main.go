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
	api "github.com/rainbowmga/timetravel/handler/v1"
	// "github.com/rainbowmga/timetravel/handler/v2" // v2 can be added later
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func main() {
	// --- Step 1: Initialize SQLite DB for durability ---
	dbPath := "./db/timetravel.db"
	db := gateways.ConnectDB(dbPath)
	defer db.Close()
	log.Println("SQLite DB connected successfully!")

	// --- Initialize router ---
	router := mux.NewRouter()

	// --- Health Check ---
	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})

	// --- Initialize in-memory service (can switch to DB-backed later) ---
	service := controller.NewInMemoryRecordService()
	apiHandler := api.NewAPI(&service)
	apiHandler.CreateRoutes(apiRoute)

	// --- Start HTTP Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	address := "127.0.0.1:" + port
	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("listening on %s", address)
	log.Fatal(srv.ListenAndServe())
}
