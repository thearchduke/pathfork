package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func TestCheckPassword(t *testing.T) {
	if valid := checkPassword("password", "password"); valid {
		t.Error("plaintext password storage is bad")
	}
}

func TestHashPassword(t *testing.T) {
	hashed, _ := HashPassword("password")
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte("password"))
	if err != nil {
		t.Errorf("bcrypt error: %v", err.Error())
	}
}

func TestAuthenticator(t *testing.T) {
	store := sessions.NewCookieStore([]byte("whatever"))
	r, _ := http.NewRequest("POST", "/auth", nil)
	w := httptest.NewRecorder()
	authenticator := NewAuthenticator(r, w, store)
	if _, ok := authenticator.Manager.Session.Values["userEmail"]; ok {
		t.Error("userEmail is set at init for some reason")
	}
	//TODO these tests are broken
	authenticator.Manager.SetUser("tynanburke@gmail.com")
	if authenticator.Manager.Session.Values["userEmail"] != "tynanburke@gmail.com" {
		t.Error("SetUser not working")
	}
	authenticator.LogUserOut()
	if _, ok := authenticator.Manager.Session.Values["userEmail"]; ok {
		t.Error("logUserOut is failing for some reason")
	}
}

func TestTokens(t *testing.T) {
	newToken := NewToken("tynanburke+2@gmail.com", "verify-email")
	email, verified := VerifyToken("verify-email", newToken)
	if !verified {
		t.Error("Token is not verifying")
	}
	if email != "tynanburke+2@gmail.com" {
		fmt.Println(email)
		t.Error("b64 decode failing")
	}
	_, verified = VerifyToken("flooglemorp", newToken)
	if verified {
		t.Error("False positive on token verification")
	}
	newTSToken := NewTSToken("tynanburke@gmail.com", "csrf")
	identifier, valid := VerifyTSToken("csrf", newTSToken, 0)
	if !(!valid && identifier == "expired") {
		t.Errorf("Expiration invalidation failing with valid=%v, ident=%v", valid, identifier)
	}
	identifier, valid = VerifyTSToken("csrf", newTSToken, 1)
	if !(valid && identifier == "tynanburke@gmail.com") {
		t.Errorf("Validation failing with valid=%v, ident=%v", valid, identifier)
	}
}
