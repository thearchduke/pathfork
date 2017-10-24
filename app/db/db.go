package db

import (
	"database/sql"
	"fmt"

	"github.com/golang/glog"
)

type DB struct {
	DB *sql.DB
}

func New() *DB {
	return &DB{}
}

func (db *DB) Open(sourceName string) {
	database, err := sql.Open("postgres", sourceName)
	if err != nil {
		panic(err)
	}
	db.DB = database
}

type Insertable interface {
	GetInsertStr() string
	GetInsertArgs() []interface{}
}

func (db *DB) Insert(i Insertable, tx *sql.Tx) (int, error) {
	var newId int
	insertStr := i.GetInsertStr()
	args := i.GetInsertArgs()
	err := tx.QueryRow(insertStr, args...).Scan(&newId)
	if err != nil && err.Error() != "sql: no rows in result set" {
		glog.Errorf("Insertion error, rollback: %v", err.Error())
		tx.Rollback()
		return 0, err
	}
	return newId, nil
}

type Deleteable interface {
	GetDeleteStr() string
	GetDeleteArgs() []interface{}
}

func (db *DB) Delete(d Deleteable, tx *sql.Tx) error {
	deleteStr := d.GetDeleteStr()
	args := d.GetDeleteArgs()
	stmt, err := tx.Prepare(deleteStr)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(args...)
	if err != nil {
		glog.Error("Transaction error, rollback: %v", err.Error())
		tx.Rollback()
		return err
	}
	return nil
}

type Queryable interface {
	GetQueryStr() string
	GetQueryArgs() []interface{}
	ObjFromRow(*DB, *sql.Rows) (Insertable, error)
}

func (db *DB) Query(q Queryable) ([]Insertable, error) {
	rows, err := db.DB.Query(q.GetQueryStr(), q.GetQueryArgs()...)
	if err != nil {
		glog.Errorf("Error on db.Query: %v", err)
		return nil, err
	}
	defer rows.Close()
	output := make([]Insertable, 0)
	for rows.Next() {
		obj, err := q.ObjFromRow(db, rows)
		if err != nil {
			return nil, err
		}
		output = append(output, obj)
	}
	return output, nil
}

type Updatable interface {
	GetUpdateStr() string
	GetUpdateArgs() []interface{}
}

func (db *DB) Update(u Updatable, tx *sql.Tx) error {
	updateStr := u.GetUpdateStr()
	args := u.GetUpdateArgs()
	stmt, err := tx.Prepare(updateStr)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(args...)
	if err != nil {
		glog.Error("Transaction error, rollback: %v", err.Error())
		tx.Rollback()
		return err
	}
	return nil
}

type basicDelete struct {
	ObjName string
	Id      string
}

func (d basicDelete) GetDeleteStr() string {
	return fmt.Sprintf("DELETE FROM tbl_%v WHERE %v_id=$1", d.ObjName, d.ObjName)
}

func (d basicDelete) GetDeleteArgs() []interface{} {
	return []interface{}{d.Id}
}

func DoBasicDelete(id int, objName string, database *DB) (bool, error) {
	d := basicDelete{
		ObjName: objName,
		Id:      fmt.Sprintf("%v", id),
	}
	tx, err := database.DB.Begin()
	if err != nil {
		glog.Errorf("%v delete database error: %v", objName, err.Error())
		return false, err
	}
	err = database.Delete(d, tx)
	if err != nil {
		glog.Errorf("%v delete database error: %v", objName, err.Error())
		return false, err
	}
	err = tx.Commit()
	return err == nil, err
}
