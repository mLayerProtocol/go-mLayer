package constants

import (
	"html/template"
	"time"
)


const (
    NETWORK_NAME = "mlayer" // time interval within which to accept a handshake
	VALID_HANDSHAKE_SECONDS = 15 // time interval within which to accept a handshake
    TOPIC_INTEREST_TTL = 12 * time.Hour // duration a node is considered still interested in a topic after initialy showing interest
)



var VALID_PROTOCOLS = []string{"/mlayer/1.0.0"}

const (
	DefaultPKeyPath string = "./data/.local.key"
)
const (
	DefaultRPCPort            string = "9525"
	DefaultWebSocketAddress   string = "0.0.0.0:8088"
	DefaultRestAddress        string = ":9531"
	DefaultDataDir            string = "./data"
	DefaultMLBlockchainAPIUrl string = ":9520"
    DefaultQuicPort        uint16 = 9533
	DefaultProtocolVersion           string = "/mlayer/1.0.0"
)

type ContextPath string;
const (
    ConnectedSubscribersMap  ContextPath  = "connected-subscribers-map"
)

type NodeType uint
const (
	ValidatorNodeType NodeType = 1
    SentryNodeType     NodeType = 2
)

const ADDRESS_ZERO = "0000000000000000000000000000000000000000"
const MaxBlockSize = 1000

const (
	ErrorUnauthorized = "4001"
	ErrorBadRequest   = "4000"
	ErrorForbidden    = "4003"
)


// Channel Ids within main context
type ChannelId string
const (
	ConfigKey                       ChannelId = "Config"
	SQLDB                         ChannelId  = "sqldb"
)

// State store key
const (
	CurrentDeliveryProofBlockStateKey string = "/df-block/current-state"
	CurrentSubscriptionBlockStateKey  string = "/sub-block/current-state"
)

type Protocol string

const (
	WS   Protocol = "ws"
	MQTT Protocol = "mqtt"
	RPC  Protocol = "rpc"
)

type ApplicationCategory int16

const (
	CategoryGeneral       ApplicationCategory = 1
	CategoryMobility      ApplicationCategory = 2
	CategoryEnergy        ApplicationCategory = 3
	CategoryEnvironmental ApplicationCategory = 4
	CategoryHealthcare    ApplicationCategory = 5
	CategorySmartCity     ApplicationCategory = 6
	CategorySmartHome     ApplicationCategory = 7
	CategoryGeoLocation   ApplicationCategory = 8
	CategoryP2PMessaging  ApplicationCategory = 9
	CategorySharedCompute ApplicationCategory = 10
	CategoryFileSharing   ApplicationCategory = 11
)

const SignatureMessageString string = `{"action":"%s","identifier":"%s","network":"%s","hash":"%s"}`


const MAX_SYNC_FILE_SIZE  = 100 * 1024 * 1024 // 100 MB
/* KEY MAPS
Always enter map in sorted order

Abi             = abi
Account pu        = acct
Action          = a
Actions         = as
Address         = addr
Agent           = agt
Amount          = amt
Approval        = ap
ApprovalExpiry  = apExp
Associations    = assoc
Attachments		= atts
Authority       = auth
Avatar          = ava
Body            = b
Block           = blk
BlockId         = blkId
Broadcast       = br
Chain           = c
ChainId         = cId
ChannelExpiry   = chEx
ChannelName     = chN
Channel         = ch
Channels        = chs
CID             = cid
Closed          = cl
Contract        = co
Data            = d
DataHash        = dH
Description     = desc
Device			= dev
Duration        = du
Event			= e
EventHash       = eH
File			= f
Grantor         = gr
Handle          = hand
Hash            = h
Header          = head
Identifier		= id
Index           = i
InviteOnly 		= invO
Interval		= inter
IsValid         = isVal
Length          = len
Message         = m
MessageHash     = mH
MessageSender   = mS
Name            = nm
Node            = n
OperatorAddress     = nA
NodeHeight      = nH
NodeSignature   = nS
NodeType        = nT
Origin          = o
Owner			= own
Parameters      = pa
ParentTopic = pTH
Paylaod         = pld
Platform        = p
Privilege       = privi
Proof           = pr
Proofs          = prs
ProtocolId      = proId
PublicKey		= pubK
Range			= range
ReadOnly		= rO
Receiver        = r
Ref             = ref
Roles 			= rol
Secret          = sec
Sender          = s
SenderAddress   = sA
SenderSignature = sSig
SignatureExpiry = sigExp
Signature       = sig
SignatureData   = sigD
Signer          = sigr
Size            = si
Socket          = sock
Status          = st
Subject         = su
SubjectHash     = suH
Application        	= app
Subscriber      = sub
Synced          = sync
Timestamp       = ts
TopicHash       =topH
TopicId         = topId
TopicIds        = topIds
Type            = ty
Url				= url
Validator       = v







*/

var HomeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))

/*
{\"type\": \"cosmos-sdk/StdTx\",
\"value\": {\"msg\": [{  \"type\": \"sign/MsgSignData\",  \"value\": {\t\"signer\": \"cosmos14y0pyqjay3p8dsqp2jd5rkft7vf9cdkqnrc43l\",\t\"data\": \"aGVsbG93b3JsZA==\"  }}],\"fee\": {\"amount\": [],\"gas\": \"0\"},\"memo\": \"\"}}
*/
