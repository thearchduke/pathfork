package pathfork

const StaticRoute = "/static/"

var namePathMap map[string]string

type Route struct {
	Path    string
	Handler FrontEndHandlerBuilder
	Name    string
	Public  bool
}

var FrontEndRoutes = []Route{
	Route{"/dashboard", BuildDashboardHandler, "dashboard", false},

	Route{"/character/new", BuildCharacterNewHandler, "character_new", false},
	Route{"/character/edit/", BuildCharacterEditHandler, "character_edit", false},
	Route{"/character/view/", BuildCharacterViewHandler, "character_view", false},
	Route{"/character/index/", BuildCharacterIndexHandler, "character_index", false},
	Route{"/character/delete/", BuildCharacterDeleteHandler, "character_delete", false},

	Route{"/section/new", BuildSectionNewHandler, "section_new", false},
	Route{"/section/edit/", BuildSectionEditHandler, "section_edit", false},
	Route{"/section/view/", BuildSectionViewHandler, "section_view", false},
	Route{"/section/delete/", BuildSectionDeleteHandler, "section_delete", false},
	Route{"/work/reorder/", BuildSectionReorderHandler, "section_reorder", false},

	Route{"/setting/new", BuildSettingNewHandler, "setting_new", false},
	Route{"/setting/edit/", BuildSettingEditHandler, "setting_edit", false},
	Route{"/setting/view/", BuildSettingViewHandler, "setting_view", false},
	Route{"/setting/index/", BuildSettingIndexHandler, "setting_index", false},
	Route{"/setting/delete/", BuildSettingDeleteHandler, "setting_delete", false},

	Route{"/work/new", BuildWorkNewHandler, "work_new", false},
	Route{"/work/edit/", BuildWorkEditHandler, "work_edit", false},
	Route{"/work/view/", BuildWorkViewHandler, "work_view", false},
	Route{"/work/export/", BuildWorkExportHandler, "work_export", false},
	Route{"/work/delete/", BuildWorkDeleteHandler, "work_delete", false},

	Route{"/about", BuildAboutHandler, "about", true},
	Route{"/contact", BuildContactHandler, "contact", true},
	Route{"/auth", BuildAuthHandler, "auth", true},
	Route{"/reset", BuildResetHandler, "reset", true},
	Route{"/", BuildHomeHandler, "home", true},
}

var publicRoutes map[string]bool

func InitRoutes() {
	namePathMap = make(map[string]string)
	publicRoutes = make(map[string]bool)
	for _, route := range FrontEndRoutes {
		namePathMap[route.Name] = route.Path
		if route.Public {
			publicRoutes[route.Path] = true
		}
	}
	publicRoutes[""] = true
}
