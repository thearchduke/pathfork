package models

import (
	"database/sql"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/bradfitz/slice"
	"github.com/golang/glog"
)

type Setting struct {
	Id        int
	Name      string
	Blurb     string
	UserEmail string
	DB        *db.DB
	Body      string
}

var settingListColumnStr = "SELECT tbl_setting.setting_id, name, tbl_setting.blurb, tbl_setting.user_email FROM tbl_setting"
var settingDetailColumnStr = "SELECT tbl_setting.setting_id, name, tbl_setting.blurb, tbl_setting.body, tbl_setting.user_email FROM tbl_setting"

func (s *Setting) VerifyPermission(sm sessionManager.SessionManager) bool {
	return s.UserEmail == sm.GetUserEmail()
}

func (s *Setting) GetInsertStr() string {
	return `
INSERT INTO tbl_setting(name, blurb, body, user_email)
VALUES ($1, $2, $3, $4)
RETURNING setting_id;
`
}

func (s *Setting) GetInsertArgs() []interface{} {
	return []interface{}{s.Name, db.ToNullString(s.Blurb), db.ToNullString(s.Body), s.UserEmail}
}

func (s *Setting) GetUpdateStr() string {
	return `
UPDATE tbl_setting
SET name=$1, blurb=$2, body=$3
WHERE setting_id=$4
`
}

func (s *Setting) GetUpdateArgs() []interface{} {
	return []interface{}{s.Name, s.Blurb, s.Body, s.Id}
}

func (s *Setting) Save(tx *sql.Tx) error {
	return s.DB.Update(s, tx)
}

func GetSettingById(id int, database *db.DB) Verifiable {
	query := settingByIdQuery{Id: id}
	settingInt, err := database.Query(query)
	if err != nil {
		glog.Error(err.Error())
		return nil
	}
	if len(settingInt) == 0 {
		return nil
	}
	return settingInt[0].(*Setting)
}

type settingByIdQuery struct {
	Id int
}

func (q settingByIdQuery) GetQueryStr() string {
	return settingDetailColumnStr + " where setting_id=$1"
}

func (q settingByIdQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.Id}
}

func (q settingByIdQuery) ObjFromRow(database *db.DB, r *sql.Rows) (db.Insertable, error) {
	return settingDetailFromRow(database, r)
}

func settingListFromRow(database *db.DB, r *sql.Rows) (db.Insertable, error) {
	setting := Setting{DB: database}
	nullBlurb := sql.NullString{}
	if err := r.Scan(&setting.Id, &setting.Name, &nullBlurb, &setting.UserEmail); err != nil {
		glog.Error(err.Error())
		return nil, err
	}
	setting.Blurb = nullBlurb.String
	return &setting, nil
}

func settingDetailFromRow(database *db.DB, r *sql.Rows) (db.Insertable, error) {
	setting := Setting{DB: database}
	nullBlurb := sql.NullString{}
	nullBody := sql.NullString{}
	if err := r.Scan(&setting.Id, &setting.Name, &nullBlurb, &nullBody, &setting.UserEmail); err != nil {
		glog.Error(err.Error())
		return nil, err
	}
	setting.Blurb = nullBlurb.String
	setting.Body = nullBody.String
	return &setting, nil
}

func getSettingsForLeft(leftName string, leftId int, database *db.DB, detailLevel string) []*Setting {
	columnStr := ""
	var objFromRow func(db *db.DB, r *sql.Rows) (db.Insertable, error)
	if detailLevel == "list" {
		columnStr = settingListColumnStr
		objFromRow = settingListFromRow
	} else {
		columnStr = settingDetailColumnStr
		objFromRow = settingDetailFromRow
	}
	query := rightForLeftQuery{
		LeftName:       leftName,
		LeftId:         leftId,
		RightName:      "setting",
		ColumnStr:      columnStr,
		ObjFromRowFunc: objFromRow,
		DB:             database,
	}
	settingsInt := getRightForLeft(query)
	output := make([]*Setting, len(settingsInt))
	for i := range settingsInt {
		output[i] = settingsInt[i].(*Setting)
	}
	return output
}

func GetSettingsForWork(workId int, database *db.DB) []*Setting {
	settings := getSettingsForLeft("work", workId, database, "list")
	slice.Sort(settings, func(i, j int) bool {
		return settings[i].Name < settings[j].Name
	})
	return settings
}

func GetSettingsForSection(sectionId int, database *db.DB) []*Setting {
	return getSettingsForLeft("section", sectionId, database, "list")
}

func GetSettingsForUser(userEmail string, database *db.DB) []*Setting {
	query := settingsForUserQuery{UserEmail: userEmail}
	settingsInt, err := database.Query(query)
	if err != nil {
		glog.Errorf("Error on GetSettingsForUser: %v", err.Error())
		return nil
	}
	output := make([]*Setting, len(settingsInt))
	for i := range settingsInt {
		output[i] = settingsInt[i].(*Setting)
	}
	return output
}

type settingsForUserQuery struct {
	UserEmail string
}

func (q settingsForUserQuery) GetQueryStr() string {
	return settingListColumnStr + " WHERE user_email=$1"
}

func (q settingsForUserQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.UserEmail}
}

func (q settingsForUserQuery) ObjFromRow(database *db.DB, r *sql.Rows) (db.Insertable, error) {
	return settingListFromRow(database, r)
}

func DeleteSetting(settingId int, database *db.DB) (bool, error) {
	return db.DoBasicDelete(settingId, "setting", database)
}
