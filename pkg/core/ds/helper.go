package ds

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
)

const (
	SyncedBlockKey = "/syncedBlock"
)
func GetLastSyncedBlock(ctx *context.Context) (*big.Int, error) {
	SystemStore, ok := (*ctx).Value(constants.SystemStore).(*Datastore)
	if !ok {
		return big.NewInt(0), fmt.Errorf("GetLastSyncBlock: unable to load SystemStore from context")
	}
	lastBlockByte, err := SystemStore.Get(*ctx, Key(SyncedBlockKey))
	if err != nil && err != datastore.ErrNotFound {
		return big.NewInt(0), fmt.Errorf("GetLastSyncedBlock: %v", err)
	}
	return new(big.Int).SetBytes(lastBlockByte), nil
}

func SetLastSyncedBlock(ctx *context.Context, value *big.Int) (error) {
	SystemStore, ok := (*ctx).Value(constants.SystemStore).(*Datastore)
	if !ok {
		return  fmt.Errorf("SetLastSyncedBlock: unable to load SystemStore from context")
	}
	return SystemStore.Set(*ctx, Key(SyncedBlockKey), value.Bytes(), true)
}