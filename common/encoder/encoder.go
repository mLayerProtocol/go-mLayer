package encoder

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
	"github.com/vmihailenco/msgpack/v5"
)

var logger = &log.Logger

func MsgPackStruct(msg interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json")
	err := enc.Encode(msg)
	return buf.Bytes(), err
}

func MsgPackUnpackStruct[T interface{}](b []byte, message *T) error {
	buf := bytes.NewBuffer(b)
	dec := msgpack.NewDecoder(buf)
	// dec.UseLooseInterfaceDecoding(true)
	dec.SetCustomStructTag("json")
	err := dec.Decode(message)
	return err
}

// func EncodeNumber(b []byte, message interface{}) error {
// 	buf := bytes.NewBuffer(b)
// 	dec := msgpack.NewDecoder(buf)
// 	dec.SetCustomStructTag("json")
// 	err := dec.Decode(&message)
// 	return err
// }

func NumberToByte(i uint64) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, i)
	return buf.Bytes()
}

func NumberFromByte(buf []byte) uint64 {
	return (binary.BigEndian.Uint64(buf))
}

type EncoderDataType string

const (
	StringEncoderDataType  EncoderDataType = "string"
	ByteEncoderDataType    EncoderDataType = "byte"
	BigNumEncoderDataType  EncoderDataType = "bignum"
	IntEncoderDataType     EncoderDataType = "int"
	HexEncoderDataType     EncoderDataType = "hex"
	BoolEncoderDataType    EncoderDataType = "bool"
	AddressEncoderDataType EncoderDataType = "address"
)

type EncoderParam struct {
	Type  EncoderDataType
	Value any
}

func EncodeBytes(args ...EncoderParam) (data []byte, err error) {
	defer func() {
		// recover from panic if one occurred. Set err to nil otherwise.
		if pan := recover(); pan != nil {
			logger.Errorf("encodbytes: %v", fmt.Errorf("%v", pan))
		}
	}()
	m := make(map[int][]byte)
	var index []int
	var buffer bytes.Buffer
	for i, arg := range args {
		index = append(index, i)
		if arg.Type == ByteEncoderDataType {
			if _, ok := arg.Value.(json.RawMessage); ok {
				m[i] = []byte(arg.Value.(json.RawMessage))
			} else {
				m[i] = (arg.Value.([]byte))
		
			}
		}
		if arg.Type == StringEncoderDataType {
			m[i] = []byte(fmt.Sprintf("%v", arg.Value))
		}
		if arg.Type == IntEncoderDataType {
			num, err := strconv.ParseUint(fmt.Sprintf("%v", arg.Value), 10, 64)
			if err != nil {
				return []byte(""), err
			}
			m[i] = NumberToByte(num)
			// if num == 0 {
			// 	m[i] = []byte{}
			// } else {
			// 	m[i] = NumberToByte(num)
			// }
		}
		if arg.Type == BoolEncoderDataType {
			val := 0
			if arg.Value == true {
				val = 1
			}
			num, err := strconv.ParseUint(fmt.Sprintf("%v", val), 10, 64)
			if err != nil {
				return []byte(""), err
			}
			m[i] = []byte(NumberToByte(num))
		}
		if arg.Type == BigNumEncoderDataType {
			bigNum := new(big.Int)
			bigNum.SetString(fmt.Sprintf("%v", arg.Value), 10)
			m[i] = bigNum.Bytes()
		}
		if arg.Type == HexEncoderDataType {
			hexString := fmt.Sprintf("%v", arg.Value)
			if strings.HasPrefix(hexString, "0x") {
				b, err := hexutil.Decode(fmt.Sprintf("%v", arg.Value))
				if err != nil {
					return []byte(""), err
				}
				m[i] = b
			} else {
				b, err := hex.DecodeString(fmt.Sprintf("%v", arg.Value))
				if err != nil {
					return []byte(""), err
				}
				m[i] = b
			}			
		}
		if arg.Type == AddressEncoderDataType {
			v := fmt.Sprintf("%v", arg.Value)
			
			m[i] = []byte(v)
			// if err != nil {
			// 	return nil, err
			// }
			// if strings.HasPrefix(addr.Addr, "0x") {
			// 	// treat as hex
			// 	b, err := hexutil.Decode(fmt.Sprintf("%v", arg.Value))
			// 	if err != nil {
			// 		return []byte(""), err
			// 	}
			// 	m[i] = b
			// } else {
			// 	toLower := strings.ToLower(fmt.Sprintf("%v", arg.Value))
			// 	values := strings.Split(strings.Trim(fmt.Sprintf("%v", toLower), " "), ":")
			// 	var addrBuffer bytes.Buffer
			// 	addrBuffer.Write([]byte(values[0]))
			// 	addrBuffer.Write([]byte(values[1]))
			// 	if len(values) == 3 {
			// 		chain, err := strconv.ParseUint(values[2], 10, 64)
			// 		if err != nil {
			// 			return []byte(""), err
			// 		}
			// 		addrBuffer.Write(NumberToByte(chain))
			// 	}
			// 	m[i] = addrBuffer.Bytes()
			// }
		}

	}

	// TODO: sort the byte slice in an efficient way so that the order of args doesnt matter
	for _, n := range index {
		buffer.Write(m[n])
	}
	// logger.Debugf("LOG MEssage  =========> %v \n %v \n %v", index, buffer, args)
	return buffer.Bytes(), nil
}

func AddBase64Padding(value string) string {
    m := len(value) % 4
    if m != 0 {
        value += strings.Repeat("=", 4-m)
    }
    return value
}

func ToBase64Padded(data []byte) (string) {
	rsl := base64.StdEncoding.EncodeToString(data)
	return AddBase64Padding(rsl)
}

func ExtractHRP(address string) (string, error) {
    // The separator for Bech32 is "1", so we split the string based on that.
    parts := strings.SplitN(address, "1", 2)
    if len(parts) < 2 {
        return "", fmt.Errorf("invalid Bech32 address: %s", address)
    }
    return parts[0], nil
}