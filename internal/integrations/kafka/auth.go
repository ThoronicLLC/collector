package kafka

import (
  "fmt"
  "github.com/ThoronicLLC/collector/internal/integrations/kafka/gssapi"
  "github.com/jcmturner/gokrb5/v8/client"
  "github.com/jcmturner/gokrb5/v8/config"
  kt "github.com/jcmturner/gokrb5/v8/keytab"
  "github.com/segmentio/kafka-go/sasl"
  "github.com/segmentio/kafka-go/sasl/scram"
)

type AuthConfig struct {
  ScramSha256    AuthScramSha256Config    `json:"scram_sha_256"`
  ScramSha512    AuthScramSha512Config    `json:"scram_sha_512"`
  GssApiKeytab   AuthGssApiKeytabConfig   `json:"gssapi_keytab"`
  GssApiPassword AuthGssApiPasswordConfig `json:"gssapi_password"`
}

// AuthScramSha256Config is the configuration for the SCRAM-SHA-256 authentication mechanism
type AuthScramSha256Config struct {
  Enabled  bool   `json:"enabled"`
  Username string `json:"username" validate:"required_if:Enabled,true"`
  Password string `json:"password" validate:"required_if:Enabled,true"`
}

// AuthScramSha512Config is the configuration for the SCRAM-SHA512 authentication mechanism
type AuthScramSha512Config struct {
  Enabled  bool   `json:"enabled"`
  Username string `json:"username" validate:"required_if:Enabled,true"`
  Password string `json:"password" validate:"required_if:Enabled,true"`
}

// AuthGssApiKeytabConfig is the configuration for GSSAPI with a keytab authentication mechanism
type AuthGssApiKeytabConfig struct {
  Enabled     bool   `json:"enabled"`
  Username    string `json:"username" validate:"required_if:Enabled,true"`
  KeytabFile  string `json:"keytab_file" validate:"required_if:Enabled,true|filePath"`
  Realm       string `json:"realm" validate:"required_if:Enabled,true"`
  ServiceName string `json:"service_name" validate:"required_if:Enabled,true"`
  ConfigFile  string `json:"config_file" validate:"required_if:Enabled,true|filePath"`
}

// AuthGssApiPasswordConfig is the configuration for GSSAPI with password authentication mechanism
type AuthGssApiPasswordConfig struct {
  Enabled     bool   `json:"enabled"`
  Username    string `json:"username" validate:"required_if:Enabled,true"`
  Password    string `json:"password" validate:"required_if:Enabled,true"`
  Realm       string `json:"realm" validate:"required_if:Enabled,true"`
  ServiceName string `json:"service_name" validate:"required_if:Enabled,true"`
  ConfigFile  string `json:"config_file" validate:"required_if:Enabled,true|filePath"`
}

// newMechanism creates a new SASL mechanism based on the configuration
func newMechanism(kConf AuthConfig) (sasl.Mechanism, error) {
  switch {
  case kConf.ScramSha256.Enabled:
    return newMechanismScramSha256(kConf.ScramSha256)
  case kConf.ScramSha512.Enabled:
    return newMechanismScramSha512(kConf.ScramSha512)
  case kConf.GssApiPassword.Enabled:
    return newMechanismGSSAPIWithPassword(kConf.GssApiPassword)
  case kConf.GssApiKeytab.Enabled:
    return newMechanismGSSAPIWithKeytab(kConf.GssApiKeytab)
  }

  return nil, nil
}

// newMechanismGSSAPIWithPassword creates a new GSSAPI mechanism with a password
func newMechanismGSSAPIWithPassword(conf AuthGssApiPasswordConfig) (sasl.Mechanism, error) {
  cfg, err := config.Load(conf.ConfigFile)
  if err != nil {
    return nil, fmt.Errorf("config.Load(): %w", err)
  }

  return gssapi.GoKRB5v8(client.NewWithPassword(conf.Username, conf.Realm, conf.Password, cfg), conf.ServiceName), nil
}

// newMechanismGSSAPIWithKeytab creates a new GSSAPI mechanism with a keytab
func newMechanismGSSAPIWithKeytab(conf AuthGssApiKeytabConfig) (sasl.Mechanism, error) {
  cfg, err := config.Load(conf.ConfigFile)
  if err != nil {
    return nil, fmt.Errorf("config.Load(): %w", err)
  }
  ktFromFile, err := kt.Load(conf.KeytabFile)
  if err != nil {
    return nil, fmt.Errorf("kt.Load(): %w", err)
  }
  return gssapi.GoKRB5v8(client.NewWithKeytab(conf.Username, conf.Realm, ktFromFile, cfg), conf.ServiceName), nil
}

// newMechanismScramSha256 creates a new scram sha256 mechanism
func newMechanismScramSha256(conf AuthScramSha256Config) (sasl.Mechanism, error) {
  return scram.Mechanism(scram.SHA256, conf.Username, conf.Password)
}

// newMechanismScramSha512 creates a new scram sha512 mechanism
func newMechanismScramSha512(conf AuthScramSha512Config) (sasl.Mechanism, error) {
  return scram.Mechanism(scram.SHA512, conf.Username, conf.Password)
}
