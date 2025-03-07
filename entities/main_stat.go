package entities

type MainStat struct {
	Accounts     uint64    `json:"accounts"`
	MessageCost string `json:"message_cost"`
	// TopicBalance uint64    `json:"topic_balance"`
	TotalValueLocked string    `json:"tvl"`
	SubscriptionCount int64 `json:"subC"`
	EventCount uint64 `json:"event_count"`
	TotalEventsValue string `json:"total_events_value"`
	TopicCount uint64 `json:"topC"`
	AuthCount uint64 `json:"authC"`
	AppKeyCount uint64 `json:"appKeyC"`
}
