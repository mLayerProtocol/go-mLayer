package models

import (
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"gorm.io/gorm"
)

type SubnetState struct {
	entities.Subnet
	BaseModel
}

func (d *SubnetState) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		hash, err := entities.GetId(*d, d.ID)
		if err != nil {
			panic(err)
		}
		d.ID = hash
	}
	return nil
}

func (d SubnetState) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(&d.Subnet)
	return b
}

type SubnetEvent struct {
	entities.Event
	BaseModel
	// SubnetID     uint64
	// Subnet		SubnetState
}
