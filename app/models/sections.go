package models

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/bradfitz/slice"
	"github.com/golang/glog"
)

const sectionListColumnStr = "select tbl_section.section_id, tbl_section.title, tbl_section.blurb, tbl_section.user_email, tbl_section.work_id, tbl_section.section_order, tbl_section.is_snippet, tbl_section.word_count from tbl_section"
const sectionDetailColumnStr = "select tbl_section.section_id, tbl_section.title, tbl_section.blurb, tbl_section.body, user_email, tbl_section.work_id, tbl_section.section_order, tbl_section.is_snippet, tbl_section.word_count from tbl_section"

type Section struct {
	Title     string
	Blurb     string
	Body      string
	DB        *db.DB
	UserEmail string
	Id        int
	WorkId    int
	Order     int64
	Snippet   bool
	WordCount int
}

func NewSection(title, blurb, body, workId, email string) *Section {
	i, _ := strconv.Atoi(workId)
	return &Section{
		Title:     title,
		Blurb:     blurb,
		Body:      body,
		WorkId:    i,
		UserEmail: email,
		WordCount: 0,
	}
}

func (s *Section) VerifyPermission(sm sessionManager.SessionManager) bool {
	return s.UserEmail == sm.GetUserEmail()
}

func (s *Section) GetInsertStr() string {
	return `
INSERT INTO tbl_section(title, blurb, body, work_id, section_order, user_email, is_snippet, word_count)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning section_id;`
}

func (s *Section) GetInsertArgs() []interface{} {
	return []interface{}{s.Title, db.ToNullString(s.Blurb), db.ToNullString(s.Body), s.WorkId, db.ToNullInt(s.Order), s.UserEmail, s.Snippet, s.WordCount}
}

func (s *Section) GetUpdateStr() string {
	return `
UPDATE tbl_section
SET title=$1, blurb=$2, body=$3, section_order=$4, is_snippet=$5, word_count=$6
WHERE section_id=$7
`
}

func (s *Section) GetUpdateArgs() []interface{} {
	return []interface{}{s.Title, s.Blurb, s.Body, s.Order, s.Snippet, s.WordCount, s.Id}
}

func (s *Section) Save(tx *sql.Tx) error {
	return s.DB.Update(s, tx)
}

func GetSectionById(id int, database *db.DB) Verifiable {
	query := sectionByIdQuery{Id: id}
	sectionInt, err := database.Query(query)
	if err != nil {
		glog.Error(err)
		return nil
	}
	if len(sectionInt) == 0 {
		return nil
	}
	return sectionInt[0].(*Section)
}

type sectionByIdQuery struct {
	Id int
}

func (q sectionByIdQuery) GetQueryStr() string {
	return sectionDetailColumnStr + " where section_id=$1"
}

func (q sectionByIdQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.Id}
}

func (q sectionByIdQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	return sectionDetailFromRow(db, r)
}

func sectionListFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	section := Section{DB: db}
	nullBlurb := sql.NullString{}
	nullOrder := sql.NullInt64{}
	if err := r.Scan(&section.Id, &section.Title, &nullBlurb, &section.UserEmail, &section.WorkId, &nullOrder, &section.Snippet, &section.WordCount); err != nil {
		glog.Error(err.Error())
		return nil, err
	}
	section.Blurb = nullBlurb.String
	section.Order = nullOrder.Int64
	return &section, nil
}

func sectionDetailFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	section := Section{DB: db}
	nullBlurb := sql.NullString{}
	nullBody := sql.NullString{}
	nullOrder := sql.NullInt64{}
	if err := r.Scan(&section.Id, &section.Title, &nullBlurb, &nullBody, &section.UserEmail, &section.WorkId, &nullOrder, &section.Snippet, &section.WordCount); err != nil {
		glog.Error(err.Error())
		return nil, err
	}
	section.Blurb = nullBlurb.String
	section.Body = nullBody.String
	section.Order = nullOrder.Int64
	return &section, nil
}

func GetSectionsForWork(workId int, db *db.DB) ([]*Section, []*Section) {
	query := sectionsForWorkQuery{Id: workId}
	sectionInt, err := db.Query(query)
	if err != nil {
		fmt.Printf("Error on GetSectionsForWork: %v", err.Error())
		return nil, nil
	}
	allSections := make([]*Section, len(sectionInt))
	for i := range sectionInt {
		allSections[i] = sectionInt[i].(*Section)
	}
	sections := []*Section{}
	snippets := []*Section{}
	for i := range allSections {
		if allSections[i].Snippet == true {
			snippets = append(snippets, allSections[i])
		} else {
			sections = append(sections, allSections[i])
		}
	}
	slice.Sort(sections, func(i, j int) bool {
		return sections[i].Order < sections[j].Order
	})
	return sections, snippets
}

type sectionsForWorkQuery struct {
	Id int
}

func (q sectionsForWorkQuery) GetQueryStr() string {
	return `
select section_id, title, blurb, section_order, is_snippet, word_count from tbl_section where work_id=$1`
}

func (q sectionsForWorkQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.Id}
}

func (q sectionsForWorkQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	section := Section{DB: db}
	nullBlurb := sql.NullString{}
	nullOrder := sql.NullInt64{}
	if err := r.Scan(&section.Id, &section.Title, &nullBlurb, &nullOrder, &section.Snippet, &section.WordCount); err != nil {
		glog.Error(err.Error())
		return nil, err
	}
	section.Blurb = nullBlurb.String
	section.Order = nullOrder.Int64
	return &section, nil
}

func GetSectionsForCharacter(characterId int, database *db.DB) []*Section {
	query := leftForRightQuery{
		RightName:      "character",
		DB:             database,
		LeftName:       "section",
		RightId:        characterId,
		ColumnStr:      sectionListColumnStr,
		ObjFromRowFunc: sectionListFromRow,
	}
	sectionsInt := getLeftForRight(query)
	output := make([]*Section, len(sectionsInt))
	for i := range sectionsInt {
		output[i] = sectionsInt[i].(*Section)
	}
	return output
}

func GetSectionsForSetting(settingId int, database *db.DB) []*Section {
	query := leftForRightQuery{
		RightName:      "setting",
		DB:             database,
		LeftName:       "section",
		RightId:        settingId,
		ColumnStr:      sectionListColumnStr,
		ObjFromRowFunc: sectionListFromRow,
	}
	sectionsInt := getLeftForRight(query)
	output := make([]*Section, len(sectionsInt))
	for i := range sectionsInt {
		output[i] = sectionsInt[i].(*Section)
	}
	return output
}

func DeleteSection(sectionId int, database *db.DB) (bool, error) {
	return db.DoBasicDelete(sectionId, "section", database)
}

func ReorderSectionsFromFormValue(rawOrder string, workId int, tx *sql.Tx) error {
	splitOrder := strings.Split(rawOrder, ",")
	updateStr := `UPDATE tbl_section SET section_order = mt.section_order
	FROM (values `
	updateArgs := make([]interface{}, len(splitOrder)*2)
	for i := range splitOrder {
		nForSql := i*2 + 1
		updateStr += fmt.Sprintf("($%v::integer, $%v::integer)", nForSql, nForSql+1)
		if i < len(splitOrder)-1 {
			updateStr += ", "
		}
		idAndOrder := strings.Split(splitOrder[i], "-")
		updateArgs[i*2] = idAndOrder[0]
		updateArgs[i*2+1] = idAndOrder[1]
	}
	updateStr += `
	) AS mt(section_id, section_order)
	WHERE mt.section_id::integer=tbl_section.section_id::integer `
	updateStr += fmt.Sprintf("AND tbl_section.work_id::integer=$%v", len(splitOrder)*2+1)
	updateArgs = append(updateArgs, workId)
	stmt, err := tx.Prepare(updateStr)
	if err != nil {
		glog.Errorf("Reorder database error: %v", err.Error())
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(updateArgs...)
	if err != nil {
		glog.Error("Transaction error, rollback: %v", err.Error())
		tx.Rollback()
		return err
	}
	return nil
}
