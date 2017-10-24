package models

import (
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
)

type Verifiable interface {
	VerifyPermission(sessionManager.SessionManager) bool
}
