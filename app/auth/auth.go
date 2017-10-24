package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/jtyburke/pathfork/app/config"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/bwmarrin/go-alone"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	Manager sessionManager.SessionManager
}

func NewAuthenticator(r *http.Request, w http.ResponseWriter, s *sessions.CookieStore) Authenticator {
	return Authenticator{
		Manager: sessionManager.New(r, w, s),
	}
}

func IsLoggedIn(r *http.Request, s *sessions.CookieStore) bool {
	session, err := s.Get(r, config.SessionCookieName)
	if err != nil {
		glog.Error(err)
	}
	if _, ok := session.Values["userEmail"]; ok {
		return true
	}
	return false
}

func HashPassword(raw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(raw), 10)
	if err != nil {
		glog.Errorf("Hash error: %v", err.Error())
		return "", err
	}
	return string(bytes), nil
}

func (p *Authenticator) LogUserIn(userEmail string, hashedPassword string, rawPassword string) bool {
	if valid := checkPassword(rawPassword, hashedPassword); !valid {
		return false
	}
	p.Manager.SetUser(userEmail)
	if err := p.Manager.Save(); err != nil {
		glog.Error(err.Error())
		return false
	}
	return true
}

func checkPassword(rawPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	return err == nil
}

func (p *Authenticator) LogUserOut() {
	p.Manager.DeleteUser()
}

func NewSigner() *goalone.Sword {
	signer := goalone.New([]byte(config.HMACKey))
	return signer
}

func VerifyToken(kind, token string) (string, bool) {
	decoded, _ := base64.URLEncoding.DecodeString(token)
	signer := NewSigner()
	data, err := signer.Unsign([]byte(decoded))
	if err != nil {
		return "", false
	}
	split := strings.Split(string(data), "||")
	if err != nil || fmt.Sprintf("%v||%v", split[0], kind) != string(data) {
		return "", false
	}
	return split[0], true
}

func NewToken(email, kind string) string {
	signer := NewSigner()
	token := signer.Sign([]byte(fmt.Sprintf("%v||%v", email, kind)))
	encoded := base64.URLEncoding.EncodeToString(token)
	return string(encoded)
}
