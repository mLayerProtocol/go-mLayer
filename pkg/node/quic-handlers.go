package node

import (
	"bytes"
	"context"
	"encoding/hex"

	"fmt"
	"io"
	"log"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/system"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
	"github.com/quic-go/quic-go"
)

func HandleQuicConnection(ctx *context.Context, cfg *configs.MainConfiguration, connection quic.Connection) {
	for {
		logger.Infof("NewRemoteStreamStarted: %s", connection.RemoteAddr().String())
		newStream, err := connection.AcceptStream(*ctx)
		if err != nil {
			if  _, ok := err.(*quic.StreamError); ok {
				continue
			} else {
				return
			}
		}
		go func(stream quic.Stream) {
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
					logger.Error("HandleQuicConnection/writedata: ", err) // Handle error
					return
				}
			}
			if len(data.Bytes()) == 0 {
				return
			} 
			payload, err := p2p.UnpackP2pPayload(data.Bytes())

			if err != nil {
				
				logger.Errorf("HandleQuicConnection/UnpackP2pPayload: %v, %d", err, data.Len())
				return
			}
			// if !payload.IsValid(cfg.ChainId) {
			// 	logger.Error(fmt.Errorf("HandleQuicConnection: invalid payload signature for action %d", payload.Action))
			// 	return
			// }
			p2p.NodeQuicPool.AddConnection(*ctx, connection.RemoteAddr().String(), hex.EncodeToString(payload.Signer), connection)
			response := p2p.NewP2pPayload(cfg, p2p.P2pActionResponse, []byte{})
			response.Id = payload.Id
			switch payload.Action {
			case p2p.P2pActionPostEvent:
				eventPayload := entities.EventPayload{}
				err := encoder.MsgPackUnpackStruct(payload.Data, &eventPayload)
				if err == nil {
					event, err := entities.UnpackEvent(eventPayload.Event, entities.GetModelTypeFromEventType(eventPayload.EventType))
					if err == nil {
						resp, err := service.HandleNewPubSubEvent(*event, cfg.Context)
						if err == nil {
							stateBytes, err := encoder.MsgPackStruct(resp.State)
							// stateBytes, err := dsquery.GetStateBytesFromEventPath(event.GetPath())
							if err == nil {
								delta := map[string]interface{}{}
								state := map[string]interface{}{}
								prevState := map[string]interface{}{}
								encoder.MsgPackUnpackStruct(stateBytes, &state)
								var previousStateBytes []byte
								logger.Infof("STATEMAP %+v", resp.State)
								if event.PreviousEvent.ID != "" && event.PreviousEvent.EntityPath.Model == event.GetPath().Model {
									previousStateBytes, err = dsquery.GetStateBytesFromEventPath(&event.PreviousEvent)
									logger.Infof("STATEMAP %v", previousStateBytes)
									if err == nil {
										encoder.MsgPackUnpackStruct(previousStateBytes, &prevState)
										delta = utils.GetDifference(prevState, state)
									} else {
										delta = state
									}
								} else {
									delta = state
								}
								b, err := encoder.MsgPackStruct(delta)
								 if err == nil {
									// logger.Infof("STATEDELTA %v", delta)
									// pack, err := encoder.MsgPackStruct(delta)
									//if err == nil {
										var previousHash []byte
										if len(fmt.Sprint(prevState["h"])) > 0 {
											previousHash, _ = hex.DecodeString(fmt.Sprint(prevState["h"]))
										}
										resp := entities.EventDelta{Delta: b, PreviousHash: previousHash}
										resp.Hash, _ = resp.GetHash()
										sig, _ := crypto.SignSECP(resp.Hash, cfg.PrivateKeySECP)
										resp.Signatures = []entities.SignatureData{
											{Signature: entities.HexString(hex.EncodeToString(sig)), PublicKey: entities.PublicKeyString(cfg.PublicKeySECPHex)},
										}
										// resp.Signature, _ = crypto.SignSECP(resp.Hash, cfg.PrivateKeySECP)
										// resp.Validator = cfg.PublicKeySECP
										response.Data, _ = encoder.MsgPackStruct(resp)
									//}

								 }
							}
						}
					}
				}
				if err != nil {
					logger.Errorf("ErrorProccessingEvent: %v", err)
					response.Error = err.Error()
					response.ResponseCode = 500
				}
			case p2p.P2pActionNotifyValidEvent:
				
				eventPath := entities.EventPath{}
				err := encoder.MsgPackUnpackStruct(payload.Data, &eventPath)
				
				if err == nil {
					if v, err := system.Mempool.GetData(eventPath.ID); err == nil {

						dstate := dsquery.DataStates{}
						err = encoder.MsgPackUnpackStruct(*v, &dstate)
						if err == nil {
							logger.Infof("P2pActionNotifyValidEvent %+v", dstate)
							err = dstate.Commit(nil, nil, nil, eventPath.ID, err)
							if err == nil {
								response.Data = eventPath.MsgPack()
							}
						} else {
							logger.Errorf("P2pActionNotifyValidEvent %v", err)
						}
					}
				} else {
					logger.Errorf("P2pActionNotifyValidEventDsError %v", err)
				}
				if err != nil {
					logger.Errorf("P2pActionNotifyValidEvent %v", err)
					response.Error = err.Error()
					response.ResponseCode = 500
				}
			case p2p.P2pActionSyncState:
				logger.Infof("SyncingValidState")
				stateDelta := entities.EventDelta{}
				err := encoder.MsgPackUnpackStruct(payload.Data, &stateDelta)
				if err == nil {
					valid := true
					hash, _ := stateDelta.GetHash()
					for _, sig := range stateDelta.Signatures {
						if v, _ := crypto.VerifySignatureSECP(sig.PublicKey.GetBytes(), hash, sig.Signature.GetBytes()); !v {
							valid = false
						}
					}
					if valid {
						delta := map[string]interface{}{}
						encoder.MsgPackUnpackStruct([]byte(stateDelta.Delta), delta)
						id := fmt.Sprint(delta["id"])
						model := entities.GetModelTypeFromEventType(stateDelta.Event.EventType)
						var newState = entities.GetStateModelFromEntityType(model)
						//var prevEventPath *entities.EventPath
						if stateDelta.PreviousHash != nil {
							// prevEvent := &entities.Event{}
							newState, _, err = service.SyncStateFromPeer(fmt.Sprint(delta["id"]), model, cfg, string(stateDelta.Event.Validator))
							if err != nil {
								logger.Errorf("ErrorSyncingOldState")
								
							}
							// if prevEvent != nil {
							// 	prevEventPath = prevEvent.GetPath()
							// }
						}
						if err != nil {
							utils.MapToStruct(delta, &newState)
							dataState := dsquery.NewDataStates(stateDelta.Event.ID, cfg)
							dataState.AddCurrentState(model, id, newState)
							// dataState.AddEvent(stateDelta.Event)
							err = dataState.Commit(nil, nil, nil, stateDelta.Event.ID, nil)
						}
					}
				}
				if err != nil {
					response.Error = err.Error()
					response.ResponseCode = 500
				}

			default:
				response, err = p2p.ProcessP2pPayload(cfg, payload, false)

			}

			if err != nil {
				logger.Error("HandleQuicConnection/processP2pPayload: ", err)
				return
			}
			_, err = stream.Write(response.MsgPack())
			if err != nil {
				log.Fatalf("Failed to send file: %v", err)
			}
		}(newStream)
	}

}
