package p2p

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/quic-go/quic-go"
)

var ValidCerts  = map[string]string{}

func  SendSecureQuicRequest(config *configs.MainConfiguration, maddr multiaddr.Multiaddr, validSigner entities.PublicKeyString,  message []byte) ([]byte, error) {
	ip, _, err := extractQuicAddress(config, []multiaddr.Multiaddr{maddr})
	if err != nil {
		return nil, err
	}
	addr := ip
	
	if ValidCerts[ip] == "" || ValidCerts[fmt.Sprintf("%s/addr", ip)] == "" {
		// get the cert
		certPayload := NewP2pPayload(config, P2pActionGetCert, []byte{'0'})
		err := certPayload.Sign(config.PrivateKeyEDD)
		if err != nil {
			return nil, err
		}
			
		certResponse, err := certPayload.SendP2pRequestToAddress(config.PrivateKeyEDD, maddr, DataRequest)
		if err != nil {
			return nil, err
		}
		logger.Infof("ValidSignerBefore %s", validSigner)
		validSigner, _ = chain.NetworkInfo.GetValidatorKeys(validSigner)
		if len(validSigner) == 0 {
			return nil, fmt.Errorf("invalid signer public key")
		}
		logger.Infof("ValidSignerAfter %s", validSigner)
		if certResponse.IsValid(config.ChainId) && strings.EqualFold(hex.EncodeToString(certResponse.Signer), string(validSigner)) {
			crd := CertResponseData{}
			err := encoder.MsgPackUnpackStruct(certResponse.Data,&crd)
			if err != nil {
				return nil, err
			}
			addr = fmt.Sprintf("%s%s", addr[:(strings.Index(addr,":"))], crd.QuicHost[strings.Index(crd.QuicHost,":"):])
			ValidCerts[addr] = hex.EncodeToString(crd.CertHash)
			ValidCerts[ip] = hex.EncodeToString(crd.CertHash)
			ValidCerts[fmt.Sprintf("%s/addr", ip)] = addr
			
		} else {
			return nil, fmt.Errorf("quic: invalid signer")
		}
	} else {
		addr = ValidCerts[fmt.Sprintf("%s/addr", ip)]
	}
	
	b, err := sendQuicRequest(addr, message, false)
	
	if err == ErrInvalidCert {
		ValidCerts[addr] = ""
	}
	return b, err
}

func  SendInsecureQuicRequest(addr string, message []byte) ([]byte, error) {
	return sendQuicRequest(addr, message, true)
}
var ErrInvalidCert = fmt.Errorf("invalid certficate")
func  sendQuicRequest(addr string, message []byte, insecure bool, ) ([]byte, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"mlayer-p2p"},
		VerifyPeerCertificate: func (rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error  {
			if insecure {
				return nil
			}
			for _, cert := range rawCerts {
				if(ValidCerts[addr] != hex.EncodeToString(crypto.Keccak256Hash(cert))) {
					return ErrInvalidCert
				}
			}
			return nil
		},
	}
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		logger.Debugf("sendQuicRequest/DialAddr: %s, %v", addr, err)
		return nil, err
	}
	defer conn.CloseWithError(0, "")

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		logger.Debugf("sendQuicRequest.OpenStream: %v", err)
		return nil, err
	}
	defer stream.Close()
	_, err = stream.Write([]byte(message))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	data := bytes.Buffer{}
	for {
		n, err := stream.Read(buf)  // Read into the buffer
		
		data.Write(buf[:n])
		if n == 0 || err == io.EOF {
			break  // End of file, stop reading
		}
		if err != nil  {
			return nil, err
		}
	}
	return data.Bytes(), nil
}