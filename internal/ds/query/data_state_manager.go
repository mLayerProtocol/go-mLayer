package query

import (
	"context"
	"path/filepath"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
)

type DataStates struct {
	Events map[string]entities.Event
	CurrentStates map[entities.EntityPath]interface{}
	HistoricState map[entities.EntityPath][]byte
	Config *configs.MainConfiguration
	DataCount uint16
}

func NewDataStates(cfg *configs.MainConfiguration) *DataStates {
	return &DataStates{
		Events: make(map[string]entities.Event),
		CurrentStates: make(map[entities.EntityPath]interface{}),
		HistoricState: make(map[entities.EntityPath][]byte),
		Config: cfg,
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
	path := entities.EntityPath{Model: model, Hash: id}
	logger.Infof("VENTMOREREC %+v", path)
	if ds.CurrentStates[path] == nil {
		ds.CurrentStates[path] =  state
		ds.DataCount++
	} else {
		utils.UpdateStruct(state, ds.CurrentStates[path])
	}
}

func (ds *DataStates) AddHistoricState( model entities.EntityModel, id string, state []byte) {
	path := entities.EntityPath{Model: model, Hash: id}
	if ds.HistoricState[path] == nil {
		ds.DataCount++
	}
	ds.HistoricState[path] = state
	
}

func (ds *DataStates) Commit(stateTx *datastore.Txn, eventTx *datastore.Txn, messageTx *datastore.Txn) (err  error ) {
	_stateTxn, err := InitTx(stores.StateStore, stateTx)
	if err != nil {
		return  err
	}
	if stateTx == nil {
		defer _stateTxn.Discard(context.Background())
	}

	_eventTxn, err := InitTx(stores.EventStore, eventTx)
	if err != nil {
		return  err
	}
	if eventTx == nil {
		defer _eventTxn.Discard(context.Background())
	}

	_messageTxn, err := InitTx(stores.MessageStore, messageTx)
	if err != nil {
		return  err
	}
	if messageTx == nil {
		defer _messageTxn.Discard(context.Background())
	}

	for k, v := range ds.CurrentStates {
		switch k.Model {
		case entities.SubnetModel:
			state := v.(entities.Subnet)
			_, err = UpdateSubnetState(k.Hash, &state, &_stateTxn, true)
		case entities.AuthModel:
			state := v.(entities.Authorization)
			_, err = CreateAuthorizationState(&state, &_stateTxn)
		case entities.TopicModel:
			state := v.(entities.Topic)
			_, err = UpdateTopicState(k.Hash, &state, &_stateTxn, true)
		case entities.SubscriptionModel:
			state := v.(entities.Subscription)
			_, err = CreateSubscriptionState(&state, &_stateTxn)
		case entities.MessageModel:
			state := v.(entities.Message)
			_, err = CreateMessageState(&state, &_messageTxn)
		}
		if err != nil {
			return err
		}
	}

	for k, v := range ds.HistoricState {
		 err = SaveHistoricState(k.Model, k.Hash, v)
		if err != nil {
			return err
		}

	}
	if len(ds.Events) == 0 {
		panic("No events")
	}
	for _, v := range ds.Events {
		logger.Debugf("EVENSTTOSAVE", v.Hash)
		err = UpdateEvent(&v, &_eventTxn, true)
		
	   if err != nil {
		   return err
	   }
	  //  err = IncrementCounters(v.Cycle, v.Validator, v.Subnet, &_eventTxn)

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
	if eventTx == nil && err == nil {
		err = _eventTxn.Commit(context.Background())
		if err != nil {
			logger.Errorf("COMMITEDEVENTError %v", err)
		}
	}	
	if err == nil {
		go utils.WriteBytesToFile(filepath.Join(ds.Config.DataDir, "log.txt"), []byte("newMessage" + "\n"))
	} else{
		logger.Error("DatastateCommitError", err)
	}
	return err
}