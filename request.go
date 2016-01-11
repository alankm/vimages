package vimages

import (
	"database/sql"
	"net/url"

	"github.com/alankm/privileges"
)

var emptyPayloadChecksum = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

type Request struct {
	tx     *sql.Tx
	s      *privileges.Session
	params map[string][]string

	Username string
	Hashword string
	Method   string
	Path     string
	URL      string
	Checksum string
	Header   map[string][]string
}

func (r *Request) Empty() bool {

	return r.Checksum == emptyPayloadChecksum

}

func (r *Request) GetHeader(header string) string {

	val, _ := r.Header[header]
	if val == nil || len(val) == 0 {
		return ""
	}
	return val[0]

}

func (r *Request) GetParam(param string) string {

	if r.params == nil {
		url, err := url.Parse(r.URL)
		if err != nil {
			return ""
		}

		r.params = url.Query()
	}

	val, _ := r.params[param]
	if val == nil || len(val) == 0 {
		return ""
	}
	return val[0]

}
