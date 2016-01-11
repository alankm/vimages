package vimages

import "database/sql"

type record struct {
	path  string
	dir   bool
	owner string
	group string
	rules string
	chk   string
	date  uint64
	auth  string
	desc  string
}

func recordFromRows(rows *sql.Rows) (*record, error) {
	rec := new(record)
	err := rows.Scan(&rec.path, &rec.dir, &rec.owner, &rec.group, &rec.rules, &rec.chk, &rec.date, &rec.auth, &rec.desc)
	return rec, err
}
