package pages

import (
	"fmt"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/bradfitz/slice"
)

func GetWorkViewPage(sm sessionManager.SessionManager, verifiable interface{}) WebPage {
	work := verifiable.(*models.Work)
	sm.SetCurrentWork(work.Id, work.Title)
	sections, snippets := models.GetSectionsForWork(work.Id, work.DB)
	return WebPage{
		Title:          fmt.Sprintf("View work: %v", work.Title),
		Name:           "work_view",
		Work:           work,
		SectionsList:   sections,
		CharactersList: models.GetCharactersForWork(work.Id, work.DB),
		SettingsList:   models.GetSettingsForWork(work.Id, work.DB),
		SnippetsList:   snippets,
		Universals:     getUniversals(sm),
	}
}

func GetWorkEditPage(sm sessionManager.SessionManager, database *db.DB, verifiable interface{}) WebPage {
	work := verifiable.(*models.Work)
	charsMap := forms.CharsToFormOptions(
		models.GetCharactersForUser(sm.GetUserEmail(), database),
		models.GetCharactersForWork(work.Id, database)...,
	)
	settingsMap := forms.SettingsToFormOptions(
		models.GetSettingsForUser(sm.GetUserEmail(), database),
		models.GetSettingsForWork(work.Id, database)...,
	)
	form := forms.NewWorkForm(charsMap, settingsMap)
	form.Fields["title"].SetData(work.Title)
	form.Fields["blurb"].SetData(work.Blurb)
	return WebPage{
		Title:      fmt.Sprintf("Edit work: %v", work.Title),
		Headline:   work.Title,
		Name:       "work_edit",
		Work:       work,
		Form:       form,
		Universals: getUniversals(sm),
		DeleteForm: forms.NewDeleteForm(work.Id),
	}
}

func GetWorkNewPage(sm sessionManager.SessionManager, database *db.DB, args ...string) WebPage {
	charsMap := forms.CharsToFormOptions(
		models.GetCharactersForUser(sm.GetUserEmail(), database),
	)
	settingsMap := forms.SettingsToFormOptions(
		models.GetSettingsForUser(sm.GetUserEmail(), database),
	)
	return WebPage{
		Headline:   "How exciting! You're starting a new work.",
		Title:      "Add a work",
		Name:       "work_new",
		Work:       &models.Work{},
		Form:       forms.NewWorkForm(charsMap, settingsMap),
		NewObj:     true,
		Universals: getUniversals(sm),
	}
}

func GetWorkExportPage(sm sessionManager.SessionManager, work *models.Work, sections []*models.Section,
	snippets []*models.Section, settings []*models.Setting, characters []*models.Character) WebPage {
	slice.Sort(sections, func(i, j int) bool {
		return sections[i].Order < sections[j].Order
	})
	return WebPage{
		Work:           work,
		SectionsList:   sections,
		CharactersList: characters,
		SettingsList:   settings,
		SnippetsList:   snippets,
	}
}
