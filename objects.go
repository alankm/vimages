package vimages

import (
	"database/sql"
	"strings"

	"github.com/alankm/privileges"
)

type image struct {
}

type folder struct {
	tx      *sql.Tx
	s       *privileges.Session
	path    string
	rules   *privileges.Rules
	command string
	name    string
}

func newFolder(tx *sql.Tx, s *privileges.Session, path string, rules *privileges.Rules) *folder {
	f := new(folder)
	f.tx = tx
	f.path = path
	f.rules = rules
	return f
}

func (f *folder) createFolder(name string) {
	f.command = "New Folder"
	f.name = name
}

func (f *folder) createImage(name string) {
	f.command = "New Image"
	f.name = name
}

func (f *folder) deleteObject(name string) {
	f.command = "Delete"
	f.name = name
}

func (f *folder) Rules() *privileges.Rules {
	return f.rules
}

// Read on a folder returns a map of all children and their attributes
func (f *folder) Read() interface{} {

	if !f.s.CanExec(f) {
		return errDenied
	}

	rows, err := f.tx.Query("SELECT * FROM vimages WHERE path LIKE ? AND path NOT LIKE ?", f.path+"/%", f.path+"/%/")
	if err != nil {
		return err
	}
	defer rows.Close()

	read := make(map[string]*record)

	for rows.Next() {
		rec, err := recordFromRows(rows)
		if err != nil {
			return err
		}
		vals := strings.Split(rec.path, "/")
		path := vals[len(vals)-1]
		rec.path = path
		read[path] = rec
	}

	return read

}

func (f *folder) Write() interface{} {

	if !f.s.CanExec(f) {
		return errDenied
	}

	switch f.command {
	case "New Folder":
	case "New Image":
	case "Delete":
	default:
	}

	//
	// TODO: implement write
	//
	return nil
}

// Exec on a folder returns a map of all children and their Privileged objects
func (f *folder) Exec() interface{} {

	rows, err := f.tx.Query("SELECT * FROM vimages WHERE path LIKE ? AND path NOT LIKE ?", f.path+"/%", f.path+"/%/")
	if err != nil {
		return err
	}
	defer rows.Close()

	exec := make(map[string]privileges.Privileged)
	for rows.Next() {
		rec, err := recordFromRows(rows)
		if err != nil {
			return err
		}
		vals := strings.Split(rec.path, "/")
		path := vals[len(vals)-1]
		rules, err := privileges.NewRules(rec.owner, rec.group, rec.rules)
		if err != nil {
			return err
		}
		if rec.dir {
			exec[path] = &folder{
				tx:    f.tx,
				s:     f.s,
				path:  path,
				rules: rules,
			}
		} else {
			//  TODO: exec[rec.path] = .....
		}

	}

	return exec
}
