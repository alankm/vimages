package vimages

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/alankm/privileges"
)

var vimages *Vimages

func TestMain(m *testing.M) {
	os.Remove("./test.db")
	os.Remove("./priv.db")
	defer os.Remove("./test.db")
	defer os.Remove("./priv.db")

	priv, _ := privileges.New("./priv.db")
	vimages, _ = New("./test.db", priv)

	m.Run()
}

func printDB(tx *sql.Tx) {
	fmt.Println("\n\nDB:")
	rows, _ := tx.Query("SELECT * FROM vimages")
	for rows.Next() {
		rec, _ := recordFromRows(rows)
		fmt.Println(rec.path)
	}
	fmt.Println("")
}

func TestPost000(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects/a.img"
	req.params = make(map[string][]string)
	req.Header = make(map[string][]string)
	req.Method = "POST"

	post(req)
	tx.Commit()
}

func TestPost001(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects/a.img/b.img"
	req.params = make(map[string][]string)
	req.Header = make(map[string][]string)
	req.Method = "POST"

	resp := post(req)
	if resp != respCollission {
		t.Error(nil)
	}
	tx.Commit()
}

func TestPost002(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects/a.img/b.img"
	req.params = make(map[string][]string)
	req.params["overwrite"] = make([]string, 1)
	req.params["overwrite"][0] = "true"
	req.Header = make(map[string][]string)
	req.Method = "POST"

	post(req)
	tx.Commit()
}

func TestPost003(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects"
	req.params = make(map[string][]string)
	req.Header = make(map[string][]string)
	req.Method = "POST"

	resp := post(req)
	if resp != respCollission {
		t.Error(nil)
	}
	tx.Commit()
}

func TestPost004(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects"
	req.params = make(map[string][]string)
	req.params["overwrite"] = make([]string, 1)
	req.params["overwrite"][0] = "true"
	req.Header = make(map[string][]string)
	req.Method = "POST"

	post(req)
	tx.Commit()
}

func TestPost005(t *testing.T) {

	s, _ := vimages.users.Login("guest", "")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects"
	req.params = make(map[string][]string)
	req.params["overwrite"] = make([]string, 1)
	req.params["overwrite"][0] = "true"
	req.Header = make(map[string][]string)
	req.Method = "POST"

	resp := post(req)
	if resp != respAccessDenied {
		t.Error(nil)
	}
	tx.Commit()
}

func TestPost006(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects"
	req.params = make(map[string][]string)
	req.params["overwrite"] = make([]string, 1)
	req.params["overwrite"][0] = "true"
	req.Header = make(map[string][]string)
	req.Method = "POST"

	post(req)
	tx.Commit()
}

func TestDelete000(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users/alankm/projects"
	req.params = make(map[string][]string)
	req.Header = make(map[string][]string)
	req.Method = "DELETE"

	delete(req)
	tx.Commit()
}

func TestDelete001(t *testing.T) {

	s, _ := vimages.users.Login("guest", "")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users"
	req.params = make(map[string][]string)
	req.Header = make(map[string][]string)
	req.Method = "DELETE"

	resp := delete(req)
	switch resp {
	case respAccessDenied:
		fmt.Println("A")
	case respInternalError:
		fmt.Println("B")
	default:
		fmt.Println("C")
	}

	if resp != respAccessDenied {
		fmt.Println(resp)
		t.Error(nil)
	}
	tx.Commit()
}

func TestDelete002(t *testing.T) {

	s, _ := vimages.users.Login("root", "guest")
	req := new(Request)
	req.s = s
	tx, _ := vimages.db.Begin()
	req.tx = tx
	req.Path = "/users"
	req.params = make(map[string][]string)
	req.Header = make(map[string][]string)
	req.Method = "DELETE"

	delete(req)
	tx.Commit()
}
