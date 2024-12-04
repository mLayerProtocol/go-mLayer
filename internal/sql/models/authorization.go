package models

import (
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"gorm.io/gorm"
)



type AuthorizationEvent struct {
	// Event `msgpack:",noinline"`
	entities.Event	 `msgpack:",noinline"`
	//IsValid   bool `gorm:"default:false" json:"isVal"`
	// EventType int16 `json:"t"`
	// Payload entities.ClientPayload  `json:"pld"  msgpack:",noinline"`
	BaseModel `msgpack:",noinline"`
}
func (AuthorizationEvent) TableName() string {
    return "authorization_events"
}

type AuthorizationState struct {
	entities.Authorization `msgpack:",noinline"`
	BaseModel
}
func (AuthorizationState) TableName() string {
    return "authorization_states"
}

func (d *AuthorizationState) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	d.ID, err = entities.GetId(*d)
	if err != nil {
		panic(err)
	}
	return nil
  }

func (auth AuthorizationState) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(&auth.Authorization)
	return b
}