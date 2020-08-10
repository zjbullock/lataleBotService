package handler

import (
	"context"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"net/http"
)

type Funcs struct {
	Ctx      context.Context
	Schema   *graphql.Schema
}

func (h *Funcs) GraphQL(w http.ResponseWriter, r *http.Request) {
	server := &relay.Handler{Schema: h.Schema}
	server.ServeHTTP(w, r.WithContext(h.Ctx))
}