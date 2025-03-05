package service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/mlayerprotocol/go-mlayer/smartlet"
	"github.com/sirupsen/logrus"
)

var (
	ErrorKeyNotFound = errors.New("key not found")
	ErrorKeyExist    = errors.New("key already exists")
)








// type readOnlyKeystore interface {
// 	Get(keys [][16]byte) (result map[(smartlet.Key)][]byte, err error)
// 	List(key [16]byte, prefix []byte) (result [][]byte, err error)
// }

// type readWriteKeystore interface {
// 	Get(keys [][16]byte) (result map[(smartlet.Key)][]byte, err error)
// 	List(key [16]byte, prefix []byte) (result [][]byte, err error)
// 	Set(key smartlet.Key, value []byte, replaceIfExists bool) error
// 	Append(key smartlet.Key, value [16]byte, prefix []byte) error
// }




const (
	COUNT_KEY string = "1"
)



type smartletKeyStore struct {
	db         *badger.DB
	 UpdateData *[]smartlet.UpdateData
	logger     *logrus.Logger
	app        *smartlet.App
	mutex          *sync.Mutex
	finalized  bool
}



func newKeyStore(app *smartlet.App, storeDb  *badger.DB) smartletKeyStore {
	return smartletKeyStore{storeDb, &[]smartlet.UpdateData{},  logger, app, &sync.Mutex{}, false}
}


func (ks smartletKeyStore) GetUpdateKeys() (keys []smartlet.Key) {
	for _, d := range *ks.UpdateData {
		for k, _ := range d.Data {
			keys = append(keys, smartlet.Key(k))
		}
		
	}
	return keys
}

func(ks smartletKeyStore) IsFinalized() bool {
	return ks.finalized
}

func (ks smartletKeyStore) Commit() (err error) {
	return ks.update(true)
}

func (ks smartletKeyStore) Get(keys [][16]byte) (result map[(smartlet.Key)][]byte, err error) {
	err = ks.db.View(func(txn *badger.Txn) error {
		for _, key := range keys {
			ksKey := toKeystoreKey(key[:])
			item, err := txn.Get(append([]byte(fmt.Sprintf("%s/%s", ks.app.Topic.Subnet, ks.app.Topic.ID)), ksKey...))
			if err != nil && badger.ErrKeyNotFound != err {
				ks.logger.Debugf("ApplicationGetError: %s", err)
				return err
			}
			if err := item.Value(func(val []byte) error {
				result[key] = val
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ks smartletKeyStore) List(key [16]byte, prefix []byte) (result [][]byte, err error) {
	err = ks.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // Only keys, no values
		opts.Reverse = true

		it := txn.NewIterator(opts)
		defer it.Close()
		searchKey := append([]byte(fmt.Sprintf("%s/", ks.app.Topic.ID)), key[:]...)
		if len(prefix) > 0 {
			searchKey = append(searchKey, append([]byte("/"), toKeystoreKey(prefix)...)...)
		}
		searchKey = append(searchKey, append([]byte("/"), []byte("~~~~999999999999")...)...)
		// Seek to the prefix
		for it.Seek(searchKey); it.ValidForPrefix(searchKey); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			bytes.LastIndex(key, []byte("/"))
			result = append(result, key[bytes.LastIndex(key, []byte("/"))+1:])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ks smartletKeyStore) Set(key smartlet.Key, value []byte, replaceIfExists bool) (err error) {
	// for key, value := range data {
	map1 := map[smartlet.Key]interface{}{}
	// for k, v := range data {
	// 	map1[k] = v
	// }
	map1[key] = value
	
	*ks.UpdateData = append(*ks.UpdateData, smartlet.UpdateData{Data: map1, Replace: replaceIfExists, Prefix: nil})
	return ks.update(false)
}

// func (app *app) Append(data map[(Key)][]byte, replaceIfExists bool) ( err error) {
func (ks smartletKeyStore) Append(key smartlet.Key, value [16]byte, prefix []byte) (err error) {
	// for key, value := range data {

	map1 := map [smartlet.Key]interface{}{}
	//for k, v := range data {
	map1[key] = [][]byte{value[:]}

	//}
	if prefix == nil {
		prefix = []byte{}
	}
	*ks.UpdateData = append(*ks.UpdateData, smartlet.UpdateData{Data: map1, Replace: true, Prefix: prefix})
	return ks.update(false)
}



func toKeystoreKey (k []byte) []byte {
	return []byte(hex.EncodeToString(k))
}


func (ks smartletKeyStore) update(commit bool) (err error) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	if ks.finalized {
		return nil
	}
	if commit {
		defer func() {
			ks.finalized = true
		}()
	}
	// finalKeys := []finalKey{}
	
	err = ks.db.Update(func(txn *badger.Txn) error {
		if !commit {
			logger.Info("Discarding.....")
			defer txn.Discard()
		} 
		// else {
		// 	logger.Info("Committing.....")
		// }
		countTracker := map [smartlet.Key]int{}
		txn.Get([]byte(COUNT_KEY))
		logger.Infof("UpdatingStore: %v, %d", commit, len(*ks.UpdateData))
		for _, updateData := range *ks.UpdateData {
			for key, value := range updateData.Data {
				// key1 := []byte(hex.EncodeToString(keyByte[:]))
				// key := [16]byte(key1[:16])
				ksKey := toKeystoreKey(key[:])
				keyByte := append([]byte(fmt.Sprintf("/%s/%s/", ks.app.Topic.Subnet, ks.app.Topic.ID)), ksKey...)
				final := [][]byte{}
				validValue := false
				isArray := false
				if _, ok := value.([]byte); ok {
					final = [][]byte{value.([]byte)}
					validValue = true
				}
				if _, ok := value.([][]byte); ok {
					countKey := append([]byte(fmt.Sprintf("/%s/%s/%s/", ks.app.Topic.Subnet, ks.app.Topic.ID, COUNT_KEY)), ksKey...)
					countTracker[key] = 0
					if countByte, err := txn.Get(countKey); err == nil {
						countByte.Value(func(val []byte) error {
							countTracker[key], err = strconv.Atoi(string(val))
							return nil
						})
					}
					if countTracker[key] >= 200 {
						return fmt.Errorf("List length limit reached")
					}
					countTracker[key]++
					validValue = true
					isArray = true
					final = value.([][]byte)
					updateData.Replace = true
					if updateData.Prefix != nil && len(updateData.Prefix) > 0 {
						
						keyByte = append(keyByte, append([]byte("/"), toKeystoreKey(updateData.Prefix)...)...)
					}
				}
				

				if !validValue {
					return errors.New("value must be byte slice  ([]byte) or array of byte slice ([][]byte)")
				}

				for _, b := range final {
					if updateData.Replace {
						if isArray {
							key :=  append(keyByte, append([]byte("/"), b...)...)
							logger.Infof("AppendingToFinalKeys: %v", commit)
							// finalKeys =append(finalKeys, finalKey{key: &key, value: &[]byte{}})
							if err2 := txn.Set(key, []byte{}); err2 != nil {
								return err2
							}
						} else {
							
							// finalKeys = append(finalKeys, finalKey{key: &keyByte, value: &b })
							
							if err2 := txn.Set(keyByte, b); err2 != nil {
								logger.Errorf("SetError::: %v", err2)
								return err2
							}
							// Verify immediately in same transaction
								_, err := txn.Get(keyByte)
								if err != nil {
									fmt.Printf("Immediate get error: %v\n", err)
								}
						}
					} else {
						if _, err := txn.Get(keyByte); err == nil {
							return ErrorKeyExist
						}
						logger.Infof("AppendingToFinalKeys: %v", commit)
						// finalKeys = append(finalKeys, finalKey{key: &keyByte, value: &b })
						if err2 := txn.Set(keyByte, b); err2 != nil {
							return err2
						}
					}
					 logger.Debugf("TopicKeySmartletKeys: %s", string(keyByte))
				}
			}
		}
		for k, v := range countTracker {
			ksKey := toKeystoreKey(k[:])
			txn.Set(append([]byte(fmt.Sprintf("%s/%s/%s/", ks.app.Topic.Subnet,  ks.app.Topic.ID, COUNT_KEY)),ksKey...), []byte(fmt.Sprint(v)))
		}
		
		if commit {
			logger.Infof("KeyCOunt %d", len(*ks.UpdateData))
		}
			
		// 	// err =  txn.Commit()
		// 	// logger.Infof("KeyCOunt %s", len(finalKeys))
		// 	// err = ks.db.Update(func(txn *badger.Txn) error {
		// 	// 	for _, vv := range finalKeys {
		// 	// 		logger.Infof("SavingKEEEYEEYE %s", string(*vv.key))
		// 	// 		err  = txn.Set(*vv.key, *vv.value)
		// 	// 	}
		// 	// // 	 err  := txn.Set([]byte("/534e4554-0000-0000-0000-000000000000/746f7069-0000-0000-0000-000000000000/911"), []byte{})
		// 	// 	if err != nil {
		// 	// 		logger.Errorf("ERROEGET %v", err)
		// 	// 	}
		// 	// 	return err
		// 	// })
		// 	// logger.Debugf("UPDATEERRORRRR.... %v", err)
		// 	//  err = ks.db.View(func(txn *badger.Txn) error {
		// 	// 	for _, vv := range finalKeys {
				
		// 	// 		_, err  = txn.Get(*vv.key)
		// 	// 		return err
		// 	// 	}
				
		// 	// 	if err != nil {
		// 	// 		logger.Errorf("ERROEGET %v", err)
		// 	// 	}
		// 	// 	return err
		// 	// })
		// 	// logger.Debugf("COMMITING.... %v, %v", err, ks.db.Opts())
		// 	return err
		// } else {
		// 	// logger.Infof("KeyCOuntDiscard %s", len(finalKeys))
		// 	txn.Discard()
		// }
		return err
	})
	if err != nil {
		if !commit && strings.Contains(err.Error(), "discarded txn") {
			return nil
		}
	}
	
	return err
}
