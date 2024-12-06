package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"gorm.io/gorm"

	dsquery "github.com/ipfs/go-datastore/query"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto/schnorr"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	p2p "github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
)

// Validator node only
// Keep a record of all messages sent within a cycle per subnet
func TrackReward(ctx *context.Context) {
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !cfg.Validator {
		return
	}
	defer TrackReward(ctx) 
	defer time.Sleep(5 * time.Second)
	
	if !chain.NetworkInfo.Synced {
		return
	}
	
	
	
	// validator := (*cfg).PublicKey  
	

	currentCycle, err := chain.DefaultProvider(cfg).GetCurrentCycle()
	
	if err != nil {
		// wait and retry
		// time.Sleep(5 * time.Second)
		// TrackReward(ctx) 
		logger.Debugf("TrackReward: Unable to get current cycle")
		return;
		
	}
	lastCycleClaimedKey :=  datastore.NewKey("c")
	lastClaimedCycle := uint64(0)
	//batch, err :=	claimedRewardStore.Batch(*ctx)
	// if err != nil {
	// 	panic(err)
	// }
	lastClaimed, err := stores.ClaimedRewardStore.Get(*ctx, lastCycleClaimedKey)
		if err != nil  {
			if err != datastore.ErrNotFound {
				logger.Error(err)
				return;
			}
		} else {
			lastClaimedCycle = encoder.NumberFromByte(lastClaimed)
		}
	if lastClaimedCycle >= currentCycle.Uint64() {
		return
	}
	if lastClaimedCycle + 1 < currentCycle.Uint64() {
		logger.Debug("Tracking Reward Batches...")
	}
	for i := lastClaimedCycle+1; i < currentCycle.Uint64(); i++ {
		// TODO loop through index till no data
		index := 1
		for {
			rewardBatch, err := generateBatch(i, index, ctx)
			if rewardBatch == nil && err == query.ErrorNotFound {
				break
			}
			index++
			if err != nil {
				break
			}
			if rewardBatch.Closed {
				// save batch
				unclaimedBatchKey :=  datastore.NewKey(fmt.Sprintf("u/%s", rewardBatch.Id))
				err :=	stores.ClaimedRewardStore.Put(*ctx, unclaimedBatchKey, rewardBatch.MsgPack())
				if err != nil {
					panic(err)
				}
				go processSentryRewardBatch(*ctx, cfg, rewardBatch)
			}
		}
		stores.ClaimedRewardStore.Set(*ctx, lastCycleClaimedKey,  encoder.NumberToByte(i), true)
}
			
	
	// key := fmt.Sprintf("%s", validator)
	// for {
	// 	result, err := claimedRewardStore.Query(*ctx, query.Query{
	// 		Prefix: key,
	// 	})
	// 	if err != nil {
	// 		continue
	// 	}
	// 	for {
	// 		select {
	// 		case rsl := <-result.Next():
	// 			rsl.Value

	// 		}
	// 	}
		
	// }
}

func generateBatch(cycle uint64, index int, ctx *context.Context) (*entities.RewardBatch, error) {
	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		return nil, fmt.Errorf("reward store not loaded")
	}
	
	subnetList := []models.EventCounter{}
		claimed := false
		err := query.GetManyWithLimit(models.EventCounter{Cycle: &cycle, Validator: entities.PublicKeyString(cfg.PublicKeyEDDHex), Claimed: &claimed }, &subnetList, &map[string]query.Order{"count": query.OrderDec}, entities.MaxBatchSize, index*entities.MaxBatchSize)
		if err != nil {
			return nil, err
		}
		// defer subnetList.Close()
		// logger.Debugf("ListLen: %d", len(subnetList))
		if len(subnetList) == 0 {
			return nil, query.ErrorNotFound //do not change because error string "empty" is checked above
		}
		cost, err := p2p.GetCycleMessageCost(*ctx, cycle)
		if err != nil {
			logger.Errorf("GetCycleMessageCost: %v", err)
			return nil, err
		}
		
		rewardBatch := entities.NewRewardBatch(cfg, cycle, index, cost, len(subnetList), cfg.PublicKeySECP)
		for  _, rsl := range subnetList {
				rewardBatch.Append(entities.SubnetCount{
					Subnet: rsl.Subnet,
				EventCount: *rsl.Count,
				})
				if rewardBatch.Closed {
					break
				}
		}
		return rewardBatch, nil
}

func processSentryRewardBatch(ctx context.Context, cfg *configs.MainConfiguration, batch *entities.RewardBatch) {
	logger.Debugf("Processing Batch....: %v", batch.Id)

	hashNumber :=  new(big.Int).SetBytes(batch.DataHash)
	logger.Debugf("Hash: %s", hashNumber)
		totalLicenses, err := chain.DefaultProvider(cfg).GetSentryActiveLicenseCount(big.NewInt(int64(batch.Cycle)))
		if err != nil {
			logger.Error(err)
			return
		}
		if totalLicenses.Cmp(big.NewInt(0)) == 0 {
			logger.Debugf("No active license found")
			return
		}
		salt := new(big.Int).Mod(hashNumber, big.NewInt(1000))
		salt = new(big.Int).Add(salt, big.NewInt(1))
		logger.Debugf("Salt: %s", new(big.Int).Mod(hashNumber, big.NewInt(1000)))
		startLicence := new(big.Int).Mod(new(big.Int).Div(hashNumber, salt), totalLicenses).Add(big.NewInt(1000), big.NewInt(0))
		
		// var licenses = [15]*big.Int{}
		//1. Identify nodes licences associated with the batch hash
		//2. loop through licence and request commitment from nodes by sending them the batch data. The nodes will call "schnorr.ComputeNonce(pk, msg)"
				// and return the nonce publicKey to you
		noncePubKeys := []*btcec.PublicKey{}
		sentryPubKeys := []*btcec.PublicKey{}
		signatures := [][]byte{}
		operatorPubKeys := []string{}
		max := 30
		if totalLicenses.Uint64() < 120 {
			max = int(totalLicenses.Uint64() / 3)
		}
		if totalLicenses.Uint64() < 3 {
			max = int(totalLicenses.Uint64()) - 1
		}
		logger.Debugf("Max Proofs Needed: %d", max)
		for i := 0; i<max*2; i++ {
			decodedLicence :=  new(big.Int).Mod(utils.Lcg(startLicence.Uint64() + (uint64(i) * salt.Uint64())), totalLicenses)
			decodedLicence = new(big.Int).Add(decodedLicence, big.NewInt(1000))
			logger.Debugf("Start Licence: %s", decodedLicence)
			operatorBytes, err := chain.Provider(cfg.ChainId).GetSentryLicenseOperator(decodedLicence)
			if err != nil {
				logger.Debug(err)
				continue
			} else {
				if len(operatorBytes) > 0 {
					operator := hex.EncodeToString(operatorBytes)
					if operator == hex.EncodeToString(cfg.PublicKeySECP) {
						continue
					}
					// get the operators address from the dht
					
				// validator, err := p2p.GetOperatorMultiAddress(hex.EncodeToString(operator), cfg.ChainId)
				// if err != nil {
				// 	continue
				// }

				
				// clear out the batch data. Its unncessary to send as the receiving node wont use it
				batchCopy := *batch
				tmpBatch := &batchCopy
				tmpBatch.Data = []entities.SubnetCount{}
				
				logger.Debugf("BATCHDATA: %v", batch)
				// request commitment from validator
				payload := p2p.NewP2pPayload(cfg, p2p.P2pActionGetCommitment, tmpBatch.MsgPack())
				
				response, err := payload.SendDataRequest( operator) // nonce public key
				if err != nil {
					logger.Error(err)
					continue
				}
				
				logger.Debugf("IsValid Response: %s, %v", response.Id, response.IsValid(cfg.ChainId))
				
				if response.IsValid(cfg.ChainId) {
					
					noncePubKey, err := btcec.ParsePubKey(response.Data)
					
					if err != nil {
						logger.Errorf("Response Nonce: %v, %v", err, response.Data)
				
						continue
					}
					
					pubKey, err := btcec.ParsePubKey(operatorBytes)
					if err != nil {
						logger.Errorf("Response Nonce: %v, %v", err, response.Data)
				
						continue
					}
					
					noncePubKeys = append(noncePubKeys, noncePubKey)
					sentryPubKeys = append(sentryPubKeys, pubKey)
					operatorPubKeys = append(operatorPubKeys, operator)
					
				}
				
				
				if len(noncePubKeys) == max {
					
					// hash, _ := batch.GetHash(cfg.ChainId)
					// if err != nil {

					// }
					proofData := batch.GetProofData(cfg.ChainId)
					hash, _ := proofData.GetHash()
					_, challenge, commitment := schnorr.ComputeSigningParams(sentryPubKeys, noncePubKeys, hash )
					signReq := entities.SignatureRequestData{
						ProofHash: hash[:],
						// AggPubKey: aggPubKey.SerializeCompressed(),
						Challenge: challenge,
						// Commitment: commitment,
					}
					logger.Debugf("ReceivedAllNonces: challange:%s, commit:%s", hex.EncodeToString(challenge), commitment)
					signingPayload := p2p.NewP2pPayload(cfg, p2p.P2pActionGetSentryProof, signReq.MsgPack())
					if err != nil {
						continue
					}
					for _, validator := range operatorPubKeys {
						signResp, err := signingPayload.SendDataRequest(validator)
						if err != nil {
							continue
						}
						if signResp.Error != "" {
							logger.Errorf("GetProofRequestError %s", signResp.Error)
							continue
						}
						logger.Debugf("IsValid Response For Signature: %s, %v", response.Id, response.IsValid(cfg.ChainId))
						if signResp.IsValid(cfg.ChainId) {
							// we have received a valid signature from this node
							signatures = append(signatures, signResp.Data)
						}
					}
					if len(signatures) == max { // all 20 have signed
						// aggregate signature
						logger.Debugf("Signatures: %v", signatures)
						aggSig := schnorr.AggregateSignatures(signatures)
						// _, err := chain.DefaultProvider(cfg).ClaimReward(entities.ClaimData{
						// 	SubnetRewardCount: batch.Data,
						// 	Signature: [32]byte(aggSig),
						// 	Commitment: commitment,
						// 	PubKeys: sentryPubKeys,
						// 	Cycle: batch.Cycle,
						// })
						// if err != nil {
						// 	// retry again
						// 	logger.Errorf("processSentryRewardBatch: %v", err)
						// } else {
							// mark it as finalized
							// store it as a pending claim
							signers := [][]byte{}
						logger.Debugf("RPCCONFIG: %v", cfg.EvmRpcConfig["31337"])		
							for _, k := range sentryPubKeys {
								signers = append(signers, k.SerializeCompressed())
							}
							proofData.Signature  = aggSig	
							proofData.Signers = signers
							proofData.Commitment = []byte(commitment)
							pendingClaimsKey :=  datastore.NewKey(fmt.Sprintf("validClaim/%s",  hex.EncodeToString(hash[:])))
							err = stores.ClaimedRewardStore.Put(ctx, pendingClaimsKey, batch.MsgPack())
							if err != nil {
								logger.Error(err)
							} else {
								for _, d := range batch.Data {
									logger.Debugf("[")
									logger.Debugf("{'subnetId':'0x%s', 'amount':'%s'},", strings.ReplaceAll(d.Subnet, "-", ""), new(big.Int).SetBytes(d.Cost) )
									logger.Debugf("]")
								}
								for _, k := range sentryPubKeys {
									logger.Debugf("[")
									logger.Debugf("{'x':'%s','y':'%s'},", k.X(), k.Y() )
									logger.Debugf("]")
								}
								logger.Debug("Cycle: ", batch.Cycle)
								logger.Debug("Index: ", batch.Index)
								logger.Debug("Cost: ", new(big.Int).SetBytes(batch.TotalValue))
								logger.Debug("Validator: ", hex.EncodeToString(batch.Validator))
								logger.Debug("DataHash: ", hex.EncodeToString(proofData.DataHash))
								logger.Debug("ProofHash: ", hex.EncodeToString(hash[:]))
								logger.Debug("commitment: ", hex.EncodeToString(proofData.Commitment))
								logger.Debug("signature: ", hex.EncodeToString(proofData.Signature))
								logger.Debugf("processSentryRewardBatch: Successful..... %d/%s, %v, %v", batch.Cycle, batch.Id, sentryPubKeys[0], commitment)
							}
						// }

					}
					
					break
				}
			}
			
		}
		// get the node that owns the license
	}
					
	//3. make a list of the public keys 20 nodes that respond. then get the signing parameters with schnoor.ComputeSigningParams(pks,noncePks,message)
	//4. send the signing params to each node on the list. Each node will call ComputeSignature the return the signature
	//5. aggregate the signatures with schnorr.AggregateSignatures()

}

// Validator only
// Claim rewards after each cycle
func ProcessPendingClaims(ctx *context.Context) {
	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		logger.Errorf("error: failed loading config")
	}
	if !cfg.Validator {
		return
	}
	waitPeriod := 10 * time.Second
	defer ProcessPendingClaims(ctx) 
	defer time.Sleep(waitPeriod)
	// if !chain.NetworkInfo.Synced {
	// 	return
	// }
	
	
	pendingClaimsKey :=  datastore.NewKey("validClaim/")

	results, err := stores.ClaimedRewardStore.Query(*ctx, dsquery.Query{
		Prefix: pendingClaimsKey.String(),
	})
	
	if err != nil {
		logger.Error("ProcessPendingClaim/claimRwardStore", err)
		return
	}
	
	batchLoop:
	for {
			result, ok := <-results.Next()
			if !ok {
				return
			}
			
			
			batch, err := entities.UnpackRewardBatch(result.Entry.Value)
			if err != nil {
				continue
			}
			
			
			proofData := batch.GetProofData(batch.ChainId)
			signers := []schnorr.Point{}
			sIds := []string{}
			for _, sub := range batch.Data {
				sIds = append(sIds, sub.Subnet)
			}
			for _, p := range proofData.Signers {
				pubK, err := btcec.ParsePubKey(p)
				if err != nil {
					continue batchLoop
				}
				signers = append(signers, schnorr.Point{X: pubK.X(), Y: pubK.Y()})
			}
			
			cycle := new(big.Int).SetBytes(utils.ToUint256(big.NewInt(int64(proofData.Cycle))))
			index :=  new(big.Int).SetBytes(utils.ToUint256(big.NewInt(int64(proofData.Index))))
			claimed, err := chain.Provider(cfg.ChainId).Claimed(batch.Validator, cycle, index);
			if err != nil {
				continue
			}
			
			if !claimed {
				claimData := &entities.ClaimData{
					Cycle: new(big.Int).SetBytes(utils.ToUint256(big.NewInt(int64(proofData.Cycle)))),
					Signature: proofData.Signature,
					Commitment: proofData.Commitment,
					Signers: signers,
					Index: new(big.Int).SetBytes(utils.ToUint256(big.NewInt(int64(proofData.Index)))),
					Validator: proofData.Validator,
					ClaimData: batch.Data,
					TotalCost: new(big.Int).SetBytes(utils.ToUint256(new(big.Int).SetBytes(proofData.TotalCost))),
				}
				
				_, err = chain.Provider(cfg.ChainId).ClaimReward(claimData)
				
				
			}
			
			if claimed || err == nil {
				stores.ClaimedRewardStore.Delete(*ctx, datastore.NewKey(result.Key))
				var claimed = true
				result := sql.SqlDb.Where(
					models.EventCounter{Cycle: &batch.Cycle,
						Validator: entities.PublicKeyString(cfg.PublicKeyEDDHex),
						}).Where("subnet IN ?", sIds).Updates(models.EventCounter{Claimed: &claimed})
				if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
					continue
				}
			}
	}
}

// get the cycle from the store else from the network
