package pathfork

import (
	"fmt"
	"net/http"
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

type SectionViewHandler pathforkFrontEndHandler

func (h SectionViewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	cvi := crudViewInput{
		GetByIdFunc:     models.GetSectionById,
		GetViewPageFunc: pages.GetSectionViewPage,
		TemplateName:    "section_view",
	}
	HandleCrudView(r, w, h.db, h.tr, manager, cvi)
}

func (h SectionViewHandler) Methods() []string {
	return h.methods
}

func BuildSectionViewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SectionViewHandler{
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

func handleSectionForm(section *models.Section, r *http.Request, page pages.WebPage, manager sessionManager.SessionManager) {
	section.Title = r.FormValue("title")
	section.Blurb = r.FormValue("blurb")
	section.Body = r.FormValue("body")
	section.Snippet = false
	if r.FormValue("snippet") == "on" {
		section.Snippet = true
	}
	section.UserEmail = manager.GetUserEmail()
}

type SectionEditHandler pathforkFrontEndHandler

func (h SectionEditHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	params := crudEditInput{
		GetByIdFunc:     models.GetSectionById,
		GetEditPageFunc: pages.GetSectionEditPage,
		TemplateName:    "section_edit",
		UpdateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager, obj db.Updatable) (db.Insertable, error) {
			section := obj.(*models.Section)
			handleSectionForm(section, r, page, manager)
			charsToInsert, charsToDelete, err := forms.GetRelationUpdateIds(
				r, "currentCharIds", "characters",
			)
			settingsToInsert, settingsToDelete, err := forms.GetRelationUpdateIds(
				r, "currentSettingIds", "settings",
			)
			tx, err := h.db.DB.Begin()
			if err == nil {
				if err := section.Save(tx); err != nil {
					fmt.Printf("Error saving section on SectionEditHandler: %v", err.Error())
					return nil, err
				}
				if err := models.UpdateSectionsCharsRelations(
					h.db, tx, section.Id, charsToInsert, charsToDelete,
				); err != nil {
					glog.Errorf("Problem saving section relations: %v", err.Error())
					return nil, err
				}
				allChars, _ := utils.StringsToInts(r.Form["characters"])
				if err := models.UpdateWorksCharsNoConflict(
					h.db, tx, section.WorkId, allChars,
				); err != nil {
					glog.Errorf("Problem saving section relations: %v", err.Error())
					return nil, err
				}
				if err := models.UpdateSectionsSettingsRelations(
					h.db, tx, section.Id, settingsToInsert, settingsToDelete,
				); err != nil {
					glog.Errorf("Problem saving section relations: %v", err.Error())
					return nil, err
				}
				allSettings, _ := utils.StringsToInts(r.Form["settings"])
				if err := models.UpdateWorksSettingsNoConflict(
					h.db, tx, section.WorkId, allSettings,
				); err != nil {
					glog.Errorf("Problem saving section relations: %v", err.Error())
					return nil, err
				}
				tx.Commit()
				return section, nil
			}
			return nil, err
		},
	}
	response := HandleCrudEdit(r, w, h.db, h.tr, manager, params)
	if r.Method == "POST" && response.Error == nil {
		section := response.Obj.(*models.Section)
		http.Redirect(w, r, URLFor("section_view")+fmt.Sprintf("%v", section.Id), http.StatusFound)
	}
}

func (h SectionEditHandler) Methods() []string {
	return h.methods
}

func BuildSectionEditHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SectionEditHandler{
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

type SectionNewHandler pathforkFrontEndHandler

func (h SectionNewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	query := r.URL.Query()
	workIdQ, ok := query["workId"]
	if !ok {
		manager.AddFlash("Sorry, that link must have been bad.")
		http.Redirect(w, r, URLFor("dashboard"), http.StatusFound)
		return
	}
	workId := workIdQ[0]
	params := crudCreateInput{
		GetCreatePageFunc: pages.GetSectionNewPage,
		CreateFuncArgs:    []string{workId},
		TemplateName:      "section_edit",
		SuccessRedirect:   URLFor("work_view") + workId,
		CreateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager) (db.Insertable, error) {
			newSection := &models.Section{}
			handleSectionForm(newSection, r, page, manager)
			workIdInt, _ := strconv.Atoi(workId)
			newSection.WorkId = workIdInt
			newSection.Order = 10000
			tx, err := h.db.DB.Begin()
			if err != nil {
				glog.Error(err.Error())
				return nil, err
			}
			id, err := h.db.Insert(newSection, tx)
			if err != nil {
				glog.Error(err.Error())
				return nil, err
			}
			newSection.Id = id
			charIdsToInsert, _ := utils.StringsToInts(r.Form["characters"])
			err = models.UpdateSectionsCharsRelations(h.db, tx, id, charIdsToInsert, []int{})
			err = models.UpdateWorksCharsNoConflict(h.db, tx, newSection.WorkId, charIdsToInsert)
			settingIdsToInsert, _ := utils.StringsToInts(r.Form["settings"])
			err = models.UpdateSectionsSettingsRelations(h.db, tx, id, settingIdsToInsert, []int{})
			err = models.UpdateWorksSettingsNoConflict(h.db, tx, newSection.WorkId, settingIdsToInsert)
			tx.Commit()
			return newSection, err
		},
	}
	HandleCrudCreate(r, w, h.db, h.tr, manager, params)
}

func (h SectionNewHandler) Methods() []string {
	return h.methods
}

func BuildSectionNewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SectionNewHandler{
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

type SectionDeleteHandler pathforkFrontEndHandler

func (h SectionDeleteHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	response := getCrudStarterResponse(r, w, h.db, manager, models.GetSectionById)
	if response.RedirectCode != 0 {
		if response.FlashMsg != "" {
			manager.AddFlash(response.FlashMsg)
		}
		http.Redirect(w, r, URLFor("dashboard"), response.RedirectCode)
		return
	}
	section := response.Obj.(*models.Section)
	form := forms.NewDeleteForm(section.Id)
	form.Populate(r)
	if r.Method == "POST" && form.Validate() {
		idToDelete, _ := strconv.Atoi(r.FormValue("object_id"))
		success, err := models.DeleteSection(idToDelete, h.db)
		if err != nil || !success {
			glog.Error(err)
			http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("section_view"), section.Id), 301)
			return
		}
	}
	manager.AddFlash("Alright, I got rid of that section for you.")
	http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("work_view"), section.WorkId), 301)
}

func (h SectionDeleteHandler) Methods() []string {
	return h.methods
}

func BuildSectionDeleteHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SectionDeleteHandler{
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

type SectionReorderHandler pathforkFrontEndHandler

func (h SectionReorderHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
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
	page := pages.GetSectionReorderPage(manager, h.db, work)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err == nil && r.FormValue("section-order") != "" {
			tx, err := h.db.DB.Begin()
			if err == nil {
				err = models.ReorderSectionsFromFormValue(r.FormValue("section-order"), work.Id, tx)
				if err == nil {
					manager.AddFlash("OK, that's been reordered.")
					tx.Commit()
					http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("work_view"), work.Id), 302)
					return
				}
			}
		}
		manager.AddFlash("Sorry, something went wrong :(")
	}
	err := h.tr.RenderPage(w, "section_reorder", page)
	if err != nil {
		glog.Error(err.Error())
		http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("work_view"), work.Id), response.RedirectCode)
		return
	}
}

func (h SectionReorderHandler) Methods() []string {
	return h.methods
}

func BuildSectionReorderHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return SectionReorderHandler{
		tr:           tr,
		methods:      []string{"GET", "POST"},
		db:           db,
		sessionStore: store,
	}
}
