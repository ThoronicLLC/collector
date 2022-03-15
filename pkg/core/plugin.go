package core

import "encoding/json"

type PluginConfig struct {
  Name     string          `json:"name" yaml:"name"`
  Settings json.RawMessage `json:"settings,omitempty" yaml:"settings,omitempty"`
}
