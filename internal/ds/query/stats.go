package query

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
)






func IncrementStats(event *entities.Event, tx *datastore.Txn) (err error) {
	ds := stores.NetworkStatsStore
	txn, err  := InitTx(ds, tx)
	if tx == nil {
		defer txn.Discard(context.Background())
	}
	for _, keyString := range entities.GetBlockStatsKeys(event) {
		var count *big.Int
		key := datastore.NewKey(keyString)

		if keyString == entities.RecentEventKey(event.Cycle) {
			err  = txn.Put(context.Background(),  key, []byte(event.ID))
			if err != nil {
				return err
			}
			continue
		}
		
		value, err := txn.Get(context.Background(), key)
		if err != nil {
			if !IsErrorNotFound(err) {
				logger.Infof("STATDSERROR %v", err)
				return err
			}
			count = big.NewInt(1)
		} else {
			count = new(big.Int).Add(new(big.Int).SetBytes(value), big.NewInt(1))
		}
		err  = txn.Put(context.Background(),  key, count.Bytes())
		if err != nil {
			logger.Infof("STATDSERROR2 %v", err)
			return err
		}
	}
	if tx == nil {
		return txn.Commit(context.Background())
	}
	
	return nil
}

type BlockStatsQuery struct {
	Cycle *uint64
	Epoch *uint64
	Block *uint64
	EventType *entities.EntityModel
}

func GetCycleRecentEvent(cycle uint64) string {
	 keyString := entities.RecentEventKey(cycle)
	d, err := stores.NetworkStatsStore.Get(context.Background(), datastore.NewKey(keyString))
	if err != nil {
		return ""
	}
	return string(d)
}
func GetStats(query BlockStatsQuery, limit *QueryLimit) (uint64, error) {
	key := ""
	if query.Block != nil {
		if query.EventType != nil {
			key = fmt.Sprintf("b/%015d/%s", *query.Block, *query.EventType)
		} else {
			key = fmt.Sprintf("b/%015d", *query.Block)
		}
	} else {
		if query.Cycle != nil {
			key = fmt.Sprintf("c/%015d", *query.Cycle)
		} else if query.Epoch != nil {
			key = fmt.Sprintf("e/%015d", *query.Epoch)
		} else {
			if query.EventType != nil {
				key = "m"
			} else {
				key = fmt.Sprintf("m/%015d", *query.Block)
			}
		}
	}

	value, err := stores.NetworkStatsStore.Get(context.Background(), datastore.NewKey(key))
	if err != nil && !IsErrorNotFound(err) {
		return 0, err
	}
	if IsErrorNotFound(err) {
		return 0, nil
	}
	return new(big.Int).SetBytes(value).Uint64(), nil
}

