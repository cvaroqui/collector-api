package main

import "net/http"

//
// GetNode     godoc
// @Summary      Show a node
// @Description  Show a node by index, id or name
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         nodes
// @Accept       json
// @Produce      json
// @Success      200  {array}   Node
// @Failure      404  {string}  string  "Not Found"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /nodes/{id}  [get]
//
func getNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("node").(Node)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []Node{n})
}
