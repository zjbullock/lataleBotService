package resolvers

import (
	"github.com/juju/loggo"
	"lataleBotService/services"
)

type Resolver struct {
	Services struct {
		Adventure services.Adventure
	}
	Log loggo.Logger
}