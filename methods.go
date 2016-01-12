package vimages

import (
	"fmt"
	"strings"
)

func post(r *Request) *Response {
	d, err := newObject(r.tx, r.s, r.GetParam("overwrite") == "true", r.Path, r.Checksum, r.Header)
	if err != nil {
		switch err {
		case errDenied:
			return respAccessDenied
		case errCollission:
			return respCollission
		default:
			return respInternalError
		}
	}

	resp := new(Response)
	resp.Delete = d
	return resp
}

func put(r *Request) *Response {
	obj, err := getObject(r.tx, r.s, r.Path)
	if err != nil {
		switch err {
		case errExist:
			return respBadTarget
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	}

	switch r.GetParam("command") {
	case "chmod":
		val, err := r.s.Write(obj, "CHMOD")
		if err != nil {
			switch err {
			case errDenied:
				return respAccessDenied
			default:
				return respInternalError
			}
		}
		if val == nil {
			return new(Response)
		}
		switch val.(error) {
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	case "chown":
		val, err := r.s.Write(obj, "CHOWN")
		if err != nil {
			switch err {
			case errDenied:
				return respAccessDenied
			default:
				return respInternalError
			}
		}
		if val == nil {
			return new(Response)
		}
		switch val.(error) {
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	case "chgrp":
		val, err := r.s.Write(obj, "CHGRP")
		if err != nil {
			switch err {
			case errDenied:
				return respAccessDenied
			default:
				return respInternalError
			}
		}
		if val == nil {
			return new(Response)
		}
		switch val.(error) {
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	default:
	}

	switch obj.otype() {
	case "folder":
		return respBadPut
	case "image":
		val, ok := r.Header["Author"]
		if ok {
			val, err := r.s.Write(obj, "AUTHOR", val[0])
			if err != nil {
				switch err {
				case errDenied:
					return respAccessDenied
				default:
					return respInternalError
				}
			}
			if val != nil {
				if val.(error) != nil {
					switch err {
					case errDenied:
						return respAccessDenied
					default:
						return respInternalError
					}
				}
			}
		}

		val, ok = r.Header["Description"]
		if ok {
			val, err := r.s.Write(obj, "DESCRIPTION", val[0])
			if err != nil {
				switch err {
				case errDenied:
					return respAccessDenied
				default:
					return respInternalError
				}
			}
			if val != nil {
				if val.(error) != nil {
					switch err {
					case errDenied:
						return respAccessDenied
					default:
						return respInternalError
					}
				}
			}
		}

		val, ok = r.Header["Date"]
		if ok {
			val, err := r.s.Write(obj, "DATE", val[0])
			if err != nil {
				switch err {
				case errDenied:
					return respAccessDenied
				default:
					return respInternalError
				}
			}
			if val != nil {
				if val.(error) != nil {
					switch err {
					case errDenied:
						return respAccessDenied
					default:
						return respInternalError
					}
				}
			}
		}

		return new(Response)

	default:
	}

	return respInternalError
}

func get(r *Request) *Response {
	obj, err := getObject(r.tx, r.s, r.Path)
	if err != nil {
		switch err {
		case errExist:
			return respBadTarget
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	}

	//
	switch obj.otype() {
	case "image":
		attr := ""
		if r.GetParam("attributes") == "true" {
			attr = "attributes"
		}
		val, err := r.s.Read(obj, attr)
		if err != nil {
			switch err {
			case errDenied:
				return respAccessDenied
			default:
				return respInternalError
			}
		}

		err, ok := val.(error)
		if ok {
			switch err {
			case errDenied:
				return respAccessDenied
			default:
				return respInternalError
			}
		}

		resp := new(Response)
		chk, ok := val.(string)
		if ok {
			resp.Send = chk
		} else {
			resp.Response = val.(imgData)
		}
		return resp

	case "folder":
		val, err := r.s.Read(obj, r.GetParam("format"))
		if err != nil {
			switch err {
			case errDenied:
				return respAccessDenied
			default:
				return respInternalError
			}
		}

		err, ok := val.(error)
		if ok {
			switch err {
			case errDenied:
				return respAccessDenied
			default:
				return respInternalError
			}
		}
		resp := new(Response)
		resp.Response = val
		return resp
	default:
		return respInternalError
	}

}

func delete(r *Request) *Response {
	i := strings.LastIndex(r.Path, "/")
	path := r.Path[:i]

	obj, err := getObject(r.tx, r.s, path)
	if err != nil {
		switch err {
		case errExist:
			return respBadTarget
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	}

	//
	deleted, err := r.s.Write(obj, "DELETE", r.Path[i+1:])
	if err != nil {
		if err.Error() == "access denied" {
			return respAccessDenied
		}
		switch err {
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	}

	err, ok := deleted.(error)
	if ok {
		fmt.Println(err)
		switch err {
		case errDenied:
			return respAccessDenied
		default:
			return respInternalError
		}
	}

	resp := new(Response)
	if deleted != nil {
		resp.Delete = deleted.([]string)
	}

	return resp

}
