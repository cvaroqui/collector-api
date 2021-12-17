package routes

import (
	"fmt"
	"net/http"

	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
)

func GetNodeTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("nodeTag").(tables.NodeTag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []tables.NodeTag{n})
}

func GetNodesTags(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("node_tags").Request()
	rq.AutoJoin("nodes")
	td, err := rq.MakeReadTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func GetNodeTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, _ := ctx.Value("node").(tables.Node)
	rq := db.Tab("node_tags").Request()
	rq.AutoJoin("nodes")
	rq.Where("nodes.id = ?", n.ID)
	td, err := rq.MakeReadTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func GetNodeCandidateTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, _ := ctx.Value("node").(tables.Node)
	exclude := db.DB().Table("node_tags").Where("node_id = ?", n.NodeID).Select("tag_id")

	rq := db.Tab("tags").Request()
	rq.Where("tags.tag_id NOT IN (?)", exclude)
	td, err := rq.MakeReadTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func GetTagNodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tag, _ := ctx.Value("tag").(tables.Tag)
	rq := db.Tab("nodes").Request()
	rq.AutoJoin("tags")
	rq.Where("node_tags.tag_id = ?", tag.TagID)
	td, err := rq.MakeReadTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
