package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserIDFromRequest_LegacyFallbackDisabled(t *testing.T) {
	h := &Handler{
		allowLegacyUserID: false,
		sessionService:    newMockSessionService(),
	}

	req := httptest.NewRequest("GET", "/api/calendar?user_id=42", nil)
	rr := httptest.NewRecorder()
	if _, err := h.getUserIDFromRequest(rr, req); err == nil {
		t.Fatalf("expected auth error when legacy fallback is disabled")
	}
}

func TestGetUserIDFromRequest_LegacyFallbackEnabled(t *testing.T) {
	h := &Handler{
		allowLegacyUserID: true,
		sessionService:    newMockSessionService(),
	}

	req := httptest.NewRequest("GET", "/api/calendar?user_id=42", nil)
	rr := httptest.NewRecorder()
	userID, err := h.getUserIDFromRequest(rr, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected userID 42, got %d", userID)
	}
}

func TestGetUserIDFromRequest_AccessCookie(t *testing.T) {
	h, _, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)

	req := httptest.NewRequest("GET", "/api/calendar", nil)
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()

	userID, err := h.getUserIDFromRequest(rr, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 1 {
		t.Fatalf("expected userID 1, got %d", userID)
	}
}

func TestGetUserIDFromRequest_SilentRefresh(t *testing.T) {
	h, _, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	// Drop access mapping so Resolve falls through to refresh.
	delete(h.sessionService.(*mockSessionService).access, access)

	req := httptest.NewRequest("GET", "/api/calendar", nil)
	addAuthCookies(req, "expired", refresh)
	rr := httptest.NewRecorder()

	userID, err := h.getUserIDFromRequest(rr, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 1 {
		t.Fatalf("expected userID 1, got %d", userID)
	}
	cookies := rr.Result().Cookies()
	var gotAccess, gotRefresh bool
	for _, c := range cookies {
		if c.Name == accessCookieName && c.Value != "" {
			gotAccess = true
		}
		if c.Name == refreshCookieName && c.Value != "" {
			gotRefresh = true
		}
	}
	if !gotAccess || !gotRefresh {
		t.Fatalf("expected rotated cookies, got %#v", cookies)
	}
}

func TestIssueAuthCookiesClearsLegacy(t *testing.T) {
	h, _, _, _ := testHandler(1)
	rr := httptest.NewRecorder()
	if err := h.issueAuthCookies(rr, 1); err != nil {
		t.Fatal(err)
	}
	var clearedLegacy bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == legacySessionName && c.MaxAge < 0 {
			clearedLegacy = true
		}
	}
	if !clearedLegacy {
		t.Fatal("expected legacy session cookie cleared")
	}
}

func TestBrowserLogoutRevokesRefresh(t *testing.T) {
	h, _, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	req := httptest.NewRequest(http.MethodPost, "/api/browser-auth/logout", nil)
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.BrowserLogoutAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := h.sessionService.(*mockSessionService).refresh[refresh]; ok {
		t.Fatal("refresh should be revoked")
	}
}
