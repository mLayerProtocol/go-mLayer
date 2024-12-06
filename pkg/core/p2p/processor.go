package p2p

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/channelpool"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto/schnorr"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
	"github.com/multiformats/go-multiaddr"
	"github.com/quic-go/quic-go"
	// rest "messagingprotocol/pkg/core/rest"
	// dhtConfig "github.com/libp2p/go-libp2p-kad-dht/internal/config"
)

/**

**/
// func isChannelClosed(ch interface{}) bool {
// 	// Reflect on the channel to check its state
// 	c := reflect.ValueOf(ch)
// 	if c.Kind() != reflect.Chan {
// 		return false
// 	}
// 	_, ok := c.TryRecv()
// 	return !ok
// }

/*
**
Publish Events to a specified p2p broadcast channel
****
*/
type CertResponseData struct {
	CertHash json.RawMessage `json:"cer"`
	QuicHost string          `json:"quic"`
}



func publishChannelEventToNetwork(channelPool chan *entities.Event, pubsubChannel *entities.Channel, mainCtx *context.Context) {
	_, cancel := context.WithCancel(context.Background())
	cfg, ok := (*mainCtx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	defer cancel()
	if !ok {
		logger.Fatalf("Unable to read config")
		return
	}
	for {
		if pubsubChannel == nil {
			continue
		}
		event, ok := <-channelPool

		if !ok {
			logger.Fatalf("Channel pool closed. %v", &channelPool)
			panic("Channel pool closed")
		}
		if cfg.Validator {
			if !ok {
				logger.Errorf("Outgoing channel closed. Please restart server to try or adjust buffer size in config")
				return
			}
			pack := event.MsgPack()
			// if err != nil {
			// 	logger.Error(err)
			// 	continue
			// }

			// event, errT := entities.UnpackEvent(pack, &entities.Authorization{})
			// if errT != nil {
			// 	logger.Errorf("Error receiving event  %v\n", errT)
			// 	continue;
			// }

			// eT := entities.Event{
			// 	Payload: ,
			// }
			// // auth := models.AuthorizationEvent{}
			// err = entities.UnpackEvent(pack, &eT)
			// if err != nil {
			// 	logger.Errorf("Failed to UNPCAKC %v", err)
			// 	continue
			// }
			//payload := entities.AuthorizationPayload{}
			// dbByte, _ := json.Marshal(auth.Event.Payload)
			// _	= json.Unmarshal(dbByte, &payload)

			// auth.Payload = payload
			//  auth.Payload.ClientPayload.Data = auth.Payload.ClientPayload.Data
			//  logger.Debugf("Payload----> %v", event.Payload.ClientPayload.Data)
			// 	// auth.Event.Payload = payload
			// 	b, err := (auth).EncodeBytes()
			// 	if err != nil {
			// 		logger.Errorf("Failed to ENCODE %v", err)
			// 		continue
			// 	}

			err := pubsubChannel.Publish(entities.NewPubSubMessage(pack))

			if err != nil {
				logger.Errorf("Unable to publish message. Please restart server to try again or adjust buffer size in config. Failed with error %v", err)
				return
			}

		}
	}

}

func ProcessEventsReceivedFromOtherNodes(modelType entities.EntityModel, fromPubSubChannel *entities.Channel, mainCtx *context.Context) {
	// time.Sleep(5 * time.Second)

	_, cancel := context.WithCancel(context.Background())

	defer cancel()

	for {
		if fromPubSubChannel == nil || fromPubSubChannel.Messages == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		logger.Infof("Listening for broadcasted \"%s\" events...", modelType)

		message, ok := <-fromPubSubChannel.Messages
		if !ok {
			logger.Fatalf("Primary Message channel closed. Please restart server to try or adjust buffer size in config")
			return
		}
		

		event, errT := entities.UnpackEvent(message.Data, modelType)
		logger.Debugf("ReceivedEvent \"%s\" in Subnet: %s", event.ID, event.Subnet)
		if errT != nil {
			logger.Errorf("Error receiving event  %v\n", errT)
			continue
		}

	
		logger.Debugf("ProcessingEvent \"%s\" in Subnet: %s", event.ID, event.Subnet)
		// event.ID, _ = event.GetId()
		channelpool.EventProcessorChannel <- event
		// go process(event, mainCtx)
	}
	logger.Debugf("Done processing event: %s", modelType)
	// for {
	// 	select {

	// 	case authEvent, ok := <-authorizationPubSub.Messages:
	// 		if !ok {
	// 			cancel()
	// 			logger.Fatalf("Primary Message channel closed. Please restart server to try or adjust buffer size in config")
	// 			return
	// 		}
	// 		// !validating message
	// 		// !if not a valid message continue
	// 		// _, err := inMessage.MsgPack()
	// 		// if err != nil {
	// 		// 	continue
	// 		// }
	// 		//TODO:
	// 		// if not a valid message, continue

	// 		logger.Debugf("Received new message %s\n", authEvent.ToString())
	// 		cm := models.AuthorizationEvent{}
	// 		err = encoder.MsgPackUnpackStruct(authEvent.Data, cm)
	// 		if err != nil {

	// 		}
	// 		*incomingAuthorizationC <- &cm
	// 	case inMessage, ok := <-batchPubSub.Messages:
	// 		if !ok {
	// 			cancel()
	// 			logger.Fatalf("Primary Message channel closed. Please restart server to try or adjust buffer size in config")
	// 			return
	// 		}
	// 		// !validating message
	// 		// !if not a valid message continue
	// 		// _, err := inMessage.MsgPack()
	// 		// if err != nil {
	// 		// 	continue
	// 		// }
	// 		//TODO:
	// 		// if not a valid message, continue

	// 		logger.Debugf("Received new message %s\n", inMessage.ToString())
	// 		cm, err := entities.MsgUnpackClientPayload(inMessage.Data)
	// 		if err != nil {

	// 		}
	// 		*incomingMessagesC <- &cm
	// 	case sub, ok := <-subscriptionPubSub.Messages:
	// 		if !ok {
	// 			cancel()
	// 			logger.Fatalf("Primary Message channel closed. Please restart server to try or adjust buffer size in config")
	// 			return
	// 		}
	// 		// logger.Debug("Received new message %s\n", inMessage.Message.Body.Message)
	// 		cm, err := entities.UnpackSubscription(sub.Data)
	// 		if err != nil {

	// 		}
	// 		logger.Debug("New subscription updates:::", string(cm.ToJSON()))
	// 		// *incomingMessagesC <- &cm
	// 		cm.Broadcasted = false
	// 		*publishedSubscriptionC <- &cm
	// 	}
	// }
}

type IState interface {
	MsgPack() []byte
}

// process the payload based on the type of request
func processP2pPayload(config *configs.MainConfiguration, payload *P2pPayload, mustSign bool) (response *P2pPayload, err error) {
	ctx := MainContext

	response = NewP2pPayload(config, P2pActionResponse, []byte{})
	response.Id = payload.Id
	switch payload.Action {
	case P2pActionGetEvent:
		eventPath, err := entities.UnpackEventPath(payload.Data)
		if err != nil {
			response.ResponseCode = 500
			response.Error = "Invalid payload data"
			logger.Debugf("processP2pPayload: %v", err)
		}
		event, err := dsquery.GetEventFromPath(eventPath)
		if err != nil {
			logger.Errorf("EventFromPathError: %v,%v", err, eventPath)
			if dsquery.IsErrorNotFound(err) {
				response.ResponseCode = 404
				response.Error = "Event not found"
			} else {
				response.ResponseCode = 500
				response.Error = err.Error()
			}
		} else {
			logger.Infof("FOUND_EVENT %s", event.ID)
			// d := models.GetStateModelFromModelType(eventPath.Model)
			//result := []IState{}
			states := []json.RawMessage{}
			state, err := dsquery.GetStateBytesFromEventPath(eventPath)
			if err != nil && !dsquery.IsErrorNotFound(err) {
				logger.Errorf("GettingEventStateError: %v", err)
			}
			if state != nil {
				states = append(states, state)
			}
			
			// if err == nil {
			// 	for _, st := range result {
			// 		states = append(states, st.MsgPack())
			// 	}
			
				data := P2pEventResponse{Event: event.MsgPack(), States: states}
			// 	logger.Debugf("EventReseponse: %v", (&data).MsgPack())
			 	response.Data = (&data).MsgPack()
			// }
		}
	case P2pActionGetState:
		logger.Error("ReceivedGetState Request...")
		ePath, err := entities.UnpackEntityPath(payload.Data)
		if err != nil {
			response.ResponseCode = 500
			response.Error = "Invalid payload data"
			logger.Debugf("processP2pPayload: %v", err)
		}
		logger.Errorf("ReceivedGetState Request... %v", ePath)
		state, err := dsquery.GetStateFromEntityPath(ePath)
		
		if err != nil {
			logger.Errorf("P2pActionGetState: %v", err)
			if dsquery.IsErrorNotFound(err) {
				response.ResponseCode = 404
				response.Error = "State not found"
			} else {
				response.ResponseCode = 500
				response.Error = err.Error()
			}
		} else {
			logger.Errorf("FoundRequestedState.. %v", ePath)
			mapp := map[string]interface{}{}
			err := encoder.MsgPackUnpackStruct(state, &mapp)
			if err != nil {
				response.ResponseCode = 404
				response.Error = "State not found"
				break
			}
			pathMap := mapp["e"]
			logger.Infof("TopicEvent::: %v", pathMap)
			// := entities.EventPathFromString(eventPath)
			//path := eventPath.(entities.EventPath)
			// ev, err := dsquery.GetEventFromPath(&d.Event)
			eventPath := entities.EventPath{EntityPath: entities.EntityPath{
				Hash: string(pathMap.(map[string]string)["h"]),
				Model:  entities.EntityModel(string(pathMap.(map[string]string)["mod"])),
				Validator: entities.PublicKeyString(string(pathMap.(map[string]string)["mod"])),
			}}
			event, err := dsquery.GetEventFromPath(&eventPath)
			if err == nil {
				states := []json.RawMessage{}
				states = append(states, state)
				data := P2pEventResponse{Event: event.MsgPack(), States: states}
				response.Data = (&data).MsgPack()
			} else {
				response.ResponseCode = 500
				response.Error = err.Error()
			}

		}
	case P2pActionSyncCycle:

		blocks := Range{}
		encoder.MsgPackUnpackStruct(payload.Data, &blocks)

		var buffer bytes.Buffer

		// fromBlock :=  new(big.Int).SetBytes(blocks.From)
		// toBlock :=  new(big.Int).SetBytes(blocks.To)
		// var where =  fmt.Sprintf("block_number >= %d AND block_number <= %d",  fromBlock.Uint64(), toBlock.Uint64())
		//  var where = "1=1"
		
		endBlock := new(big.Int).Sub(chain.NetworkInfo.CurrentBlock, big.NewInt(5))
		from := new(big.Int).SetBytes(blocks.From)
		if from.Cmp(endBlock) == 1 {
				// Path does not exist
				response.ResponseCode = apperror.BadRequestError
				response.Error = "invalid from in range"
				break
		}
		to := new(big.Int).SetBytes(blocks.To)
		if to.Cmp(endBlock) == 1 {
			to = endBlock
		}
		fromCycle, err := chain.DefaultProvider(cfg).GetCycle(from)
			if err != nil {
				response.ResponseCode = apperror.InternalError
				response.Error = "could not get block cycle"
				break
			}
			
		toCycle, err := chain.DefaultProvider(cfg).GetCycle(to)
			if err != nil {
				response.ResponseCode = apperror.InternalError
				response.Error = "could not get block cycle"
				break
			}

		for i := fromCycle.Uint64(); i <= toCycle.Uint64(); i++ {
			if i != 102 {
				continue
			}
	 		cycleDir := filepath.Join(cfg.ArchiveDir, fmt.Sprint(i))
			 _, err = os.Stat(cycleDir)
			 if err != nil && !os.IsNotExist(err) {
				response.ResponseCode = apperror.InternalError
				response.Error = err.Error()
				break
			}
			 files, err := utils.ListFilesInDir(cycleDir)
			 if err != nil {
				response.ResponseCode = 500
				response.Error = err.Error()
			 }
			
			 for _, file := range files {
				if !strings.HasSuffix(file, ".dat") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(cycleDir, file))
				if err != nil {
					fmt.Println("Error reading file:", err)
					response.ResponseCode = 500
					response.Error = err.Error()
				}
				buffer.Write(data)
			 }
			 logger.Infof("SyncCycle: %d, %d, %v", i,  new(big.Int).SetBytes(blocks.To).Uint64(),buffer.Len())
			if err != nil {
					response.ResponseCode = 404
					response.Error = "Event not found"
					break
			}
			
		}

		// for i := from.Uint64(); i <= to.Uint64(); i++ {
		
		// 	cycle, err := chain.DefaultProvider(cfg).GetCycle(new(big.Int).SetUint64(i))
		// 	if err != nil {
		// 		response.ResponseCode = apperror.InternalError
		// 		response.Error = "could not get block cycle"
		// 		break
		// 	}
			
		// 	 cycleDir := filepath.Join(cfg.ArchiveDir, cycle.String())
		// 	 _, err = os.Stat(cycleDir)
		// 	 if err != nil && !os.IsNotExist(err) {
		// 		response.ResponseCode = apperror.InternalError
		// 		response.Error = err.Error()
		// 		break
		// 	}
		// 	 files, err := utils.ListFilesInDir(cycleDir)
		// 	 if err != nil {
		// 		response.ResponseCode = 500
		// 		response.Error = err.Error()
		// 	 }
			
		// 	 for _, file := range files {
		// 		if !strings.HasSuffix(file, ".dat") {
		// 			continue
		// 		}
		// 		data, err := os.ReadFile(filepath.Join(cycleDir, file))
		// 		if err != nil {
		// 			fmt.Println("Error reading file:", err)
		// 			response.ResponseCode = 500
		// 			response.Error = err.Error()
		// 		}
		// 		buffer.Write(data)
				
		// 	 }
		// 	 logger.Infof("SyncBlock: %d, %d, %v", i,  new(big.Int).SetBytes(blocks.To).Uint64(), err)
		// 	if err != nil {
		// 			response.ResponseCode = 404
		// 			response.Error = "Event not found"
		// 			break
		// 	}
			
		// }
			
		response.Data = buffer.Bytes()
		// if err != nil {
		// 	logger.Error("GZIP", err)
		// 	response.ResponseCode = 404
		// 	response.Error = err.Error()
		// }

	case P2pActionGetCommitment:

		realBatch, err := entities.UnpackRewardBatch(payload.Data)
		batchCopy := *realBatch
		batch := &batchCopy
		batch.Clear()
		if err != nil {
			response.ResponseCode = 500
			response.Error = err.Error()
		}
		//  cycleKey :=  fmt.Sprintf("%s/%d", response.Signer, batch.Cycle)
		// subnetList := []models.EventCounter{}
		claimed := false
		subnetList, err := dsquery.GetCycleCounts( batch.Cycle,  entities.PublicKeyString(hex.EncodeToString(payload.Signer)),  &claimed, nil, &dsquery.QueryLimit{Limit: entities.MaxBatchSize, Offset: batch.Index*entities.MaxBatchSize} )
		// err = query.GetManyWithLimit(models.EventCounter{Cycle: &batch.Cycle, Validator: entities.PublicKeyString(hex.EncodeToString(payload.Signer)), Claimed: &claimed}, &subnetList, &map[string]query.Order{"count": query.OrderDec}, entities.MaxBatchSize, batch.Index*entities.MaxBatchSize)
		
		if err != nil {
			return nil, err
		}
		if len(subnetList) == 0 {
			response.ResponseCode = 500
			response.Error = "empty list"
			break
		}
		if subnetList[0].Subnet != realBatch.DataBoundary[0].Subnet {
			response.ResponseCode = 500
			response.Error = "upper data boundary dont match"
			break
		}
		if subnetList[len(subnetList)-1].Subnet != realBatch.DataBoundary[1].Subnet {
			response.ResponseCode = 500
			response.Error = "lower data boundary dont match"
			break
		}

		for _, rsl := range subnetList {
			// if start  == i {
			// if rsl.Subnet != batch.DataBoundary[0].Subnet {
			// 	response.ResponseCode = 500
			// 	response.Error = "data boundary dont match"
			// 	break
			// }
			batch.Append(entities.SubnetCount{
				Subnet:     rsl.Subnet,
				EventCount: *rsl.Count,
			})

			// }
			// if i > start + 99 {
			// 	break
			// }
			// i++
		}

		claimHash := [32]byte{}
		if len(batch.Data) > 0 && len(response.Error) == 0 {
			//logger.Debugf("BATCHINGOF %s", realBatch.GetProofData(config.ChainId).DataHash)
			claimHash, err = realBatch.GetProofData(config.ChainId).GetHash()
			logger.Debugf("ValidDataHash %v, %v", [32]byte(batch.DataHash) == [32]byte(realBatch.DataHash), realBatch)
			if err != nil {
				response.ResponseCode = 500
				response.Error = err.Error()
				logger.Errorf("Error getting hash: %v", err)
			}
			if [32]byte(batch.DataHash) != [32]byte(realBatch.DataHash) {
				response.ResponseCode = 400
				response.Error = "Invalid batch hash"
			}
		} else {
			response.ResponseCode = 400
			response.Error = "Invalid batch hash"
		}

		validCommitmentKey := datastore.NewKey(fmt.Sprintf("commit/%s", hex.EncodeToString(claimHash[:])))
		logger.Debugf("CommitmentKey1: %s", validCommitmentKey.String())

		if response.ResponseCode == 0 {
			pk, _ := btcec.PrivKeyFromBytes(config.PrivateKeySECP)
			nonce, noncePublicKey := schnorr.ComputeNonce(pk, claimHash)
			err = stores.ClaimedRewardStore.Put(*ctx, validCommitmentKey, nonce.Bytes())
			if err != nil {
				logger.Errorf("FailedStoringComittemnt: %v", err)
				response.ResponseCode = 500
				response.Error = "Internal error"
			} else {
				response.Data = noncePublicKey.SerializeCompressed()
			}
			logger.Debugf("NoncePubKey %s", hex.EncodeToString(noncePublicKey.SerializeCompressed()))
		}

	case P2pActionGetSentryProof:
		logger.Debug("ReceivedProoftRequest")

		sigData, err := entities.UnpackSignatureRequestData(payload.Data)

		if err != nil {
			response.ResponseCode = 500
			response.Error = err.Error()
		}

		validCommitmentKey := datastore.NewKey(fmt.Sprintf("commitment/%s", hex.EncodeToString(sigData.ProofHash)))
		logger.Debugf("CommitmentKey2: %s", validCommitmentKey.String())

		nonce, err := stores.ClaimedRewardStore.Get(*ctx, validCommitmentKey)
		if err != nil {
			response.ResponseCode = 500
			response.Error = "Internal error"
			logger.Debugf("Error getting commitment from store")
		}
		if err == nil && response.ResponseCode == 0 {

			pk, _ := btcec.PrivKeyFromBytes(config.PrivateKeySECP)
			// nonce, _ := schnorr.ComputeNonce(pk, [32]byte(sigData.BatchHash))
			sig := schnorr.ComputeSignature(pk, new(big.Int).SetBytes(nonce), sigData.Challenge)
			//  cycleKey :=  fmt.Sprintf("%s/%d", response.Signer, batch.Cycle)
			response.Data = sig
			/// TODO save the nonepublickey with the claimhash in badger
			logger.Debugf("NoncePubKey %s", hex.EncodeToString(sig))
		}

		// if err != nil {
		// 	response.ResponseCode = 500
		// 	response.Error = "Invalid payload data"
		// 	logger.Debugf("processP2pPayload: %v", err)
		// }

		// 1. Get the reward batch data
		// 2. Loop through the Data field and check your /validator/cycle/subnetId/{batchId} to get the last time a proof was requested
		// 3. If this is less than 10 minutes ago, respond with error - proof requested too early
		// 4. If non exists or most recent is more than 10 minutes
	case P2pActionGetCert:
		certData := crypto.GetOrGenerateCert(ctx)
		//cert, _ := hex.DecodeString(certData.Cert)
		keyByte, _ := hex.DecodeString(certData.Key)
		certByte, _ := hex.DecodeString(certData.Cert)
		tlsConfig, _ := crypto.GenerateTLSConfig(keyByte, certByte)
		resp := CertResponseData{
			CertHash: crypto.Keccak256Hash(tlsConfig.Certificates[0].Certificate[0]),
			QuicHost: config.QuicHost,
		}
		response.Data, err = encoder.MsgPackStruct(resp)
		if err != nil {
			response.Error = err.Error()
			response.ResponseCode = 400
			break
		}
		response.Sign(config.PrivateKeyEDD)
	case P2pActionGetHandshake:
		lastSync, err := ds.GetLastSyncedBlock(ctx)
		if err != nil {
			lastSync = big.NewInt(0)
		}
		// TODO  remove when we have several bootstrap nodes
		if cfg.NoSync {
			lastSync = chain.NetworkInfo.CurrentBlock
		}
		handshake, err := NewNodeHandshake(cfg, cfg.ProtocolVersion, cfg.PrivateKeySECP, cfg.PublicKeyEDD, utils.IfThenElse(cfg.Validator, constants.ValidatorNodeType, constants.SentryNodeType), lastSync, payload.Id)
		if err != nil {
			response.Error = "invalid action type"
			response.ResponseCode = 400
			break
		}

		response.Data = handshake.MsgPack()
	default:
		response.Error = "invalid action type"
		response.ResponseCode = 400

	}
	if mustSign && len(response.Signature) == 0 {
		response.Sign(config.PrivateKeyEDD)
	}
	return response, err
}

// func generateImportScript(model any, fromBlock uint64, toBlock uint64) ([]byte, error) {

// 	sql, err := query.GenerateImportScript(sql.SqlDb, models.SubnetEvent{}, sql.SqlDb.Where("block_number >= ? AND block_number <= ?",  fromBlock, toBlock), "", config )
// 				if err != nil {
// 					logger.Debugf("SQLFILEERROR: %v", err)
// 				}
// 				d, err := utils.CompressToGzip(sql)
// 				if err != nil {
// 					return nil, err
// 				}
// 				return d, nil
// }

func HandleQuicConnection(ctx *context.Context, cfg *configs.MainConfiguration, connection quic.Connection) {

	stream, err := connection.AcceptStream(*ctx)

	if err != nil {
		logger.Fatal(err)
	}
	defer stream.Close()

	// Read the client's request (the filename)
	buf := make([]byte, 1024)
	data := bytes.Buffer{}
	for {

		n, err := stream.Read(buf) // Read into the buffer
		data.Write(buf[:n])
		if n < len(buf) || n == 0 || err == io.EOF {
			break // End of file, stop reading
		}
		if err != nil {
			logger.Error(err) // Handle error
			return
		}
	}

	payload, err := UnpackP2pPayload(data.Bytes())

	if err != nil {
		logger.Error(err)
		return
	}
	if !payload.IsValid(cfg.ChainId) {
		logger.Error(fmt.Errorf("HandleQuicConnection: invalid payload signature for action %d", payload.Action))
		return
	}
	response, err := processP2pPayload(cfg, payload, false)
	if err != nil {
		logger.Error(err)
		return
	}
	_, err = stream.Write(response.MsgPack())
	if err != nil {
		log.Fatalf("Failed to send file: %v", err)
	}
}

func parseQuicAddress(cfg *configs.MainConfiguration, maddrs []multiaddr.Multiaddr) (string, multiaddr.Multiaddr, error) {
	var idx int
	var found bool
	for i, addr := range maddrs {
		if len(cfg.SyncHost) > 0 {
			if !strings.Contains(addr.String(), cfg.SyncHost) {
				continue
			}
		} else {
			if  !isRemote(addr) {
				continue
			}
		}
		if strings.Contains(addr.String(), "/quic-v1/") {
			idx = i
			found = true
			break
		}
	}
	if !found {
		for j, addr := range maddrs {
			if strings.Contains(addr.String(), "/quic-v1/") {
				idx = j
				found = true
				break
			}
		}
	}
	if !found {
		return "", nil, fmt.Errorf("invalid quic address")
	}
	ip, err := extractIP(maddrs[idx])
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("%s%s", ip, cfg.QuicHost[strings.Index(cfg.QuicHost, ":"):]), maddrs[idx], nil

}
func extractIP(maddr multiaddr.Multiaddr) (string, error) {

	// Extract the transport protocol and address parts
	components := maddr.Protocols()

	for _, component := range components {
		// Check if the protocol is IP4, IP6, or DNS
		if component.Name == "ip4" || component.Name == "ip6" || component.Name == "dns" {
			// Extract the value for the IP or hostname
			addrValue, err := maddr.ValueForProtocol(component.Code)
			if err != nil {
				return "", err
			}
			return addrValue, nil
		}
	}

	return "", fmt.Errorf("no valid IP or DNS found in the multiaddress")
}
