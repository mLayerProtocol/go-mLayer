package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	// ID string `gorm:"primaryKey" json:"id,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Note: Gorm will fail if the function signature
//  does not include `*gorm.DB` and `error`

var Models = []interface{}{
	EventCounter{},
	Config{},
	TopicState{},
	TopicEvent{},
	MessageState{},
	MessageEvent{},
	AuthorizationState{},
	AuthorizationEvent{},
	SubscriptionState{},
	SubscriptionEvent{},
	BlockStat{},

	ApplicationState{},
	ApplicationEvent{},

	WalletState{},
	WalletEvent{},
	
	MigrationState{},
}

var SyncModels = []any{
	ApplicationEvent{},
	ApplicationState{},
	 AuthorizationEvent{},
	AuthorizationState{},
	TopicEvent{},
	TopicState{},
	SubscriptionEvent{},
	SubscriptionState{},
	MessageEvent{},
	MessageState{},
	EventCounter{},
}

var SyncEvents = []any{
	ApplicationEvent{},
	AuthorizationEvent{},
	TopicEvent{},
	SubscriptionEvent{},
	MessageEvent{},
}
