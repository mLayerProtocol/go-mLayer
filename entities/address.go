package entities

import (
	// "errors"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
)

type PublicKeyString string
type DIDString string
type DeviceString DIDString

func (s PublicKeyString) Bytes() []byte {
	if strings.HasPrefix(string(s), "0x") {
		s = PublicKeyString(strings.Replace(string(s), "0x", "", 1))
	}
	b, _ := hex.DecodeString(string(s))
	return b
}
func (address *DIDString) ToString() string {

	return string(*address)
}

type DID struct {
	Prefix string  `json:"pre"`
	Addr   string `json:"addr"`
	// Platform string    `json:"p"`
	Chain string `json:"ch"`
}

func (address *DID) ToJSON() []byte {
	m, e := json.Marshal(address)
	if e != nil {
		logger.Errorf("Unable to parse address to []byte")
	}
	return m
}

func (address *DID) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(address)
	return b
}

func (address DID) ToDeviceString() DeviceString {
	if address.Prefix == "" {
		address.Prefix = "did"
	}
	return DeviceString(address.ToString())
}

func (address DID) ToDIDString() DIDString {
	if address.Prefix == "" {
		address.Prefix = "did"
	}
	return DIDString(address.ToString())
}


func StringToDeviceString(str string) (DeviceString) {
	return AddressFromString(str).ToDeviceString()
}

func AddressFromBytes(b []byte) (DID, error) {
	var address DID
	err := json.Unmarshal(b, &address)
	return address, err
}
func MsgUnpack(b []byte) (DID, error) {
	var address DID
	err := encoder.MsgPackUnpackStruct(b, &address)
	return address, err
}

func (address *DID) GetHash() []byte {
	return crypto.Keccak256Hash(address.ToBytes())
}

func (address DID) ToString() DIDString {
	values := []string{}
	values = append(values, address.Prefix)
	values = append(values, ":")
	values = append(values, address.Addr)
	if address.Chain != "" {
		values = append(values, fmt.Sprintf("#%s", address.Chain))
	}
	return DIDString(strings.Join(values, ""))
}

func (address *DID) ToBytes() []byte {
	// var buffer bytes.Buffer
	// // buffer.Write([]byte(address.Platform))
	// buffer.Write([]byte("did:"))
	// if strings.HasPrefix(address.Addr, "0x") {
	// 	// treat as hex
	// 	b, err := hexutil.Decode(faddress.Addr)
	// 	if err != nil {
	// 		return []byte(""), err
	// 	}
	// 	buffer.Write(b)
	// } else {
	// buffer.Write([]byte(address.Addr))
	// if(address.Chain != "") {
	// 	// binary.Write(&buffer, binary.BigEndian, address.Chain)
	// 	buffer.Write([]byte(fmt.Sprintf("#%s", address.Chain)))
	// }
	return []byte(address.ToString())
}

func DIDFromString(s string) (DID) {
	return AddressFromString(s)
	//return Address{Addr: values[0], Prefix: "", Chain: uint64(chain)}, nil
}

func AccountFromString(s string) (DID) {
	return buildAddress(s, "mid")
	//return Address{Addr: values[0], Prefix: "", Chain: uint64(chain)}, nil
}

func AddressFromString(s string) (DID) {
	return buildAddress(s, "did")
	//return Address{Addr: values[0], Prefix: "", Chain: uint64(chain)}, nil
}
func buildAddress(s string, prefix string) DID {
	addr := DID{Prefix: prefix}
	values := strings.Split(strings.Trim(string(s), " "), ":")
	if len(values) == 0 {
		return DID{}
	}

	if len(values) == 1 {
		addr.Addr = values[0]
	}
	if len(values) == 2 {
		addr.Addr = values[1]
		addr.Prefix = values[0]
	}
	values2 := strings.Split(addr.Addr, "#")
	if len(values2) > 1 {
		addr.Addr = values2[0]
		addr.Chain = values2[1]
	}
	return addr
}