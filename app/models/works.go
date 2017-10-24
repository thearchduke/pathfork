package models

import (
	"database/sql"
	"fmt"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/golang/glog"
)

var workListColumnStr = "select tbl_work.work_id, tbl_work.title, tbl_work.blurb, tbl_work.user_email from tbl_work"

type Work struct {
	Title     string
	Blurb     string
	DB        *db.DB
	UserEmail string
	Id        int
}

func NewWork(title string, blurb string, email string) *Work {
	return &Work{
		Title:     title,
		Blurb:     blurb,
		UserEmail: email,
	}
}

func (w *Work) VerifyPermission(sm sessionManager.SessionManager) bool {
	return w.UserEmail == sm.GetUserEmail()
}

func (w *Work) GetInsertStr() string {
	return `
INSERT INTO tbl_work(title, blurb, user_email)
VALUES ($1, $2, $3) returning work_id`
}

func (w *Work) GetInsertArgs() []interface{} {
	return []interface{}{
		w.Title,
		db.ToNullString(w.Blurb),
		w.UserEmail,
	}
}

func (w *Work) GetUpdateStr() string {
	return `
UPDATE tbl_work
SET title=$1, blurb=$2
WHERE work_id=$3
`
}

func (w *Work) GetUpdateArgs() []interface{} {
	return []interface{}{w.Title, w.Blurb, w.Id}
}

func (w *Work) Save(tx *sql.Tx) error {
	return w.DB.Update(w, tx)
}

func GetWorkById(id int, db *db.DB) Verifiable {
	query := workByIdQuery{Id: id}
	workInt, err := db.Query(query)
	if err != nil {
		glog.Error(err)
		return nil
	}
	if len(workInt) == 0 {
		return nil
	}
	return workInt[0].(*Work)
}

type workByIdQuery struct {
	Id int
}

func (q workByIdQuery) GetQueryStr() string {
	return workListColumnStr + " where work_id=$1"
}

func (q workByIdQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.Id}
}

func (q workByIdQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	return workFromRow(db, r)
}

func workFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	work := Work{DB: db}
	nullBlurb := sql.NullString{}
	if err := r.Scan(&work.Id, &work.Title, &nullBlurb, &work.UserEmail); err != nil {
		return nil, err
	}
	work.Blurb = nullBlurb.String
	return &work, nil
}

func parseMultiworkQuery(workInt []db.Insertable, err error) []*Work {
	if err != nil {
		fmt.Printf("Error on multiwork query: %v", err)
		return nil
	}
	output := make([]*Work, len(workInt))
	for i := range workInt {
		output[i] = workInt[i].(*Work)
	}
	return output
}

func GetWorksForUser(email string, db *db.DB) []*Work {
	query := worksForUserQuery{Email: email}
	return parseMultiworkQuery(db.Query(query))
}

type worksForUserQuery struct {
	Email string
}

func (q worksForUserQuery) GetQueryStr() string {
	return workListColumnStr + " where user_email=$1"
}

func (q worksForUserQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.Email}
}

func (q worksForUserQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	return workFromRow(db, r)
}

func GetWorksForCharacter(characterId int, database *db.DB) []*Work {
	query := leftForRightQuery{
		RightName:      "character",
		DB:             database,
		LeftName:       "work",
		RightId:        characterId,
		ColumnStr:      workListColumnStr,
		ObjFromRowFunc: workFromRow,
	}
	worksInt := getLeftForRight(query)
	output := make([]*Work, len(worksInt))
	for i := range worksInt {
		output[i] = worksInt[i].(*Work)
	}
	return output
}

func GetWorksForSetting(settingId int, database *db.DB) []*Work {
	query := leftForRightQuery{
		RightName:      "setting",
		DB:             database,
		LeftName:       "work",
		RightId:        settingId,
		ColumnStr:      workListColumnStr,
		ObjFromRowFunc: workFromRow,
	}
	worksInt := getLeftForRight(query)
	output := make([]*Work, len(worksInt))
	for i := range worksInt {
		output[i] = worksInt[i].(*Work)
	}
	return output
}

func DeleteWork(workId int, database *db.DB) (bool, error) {
	return db.DoBasicDelete(workId, "work", database)
}
