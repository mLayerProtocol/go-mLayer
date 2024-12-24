package ws

import (
	// "errors"

	"context"
	"encoding/json"
	"fmt"
	"strings"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/channelpool"
	"github.com/mlayerprotocol/go-mlayer/pkg/client"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
)

var logger = &log.Logger

type Flag string

// Main websocket service that receives messages and directs them to the neccessry processing channel

// type msgError struct {
// 	code int
// 	message string
// }

const (
	SubescriptionRequest client.RequestType = "WRITE:__subscribe__"
)



type WsService struct {
	Ctx                       *context.Context
	Cfg                       *configs.MainConfiguration
	// ClientHandshakeChannel    *chan *entities.ClientHandshake
	// ClientSubscriptionChannel *chan *entities.ClientWsSubscription
	RequestProcess            *client.ClientRequestProcessor
}

type WsPayload struct {
	Id            string                 `json:"id"`
	RequestType   client.RequestType     `json:"rTy"`
	Params        map[string]interface{} `json:"params"`
	ClientPayload entities.ClientPayload `json:"pl"`
}

func (p *WsPayload) MsgPack() ([]byte, error) {
	return encoder.MsgPackStruct(p)
}

func MsgUnpackWsPayload(data []byte, p *WsPayload) error {
	return encoder.MsgPackUnpackStruct(data, p)
}

func NewWsService(mainCtx *context.Context) *WsService {
	cfg, _ := (*mainCtx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	//clientVerificationc, _ := (*mainCtx).Value(constants.ClientHandShackChId).(*chan *entities.ClientHandshake)
	return &WsService{
		Ctx:                    mainCtx,
		Cfg:                    cfg,
		// ClientHandshakeChannel: clientVerificationc,
		RequestProcess:         client.NewClientRequestProcess(mainCtx),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (p *WsService) HandleConnection(w http.ResponseWriter, r *http.Request) {
	logger.Debug("New Attempted WebSocket connection")
	// authHeader := r.Header.Get("Authorization")
	// token := authHeader[strings.Index(authHeader, " "):]

	// tokenBytes, err := hex.DecodeString(token)

	// if err != nil {
	// 	tokenBytes, err = base64.StdEncoding.DecodeString(token)
	// 	if err != nil {
	// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 		return
	// 	}
	// }

	// handshake, err := entities.ClientHandshakeFromJson(tokenBytes)
	// if err != nil {
	// 	http.Error(w, "Invalid handshake in Authorization header", http.StatusUnauthorized)
	// 	return
	// }

	// _, err = service.ConnectClient(p.Cfg, handshake, constants.WS)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusUnauthorized)
	// 	return
	// }
	handshake := entities.ClientHandshake{}
	c, err := upgrader.Upgrade(w, r, nil)
	//handshake.ClientSocket = c

	logger.Debug("New ServeWebSocket c : ", c.RemoteAddr())

	if err != nil {
		logger.Debug("WS connection error:", err)
		return
	}
	defer c.Close()
	// var mt int
	for {
		// b := bytes.Buffer{}

		// for {
		// 	t, message, err := c.ReadJSON()
		// 	logger.Debugf("READWSMESSAGE: type %v, message  %v, err %v", t, len(b.Bytes()), err)
		// 	// mt = t
		// 	if len(message) > 0 {
		// 		b.Write(message)
		// 	}
		
		// 	if err != nil  {
		// 		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		// 			// remove from subscribers
		// 		} 
		// 		logger.Println("read:", err)
		// 		break

		// 	}
		// }
		// logger.Debug("READWSMESSAGE", b.String())

			
			
		
			// err = json.Unmarshal(b.Bytes(), &payload)
			// if err != nil {
			// 	err = c.WriteJSON(entities.ClientResponse{
			// 		Error:        err.Error(),
			// 		ResponseCode: int(apperror.BadRequestError),
			// 		Id:           payload.Id,
			// 	})
			// 	if err != nil {
			// 		logger.Error(err)
			// 	}
			// 	continue
			// }
			payload := WsPayload{}
			err = c.ReadJSON(&payload)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, ) || strings.Contains(err.Error(), "close 1006") {
					// remove from subscribers
					return
				} 
				logger.Println("ReadERROR:", err, payload)
			}
		
			clientPayload := entities.ClientPayload{}
			pBytes, err := json.Marshal(payload.ClientPayload)
			if err != nil {
				err = c.WriteJSON(entities.ClientResponse{
					Error:        err.Error(),
					ResponseCode: int(apperror.BadRequestError),
					Id:           payload.Id,
				})
				if err != nil {
					logger.Error("HandleConnection/WriteJSON: ", err)
				}
				continue
			}
			logger.Debugf("Received PAYLOAD %v", payload)
			err = json.Unmarshal(pBytes, &clientPayload)
			if err != nil {
				err = c.WriteJSON(entities.ClientResponse{
					Error:        err.Error(),
					ResponseCode: int(apperror.BadRequestError),
					Id:           payload.Id,
				})
				if err != nil {
					logger.Error("HandleConnection/WriteJSON2: ",err)
				}
				continue
			}
		
			// check if the user is trying to request
			
			if payload.RequestType == SubescriptionRequest {
			
				filter := map[string][]string{}
				for key, value := range payload.Params {
					filter[key] = []string{}
					for _, val := range value.([]interface{}) {
						filter[key] = append(filter[key], fmt.Sprintf("%s", val))
					}
					
				}
				
				channelpool.ClientWsSubscriptionChannel <- &entities.ClientWsSubscription{
					Conn:   c,
					Filter: filter,
					Id: payload.Id,
					Account: string(handshake.Account),
				}

				resp := entities.ClientResponse{
					ResponseCode: 200,
					Id:           payload.Id,
				}
				err = c.WriteJSON(resp)
				if err != nil {
					resp = entities.ClientResponse{
						Error:        err.Error(),
						ResponseCode: int(apperror.BadRequestError),
						Id:           payload.Id,
					}
					c.WriteJSON(resp)
					continue
				}
				continue
			}
			response, err := p.RequestProcess.Process(payload.RequestType, payload.Params, clientPayload)
			if err != nil {
				c.WriteJSON(entities.ClientResponse{
					ResponseCode: int(apperror.BadRequestError),
					Error:           err.Error(),
					Id:        payload.Id,
				})
				continue
			}
			logger.Debugf("Sending PAYLOAD %v", entities.ClientResponse{
				ResponseCode: 200,
				Id:           payload.Id,
				Data:         response,
			})
			c.WriteJSON(entities.ClientResponse{
				ResponseCode: 200,
				Id:           payload.Id,
				Data:         response,
			})
			// else {
			// 	err = c.WriteMessage(mt, (append(message, []byte("recieved Signature")...)))
			// 	if err != nil {
			// 		logger.Println("Error:", err)
			// 	} else {
			// 		if(!isVerifed) {
			// 			verifiedRequest, err := service.ConnectClient(message, constants.WS, c,)
			// 			if (err != nil) {
			// 				c.Close()
			// 				continue
			// 			}
			// 			*p.ClientHandshakeChannel <- verifiedRequest

			// 			logger.Debugf("message:", string(message))
			// 			logger.Debugf("recv: %s - %d - %s\n", message, mt, c.RemoteAddr())
			// 			continue
			// 		}
			// 		// process message
			// 	}

			 }
		
	//}

}
