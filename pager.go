package main

import (
	"context"
	"net/http"
	"strconv"
)

const (
	DefaultOffset = 0
	DefaultLimit  = 20
	offsetKey     = "offset"
	limitKey      = "limit"
)

func pager(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			limit, offset int
			err           error
		)
		limitParam := r.URL.Query().Get("limit")
		if limitParam == "" {
			limit = DefaultLimit
		} else if limit, err = strconv.Atoi(limitParam); err != nil {
			limit = DefaultLimit
		}
		ctx := context.WithValue(r.Context(), "limit", limit)

		offsetParam := r.URL.Query().Get("offset")
		if offsetParam == "" {
			offset = DefaultOffset
		} else if offset, err = strconv.Atoi(offsetParam); err != nil {
			offset = DefaultOffset
		}
		ctx = context.WithValue(ctx, "offset", offset)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func page(r *http.Request) (offset int, limit int) {
	ctx := r.Context()
	offset, _ = ctx.Value(offsetKey).(int)
	limit, _ = ctx.Value(limitKey).(int)
	return
}
