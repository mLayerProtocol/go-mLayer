package models

import (
	"github.com/google/uuid"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"gorm.io/gorm"
)

type BlockStat struct {
	entities.BlockStats
	BaseModel
}

func (t *BlockStat) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New().String()

	return nil
}
