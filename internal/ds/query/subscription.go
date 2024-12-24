package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
)

// func GetAuthorizationById(did string) (*entities.Authorization, error) {
// 	var stateData []byte
// 	stateData, err := GetStateById(did, entities.AuthModel)
// 	if err != nil {
// 		return nil, err
// 	}
// 	data, err := entities.UnpackAuthorization(stateData)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &data, errp
// }

// func GetSubscriptionById(did string) (*entities.Subscription, error) {
// 	var stateData []byte
// 	stateData, err := GetStateById(did, entities.TopicModel)
// 	if err != nil {
// 		return nil, err
// 	}
// 	data, err := entities.UnpackSubscription(stateData)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &data, err
// }
func GetSubscriptions( filter entities.Subscription, limits *QueryLimit, txn *datastore.Txn) (data []*entities.Subscription, err error) {
	defer utils.TrackExecutionTime(time.Now(), "GetSubscriptions::")

	ds :=  stores.StateStore
	var rsl query.Results
	if limits == nil {
		limits = &QueryLimit{}
	}
	key := filter.SubscriberKey()
	if filter.Status != nil {
		key = filter.SubscriptionStatusKey()
	}
	if filter.Status != nil && filter.Topic == "" && filter.Subscriber != "" {
		key = filter.SubscribedTopicsKey()
	}
	
	if txn != nil {
		rsl,  err = (*txn).Query(context.Background(), query.Query{
			Prefix: key,
			Limit:  limits.Limit,
			Offset: limits.Offset,
		})
	} else {
		rsl,  err = ds.Query(context.Background(), query.Query{
			Prefix: key,
			Limit:  limits.Limit,
			Offset: limits.Offset,
		})
	}
	if err != nil {
		return nil, err
	}
	entries, _ := rsl.Rest()
	
	for _, entry := range entries { 
		
		value, qerr := GetSubscriptionByEvent(entities.EventPath{EntityPath: entities.EntityPath{ ID: string(entry.Value)}})
		if qerr != nil {
			continue
		}
		logger.Debugf("Getting Subscription Event ID...: %s,",string(value.Event.ID))
		if value.Timestamp == nil {
			if index := strings.LastIndex(entry.Key, string("/")); index != -1 {
				layout := "200602011504000"
				// Parse the date string
				date, _ := time.Parse(layout, entry.Key[index+1:])
				logger.Errorf("ERROR %v, %v", date,  entry.Key[index+1:])
				ts64 := uint64(date.UnixMilli())
				value.Timestamp = &ts64
				
			} else {
				fmt.Println("Character not found.")
			}
		}
		
		logger.Debugf("Getting Subscription Event ID2...: %s,",string(value.Event.ID))
		data = append(data, value)
		err = qerr
	}
	// logger.Debugf("Getting Authorizations Entries...: %d, %v", len(entries), err)
	if err != nil {
		logger.Errorf("SubscriptionError: Agent not subscribed")
		return nil, err
	}
	return data, err
}

func CreateSubscriptionState(newState *entities.Subscription, tx *datastore.Txn) (sub *entities.Subscription, err error) {
	if newState.Subscriber == "" || newState.Topic == "" || newState.Subnet == "" {
		return nil, fmt.Errorf("new state must include acc, snet and agent fields")
	}
	ds := stores.StateStore
	txn, err := InitTx(ds, tx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}
	_refKey := newState.RefKey()
	refKey := &_refKey
	logger.Infof("CreatingSubscriptionState: %v", err)
	id := newState.ID
	if len(id) == 0 {
		id, err =  entities.GetId(newState, newState.ID)
	}
	newState.ID = id
	if err != nil {
		logger.Errorf("ERRORRRRR: %v", err)
		return nil, err
	}

	subscriptions,  err := GetSubscriptions(entities.Subscription{Topic: newState.Topic, Subscriber: newState.Subscriber}, DefaultQueryLimit, &txn)
	if err != nil && !IsErrorNotFound(err) {
		return nil, err
	}
	
	for _, sub := range subscriptions {
		if !strings.EqualFold(sub.RefKey(), _refKey) {
			err = txn.Delete(context.Background(), datastore.NewKey(sub.RefKey()))
		}
		if sub.Status != newState.Status {
			// delete the subscription
			err = txn.Delete(context.Background(), datastore.NewKey(sub.SubscribedTopicsKey()))
		}
	}

	accountRsl,  err := txn.Query(context.Background(), query.Query{
		Prefix: newState.SubscriberKey(),
	})
	if err != nil && !IsErrorNotFound(err) {
		return nil, err
	}
	entries, _ := accountRsl.Rest()
	for _, entry := range entries { 

		 txn.Delete(context.Background(), datastore.NewKey(entry.Key))
	}
	
	accountStatusRsl,  err := txn.Query(context.Background(), query.Query{
		Prefix: newState.SubscriptionStatusKey(),
	})
	if err != nil && !IsErrorNotFound(err) {
		return nil, err
	}
	entries, _ = accountStatusRsl.Rest()
	for _, entry := range entries { 
		txn.Delete(context.Background(), datastore.NewKey(entry.Key))
	}
	
	if newState.Timestamp == nil {
		now := uint64(time.Now().UnixMilli())
		newState.Timestamp = &now
	}
	// delete the ref key
	
	if len(newState.Ref) > 0 {
		val, err := txn.Get(context.Background(), datastore.NewKey(_refKey))
		if err != nil && !IsErrorNotFound(err) {
			return  nil, err
		}
		if len(val) > 0  && string(val) != newState.ID {
			return nil, fmt.Errorf("\"%s\" ref already exists", entities.SubnetModel)
		}
	} else {
		refKey = nil
	}
	
	err = CreateState(CreateStateParam{
		ModelType: entities.SubscriptionModel,
		ID: id,
		IDKey: newState.Key(),
		DataKey: newState.DataKey(),
		RefKey: refKey,
		Keys: newState.GetKeys(),
		Data: newState.MsgPack(),
		EventHash: newState.Event.ID,
	}, tx)
	if err != nil {
		logger.Errorf("ERRORRRRR: %v", err)
		return nil, err
	}
	return newState, err
}

// func UpdateSubscriptionState(newState *entities.Subscription, tx *datastore.Txn) (*entities.Subscription, error) {
// 	ds := stores.StateStore
// 	txn, err := InitTx(ds, tx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if tx == nil {
// 		defer txn.Discard(context.Background())
// 	}
// 	if newState.Topic == "" || newState.Subscriber == "" || newState.Subnet == "" {
// 		return nil, fmt.Errorf("new state must include acc, snet and agent field")
// 	}
	
// 	accountRsl,  err := txn.Query(context.Background(), query.Query{
// 		Prefix: newState.SubscriberKey(),
// 	})
// 	if err != nil && !IsErrorNotFound(err) {
// 		return nil, err
// 	}
// 	entries, _ := accountRsl.Rest()
// 	for _, entry := range entries { 
// 		txn.Delete(context.Background(), datastore.NewKey(entry.Key))
// 	}
// 	return CreateSubscriptionState(newState, tx)
// 	// err = UpdateState(id, NewStateParam{
// 	// 	OldIDKey: newState.Key(),
// 	// 	DataKey: newState.DataKey(),
// 	// 	Data: newState.MsgPack(),
// 	// 	EventHash: newState.Event.ID,
// 	// 	RefKey: &newState.Ref,
// 	// 	OldRefKey: &oldTopic.Ref,
// 	// }, tx)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
	
// 	// return newState, nil
// }


func GetSubscriptionStates(subnet string, topic string, subscriber entities.DeviceString, limits QueryLimit) (rsl []*entities.Subscription, err error) {
	ds :=  stores.StateStore
	result,  err := ds.Query(context.Background(), query.Query{
		Prefix: (&entities.Subscription{Subnet: subnet, Topic: topic, Subscriber: entities.DIDString(subscriber)}).SubscriberKey(),
		Limit:  limits.Limit,
		Offset: limits.Offset,
	})
	if err != nil {
		return nil, err
	}
	entries, _ := result.Rest()
	for _, entry := range entries { 
		value, qerr := GetSubscriptionByEvent(entities.EventPath{EntityPath: entities.EntityPath{ ID: string(entry.Value)}})
		if qerr != nil {
			continue
		}
		rsl = append(rsl, value)
		
	}	
	return rsl, err
}

func GetSubscriptionByEvent( event entities.EventPath) (*entities.Subscription, error) {
	ds :=  stores.StateStore
	
	value, err := ds.Get(context.Background(), datastore.NewKey((&entities.Subscription{Event: event}).DataKey()))
	if err != nil {
		return nil, err
	}
	data, err := entities.UnpackSubscription(value)
	if err != nil {
		return nil, err
	}
	return &data, err
}




// auth/agt/did:0x99E904417f7e69505c738CB24F66EBeF688AB19d/fb6d5a3d-3d1c-4051-9577-9bd9d13fd20e/did:0x59fD8f94dDd1Fe6066d300F74afD5E3a01970e43/20241111093301000
// auth/agt/did:0x59fD8f94dDd1Fe6066d300F74afD5E3a01970e43/fb6d5a3d-3d1c-4051-9577-9bd9d13fd20e/did:0x73d67D769f10b860e51B5234D467624930D36Ec1