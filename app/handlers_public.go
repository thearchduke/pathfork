package pathfork

import (
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/sessions"

	"bitbucket.org/jtyburke/pathfork/app/auth"
	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/messages"
	"bitbucket.org/jtyburke/pathfork/app/models"
	"bitbucket.org/jtyburke/pathfork/app/pages"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
)

type pathforkFrontEndHandler struct {
	tr           *TemplateRenderer
	methods      []string
	db           *db.DB
	sessionStore *sessions.CookieStore
}

// HomeHandler is the handler for the homepage
type HomeHandler pathforkFrontEndHandler

func (h HomeHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Could not find page", 404)
		return
	}
	manager := sessionManager.New(r, w, h.sessionStore)
	page := pages.GetHomePage(manager)
	refreshPage := false
	if r.Method == "POST" {
		form := page.Form
		form.Populate(r)
		valid := form.Validate()
		passwordMatch := r.FormValue("password") == r.FormValue("repeatPassword")
		if valid && passwordMatch {
			newEmail := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
			newUser, err := models.NewUser(newEmail, r.FormValue("password"))
			if err != nil {
				glog.Errorf("User creation error %v", err.Error())
				manager.AddFlash("Sorry, we're having trouble saving that password.")
				refreshPage = true
			} else {
				tx, err := h.db.DB.Begin()
				if err == nil {
					if _, err := h.db.Insert(newUser, tx); err != nil {
						if err.Error() == `pq: duplicate key value violates unique constraint "tbl_user_pkey"` {
							glog.Errorf("Duplicate user user: %v", r.FormValue("email"))
							manager.AddFlash("Looks like that email's already in use. Try logging in above?")
							refreshPage = true
						} else {
							glog.Errorf("Error inserting user: %v", err.Error())
							manager.AddFlash("Something went wrong with our database. Still getting all the kinks out!")
							refreshPage = true
						}
					} else {
						err = messages.SendVerificationEmail(newUser.Email)
						if err == nil {
							manager.AddFlash("Thanks for signing up! Please follow the verification link you've been emailed to get started.")
							refreshPage = true
							tx.Commit()
						} else {
							glog.Error(err)
							manager.AddFlash("Looks like something went wrong with our email provider. Please try again later.")
							refreshPage = true
							tx.Rollback()
						}
					}
				} else {
					glog.Error(err)
					manager.AddFlash("Looks like there's a database error.")
					refreshPage = true
				}
			}
		} else if !passwordMatch {
			form.AddError("Passwords must match.")
			form.Fields["password"].SetData("")
			form.Fields["repeatPassword"].SetData("")
		}
	}
	if refreshPage {
		page = pages.GetHomePage(manager)
	}
	err := h.tr.RenderPage(w, "home", page)
	if err != nil {
		glog.Error(err.Error())
	}
}

func (h HomeHandler) Methods() []string {
	return h.methods
}

func BuildHomeHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return HomeHandler{
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

type AboutHandler pathforkFrontEndHandler

func (h AboutHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	err := h.tr.RenderPage(w, "about", pages.GetAboutPage(manager))
	if err != nil {
		glog.Error(err.Error())
	}
}

func (h AboutHandler) Methods() []string {
	return h.methods
}

func BuildAboutHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return AboutHandler{
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

type ContactHandler pathforkFrontEndHandler

func (h ContactHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	page := pages.GetContactPage(manager)
	if r.Method == "POST" {
		form := page.Form
		form.Populate(r)
		if form.Validate() {
			err := messages.SendContactFormEmail(
				r.FormValue("email"), r.FormValue("message"),
			)
			if err != nil {
				glog.Errorf("Error sending contact form email: %v", err.Error())
				manager.AddFlash("Hm, looks like that didn't go through.")
				page.RefreshUniversals(manager)
			} else {
				glog.Info("Contact form message sent")
				manager.AddFlash("Thanks for the note!")
				page.RefreshUniversals(manager)
			}
		}
	}
	err := h.tr.RenderPage(w, "contact", page)
	if err != nil {
		glog.Error(err.Error())
	}
}

func (h ContactHandler) Methods() []string {
	return h.methods
}

func BuildContactHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return ContactHandler{
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

type AuthHandler pathforkFrontEndHandler

func (h AuthHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	action, ok := query["action"]
	authenticator := auth.NewAuthenticator(r, w, h.sessionStore)
	if r.Method == "POST" && ok && action[0] == "login" {
		form := forms.NewSigninForm()
		form.Populate(r)
		if form.Validate() {
			email := strings.ToLower(r.FormValue("email"))
			user := models.GetUserByEmail(email, h.db)
			if user != nil && user.Verified {
				pw := r.FormValue("password")
				if valid := authenticator.LogUserIn(user.Email, user.Password, pw); !valid {
					authenticator.Manager.AddFlash("Sorry, your login credentials were invalid.")
				}
			} else {
				glog.Infof("Could not find user %v", email)
				authenticator.Manager.AddFlash("Those credentials were incorrect.")
			}
		}
	} else if r.Method == "GET" && ok && action[0] == "logout" {
		authenticator.LogUserOut()
	} else if r.Method == "GET" && ok && action[0] == "verify" {
		token, ok := query["token"]
		if ok {
			email, valid := auth.VerifyToken("verify-email", token[0])
			if valid {
				user := models.GetUserByEmail(email, h.db)
				tx, _ := h.db.DB.Begin()
				err := models.VerifyUser(user, tx)
				if err != nil {
					authenticator.Manager.AddFlash("Looks like there was a database error verifying your email. Ugh!")
				} else {
					tx.Commit()
					authenticator.Manager.AddFlash("You're all set to go! Go ahead and log in above.")
				}
			}
		} else {
			authenticator.Manager.AddFlash("Sorry, that verification link isn't valid.")
		}
	}
	next, ok := query["next"]
	toForward := URLFor("dashboard")
	if ok && len(next) > 1 {
		toForward = next[0]
	}
	http.Redirect(w, r, toForward, 302)
}

func (h AuthHandler) Methods() []string {
	return h.methods
}

func BuildAuthHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return AuthHandler{
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

type ResetHandler pathforkFrontEndHandler

func (h ResetHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	manager := sessionManager.New(r, w, h.sessionStore)
	query := r.URL.Query()
	action, ok := query["action"]
	if !ok {
		http.Redirect(w, r, URLFor("home"), 302)
		return
	}
	if action[0] == "request-reset" {
		page := pages.GetRequestResetPasswordPage(manager)
		if r.Method == "GET" {
			page := pages.GetRequestResetPasswordPage(manager)
			err := h.tr.RenderPage(w, "request_reset_password", page)
			if err != nil {
				glog.Error(err.Error())
				manager.AddFlash("Sorry, that link doesn't work.")
				http.Redirect(w, r, URLFor("home"), 302)
				return
			}
		} else if r.Method == "POST" {
			form := page.Form
			form.Populate(r)
			if ok && form.Validate() {
				user := models.GetUserByEmail(r.FormValue("email"), h.db)
				if user == nil {
					manager.AddFlash("We don't have a record of that email address.")
					// This is not an ultra-secure message! But user-friendly
					http.Redirect(w, r, URLFor("home"), 302)
					return
				}
				err := messages.SendResetPasswordEmail(r.FormValue("email"))
				glog.Info(auth.NewToken(r.FormValue("email"), "reset-password"))
				msg := "OK, check your inbox for the reset email."
				if err != nil {
					msg = "Sorry, something went wrong with our email provider. Please try again later."
				}
				manager.AddFlash(msg)
				http.Redirect(w, r, URLFor("home"), 302)
				return
			}
		}
	} else if action[0] == "reset" {
		token, ok := query["token"]
		email, valid := auth.VerifyToken("reset-password", token[0])
		if !ok || !valid {
			manager.AddFlash("Sorry, that URL isn't valid.")
			http.Redirect(w, r, URLFor("home"), 302)
			return
		}
		page := pages.GetResetPasswordPage(manager, token[0])
		if r.Method == "GET" {
			err := h.tr.RenderPage(w, "reset_password", page)
			if err != nil {
				glog.Error(err.Error())
				manager.AddFlash("Sorry, that link doesn't work.")
				http.Redirect(w, r, URLFor("home"), 302)
				return
			}
		} else if r.Method == "POST" {
			form := page.Form
			form.Populate(r)
			passwordMatch := r.FormValue("newPassword") == r.FormValue("repeatPassword")
			valid := form.Validate()
			if passwordMatch && valid {
				user := models.GetUserByEmail(email, h.db)
				tx, _ := h.db.DB.Begin()
				err := models.UpdatePassword(user, r.FormValue("newPassword"), tx)
				if err != nil {
					manager.AddFlash("Looks like there was a database error resetting your password. Ugh!")
				} else {
					tx.Commit()
					manager.AddFlash("Your password has been reset! Go ahead and log in above.")
				}
				http.Redirect(w, r, URLFor("home"), 302)
				return
			} else {
				form.AddError("The passwords must match.")
				err := h.tr.RenderPage(w, "reset_password", page)
				if err != nil {
					glog.Error(err.Error())
					manager.AddFlash("Sorry, something went wrong.")
					http.Redirect(w, r, URLFor("reset_password"), 302)
					return
				}
			}
		}
	}
	http.Redirect(w, r, URLFor("home"), 302)
	return
}

func (h ResetHandler) Methods() []string {
	return h.methods
}

func BuildResetHandler(tr *TemplateRenderer, db *db.DB, store *sessions.CookieStore) FrontEndHandler {
	return ResetHandler{
		tr:           tr,
		methods:      []string{"GET", "POST"},
		db:           db,
		sessionStore: store,
	}
}
