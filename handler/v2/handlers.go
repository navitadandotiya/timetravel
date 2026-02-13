package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/observability"
)

// API wraps the v2 SQLite controller
type API struct {
	Controller *controller.SQLiteRecordController
}

// NewAPI initializes the v2 API with the SQLite controller
func NewAPI(c *controller.SQLiteRecordController) *API {
	return &API{Controller: c}
}

// CreateRoutes registers v2 endpoints
func (api *API) CreateRoutes(router *mux.Router) {
	router.HandleFunc("/records/{policyholder_id}", api.PostRecord).Methods("POST")
	router.HandleFunc("/records/{policyholder_id}", api.GetRecord).Methods("GET")
	router.HandleFunc("/health", api.HealthCheck).Methods("POST")
}

// POST /api/v2/records/{policyholder_id}
// Creates or updates a policyholder record
func (api *API) PostRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pidStr := vars["policyholder_id"]
	policyholderID, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || policyholderID <= 0 {
		observability.DefaultLogger.Warn("post_record invalid policyholder_id", "policyholder_id", pidStr, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid policyholder_id"})
		return
	}

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		observability.DefaultLogger.Warn("post_record invalid JSON", "policyholder_id", policyholderID, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON payload"})
		return
	}

	ctx := r.Context()
	record, err := api.Controller.GetRecord(ctx, int(policyholderID))
	if err != nil {
		if err == controller.ErrRecordDoesNotExist {
			// Create new record
			record = entity.PolicyholderRecord{ID: policyholderID, Data: data}
			if err := api.Controller.CreateRecord(ctx, record); err != nil {
				observability.DefaultLogger.Error("post_record create failed", "policyholder_id", policyholderID, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			observability.DefaultLogger.Info("record_created", "api", "v2", "policyholder_id", policyholderID)
			// CreatedAt/UpdatedAt set by service; re-fetch or use zero time for created
			record, _ = api.Controller.GetRecord(ctx, int(policyholderID))
		} else {
			observability.DefaultLogger.Warn("post_record get failed", "policyholder_id", policyholderID, "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	} else {
		// Update existing: convert map[string]string to map[string]*string
		updates := make(map[string]*string, len(data))
		for k, v := range data {
			s := v
			updates[k] = &s
		}
		record, err = api.Controller.UpdateRecord(ctx, int(policyholderID), updates)
		if err != nil {
			observability.DefaultLogger.Error("post_record update failed", "policyholder_id", policyholderID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		observability.DefaultLogger.Info("record_updated", "api", "v2", "policyholder_id", policyholderID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"policyholder_id": policyholderID,
		"record_id":       record.ID,
		"data":            record.Data,
		"updated_at":      record.UpdatedAt.Format(time.RFC3339),
	})
}

// GET /api/v2/records/{policyholder_id}
// Retrieves a policyholder record
func (api *API) GetRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pidStr := vars["policyholder_id"]
	policyholderID, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || policyholderID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid policyholder_id"})
		return
	}

	record, err := api.Controller.GetRecord(r.Context(), int(policyholderID))
	if err != nil {
		observability.DefaultLogger.Warn("get_record failed", "policyholder_id", policyholderID, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	observability.DefaultLogger.Info("record_fetched", "api", "v2", "policyholder_id", policyholderID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"policyholder_id": policyholderID,
		"record_id":       record.ID,
		"data":            record.Data,
		"created_at":      record.CreatedAt.Format(time.RFC3339),
		"updated_at":      record.UpdatedAt.Format(time.RFC3339),
	})
}

// POST /api/v2/health
func (api *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
