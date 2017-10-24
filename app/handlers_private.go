package pathfork

import (
	"fmt"
	"net/http"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/pages"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/gorilla/sessions"
)

type DashboardHandler pathforkFrontEndHandler

func (h DashboardHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	works := models.GetWorksForUser(manager.GetUserEmail(), h.db)
	page := pages.GetDashboardPage(manager, works)
	if err := h.tr.RenderPage(w, "dashboard", page); err != nil {
		fmt.Printf("Error with DashboardHandler page render: %v", err.Error())
		// flash error
		http.Redirect(w, r, URLFor("home"), 302)
	}
}

func (h DashboardHandler) Methods() []string {
	return h.methods
}

func BuildDashboardHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return DashboardHandler{
		tr:           tr,
		methods:      []string{"GET"},
		db:           db,
		sessionStore: store,
	}
}
