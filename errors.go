package vimages

import "errors"

var (
	errDenied = errors.New("access denied")
	errMethod = errors.New("unsupported method")
)
