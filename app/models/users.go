package models

import (
	"database/sql"

	"bitbucket.org/jtyburke/pathfork/app/auth"
	"bitbucket.org/jtyburke/pathfork/app/db"
	"github.com/golang/glog"
)

type User struct {
	Email     string
	Password  string
	DB        *db.DB
	Confirmed bool
	Verified  bool
}

func NewUser(email, pw string) (*User, error) {
	password, err := auth.HashPassword(pw)
	if err != nil {
		return nil, err
	}
	user := &User{
		Email:    email,
		Password: password,
	}
	return user, nil
}

func (u *User) GetInsertStr() string {
	return "INSERT INTO tbl_user(email, pw) VALUES ($1, $2) returning 0;"
}

func (u *User) GetInsertArgs() []interface{} {
	return []interface{}{u.Email, u.Password}
}

func GetUserByEmail(email string, db *db.DB) *User {
	query := userByEmailQuery{Email: email}
	userInt, err := db.Query(query)
	if err != nil {
		glog.Error(err.Error())
		return nil
	}
	if len(userInt) == 0 {
		return nil
	}
	user := userInt[0].(*User)
	return user
}

type userByEmailQuery struct {
	Email string
}

func (q userByEmailQuery) GetQueryStr() string {
	return "select * from tbl_user where email=$1"
}

func (q userByEmailQuery) GetQueryArgs() []interface{} {
	return []interface{}{q.Email}
}

func (q userByEmailQuery) ObjFromRow(db *db.DB, r *sql.Rows) (db.Insertable, error) {
	user := User{DB: db}
	if err := r.Scan(&user.Email, &user.Password, &user.Verified); err != nil {
		return nil, err
	}
	return &user, nil
}

func VerifyUser(u *User, tx *sql.Tx) error {
	stmt, err := tx.Prepare("Update tbl_user set verified=true where email=$1")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.Email)
	if err != nil {
		glog.Error("Transaction error, rollback: %v", err.Error())
		tx.Rollback()
		return err
	}
	return nil
}

func UpdatePassword(u *User, password string, tx *sql.Tx) error {
	stmt, err := tx.Prepare("Update tbl_user set pw=$1 where email=$2")
	if err != nil {
		return err
	}
	hashed, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(hashed, u.Email)
	if err != nil {
		glog.Error("Transaction error, rollback: %v", err.Error())
		tx.Rollback()
		return err
	}
	return nil
}
