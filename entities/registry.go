package entities

type CapabilityType string

const (
	NLP = "nlp"
	ComputerVision = "computerVision"
)

type CapabilityScopeType interface{}




type Capability struct {
	CapabilityType CapabilityType  `json:"capabilityType,omitempty"`
	Scope  any `json:"scope,omitempty"`
}


// type GPURequirements struct {
// 	MinimumVRAM    string   `json:"minimumVRAM"`
// 	Architectures  []string `json:"architectures"`
// }

//	type ResourceRequirements struct {
//		MinimumMemory     string          `json:"minimumMemory"`
//		RecommendedMemory string          `json:"recommendedMemory"`
//		MinimumCPU        string          `json:"minimumCPU"`
//		RecommendedCPU    string          `json:"recommendedCPU"`
//		GPUSupport        bool            `json:"gpuSupport"`
//		GPURequirements   *GPURequirements `json:"gpuRequirements,omitempty"`
type ServiceType string

// }
const (
	AccountService      ServiceType = "account"
	OrganizationService  ServiceType = "organization"
	AIService            ServiceType = "ai"
	DeviceService        ServiceType = "device"
	SmartContractService ServiceType = "smart_contract"
	ApiService          ServiceType = "api"
	BotService         ServiceType = "bots"
	FeedService         ServiceType = "data_feeds"
)

type DataEncryption struct {
	AtRest    bool   `json:"atRest"`
	InTransit bool   `json:"inTransit"`
	Standard  string `json:"standard"`
}

type AccessControls struct {
	Type        string   `json:"type"`
	Permissions []string `json:"permissions"`
}

type SecurityProfile struct {
	AuthenticationRequired bool           `json:"authenticationRequired"`
	AuthMethods            []string       `json:"authMethods"`
	DataEncryption         DataEncryption `json:"dataEncryption"`
	AccessControls         AccessControls `json:"accessControls"`
	AuditLogging           bool           `json:"auditLogging"`
}

type Key struct {
	KeyType     string `json:"keyType"`
	Key         []byte `json:"key"`
	Description string `json:"description"`
}
type Keys struct {
	IdentityKey Key   `json:"idKeys"`
	AppKeys  []Key `json:"appKeys"`
}

type Creator struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Contact      string `json:"contact"`
}

type ServiceMetadata struct {
	ExternalID     string    `json:"externalId"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Creator     Creator   `json:"creator"`
	License     string    `json:"license"`
	Created     uint64 `json:"created"`
	LastUpdated uint64 `json:"lastUpdated"`
	Status      string    `json:"status"` // active, deprecated, testing, maintenance
}

type LatencyMetrics struct {
	Average float64 `json:"average"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
}

type ThroughputMetrics struct {
	RequestsPerSecond     int `json:"requestsPerSecond"`
	MaxConcurrentRequests int `json:"maxConcurrentRequests"`
}

type ReliabilityMetrics struct {
	Uptime    float64 `json:"uptime"`
	ErrorRate float64 `json:"errorRate"`
	MTTR      float64 `json:"mttr"` // Mean Time To Recovery
}

type PerformanceMetrics struct {
	Latency     LatencyMetrics     `json:"latency"`
	Throughput  ThroughputMetrics  `json:"throughput"`
	Reliability ReliabilityMetrics `json:"reliability"`
}

type RateLimit struct {
	RequestsPerMinute int `json:"requestsPerMinute"`
	BurstLimit        int `json:"burstLimit"`
}

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIEndpoint struct {
	URL            string      `json:"url"`
	Method         string      `json:"method"` // GET, POST, PUT, DELETE
	Authentication string      `json:"authentication"`
	RateLimit      RateLimit   `json:"rateLimit"`
	Parameters     []Parameter `json:"parameters"`
}

type API struct {
	BaseURL       string        `json:"baseURL"`
	Version       string        `json:"version"`
	Endpoints     []APIEndpoint `json:"endpoints"`
	Documentation string        `json:"documentation"`
}

type PricingRate struct {
	Base float64 `json:"base"`
	Unit string  `json:"unit"`
}

type PricingTier struct {
	Name  string  `json:"name"`
	Limit int     `json:"limit"`
	Price float64 `json:"price"`
}

type Pricing struct {
	Model          string        `json:"model"` // free, subscription, pay-per-use
	Rates          *PricingRate  `json:"rates,omitempty"`
	Tiers          []PricingTier `json:"tiers,omitempty"`
	CurrencySymbol string        `json:"currency,omitempty"`       // eg. token symbol or currency symbol
	CurrencyId     string        `json:"currencyId,omitempty"`     // eg. contract address or iso symbol
	CurrencyDomain string        `json:"currencyDomain,omitempty"` // e.g. ethereum, solana, base, usa, mexico
	CurrencyType   string        `json:"CurrencyType,omitempty"`   // fiat or crypto
}

type DataPrivacy struct {
	GDPR  bool `json:"gdpr"`
	CCPA  bool `json:"ccpa"`
	HIPAA bool `json:"hipaa"`
}

type Compliance struct {
	Certifications []string    `json:"certifications"`
	DataPrivacy    DataPrivacy `json:"dataPrivacy"`
	AuditReports   []string    `json:"auditReports"`
}

type Dependencies struct {
	Required  []string `json:"required"`
	Optional  []string `json:"optional"`
	Conflicts []string `json:"conflicts"`
}

type Dataset struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    string `json:"size"`
	Type    string `json:"type"`
}

type EvaluationMetric struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Description string  `json:"description"`
}

type Training struct {
	Dataset           Dataset            `json:"dataset"`
	Method            string             `json:"method"`
	EvaluationMetrics []EvaluationMetric `json:"evaluationMetrics"`
}

// type AdminAuthority struct {
// 	ExternalID string `json:"externalId"`
// 	AuthorizedAccount string `json:"authorizedAccount"`
// 	ServiceAccount  string `json:"entityAccount"`
// 	Timestamp uint64 `json:"ts"`
// 	Signature json.RawMessage `json:"sign"`
// }

// func (aa *AdminAuthority) EncodeBytes() ([]byte, error)  {
// 	return encoder.EncodeBytes(
// 		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: aa.AuthorizedAccount},
// 		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: aa.ExternalID},
// 		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: aa.Timestamp},
// 	)
// }

type AgentTopic struct {
	ID string `json:"id"`
	SampleMessage string  `json:"sample_message"`
	Public bool  `json:"public"`
}

type RegistryService struct {
	// REQUIRED
	 Version      float32           `json:"_v"`
	
	// Reference string `json:"ref"`
	Id string `json:"_id,omitempty"`
	Owner AccountString `json:"account,omitempty"`
	ServiceType   ServiceType        `json:"_t"`
	//  AdminAuthority   AdminAuthority        `json:"adminAuthority"`
	Metadata     *ServiceMetadata     `json:"metadata"`
	Capabilities []Capability `json:"capabilities,omitempty"`
	// Resources    ResourceRequirements `json:"resources"`
	Security     SecurityProfile    `json:"security"`
	Performance  PerformanceMetrics `json:"performance"`
	API          API                `json:"api"`
	Pricing      Pricing            `json:"pricing"`
	Compliance   Compliance         `json:"compliance"`
	Dependencies Dependencies       `json:"dependencies"`
	Training     Training           `json:"training"`
	InterfaceKeys		[]string		`json:"interfaceKeys"`
	Topics		[]AgentTopic		`json:"topics"`
}
