package models

import "github.com/mlayerprotocol/go-mlayer/entities"

type EventCounter struct {
	Id        uint      `json:"id" gorm:"unique;primaryKey;autoIncrement"`
	Count *uint64		
	Cycle *uint64			`gorm:"uniqueIndex:idx_uniq_cyc_sub_val;"`
	Application string		`gorm:"uniqueIndex:idx_uniq_cyc_sub_val;"`
	Validator entities.PublicKeyString `gorm:"uniqueIndex:idx_uniq_cyc_sub_val;"`
	Claimed *bool `gorm:"default:false;"`
}
