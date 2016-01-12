package vimages

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/alankm/privileges"
)

var (
	errBadWrite   = errors.New("call to write without proper setup")
	errInternal   = errors.New("internal error")
	errExist      = errors.New("object doesn't exist")
	errBadDate    = errors.New("date must be an unsigned integer")
	errCollission = errors.New("collission detected with no overwrite set")
)

type shortData struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type longData struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Rules string `json:"permissions"`
}

type object interface {
	Rules() *privileges.Rules
	Read(...string) interface{}
	Write(...string) interface{}
	Exec(...string) interface{}

	otype() string
	delete() ([]string, error)
}

func newObject(tx *sql.Tx, s *privileges.Session, overwrite bool, path, checksum string, headers map[string][]string) ([]string, error) {

	var deleted []string
	rules, _ := privileges.NewRules("root", "root", "0755")
	var obj object = newFolder(tx, s, "", rules)
	route := strings.Split(path[1:], "/")
	ctype := "folder"

	var date uint64
	a, ok := headers["Date"]
	if ok {
		date, _ = strconv.ParseUint(a[0], 10, 64)
	}

	author := ""
	a, ok = headers["Author"]
	if ok {
		author = a[0]
	}

	description := ""
	a, ok = headers["Description"]
	if ok {
		description = a[0]
	}

	for i, o := range route {

		if i == len(route)-1 && checksum != emptyPayloadChecksum {
			ctype = "image"
		}

		val, err := s.Exec(obj)
		if err != nil {
			return nil, err
		}

		children, ok := val.(map[string]object)
		if !ok {
			return nil, errInternal
		}

		gid, _ := s.Gid("", "")
		mod, _ := s.Umask("")
		rule, _ := privileges.NewRules(s.User, gid, mod)
		path := ""
		for j := 0; j <= i; j++ {
			path = path + "/" + route[j]
		}

		if children[o] == nil { // make an entry
			if !s.CanWrite(obj) {
				return nil, errDenied
			}
			if ctype == "folder" {
				folder := newFolder(tx, s, path, rule)
				children[o] = folder
				_, err = tx.Exec("INSERT INTO vimages VALUES(?,?,?,?,?,?,?,?,?)", path, true, s.User, gid, mod, "", 0, "", "")
				if err != nil {
					return nil, err
				}
			} else { // image
				_, err = tx.Exec("INSERT INTO vimages VALUES(?,?,?,?,?,?,?,?,?)", path, false, s.User, gid, mod, checksum, date, author, description)
				if err != nil {
					return nil, err
				}
				return deleted, nil
			}

		} else if children[o].otype() != ctype { // overwrite?
			if !overwrite {
				return nil, errCollission
			}
			if !s.CanWrite(obj) {
				return nil, errDenied
			}
			// overwrite EVERYTHING
			d, err := children[o].delete()
			if err != nil {
				return nil, err
			}
			deleted = append(deleted, d...)

			if ctype == "folder" {
				folder := newFolder(tx, s, path, rule)
				children[o] = folder
				_, err = tx.Exec("INSERT INTO vimages VALUES(?,?,?,?,?,?,?,?,?)", path, true, s.User, gid, mod, checksum, date, author, description)
				if err != nil {
					return nil, err
				}
			} else { // image
				_, err = tx.Exec("INSERT INTO vimages VALUES(?,?,?,?,?,?,?,?,?)", path, false, s.User, gid, mod, checksum, date, author, description)
				if err != nil {
					return nil, err
				}
				return deleted, nil
			}

		} else if i == len(route)-1 { // overwrite if final image
			if !overwrite {
				return nil, errCollission
			}
			if s.CanWrite(children[o]) {
				d, err := children[o].delete()
				if err != nil {
					return nil, err
				}
				deleted = append(deleted, d...)
				_, err = tx.Exec("DELETE FROM vimages WHERE path=?", path)
				if err != nil {
					return nil, err
				}
				_, err = tx.Exec("INSERT INTO vimages VALUES(?,?,?,?,?,?,?,?,?)", path, false, s.User, gid, mod, checksum, date, author, description)
				if err != nil {
					return nil, err
				}
				return deleted, nil

			} else {
				return nil, errDenied
			}
		}

		obj = children[o]

	}

	return deleted, nil

}

func getObject(tx *sql.Tx, s *privileges.Session, path string) (object, error) {
	rules, _ := privileges.NewRules("root", "root", "0755")
	var obj object = newFolder(tx, s, "", rules)
	if len(path) < 2 {
		return obj, nil
	}
	route := strings.Split(path[1:], "/")
	for _, o := range route {
		val, err := s.Exec(obj)
		if err != nil {
			return nil, err
		}

		children, ok := val.(map[string]object)
		if !ok {
			return nil, errInternal
		}

		obj = children[o]
		if obj == nil {
			return nil, errExist
		}
	}

	return obj, nil

}

type folder struct {
	tx    *sql.Tx
	s     *privileges.Session
	path  string
	rules *privileges.Rules
}

func newFolder(tx *sql.Tx, s *privileges.Session, path string, rules *privileges.Rules) *folder {
	f := new(folder)
	f.s = s
	f.tx = tx
	f.path = path
	f.rules = rules
	return f
}

func (f *folder) otype() string {
	return "folder"
}

func (f *folder) Rules() *privileges.Rules {
	return f.rules
}

func (f *folder) Read(args ...string) interface{} {

	if !f.s.CanExec(f) {
		return errDenied
	}

	ret, err := f.children()
	if err != nil {
		return err
	}

	var names []string
	for _, child := range ret {
		splits := strings.Split(child.path, "/")
		names = append(names, splits[len(splits)-1])
	}

	sort.Strings(names)

	if args != nil && len(args) > 0 && args[0] == "long" {
		var list []longData
		for _, name := range names {
			rec := ret[name]
			t := "image"
			if rec.dir {
				t = "folder"
			}
			rule, _ := privileges.NewRules(rec.owner, rec.group, rec.rules)
			list = append(list, longData{
				Name:  name,
				Type:  t,
				Rules: rule.Symbolic(rec.dir),
			})
		}
		return list
	} else {
		var list []shortData
		for _, name := range names {
			rec := ret[name]
			t := "image"
			if rec.dir {
				t = "folder"
			}
			list = append(list, shortData{
				Name: name,
				Type: t,
			})
		}
		return list
	}

}

func (f *folder) Write(args ...string) interface{} {

	if !f.s.CanExec(f) {
		return errDenied
	}

	if args == nil || len(args) != 2 {
		return errBadWrite
	}

	switch args[0] {
	case "CHMOD":
		if f.s.CanChown(f.rules, args[1]) {
			f.tx.Exec("UPDATE vimages SET permissions=? WHERE path=?", args[1], f.path)
			return nil
		}
		return errDenied
	case "CHOWN":
		if f.s.CanChown(f.rules, args[1]) {
			f.tx.Exec("UPDATE vimages SET owner_user=? WHERE path=?", args[1], f.path)
			return nil
		}
		return errDenied
	case "CHGRP":
		if f.s.CanChgrp(f.rules, args[1]) {
			f.tx.Exec("UPDATE vimages SET owner_group=? WHERE path=?", args[1], f.path)
			return nil
		}
		return errDenied
	case "DELETE":
		exec := f.Exec("")
		children, ok := exec.(map[string]object)
		if !ok {
			return exec
		}

		fmt.Println("DELETING...")
		fmt.Println(args[1])
		fmt.Println("")

		target := children[args[1]]
		if target == nil {
			return errBadWrite
		}

		deleted, err := target.delete()
		if err != nil {
			return err
		}

		return deleted
	default:
		return errBadWrite
	}

}

func (f *folder) Exec(args ...string) interface{} {

	records, err := f.children()
	if err != nil {
		return err
	}

	exec := make(map[string]object)
	for _, rec := range records {
		vals := strings.Split(rec.path, "/")
		path := vals[len(vals)-1]
		rules, err := privileges.NewRules(rec.owner, rec.group, rec.rules)
		if err != nil {
			return err
		}
		if rec.dir {
			exec[path] = newFolder(f.tx, f.s, rec.path, rules)
		} else {
			exec[path] = newImage(f.tx, f.s, rec.path, rec.chk, rules, rec)
		}
	}

	return exec

}

func (f *folder) delete() ([]string, error) {

	var deleted []string
	children, err := f.s.Exec(f)
	if err != nil {
		return nil, err
	}

	for _, child := range children.(map[string]object) {
		d, err := child.delete()
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, d...)
	}

	_, err = f.tx.Exec("DELETE FROM vimages WHERE path=?", f.path)
	if err != nil {
		return nil, err
	}

	return deleted, nil

}

func (f *folder) children() (map[string]*record, error) {

	rows, err := f.tx.Query("SELECT * FROM vimages WHERE path LIKE ? AND path NOT LIKE ?", f.path+"/%", f.path+"/%/%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ret = make(map[string]*record)
	for rows.Next() {
		rec, err := recordFromRows(rows)
		if err != nil {
			return nil, err
		}
		ret[rec.path] = rec
	}

	return ret, nil

}

type imgData struct {
	Path        string `json:"path"`
	Date        uint64 `json:"date"`
	Checksum    string `json:"checksum"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

type image struct {
	tx    *sql.Tx
	s     *privileges.Session
	path  string
	rules *privileges.Rules
	chk   string
	rec   *record
}

func newImage(tx *sql.Tx, s *privileges.Session, path, chk string, rules *privileges.Rules, rec *record) *image {
	i := new(image)
	i.tx = tx
	i.path = path
	i.rules = rules
	i.chk = chk
	i.rec = rec
	return i
}

func (i *image) otype() string {
	return "image"
}

func (i *image) Rules() *privileges.Rules {
	return i.rules
}

func (i *image) Read(args ...string) interface{} {
	if args != nil && len(args) > 0 && args[0] == "attributes" {
		var data = imgData{
			Path:        i.rec.path,
			Date:        i.rec.date,
			Checksum:    i.rec.chk,
			Author:      i.rec.auth,
			Description: i.rec.desc,
		}
		return data
	} else {
		return i.chk
	}
}

func (i *image) Write(args ...string) interface{} {

	if args == nil || len(args) != 2 {
		return errBadWrite
	}

	switch args[0] {
	case "CHMOD":
		if i.s.CanChown(i.rules, args[1]) {
			i.tx.Exec("UPDATE vimages SET permissions=? WHERE path=?", args[1], i.path)
			return nil
		}
		return errDenied
	case "CHOWN":
		if i.s.CanChown(i.rules, args[1]) {
			i.tx.Exec("UPDATE vimages SET owner_user=? WHERE path=?", args[1], i.path)
			return nil
		}
		return errDenied
	case "CHGRP":
		if i.s.CanChgrp(i.rules, args[1]) {
			i.tx.Exec("UPDATE vimages SET owner_group=? WHERE path=?", args[1], i.path)
			return nil
		}
		return errDenied
	case "AUTHOR":
		i.tx.Exec("UPDATE vimages SET author=? WHERE path=?", args[1], i.path)
		return nil
	case "DESCRIPTION":
		i.tx.Exec("UPDATE vimages SET description=? WHERE path=?", args[1], i.path)
		return nil
	case "DATE":
		date, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return errBadDate
		}
		i.tx.Exec("UPDATE vimages SET date=? WHERE path=?", date, i.path)
		return nil
	default:
		return errInternal
	}

}

func (i *image) Exec(args ...string) interface{} {
	return nil
}

func (i *image) delete() ([]string, error) {

	var deleted []string

	_, err := i.tx.Exec("DELETE FROM vimages WHERE path=?", i.path)
	if err != nil {
		return nil, err
	}

	rows, err := i.tx.Query("SELECT * FROM vimages WHERE checksum=?", i.chk)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		deleted = append(deleted, i.chk)
	}

	return deleted, nil

}
