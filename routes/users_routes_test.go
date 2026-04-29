package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestRegisterUsersRoutes_PutProfileHasPriorityOverID(t *testing.T) {
	router := mux.NewRouter()
	RegisterUsersRoutes(router)

	profileReq := httptest.NewRequest("PUT", "/users/profile", nil)
	profileMatch := &mux.RouteMatch{}
	if !router.Match(profileReq, profileMatch) {
		t.Fatalf("expected route match for PUT /users/profile")
	}

	pathTemplate, err := profileMatch.Route.GetPathTemplate()
	if err != nil {
		t.Fatalf("failed to get path template: %v", err)
	}
	if pathTemplate != "/users/profile" {
		t.Fatalf("expected /users/profile route, got %s", pathTemplate)
	}

	idReq := httptest.NewRequest("PUT", "/users/123", nil)
	idMatch := &mux.RouteMatch{}
	if !router.Match(idReq, idMatch) {
		t.Fatalf("expected route match for PUT /users/123")
	}

	idPathTemplate, err := idMatch.Route.GetPathTemplate()
	if err != nil {
		t.Fatalf("failed to get id path template: %v", err)
	}
	if idPathTemplate != "/users/{id}" {
		t.Fatalf("expected /users/{id} route, got %s", idPathTemplate)
	}
}
