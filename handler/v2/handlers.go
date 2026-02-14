package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/observability"
)

// API wraps the SQLite v2 controller/service
type API struct {
	Controller *controller.SQLiteRecordController
}

// NewAPI initializes the v2 API
func NewAPI(ctrl *controller.SQLiteRecordController) *API {
	return &API{Controller: ctrl}
}

// CreateRoutes registers v2 endpoints
func (api *API) CreateRoutes(router *mux.Router) {
	router.HandleFunc("/records/{policyholder_id}", api.UpsertRecord).Methods("POST")
	router.HandleFunc("/records/{policyholder_id}", api.GetRecord).Methods("GET")
	router.HandleFunc("/health", api.HealthCheck).Methods("POST")
	router.HandleFunc("/records/{policyholder_id}/versions", api.ListVersions).Methods("GET")
	router.HandleFunc("/records/{policyholder_id}/versions/{version}", api.GetVersion).Methods("GET")
}

// UpsertRecord creates or updates a record
func (api *API) UpsertRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pidStr := vars["policyholder_id"]
	policyholderID, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || policyholderID <= 0 {
		respondError(w, http.StatusBadRequest, "invalid policyholder_id")
		return
	}

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	ctx := r.Context()
	record, err := api.Controller.UpsertRecord(ctx, policyholderID, data)
	if err != nil {
		observability.DefaultLogger.Error("upsert_record failed", "policyholder_id", policyholderID, "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	observability.DefaultLogger.Info("record_upserted", "policyholder_id", policyholderID, "version", record.Version)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"policyholder_id": policyholderID,
		"record_id":       record.ID,
		"version":         record.Version,
		"data":            record.Data,
		"created_at":      record.CreatedAt.Format(time.RFC3339),
		"updated_at":      record.UpdatedAt.Format(time.RFC3339),
	})
}

// GetRecord retrieves a policyholder record
func (api *API) GetRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pidStr := vars["policyholder_id"]
	policyholderID, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || policyholderID <= 0 {
		respondError(w, http.StatusBadRequest, "invalid policyholder_id")
		return
	}

	record, err := api.Controller.GetRecord(r.Context(), policyholderID)
	if err != nil {
		if err == controller.ErrRecordDoesNotExist {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	observability.DefaultLogger.Info("record_fetched", "policyholder_id", policyholderID, "version", record.Version)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"policyholder_id": policyholderID,
		"record_id":       record.ID,
		"version":         record.Version,
		"data":            record.Data,
		"created_at":      record.CreatedAt.Format(time.RFC3339),
		"updated_at":      record.UpdatedAt.Format(time.RFC3339),
	})
}

// HealthCheck returns simple OK
func (api *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// Helper for JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Helper for error JSON
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func (api *API) GetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["policyholder_id"])
	version, _ := strconv.Atoi(vars["version"])

	data, err := api.Controller.GetVersion(r.Context(), id, version)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"policyholder_id": id,
		"version":         version,
		"data":            data,
	})
}

func (api *API) ListVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["policyholder_id"])

	versions, err := api.Controller.ListVersions(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"policyholder_id": id,
		"versions":        versions,
	})
}
