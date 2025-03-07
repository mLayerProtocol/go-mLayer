package query

import (
	"context"
	"fmt"
	"reflect"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/channelpool"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
)

type DataStates struct {
	Events map[string]entities.Event
	CurrentStates map[entities.EntityPath]interface{}
	HistoricState map[entities.EntityPath][]byte
	DhtSync map[string][]byte
	Config *configs.MainConfiguration	`json:"-"`
	DataCount uint16
}



func NewDataStates(eventId string, cfg *configs.MainConfiguration) *DataStates {
	return &DataStates{
		Events: make(map[string]entities.Event),
		CurrentStates: make(map[entities.EntityPath]interface{}),
		HistoricState: make(map[entities.EntityPath][]byte),
		Config: cfg,
		DhtSync: make(map[string][]byte),
	}
}

func (ds *DataStates) Empty() bool {
	return ds.DataCount == 0
}

func (ds *DataStates) AddEvent( event entities.Event) {
	if len(event.Error) > 0 {
		event.IsValid = utils.FalsePtr()
	}
	if ds.Events[event.ID].ID == "" {
		id, err := event.GetId()
		if err != nil {
			panic(err)
		}
		ds.Events[id] = event
		ds.DataCount++
	} else {
		var s = ds.Events[event.ID];
		utils.UpdateStruct(&event, &s)
		ds.Events[event.ID] = s;
	}
	
}

func (ds *DataStates) AddCurrentState( model entities.EntityModel, id string, state interface{}) {
	path := entities.EntityPath{Model: model, ID: id}
	logger.Infof("VENTMOREREC %+v", path)
	if ds.CurrentStates[path] == nil {
		ds.CurrentStates[path] =  state
		ds.DataCount++
	} else {
		utils.UpdateStruct(state, ds.CurrentStates[path])
	}
}

func (ds *DataStates) AddToDhtSync( dhtKeyPrefix string, dataKey string, value []byte) {
	ds.DhtSync[fmt.Sprintf("/dht/%s/%s", dhtKeyPrefix, dataKey)] = value
}

func (ds *DataStates) AddHistoricState( model entities.EntityModel, id string, state []byte) {
	path := entities.EntityPath{Model: model, ID: id}
	if ds.HistoricState[path] == nil {
		ds.DataCount++
	}
	ds.HistoricState[path] = state
	
}

func (ds *DataStates) error(updateError error, eventId string, eventTx *datastore.Txn, wb *badger.WriteBatch) (err  error ) {
	// _eventTxn, err := InitTx(stores.EventStore, eventTx)
	// if err != nil {
	// 	return  err
	// }
	if eventTx == nil && wb == nil {
		wb = stores.EventStore.DB.NewWriteBatch()
	}
	for _, v := range ds.Events {
		logger.Debugf("EVENSTTOSAVE: %v", v.Hash)
		if v.ID == eventId {
			v.Error = updateError.Error()
			v.IsValid = utils.FalsePtr()
			v.Synced = utils.TruePtr()
			err = UpdateEvent(&v, eventTx, wb, true)
		}
		
	   if err != nil {
		   return err
	   }
	if eventTx == nil && err == nil {
		err = wb.Flush()
		if err != nil {
			logger.Errorf("COMMITEDEVENTError %v", err)
			return err
		}
	}	
   }
   return nil
}

func (ds *DataStates) Save(key string) error {
	b, err := encoder.MsgPackStruct(*ds)
	if err != nil {
		return err
	}
	go func() {
		channelpool.MempoolC <- &entities.KeyByteValue{Key: key, Value: b};
	}()
	return nil
}

// func DataStateFromKey(key string) (DataStates, error) {
// 	stores.MempoolStore.DB.View(func(txn *badger.Txn) error {
// 		txn.Get(ctx)
// 	})
// 	b, err := encoder.MsgPackStruct(*ds)
// 	if err != nil {
// 		return err
// 	}
// 	go func() {
// 		channelpool.MempoolC <- &entities.KeyByteValue{Key: key, Value: b};
// 	}()
// 	return nil
// }

func (ds *DataStates) Commit(stateTx *datastore.Txn, eventTx *datastore.Txn, messageTx *datastore.Txn, mainEventId string, stateError error) (err  error ) {
	_stateTxn, err := InitTx(stores.StateStore, stateTx)
	if err != nil {
		return  err
	}
	if stateTx == nil {
		defer _stateTxn.Discard(context.Background())
	}
	var _eventTxn datastore.Txn
	var _eventWB *badger.WriteBatch
	if eventTx != nil {
		_eventTxn, err = InitTx(stores.EventStore, eventTx)
		if err != nil {
			return  err
		}
	} else {
		_eventWB = stores.EventStore.DB.NewWriteBatch()
	}
	// if eventTx == nil {
	// 	defer _eventTxn.Discard(context.Background())
	// }

	if stateError != nil {
		return ds.error(stateError, mainEventId, eventTx, _eventWB)
	}

	_messageTxn, err := InitTx(stores.MessageStore, messageTx)
	if err != nil {
		return  err
	}
	if messageTx == nil {
		defer _messageTxn.Discard(context.Background())
	}

	for k, v := range ds.CurrentStates {
		
		// var state interface{}
		
		 
		switch k.Model {
		case entities.ApplicationModel:
			state := entities.Application{}
			if reflect.TypeOf(v).Kind() == reflect.Map {
				b, _ := encoder.MsgPackStruct(v)
				encoder.MsgPackUnpackStruct(b, &state)
			} else {
				state = v.(entities.Application)
			}
			_, err = UpdateApplicationState(k.ID, &state, &_stateTxn, true)
		case entities.AuthModel:
			state := entities.Authorization{}
			if reflect.TypeOf(v).Kind() == reflect.Map {
				b, _ := encoder.MsgPackStruct(v)
				encoder.MsgPackUnpackStruct(b, &state)
			} else {
				state = v.(entities.Authorization)
			}
			_, err = CreateAuthorizationState(&state, &_stateTxn)
		case entities.TopicModel:
			state := entities.Topic{}
			if reflect.TypeOf(v).Kind() == reflect.Map {
				b, _ := encoder.MsgPackStruct(v)
				encoder.MsgPackUnpackStruct(b, &state)
			} else {
				state = v.(entities.Topic)
			}
			_, err = UpdateTopicState(k.ID, &state, &_stateTxn, true)
		case entities.SubscriptionModel:
			state := entities.Subscription{}
			if reflect.TypeOf(v).Kind() == reflect.Map {
				b, _ := encoder.MsgPackStruct(v)
				encoder.MsgPackUnpackStruct(b, &state)
			} else {
				state = v.(entities.Subscription)
			}
			_, err = CreateSubscriptionState(&state, &_stateTxn)
		case entities.MessageModel:
			state := entities.Message{}
			if reflect.TypeOf(v).Kind() == reflect.Map {
				b, _ := encoder.MsgPackStruct(v)
				encoder.MsgPackUnpackStruct(b, &state)
			} else {
				state = v.(entities.Message)
			}
			_, err = CreateMessageState(&state, &_messageTxn)
		}
		if err != nil {
			return err
		}
	}

	for k, v := range ds.HistoricState {
		 err = SaveHistoricState(k.Model, k.ID, v)
		if err != nil {
			return err
		}

	}

	for k, v := range ds.DhtSync {
		err	 = _stateTxn.Put(context.Background(),datastore.NewKey(k), v)
	   if err != nil {
		   return err
	   }

   }

	if len(ds.Events) == 0 {
		// panic("No events")
		logger.Warnf("No Event Data")
	}
	for _, v := range ds.Events {
		logger.Debugf("EVENSTTOSAVE: %v", v.Hash)
		
		err = UpdateEvent(&v, eventTx,  _eventWB, true)
		
	   if err != nil {
		   return err
	   }
	  //  err = IncrementCounters(v.Cycle, v.Validator, v.Application, &_eventTxn)

   }
   if stateTx == nil && err == nil {
		err = _stateTxn.Commit(context.Background())
		if err != nil {
			logger.Errorf("COMMITEDESTATE %v", err)
		}
   }
   if messageTx == nil && err == nil {
		err = _messageTxn.Commit(context.Background())
		if err != nil {
			logger.Errorf("COMMITEDEVENT %v", err)
		}
	}	
	if eventTx != nil && err == nil {
		if len(ds.Events) == 0 {
			_eventTxn.Discard(context.Background())
		} else {
		err = _eventTxn.Commit(context.Background())
		if err != nil {
			logger.Errorf("COMMITEDEVENTError %v", err)
		}
		}
	}
	if _eventWB != nil && err == nil {
		if len(ds.Events) == 0 {
			_eventWB.Cancel()
		} else {
		err = _eventWB.Flush()
		if err != nil {
			logger.Errorf("COMMITEDEVENTError %v", err)
		}
		}
	}	
	if err == nil {
		// go utils.WriteBytesToFile(filepath.Join(ds.Config.DataDir, "log.txt"), []byte("newMessage" + "\n"))
	} else{
		logger.Error("DatastateCommitError", err)
	}
	return err
}