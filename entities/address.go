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
type HexString string
type PublicKeyString HexString
type AddressString string
type DeviceString AddressString
type AccountString AddressString

func (a *HexString) GetBytes()  ([]byte) {
 	b, _ := hex.DecodeString(string(*a))
	return b
}

func (a *PublicKeyString) GetBytes()  ([]byte) {
	b, _ := hex.DecodeString(string(*a))
   return b
}

type Account struct {
	Address
}
type Device struct {
	Address
}

func (a AddressString) EncodeBytes() []byte  {
	return []byte(strings.ToLower(string(a)))
}
func (a Address) EncodeBytes() []byte  {
	return a.ToString().EncodeBytes()
}
func (a AddressString) IsAccount() bool {
	return strings.HasPrefix(string(a), "mid:") && len(string(a)) > 10
}

func (a Address) IsAccount() bool {
	return strings.EqualFold(a.Prefix, "mid") && len(string(a.Addr)) > 10
}

func (a Address) IsDevice() bool {
	return strings.EqualFold(a.Prefix, "did") && len(string(a.Addr)) > 10
}


func (a AddressString) IsDevice() bool {
	return strings.HasPrefix(string(a), "did:") && len(string(a)) > 10
}

func (a AccountString) IsValid() bool {
	return strings.HasPrefix(string(a), "mid:") && len(string(a)) > 10
}

func (a DeviceString) IsValid() bool {
	return strings.HasPrefix(string(a), "did:") && len(string(a)) > 10
}

func (a Account) ToString() AccountString {
	if a.Address.Prefix != "mid" {
		a.Address.Prefix = "mid"
	}
	return AccountString(a.Address.ToString())
}

func (a Device) ToString() DeviceString {
	if a.Address.Prefix != "did" {
		a.Address.Prefix = "did"
	}
	return DeviceString(a.Address.ToString())
}

func (a Account) IsValid() bool {
	return a.ToString().IsValid()
}

func (a Device) IsValid() bool {
	return a.ToString().IsValid()
}

func (s PublicKeyString) Bytes() []byte {
	if strings.HasPrefix(string(s), "0x") {
		s = PublicKeyString(strings.Replace(string(s), "0x", "", 1))
	}
	b, _ := hex.DecodeString(string(s))
	return b
}
func (address *AddressString) ToString() string {

	return string(*address)
}

type Address struct {
	Prefix string  `json:"pre"`
	Addr   string `json:"addr"`
	// Platform string    `json:"p"`
	Chain string `json:"ch"`
}


func (address *Address) ToJSON() []byte {
	m, e := json.Marshal(address)
	if e != nil {
		logger.Errorf("Unable to parse address to []byte")
	}
	return m
}
func (address *Address) ToAccount() Account {
	address.Prefix = "mid"
	return Account{Address:  *address}
}

func (address *Address) ToDevice() Device {
	address.Prefix = "did"
	return Device{Address:  *address}
}

func (address *Address) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(address)
	return b
}

func (address Address) ToDeviceString() DeviceString {
	if address.Prefix == "" {
		address.Prefix = "did"
	}
	return DeviceString(address.ToString())
}

func (address Address) ToAddressString() AddressString {
	if address.Prefix == "" {
		address.Prefix = "did"
	}
	return AddressString(address.ToString())
}


// func StringToAddressString(str string) (AddressString, error) {
// 	return AddressFromString(str).ToDeviceString()
// }

func AddressFromBytes(b []byte) (Address, error) {
	var address Address
	err := json.Unmarshal(b, &address)
	return address, err
}
func MsgUnpack(b []byte) (Address, error) {
	var address Address
	err := encoder.MsgPackUnpackStruct(b, &address)
	return address, err
}

func (address *Address) GetHash() []byte {
	return crypto.Keccak256Hash(address.ToBytes())
}

func (address Address) ToString() AddressString {
	values := []string{}
	values = append(values, address.Prefix)
	values = append(values, ":")
	values = append(values, address.Addr)
	if address.Chain != "" {
		values = append(values, fmt.Sprintf("#%s", address.Chain))
	}
	return AddressString(strings.Join(values, ""))
}

func (address *Address) ToBytes() []byte {
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

func AddressFromString(s string) (*Address, error) {
	return buildAddress(s, "")
	//return Address{Addr: values[0], Prefix: "", Chain: uint64(chain)}, nil
}

func AccountFromString(s string) (*Account,  error) {
	add, err := buildAddress(s, "mid")
	if err != nil {
		return nil, err
	}
	acc := add.ToAccount()
	return &acc, nil
	//return Address{Addr: values[0], Prefix: "", Chain: uint64(chain)}, nil
}

func DeviceFromString(s string) (*Device, error) {
	add, err := buildAddress(s, "did")
	if err != nil {
		return nil, err
	}
	acc := add.ToDevice()
	return &acc, nil
}
func buildAddress(s string, prefix string) (*Address, error) {
	addr := Address{Prefix: prefix}
	values := strings.Split(strings.Trim(string(s), " "), ":")
	if len(values) == 0 {
		return nil, fmt.Errorf("empty string")
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
	if len(prefix) > 0 && !strings.EqualFold(addr.Prefix, prefix) {
		return nil, fmt.Errorf("invalid address prefix")
	}
	return &addr, nil
}