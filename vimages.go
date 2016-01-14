package vimages

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	"github.com/alankm/privileges"
	"github.com/mattn/go-sqlite3"
)

func init() {
	sql.Register("sqlite3_with_fk",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				_, err := conn.Exec("PRAGMA foreign_keys = ON", nil)
				return err
			},
		})
}

type Vimages struct {
	path  string
	db    *sql.DB
	users *privileges.Privileges
}

func New(path string, users *privileges.Privileges) (*Vimages, error) {

	v := new(Vimages)
	v.path = path
	v.users = users
	err := v.setup()
	return v, err

}

func (v *Vimages) setup() error {

	var err error
	v.db, err = sql.Open("sqlite3_with_fk", v.path)
	if err != nil {
		return err
	}

	_, err = v.db.Exec("CREATE TABLE IF NOT EXISTS vimages (" +
		"path VARCHAR(1024) PRIMARY KEY, " +
		"directory BOOLEAN NULL, " +
		"owner_user VARCHAR(64) NULL, " +
		"owner_group VARCHAR(64) NULL, " +
		"permissions VARCHAR(4), " +
		"checksum VARCHAR(64) NULL, " +
		"date INTEGER NULL, " +
		"author VARCHAR(64) NULL, " +
		"description VARCHAR(1024) NULL" +
		");")
	if err != nil {
		v.db.Close()
		return err
	}

	return nil

}

func (v *Vimages) Restore(snapshot []byte) error {

	v.db.Close()
	ioutil.WriteFile(v.path, snapshot, 0775)
	v.db, _ = sql.Open("sqlite3_fk", v.path)
	return v.setup()

}

func (v *Vimages) Snapshot() ([]byte, error) {
	return ioutil.ReadFile(v.path)
}

func (v *Vimages) Do(request *Request) *Response {

	var err error
	request.s, err = v.users.LoginHash(request.Username, request.Hashword)
	if err != nil {
		return respPrivilegesError
	}
	defer request.s.Logout()

	request.tx, err = v.db.Begin()
	if err != nil {
		return respInternalError
	}

	var response *Response

	switch request.Method {
	case "POST":
		response = post(request)
	case "PUT":
		response = put(request)
	case "GET":
		response = get(request)
	case "DELETE":
		response = delete(request)
	default:
		defer request.tx.Rollback()
		return respUnsupportedMethod
	}

	if response.Fail {
		defer request.tx.Rollback()
	} else {
		defer request.tx.Commit()
	}

	fmt.Println(response)

	return response

}
