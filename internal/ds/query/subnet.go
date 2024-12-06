package query

import (
	"context"
	"strings"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
)



func GetSubnetStateById(did string) (*entities.Subnet, error) {
	var stateData []byte

	stateData, err := GetStateById(did, entities.SubnetModel)
	if err != nil {
		return nil, err
	}
	data, err := entities.UnpackSubnet(stateData)
	if err != nil {
		return nil, err
	}
	return &data, err
}

func CreateSubnetState(newState *entities.Subnet, tx *datastore.Txn) (sub *entities.Subnet, err error) {
	ds := stores.StateStore
	if newState.ID == "" {
		newState.ID, err = entities.GetId(newState)
	}
	if err != nil {
		return nil, err
	}
	logger.Infof("CreatingSubnet... %s", newState.ID)
	stateBytes := newState.MsgPack()
	keys := newState.GetKeys()
	txn, err := InitTx(ds, tx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}

	// return txn.Set(key.Bytes(), value)
	for _, key := range keys {
		logger.Infof("NewStateKey: %v, %v", key, newState.Event.Hash)
		if strings.EqualFold(key, newState.DataKey()) {
			if err := txn.Put(context.Background(), datastore.NewKey(key), stateBytes); err != nil {
				return nil, err
			}
			continue
		}
		if strings.EqualFold(key, newState.Key()) {

			if err := txn.Put(context.Background(), datastore.NewKey(key), []byte(newState.Event.Hash)); err != nil {
				return nil, err
			}

			continue
		}

		if err := txn.Put(context.Background(), datastore.NewKey(key), []byte(newState.ID)); err != nil {
			return nil, err
		}
	}
	if tx == nil {
		err = txn.Commit(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return newState, nil
}

func UpdateSubnetState(id string, newState *entities.Subnet, tx *datastore.Txn) (*entities.Subnet, error) {
	ds := stores.StateStore
	txn, err := InitTx(ds, tx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}

	// state.ID, err = entities.GetId(state)
	// if err != nil {
	// 	return  nil, err
	// }
	stateBytes := newState.MsgPack()

	oldState, err := GetSubnetStateById(id)

	if err != nil {
		return nil, err
	}

	logger.Debugf("UpdateSubnet %v, %v, %v, %v", id, oldState.ID, newState.ID, newState.Event.Hash)

	if err := txn.Put(context.Background(), datastore.NewKey(newState.DataKey()), stateBytes); err != nil {
		logger.Errorf("error updateing state key: %v", err)
		return nil, err
	}

	if err := txn.Put(context.Background(), datastore.NewKey(oldState.Key()), []byte(newState.Event.Hash)); err != nil {
		logger.Errorf("error updateing state key: %v", err)
		return nil, err
	}

	if oldState.Ref != newState.Ref {
		err := txn.Delete(context.Background(), datastore.NewKey(oldState.RefKey()))
		if err != nil {
			return nil, err
		}
		if err := txn.Put(context.Background(), datastore.NewKey(newState.RefKey()), []byte(newState.Event.Hash)); err != nil {
			logger.Errorf("error updateing subnet ref: %v", err)
			return nil, err
		}
	}
	if tx == nil {
		if err := txn.Commit(context.Background()); err != nil {
			return nil, err
		}
	}

	logger.Infof("SubnetKey: %s", newState.Event.Hash)

	// if err := txn.Commit(context.Background()); err != nil {
	// 	return nil, err
	// }
	return newState, nil
}

func GetAccountSubnets(account entities.DIDString, limit QueryLimit) (data []*entities.Subnet, err error) {
	ds := stores.StateStore
	key := (&entities.Subnet{Account: account}).AccountSubnetsKey()
	rsl, err := ds.Query(context.Background(), query.Query{
		Prefix: key,
		Limit:  limit.Limit,
		Offset: limit.Offset,
	})
	if err != nil {
		return nil, err
	}
	entries, _ := rsl.Rest()
	logger.Debugf("EntriesLen for key: %v,  %v", key, len(entries))
	for _, entry := range entries {
		keyString := strings.Split(entry.Key, "/")
		id := keyString[len(keyString)-1]

		value, qerr := GetSubnetStateById(id)
		if qerr != nil {
			logger.Debugf("KeyString for key: %v", qerr)
			continue
		}
		data = append(data, value)
		err = qerr
	}
	if err != nil {
		return nil, err
	}
	return data, err
}


