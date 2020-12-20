package resolvers

import "lataleBotService/models"

type traitResolver struct {
	trait models.Trait
}

func (t *traitResolver) Name() string {
	return t.trait.Name
}

func (t *traitResolver) Description() string {
	return t.trait.Description
}
