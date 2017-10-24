package pages

import (
	"fmt"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
)

func GetSectionViewPage(sm sessionManager.SessionManager, verifiable interface{}) WebPage {
	section := verifiable.(*models.Section)
	characters := models.GetCharactersForSection(section.Id, section.DB)
	settings := models.GetSettingsForSection(section.Id, section.DB)
	return WebPage{
		Title:          fmt.Sprintf("View section: %v", section.Title),
		Name:           "section_view",
		Section:        section,
		Universals:     getUniversals(sm),
		CharactersList: characters,
		SettingsList:   settings,
	}
}

func GetSectionEditPage(sm sessionManager.SessionManager, database *db.DB, verifiable interface{}) WebPage {
	section := verifiable.(*models.Section)
	charsMap := forms.CharsToFormOptions(
		models.GetCharactersForUser(sm.GetUserEmail(), database),
		models.GetCharactersForSection(section.Id, database)...,
	)
	settingsMap := forms.SettingsToFormOptions(
		models.GetSettingsForUser(sm.GetUserEmail(), database),
		models.GetSettingsForSection(section.Id, database)...,
	)
	form := forms.NewSectionForm(charsMap, settingsMap)
	form.Fields["title"].SetData(section.Title)
	form.Fields["blurb"].SetData(section.Blurb)
	form.Fields["body"].SetData(section.Body)
	if section.Snippet == true {
		form.Fields["snippet"].SetData("on")
	}
	return WebPage{
		Title:      fmt.Sprintf("Edit section: %v", section.Title),
		Name:       "section_edit",
		Section:    section,
		Form:       form,
		Universals: getUniversals(sm),
		DeleteForm: forms.NewDeleteForm(section.Id),
	}
}

func GetSectionReorderPage(sm sessionManager.SessionManager, database *db.DB, work *models.Work) WebPage {
	sections, _ := models.GetSectionsForWork(work.Id, database)
	return WebPage{
		Title:        fmt.Sprintf("Reorder sections for %v", work.Title),
		Name:         "section_reorder",
		SectionsList: sections,
		Work:         work,
		Universals:   getUniversals(sm),
	}
}

func GetSectionNewPage(sm sessionManager.SessionManager, database *db.DB, args ...string) WebPage {
	charsMap := forms.CharsToFormOptions(
		models.GetCharactersForUser(sm.GetUserEmail(), database),
	)
	settingsMap := forms.SettingsToFormOptions(
		models.GetSettingsForUser(sm.GetUserEmail(), database),
	)
	form := forms.NewSectionForm(charsMap, settingsMap)
	workId := args[0]
	return WebPage{
		Title:      "New section",
		Name:       "section_edit",
		Section:    &models.Section{},
		NewObj:     true,
		ParentId:   workId,
		Form:       form,
		Universals: getUniversals(sm),
	}
}
