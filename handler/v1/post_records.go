package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/observability"
)

// POST /records/{id}
// if the record exists, the record is updated.
// if the record doesn't exist, the record is created.
func (a *API) PostRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		observability.DefaultLogger.Warn("post_records invalid id", "id", id, "error", err)
		_ = writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		return
	}

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		observability.DefaultLogger.Warn("post_records invalid JSON", "id", idNumber, "error", err)
		_ = writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		return
	}

	// first retrieve the record
	record, err := a.records.GetRecord(
		ctx,
		int(idNumber),
	)

	if !errors.Is(err, controller.ErrRecordDoesNotExist) { // record exists
		record, err = a.records.UpdateRecord(ctx, int(idNumber), body)
		if err == nil {
			observability.DefaultLogger.Info("record_updated", "api", "v1", "id", idNumber)
		}
	} else { // record does not exist
		recordMap := map[string]string{}
		for key, value := range body {
			if value != nil {
				recordMap[key] = *value
			}
		}
		record = entity.Record{ID: int(idNumber), Data: recordMap}
		err = a.records.CreateRecord(ctx, record)
		if err == nil {
			observability.DefaultLogger.Info("record_created", "api", "v1", "id", idNumber)
		}
	}

	if err != nil {
		observability.DefaultLogger.Error("post_records failed", "id", idNumber, "error", err)
		_ = writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	_ = writeJSON(w, record, http.StatusOK)
}
