package main

import (
	. "cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/graph-gophers/graphql-go"
	"github.com/juju/loggo"
	"google.golang.org/api/option"
	"io/ioutil"
	"lataleBotService/datasource"
	"lataleBotService/handler"
	"lataleBotService/repositories"
	"lataleBotService/resolvers"
	"lataleBotService/router"
	"lataleBotService/server"
	"lataleBotService/services"
	"net/http"
	"os"
	"strings"
)

var (
	ctx          = context.Background()
	l            loggo.Logger
	handlerFuncs *handler.Funcs
	ds           datasource.Datasource
	originList   []string
)

func init() {
	l.SetLogLevel(loggo.DEBUG)
	serv := server.NewServer(l)
	schemaString, err := serv.GetSchema("./server/graphql/", l)
	if err != nil {
		l.Criticalf("error occurred while fetching graphql schema: %v", err)
	}
	configFile, err := os.Open("./config.json")
	if err != nil {
		l.Errorf("error opening credentials file: %v", err)
		return
	}
	l.Errorf("configFile: %v", configFile)
	defer configFile.Close()
	var configMap map[string]interface{}
	byteValue, _ := ioutil.ReadAll(configFile)
	json.Unmarshal([]byte(byteValue), &configMap)
	client, err := NewClient(ctx, configMap["project_id"].(string), option.WithCredentialsFile("./credentials.json"), option.WithGRPCConnectionPool(10))
	if err != nil {
		l.Errorf("error initializing Fire Store client with projectId: %s. Received error: %v", configMap["project_id"].(string), err)
		return
	}
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

	origins := configMap["origins"].(string)
	originList = strings.Split(origins, ",")
}

func main() {
	r := router.NewRouter(handlerFuncs)
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})
	allowedOrigins := handlers.AllowedOrigins(originList)
	allowedMethods := handlers.AllowedMethods([]string{http.MethodPost, http.MethodOptions, http.MethodGet})
	l.Criticalf(http.ListenAndServe(":8080", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)).Error())
}
