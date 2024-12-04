package configs

var TestNetConfig = MainConfiguration{
	AddressPrefix:   "mldev",
	ChainId:         "84532",
	ProtocolVersion: "/mlayer/1.0.0",
	LogLevel:        "info",
	DataDir:         "./data/84532/",
	ListenerAdresses: []string{
		"/ip4/0.0.0.0/udp/5002/quic-v1",
		"/ip4/0.0.0.0/udp/5002/quic-v1/webtransport",
		"/ip4/0.0.0.0/tcp/6001",
		"/ip4/0.0.0.0/tcp/7001/ws",
     	"/ip6/::1/tcp/7001/ws",
	},
	BootstrapPeers: []string{
		//  "/ip4/154.12.228.25/tcp/7001/ws/p2p/12D3KooWFipGipTgu1XxtqpV1wUXcosTjK351Yip7Nj32npo68in",
		// "/ip4/154.12.228.25/udp/5002/quic-v1/p2p/12D3KooWFipGipTgu1XxtqpV1wUXcosTjK351Yip7Nj32npo68in",
		// "/ip4/154.12.228.25/tcp/6001/p2p/12D3KooWFipGipTgu1XxtqpV1wUXcosTjK351Yip7Nj32npo68in",
		// "/ip4/154.12.228.25/udp/5002/quic-v1/webtransport/certhash/uEiA-5eiPdqljQ0S7mQhfcjn63eLwp2yvx3IjMlerBP9n2A/certhash/uEiDByOLXkNBStJPkb5UE9g9wZKc1l0dCI76IiRTmAV3iAQ/p2p/12D3KooWFipGipTgu1XxtqpV1wUXcosTjK351Yip7Nj32npo68in",
	},
	BootstrapNode: false,
	EvmRpcConfig: map[string]EthConfig{
		"84532": {
			Name:                  "base-sepolia",
			Http:                  "https://snowy-multi-liquid.base-sepolia.quiknode.pro/7dac4517f70845dc1d5ee3ffe539fa43352fce9f",
			Wss:                   "wss://snowy-multi-liquid.base-sepolia.quiknode.pro/7dac4517f70845dc1d5ee3ffe539fa43352fce9f",
			TokenContract:         "0xEdC160695971977326Ff10f285a6cd7dA6B2186c",
			XTokenContract:        "0xBf58C54DA1c778D3f77c47332C1554bda1D95ea0",
			ChainInfoContract:     "0x7b45C5Bf6b4f27E9ac0F9a6907656c2BE342c16F",
			SentryNodeContract:    "0x9856c3B8d03937862C57b2330aF088684CA196c1",
			ValidatorNodeContract: "0x58E549288E64e4A1bcF80aeCfa3bb002E6C4742b",
			SubnetContract:        "0x331bd4973dAC41F20aAB98856bB2cF3b691419a6",
		},
	},
	SQLDB: SqlConfig{
		DbDialect:         "sqlite",
		// DbHost:            "localhost",
		// DbPort:            5432,
		// DbSSLMode:         "enable",
		// DbTimezone:        "America/Chicago",
		// DbDatabase:        "mlayer",
		// DbUser:            "dev2",
		// DbPassword:        "",
		// DbMaxOpenConns:    100,
		// DbMaxIdleConns:    10,
		// DbMaxConnLifetime: 3600,
	},
}
