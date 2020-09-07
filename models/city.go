package models

type City struct {
	ID            int      `json:"id" firestore:"id"`
	EquipmentShop []string `json:"weapons" firestore:"weapons"`
	ConsumeShop   []string `json:"consumables" firestore:"consumables"`
}
