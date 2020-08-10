package router

import (
	"github.com/gorilla/mux"
	"lataleBotService/handler"
)

func NewRouter(handler *handler.Funcs) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes(handler) {
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	return router
}