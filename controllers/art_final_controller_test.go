package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func artFinalAccessRequest(userID int, role string, permissions []string) *http.Request {
	r := httptest.NewRequest("GET", "/art-final/dashboard", nil)
	ctx := context.WithValue(r.Context(), "user_id", userID)
	ctx = context.WithValue(ctx, "role", role)
	ctx = context.WithValue(ctx, "permissions", permissions)
	return r.WithContext(ctx)
}

func TestArtFinalAccessIsExplicitAndScoped(t *testing.T) {
	sales := artFinalAccess(artFinalAccessRequest(42, "standard", []string{"vendas"}))
	if sales.UserID != 42 || !sales.Sales || sales.Manage || sales.Corel || sales.Media {
		t.Fatalf("unexpected sales access: %+v", sales)
	}
	production := artFinalAccess(artFinalAccessRequest(7, "standard", []string{"producao"}))
	if !production.ProductionProfile || !production.Production || production.Stories || !production.Engraving {
		t.Fatalf("unexpected production access: %+v", production)
	}
	admin := artFinalAccess(artFinalAccessRequest(1, "admin", nil))
	if !admin.Admin || !admin.Manage || !admin.Layout || !admin.Corel || !admin.Media || !admin.Production || !admin.Stories {
		t.Fatalf("unexpected admin access: %+v", admin)
	}
}

func TestMarketingIsSeparatedFromArtFinal(t *testing.T) {
	marketing := artFinalAccess(artFinalAccessRequest(8, "standard", []string{"marketing"}))
	if !marketing.Marketing || !marketing.Media || !marketing.Stories || marketing.Manage || marketing.Layout {
		t.Fatalf("unexpected marketing access: %+v", marketing)
	}
	art := artFinalAccess(artFinalAccessRequest(9, "standard", []string{"arte_final"}))
	if !art.Manage || art.Media || art.Stories || art.Marketing {
		t.Fatalf("marketing leaked into art-final: %+v", art)
	}
}
