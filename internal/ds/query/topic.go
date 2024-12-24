package query

import (
	"context"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
)


func GetTopicById(did string) (*entities.Topic, error) {
	var stateData []byte
	stateData, err := GetStateById(did, entities.TopicModel)
	if err != nil {
		return nil, err
	}
	data, err := entities.UnpackTopic(stateData)
	if err != nil {
		return nil, err
	}
	return &data, err
}

func GetAccountTopics( topic entities.Topic, limits *QueryLimit, txn *datastore.Txn) (data []*entities.Topic, err error) {
	ds :=  stores.StateStore
	var rsl query.Results
	if limits == nil {
		limits = DefaultQueryLimit
	}
	if txn != nil {
		rsl,  err = (*txn).Query(context.Background(), query.Query{
			Prefix: topic.GetAccountTopicsKey(),
			Limit:  limits.Limit,
			Offset: limits.Offset,
		})
	} else {
		rsl,  err = ds.Query(context.Background(), query.Query{
			Prefix: topic.GetAccountTopicsKey(),
			Limit:  limits.Limit,
			Offset: limits.Offset,
		})
	}

	data  = []*entities.Topic{}
	
	if err != nil {
		return nil, err
	}
	logger.Debugf("Getting Topics...: %s", topic.GetAccountTopicsKey())
	entries, _ := rsl.Rest()
	for _, entry := range entries { 
		
		//keyString := strings.Split(entry.Key, "/")
		logger.Debugf("Getting Topics Entries ID...: %s",string(entry.Value))
		value, qerr := GetTopicById(string(entry.Value))
		if qerr != nil {
			continue
		}
		data = append(data, value)
		err = qerr
	}
	logger.Debugf("Getting Topics Entries...: %d, %v", len(entries), err)
	if err != nil {
		return nil, err
	}
	return data, err
}

func CreateTopicState(newState *entities.Topic, tx *datastore.Txn) (sub *entities.Topic, err error) {
	refKey := newState.RefKey()
	if len(newState.Ref) == 0 {
		refKey = ""
	}
	logger.Infof("CreatingTopicState: %v",  newState)
	id := newState.ID
	if id == "" {
		id, err =  entities.GetId(newState, newState.ID)
	}
	if err != nil {
		logger.Errorf("ERRORRRRR: %v", err)
		return nil, err
	}
	err = CreateState(CreateStateParam{
		ModelType: entities.TopicModel,
		ID: id,
		IDKey: newState.Key(),
		DataKey: newState.DataKey(),
		RefKey: &refKey,
		Keys: newState.GetKeys(),
		Data: newState.MsgPack(),
		EventHash: newState.Event.ID,
	}, tx)
	if err != nil {
		logger.Errorf("ERRORRRRR: %v", err)
		return nil, err
	}
	return newState, err
	// ds := stores.StateStore
	// newState.ID, err = entities.GetId(newState)
	// if err != nil {
	// 	return nil, err
	// }
	// stateBytes := newState.MsgPack()
	// keys := newState.GetKeys()
	// txn, err := InitTx(ds, tx)
	// if err != nil {
	// 	return nil, err
	// }
	// if tx == nil {
	// 	defer txn.Discard(context.Background())
	// }

	// // return txn.Set(key.Bytes(), value)
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

	// 	if err := txn.Put(context.Background(), datastore.NewKey(key), []byte(newState.ID)); err != nil {
	// 		return nil, err
	// 	}
	// }
	// if tx == nil {
	// 	err = txn.Commit(context.Background())
	// }

	// if err != nil {
	// 	return nil, err
	// }
	// return newState, nil
}

func UpdateTopicState(id string, newState *entities.Topic, tx *datastore.Txn, create bool) (*entities.Topic, error) {
	id, err := entities.GetId(*newState, id)
	if err != nil {
		return nil, err
	}
	oldTopic, err :=  GetTopicById(id)
	if err != nil {
		if IsErrorNotFound(err) && create {
			return CreateTopicState(newState, tx)
		}
		return nil, err
	}
	
	err = UpdateState(id, NewStateParam{
		OldIDKey:  fmt.Sprintf("%s/id/%s", entities.TopicModel, id),
		DataKey: newState.DataKey(),
		Data: newState.MsgPack(),
		EventHash: newState.Event.ID,
		RefKey: &newState.Ref,
		OldRefKey: &oldTopic.Ref,
	}, tx)
	if err != nil {
		return nil, err
	}
	
	return newState, nil
}




func GetTopicByEvent( event entities.EventPath) (*entities.Topic, error) {
	ds :=  stores.StateStore
	
	value, err := ds.Get(context.Background(), datastore.NewKey((&entities.Topic{Event: event}).DataKey()))
	if err != nil {
		return nil, err
	}
	data, err := entities.UnpackTopic(value)
	if err != nil {
		return nil, err
	}
	return &data, err
}