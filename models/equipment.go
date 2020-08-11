package models

type Equipment struct {
	Weapon *string `json:"weapon" firestore:"weapon"`
	Body   *string `json:"body" firestore:"body"`
	Glove  *string `json:"glove" firestore:"glove"`
	Shoes  *string `json:"shoes" firestore:"shoes"`
}
