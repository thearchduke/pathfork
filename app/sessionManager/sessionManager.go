package sessionManager

import (
	"net/http"

	"bitbucket.org/jtyburke/pathfork/app/config"

	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

type SessionManager struct {
	Session *sessions.Session
	r       *http.Request
	w       http.ResponseWriter
}

func (s SessionManager) Save() error {
	return s.Session.Save(s.r, s.w)
}

func (s SessionManager) AddFlash(msg string) error {
	s.Session.AddFlash(msg)
	return s.Save()
}

func (s SessionManager) GetUserEmail() string {
	return s.Session.Values["userEmail"].(string)
}

func (s SessionManager) SetUser(email string) error {
	s.Session.Values["userEmail"] = email
	return s.Save()
}

func (s SessionManager) DeleteUser() error {
	delete(s.Session.Values, "userEmail")
	return s.Save()
}

func (s SessionManager) SetCurrentWork(id int, title string) error {
	s.Session.Values["workId"] = id
	s.Session.Values["workTitle"] = title
	return s.Save()
}

func (s SessionManager) GetCurrentWork() (int, string) {
	id := s.Session.Values["workId"]
	if id == nil {
		id = 0
	}
	title := s.Session.Values["workTitle"]
	if title == nil {
		title = ""
	}
	return id.(int), title.(string)
}

func New(r *http.Request, w http.ResponseWriter, s *sessions.CookieStore) SessionManager {
	session, err := s.Get(r, config.SessionCookieName)
	if err != nil {
		glog.Error(err)
	}
	return SessionManager{
		Session: session,
		r:       r,
		w:       w,
	}
}
