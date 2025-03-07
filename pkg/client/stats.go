package client

import (
	// "errors"

	"context"
	"math/big"
	"strconv"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"

	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
)

func GetBlockStats(startBlock uint64, limit *entities.QueryLimit) (*[]models.BlockStat, error) {
	var blockStat []models.BlockStat
	

	if limit == nil {
		limit = dsquery.DefaultQueryLimit
	}
	if startBlock <= uint64(limit.Offset) {
		limit.Offset = 0
	}
	if startBlock == 0 {
		startBlock = chain.NetworkInfo.CurrentBlock.Uint64()
	}
	start  := startBlock - uint64(limit.Offset)
	if start <= uint64(limit.Limit) {
		limit.Limit = int(start)
	}
	for i := start; i > start-uint64(limit.Limit); i-- {
		
		v, err := dsquery.GetStats(dsquery.BlockStatsQuery{
			Block: &start, 

		}, nil)
		if err != nil {
			continue
		}
		stat := models.BlockStat{
			BlockStats: entities.BlockStats{
				BlockNumber: i,
				EventCount: v,
			},
		}
		models := []entities.EntityModel{entities.ApplicationModel, entities.AuthModel, entities.TopicModel, entities.SubscriptionModel, entities.MessageModel}
		for _, m := range models {
			c, err := dsquery.GetStats(dsquery.BlockStatsQuery{
				Block: &i, 
				EventType:  &m,
			}, nil)
			if err != nil {
				continue
			}
			switch m {
			case entities.ApplicationModel:
				stat.ApplicationCount = c
			case entities.AuthModel:
				stat.AuthorizationCount = c
			case entities.TopicModel:
				stat.TopicCount = c
			case entities.SubscriptionModel:
				stat.SubscriptionCount = c
			case entities.MessageModel:
				stat.MessageCount = c
			}
			
		}
		blockStat = append(blockStat, stat)
	}
	
	return &blockStat, nil
}

func GetCycleStats(startCycle uint64, limit *entities.QueryLimit) (*[]models.BlockStat, error) {
	var blockStat []models.BlockStat
	

	if limit == nil {
		limit = dsquery.DefaultQueryLimit
	}
	if startCycle <= uint64(limit.Offset) {
		limit.Offset = 0
	}
	if startCycle == 0 {
		startCycle = chain.NetworkInfo.CurrentCycle.Uint64()
	}
	start  := startCycle - uint64(limit.Offset)
	if start <= uint64(limit.Limit) {
		limit.Limit = int(start)
	}
	for i := start; i > start-uint64(limit.Limit); i-- {
		
		v, err := dsquery.GetStats(dsquery.BlockStatsQuery{
			Cycle: &i, 

		}, nil)
		if err != nil {
			continue
		}
		ev := dsquery.GetCycleRecentEvent(i)
		if err != nil {
			continue
		}
		stat := models.BlockStat{
			BlockStats: entities.BlockStats{
				Cycle: i,
				EventCount: v,
				Event: ev,
			},
		}
		// models := []entities.EntityModel{entities.ApplicationModel, entities.AuthModel, entities.TopicModel, entities.SubscriptionModel, entities.MessageModel}
		// for _, m := range models {
		// 	c, err := dsquery.GetStats(dsquery.BlockStatsQuery{
		// 		Cycle: &i, 
		// 		EventType:  &m,
		// 	}, nil)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	switch m {
		// 	case entities.ApplicationModel:
		// 		stat.ApplicationCount = c
		// 	case entities.AuthModel:
		// 		stat.AuthorizationCount = c
		// 	case entities.TopicModel:
		// 		stat.TopicCount = c
		// 	case entities.SubscriptionModel:
		// 		stat.SubscriptionCount = c
		// 	case entities.MessageModel:
		// 		stat.MessageCount = c
		// 	}
			
		// }
		
		blockStat = append(blockStat, stat)
	}
	
	return &blockStat, nil
}

func GetMainStats(cfg *configs.MainConfiguration) (*entities.MainStat, error) {
	// var mainStat []entities.MainStat
	var accountCount uint64
	var agentCount uint64
	// err := query.GetTx().Model(&models.AuthorizationState{}).Group("account").Count(&accountCount).Error
	totalEventsCount, err := dsquery.GetNetworkCounts(nil, dsquery.DefaultQueryLimit)
	if err != nil {
		if dsquery.IsErrorNotFound(err) {
			return &entities.MainStat{}, nil
		}
		return &entities.MainStat{}, err
	}



	agentCounBytes, err  := stores.StateStore.Get(context.Background(), datastore.NewKey(entities.AppKeyCountKey()))
	
	if err != nil && !dsquery.IsErrorNotFound(err) {
		return  &entities.MainStat{}, err
	}
	if len(agentCounBytes) > 0 {
		if agentCountInt, err := strconv.Atoi(string(agentCounBytes)); err != nil {
			return  &entities.MainStat{}, err
		} else {
			agentCount  = uint64(agentCountInt)
		}
	}
	
	// err = query.GetTx().Model(&models.ApplicationState{}).Select("COALESCE(sum(balance), 0)").Row().Scan(&appBalanceTotal)
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, err
	// }
	// err = query.GetTx().Model(&models.MessageState{}).Count(&messages).Error
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, err
	// }

	// err = query.GetTx().Model(&models.MessageState{}).Count(&messageCount).Error
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, err
	// }

	appBal, _ := chain.DefaultProvider(cfg).GetTotalValueLockedInApplications()
	if appBal == nil {
		appBal = big.NewInt(0)
	}
	
	msgCost, err := chain.Provider(cfg.ChainId).GetCurrentMessagePrice()
	if err != nil {
		panic(err)
	}
	
	
	msgCount := uint64(0)
	if len(totalEventsCount) > 0 {
		msgCount = *(totalEventsCount[0].Count)
	} 
	
	accountCount, _ =  dsquery.GetNumAccounts()
	return &entities.MainStat{
		Accounts:  accountCount,
		MessageCost:  msgCost.String(),
		TotalValueLocked: appBal.String(),
		EventCount: msgCount,
		TotalEventsValue: big.NewInt(1).Mul(msgCost, new(big.Int).SetUint64(msgCount)).String(),
		AppKeyCount: agentCount,
	}, nil
}
