package main

import (
	. "cloud.google.com/go/firestore"
	"context"
	"github.com/gorilla/handlers"
	"github.com/graph-gophers/graphql-go"
	"github.com/juju/loggo"
	"google.golang.org/api/option"
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
	ds           datasource.Datasource
)

func init() {
	l.SetLogLevel(loggo.DEBUG)
	serv := server.NewServer(l)
	schemaString, err := serv.GetSchema("./server/graphql/", l)
	if err != nil {
		l.Criticalf("error occurred while fetching graphql schema: %v", err)
	}
	client, err := NewClient(ctx, globals.PROJECTID, option.WithCredentialsFile("./credentials-prod.json"), option.WithGRPCConnectionPool(10))
	if err != nil {
		l.Errorf("error initializing Fire Store client with projectId: %s. Received error: %v", globals.PROJECTID, err)
		return
	}
	//ds = datasource.NewDataSource(l, ctx, globals.PROJECTID)
	ds = datasource.NewDataSource(l, ctx, client)
	repos := struct {
		area    repositories.AreasRepository
		classes repositories.ClassRepository
		user    repositories.UserRepository
		levels  repositories.LevelRepository
		equips  repositories.EquipmentRepository
		config  repositories.ConfigRepository
		party   repositories.PartyRepository
	}{
		area:    repositories.NewAreaRepo(l, ds),
		classes: repositories.NewClassRepo(l, ds),
		user:    repositories.NewUserRepo(l, ds),
		levels:  repositories.NewLevelRepo(l, ds),
		equips:  repositories.NewEquipmentRepo(l, ds),
		config:  repositories.NewConfigRepo(l, ds),
		party:   repositories.NewPartiesRepo(l, ds),
	}
	service := struct {
		Adventure services.Adventure
		Manage    services.Manage
		Damage    services.Damage
	}{
		Adventure: services.NewAdventureService(repos.area, repos.classes, repos.user, repos.equips, repos.levels, repos.config, repos.party, l),
		Manage:    services.NewManageService(repos.area, repos.levels, repos.classes, repos.user, repos.equips, repos.config, l),
		Damage:    services.NewDamageService(l),
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
