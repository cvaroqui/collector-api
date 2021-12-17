package routes

import (
	"net/http"

	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
)

func DelTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tag, ok := ctx.Value("tag").(tables.Tag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	// TODO: priv check
	db.DB().Where("tags.id = ?", tag.ID).Delete(&tables.Tag{})
	jsonEncode(w, []tables.Tag{tag})
}

func GetTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("tag").(tables.Tag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []tables.Tag{n})
}
