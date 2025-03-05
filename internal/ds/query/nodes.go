package query

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
)


func GetInterestedNodes( id string, callback func (publicKey string ) ) (error) {
	return stores.NodeTopicsStore.DB.View(
		
		func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 5000
			opts.PrefetchValues = false // Fetch both keys and values
			it := txn.NewIterator(opts)
			defer it.Close()

			for  it.Seek([]byte(fmt.Sprintf("%s/", id))); it.Valid(); it.Next() {
				item := it.Item()
				k := item.Key()
				parts := strings.Split(string(k), "/")
				go callback(parts[1])
			}
			return nil
		})
}


func IsNodeInterested (publickey string, id string,  modelType entities.EntityModel) (bool, error) {
	_, err := stores.NodeTopicsStore.Get(context.Background(),  datastore.NewKey(fmt.Sprintf("/%s/%s/%s", id, modelType, publickey)))
	if err != nil {
		if IsErrorNotFound(err) {
			return false, nil
		} else {
			return false,  err
		}
	}
	return true, err
}

func SetNodeInterest( publicKey string, ids []string, modelType entities.EntityModel) (error) {
	batch, err := stores.NodeTopicsStore.Batch(context.Background())
	if err !=nil {
		return err
	}
	for _, id := range ids {
		err := batch.Put(context.Background(),  datastore.NewKey(fmt.Sprintf("/%s/%s/%s", id, modelType, publicKey)),  []byte{})
		if err != nil && err != ds.ErrKeyExists {
			return err
		}
	}
	return batch.Commit(context.Background())
	
}

func UnsetNodeInterest( publicKey string, topic string) (error) {
	return stores.NodeTopicsStore.Delete(context.Background(), datastore.NewKey(fmt.Sprintf("/%s/%s", topic, publicKey)))
}

func SetOwnInterests(ids []string, modelType entities.EntityModel) (error) {
	batch, err := stores.NodeTopicsStore.Batch(context.Background())
	if err !=nil {
		return err
	}
	exp := strconv.Itoa(int(time.Now().UnixMilli())+int(constants.TOPIC_INTEREST_TTL.Milliseconds()))
	for _, id := range ids {
		err := batch.Put(context.Background(),  datastore.NewKey(fmt.Sprintf("/%s/%s", modelType, id)),  []byte(exp))
		if err != nil && err != ds.ErrKeyExists {
			logger.Errorf("SetOwnInterests/Put: %v", err)
			return err
		}
	}
	return batch.Commit(context.Background())
}

func FilterOwnInterests(ids []string, modelType entities.EntityModel) ([]string, error) {
	intr := []string{}
	for _, id := range ids {
		v, err := stores.NodeTopicsStore.Get(context.Background(),  datastore.NewKey(fmt.Sprintf("/%s/%s", modelType, id)))
		if err != nil {
			if  !IsErrorNotFound(err) {
				logger.Errorf("FilterOwnInterests/Put: %v", err)
				return nil, err
			}
			continue
		}
		exp, err := strconv.Atoi(string(v))
		if err == nil {
			if int(time.Now().UnixMilli()) > exp - int((10 * time.Minute).Milliseconds()) {
				continue
			}
		}
		intr = append(intr, id)
	}
	return intr, nil
}

func IsInterestedIn(id string, modelType entities.EntityModel) (bool, error) {
	v, err := stores.NodeTopicsStore.Get(context.Background(),  datastore.NewKey(fmt.Sprintf("/%s/%s", modelType, id)))
	if err != nil {
		if  !IsErrorNotFound(err) {
			logger.Errorf("FilterOwnInterests/Put: %v", err)
			return false, err
		}
		return false, nil
	}
	exp, err := strconv.Atoi(string(v))
		if err == nil {
			if int(time.Now().UnixMilli()) > exp - int((10 * time.Minute).Milliseconds()) {
				return false, nil
			}
		}
	return true, nil
}

func UnsetOwnInterest( id string, modelType entities.EntityModel) (error) {
	return stores.NodeTopicsStore.Delete(context.Background(), datastore.NewKey(fmt.Sprintf("/%s/%s", modelType, id)))
}

