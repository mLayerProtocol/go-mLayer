package query

import (
	"context"
	"fmt"
	"strconv"

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

func GetAccountAuthorizations( auth entities.Authorization, limits *QueryLimit, txn *datastore.Txn) (data []*entities.Authorization, err error) {
	ds :=  stores.StateStore
	var rsl query.Results
	if limits == nil {
		limits = DefaultQueryLimit
	}
	logger.Debugf("AccountAuthKEYYYY: %s", auth.AccountAuthorizationsKey())
	if txn != nil {
		rsl,  err = (*txn).Query(context.Background(), query.Query{
			Prefix: auth.AccountAuthorizationsKey(),
			Limit:  limits.Limit,
			Offset: limits.Offset,
		})
	} else {
		rsl,  err = ds.Query(context.Background(), query.Query{
			Prefix: auth.AccountAuthorizationsKey(),
			Limit:  limits.Limit,
			Offset: limits.Offset,
		})
	}

	
	
	if err != nil {
		
		return data, err
	}
	// logger.Debugf("Getting Authorizations...: %s", auth.AccountAuthorizationsKey())
	entries, _ := rsl.Rest()
	for _, entry := range entries { 
		
		//keyString := strings.Split(entry.Key, "/")
	logger.Debugf("Getting Authorizations Entries ID...: %s,",string(entry.Key))
		value, qerr := GetAuthorizationByEvent(entities.EventPath{EntityPath: entities.EntityPath{ Hash: string(entry.Value)}})
		if qerr != nil {
			logger.Infof("AuthERROR: %+v", qerr)
			continue
		}
		logger.Debugf("Getting Authorizations Event ID...: %s,",string(value.Event.Hash))
		data = append(data, value)
		err = qerr
	}
	// logger.Debugf("Getting Authorizations Entries...: %d, %v", len(entries), err)
	
	return data, err
}

func CreateAuthorizationState(newState *entities.Authorization, tx *datastore.Txn) (sub *entities.Authorization, err error) {
	if newState.Account == "" || newState.Agent == "" || newState.Subnet == "" {
		return nil, fmt.Errorf("new state must include acc, snet and agent fields")
	}
	ds := stores.StateStore
	
	newState.ID, err = entities.GetId(newState, newState.ID)
	
	if err != nil {
		return nil, err
	}
	stateBytes := newState.MsgPack()
	keys := newState.GetKeys()
	txn, err := InitTx(ds, tx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}
	
	// Delete old account keys
	accountRsl,  err := txn.Query(context.Background(), query.Query{
		Prefix: newState.AccountAuthorizationsKey(),
	})
	if err != nil && !IsErrorNotFound(err) {
		return nil, err
	}
	entries, _ := accountRsl.Rest()
	
	if len(entries) == 0 {
		agentCount := 0
		if agentCountBytes, err := txn.Get(context.Background(), datastore.NewKey(entities.AgentCountKey())); err != nil {
			if !IsErrorNotFound(err) {
				return nil, err
			}
		} else {
			if len(agentCountBytes) > 0 {
				agentCount, err = strconv.Atoi(string(agentCountBytes))
				if err != nil {
					return nil, err
				}
				
				
			}			
		}
		agentCount++
		txn.Put(context.Background(), datastore.NewKey(entities.AgentCountKey()), []byte(fmt.Sprint(agentCount)) )
		
	}
	

	for _, entry := range entries { 
		txn.Delete(context.Background(), datastore.NewKey(entry.Key))
	}
	agentRsl,  err := txn.Query(context.Background(), query.Query{
		Prefix: newState.AuthorizedAgentStateKey(),
	})
	if err != nil && !IsErrorNotFound(err) {
		return nil, err
	}
	entries2, _ := agentRsl.Rest()
	for _, entry := range entries2 { 
		txn.Delete(context.Background(), datastore.NewKey(entry.Key))
	}

	id, err :=  entities.GetId(newState, newState.ID)
	if err != nil {
		logger.Errorf("ERRORRRRR: %v", err)
		return nil, err
	}
	err = CreateState(CreateStateParam{
		ModelType: entities.AuthModel,
		ID: id,
		IDKey: newState.Key(),
		DataKey: newState.DataKey(),
		RefKey: nil,
		Keys: keys,
		Data: stateBytes,
		EventHash: newState.Event.Hash,
		RestKeyValue: []byte(newState.Event.Hash),
	}, tx)
	if err != nil {
		return nil, err
	}
	// for _, key := range keys {
	// 	logger.Infof("NewStateKey: %v, %v", key, newState.Event.Hash)
	// 	if strings.EqualFold(key, newState.DataKey()) {
	// 		if err := txn.Put(context.Background(), datastore.NewKey(key), stateBytes); err != nil {
	// 			return nil, err
	// 		}
	// 		continue
	// 	}
		
	// 	if strings.EqualFold(key, newState.Key()) {

	// 		if err := txn.Put(context.Background(), datastore.NewKey(key), []byte(newState.Event.Hash)); err != nil {
	// 			return nil, err
	// 		}

	// 		continue
	// 	}

	// 	if err := txn.Put(context.Background(), datastore.NewKey(key), []byte(newState.Event.Hash)); err != nil {
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

// func UpdateAuthorizationState(newState *entities.Authorization, tx *datastore.Txn) (*entities.Authorization, error) {
// 	ds := stores.StateStore
// 	txn, err := InitTx(ds, tx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if tx == nil {
// 		defer txn.Discard(context.Background())
// 	}
// 	if newState.Account == "" || newState.Agent == "" || newState.Subnet == "" {
// 		return nil, fmt.Errorf("new state must include acc, snet and agent field")
// 	}

// 	stateBytes := newState.MsgPack()

// 	// Delete old account keys
// 	accountRsl,  err := txn.Query(context.Background(), query.Query{
// 		Prefix: newState.AccountAuthorizationsKey(),
// 	})
// 	entries, _ := accountRsl.Rest()
// 	for _, entry := range entries { 
// 		txn.Delete(context.Background(), datastore.NewKey(entry.Key))
// 	}
// 	agentRsl,  err := txn.Query(context.Background(), query.Query{
// 		Prefix: newState.AuthorizedAgentStateKey(),
// 	})
// 	entries2, _ := agentRsl.Rest()
// 	for _, entry := range entries2 { 
// 		txn.Delete(context.Background(), datastore.NewKey(entry.Key))
// 	}

// 	oldState, err := GetAccountAuthorizations(*newState, nil, &txn)

// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(oldState) == 0 {
// 		return nil, datastore.ErrNotFound
// 	}

// 	if err := txn.Put(context.Background(), datastore.NewKey(newState.DataKey()), stateBytes); err != nil {
// 		logger.Errorf("error updateing state key: %v", err)
// 		return nil, err
// 	}

// 	if err := txn.Put(context.Background(), datastore.NewKey(newState.Key()), []byte(newState.Event.Hash)); err != nil {
// 		logger.Errorf("error updateing state key: %v", err)
// 		return nil, err
// 	}

// 	if tx == nil {
// 		if err := txn.Commit(context.Background()); err != nil {
// 			return nil, err
// 		}
// 	}

	
// 	return newState, nil
// }


func GetAgentAuthorizationStates(subnet string, agent entities.DeviceString, limits QueryLimit) (rsl []*entities.Authorization, err error) {
	ds :=  stores.StateStore
	result,  err := ds.Query(context.Background(), query.Query{
		Prefix: (&entities.Authorization{Subnet: subnet, Agent: agent}).AuthorizedAgentStateKey(),
		Limit:  limits.Limit,
		Offset: limits.Offset,
	})
	if err != nil {
		return nil, err
	}
	entries, _ := result.Rest()
	for _, entry := range entries { 
		value, qerr := GetAuthorizationByEvent(entities.EventPath{EntityPath: entities.EntityPath{ Hash: string(entry.Value)}})
		if qerr != nil {
			continue
		}
		rsl = append(rsl, value)
		
	}	
	return rsl, err
}

func GetAuthorizationByEvent( event entities.EventPath) (*entities.Authorization, error) {
	ds :=  stores.StateStore
	
	value, err := ds.Get(context.Background(), datastore.NewKey((&entities.Authorization{Event: event}).DataKey()))
	if err != nil {
		return nil, err
	}
	data, err := entities.UnpackAuthorization(value)
	if err != nil {
		return nil, err
	}
	return &data, err
}

// auth/agt/did:0x99E904417f7e69505c738CB24F66EBeF688AB19d/fb6d5a3d-3d1c-4051-9577-9bd9d13fd20e/did:0x59fD8f94dDd1Fe6066d300F74afD5E3a01970e43/20241111093301000
// auth/agt/did:0x59fD8f94dDd1Fe6066d300F74afD5E3a01970e43/fb6d5a3d-3d1c-4051-9577-9bd9d13fd20e/did:0x73d67D769f10b860e51B5234D467624930D36Ec1