package pages

import (
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
)

func GetDashboardPage(sm sessionManager.SessionManager, works []*models.Work) WebPage {
	return WebPage{
		Title:      "Dashboard",
		Name:       "dashboard",
		WorksList:  works,
		Universals: getUniversals(sm),
	}
}
