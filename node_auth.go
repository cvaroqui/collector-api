package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type authNode struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Nodename  string         `gorm:"column:nodename; index" json:"nodename"`
	NodeID    string         `gorm:"column:node_id; uniqueIndex; size:36" json:"node_id"`
	UUID      string         `gorm:"column:uuid; size:36" json:"uuid"`
}

func nodeAuth(expire time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				nodeAuthFailed(w)
				return
			}
			data := make([]authNode, 0)
			result := db.Table("auth_node").Where("nodename = ? and uuid = ?", user, pass).Find(&data)
			if result.Error != nil {
				nodeAuthFailed(w)
				return
			}
			if len(data) == 0 {
				nodeAuthFailed(w)
				return
			}
			ctx := context.WithValue(r.Context(), "authNode", data[0])
			ctx = context.WithValue(ctx, "exp", expire)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getNodeToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exp, _ := ctx.Value("exp").(time.Duration)
	data, _ := ctx.Value("authNode").(authNode)
	n, _ := getNodeByNodeID(data.NodeID)
	expireAt := time.Now().Add(exp)
	claims := map[string]interface{}{
		"user_id":    data.NodeID,
		"exp":        expireAt.Unix(),
		"authorized": true,
		"nodename":   data.Nodename,
		"node_id":    data.NodeID,
		"app":        n.App,
	}
	_, tokenString, err := tokenAuth.Encode(claims)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	jsonEncode(w, map[string]interface{}{
		"expire_at": expireAt,
		"token":     tokenString,
	})
}

func nodeAuthFailed(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="nodes"`))
	w.WriteHeader(http.StatusUnauthorized)
}
