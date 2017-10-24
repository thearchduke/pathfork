package models

import (
	"database/sql"
	"fmt"

	"bitbucket.org/jtyburke/pathfork/app/db"
	"github.com/bradfitz/slice"
)

type sectionDetailForExportQuery struct {
	WorkId int
}

func (q sectionDetailForExportQuery) GetQueryStr() string {
	return sectionDetailColumnStr + " where tbl_section.work_id=$1"
}

func (q sectionDetailForExportQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.WorkId}
}

func (q sectionDetailForExportQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	return sectionDetailFromRow(db, r)
}

func GetSectionDetailForExport(workId int, database *db.DB) ([]*Section, []*Section) {
	query := sectionDetailForExportQuery{WorkId: workId}
	sectionInt, err := database.Query(query)
	if err != nil {
		fmt.Printf("Error on GetSectionDetailForExport: %v", err.Error())
		return nil, nil
	}
	allSections := make([]*Section, len(sectionInt))
	for i := range sectionInt {
		allSections[i] = sectionInt[i].(*Section)
	}
	slice.Sort(allSections, func(i, j int) bool {
		if allSections[i].Snippet == true && allSections[j].Snippet == false {
			return false
		}
		return true
	})
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

func GetCharactersForWorkExport(workId int, database *db.DB) []*Character {
	return getCharactersForLeft("work", workId, database, "detail")
}

func GetSettingsForWorkExport(workId int, database *db.DB) []*Setting {
	return getSettingsForLeft("work", workId, database, "detail")
}
