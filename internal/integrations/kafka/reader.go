package kafka

import (
  "context"
  "fmt"
  "github.com/segmentio/kafka-go"
  "time"
)

type Reader struct {
  reader *kafka.Reader
  ctx    context.Context
}

type ReaderConfig struct {
  Ctx        context.Context
  AuthConfig AuthConfig
  Brokers    []string
  Topic      string
  GroupID    string
  MinBytes   int
  MaxBytes   int
}

// NewReader creates a new kafka reader client
func NewReader(kConf ReaderConfig) (*Reader, error) {
  // Configure the auth mechanism based on the supplied config
  mechanism, err := newMechanism(kConf.AuthConfig)
  if err != nil {
    return nil, fmt.Errorf("newMechanism(): %w", err)
  }

  // Create a read dialer with the configured mechanism
  readDialer := &kafka.Dialer{
    Timeout:       10 * time.Second,
    DualStack:     true,
    SASLMechanism: mechanism,
  }

  // Initialize the reader with the broker addresses, topic, groupID, and dialer
  r := kafka.NewReader(kafka.ReaderConfig{
    Brokers:  kConf.Brokers,
    Topic:    kConf.Topic,
    GroupID:  kConf.GroupID,
    MinBytes: kConf.MinBytes, // 10e3 10KB
    MaxBytes: kConf.MaxBytes, // 10e6 = 10MB
    Dialer:   readDialer,
  })

  return &Reader{
    reader: r,
    ctx:    kConf.Ctx,
  }, nil
}

// ReadMessage reads a message from the kafka topic
func (k *Reader) ReadMessage() (kafka.Message, error) {
  return k.reader.ReadMessage(k.ctx)
}

// Close closes the reader
func (k *Reader) Close() error {
  err := k.reader.Close()
  if err != nil {
    return fmt.Errorf("reader.Close(): %w", err)
  }

  return nil
}
