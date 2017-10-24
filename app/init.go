package pathfork

import (
	"bitbucket.org/jtyburke/pathfork/app/config"
	"bitbucket.org/jtyburke/pathfork/app/db"

	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

func InitApp() (*TemplateRenderer, *db.DB, *sessions.CookieStore) {
	glog.Info("Caching templates")
	tr := NewTemplateRenderer()
	glog.Info("Loading routes")
	InitRoutes()
	glog.Info("Opening database connection")
	db := db.New()
	db.Open(config.PostgresUrl)
	store := sessions.NewCookieStore([]byte(config.SessionSecretKey))
	return tr, db, store
}
