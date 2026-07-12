package routes

import (
	"donapresentes/controllers"
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterArtFinalRoutes registers all art final (design/layout approval workflow) HTTP endpoints on the given router.
// Routes: GET /art-final/dashboard, POST /art-final/tasks, PUT /art-final/tasks/{id},
// POST /art-final/stories, PATCH /art-final/stories/{id}/check, DELETE /art-final/stories/{id},
// GET /art-final/layout/requests, POST /art-final/layout/requests,
// GET /art-final/layout/requests/{id}, PATCH /art-final/layout/requests/{id}/transition,
// POST /art-final/layout/requests/{id}/messages, POST /art-final/layout/items/{itemId}/versions,
// PATCH /art-final/layout/versions/{versionId}/decision,
// PUT /art-final/layout/items/{itemId}/jobs/{kind},
// PATCH /art-final/layout/items/{itemId}/product-received,
// PATCH /art-final/stories/{id}/lifecycle.
func RegisterArtFinalRoutes(r *mux.Router) {
	r.HandleFunc("/art-final/dashboard", controllers.GetArtFinalDashboard).Methods(http.MethodGet)
	r.HandleFunc("/art-final/tasks", controllers.CreateArtFinalTask).Methods(http.MethodPost)
	r.HandleFunc("/art-final/tasks/{id}", controllers.UpdateArtFinalTask).Methods(http.MethodPut)
	r.HandleFunc("/art-final/stories", controllers.CreateArtFinalStory).Methods(http.MethodPost)
	r.HandleFunc("/art-final/stories/{id}/check", controllers.CheckArtFinalStory).Methods(http.MethodPatch)
	r.HandleFunc("/art-final/stories/{id}", controllers.DeleteArtFinalStory).Methods(http.MethodDelete)
	r.HandleFunc("/art-final/layout/requests", controllers.ListLayoutRequests).Methods(http.MethodGet)
	r.HandleFunc("/art-final/layout/requests", controllers.CreateLayoutRequest).Methods(http.MethodPost)
	r.HandleFunc("/art-final/layout/requests/{id}", controllers.GetLayoutRequest).Methods(http.MethodGet)
	r.HandleFunc("/art-final/layout/requests/{id}/transition", controllers.TransitionLayoutRequest).Methods(http.MethodPatch)
	r.HandleFunc("/art-final/layout/requests/{id}/messages", controllers.AddLayoutMessage).Methods(http.MethodPost)
	r.HandleFunc("/art-final/layout/items/{itemId}/versions", controllers.AddLayoutVersion).Methods(http.MethodPost)
	r.HandleFunc("/art-final/layout/versions/{versionId}/decision", controllers.DecideLayoutVersion).Methods(http.MethodPatch)
	r.HandleFunc("/art-final/layout/items/{itemId}/jobs/{kind}", controllers.UpsertLayoutJob).Methods(http.MethodPut)
	r.HandleFunc("/art-final/layout/items/{itemId}/product-received", controllers.ConfirmLayoutProduct).Methods(http.MethodPatch)
	r.HandleFunc("/art-final/stories/{id}/lifecycle", controllers.UpdateStoryLifecycle).Methods(http.MethodPatch)
}
