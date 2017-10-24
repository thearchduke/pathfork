package models

import (
	"database/sql"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/bradfitz/slice"
	"github.com/golang/glog"
)

type Character struct {
	Id        int
	Name      string
	Blurb     string
	UserEmail string
	DB        *db.DB
	Body      string
}

var characterDetailColumnStr = "SELECT tbl_character.character_id, tbl_character.name, tbl_character.blurb, tbl_character.body, tbl_character.user_email FROM tbl_character"
var characterListColumnStr = "SELECT tbl_character.character_id, tbl_character.name, tbl_character.blurb FROM tbl_character"

func (c *Character) VerifyPermission(sm sessionManager.SessionManager) bool {
	return c.UserEmail == sm.GetUserEmail()
}

func (c *Character) GetInsertStr() string {
	return `
INSERT INTO tbl_character(name, blurb, body, user_email)
VALUES ($1, $2, $3, $4)
RETURNING character_id
`
}

func (c *Character) GetInsertArgs() []interface{} {
	return []interface{}{c.Name, db.ToNullString(c.Blurb), db.ToNullString(c.Body), c.UserEmail}
}

func (c *Character) GetUpdateStr() string {
	return `
UPDATE tbl_character
SET name=$1, blurb=$2, body=$3
WHERE character_id=$4
`
}

func (c *Character) GetUpdateArgs() []interface{} {
	return []interface{}{c.Name, c.Blurb, c.Body, c.Id}
}

func (c *Character) Save(tx *sql.Tx) error {
	return c.DB.Update(c, tx)
}

func GetCharacterDetail(id int, db *db.DB) Verifiable {
	query := characterDetailQuery{Id: id}
	charInt, err := db.Query(query)
	if err != nil {
		glog.Errorf("Error with GetCharacterDetail: %v", err.Error())
		return nil
	}
	if len(charInt) == 0 {
		return nil
	}
	return charInt[0].(*Character)
}

type characterDetailQuery struct {
	Id int
}

func (q characterDetailQuery) GetQueryStr() string {
	return characterDetailColumnStr + " WHERE character_id=$1"
}

func (q characterDetailQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.Id}
}

func (q characterDetailQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	return characterDetailFromRow(db, r)
}

func characterDetailFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	character := Character{DB: db}
	nullBlurb := sql.NullString{}
	nullBody := sql.NullString{}
	if err := r.Scan(&character.Id, &character.Name, &nullBlurb, &nullBody, &character.UserEmail); err != nil {
		glog.Errorf("Error with characterFromRow: %v", err.Error())
		return nil, err
	}
	character.Blurb = nullBlurb.String
	character.Body = nullBody.String
	return &character, nil
}

func characterListFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	character := Character{DB: db}
	nullBlurb := sql.NullString{}
	if err := r.Scan(&character.Id, &character.Name, &nullBlurb); err != nil {
		glog.Errorf("Error with characterFromRow: %v", err.Error())
		return nil, err
	}
	character.Blurb = nullBlurb.String
	return &character, nil
}

func getCharactersForLeft(leftName string, leftId int, database *db.DB, detailLevel string) []*Character {
	columnStr := ""
	var objFromRow func(db *db.DB, r *sql.Rows) (db.Insertable, error)
	if detailLevel == "list" {
		columnStr = characterListColumnStr
		objFromRow = characterListFromRow
	} else {
		columnStr = characterDetailColumnStr
		objFromRow = characterDetailFromRow
	}
	query := rightForLeftQuery{
		LeftName:       leftName,
		LeftId:         leftId,
		RightName:      "character",
		ColumnStr:      columnStr,
		ObjFromRowFunc: objFromRow,
		DB:             database,
	}
	charsInt := getRightForLeft(query)
	output := make([]*Character, len(charsInt))
	for i := range charsInt {
		output[i] = charsInt[i].(*Character)
	}
	return output
}

func GetCharactersForWork(workId int, database *db.DB) []*Character {
	characters := getCharactersForLeft("work", workId, database, "list")
	slice.Sort(characters, func(i, j int) bool {
		return characters[i].Name < characters[j].Name
	})
	return characters
}

func GetCharactersForSection(sectionId int, database *db.DB) []*Character {
	return getCharactersForLeft("section", sectionId, database, "list")
}

func GetCharactersForUser(userEmail string, database *db.DB) []*Character {
	query := charactersForUserQuery{UserEmail: userEmail}
	charactersInt, err := database.Query(query)
	if err != nil {
		glog.Errorf("Error on GetCharactersForUser: %v", err.Error())
		return nil
	}
	output := make([]*Character, len(charactersInt))
	for i := range charactersInt {
		output[i] = charactersInt[i].(*Character)
	}
	return output
}

type charactersForUserQuery struct {
	UserEmail string
}

func (q charactersForUserQuery) GetQueryStr() string {
	return characterListColumnStr + " WHERE user_email=$1 ORDER BY name"
}

func (q charactersForUserQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.UserEmail}
}

func (q charactersForUserQuery) ObjFromRow(database *db.DB, r *sql.Rows) (db.Insertable, error) {
	return characterListFromRow(database, r)
}

func DeleteCharacter(characterId int, database *db.DB) (bool, error) {
	return db.DoBasicDelete(characterId, "character", database)
}
