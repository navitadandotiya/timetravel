package api

import (
	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/controller"
)

type API struct {
	records controller.RecordService
}

func NewAPI(records controller.RecordService) *API {
	return &API{records}
}

// generates all api routes
func (a *API) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
}
