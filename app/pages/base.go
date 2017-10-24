package pages

import (
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/gorilla/sessions"
)

type universals struct {
	Flashes []string
	Session *sessions.Session
}

type WebPage struct {
	Title          string
	Headline       string
	Body           string
	Footer         string
	Styles         string
	Scripts        string
	Name           string
	Form           *forms.Form
	WorksList      []*models.Work
	Work           *models.Work
	SectionsList   []*models.Section
	Section        *models.Section
	CharactersList []*models.Character
	Character      *models.Character
	SettingsList   []*models.Setting
	Setting        *models.Setting
	NewObj         bool
	ParentId       string
	Universals     universals
	SectionsByWork map[*models.Work][]*models.Section
	DeleteForm     *forms.Form
	Token          string
	SnippetsList   []*models.Section
}

func (w WebPage) RefreshUniversals(sm sessionManager.SessionManager) {
	w.Universals = getUniversals(sm)
}

func getFlashes(sm sessionManager.SessionManager) []string {
	flashes := sm.Session.Flashes()
	output := make([]string, len(flashes))
	for i := range flashes {
		output[i] = flashes[i].(string)
	}
	sm.Save()
	return output
}

func getUniversals(sm sessionManager.SessionManager) universals {
	return universals{
		Session: sm.Session,
		Flashes: getFlashes(sm),
	}
}
