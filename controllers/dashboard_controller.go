package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	apperrors "donapresentes/errors"
	"donapresentes/middleware"
	"donapresentes/services"

	"github.com/gorilla/mux"
)

var dashboardService *services.DashboardService

func InitDashboardService(service *services.DashboardService) {
	dashboardService = service
}

type DashboardResponse = services.DashboardResponse

func GetVendedorDashboard(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	dashboard, err := dashboardService.GetDashboardData(id)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}
