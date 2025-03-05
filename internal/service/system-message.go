package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/system"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
)

func HandleNewNodeSystemMessageEvent(event *entities.Event, ctx *context.Context) (err error) {
	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		panic("Unable to load config from context")
	}
	
	if !event.IsLocal(cfg) {
		err = ValidateEvent(*event)
		if err != nil {
			return err
		}
		message := event.Payload.Data.(entities.SystemMessage)
		
		switch message.Type {
		case entities.AnnounceTopicInterest:
			data := entities.NodeInterest{}
			err = encoder.MsgPackUnpackStruct(message.Data, &data)
			if err != nil {
				logger.Errorf("HandleNewNodeSystemMessageEventErrorEncoder: %v", err)
			}
			ownInterest := []string{}
			err := system.RegisterNodeInterest(event.Validator, data.Ids, data.Type, &ownInterest)
			if err != nil {
				return err
			}
			if len(ownInterest) > 0 {
				payload := p2p.NewP2pPayload(cfg, p2p.P2pActioNotifyTopicInterest, (&entities.NodeInterest{
					Ids: ownInterest,
					Type: data.Type,
					Expiry: int(time.Now().Add(constants.TOPIC_INTEREST_TTL).UnixMilli()),
				}).MsgPack())
				// payload.Sign(cfg.PrivateKeyEDD)
				_, err := payload.SendDataRequest(string(event.Validator))
				if err != nil {
					logger.Errorf("HandleNewNodeSystemMessageEventError: %v", err)
				}


			}

			return err
		case entities.AnnounceSelf:
			mad := p2p.NodeMultiAddressData{}
		
			err = encoder.MsgPackUnpackStruct(message.Data, &mad)
			if err != nil {
				logger.Errorf("HandleAnnounceSelfErrorEncoder: %v", err)
			}
			 logger.Infof("NewValidBroadcastedAddressData: %v", mad.Signer)
			if mad.IsValid(cfg.ChainId) {
				// err := stores.SystemStore.Set(context.Background(), datastore.NewKey(fmt.Sprintf("/mad/%s", event.Validator)), message.Data, true)
				// if err != nil {
				// 	logger.Errorf("MadStoreError: %s ::: %v", fmt.Sprintf("/mad/%s", event.Validator),  err)
				// }
				// (&p2p.ValidMads).Update(&mad)
				mad.Sync()
				
				// remoteAddress := fmt.Sprintf("%s:%d", mad.Hostname, mad.QuicPort)
				// if mad.Hostname == "" {
				// 	remoteAddress = fmt.Sprintf("%s:%d", mad.IP, mad.QuicPort)
				// } 
				// p2p.ValidCerts[remoteAddress] = ""
				// p2p.ValidCerts[fmt.Sprintf("%s/addr", event.Validator)] = ""
				// p2p.ValidCerts[hex.EncodeToString(mad.CertHash)] = ""
			} else {
				return fmt.Errorf("invalid mad")
			}
			return err
		}

	}

	return fmt.Errorf("HandleNewNodeSystemMessageEvent: %v", "Invalid system message type")

}
