package entities

import (
	// "errors"

	"context"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)



type EntityModel string

const (
	AuthModel         EntityModel = "auth"
	TopicModel        EntityModel = "top"
	SubscriptionModel EntityModel = "sub"
	MessageModel      EntityModel = "msg"
	ApplicationModel       EntityModel = "app"
	TopicInterestModel       EntityModel = "topInt"
	WalletModel       EntityModel = "wal"
	SystemModel		EntityModel = "sys"
	
)

var EntityModels = []EntityModel{
	AuthModel, TopicModel, SubscriptionModel, MessageModel, ApplicationModel, WalletModel, SystemModel,
}


type QueryLimit struct {
	Limit  int `json:"lim"`
	Offset int 	`json:"off"`
}


type KeyByteValue struct {
	Key   string
	Value []byte
}
/*
*
Event paths define the unique path to an event and its relation to the entitie

*
*/
type EntityPath struct {
	Version float32 `json:"_v"`
	Model     EntityModel      `json:"mod"`
	ID      string          `json:"id"`
	Validator PublicKeyString `json:"val"`
	Index int64 `json:"idx"`
	NodeCount int32 `json:"n"`
	Timestamp      uint64          `json:"ts"`
}

type EventPath struct {
	EntityPath
	// AuthEvent *EntityPath
}

func (e *EntityPath) ToString() string {
	if e == nil || e.ID == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s/%d/%d/%d", e.Validator, e.Model, e.ID, e.Index, e.NodeCount, e.Timestamp)
}

func (e *EntityPath) EncodeBytes() ([]byte, error) {
	b, _ := encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(e.ID)},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value:e.Model},
		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value:e.Validator},
		// encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value:e.Index},
		// encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value:e.NodeCount},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value:e.Timestamp},
	)
	return b, nil
}
// func (e *EventPath) EncodeBytes() ([]byte, error) {
// 	bb, err := e.EntityPath.EncodeBytes()
// 	if err != nil {
// 		return nil, err
// 	}
// 	params := []encoder.EncoderParam{{Type: encoder.ByteEncoderDataType, Value: bb}}
// 	if e.AuthEvent != nil {
// 		a, _ := e.AuthEvent.EncodeBytes()
// 		params = append(params, encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: a})
// 	}
// 	b, _ := encoder.EncodeBytes(
// 		params...
// 	)
// 	return b, nil
// }
func (e *EntityPath) ToHexHash() (string) {
	b, err := e.EncodeBytes()
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func NewEntityPath(validator PublicKeyString, model EntityModel, id string) *EntityPath {
	return &EntityPath{Model: model, ID: id, Validator: validator}
}
func NewEventPath(validator PublicKeyString, model EntityModel, id string) *EventPath {
	return &EventPath{EntityPath{Model: model, ID: id, Validator: validator}}
}

func (e *EntityPath) MsgPack() ([]byte) {
	b, _ := encoder.MsgPackStruct(e)
	return b
}

func UnpackEntityPath(b []byte) (*EntityPath, error) {
	var p EntityPath
	err := encoder.MsgPackUnpackStruct(b, &p)
	return &p, err
}
func UnpackEventPath(b []byte) (*EventPath, error) {
	var p EventPath
	err := encoder.MsgPackUnpackStruct(b, &p.EntityPath)
	return &p, err
}

func EntityPathFromString(path string) *EntityPath {
	parts := strings.Split(path, "/")
	// assoc, err := strconv.Atoi(parts[0])
	// if err != nil {
	// 	return nil, err
	// }
	switch len(parts) {
	case 0:
		return &EntityPath{}
	case 1:
		return &EntityPath{
			//Relationship: EventAssoc(assoc),
			Model: EntityModel(""),
			ID:  parts[0],
		}
	case 2:
		return &EntityPath{
			//Relationship: EventAssoc(assoc),
			Model:     EntityModel(""),
			ID:      parts[1],
			Validator: PublicKeyString(parts[0]),
		}
	default:
		return &EntityPath{
			Validator: PublicKeyString(parts[0]),
			Model:     EntityModel(parts[1]),
			ID:      parts[2],
		}
	}
}
func EventPathFromString(path string) *EventPath {
	b := EntityPathFromString(path)
	return &EventPath{EntityPath: *b}
}

func (eP EntityPath) GormDataType() string {
	return "varchar"
}
func (eP EntityPath) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {

	asString := eP.ToString()
	return clause.Expr{
		SQL:                "?",
		Vars:               []interface{}{asString},
		WithoutParentheses: false,
	}
}

func (sD *EntityPath) Scan(value interface{}) error {
	data, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Value not instance of string:", value))
	}

	*sD = *EntityPathFromString(data)
	return nil
}

func (sD *EntityPath) Value() (driver.Value, error) {
	return sD.ToString(), nil
}

type EventInterface interface {
	EncodeBytes() ([]byte, error)
	GetValidator() PublicKeyString
	GetSignature() string
	ValidateData(config *configs.MainConfiguration)  (authState any, err error)
}

// type StateEvents struct {
// 	Application EventPath `json:"app"`
// 	Authorization EventPath `json:"auth"`
// 	Topic EventPath`json:"top"`
// 	Subscription EventPath`json:"sub"`
// 	Message EventPath `json:"message"`
// }
type StateEvents struct {
	ID string `json:"id"`
	Event EventPath `json:"e"`
}

func (sh StateEvents) EncodeBytes() ([]byte) {
	b, _ := encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(sh.Event.ID)},
	)
	return b
}

type EventDelta struct {
	Delta json.RawMessage `json:"d"`
	Event Event `json:"e"`
	PreviousHash json.RawMessage `json:"ph"`
	Hash json.RawMessage `json:"hash"`
	Signatures []SignatureData `json:"sig"`
	Error string  `json:"err"`
}

func (e EventDelta) MsgPack() ([]byte) {
	p, _ := encoder.MsgPackStruct(e)
	return p
}

func (e EventDelta) EncodeBytes() ([]byte, error) {
	b, err :=  e.Event.EncodeBytes()
	if err != nil {
		return nil, err
	}
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: e.Delta},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: e.PreviousHash},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType,Value: b},
	)
}
func (e EventDelta) GetHash() ([]byte, error) {
	b, err := e.EncodeBytes()
		if err != nil {
			return nil, err
		}
		return crypto.Sha256(b), nil
}

type EventProcessorResponse struct {
	State interface{} `json:"state"`
	Event EventPath `json:"e"`
	Hash  string `json:"hash"`
	Signature json.RawMessage `json:"sign"`
	Validator json.RawMessage `json:"val"`
	Error string  `json:"err"`
}
// type RemoteEvent struct {
// 	Event *Event
// 	ResponseChannel *chan *EventDelta
// }
type Event struct {
	Version float32 `json:"_v"`
	// Primary
	ID string `gorm:"primaryKey;type:uuid;not null" json:"id,omitempty"`

	Payload           ClientPayload `json:"pld"  msgpack:",noinline"`
	Nonce             string        `json:"nonce,omitempty" gorm:"type:varchar(80);default:null" msgpack:",noinline"`
	Timestamp         uint64        `json:"ts"`
	EventType         constants.EventType        `json:"t,omitempty"`
	
	PreviousEvent EventPath     `json:"preE,omitempty" gorm:"type:varchar;default:null"`
	AuthEvent     EventPath     `json:"authE,omitempty" gorm:"type:varchar;default:null"`
	PayloadHash       string        `json:"pH,omitempty" gorm:"type:char(64);unique"`
	StateEvents	[]StateEvents `json:"sEs"  msgpack:",noinline"`
	// StateHash string `json:"sh"`
	// Secondary
	Error       string          `json:"err,omitempty"`
	Hash        string          `json:"h,omitempty" gorm:"type:char(64)"`
	Signature   string          `json:"sig,omitempty"`
	Broadcasted bool            `json:"br,omitempty"`
	BlockNumber uint64          `json:"blk,omitempty"`
	Cycle   	uint64			`json:"cy,omitempty"`
	Epoch		uint64			`json:"ep,omitempty"`
	IsValid     *bool            `json:"isVal,omitempty" gorm:"default:false"`
	Synced      *bool            `json:"sync,omitempty" gorm:"default:false"`
	Validator   PublicKeyString `json:"val,omitempty"`
	Application   	string			`json:"app,omitempty"`
	Index int64 `json:"vec,omitempty"`
	NodeCount int32 `json:"n,omitempty"`

	Total int `json:"total,omitempty"`
	// Associations      []string      `json:"assoc,omitempty" gorm:"type:text[]"` // deprecated
}

func (g *Event) GetKeys() (keys []string)  {
	keys = append(keys, g.ApplicationKey())
	keys = append(keys, g.BlockKey())
	// keys = append(keys, fmt.Sprintf("cy/%d/%d/%s/%s", g.Cycle, utils.IfThenElse(g.Synced, 1,0), g.Application, g.ID))
	
	keys = append(keys, g.DataKey())
	
	// keys = append(keys, fmt.Sprintf("%s/%s/%s", EntityModel, g.Application, g.ID))
	// keys = append(keys,fmt.Sprintf("hash/%s",  g.GetIdHash()))
	// keys = append(keys,fmt.Sprintf("%s/%s", TopicModel, g.Hash))
	// keys = append(keys,fmt.Sprintf("prev/%s", g.PreviousEvent.ID ))
	
	// keys = append(keys, fmt.Sprintf("cy/%d/%s/%s", g.Cycle, GetModelTypeFromEventType(constants.EventType(g.EventType)), g.ID))
	return keys;
}
func (e *Event) DataKey() string {
	return fmt.Sprintf("id/%s",  e.ID)
}

func (e *Event) VectorKey(topic string) string {
	return fmt.Sprintf("vec/%s/%s", e.Validator, topic)
}

func (e *Event) ApplicationKey()  string {
	return  fmt.Sprintf("app/%s/%015d", e.Application, e.Cycle)
}
func (e *Event) BlockKey() string {
	// if e.ID != "" {
	// 	return  fmt.Sprintf("bl/%d/%d/%s/%s/%s", e.BlockNumber, utils.IfThenElse(e.Synced == nil || !*e.Synced, 0,1), GetModelTypeFromEventType(constants.EventType(e.EventType)), e.Application, e.ID)
	// }
	// if e.Application != "" {
	// 	return  fmt.Sprintf("bl/%d/%d/%s/%s", e.BlockNumber, utils.IfThenElse(e.Synced == nil || !*e.Synced, 0,1), GetModelTypeFromEventType(constants.EventType(e.EventType)), e.Application)
	// }
	if e.ID != "" {
		return utils.CreateKey("/", "bl",fmt.Sprintf("%020d/%015d",  e.BlockNumber, time.Now().UnixMilli()),  e.ID)
		// return  fmt.Sprintf("/bl/%020d/%015d/%s", e.BlockNumber, time.Now().UnixMilli(), e.ID)
	}
	// if e.EventType != 0 {
	// 	return  fmt.Sprintf("bl/%d/%d", e.BlockNumber, utils.IfThenElse(e.Synced == nil || !*e.Synced,0,1))
	// }
	// if e.Synced != nil {
	// 	return  fmt.Sprintf("/bl/%020d/", e.BlockNumber, utils.IfThenElse(e.Synced == nil || !*e.Synced, 0,1))
	// }
	if e.BlockNumber != 0 {
		return  fmt.Sprintf("/bl/%020d", e.BlockNumber)
	}    
	return "bl"
}
func (d *Event) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		// //hash, _ := d.GetId()
		// if d.Signature == "" {
		// 	return fmt.Errorf("signature is required")
		// }
		// hash, err := hex.DecodeString(d.Signature[:16])
		// if err != nil {
		// 	return err
		// }
		// u, err := uuid.FromBytes(hash)
		// if err != nil {
		// 	return err
		// }

		d.ID, err = d.GetId()
		return err
	}
	if d.Nonce == "" && d.Payload.Nonce > 0 {
		d.Nonce = fmt.Sprintf("%s:%d", string(d.Payload.Account), d.Payload.Nonce)
	}
	return nil
}

func (d *Event) GetId() (string, error) {
	if d.Signature == "" {
		return  "", fmt.Errorf("signature is required")
		}
	hash, err := hex.DecodeString(d.GetIdHash())
		if err != nil {
			return "", err
		}
		u, err := uuid.FromBytes(hash)
		if err != nil {
			return "", err
		}

		return u.String(), nil
}

func (d *Event) GetIdHash() (string)  {
	if d.Signature == "" {
		return ""
	}
	return d.Signature[:32]
}


func (e *Event) ToJSON() []byte {
	m, err := json.Marshal(e)
	if err != nil {
		logger.Errorf("Unable to parse event to []byte")
	}
	return m
}

func (e *Event) MsgPack() []byte {
	defer utils.TrackExecutionTime(time.Now(), "EventPacked::")
	b, _ := encoder.MsgPackStruct(e)
	return b
}
func (e *Event) GetDataModelType() EntityModel {
	return GetModel(e.Payload.Data)
}

func (e *Event) IsLocal(cfg *configs.MainConfiguration) bool {
	return e.Validator == PublicKeyString(cfg.PublicKeyEDDHex)
}


func GetModel(ent any) EntityModel {
	var model EntityModel
	val := reflect.ValueOf(ent)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	switch val.Interface().(type) {
		case Application:
			// logger.Debug(val)
			model = ApplicationModel
		case Authorization:
			model = AuthModel
		case Topic:
			model = TopicModel
		case Subscription:
			model = SubscriptionModel
		case Message:
			model = MessageModel
		case SystemMessage:
			model = SystemModel
	}
	return model
}

func (e *Event) GetPath() *EventPath {
	id, err := e.GetId()
	if err != nil {
		return nil
	}
	path := NewEventPath(e.Validator, e.GetDataModelType(), id)
	path.Timestamp = e.Timestamp
	path.NodeCount = e.NodeCount
	if e.Index > 0 {
		path.Index = e.Index
	}
	return path
}
func (e *Event) Lean() Event {
	
	return Event{
		ID: e.ID,
		AuthEvent: e.AuthEvent,
		PreviousEvent: e.PreviousEvent,
		Timestamp: e.Timestamp,
		Validator: e.Validator,
		Hash: e.Hash,
	}
}

func UnpackEvent(b []byte, model EntityModel) (*Event, error) {
	// e.Payload = payload
	e := Event{}
	if err := encoder.MsgPackUnpackStruct(b, &e); err != nil {
		return nil, err
	}
	
	
	dBytes, err := encoder.MsgPackStruct(e.Payload.Data)
	if err != nil {
		return nil, err
	}
	switch model {
	case AuthModel:
		r := Authorization{}
		encoder.MsgPackUnpackStruct(dBytes, &r)
		e.Payload.Data = r
	case ApplicationModel:
		r := Application{}
		encoder.MsgPackUnpackStruct(dBytes, &r)
		e.Payload.Data = r
	case TopicModel:
		r := Topic{}
		encoder.MsgPackUnpackStruct(dBytes, &r)
		logger.Debugf("PAYLOADDDDD %v", r)
		e.Payload.Data = r
	case SubscriptionModel:
		r := Subscription{}
		encoder.MsgPackUnpackStruct(dBytes, &r)
		e.Payload.Data = r
	case MessageModel:
		r := Message{}
		encoder.MsgPackUnpackStruct(dBytes, &r)
		e.Payload.Data = r
	case SystemModel:
		r := SystemMessage{}
		encoder.MsgPackUnpackStruct(dBytes, &r)
		e.Payload.Data = r
	}
	
	_, err2 := e.Payload.EncodeBytes()
	if err2 != nil {
		logger.Errorf("EncodeBytesError:: %o", err2)
	}
	// copier.Copy(e.Payload, &pl)
	newEvent := Event{
		Payload: e.Payload,
	}
	copier.Copy(&e.Payload, &newEvent.Payload)

	return &e, nil
}


func UnpackEventGeneric(b []byte) (*Event, error) {
	// e.Payload = payload
	e := Event{}
	if err := encoder.MsgPackUnpackStruct(b, &e); err != nil {
		return nil, err
	}
	return &e, nil
}



func EventFromJSON(b []byte) (Event, error) {
	var e Event
	// if err := json.Unmarshal(b, &message); err != nil {
	// 	panic(err)
	// }
	err := json.Unmarshal(b, &e)
	return e, err
}

func (e Event) GetHash() ([]byte, error) {
	b, err := e.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	return crypto.Sha256(b), nil
}

func (e Event) GetHIdHash() ([]byte, error) {
	b, err := e.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	return crypto.Sha256(b), nil
}

func (e Event) ToString() string {
	values := []string{}
	d, _ := json.Marshal(e.Payload)
	values = append(values, fmt.Sprintf("%d", d))
	values = append(values, e.ID)
	values = append(values, fmt.Sprintf("%d", e.BlockNumber))
	values = append(values, fmt.Sprintf("%d", e.EventType))
	// values = append(values, strings.Join(e.Associations, ","))
	values = append(values, e.PreviousEvent.ToString())
	values = append(values, e.AuthEvent.ToString())
	values = append(values, fmt.Sprintf("%d", e.Timestamp))
	return strings.Join(values, "")
}

func (e Event) EncodeBytes() ([]byte, error) {

	d, err := e.Payload.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	// previousEvent := []byte{}
	// if e.PreviousEvent.ID  != "" {
	// 	previousEvent, err =  e.PreviousEvent.ToByteHash()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	stateEventsBytes := []byte{}
	for _, v := range e.StateEvents {
		stateEventsBytes = append(stateEventsBytes, v.EncodeBytes()...)
	}
	
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: d},
		// encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: strings.Join(e.Associations, "")},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(e.AuthEvent.ID)},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: e.BlockNumber},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: e.Cycle},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: e.Epoch},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: e.EventType},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(e.PreviousEvent.ID)},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(e.Application)},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: stateEventsBytes},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: e.Timestamp},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: e.Index},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: e.NodeCount},
		
	)
}

func (e Event) GetValidator() PublicKeyString {
	return e.Validator
}
func (e Event) GetSignature() string {
	return e.Signature
}


func GetEventEntityFromModel(eventType EntityModel) *Event {
	// cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	// check if client payload is valid
	// if err := payload.Validate(PublicKeyString(cfg.NetworkPublicKey)); err != nil {
	// 	return  err
	// }

	//Perfom checks base on event types
	event := &Event{Payload: ClientPayload{}}
	switch eventType {
	case AuthModel:
		event.Payload.Data = Authorization{}

	case TopicModel:
		event.Payload.Data = Topic{}

	case SubscriptionModel:
		event.Payload.Data = Subscription{}

	case MessageModel:
		event.Payload.Data = Message{}

	case ApplicationModel:
		event.Payload.Data = Application{}

	case SystemModel:
		event.Payload.Data = SystemMessage{}

	case WalletModel:
		event.Payload.Data = Wallet{}
	}

	return event

}



func GetStateModelFromEntityType(entityType EntityModel) any {
	switch entityType {
	case AuthModel:

		return Authorization{}

	case TopicModel:
		return Topic{}

	case SubscriptionModel:
		return Subscription{}

	case MessageModel:
		return Message{}

	case ApplicationModel:
		return Application{}
	case SystemModel:
		return SystemMessage{}

	case WalletModel:
		return Wallet{}
	}

	return 0

}


func GetModelTypeFromEventType(eventType constants.EventType ) EntityModel {
	if eventType < 50 {
		return  SystemModel	
	}
	if eventType < 600 {
		return  ApplicationModel	
	}
	if eventType < 700 {	
		return  AuthModel	
	}
	if eventType < 1100 {
		
		return  TopicModel	
	}
	if eventType < 1200 {
		return  SubscriptionModel	
	}
	if eventType < 1300 {
		return  MessageModel	
	}
	
	if eventType < 1500 {
		return  WalletModel	
	}
	return ""

}

func CycleCounterKey(cycle uint64, validator *PublicKeyString, claimed *bool, app *string) string {
	if validator == nil && claimed == nil && app == nil {
		return fmt.Sprintf("cy/%015d/n", cycle)
	}
	if claimed == nil {
		return fmt.Sprintf("cy/%015d/%s", cycle, *validator )
	}
	if app == nil {
		return fmt.Sprintf("cy/%015d/%s/%d", cycle, *validator, utils.IfThenElse(*claimed, 1, 0) )
	}
	return fmt.Sprintf("cy/%015d/%s/%d/%s", cycle, *validator, utils.IfThenElse(*claimed, 1, 0), *app)
}


func NetworkCounterKey(app *string) string {
	if app == nil {
		return "net/"
	}
	return fmt.Sprintf("net/%s", *app)
}
func CycleApplicationKey(cycle uint64, app string) string {
	return fmt.Sprintf("cyc/%015d/%s", cycle, app)
}


type EventPayload struct {
		EventType         constants.EventType        `json:"t,omitempty"`
		Event json.RawMessage  `json:"e,omitempty"`
}
