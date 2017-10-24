package pages

import (
	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/bradfitz/slice"
)

func GetSettingViewPage(sm sessionManager.SessionManager, verifiable interface{}) WebPage {
	setting := verifiable.(*models.Setting)
	works := models.GetWorksForSetting(setting.Id, setting.DB)
	sections := models.GetSectionsForSetting(setting.Id, setting.DB)
	slice.Sort(sections, func(i, j int) bool {
		return sections[i].WorkId < sections[j].WorkId
	})
	sectionsByWork := map[*models.Work][]*models.Section{}
	for i := range works {
		sectionsByWork[works[i]] = []*models.Section{}
		for j := range sections {
			if sections[j].WorkId != works[i].Id {
				break
			}
			sectionsByWork[works[i]] = append(sectionsByWork[works[i]], sections[j])
		}
	}
	return WebPage{
		Title:          setting.Name,
		Headline:       setting.Name,
		Name:           "setting_view",
		Setting:        setting,
		Universals:     getUniversals(sm),
		SectionsByWork: sectionsByWork,
	}
}

func GetSettingEditPage(sm sessionManager.SessionManager, database *db.DB, verifiable interface{}) WebPage {
	setting := verifiable.(*models.Setting)
	form := forms.NewSettingForm()
	form.Fields["name"].SetData(setting.Name)
	form.Fields["blurb"].SetData(setting.Blurb)
	form.Fields["body"].SetData(setting.Body)
	return WebPage{
		Title:      setting.Name,
		Headline:   setting.Name,
		Name:       "setting_edit",
		Setting:    setting,
		Universals: getUniversals(sm),
		Form:       form,
		DeleteForm: forms.NewDeleteForm(setting.Id),
	}
}

func GetSettingIndexPage(sm sessionManager.SessionManager, settings []*models.Setting) WebPage {
	slice.Sort(settings, func(i, j int) bool {
		return settings[i].Name < settings[j].Name
	})
	return WebPage{
		Headline:     "Oh the places your stories will go!",
		Title:        "Settings index",
		Name:         "setting_index",
		SettingsList: settings,
		Universals:   getUniversals(sm),
	}
}

func GetSettingNewPage(sm sessionManager.SessionManager, database *db.DB, args ...string) WebPage {
	form := forms.NewSettingForm()
	workId := args[0]
	return WebPage{
		Headline:   "So tell me about this place.",
		Title:      "Add a setting",
		Name:       "setting_new",
		Setting:    &models.Setting{},
		Form:       form,
		NewObj:     true,
		Universals: getUniversals(sm),
		ParentId:   workId,
	}
}
