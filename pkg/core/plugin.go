package core

type PluginConfig struct {
	Name     string `json:"name" yaml:"name"`
	Settings []byte `json:"settings,omitempty" yaml:"settings,omitempty"`
}
