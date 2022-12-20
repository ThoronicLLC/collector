package kafka

import (
  "context"
  "fmt"
  "github.com/segmentio/kafka-go"
)

type Writer struct {
  writer *kafka.Writer
  ctx    context.Context
}

type WriterConfig struct {
  Ctx        context.Context
  AuthConfig AuthConfig
  Brokers    []string
  Topic      string
}

// NewWriter creates a new kafka writer client
func NewWriter(kConf WriterConfig) (*Writer, error) {
  // Configure the auth mechanism based on the supplied config
  mechanism, err := newMechanism(kConf.AuthConfig)
  if err != nil {
    return nil, fmt.Errorf("newMechanism(): %w", err)
  }

  // Transports are responsible for managing connection pools and other resources,
  // it's generally best to create a few of these and share them across your
  // application.
  sharedTransport := &kafka.Transport{
    SASL: mechanism,
  }

  // Initialize the writer with the broker addresses, topic, and transport
  w := &kafka.Writer{
    Addr:      kafka.TCP(kConf.Brokers...),
    Topic:     kConf.Topic,
    Balancer:  &kafka.Hash{},
    Transport: sharedTransport,
  }

  return &Writer{
    writer: w,
    ctx:    kConf.Ctx,
  }, nil
}

// WriteMessage writes the message to the kafka topic
func (k *Writer) WriteMessage(message []byte) error {
  return k.writer.WriteMessages(k.ctx, kafka.Message{
    Value: message,
  })
}

// Close closes the writer
func (k *Writer) Close() error {
  err := k.writer.Close()
  if err != nil {
    return fmt.Errorf("writer.Close(): %w", err)
  }

  return nil
}
