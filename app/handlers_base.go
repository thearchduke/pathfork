package pathfork

import (
	"errors"
	"net/http"

	"path"
	"strconv"

	"bitbucket.org/jtyburke/pathfork/app/auth"
	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/pages"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

type FrontEndHandler interface {
	HandleRequest(w http.ResponseWriter, r *http.Request)
	Methods() []string
}

type FrontEndHandlerBuilder func(*TemplateRenderer, *db.DB, *sessions.CookieStore) FrontEndHandler

// Wrappers are used to encapsulate handlers for later dependency injection
// to avoid global variables, a la https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091
// https://gist.github.com/tsenart/5fc18c659814c078378d
func WrapFrontEndHandler(builder FrontEndHandlerBuilder, tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) http.HandlerFunc {
	handler := builder(tr, db, store)
	return func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("%v from %v to %v", r.Method, r.RemoteAddr, r.URL)
		allowed := false
		methods := handler.Methods()
		for i := range methods {
			if methods[i] == r.Method {
				allowed = true
				break
			}
		}
		if !allowed {
			glog.Warningf("Unallowed request method %v sent from %v to %v", r.Method, r.Referer(), r.URL)
			http.Redirect(w, r, r.Referer(), 302)
		}
		_, public := publicRoutes[r.URL.Path]
		isLoggedIn := auth.IsLoggedIn(r, store)
		if !public && !isLoggedIn {
			manager := sessionManager.New(r, w, store)
			manager.AddFlash("Sorry, you're not allowed to access that.")
			http.Redirect(w, r, URLFor("home"), 302)
			return
		}
		handler.HandleRequest(w, r)
	}
}

type crudStarterResponse struct {
	RedirectCode int
	RedirectStr  string
	FlashMsg     string
	Error        error
	Obj          interface{}
}

func getCrudStarterResponse(r *http.Request, w http.ResponseWriter, db *db.DB, manager sessionManager.SessionManager, getByIdFunc func(int, *db.DB) models.Verifiable) crudStarterResponse {
	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		glog.Errorf("Bad object ID: %v", path.Base(r.URL.Path))
		return crudStarterResponse{
			RedirectCode: http.StatusNotFound,
			Error:        err,
			FlashMsg:     "Something went terribly wrong with this request, sorry.",
		}
	}
	verifiable := getByIdFunc(id, db)
	if verifiable == nil {
		return crudStarterResponse{
			RedirectCode: http.StatusNotFound,
			FlashMsg:     "Sorry, we couldn't find that.",
		}
	}
	if !verifiable.VerifyPermission(manager) {
		return crudStarterResponse{
			RedirectCode: http.StatusForbidden,
			FlashMsg:     "Sorry, you're not allowed to access that.",
		}
	}
	return crudStarterResponse{
		Obj: verifiable,
	}
}

type crudViewInput struct {
	GetByIdFunc     func(int, *db.DB) models.Verifiable
	GetViewPageFunc func(sessionManager.SessionManager, interface{}) pages.WebPage
	TemplateName    string
}

func HandleCrudView(
	r *http.Request, w http.ResponseWriter, db *db.DB, tr *TemplateRenderer,
	manager sessionManager.SessionManager, cvi crudViewInput) interface{} {
	response := getCrudStarterResponse(r, w, db, manager, cvi.GetByIdFunc)
	if response.RedirectCode != 0 {
		if response.FlashMsg != "" {
			manager.AddFlash(response.FlashMsg)
		}
		http.Redirect(w, r, URLFor("dashboard"), http.StatusFound)
		return nil
	}
	obj := response.Obj
	page := cvi.GetViewPageFunc(manager, obj)
	if err := tr.RenderPage(w, cvi.TemplateName, page); err != nil {
		glog.Errorf("Error with HandleCrudView page render: %v", err.Error())
		manager.AddFlash("Looks like something went wrong with our server. Sorry.")
		http.Redirect(w, r, URLFor("dashboard"), http.StatusFound)
		return nil
	}
	return obj
}

type crudCreateInput struct {
	GetCreatePageFunc func(sessionManager.SessionManager, *db.DB, ...string) pages.WebPage
	CreateFuncArgs    []string
	TemplateName      string
	SuccessRedirect   string
	CreateObjFunc     func(*http.Request, pages.WebPage, sessionManager.SessionManager) (db.Insertable, error)
}

type crudCreateOutput struct {
	NewObj db.Insertable
	Error  error
}

func HandleCrudCreate(r *http.Request, w http.ResponseWriter, db *db.DB, tr *TemplateRenderer,
	manager sessionManager.SessionManager, cci crudCreateInput) (output crudCreateOutput) {
	page := cci.GetCreatePageFunc(manager, db, cci.CreateFuncArgs...)
	if r.Method == "POST" {
		page.Form.Populate(r)
		if page.Form.Validate() {
			newObj, err := cci.CreateObjFunc(r, page, manager)
			if err != nil {
				glog.Errorf("Error inserting object related to template %v", cci.TemplateName)
				output.Error = err
			} else {
				if cci.SuccessRedirect != "" {
					http.Redirect(w, r, cci.SuccessRedirect, http.StatusFound)
				}
				output.NewObj = newObj
				return
			}
		}
	}
	if err := tr.RenderPage(w, cci.TemplateName, page); err != nil {
		glog.Errorf("Error with %v page render: %v", cci.TemplateName, err.Error())
		manager.AddFlash("Something went wrong with the server. Sorry.")
		http.Redirect(w, r, URLFor("dashboard"), http.StatusFound)
		output.Error = err
	}
	return
}

type crudEditInput struct {
	GetByIdFunc     func(int, *db.DB) models.Verifiable
	GetEditPageFunc func(sessionManager.SessionManager, *db.DB, interface{}) pages.WebPage
	TemplateName    string
	SuccessRedirect string
	UpdateObjFunc   func(*http.Request, pages.WebPage, sessionManager.SessionManager, db.Updatable) (db.Insertable, error)
}

type crudEditOutput struct {
	Obj   db.Insertable
	Error error
}

func HandleCrudEdit(r *http.Request, w http.ResponseWriter, database *db.DB, tr *TemplateRenderer,
	manager sessionManager.SessionManager, input crudEditInput) (output crudEditOutput) {
	response := getCrudStarterResponse(r, w, database, manager, input.GetByIdFunc)
	if response.RedirectCode != 0 {
		http.Redirect(w, r, URLFor("dashboard"), response.RedirectCode)
		return
	}
	obj := response.Obj
	page := input.GetEditPageFunc(manager, database, obj)
	if r.Method == "POST" {
		page.Form.Populate(r)
		if page.Form.Validate() {
			objAsUpdate := obj.(db.Updatable)
			obj, err := input.UpdateObjFunc(r, page, manager, objAsUpdate)
			if err != nil {
				glog.Errorf("Error updating object related to template %v", input.TemplateName)
				output.Error = err
			} else {
				if input.SuccessRedirect != "" {
					http.Redirect(w, r, input.SuccessRedirect, http.StatusFound)
				}
				output.Obj = obj
				return
			}
		} else {
			output.Error = errors.New("Form error")
		}
	}
	if err := tr.RenderPage(w, input.TemplateName, page); err != nil {
		glog.Errorf("Error with %v page render: %v", input.TemplateName, err.Error())
		manager.AddFlash("Something went wrong with the server. Sorry.")
		http.Redirect(w, r, URLFor("dashboard"), http.StatusFound)
		output.Error = err
	}
	return
}

func getWorkId(r *http.Request) string {
	query := r.URL.Query()
	workId := "0"
	workIdQ, ok := query["workId"]
	if ok && len(workIdQ) > 0 {
		workId = workIdQ[0]
	}
	return workId
}
