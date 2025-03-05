package stores

import (
	"context"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
)

var logger = &log.Logger

var (
	P2pDhtStore *ds.Datastore
	StateStore *ds.Datastore
	MessageStore *ds.Datastore
	RefStore *ds.Datastore
	SystemStore  *ds.Datastore
	EventStore  *ds.Datastore
	ClaimedRewardStore *ds.Datastore
	NetworkStatsStore *ds.Datastore // to be removed later
	NodeTopicsStore *ds.Datastore
	GlobalHandlerStore *ds.Datastore
	MempoolStore *ds.Datastore
	//CacheStore *ds.Datastore 
)


func InitStores(mainContext *context.Context) (_ctx context.Context,  _stores []*ds.Datastore) {

	P2pDhtStore = ds.New(mainContext,  string(constants.P2PDhtStore))
	ctx := context.WithValue(*mainContext, constants.P2PDhtStore, P2pDhtStore)
	_stores = append(_stores, P2pDhtStore)
	
	EventStore = ds.New(&ctx,   string(constants.EventStore))
	ctx = context.WithValue(ctx, constants.EventStore, EventStore)
	_stores = append(_stores, EventStore)


	StateStore = ds.New(&ctx,   string(constants.ValidStateStore))
	ctx = context.WithValue(*mainContext, constants.ValidStateStore, StateStore)
	_stores = append(_stores, StateStore)

	MessageStore = ds.New(&ctx,   string(constants.MessageStateStore))
	ctx = context.WithValue(*mainContext, constants.MessageStateStore, MessageStore)
	_stores = append(_stores, MessageStore)

	RefStore := ds.New(&ctx,   string(constants.RefDataStore))
	ctx = context.WithValue(ctx, constants.RefDataStore, RefStore)
	_stores = append(_stores, RefStore)


	SystemStore = ds.New(&ctx,  string(constants.SystemStore))
	ctx = context.WithValue(ctx, constants.SystemStore, SystemStore)
	_stores = append(_stores, SystemStore)

	ClaimedRewardStore = ds.New(&ctx,   string(constants.ClaimedRewardStore))
	ctx = context.WithValue(ctx, constants.ClaimedRewardStore, ClaimedRewardStore)
	_stores = append(_stores, ClaimedRewardStore)

	NetworkStatsStore = ds.New(&ctx,   string(constants.NetworkStatsStore))
	ctx = context.WithValue(ctx, constants.NetworkStatsStore, NetworkStatsStore)
	_stores = append(_stores, NetworkStatsStore)

	// CacheStore = ds.New(&ctx,   string(constants.CacheStore))
	// ctx = context.WithValue(ctx, constants.CacheStore, CacheStore)
	// _stores = append(_stores, CacheStore)

	NodeTopicsStore = ds.New(&ctx,   string(constants.NodeTopicsStore))
	ctx = context.WithValue(ctx, constants.NodeTopicsStore, NodeTopicsStore)
	_stores = append(_stores, NodeTopicsStore)

	GlobalHandlerStore = ds.New(&ctx,   string(constants.GlobalHandlerStore))
	ctx = context.WithValue(ctx, constants.GlobalHandlerStore, GlobalHandlerStore)
	_stores = append(_stores, GlobalHandlerStore)

	MempoolStore = ds.New(&ctx,   string(constants.MempoolStore))
	ctx = context.WithValue(ctx, constants.MempoolStore, MempoolStore)
	_stores = append(_stores, MempoolStore)

	return ctx, _stores
}