/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"sync"

	// "net/rpc/jsonrpc"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/chain/api"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"

	// "github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
	"github.com/mlayerprotocol/go-mlayer/pkg/node"
	"github.com/spf13/cobra"
)

var logger = &log.Logger



type Flag string

const (
	NETWORK_ADDRESS_PRFIX Flag = "network-address-prefix"
	CHAIN_ID Flag = "chain-id"
	PRIVATE_KEY Flag = "private-key"
	PROTOCOL_VERSION    Flag  = "protocol-version"
	RPC_PORT            Flag = "rpc-port"
	WS_ADDRESS          Flag = "ws-address"
	REST_ADDRESS        Flag = "rest-address"
	HOSTNAME        Flag = "hostname"
	QUIC_PORT        Flag = "quic-port"
	EXT_IP        Flag = "ext-ip"
	DATA_DIR            Flag = "data-dir"
	LISTENERS            Flag = "listen"
	KEYSTORE_DIR         Flag = "keystore-dir"
	KEYSTORE_PASSWORD         Flag = "keystore-password"
	NO_SYNC         Flag = "no-sync"
	SYNC_BATCH_SIZE         Flag = "sync-batch-size"
	TESTNET_MODE         Flag = "testnet"
	MAINNET_MODE         Flag = "mainnet"
	VALIDATOR_MODE         Flag = "validator"
	SENTRY_MODE         Flag = "sentry"
	ARCHIVE_MODE         Flag = "archive"
	TEST_MODE         Flag = "testing"
	VERBOSE         Flag = "verbose"
	BOOTSTRAP_NODE         Flag = "bootstrap-node"
	SYNC_HOST         Flag = "sync-host"
)
const MaxDeliveryProofBlockSize = 1000

var deliveryProofBlockMutex sync.RWMutex

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Runs goml as a daemon",
	Long: `Use this command to run goml as a daemon:

	mLayer (message layer) is an open, decentralized 
	communication network that enables the creation, 
	transmission and termination of data of all sizes, 
	leveraging modern protocols. mLayer is a comprehensive 
	suite of communication protocols designed to evolve with 
	the ever-advancing realm of cryptography. 
	Visit the mLayer [documentation](https://mlayer.gitbook.io/introduction/what-is-mlayer) to learn more
	.`,
	Run: func(cmd *cobra.Command, args []string) {
		daemonFunc(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringP(string(NETWORK_ADDRESS_PRFIX), "p", "", "The network address prefix. This determines the operational network e.g. ml=>mainnet, mldev=>devnet,mltest=>testnet")
	daemonCmd.Flags().StringP(string(PRIVATE_KEY), "k", "", "The deligated operators private key. This is the key used to sign handshakes and messages. The coresponding public key must be assigned to the validator")
	daemonCmd.Flags().StringP(string(PROTOCOL_VERSION), "", constants.DefaultProtocolVersion, "Protocol version")
	daemonCmd.Flags().StringP(string(RPC_PORT), "r", constants.DefaultRPCPort, "RPC server port")
	daemonCmd.Flags().StringP(string(WS_ADDRESS), "w", "", "ws service address")
	daemonCmd.Flags().StringP(string(REST_ADDRESS), "R", constants.DefaultRestAddress, "rest api service address")
	daemonCmd.Flags().StringP(string(DATA_DIR), "d", "", "data storage directory")
	daemonCmd.Flags().StringSliceP(string(LISTENERS), "l", []string{}, "libp2p multiaddress array eg. [\"/ip4/127.0.0.1/tcp/5000/ws\", \"/ip4/127.0.0.1/tcp/5001\"]")
	daemonCmd.Flags().StringP(string(KEYSTORE_DIR), "K", "", "path to keystore directory")
	daemonCmd.Flags().StringP(string(KEYSTORE_PASSWORD), "P", "", "password for decripting key store")
	daemonCmd.Flags().BoolP(string(NO_SYNC), "n", false, "do not sync db")
	daemonCmd.Flags().UintP(string(SYNC_BATCH_SIZE), "b", 100, "number of blocks within a sync request. Default 100")
	daemonCmd.Flags().BoolP(string(TESTNET_MODE), "", true, "Run in testnet mode")
	daemonCmd.Flags().BoolP(string(MAINNET_MODE), "", false, "Run in mainnet mode")
	daemonCmd.Flags().BoolP(string(VALIDATOR_MODE), "v", false, "Run as validator")
	daemonCmd.Flags().BoolP(string(TEST_MODE), "", false, "Run test functions")
	daemonCmd.Flags().StringP(string(QUIC_PORT), "", "", "Quic server listening port")
	daemonCmd.Flags().StringP(string(HOSTNAME), "", "" , "Server hostname")
	daemonCmd.Flags().StringP(string(EXT_IP), "", "" , "Server public ip")
	daemonCmd.Flags().BoolP(string(VERBOSE), "", false, "Sets log level to debug")
	daemonCmd.Flags().BoolP(string(BOOTSTRAP_NODE), "", false, "Run as bootstrap node")
	daemonCmd.Flags().StringP(string(SYNC_HOST), "", "", "IP address of sync node")
}

func daemonFunc(cmd *cobra.Command, _ []string) {
	// var systeError error
	testnet, _ := cmd.Flags().GetBool(string(TESTNET_MODE))
	mainnet, _ := cmd.Flags().GetBool(string(MAINNET_MODE))
	if mainnet {
		testnet = false
	}
	configs.Init(testnet)
	cfg := configs.Config
	verbose, _ := cmd.Flags().GetBool(string(VERBOSE))	
	if verbose {
		cfg.LogLevel = "debug"
	}
	log.Initialize(cfg.LogLevel)
	
	ctx := context.Background()

	chain.NetworkInfo = &chain.NetworkParams{Config: &cfg}

	
	defer func() {
		node.Start(&ctx)
	}()
	defer func () {
		// chain.Network = chain.Init(&cfg)
		logger.Println("Initializing chain...")
		time.Sleep(1*time.Second)
		chain.RegisterProvider(
			"31337", api.NewGenericAPI(),
		)
		ethAPI, err := api.NewEthAPI(cfg.ChainId, cfg.EvmRpcConfig[string(cfg.ChainId)], &cfg.PrivateKeySECP)
		if err != nil {
			logger.Fatal("APIERROR: ", err)
		}
		chain.RegisterProvider(
			"84532", ethAPI,
		)
		
		// // chain.DefaultProvider = chain.Network.Default()
		// ownerAddress, _ := hex.DecodeString(constants.ADDRESS_ZERO)
		// if cfg.Validator {
		// 	ownerAddress, err = chain.Provider(cfg.ChainId).GetValidatorLicenseOwnerAddress(cfg.PublicKeySECP)
		// } else {
		// 	ownerAddress, err = chain.Provider(cfg.ChainId).GetSentryLicenseOwnerAddress(cfg.PublicKeySECP)
		// }
		// if err != nil  {
		// 	logger.Fatalf("unable to get license owner: %v", err)
		// }
		// if hex.EncodeToString(ownerAddress) == constants.ADDRESS_ZERO {
		// 	if cfg.Validator {
		// 		logger.Fatalf("Failed to run in validator mode because no license is assigned to this operators public key (SECP).")
		// 	}
		// 	logger.Debugf("Operator not yet deligated. Running is archive mode.")
		// }
		// cfg.OwnerAddress = common.BytesToAddress(ownerAddress)
		// fmt.Println("Chain initialized!")
	}()
	defer func() {
		if r := recover(); r != nil {
			logger.Fatal(r)
		}
	
		sql.Init(&cfg)
	}()
	

	

	rpcPort, _ := cmd.Flags().GetString(string(RPC_PORT))
	wsAddress, _ := cmd.Flags().GetString(string(WS_ADDRESS))
	restAddress, _ := cmd.Flags().GetString(string(REST_ADDRESS))
	listeners, _ := cmd.Flags().GetStringSlice(string(LISTENERS))

	
	cfg.Context = &ctx

	if len(cfg.ChainId) == 0 {
		cfg.ChainId = "ml"
	}
	chainId, _ := cmd.Flags().GetString(string(CHAIN_ID))
	if len(chainId) > 0 {
		cfg.ChainId = configs.ChainId(chainId)
	}

	if len(cfg.ChainId) == 0 {
		cfg.ChainId = "ml"
	}
	prefix, _ := cmd.Flags().GetString(string(NETWORK_ADDRESS_PRFIX))
	if len(prefix) > 0 {
		cfg.AddressPrefix = prefix
	}
	
	dir, _ := cmd.Flags().GetString(string(KEYSTORE_DIR))	
	cfg = injectPrivateKey(&cfg, cmd, getKeyStoreFilePath("account", dir))
	if len(wsAddress) > 0 {
		cfg.WSAddress = wsAddress
	}
	if len(cfg.WSAddress) == 0 {
		cfg.WSAddress = constants.DefaultWebSocketAddress
	}

	

	cfg.SyncBatchSize, _ = cmd.Flags().GetUint(string(SYNC_BATCH_SIZE))
	

	cfg.NoSync, _ = cmd.Flags().GetBool(string(NO_SYNC))
	

	if len(restAddress) > 0 {
		cfg.RestAddress = restAddress
	}

	
	hostname, _ := cmd.Flags().GetString(string(HOSTNAME))
	if hostname != "" {
		cfg.Hostname = hostname
	}

	ip, _ := cmd.Flags().GetString(string(EXT_IP))
	
	// validate IP
	remoteIp, err := utils.GetPublicIPFromDial()
	if err != nil {
		remoteIp, _ = utils.GetPublicIPFromAPI()
		
	}
	
	if  remoteIp != "" {
		if ip != "" && ip != remoteIp {
			logger.Fatal(fmt.Sprintf("invalid external ip. %s provide but %s was detected", ip, remoteIp))
		}
		ip = remoteIp
	} else {
		if ip == "" {
			logger.Fatal("error detecting public IP and external ip not provided")
		}
	}
	cfg.IP = ip

	quicPort, _ := cmd.Flags().GetString(string(QUIC_PORT))
	
	if quicPort != "" {
		qp, err := strconv.Atoi(quicPort)
		if err != nil {
		logger.Fatal(err)

		}
		cfg.QuicPort = uint16(qp)
	}
	if cfg.QuicPort == 0 {
		cfg.QuicPort = constants.DefaultQuicPort
	}
	if len(cfg.Hostname) > 0 {
		cfg.QuicHost = fmt.Sprintf("%s:%d", cfg.Hostname, cfg.QuicPort)
	} else {
		cfg.QuicHost = fmt.Sprintf("%s:%d", cfg.IP, cfg.QuicPort)
	}
	




	dataDir, _ := cmd.Flags().GetString(string(DATA_DIR))
	if len(dataDir) != 0  {
		cfg.DataDir = dataDir
	}
	if len(cfg.DataDir) == 0 {
		cfg.DataDir = constants.DefaultDataDir
	}
	
	cfg.TestMode, _ = cmd.Flags().GetBool(string(TEST_MODE))
	

	if len(cfg.SQLDB.DbStoragePath) == 0 {
		cfg.SQLDB.DbStoragePath = filepath.Join( cfg.DataDir, "store", "sql")
	}
	if strings.HasPrefix(cfg.DataDir, "./") && !strings.HasPrefix(cfg.SQLDB.DbStoragePath, "./") {
		cfg.SQLDB.DbStoragePath = "./"+cfg.SQLDB.DbStoragePath
	}
	if strings.HasPrefix(cfg.DataDir, "../") && !strings.HasPrefix(cfg.SQLDB.DbStoragePath, "../") {
		cfg.SQLDB.DbStoragePath = "../"+cfg.SQLDB.DbStoragePath
	}


	protocolVersion, _ := cmd.Flags().GetString(string(PROTOCOL_VERSION))
	
	if len(protocolVersion) > 0 && protocolVersion != constants.DefaultProtocolVersion  {
		cfg.ProtocolVersion = protocolVersion
	}
	if len(cfg.ProtocolVersion) == 0 {
		cfg.ProtocolVersion = constants.DefaultProtocolVersion
	}

	if !slices.Contains(constants.VALID_PROTOCOLS, cfg.ProtocolVersion) {
		logger.Fatal("Invalid protocol version provided")
	}

	if rpcPort == constants.DefaultRPCPort && len(cfg.RPCPort) > 0 {
		rpcPort = cfg.RPCPort
	}
	if len(rpcPort) > 0 {
		cfg.RPCPort = rpcPort
	}
	if len(cfg.RPCPort) == 0 {
		cfg.RPCPort = constants.DefaultRPCPort
	}
	
	if len(listeners) > 0 {
		cfg.ListenerAdresses = listeners
	}

	validator, _ := cmd.Flags().GetBool(string(VALIDATOR_MODE))	
	if validator  {
		cfg.Validator = true
	}

	isBootstrap, _ := cmd.Flags().GetBool(string(BOOTSTRAP_NODE))
	if !cfg.BootstrapNode {
		cfg.BootstrapNode = isBootstrap
	}
	
	archiveDir := filepath.Join( cfg.DataDir, "archive")
	if strings.HasPrefix(cfg.DataDir, "./") && !strings.HasPrefix(archiveDir, "./") {
		archiveDir = "./"+archiveDir
	}
	if strings.HasPrefix(cfg.DataDir, "../") && !strings.HasPrefix(archiveDir, "../") {
		archiveDir = "../"+archiveDir
	}
	if err = os.MkdirAll(archiveDir, os.ModePerm); err!=nil {
		logger.Fatal(err)
	}
	cfg.ArchiveDir = archiveDir

	cfg.SyncHost, _ = cmd.Flags().GetString(string(SYNC_HOST))
	

	// ****** INITIALIZE CONTEXT ****** //

	ctx = context.WithValue(ctx, constants.ConfigKey, &cfg)

	// ADD EVENT  SUBSCRIPTION CHANNELS TO THE CONTEXT
	// ctx = context.WithValue(ctx, constants.IncomingAuthorizationEventChId, &channelpool.AuthorizationEvent_SubscriptionC)
	// ctx = context.WithValue(ctx, constants.IncomingTopicEventChId, &channelpool.IncomingTopicEventSubscriptionC)

	// ADD EVENT BROADCAST CHANNELS TO THE CONTEXT
	// ctx = context.WithValue(ctx, constants.BroadcastAuthorizationEventChId, &channelpool.AuthorizationEventPublishC)
	// ctx = context.WithValue(ctx, constants.BroadcastTopicEventChId, &channelpool.TopicEventPublishC)
	// ctx = context.WithValue(ctx, constants.BroadcastSubnetEventChId, &channelpool.SubnetEventPublishC)

	// // CLEANUP
	// ctx = context.WithValue(ctx, constants.IncomingMessageChId, &channelpool.IncomingMessageEvent_P2P_D_C)
	// ctx = context.WithValue(ctx, constants.OutgoingMessageChId, &channelpool.NewPayload_Cli_D_C)
	// ctx = context.WithValue(ctx, constants.OutgoingMessageDP2PChId, &channelpool.OutgoingMessageEvents_D_P2P_C)
	// incoming from client apps to daemon channel
	// ctx = context.WithValue(ctx, constants.SubscribeChId, &channelpool.Subscribers_RPC_D_C)
	// daemon to p2p channel
	// ctx = context.WithValue(ctx, constants.SubscriptionDP2PChId, &channelpool.Subscription_D_P2P_C)
	// ctx = context.WithValue(ctx, constants.ClientHandShackChId, &channelpool.ClientHandshakeC)
	// ctx = context.WithValue(ctx, constants.OutgoingDeliveryProof_BlockChId, &channelpool.OutgoingDeliveryProof_BlockC)
	// ctx = context.WithValue(ctx, constants.OutgoingDeliveryProofChId, &channelpool.OutgoingDeliveryProofC)
	// ctx = context.WithValue(ctx, constants.PubsubDeliverProofChId, &channelpool.PubSubInputBlockC)
	// ctx = context.WithValue(ctx, constants.PubSubBlockChId, &channelpool.PubSubInputProofC)
	// // receiving subscription from other nodes channel
	// ctx = context.WithValue(ctx, constants.PublishedSubChId, &channelpool.PublishedSubC)

	ctx = context.WithValue(ctx, constants.SQLDB, &sql.SqlDb)

	
	
	

}

func injectPrivateKey(cfg *configs.MainConfiguration, cmd *cobra.Command, storeFilePath string ) configs.MainConfiguration {
	operatorPrivateKey, _ := cmd.Flags().GetString(string(PRIVATE_KEY))
	// if err != nil || len(operatorPrivateKey) == 0 {
	// 	logger.Fatal("operators private_key is required. Use --private-key flag or environment var ML_PRIVATE_KEY")

	// }
	pkFlagLen := len(operatorPrivateKey)
	if pkFlagLen > 0 {
		cfg.PrivateKey = operatorPrivateKey
	}
	
	if len(cfg.PrivateKey) == 0 {
		//check the keystore
		password, _ := cmd.Flags().GetString(string(KEYSTORE_PASSWORD))
		if password == "" {
			password = os.Getenv("ML_KEYSTORE_PASSWORD")
		}
		if password == "" {
			fmt.Println("Enter your keystore password: ")
			inputPass, err := readInputSecurely()
			if err!= nil {
				logger.Fatal("provide a keystore password")
			}
			password = string(inputPass)
		}
		// ksDir, _ := cmd.Flags().GetString(string(KEYSTORE_DIR))
		// if len(ksDir) == 0 {
		// 	ksDir = cfg.KeyStoreDir
		// }
		// if len(ksDir) == 0 {
		// 	ksDir = filepath.Join(cfg.DataDir, "keystores")
		// }
		privKey, err := loadPrivateKeyFromKeyStore(string(password), storeFilePath)
		if err != nil {
			logger.Fatal(err)
		}
		
		if len(privKey) == 0 {
			// logger.Fatal("error loading private key. please check password.")
			logger.Fatal("error loading private key. please check password.")
		}
		cfg.PrivateKey = hex.EncodeToString(privKey)
	}
	
	if  len(cfg.PrivateKey) != 64 {
		logger.Fatal("--private-key must be 32 bytes long")
	}
	

	// conver private key to edd
	pk, err :=  hex.DecodeString(cfg.PrivateKey)
	if err != nil {
		logger.Fatal( err)
	}
	// SECP KEYS
	cfg.PrivateKeySECP = pk
	_, pubKey := btcec.PrivKeyFromBytes(pk)
	cfg.PublicKeySECP = pubKey.SerializeCompressed()
	cfg.PublicKeySECPHex = hex.EncodeToString(cfg.PublicKeySECP)
	
	// EDD KEYS
	cfg.PrivateKeyEDD = ed25519.NewKeyFromSeed(pk)
	cfg.PrivateKey = hex.EncodeToString(cfg.PrivateKeyEDD)
	cfg.PublicKeyEDD = cfg.PrivateKeyEDD[32:]
	cfg.PublicKeyEDDHex = hex.EncodeToString(cfg.PublicKeyEDD)
	// cfg.PublicKey = hex.EncodeToString(cfg.PublicKeyEDD)
	cfg.OperatorAddress = crypto.ToBech32Address(cfg.PublicKeySECP, string(cfg.AddressPrefix))

	return *cfg
}