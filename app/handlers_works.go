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
	"bitbucket.org/jtyburke/pathfork/app/utils"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

type WorkViewHandler pathforkFrontEndHandler

func (h WorkViewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	cvi := crudViewInput{
		GetByIdFunc:     models.GetWorkById,
		GetViewPageFunc: pages.GetWorkViewPage,
		TemplateName:    "work_view",
	}
	HandleCrudView(r, w, h.db, h.tr, manager, cvi)
}

func (h WorkViewHandler) Methods() []string {
	return h.methods
}

func BuildWorkViewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return WorkViewHandler{
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

type WorkEditHandler pathforkFrontEndHandler

func (h WorkEditHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	params := crudEditInput{
		GetByIdFunc:     models.GetWorkById,
		GetEditPageFunc: pages.GetWorkEditPage,
		TemplateName:    "work_edit",
		SuccessRedirect: func() string {
			id := path.Base(r.URL.Path)
			return URLFor("work_view") + id
		}(),
		UpdateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager, obj db.Updatable) (db.Insertable, error) {
			work := obj.(*models.Work)
			work.Title = r.FormValue("title")
			work.Blurb = r.FormValue("blurb")
			charsToInsert, charsToDelete, err := forms.GetRelationUpdateIds(
				r, "currentCharIds", "characters",
			)
			settingsToInsert, settingsToDelete, err := forms.GetRelationUpdateIds(
				r, "currentSettingIds", "settings",
			)
			tx, err := h.db.DB.Begin()
			if err == nil {
				if err := work.Save(tx); err != nil {
					fmt.Printf("Error saving work on WorkEditHandler: %v", err)
					return nil, err
				}
				if err := models.UpdateWorksCharsRelations(
					h.db, tx, work.Id, charsToInsert, charsToDelete,
				); err != nil {
					glog.Errorf("Problem saving work relations: %v", err.Error())
					return nil, err
				}
				if err := models.UpdateWorksSettingsRelations(
					h.db, tx, work.Id, settingsToInsert, settingsToDelete,
				); err != nil {
					glog.Errorf("Problem saving work relations: %v", err.Error())
					return nil, err
				}
				tx.Commit()
				return work, nil
			}
			return nil, err
		},
	}
	HandleCrudEdit(r, w, h.db, h.tr, manager, params)
}

func (h WorkEditHandler) Methods() []string {
	return h.methods
}

func BuildWorkEditHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return WorkEditHandler{
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

type WorkNewHandler pathforkFrontEndHandler

func (h WorkNewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	params := crudCreateInput{
		GetCreatePageFunc: pages.GetWorkNewPage,
		CreateFuncArgs:    nil,
		TemplateName:      "work_edit",
		CreateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager) (db.Insertable, error) {
			newWork := &models.Work{}
			newWork.Title = r.FormValue("title")
			newWork.Blurb = r.FormValue("blurb")
			newWork.UserEmail = manager.GetUserEmail()
			tx, err := h.db.DB.Begin()
			if err != nil {
				glog.Error(err.Error())
				return nil, err
			}
			id, err := h.db.Insert(newWork, tx)
			newWork.Id = id
			charIdsToInsert, _ := utils.StringsToInts(r.Form["characters"])
			err = models.UpdateWorksCharsRelations(h.db, tx, id, charIdsToInsert, []int{})
			settingIdsToInsert, _ := utils.StringsToInts(r.Form["settings"])
			err = models.UpdateWorksSettingsRelations(h.db, tx, id, settingIdsToInsert, []int{})
			tx.Commit()
			return newWork, err
		},
	}
	response := HandleCrudCreate(r, w, h.db, h.tr, manager, params)
	if r.Method == "POST" && response.NewObj != nil {
		newWork := response.NewObj.(*models.Work)
		http.Redirect(w, r, URLFor("work_view")+fmt.Sprintf("%v", newWork.Id), http.StatusFound)
	}
}

func (h WorkNewHandler) Methods() []string {
	return h.methods
}

func BuildWorkNewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return WorkNewHandler{
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

type WorkDeleteHandler pathforkFrontEndHandler

func (h WorkDeleteHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	response := getCrudStarterResponse(r, w, h.db, manager, models.GetWorkById)
	if response.RedirectCode != 0 {
		if response.FlashMsg != "" {
			manager.AddFlash(response.FlashMsg)
		}
		http.Redirect(w, r, URLFor("dashboard"), response.RedirectCode)
		return
	}
	work := response.Obj.(*models.Work)
	form := forms.NewDeleteForm(work.Id, manager)
	form.Populate(r)
	if r.Method == "POST" {
		if form.Validate() {
			idToDelete, _ := strconv.Atoi(r.FormValue("object_id"))
			success, err := models.DeleteWork(idToDelete, h.db)
			if err != nil || !success {
				glog.Error(err)
				http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("work_view"), work.Id), 301)
				return
			}
			manager.UnsetCurrentWork()
		} else {
			http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("work_view"), work.Id), 301)
			return
		}
	}
	manager.AddFlash("That work is no more.")
	http.Redirect(w, r, URLFor("dashboard"), 301)
}

func (h WorkDeleteHandler) Methods() []string {
	return h.methods
}

func BuildWorkDeleteHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return WorkDeleteHandler{
		tr:           tr,
		methods:      []string{"POST"},
		db:           db,
		sessionStore: store,
	}
}

/*
.
.
*/

type WorkExportHandler pathforkFrontEndHandler

func (h WorkExportHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	response := getCrudStarterResponse(r, w, h.db, manager, models.GetWorkById)
	if response.RedirectCode != 0 {
		if response.FlashMsg != "" {
			manager.AddFlash(response.FlashMsg)
		}
		http.Redirect(w, r, URLFor("dashboard"), response.RedirectCode)
		return
	}
	work := response.Obj.(*models.Work)
	sections, snippets := models.GetSectionDetailForExport(work.Id, h.db)
	settings := models.GetSettingsForWorkExport(work.Id, h.db)
	characters := models.GetCharactersForWorkExport(work.Id, h.db)
	err := h.tr.RenderPage(
		w, "work_export", pages.GetWorkExportPage(
			manager, work, sections, snippets, settings, characters,
		),
	)
	if err != nil {
		glog.Error(err.Error())
		manager.AddFlash("Sorry, something went wrong exporting that.")
		http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("work_view"), work.Id), 302)
	}
}

func (h WorkExportHandler) Methods() []string {
	return h.methods
}

func BuildWorkExportHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return WorkExportHandler{
		tr:           tr,
		methods:      []string{"GET"},
		db:           db,
		sessionStore: store,
	}
}
