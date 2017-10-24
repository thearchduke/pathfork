package sessionManager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
)

func TestSessionManager(t *testing.T) {
	store := sessions.NewCookieStore([]byte("whatever"))
	r, _ := http.NewRequest("POST", "/auth", nil)
	w := httptest.NewRecorder()
	manager := New(r, w, store)
	if _, ok := manager.Session.Values["userEmail"]; ok {
		t.Error("userId is set at init for some reason")
	}
	manager.SetUser("tynanburke@gmail.com")
	if manager.Session.Values["userEmail"] != "tynanburke@gmail.com" {
		t.Error("SetUser not working")
	}
	manager.DeleteUser()
	if _, ok := manager.Session.Values["userId"]; ok {
		t.Error("logUserOut is failing for some reason")
	}
	id, _ := manager.GetCurrentWork()
	if id != 0 {
		t.Error("GetCurrentWork() should be 0")
	}
	manager.SetCurrentWork(12, "a title")
	id, title := manager.GetCurrentWork()
	if id != 12 {
		t.Error("GetCurrentWork() id should be 12")
	}
	if title != "a title" {
		t.Error("GetCurrentWork() title should be 'a title'")
	}
}
