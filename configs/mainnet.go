package configs

var MainNetConfig = MainConfiguration{
	AddressPrefix:   "mldev",
	ChainId:         "8453",
	ProtocolVersion: "/mlayer/1.0.0",
	LogLevel:        "info",
	DataDir:         "data/8453/",
	ListenerAdresses: []string{
		"/ip4/0.0.0.0/udp/5002/quic-v1",
		"/ip4/0.0.0.0/udp/5002/quic-v1/webtransport",
		"/ip4/0.0.0.0/tcp/6001",
	},
	BootstrapPeers: []string{
		
	},
	BootstrapNode: false,
	EvmRpcConfig: map[string]EthConfig{
		"8453": { 
			Name:                  "base-sepolia",
			Http:                  "https://snowy-multi-liquid.base-sepolia.quiknode.pro/7dac4517f70845dc1d5ee3ffe539fa43352fce9f/",
			Wss:                   "wss://snowy-multi-liquid.base-sepolia.quiknode.pro/7dac4517f70845dc1d5ee3ffe539fa43352fce9f/",
			TokenContract:         "0xEdC160695971977326Ff10f285a6cd7dA6B2186c",
			XTokenContract:        "0xBf58C54DA1c778D3f77c47332C1554bda1D95ea0",
			ChainInfoContract:     "0x7b45C5Bf6b4f27E9ac0F9a6907656c2BE342c16F",
			SentryNodeContract:    "0x9856c3B8d03937862C57b2330aF088684CA196c1",
			ValidatorNodeContract: "0x58E549288E64e4A1bcF80aeCfa3bb002E6C4742b",
			ApplicationContract:        "0x331bd4973dAC41F20aAB98856bB2cF3b691419a6",
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
