package node

import (
	"context"
	"fmt"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
)

// Validator node only
// Keep a record of all messages sent within a cycle per subnet
func ArchiveBlocks(ctx *context.Context) error {
	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		return fmt.Errorf("unable to load config from context")
	}
	// lastBlock := uint64(0)
	// if val, err := stores.SystemStore.Get(context.Background(), datastore.NewKey("archBlk")); err != nil {
	// 	if !dsquery.IsErrorNotFound(err) {
	// 		return err
	// 	}
	// 	if err == nil {
	// 		lastBlock = new(big.Int).SetBytes(val).Uint64()
	// 	}
	// }
	currentBlock := chain.NetworkInfo.CurrentBlock
	if currentBlock != nil {
		return dsquery.ArchiveEvents(cfg)
	}
	return nil
}
