package forms

import (
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/utils"
)

func WorksToFormOptions(works []*models.Work, selectedWorks ...*models.Work) []map[string]string {
	worksMap := make([]map[string]string, len(works))
	for i, work := range works {
		worksMap[i] = map[string]string{
			"value": fmt.Sprintf("%v", work.Id),
			"text":  work.Title,
		}
		for _, selected := range selectedWorks {
			if selected.Id == work.Id {
				worksMap[i]["selected"] = "true"
			}
		}
	}
	return worksMap
}

func CharsToFormOptions(chars []*models.Character, selectedChars ...*models.Character) []map[string]string {
	charsMap := make([]map[string]string, len(chars))
	for i, char := range chars {
		charsMap[i] = map[string]string{
			"value": fmt.Sprintf("%v", char.Id),
			"text":  char.Name,
		}
		for _, selected := range selectedChars {
			if selected.Id == char.Id {
				charsMap[i]["selected"] = "true"
			}
		}
	}
	return charsMap
}

func SettingsToFormOptions(settings []*models.Setting, selectedSettings ...*models.Setting) []map[string]string {
	settingsMap := make([]map[string]string, len(settings))
	for i, setting := range settings {
		settingsMap[i] = map[string]string{
			"value": fmt.Sprintf("%v", setting.Id),
			"text":  setting.Name,
		}
		for _, selected := range selectedSettings {
			if selected.Id == setting.Id {
				settingsMap[i]["selected"] = "true"
			}
		}
	}
	return settingsMap
}

func GetCurrentIds(options []map[string]string) string {
	currentIds := ""
	for i := range options {
		if options[i]["selected"] == "true" {
			currentIds += fmt.Sprintf("%v,", options[i]["value"])
		}
	}
	if len(currentIds) > 0 {
		currentIds = currentIds[:len(currentIds)-1]
	}
	return currentIds
}

func GetRelationUpdateIds(r *http.Request, oldFieldname, newFieldname string) (toInsert, toDelete []int, err error) {
	oldIds := strings.Split(r.FormValue(oldFieldname), ",")
	newIds := r.Form[newFieldname]
	toInsert, err = utils.StringsToInts(
		utils.StringSliceDifference(newIds, oldIds),
	)
	toDelete, err = utils.StringsToInts(
		utils.StringSliceDifference(oldIds, newIds),
	)
	return toInsert, toDelete, err
}
