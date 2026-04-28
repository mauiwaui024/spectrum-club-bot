package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetUserIDFromRequest_LegacyFallbackDisabled(t *testing.T) {
	h := &Handler{
		allowLegacyUserID: false,
		sessions:          map[string]browserSession{},
	}

	req := httptest.NewRequest("GET", "/api/calendar?user_id=42", nil)
	if _, err := h.getUserIDFromRequest(req); err == nil {
		t.Fatalf("expected auth error when legacy fallback is disabled")
	}
}

func TestGetUserIDFromRequest_LegacyFallbackEnabled(t *testing.T) {
	h := &Handler{
		allowLegacyUserID: true,
		sessions:          map[string]browserSession{},
	}

	req := httptest.NewRequest("GET", "/api/calendar?user_id=42", nil)
	userID, err := h.getUserIDFromRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected userID 42, got %d", userID)
	}
}

func TestGetUserIDFromRequest_SessionAuth(t *testing.T) {
	h := &Handler{
		allowLegacyUserID: false,
		sessions: map[string]browserSession{
			"test-token": {
				UserID:    77,
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/calendar", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "test-token",
	})

	userID, err := h.getUserIDFromRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 77 {
		t.Fatalf("expected userID 77, got %d", userID)
	}
}
