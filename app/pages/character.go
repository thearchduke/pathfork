package pages

import (
	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/bradfitz/slice"
)

func GetCharacterViewPage(sm sessionManager.SessionManager, verifiable interface{}) WebPage {
	character := verifiable.(*models.Character)
	works := models.GetWorksForCharacter(character.Id, character.DB)
	sections := models.GetSectionsForCharacter(character.Id, character.DB)
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
		Title:          character.Name,
		Headline:       character.Name,
		Name:           "character_view",
		Character:      character,
		Universals:     getUniversals(sm),
		SectionsByWork: sectionsByWork,
	}
}

func GetCharacterEditPage(sm sessionManager.SessionManager, database *db.DB, verifiable interface{}) WebPage {
	character := verifiable.(*models.Character)
	form := forms.NewCharacterForm(sm)
	form.Fields["name"].SetData(character.Name)
	form.Fields["blurb"].SetData(character.Blurb)
	form.Fields["body"].SetData(character.Body)
	return WebPage{
		Title:      character.Name,
		Headline:   character.Name,
		Name:       "character_edit",
		Form:       form,
		Character:  character,
		Universals: getUniversals(sm),
		DeleteForm: forms.NewDeleteForm(character.Id, sm),
	}
}

func GetCharacterNewPage(sm sessionManager.SessionManager, database *db.DB, args ...string) WebPage {
	form := forms.NewCharacterForm(sm)
	workId := args[0]
	return WebPage{
		Headline:   "You must be the new guy.",
		Title:      "Add a character",
		Name:       "character_new",
		Character:  &models.Character{},
		Form:       form,
		NewObj:     true,
		Universals: getUniversals(sm),
		ParentId:   workId,
	}
}

func GetCharacterIndexPage(sm sessionManager.SessionManager, characters []*models.Character) WebPage {
	slice.Sort(characters, func(i, j int) bool {
		return characters[i].Name < characters[j].Name
	})
	return WebPage{
		Headline:       "Here are some folks you wrote",
		Title:          "Characters index",
		Name:           "character_index",
		CharactersList: characters,
		Universals:     getUniversals(sm),
	}
}
