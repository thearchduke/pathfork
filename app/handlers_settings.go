package pathfork

import (
	"fmt"
	"net/http"
	"path"
	"strconv"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/pages"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

type SettingViewHandler pathforkFrontEndHandler

func (h SettingViewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	cvi := crudViewInput{
		GetByIdFunc:     models.GetSettingById,
		GetViewPageFunc: pages.GetSettingViewPage,
		TemplateName:    "setting_view",
	}
	HandleCrudView(r, w, h.db, h.tr, manager, cvi)
}

func (h SettingViewHandler) Methods() []string {
	return h.methods
}

func BuildSettingViewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SettingViewHandler{
		tr:           tr,
		methods:      []string{"GET"},
		db:           db,
		sessionStore: store,
	}
}

/*
.
.
*/

type SettingEditHandler pathforkFrontEndHandler

func (h SettingEditHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	params := crudEditInput{
		GetByIdFunc:     models.GetSettingById,
		GetEditPageFunc: pages.GetSettingEditPage,
		TemplateName:    "setting_edit",
		SuccessRedirect: func() string {
			id := path.Base(r.URL.Path)
			return URLFor("setting_view") + id
		}(),
		UpdateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager, obj db.Updatable) (db.Insertable, error) {
			setting := obj.(*models.Setting)
			setting.Name = r.FormValue("name")
			setting.Blurb = r.FormValue("blurb")
			setting.Body = r.FormValue("body")
			tx, err := h.db.DB.Begin()
			if err == nil {
				if err := setting.Save(tx); err != nil {
					glog.Errorf("Error saving setting on edit handler: %v", err.Error())
					return nil, err
				}
				tx.Commit()
				return setting, nil
			}
			return nil, err
		},
	}
	HandleCrudEdit(r, w, h.db, h.tr, manager, params)
}

func (h SettingEditHandler) Methods() []string {
	return h.methods
}

func BuildSettingEditHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SettingEditHandler{
		tr:           tr,
		methods:      []string{"GET", "POST"},
		db:           db,
		sessionStore: store,
	}
}

/*
.
.
*/

type SettingNewHandler pathforkFrontEndHandler

func (h SettingNewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	workId := getWorkId(r)
	manager := sessionManager.New(r, w, h.sessionStore)
	params := crudCreateInput{
		GetCreatePageFunc: pages.GetSettingNewPage,
		CreateFuncArgs:    []string{workId},
		TemplateName:      "setting_edit",
		SuccessRedirect: func() string {
			if workId == "0" {
				return URLFor("setting_index")
			} else {
				return URLFor("work_view") + workId
			}
		}(),
		CreateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager) (db.Insertable, error) {
			newSetting := &models.Setting{}
			newSetting.Name = r.FormValue("name")
			newSetting.Blurb = r.FormValue("blurb")
			newSetting.Body = r.FormValue("body")
			newSetting.UserEmail = manager.GetUserEmail()
			workId, _ := strconv.Atoi(workId)
			tx, err := h.db.DB.Begin()
			if err == nil {
				newId, err := h.db.Insert(newSetting, tx)
				if workId != 0 {
					err = models.UpdateWorksSettingsRelations(h.db, tx, workId, []int{newId}, []int{})
					if err != nil {
						return nil, err
					}
				}
				if err == nil {
					tx.Commit()
					return newSetting, nil
				}
				glog.Error(err.Error())
			}
			glog.Error(err.Error())
			return nil, err
		},
	}
	HandleCrudCreate(r, w, h.db, h.tr, manager, params)
}

func (h SettingNewHandler) Methods() []string {
	return h.methods
}

func BuildSettingNewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SettingNewHandler{
		tr:           tr,
		methods:      []string{"GET", "POST"},
		db:           db,
		sessionStore: store,
	}
}

/*
.
.
*/

type SettingIndexHandler pathforkFrontEndHandler

func (h SettingIndexHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	settings := models.GetSettingsForUser(manager.GetUserEmail(), h.db)
	page := pages.GetSettingIndexPage(manager, settings)
	if err := h.tr.RenderPage(w, "setting_index", page); err != nil {
		glog.Errorf("Error with SettingsIndex page render: %v", err.Error())
		http.Redirect(w, r, URLFor("dashboard"), 302)
	}
}

func (h SettingIndexHandler) Methods() []string {
	return h.methods
}

func BuildSettingIndexHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SettingIndexHandler{
		tr:           tr,
		methods:      []string{"GET"},
		db:           db,
		sessionStore: store,
	}
}

/*
.
.
*/

type SettingDeleteHandler pathforkFrontEndHandler

func (h SettingDeleteHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	response := getCrudStarterResponse(r, w, h.db, manager, models.GetSettingById)
	if response.RedirectCode != 0 {
		if response.FlashMsg != "" {
			manager.AddFlash(response.FlashMsg)
		}
		http.Redirect(w, r, URLFor("dashboard"), response.RedirectCode)
		return
	}
	setting := response.Obj.(*models.Setting)
	form := forms.NewDeleteForm(setting.Id, manager)
	form.Populate(r)
	if r.Method == "POST" {
		if form.Validate() {
			idToDelete, _ := strconv.Atoi(r.FormValue("object_id"))
			success, err := models.DeleteSetting(idToDelete, h.db)
			if err != nil || !success {
				glog.Error(err)
				http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("setting_view"), setting.Id), 301)
				return
			}
		} else {
			http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("setting_view"), setting.Id), 301)
			return
		}
	}
	manager.AddFlash("That setting's gone. It was a bad neighborhood anyway.")
	http.Redirect(w, r, URLFor("setting_index"), 301)
}

func (h SettingDeleteHandler) Methods() []string {
	return h.methods
}

func BuildSettingDeleteHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SettingDeleteHandler{
		tr:           tr,
		methods:      []string{"POST"},
		db:           db,
		sessionStore: store,
	}
}
