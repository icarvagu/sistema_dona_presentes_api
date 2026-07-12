package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"donapresentes/controllers/config"
	"donapresentes/middleware"
	"donapresentes/models"
	"donapresentes/repositories"
	"donapresentes/services"

	"github.com/gorilla/mux"
)

// artFinalService provides operations for art final tasks, stories, and layout management.
var artFinalService *services.ArtFinalService

// InitArtFinalService initializes the art final service with the required repository.
func InitArtFinalService() {
	artFinalService = services.NewArtFinalService(repositories.NewArtFinalRepository(config.DB), auditService)
}

func artFinalAccess(r *http.Request) models.ArtFinalAccess {
	userID, _, role := middleware.GetUserFromRequest(r)
	admin := role == "admin"
	permissions := middleware.GetPermissionsFromRequest(r)
	hasProfile := func(key string) bool {
		for _, permission := range permissions {
			if permission == key {
				return true
			}
		}
		return false
	}
	artFinal := hasProfile("arte_final")
	manage := admin || artFinal
	production := hasProfile("producao")
	sales := hasProfile("vendas")
	purchases := hasProfile("compras")
	marketing := hasProfile("marketing")
	return models.ArtFinalAccess{
		UserID: userID, Admin: admin, ArtFinal: artFinal, Sales: sales,
		Purchases: purchases, ProductionProfile: production,
		Manage: manage, Marketing: admin || marketing, Layout: manage || sales, Corel: manage || purchases,
		Engraving: manage || production, Media: admin || marketing,
		Production: manage || production || purchases || sales || marketing,
		Stories:    admin || marketing,
	}
}

func canAccessArtFinal(access models.ArtFinalAccess) bool {
	return access.Manage || access.Marketing || access.Layout || access.Corel || access.Engraving || access.Production || access.Stories
}

func parseArtFinalDate(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetArtFinalDashboard handles GET /art-final/dashboard — returns dashboard data filtered by category, status, and date range.
func GetArtFinalDashboard(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !canAccessArtFinal(access) {
		workflowForbidden(w)
		return
	}
	from, err := parseArtFinalDate(r.URL.Query().Get("date_from"))
	if err != nil {
		workflowError(w, err)
		return
	}
	to, err := parseArtFinalDate(r.URL.Query().Get("date_to"))
	if err != nil {
		workflowError(w, err)
		return
	}
	data, err := artFinalService.Dashboard(access, models.ArtFinalFilters{Category: r.URL.Query().Get("category"), Status: r.URL.Query().Get("status"), Search: r.URL.Query().Get("search"), DateFrom: from, DateTo: to})
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}

// CreateArtFinalTask handles POST /art-final/tasks — creates a new art final task.
func CreateArtFinalTask(w http.ResponseWriter, r *http.Request) {
	var input models.ArtFinalTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	access := artFinalAccess(r)
	if (input.Panel == "media" && !access.Marketing && !access.Admin) || (input.Panel != "media" && !access.Manage) {
		workflowForbidden(w)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	id, err := artFinalService.CreateTask(input, user)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
// UpdateArtFinalTask handles PUT /art-final/tasks/{id} — updates an existing art final task.
func UpdateArtFinalTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.ArtFinalTaskInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	access := artFinalAccess(r)
	if (input.Panel == "media" && !access.Marketing && !access.Admin) || (input.Panel != "media" && !access.Manage) {
		workflowForbidden(w)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.UpdateTask(id, input, user, access); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// CreateArtFinalStory handles POST /art-final/stories — creates a new art final story.
func CreateArtFinalStory(w http.ResponseWriter, r *http.Request) {
	if !artFinalAccess(r).Stories {
		workflowForbidden(w)
		return
	}
	var input models.ArtFinalStoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	id, err := artFinalService.CreateStory(input, user)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
// CheckArtFinalStory handles PATCH /art-final/stories/{id}/check — marks a story as checked or unchecked.
func CheckArtFinalStory(w http.ResponseWriter, r *http.Request) {
	if !artFinalAccess(r).Stories {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input struct {
		Checked bool `json:"checked"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.CheckStory(id, input.Checked, user); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// DeleteArtFinalStory handles DELETE /art-final/stories/{id} — deletes an art final story.
func DeleteArtFinalStory(w http.ResponseWriter, r *http.Request) {
	if !artFinalAccess(r).Stories {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.DeleteStory(id, user); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListLayoutRequests handles GET /art-final/layout/requests — lists all layout requests for the current user.
func ListLayoutRequests(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !access.Layout && !access.Corel && !access.Engraving {
		workflowForbidden(w)
		return
	}
	data, err := artFinalService.ListLayoutRequests(access)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
// GetLayoutRequest handles GET /art-final/layout/requests/{id} — returns a single layout request by ID.
func GetLayoutRequest(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !access.Layout && !access.Corel && !access.Engraving {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	data, err := artFinalService.GetLayoutRequest(id, access)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
// CreateLayoutRequest handles POST /art-final/layout/requests — creates a new layout request.
func CreateLayoutRequest(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !access.Layout {
		workflowForbidden(w)
		return
	}
	var input models.LayoutRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	id, err := artFinalService.CreateLayoutRequest(input, user, access)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
// TransitionLayoutRequest handles PATCH /art-final/layout/requests/{id}/transition — transitions a layout request to a new status.
func TransitionLayoutRequest(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !access.Manage {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.LayoutTransitionInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.TransitionLayoutRequest(id, input.Status, input.Note, user); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// AddLayoutMessage handles POST /art-final/layout/requests/{id}/messages — adds a message to a layout request.
func AddLayoutMessage(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !access.Layout && !access.Corel && !access.Engraving {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.LayoutMessageInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.AddLayoutMessage(id, input, user, access); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// AddLayoutVersion handles POST /art-final/layout/items/{itemId}/versions — adds a new version to a layout item.
func AddLayoutVersion(w http.ResponseWriter, r *http.Request) {
	if !artFinalAccess(r).Manage {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["itemId"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.LayoutVersionInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	versionID, err := artFinalService.AddLayoutVersion(id, input, user)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": versionID})
}
// DecideLayoutVersion handles PATCH /art-final/layout/versions/{versionId}/decision — approves or rejects a layout version.
func DecideLayoutVersion(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !access.Layout {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["versionId"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.LayoutDecisionInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.DecideLayoutVersion(id, input, user, access); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// UpsertLayoutJob handles PUT /art-final/layout/items/{itemId}/jobs/{kind} — creates or updates a layout job (corel or engraving).
func UpsertLayoutJob(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	kind := mux.Vars(r)["kind"]
	if kind == "corel" && !access.Corel || kind == "engraving" && !access.Engraving {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["itemId"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.LayoutJobInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.UpsertLayoutJob(kind, id, input, user, access); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// ConfirmLayoutProduct handles PATCH /art-final/layout/items/{itemId}/product-received — confirms whether a product was received.
func ConfirmLayoutProduct(w http.ResponseWriter, r *http.Request) {
	access := artFinalAccess(r)
	if !access.Admin && !access.ProductionProfile && !access.Purchases {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["itemId"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input struct {
		Received bool `json:"received"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.ConfirmProductReceived(id, input.Received, user, access); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// UpdateStoryLifecycle handles PATCH /art-final/stories/{id}/lifecycle — updates the lifecycle stage of a story.
func UpdateStoryLifecycle(w http.ResponseWriter, r *http.Request) {
	if !artFinalAccess(r).Stories {
		workflowForbidden(w)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.StoryLifecycleInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = artFinalService.UpdateStoryLifecycle(id, input, user); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
