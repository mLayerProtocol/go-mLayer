package ds

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v4"
	ds "github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
)

func Key(key string) ds.Key {
	return ds.NewKey(key)
}

func New(mainCtx *context.Context, keyStore string) (*Datastore) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, ok := (*mainCtx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		panic(fmt.Errorf("ds.New: could not retrieve config from context"))
		// return nil, fmt.Errorf("ds.New: could not retrieve config from context")
	}
	
	dir := filepath.Join(cfg.DataDir, "store", "kv", keyStore)
	valueLogDir := filepath.Join(cfg.DataDir, "store", "kv", "logs", keyStore)
	if !strings.HasPrefix(dir, "./") && !strings.HasPrefix(dir, "../") && !filepath.IsAbs(dir) {
		dir = "./" + dir
		if strings.HasPrefix(cfg.DataDir, "../") {
			dir = "." + dir
		}
	}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(valueLogDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	opts := badger.DefaultOptions(dir).
        WithValueDir(valueLogDir).
        WithMemTableSize(1 << 30 / 5).
        WithValueLogFileSize(1 << 30).
        WithNumMemtables(10).

		
        WithNumLevelZeroTables(12).
        WithNumLevelZeroTablesStall(20).
        WithValueLogMaxEntries(1000000).
        WithNumCompactors(8).
        WithSyncWrites(false).
        WithLogger(nil). // Disable info logging
     
        WithDetectConflicts(false).
        WithBlockCacheSize(1 << 30 / 2). // 1GB block cache
        WithIndexCacheSize(1 << 30 / 2)  // 1GB index cache


    opts.WithLevelSizeMultiplier(100)       // More aggressive level sizing
    opts.WithBaseTableSize(10 << 20 * 100)       // 10MB
    opts.WithValueThreshold(1024)          // Store values > 1KB in value log
    opts.WithNumVersionsToKeep(1)          // Keep only latest version
    opts.WithBloomFalsePositive(0.01)    
	opts.WithLoggingLevel(badger.INFO)

	opt := DefaultOptions
	opt.Options = opts
	
	
	ds, err := NewDatastore(dir, &opt)
	if err != nil {
		panic(err)
	}
	return ds

	// err = db.View(func(txn *badger.Txn) error {
	// 	_, err := txn.Get([]byte("key"))
	// 	// We expect ErrKeyNotFound
	// 	fmt.Println("Error", err)
	// 	return nil
	// })

	// if err != nil {
	// 	panic(err)
	// }

	// txn := db.NewTransaction(true) // Read-write txn
	// err = txn.SetEntry(badger.NewEntry([]byte("key"), []byte("value set successfully")))
	// if err != nil {
	// 	panic(err)
	// }
	// err = txn.Commit()
	// if err != nil {
	// 	panic(err)
	// }

	// err = db.View(func(txn *badger.Txn) error {
	// 	item, err := txn.Get([]byte("key"))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	val, err := item.ValueCopy(nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Printf("Valueeeee %s\n", string(val))
	// 	return nil
	// })

	// if err != nil {
	// 	panic(err)
	// }
}
