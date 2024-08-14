package query

import (
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	db "github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
	"gorm.io/gorm"
)

// Save Subnet state only when it doesnt exist
func GetSubscriptionStateBySuscriber(subnet string, topic string, subscribers []entities.DIDString, DB *gorm.DB) (*[]models.SubscriptionState, error) {
	data := []models.SubscriptionState{}
	tx := DB
	if DB == nil {
		tx = db.SqlDb
	}
	subsc := []entities.DIDString{}
	for _, sub := range subscribers {
		if sub == "" {
			continue
		}
		subsc = append(subsc, sub)
	}
	logger.Infof("HELLOOOO::: %v, %v, %v", subnet, topic, subsc)
	err := tx.Model(models.SubscriptionState{}).Where(models.SubscriptionState{
		Subscription: entities.Subscription{ Subnet: subnet, Topic: topic },
	}).Where("subscriber IN ?", subsc).Assign(&data).Error
	if err != nil {
		return nil, err
	}
	// if DB == nil {
	// 	tx.Commit()
	// }
	return &data, nil
}
