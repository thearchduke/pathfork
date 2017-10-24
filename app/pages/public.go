package pages

import (
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
)

func GetHomePage(sm sessionManager.SessionManager) WebPage {
	return WebPage{
		Title:      "Home",
		Name:       "home",
		Form:       forms.NewSignupForm(),
		Universals: getUniversals(sm),
	}
}

func GetAboutPage(sm sessionManager.SessionManager) WebPage {
	return WebPage{
		Title:      "About",
		Name:       "about",
		Universals: getUniversals(sm),
	}
}

func GetContactPage(sm sessionManager.SessionManager) WebPage {
	return WebPage{
		Title:      "Contact",
		Name:       "contact",
		Form:       forms.NewContactForm(),
		Universals: getUniversals(sm),
	}
}

func GetRequestResetPasswordPage(sm sessionManager.SessionManager) WebPage {
	return WebPage{
		Title:      "Reset password",
		Name:       "reset_password",
		Form:       forms.NewRequestResetPasswordForm(),
		Universals: getUniversals(sm),
	}
}

func GetResetPasswordPage(sm sessionManager.SessionManager, token string) WebPage {
	return WebPage{
		Title:      "Request password reset",
		Name:       "request_reset_password",
		Form:       forms.NewResetPasswordForm(),
		Universals: getUniversals(sm),
		Token:      token,
	}
}
