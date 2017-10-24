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

type CharacterViewHandler pathforkFrontEndHandler

func (h CharacterViewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	cvi := crudViewInput{
		GetByIdFunc:     models.GetCharacterDetail,
		GetViewPageFunc: pages.GetCharacterViewPage,
		TemplateName:    "character_view",
	}
	HandleCrudView(r, w, h.db, h.tr, manager, cvi)
}

func (h CharacterViewHandler) Methods() []string {
	return h.methods
}

func BuildCharacterViewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return CharacterViewHandler{
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

type CharacterEditHandler pathforkFrontEndHandler

func (h CharacterEditHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	params := crudEditInput{
		GetByIdFunc:     models.GetCharacterDetail,
		GetEditPageFunc: pages.GetCharacterEditPage,
		TemplateName:    "character_edit",
		SuccessRedirect: func() string {
			id := path.Base(r.URL.Path)
			return URLFor("character_view") + id
		}(),
		UpdateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager, obj db.Updatable) (db.Insertable, error) {
			character := obj.(*models.Character)
			character.Name = r.FormValue("name")
			character.Blurb = r.FormValue("blurb")
			character.Body = r.FormValue("body")
			tx, err := h.db.DB.Begin()
			if err == nil {
				if err := character.Save(tx); err != nil {
					glog.Errorf("Error saving character on CharacterEditHandler: %v", err.Error())
					return nil, err
				}
				tx.Commit()
				return character, nil
			}
			return nil, err
		},
	}
	HandleCrudEdit(r, w, h.db, h.tr, manager, params)
}

func (h CharacterEditHandler) Methods() []string {
	return h.methods
}

func BuildCharacterEditHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return CharacterEditHandler{
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

type CharacterNewHandler pathforkFrontEndHandler

func (h CharacterNewHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	workId := getWorkId(r)
	manager := sessionManager.New(r, w, h.sessionStore)
	params := crudCreateInput{
		GetCreatePageFunc: pages.GetCharacterNewPage,
		CreateFuncArgs:    []string{workId},
		TemplateName:      "character_edit",
		SuccessRedirect: func() string {
			if workId == "0" {
				return URLFor("character_index")
			} else {
				return URLFor("work_view") + workId
			}
		}(),
		CreateObjFunc: func(r *http.Request, page pages.WebPage, sm sessionManager.SessionManager) (db.Insertable, error) {
			newChar := &models.Character{}
			newChar.Name = r.FormValue("name")
			newChar.Blurb = r.FormValue("blurb")
			newChar.Body = r.FormValue("body")
			newChar.UserEmail = manager.GetUserEmail()
			workId, _ := strconv.Atoi(workId)
			tx, err := h.db.DB.Begin()
			if err == nil {
				newId, err := h.db.Insert(newChar, tx)
				if err == nil && workId != 0 {
					err = models.UpdateWorksCharsRelations(h.db, tx, workId, []int{newId}, []int{})
					if err != nil {
						return nil, err
					}
				}
				if err == nil {
					tx.Commit()
					return newChar, nil
				}
				glog.Error(err.Error())
			}
			glog.Error(err.Error())
			return nil, err
		},
	}
	HandleCrudCreate(r, w, h.db, h.tr, manager, params)
}

func (h CharacterNewHandler) Methods() []string {
	return h.methods
}

func BuildCharacterNewHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return CharacterNewHandler{
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

type CharacterIndexHandler pathforkFrontEndHandler

func (h CharacterIndexHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	characters := models.GetCharactersForUser(manager.GetUserEmail(), h.db)
	page := pages.GetCharacterIndexPage(manager, characters)
	if err := h.tr.RenderPage(w, "character_index", page); err != nil {
		glog.Errorf("Error with CharacterIndex page render: %v", err.Error())
		http.Redirect(w, r, URLFor("dashboard"), 302)
	}
}

func (h CharacterIndexHandler) Methods() []string {
	return h.methods
}

func BuildCharacterIndexHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return CharacterIndexHandler{
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

type CharacterDeleteHandler pathforkFrontEndHandler

func (h CharacterDeleteHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	response := getCrudStarterResponse(r, w, h.db, manager, models.GetCharacterDetail)
	if response.RedirectCode != 0 {
		if response.FlashMsg != "" {
			manager.AddFlash(response.FlashMsg)
		}
		http.Redirect(w, r, URLFor("dashboard"), response.RedirectCode)
		return
	}
	character := response.Obj.(*models.Character)
	form := forms.NewDeleteForm(character.Id)
	form.Populate(r)
	if r.Method == "POST" && form.Validate() {
		idToDelete, _ := strconv.Atoi(r.FormValue("object_id"))
		success, err := models.DeleteCharacter(idToDelete, h.db)
		if err != nil || !success {
			glog.Error(err)
			http.Redirect(w, r, fmt.Sprintf("%v%v", URLFor("character_view"), character.Id), 301)
			return
		}
	}
	manager.AddFlash("OK, that guy won't bother you any more.")
	http.Redirect(w, r, URLFor("character_index"), 301)
}

func (h CharacterDeleteHandler) Methods() []string {
	return h.methods
}

func BuildCharacterDeleteHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return CharacterDeleteHandler{
		tr:           tr,
		methods:      []string{"POST"},
		db:           db,
		sessionStore: store,
	}
}
