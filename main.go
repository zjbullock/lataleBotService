package main

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/graph-gophers/graphql-go"
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/handler"
	"lataleBotService/repositories"
	"lataleBotService/resolvers"
	"lataleBotService/router"
	"lataleBotService/server"
	"lataleBotService/services"
	"net/http"
)

var (
	ctx          = context.Background()
	l            loggo.Logger
	handlerFuncs *handler.Funcs
)

func init() {
	l.SetLogLevel(loggo.DEBUG)
	serv := server.NewServer(l)
	schemaString, err := serv.GetSchema("./server/graphql/", l)
	if err != nil {
		l.Criticalf("error occurred while fetching graphql schema: %v", err)
	}
	repos := struct {
		area    repositories.AreasRepository
		classes repositories.ClassRepository
		user    repositories.UserRepository
	}{
		area:    repositories.NewAreaRepo(l, datasource.NewDataSource(l, ctx, globals.PROJECTID)),
		classes: repositories.NewClassRepo(l, datasource.NewDataSource(l, ctx, globals.PROJECTID)),
		user:    repositories.NewUserRepo(l, datasource.NewDataSource(l, ctx, globals.PROJECTID)),
	}
	service := struct {
		Adventure services.Adventure
		Manage    services.Manage
	}{
		Adventure: services.NewAdventureService(repos.area, repos.classes, repos.user, l),
		Manage:    services.NewManageService(repos.area, repos.classes, repos.user, l),
	}
	handlerFuncs = &handler.Funcs{
		Ctx: ctx,
		Schema: graphql.MustParseSchema(schemaString, &resolvers.Resolver{
			Services: service,
			Log:      l,
		}),
	}
}

func main() {
	r := router.NewRouter(handlerFuncs)
	allowedHeaders := handlers.AllowedHeaders([]string{"content-type"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{http.MethodPost})
	l.Criticalf(http.ListenAndServe(":8080", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)).Error())
}
