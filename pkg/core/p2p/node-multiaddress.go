package p2p

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/sirupsen/logrus"
)

/*
*
NODE HANDSHAKE MESSAGE
*
*/


type NodeMultiAddressData struct {
    Addresses []string `json:"addr"`
	Hostname string  `json:"host"`
	QuicPort uint16  `json:"quicPort"`
	CertHash json.RawMessage  `json:"certHash"`
	IP string  `json:"ip"`
	Timestamp uint64 `json:"ts"`
	ChainId configs.ChainId `json:"pre"`
	Signer json.RawMessage `json:"signr"`
	PubKeyEDD json.RawMessage `json:"pubKey"`
	Signature json.RawMessage `json:"sig"`
	config *configs.MainConfiguration `json:"-" msgpack:"-"`
	
}

func (hs *NodeMultiAddressData) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(hs)
	return b
}

func (mad *NodeMultiAddressData) QuicAddress() string {
	remoteAddress := fmt.Sprintf("%s:%d", mad.Hostname, mad.QuicPort)
		if mad.Hostname == "" {
			remoteAddress = fmt.Sprintf("%s:%d", mad.IP, mad.QuicPort)
		} 
		return remoteAddress
}


func (mad *NodeMultiAddressData) Sync() error {
	if mad.IsValid(cfg.ChainId) {
		ctx := context.Background()
		batch, err := stores.SystemStore.Batch(ctx)
		if err != nil {
			logger.Errorf("SystemBatchReader: %v", err)
			return err
		}
		keys := []string{
			hex.EncodeToString(mad.Signer),
			hex.EncodeToString(mad.PubKeyEDD),
			hex.EncodeToString(mad.CertHash),
			mad.IP,
		}
		if mad.Hostname != "" {
			keys = append(keys, mad.Hostname)
		}
		for _, key := range keys {
			
			err = batch.Put(ctx, datastore.NewKey(fmt.Sprintf("/mad/%s", key)), mad.MsgPack())
			if err != nil {
				logger.Errorf("MadStoreError: %s ::: %v", key,  err)
			}
		}
		err = batch.Commit(ctx)
		if err != nil {
			logger.Errorf("MadStoreCommitError: %v",  err)
		}
		(&ValidMads).Update(mad)
		return nil
	} else {
		return fmt.Errorf("invalid mad")
	}
}
func (mad *NodeMultiAddressData) Delete() error {
	if mad.IsValid(cfg.ChainId) {
		ctx := context.Background()
		batch, err := stores.SystemStore.Batch(ctx)
		if err != nil {
			logger.Errorf("MadStoreDeleteError: %v",  err)
		}
		keys := []string{
			hex.EncodeToString(mad.Signer),
			hex.EncodeToString(mad.PubKeyEDD),
			hex.EncodeToString(mad.CertHash),
			mad.IP,
		}
		if mad.Hostname != "" {
			keys = append(keys, mad.Hostname)
		}
		for _, key := range keys {
			err = batch.Delete(ctx, datastore.NewKey(fmt.Sprintf("/mad/%s", key)))
			if err != nil {
				logger.Errorf("MadStoreError: %s ::: %v", key,  err)
			}
		}
		err = batch.Commit(ctx)
		if err != nil {
			logger.Errorf("MadStoreCommitError: %v",  err)
		}
		if err != nil {
			return nil
		}
		(&ValidMads).Delete(mad)
		return nil
	} else {
		return fmt.Errorf("invalid mad")
	}
}
func (n NodeMultiAddressData) EncodeBytes() ([]byte, error) {
	data := []byte{}
	for _, val := range n.Addresses {
		b, _ := encoder.EncodeBytes(encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: val})
		data = append(data, b...)
	}
    return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: data},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: n.CertHash},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: n.Hostname},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: n.QuicPort},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: n.IP},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: n.ChainId.Bytes()},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: n.PubKeyEDD},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: n.Timestamp},
		
	)
}


func UnpackNodeMultiAddressData(b []byte) ( NodeMultiAddressData, error) {
	var message  NodeMultiAddressData
	err := encoder.MsgPackUnpackStruct(b, &message)
	return message, err
}


func (nma * NodeMultiAddressData) IsValid(prefix configs.ChainId) bool {
	// Important security update. Do not remove. 
	// Prevents cross chain replay attack
	nma.ChainId = prefix  // Important security update. Do not remove
	//
	// if math.Abs(float64(uint64(time.Now().UnixMilli()) - nma.Timestamp)) > float64(4 * time.Hour.Milliseconds()) {
	// 	logger.WithFields(logrus.Fields{"data": nma}).Warnf("MultiaddressDataExpired: %d", uint64(time.Now().UnixMilli()) - nma.Timestamp)
	// 	return false
	// }
	// signer, err := hex.DecodeString(string(nma.Signer));
	// if err != nil {
	// 	logger.Error("Unable to decode signer")
	// 	return false
	// }
	
	data, err := nma.EncodeBytes()
	if err != nil {
		logger.Error("Unable to encode message", err)
		return false
	}
	// signature, err := hex.DecodeString(nma.Signature);
	// if err != nil {
	// 	logger.Error(err)
	// 	return false
	// }
	// logger.Debugf("Operator4 %s", nma.Signer)
	
	isValid, err := crypto.VerifySignatureSECP(nma.Signer, data, nma.Signature)
	if err != nil {
		logger.Errorf("NodeMultiAddressData: %v", err)
		return false
	}
	
	if !isValid {
		logger.WithFields(logrus.Fields{"addresses": nma.Addresses, "signature": nma.Signature}).Warnf("Invalid signer %s", nma.Signer)
		return false
	}
	return true
}


func NewNodeMultiAddressData(config *configs.MainConfiguration, privateKeySECP []byte, addresses []string, pubKeyEDD []byte) (*NodeMultiAddressData, error) {
	//pubKey := crypto.GetPublicKeySECP(privateKey)
	certData := crypto.GetOrGenerateCert(cfg.Context)
	keyByte, _ := hex.DecodeString(certData.Key)
		certByte, _ := hex.DecodeString(certData.Cert)
		tlsConfig, _ := crypto.GenerateTLSConfig(keyByte, certByte)
		certHash :=  crypto.Keccak256Hash(tlsConfig.Certificates[0].Certificate[0])
	nma := NodeMultiAddressData{config: config, PubKeyEDD: pubKeyEDD,
		IP: cfg.IP,
		QuicPort: cfg.QuicPort,
		Hostname: cfg.Hostname,
		CertHash: certHash,
		 ChainId: config.ChainId, Addresses: addresses,   Timestamp: uint64(time.Now().UnixMilli())}
	b, err := nma.EncodeBytes();
	if(err != nil) {
		return nil, err
	}
	signature, _ := crypto.SignSECP(b, privateKeySECP)
    nma.Signature = signature
    nma.Signer = config.PublicKeySECP
	return &nma, nil
}

