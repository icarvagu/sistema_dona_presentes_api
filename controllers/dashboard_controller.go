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

// dashboardService provides dashboard data aggregation for sellers.
var dashboardService *services.DashboardService

// InitDashboardService initializes the dashboard service from a pre-created service instance.
func InitDashboardService(service *services.DashboardService) {
	dashboardService = service
}

type DashboardResponse = services.DashboardResponse

// GetSellerDashboard handles GET /dashboard/seller/{id} — returns aggregated dashboard data for a specific seller.
func GetSellerDashboard(w http.ResponseWriter, r *http.Request) {
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
