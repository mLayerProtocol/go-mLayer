package p2p

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/multiformats/go-multiaddr"
)

// var ValidCerts  = map[string]string{}

type NodeMADs map[string]*NodeMultiAddressData
var ValidMads  = NodeMADs{}

func (v *NodeMADs) Update(mad *NodeMultiAddressData) {
	(*v)[hex.EncodeToString(mad.Signer)] = mad
	(*v)[hex.EncodeToString(mad.PubKeyEDD)] = mad
	(*v)[hex.EncodeToString(mad.CertHash)] = mad
	(*v)[fmt.Sprint(mad.IP)] = mad
	if len(mad.Hostname) > 0 {
		(*v)[mad.Hostname] = mad
	}
}

func (v *NodeMADs) Delete(mad *NodeMultiAddressData) {
	logger.Debugf("DeleteMDA %v", mad.config.PublicKeySECP)
	(*v)[hex.EncodeToString(mad.Signer)] = nil
	(*v)[hex.EncodeToString(mad.PubKeyEDD)] = nil
	(*v)[hex.EncodeToString(mad.CertHash)] = nil
	(*v)[fmt.Sprint(mad.IP)] = nil
	(*v)[mad.Hostname] = nil
	
}

func (v NodeMADs) Get(key string, sync bool) *NodeMultiAddressData {
	nma := v[key]
	if nma == nil {
		b, err := stores.SystemStore.Get(context.Background(), datastore.NewKey(fmt.Sprintf("/mad/%s", key)))
		if err == datastore.ErrNotFound {
			logger.Infof("UnableToFindDtata for Key %s", key)
			return nil
		}
		nma2 := &NodeMultiAddressData{}
		err = encoder.MsgPackUnpackStruct(b, nma2)
		logger.Infof("UnableToFindDtata2 for Key %s", key)
		if err == nil {
			nma = nma2
			if sync {
				(&ValidMads).Update(nma)
			}
		}
	}
	return nma
}

func  SendSecureQuicRequest(config *configs.MainConfiguration, maddr multiaddr.Multiaddr,  validator string, message []byte) ([]byte, error) {
	ip, _, err := parseQuicAddress(config, []multiaddr.Multiaddr{maddr})
	if err != nil {
		return nil, err
	}
	//addr := ip
	mad := &NodeMultiAddressData{}
	
	if ValidMads.Get(ip, true) == nil {
		// get the cert
		certPayload := NewP2pPayload(config, P2pActionGetCert, []byte{'0'})
		err := certPayload.Sign(config.PrivateKeyEDD)
		if err != nil {
			return nil, err
		}
		logger.Infof("SendingP2PRequestTo %s", maddr)
		certResponse, err := certPayload.SendP2pRequestToAddress(config.PrivateKeyEDD, maddr, DataRequest)
		if err != nil {
			return nil, err
		}
		isValidator, _ := chain.NetworkInfo.IsValidator( hex.EncodeToString(certResponse.Signer))
		if certResponse.IsValid(config.ChainId) && isValidator {
			 err := encoder.MsgPackUnpackStruct(certResponse.Data, mad)
			 if err != nil {
				return nil, err
			 }
			 err = mad.Sync()
			 if err != nil {
				return nil, err
			 }
		} else {
			return nil, fmt.Errorf("quic: invalid signer")
		}
	} else {
		mad = ValidMads[fmt.Sprintf("%s/addr", ip)]
	}
	logger.Infof("SendingQuicRequestTo %s", mad.QuicAddress())
	b, err := sendQuicRequest(mad.QuicAddress(), validator, message, false)
	
	if err == ErrInvalidCert {
		ValidMads.Delete(mad)
	}
	return b, err
}

func  SendSecureQuicRequestToValidator(config *configs.MainConfiguration, publicKeySecP string, payload *P2pPayload) ([]byte, error) {

	// remoteAddress := ""
	// certHash := ""
	var mad *NodeMultiAddressData 
	if madData := ValidMads.Get(publicKeySecP, true); madData == nil {
		logger.Infof("No valid certificate found for %s, %d", publicKeySecP, len(ValidMads))
		madFromDht , err := GetNodeMultiAddressData(config.Context, publicKeySecP)
		if err != nil {
			// GET IT FROM P2P
			// certPayload := NewP2pPayload(config, P2pActionGetCert, []byte{'0'})
			// 	err := certPayload.Sign(config.PrivateKeyEDD)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	certResponse, err := certPayload.SendP2pRequestToAddress(config.PrivateKeyEDD, maddr, DataRequest)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	isValidator, _ := chain.NetworkInfo.IsValidator( hex.EncodeToString(certResponse.Signer))
			// 	if certResponse.IsValid(config.ChainId) && isValidator {
			// 	err := encoder.MsgPackUnpackStruct(certResponse.Data, mad)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// } else {
			// 	return nil, fmt.Errorf("quic: invalid signer")
			// }
			return nil, fmt.Errorf("quic: unable to get address for node from dht")
		} else {
			mad = madFromDht
		}
		mad.Sync()
		
	} else {
		mad = madData
	
	}
	logger.Infof("SendingQuicRequestToValidator %s at address %s", publicKeySecP, mad.QuicAddress())
	b, err := sendQuicRequest(mad.QuicAddress(), publicKeySecP, payload.MsgPack(), false)
	// logger.Infof("SentQuickRequestToValidator %s, %d", publicKeySecP, payload.Id )
	if err != nil {
		logger.Errorf("SendingQuicRequestToValidatorError: %v", err)
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("empty response from remote peer")
	}
	if err == ErrInvalidCert {
		logger.Errorf("INVALIDcERT: %v", err)
		mad.Delete()
	}
	return b, err
}

func  SendInsecureQuicRequest(config *configs.MainConfiguration, maddr multiaddr.Multiaddr, validator string,  message []byte) ([]byte, error) {
	ip, _, err := parseQuicAddress(config, []multiaddr.Multiaddr{maddr})
	if err != nil {
		return nil, err
	}
	addr := ip
	return sendQuicRequest(addr, validator, message, true)
	
}
var ErrInvalidCert = fmt.Errorf("invalid certficate")

func  sendQuicRequest(addr string, validator string, message []byte, insecure bool, ) ([]byte, error) {
	logger.Infof("SendingQuicRequest.... %d", len(message))
	conn, err  := NodeQuicPool.GetConnection(context.Background(), addr, validator, false)
	
	if err != nil {
		logger.Debugf("sendQuicRequest/ConnectToAddr: %s, %v", addr, err)
		return nil, err
	}
	logger.Debugf("OpeningStream: %v", err)
	stream, err := conn.OpenStreamSync(context.Background())
	
	if err != nil {
		logger.Debugf("sendQuicRequest.OpenStream: %v", err)
		conn, err  = NodeQuicPool.GetConnection(context.Background(), addr, validator, true)
		if err != nil {
			logger.Debugf("sendQuicRequest/ConnectToAddr: %s, %v", addr, err)
			return nil, err
		}
		stream, err = conn.OpenStreamSync(context.Background())
		// try again
		if err != nil {
			return nil, err
		}
		
	} 
	logger.Debugf("StreamOpened: %v", err)
	defer stream.Close()
	_, err = stream.Write([]byte(message))
	if err != nil {
		return nil, err
	}
	
	logger.Debugf("Writen to stream: %v", err)
	buf := make([]byte, 1024)
	data := bytes.Buffer{}
	
	// var wg sync.WaitGroup
	defer func() {
		logger.Debugf("FinishedRequest: %d", data.Len())
	}()
	// defer wg.Wait()

	resultChan := make(chan error, 1)

    // Launch reading in a goroutine
    go func() {
		for {
			n, err := stream.Read(buf)
			logger.Infof("Read %d", n)
			data.Write(buf[:n])
			if n == 0 || err == io.EOF {
				break  // End of file, stop reading
			}
			if err != nil  {
				resultChan <- err
			}
			
		}
		resultChan <- nil
    }()
	timeout := 2000 * time.Millisecond
    // Wait for either the read to complete or timeout
    select {
    case result := <-resultChan:
        return  data.Bytes(), result
    case <-time.After(timeout):
        return nil, fmt.Errorf("read timed out after %v", timeout)
    }

	// ctx, cancel := context.WithTimeout(context.Background(), 2000 * time.Millisecond)
	// defer cancel()

	// wg.Add(1)
	// go func() error {
	// 	defer wg.Done()
	// 	for {
	// 		n, err := stream.Read(buf)  // Read into the buffer
				
	// 			data.Write(buf[:n])
	// 			if n == 0 || err == io.EOF {
	// 				break  // End of file, stop reading
	// 			}
	// 			if err != nil  {
	// 				return  err
	// 			}
	// 		}
	// 		return nil
	// }()
	
	// for  range ctx.Done() {
	// 		wg.Done()	
	// 		logger.Infof("ReadTIMEOUT")
	// 		return nil, fmt.Errorf("read timeout")	
	// }
	
	// logger.Debugf("Read from stream: %d", data.Len())
	// return data.Bytes(), nil
}