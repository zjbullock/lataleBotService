package services

import (
	"github.com/juju/loggo"
	"lataleBotService/repositories"
)

type Adventure interface {

}

type adventure struct {
	areas repositories.Repository
	character repositories.Repository
	users repositories.Repository
	log loggo.Logger
}

func NewAdventureService(areas, character, users repositories.Repository, log loggo.Logger) Adventure {
	return &adventure{
		areas: areas,
		character: character,
		users: users,
		log: log,
	}
}