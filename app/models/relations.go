package models

import (
	"database/sql"
	"fmt"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"github.com/golang/glog"
)

func UpdateWorksCharsRelations(database *db.DB, tx *sql.Tx, workId int, charsToInsert, charsToDelete []int) error {
	updater := relationshipUpdater{
		TableName: "r_works_characters",
		InsertIds: charsToInsert,
		DeleteIds: charsToDelete,
		LeftName:  "work_id",
		LeftId:    workId,
		RightName: "character_id",
		Tx:        tx,
	}
	return updateRelations(database, updater)
}

func UpdateWorksCharsNoConflict(database *db.DB, tx *sql.Tx, workId int, charsToInsert []int) error {
	updater := relationshipUpdater{
		TableName:           "r_works_characters",
		InsertIds:           charsToInsert,
		DeleteIds:           []int{},
		LeftName:            "work_id",
		LeftId:              workId,
		RightName:           "character_id",
		Tx:                  tx,
		OnConflictDoNothing: true,
	}
	return updateRelations(database, updater)
}

func UpdateSectionsCharsRelations(database *db.DB, tx *sql.Tx, sectionId int, charsToInsert, charsToDelete []int) error {
	updater := relationshipUpdater{
		TableName: "r_sections_characters",
		InsertIds: charsToInsert,
		DeleteIds: charsToDelete,
		LeftName:  "section_id",
		LeftId:    sectionId,
		RightName: "character_id",
		Tx:        tx,
	}
	return updateRelations(database, updater)
}

func UpdateWorksSettingsRelations(database *db.DB, tx *sql.Tx, workId int, settingsToInsert, settingsToDelete []int) error {
	updater := relationshipUpdater{
		TableName: "r_works_settings",
		InsertIds: settingsToInsert,
		DeleteIds: settingsToDelete,
		LeftName:  "work_id",
		LeftId:    workId,
		RightName: "setting_id",
		Tx:        tx,
	}
	return updateRelations(database, updater)
}

func UpdateWorksSettingsNoConflict(database *db.DB, tx *sql.Tx, workId int, settingsToInsert []int) error {
	updater := relationshipUpdater{
		TableName:           "r_works_settings",
		InsertIds:           settingsToInsert,
		DeleteIds:           []int{},
		LeftName:            "work_id",
		LeftId:              workId,
		RightName:           "setting_id",
		Tx:                  tx,
		OnConflictDoNothing: true,
	}
	return updateRelations(database, updater)
}

func UpdateSectionsSettingsRelations(database *db.DB, tx *sql.Tx, sectionId int, settingsToInsert, settingsToDelete []int) error {
	updater := relationshipUpdater{
		TableName: "r_sections_settings",
		InsertIds: settingsToInsert,
		DeleteIds: settingsToDelete,
		LeftName:  "section_id",
		LeftId:    sectionId,
		RightName: "setting_id",
		Tx:        tx,
	}
	return updateRelations(database, updater)
}

type relationshipUpdater struct {
	TableName           string
	InsertIds           []int
	DeleteIds           []int
	LeftName            string
	LeftId              int
	RightName           string
	Tx                  *sql.Tx
	OnConflictDoNothing bool
}

func (r relationshipUpdater) GetInsertStr() string {
	output := fmt.Sprintf("INSERT INTO %v (%v, %v) VALUES", r.TableName, r.LeftName, r.RightName)
	for i := range r.InsertIds {
		nForSql := i*2 + 1
		output += fmt.Sprintf(`
($%v, $%v)
`, nForSql, nForSql+1)
		if i != len(r.InsertIds)-1 {
			output += ","
		}
	}
	if r.OnConflictDoNothing {
		output += " on conflict do nothing"
	}
	output += " returning 0"
	return output
}

func (r relationshipUpdater) GetInsertArgs() []interface{} {
	output := []interface{}{}
	for i := range r.InsertIds {
		output = append(output, r.LeftId)
		output = append(output, r.InsertIds[i])
	}
	return output
}

func (r relationshipUpdater) GetDeleteStr() string {
	output := fmt.Sprintf("DELETE FROM %v WHERE %v=$1 AND %v IN (", r.TableName, r.LeftName, r.RightName)
	for i := range r.DeleteIds {
		output += fmt.Sprintf("$%v", i+2)
		if i != len(r.DeleteIds)-1 {
			output += ","
		}
	}
	output += ")"
	return output
}

func (r relationshipUpdater) GetDeleteArgs() []interface{} {
	if len(r.DeleteIds) == 0 {
		return nil
	}
	output := make([]interface{}, len(r.DeleteIds)+1)
	output[0] = r.LeftId
	for i := range r.DeleteIds {
		output[i+1] = r.DeleteIds[i]
	}
	return output
}

func updateRelations(database *db.DB, updater relationshipUpdater) error {
	if len(updater.InsertIds) > 0 {
		_, err := database.Insert(updater, updater.Tx)
		if err != nil {
			glog.Error(err.Error())
			return err
		}
	}
	if len(updater.DeleteIds) > 0 {
		err := database.Delete(updater, updater.Tx)
		if err != nil {
			glog.Error(err.Error())
			return err
		}
	}
	return nil
}

type rightForLeftQuery struct {
	RightName      string
	DB             *db.DB
	LeftName       string
	LeftId         int
	ColumnStr      string
	ObjFromRowFunc func(db *db.DB, r *sql.Rows) (db.Insertable, error)
}

func (q rightForLeftQuery) GetQueryStr() string {
	return q.ColumnStr + fmt.Sprintf(`
JOIN r_%vs_%vs
ON tbl_%v.%v_id = r_%vs_%vs.%v_id
WHERE r_%vs_%vs.%v_id=$1`,
		q.LeftName, q.RightName,
		q.RightName, q.RightName, q.LeftName, q.RightName, q.RightName,
		q.LeftName, q.RightName, q.LeftName)
}

func (q rightForLeftQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.LeftId}
}

func (q rightForLeftQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	return q.ObjFromRowFunc(db, r)
}

func getRightForLeft(query rightForLeftQuery) []db.Insertable {
	output, err := query.DB.Query(query)
	if err != nil {
		glog.Errorf("Error on rightForLeftQuery: %v", err.Error())
		return nil
	}
	return output
}

type leftForRightQuery struct {
	RightName      string
	DB             *db.DB
	LeftName       string
	RightId        int
	ColumnStr      string
	ObjFromRowFunc func(db *db.DB, r *sql.Rows) (db.Insertable, error)
}

func (q leftForRightQuery) GetQueryStr() string {
	return q.ColumnStr + fmt.Sprintf(`
JOIN r_%vs_%vs
ON tbl_%v.%v_id = r_%vs_%vs.%v_id
JOIN tbl_%v
ON r_%vs_%vs.%v_id = tbl_%v.%v_id
WHERE tbl_%v.%v_id=$1
ORDER BY tbl_%v.%v_id`,
		q.LeftName, q.RightName,
		q.LeftName, q.LeftName, q.LeftName, q.RightName, q.LeftName,
		q.RightName,
		q.LeftName, q.RightName, q.RightName, q.RightName, q.RightName,
		q.RightName, q.RightName,
		q.LeftName, q.LeftName)
}

func (q leftForRightQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.RightId}
}

func (q leftForRightQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	return q.ObjFromRowFunc(db, r)
}

func getLeftForRight(query leftForRightQuery) []db.Insertable {
	output, err := query.DB.Query(query)
	if err != nil {
		glog.Errorf("Error on leftForRightQuery: %v", err.Error())
		return nil
	}
	return output
}
