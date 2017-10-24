package models

import (
	"strings"
	"testing"

	"bitbucket.org/jtyburke/pathfork/app/db"
)

func TestInserts(t *testing.T) {
	objects := []db.Insertable{&Section{}, &Work{}, &Character{}}
	for _, obj := range objects {
		queryStr := obj.GetInsertStr()
		queryArgs := obj.GetInsertArgs()
		numStrArgs := strings.Count(queryStr, "$")
		if numStrArgs != len(queryArgs) {
			t.Errorf("Mismatch in number of arguments for section insert string and args on %v", obj.GetInsertStr())
		}
	}
}

func TestUpdates(t *testing.T) {
	objects := []db.Updatable{&Section{}, &Work{}, &Character{}}
	for _, obj := range objects {
		queryStr := obj.GetUpdateStr()
		queryArgs := obj.GetUpdateArgs()
		numStrArgs := strings.Count(queryStr, "$")
		if numStrArgs != len(queryArgs) {
			t.Errorf("Mismatch in number of arguments for section insert string and args on %v", obj.GetUpdateStr())
		}
	}
}

func TestQueries(t *testing.T) {
	objects := []db.Queryable{
		&sectionByIdQuery{},
		&workByIdQuery{},
		&worksForUserQuery{},
		&characterDetailQuery{},
		&charactersForUserQuery{},
	}
	for _, obj := range objects {
		queryStr := obj.GetQueryStr()
		queryArgs := obj.GetQueryArgs()
		numStrArgs := strings.Count(queryStr, "$")
		if numStrArgs != len(queryArgs) {
			t.Errorf("Mismatch in number of arguments for section insert string and args on %v", obj.GetQueryStr())
		}
	}
}
