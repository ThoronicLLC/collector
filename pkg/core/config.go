package core

type Config struct {
	Input      PluginConfig   `json:"input" yaml:"input"`
	Processors []PluginConfig `json:"processors" yaml:"processors"`
	Outputs    []PluginConfig `json:"outputs" yaml:"outputs"`
}
