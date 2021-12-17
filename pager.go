package main

import (
	"net/http"
	"strconv"
)

const (
	DefaultOffset = 0
	DefaultLimit  = 20
	offsetKey     = "offset"
	limitKey      = "limit"
)

func page(r *http.Request) (offset, limit int) {
	var err error
	limitParam := r.URL.Query().Get("limit")
	if limitParam == "" {
		limit = DefaultLimit
	} else if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = DefaultLimit
	}

	offsetParam := r.URL.Query().Get("offset")
	if offsetParam == "" {
		offset = DefaultOffset
	} else if offset, err = strconv.Atoi(offsetParam); err != nil {
		offset = DefaultOffset
	}
	return
}
