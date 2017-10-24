package pathfork

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"github.com/gorilla/sessions"
)

func getTestVars() (*httptest.ResponseRecorder, *TemplateRenderer, *db.DB, *sessions.CookieStore) {
	rr := httptest.NewRecorder()
	tr := NewTemplateRenderer()
	db := db.New()
	store := sessions.NewCookieStore([]byte("whatever"))
	return rr, tr, db, store
}

func getHandlerAndStuff(builder FrontEndHandlerBuilder) (*httptest.ResponseRecorder, http.HandlerFunc) {
	InitRoutes()
	rr, tr, db, store := getTestVars()
	handler := WrapFrontEndHandler(builder, tr, db, store)
	return rr, handler
}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr, handler := getHandlerAndStuff(BuildHomeHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestForbiddenMethod(t *testing.T) {
	req, err := http.NewRequest("PUT", URLFor("home"), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr, handler := getHandlerAndStuff(BuildHomeHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != 302 {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, 302)
	}
}

func TestContactHandler(t *testing.T) {
	if _, err := http.NewRequest("GET", URLFor("contact"), nil); err != nil {
		t.Fatal(err)
	}
}

func TestAboutHandler(t *testing.T) {
	if _, err := http.NewRequest("GET", URLFor("about"), nil); err != nil {
		t.Fatal(err)
	}
}

func TestAuthHandler(t *testing.T) {
	req, err := http.NewRequest("POST", URLFor("home"), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr, handler := getHandlerAndStuff(BuildAuthHandler)
	req.Form = map[string][]string{
		"username": []string{"tynan"},
		"password": []string{"password"},
	}
	handler.ServeHTTP(rr, req) // invalid login
	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusFound)
	}
}
