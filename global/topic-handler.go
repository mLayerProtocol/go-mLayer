package global

import (
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/mlayerprotocol/go-mlayer/smartlet"
)

type Keystore struct {
	Key []byte
	Value []byte
}

type DataStore struct {
	ID []byte
	Values []Keystore
}


type NotifierType int8

const (
	Http NotifierType = 1
)

type Notifier struct {
	ID []byte
	Type NotifierType
	Endpoint string
	Payload []byte

}

// var GlobalTopicHandlers = map[string]func (ctx *context.Context) (result *Result, err error) {
// 	REGISTERY_TOPIC_REF: HandleRegisteryMessage,
// }

var GlobalTopicHandlers = map[string]smartlet.SmartletInterface{
	REGISTERY_TOPIC_REF: HandleRegisteryMessage{},
}



// type Result struct {
// 	Notifier Notifier
// 	App *smartlet.App
// }
const (
	ActionCreate uint8 = 1
)

type HandleRegisteryMessage struct {

}
func (hr HandleRegisteryMessage) PreCommit(app  smartlet.App) (_app smartlet.App, err error) {
	// app := (*ctx).Value(smartlet.Application).(*smartlet.App)
	// message := app.Event.Payload.Data.(entities.Message)
	// // validate the sender, ensure he/she has enough MLT stakes
	
	// data, err := message.GetData()
	// if err != nil {
	// 	return err
	// }
	
	// app.KeyStore.Get()
	
	// action := smartlet.ActionParams{}
	message := app.Event.Payload.Data.(entities.Message)
	// validate the sender, ensure he/she has enough MLT stakes
	
	data, err := message.GetData()
	if err != nil {
		return app, err
	}
	


	action := smartlet.ActionParams{}
	err = smartlet.DecodeActionParams(data, &action)
	if err != nil {
		return app, err
	}
	service := entities.RegistryService{}
	// service := map[string]interface{}{}
	err = encoder.MsgPackUnpackStruct(action.Params, &service)
	fmt.Printf("ENITTYDATA %v, %v", action.Action, service)
	app.SetState(action.Action, action.Params)
	service =  entities.RegistryService{}
	return app, err
}

func (hr HandleRegisteryMessage) PostCommit(app  smartlet.App) (_app  smartlet.App, err error) {
	
	


	args, exists := app.GetState(ActionCreate)
	

	if exists {
		return newRegistration(&app, args)
		
	}

	
	// cacheKey := hex.EncodeToString(service.AdminAuthority.Signature[:8])
	// if _, exists := stores.SystemCache.Get(cacheKey); exists {
	// 	return nil, fmt.Errorf("duplicate event")
	// }
	// stores.SystemCache.SetWithTTL(cacheKey, []byte{}, 15 * time.Second)

	return app, err
}

func newRegistration(app *smartlet.App, serviceData []byte) (_app smartlet.App, err error) {
	service := entities.RegistryService{}
	err = encoder.MsgPackUnpackStruct(serviceData, &service)
	if err != nil {
		return *app, err
	}
	serviceData = []byte{}
	// if service.AdminAuthority.Timestamp < app.Event.Timestamp && app.Event.Timestamp - service.AdminAuthority.Timestamp > uint64(15 * time.Second) {
	// 	return nil, fmt.Errorf("expired admin_authority")
	// }

	// if service.AdminAuthority.Timestamp > app.Event.Timestamp && service.AdminAuthority.Timestamp - app.Event.Timestamp > uint64(5 * time.Second) {
	// 	return nil, fmt.Errorf("admin_authority applied too early")
	// }
	// authorityBytes, _ := service.AdminAuthority.EncodeBytes()
	// if valid := crypto.VerifySignatureECC(service.AdminAuthority.ServiceAccount, &authorityBytes, hex.EncodeToString(service.AdminAuthority.Signature)); !valid {
	// 	return nil, fmt.Errorf("invalid admin_authority")
	// }
	// if app.Event.Payload.Account !=  entities.AccountString(service.AdminAuthority.AuthorizedAccount) {
	// 	return nil, fmt.Errorf("event account does not match authorized_account")
	// }
	// if service.Metadata.ExternalID != service.Owner {
	// 	return *app, fmt.Errorf("admin_authority.matadata_id does not match metadata.id")
	// }
	// if entities.AddressFromString(service.AdminAuthority.ServiceAccount).ToString() != entities.AccountString(app.Event.Payload.Account.ToString()) {
	// 	return *app, fmt.Errorf("account authorization string")
	// }
	// addr, _ := hex.DecodeString(utils.AddressToHex(service.AdminAuthority.ServiceAccount))
	addrString, err := entities.AddressFromString(string(app.Event.Payload.Account))
	if err != nil {
		return *app, err
	}
	addr :=  addrString.EncodeBytes()
	addr = append(addr, []byte(service.Metadata.ExternalID)...)
	serviceId, err := uuid.FromBytes(crypto.Keccak256Hash(addr)[:16])
	if err != nil {
		return *app, err
	}
	service.Id = serviceId.String()
	// registryKey := fmt.Sprintf("%s/%s/%s/%s", "r", app.Topic.ID, service.ServiceType, serviceId )
	// service.Uri = registryKey
	packed, err :=  encoder.MsgPackStruct(service)
	if err != nil {
		return *app, err
	}
	err = app.KeyStore.Set(smartlet.Key(serviceId), packed, true)
	if err != nil {
		return *app, err
	}
	app.SetResult(smartlet.NewKey([]byte("id")), serviceId[:])
	

	fmt.Printf("NewServiceID %s\n", hex.EncodeToString(serviceId[:]))
	fmt.Printf("NewServiceID %s\n", serviceId.String())
	app.KeyStore.Append(smartlet.NewKey([]byte(service.ServiceType)), serviceId, []byte(utils.CurrentTimestampAsString())) 
	// err = app.AppendStore(smartlet.NewKey([]byte(service.ServiceType)), serviceId, []byte(utils.CurrentTimestampAsString())) // allows the value to be sorted
	// app.Commit()
	// result.App = app
	// fmt.Printf("GetUpdateKeys: %v\n", app.KeyStore.GetUpdateKeys())
	return *app, err
}

