package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rainbowmga/timetravel/app"
	"github.com/rainbowmga/timetravel/conf"
	"github.com/rainbowmga/timetravel/observability"
)

// RunServer starts the HTTP server and returns an error instead of exiting
func RunServer(configPath string, envPort string) error {
	cfg := conf.LoadConfig(configPath)
	dbPath := cfg.Database.Path

	router, err := app.BuildRouter(dbPath, cfg.Database.Migrations.RunOnStartup)
	if err != nil {
		return err
	}

	port := envPort
	if port == "" {
		port = "8000"
	}

	address := ":" + port

	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	observability.DefaultLogger.Info("server listening", "address", address)
	return srv.ListenAndServe()
}

func main() {
	if err := RunServer("conf/config.yaml", os.Getenv("PORT")); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
