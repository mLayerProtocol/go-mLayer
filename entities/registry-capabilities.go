package entities

type NLPCapabilities struct {
	Languages         []string `json:"languages"`
	MaxTokens        int      `json:"maxTokens"`
	SupportedTasks   []string `json:"supportedTasks"`
	ModelArchitecture string   `json:"modelArchitecture"`
}

type ComputerVisionCapabilities struct {
	SupportedFormats []string `json:"supportedFormats"`
	MaxResolution   string   `json:"maxResolution"`
	Tasks           []string `json:"tasks"`
	ModelType       string   `json:"modelType"`
}

type AudioProcessingCapabilities struct {
	SupportedFormats []string `json:"supportedFormats"`
	MaxDuration     string   `json:"maxDuration"`
	Features        []string `json:"features"`
}