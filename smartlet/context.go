package smartlet

import (
	"context"
	"encoding/json"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/xxh3"
)

// type ContextFlag string

// const (
// 	Application ContextFlag = "app"
// )



type ActionParams struct {
	Action uint8           `json:"a"`
	Params   json.RawMessage `json:"p"`
}

func DecodeActionParams(b []byte, ap *ActionParams) error {
	return encoder.MsgPackUnpackStruct(b, ap)
}


type Key [16]byte
func NewKey(data []byte)(k [16]byte) {
	return Key(encoder.Uint64ToBytes16(xxh3.Hash(data)))
}
type KeystoreInterface interface {
	Get(keys [][16]byte) (result map[(Key)][]byte, err error)
	List(key [16]byte, prefix []byte) (result [][]byte, err error)
	Set(key Key, value []byte, replaceIfExists bool) error
	Append(key Key, value [16]byte, prefix []byte) error
	GetUpdateKeys() ([]Key)
	IsFinalized() bool
}


type Result map[Key][]byte

type SmartletInterface interface {
	PreCommit(app App) (App, error)
	PostCommit(app App) (App, error)
}

// type PreCommitApp interface {
// 	KeyStoreReader() readOnlyKeystore
	
// }
// type PostCommitApp interface {
// 	KeyStoreReader() readOnlyKeystore
// 	KeyStoreWriter() readWriteKeystore

// }


type UpdateData struct {
	Data    map[Key]interface{}
	Replace bool
	Prefix  []byte
}

type App struct {
	Event          *entities.Event
	Topic          *entities.Topic
	KeyStore  KeystoreInterface

	
	logger         *logrus.Logger
	state          *map[uint8][]byte
	ctx *context.Context
	result *Result
	
}


func NewApplication(event *entities.Event, topic *entities.Topic, logger *logrus.Logger, KeyStore KeystoreInterface, ctx *context.Context) *App {
	// keystore := KeyStore{db: db, updateData: []UpdateData{}, logger: logger}
	return &App{ event, topic, KeyStore, logger, &map[uint8][]byte{}, ctx, &Result{}}
	// app.KeyStore.SetApp(&app)
	
}

func (app *App) Finalized() bool {
	return app.KeyStore.IsFinalized()
}


// func (a App) KeyStoreReader() readOnlyKeystore {
// 	return a.KeyStore
// }

// func (a App) KeyStoreWriter() readWriteKeystore {
// 	return a.KeyStore
// }


// ephemeral state of an app
func (app App) SetState(key uint8, value []byte) {
	(*app.state)[key] = value
}

func (app App) GetState(key uint8) ([]byte, bool) {
	value, exists := (*app.state)[key]
	return value, exists
}

func (app App) GetStateKeys() (keys []uint8) {
	for k, _ := range *app.state {
		keys = append(keys, k)
	}
	return keys
}

func (app App) SetStore(key Key, value []byte, replaceIfExists bool) error{
	return app.KeyStore.Set(key, value, replaceIfExists)
}

func (app App) GetStore(keys [][16]byte) (result map[(Key)][]byte, err error) {
	return app.KeyStore.Get(keys)
}

func (app App) ListStore(key [16]byte, prefix []byte) (result [][]byte, err error) {
	return app.KeyStore.List(key, prefix)
}

func (app App) AppendStore(key Key, value [16]byte, prefix []byte) error {
	return app.KeyStore.Append(key, value, prefix)
}


func (app App) SetResult(key Key, value []byte)  {
	 (*app.result)[key] =  value
}

func (app App) GetResult() *Result  {
	return app.result
}

