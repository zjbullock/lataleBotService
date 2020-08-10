package router

import (
	"lataleBotService/handler"
	"net/http"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var (
	graphql = "/graphql"
)

func routes(handler *handler.Funcs) []route {
	return []route{
		{
			Name:        "GraphQL",
			Method:      http.MethodPost,
			Pattern:     graphql,
			HandlerFunc: handler.GraphQL,
		},
	}
}