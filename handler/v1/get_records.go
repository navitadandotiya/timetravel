package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/observability"
)

// GET /records/{id}
// GetRecord retrieves the record.
func (a *API) GetRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		observability.DefaultLogger.Warn("get_records invalid id", "id", id, "error", err)
		_ = writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		return
	}

	record, err := a.records.GetRecord(ctx, int(idNumber))
	if err != nil {
		observability.DefaultLogger.Warn("get_records not found", "id", idNumber, "error", err)
		_ = writeError(w, fmt.Sprintf("record of id %v does not exist", idNumber), http.StatusBadRequest)
		return
	}

	observability.DefaultLogger.Info("record_fetched", "api", "v1", "id", idNumber)
	_ = writeJSON(w, record, http.StatusOK)
}
