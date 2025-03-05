package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
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

func GetMessages( filter entities.Message, limits *entities.QueryLimit, txn *datastore.Txn) (data []*entities.Message, err error) {
	ds :=  stores.MessageStore
	var rsl query.Results
	if limits == nil {
		limits = DefaultQueryLimit
	}
	prefix := filter.TopicMessageKey()
	if filter.Sender != "" && filter.Receiver != "" {
		prefix = filter.MessageSenderReceiverKey()
	} else {
		if filter.Sender != "" {
			prefix = filter.MessageSenderKey()
		}
		if filter.Receiver != "" {
			prefix = filter.MessageReceiverKey()
		}
	}
	if prefix == filter.TopicMessageKey() && filter.Topic == "" {
		return []*entities.Message{}, nil
	}
	logger.Debugf("MessagehKEYYYY: %s", prefix)
	if txn != nil {
		rsl,  err = (*txn).Query(context.Background(), query.Query{
			Prefix: prefix,
			Limit:  limits.Limit,
			Offset: limits.Offset,
			Orders: []query.Order{query.OrderByKeyDescending{}},
		})
	} else {
		rsl,  err = ds.Query(context.Background(), query.Query{
			Prefix: prefix,
			Limit:  limits.Limit,
			Offset: limits.Offset,
			Orders: []query.Order{query.OrderByKeyDescending{}},
		})
	}

	
	
	if err != nil {
		return nil, err
	}
	// logger.Debugf("Getting Authorizations...: %s", auth.AccountAuthorizationsKey())
	entries, _ := rsl.Rest()
	for _, entry := range entries { 
		
		//keyString := strings.Split(entry.Key, "/")
		logger.Debugf("Getting Authorizations Entries ID...: %s,",string(entry.Key))
		keyString := strings.Split(entry.Key, "/")
		eventId := keyString[len(keyString)-1]
		value, qerr := GetMessageByEventHash(eventId)
		if qerr != nil {
			continue
		}
		data = append(data, value)
		err = qerr
	}
	// logger.Debugf("Getting Authorizations Entries...: %d, %v", len(entries), err)
	if err != nil {
		return nil, err
	}
	return data, err
}

func CreateMessageState(newState *entities.Message, tx *datastore.Txn) (sub *entities.Message, err error) {
	
	if newState.Sender == "" || (newState.Receiver == "" && newState.Topic == "") {
		return nil, fmt.Errorf("new message state must include s (sender), and (r (receiver) or top (topic)) fields")
	}
	
	
	newState.ID, err = entities.GetId(newState, newState.ID)
	if err != nil {
		logger.Infof("CREATINGMESSAGE_ERROR: %+v", err)
		return nil, err
	}
	stateBytes := newState.MsgPack()
	
	ds := stores.MessageStore
	txn, err := InitTx(ds, tx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}
	
	uniqId := newState.UniqueId()
	stateByte, checkError := txn.Get(context.Background(), datastore.NewKey(uniqId))
	if checkError == nil  && len(stateByte) > 0 {
		m, err := entities.UnpackMessage(stateByte)
		if err != nil {
			return nil, err
		}
		return &m, err
	}
	
	id, err :=  entities.GetId(newState, newState.ID)
	if err != nil {
		logger.Errorf("ERRORRRRR: %v", err)
		return nil, err
	}
	
	err = CreateState(CreateStateParam{
		ModelType: entities.MessageModel,
		ID: id,
		IDKey: newState.Key(),
		DataKey: newState.DataKey(),
		RefKey: nil,
		Keys: newState.GetKeys(),
		Data: stateBytes,
		EventHash: newState.Event.ID,
		RestKeyValue: []byte{},
	}, tx)
	if err != nil {
		return nil, err
	}
	// for _, key := range keys {
	// 	logger.Infof("NewStateKey: %v, %v", key, newState.Event.ID)
	// 	if strings.EqualFold(key, newState.DataKey()) {
	// 		if err := txn.Put(context.Background(), datastore.NewKey(key), stateBytes); err != nil {
				
	// 			return nil, err
	// 		}
	// 		continue
	// 	}
		
	// 	if strings.EqualFold(key, newState.Key()) {

	// 		if err := txn.Put(context.Background(), datastore.NewKey(key), []byte(newState.Event.ID)); err != nil {
	// 			return nil, err
	// 		}

	// 		continue
	// 	}

	// 	if err := txn.Put(context.Background(), datastore.NewKey(key), []byte{}); err != nil {
	// 		logger.Infof("CREATINGMESSAGE_ERROR1: %s,  %+v", key, err)
	// 		return nil, err
	// 	}
	// }
	if tx == nil {
		err = txn.Commit(context.Background())
		if err != nil {
			return nil, err
		}
	}

	
	return newState, nil
}



func GetMessageByEventHash( id string) (*entities.Message, error) {
	ds :=  stores.MessageStore
	
	value, err := ds.Get(context.Background(), datastore.NewKey((&entities.Message{Event: entities.EventPath{EntityPath: entities.EntityPath{ID: id}}}).DataKey()))
	if err != nil {
		return nil, err
	}
	data, err := entities.UnpackMessage(value)
	if err != nil {
		return nil, err
	}
	return &data, err
}



// auth/agt/did:0x99E904417f7e69505c738CB24F66EBeF688AB19d/fb6d5a3d-3d1c-4051-9577-9bd9d13fd20e/did:0x59fD8f94dDd1Fe6066d300F74afD5E3a01970e43/20241111093301000
// auth/agt/did:0x59fD8f94dDd1Fe6066d300F74afD5E3a01970e43/fb6d5a3d-3d1c-4051-9577-9bd9d13fd20e/did:0x73d67D769f10b860e51B5234D467624930D36Ec1