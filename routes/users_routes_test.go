package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestRegisterUsersRoutes_PutProfileHasPriorityOverID(t *testing.T) {
	router := mux.NewRouter()
	RegisterUsersRoutes(router)

	req := httptest.NewRequest("PUT", "/users/profile", nil)
	match := &mux.RouteMatch{}

	if !router.Match(req, match) {
		t.Fatalf("expected route match for PUT /users/profile")
	}

	pathTemplate, err := match.Route.GetPathTemplate()
	if err != nil {
		t.Fatalf("failed to get path template: %v", err)
	}

	if pathTemplate != "/users/profile" {
		t.Fatalf("expected /users/profile route, got %s", pathTemplate)
	}
}
