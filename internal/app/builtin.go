package app

import (
  "github.com/ThoronicLLC/collector/pkg/core"

  file_input "github.com/ThoronicLLC/collector/internal/input/file"
  kafka_input "github.com/ThoronicLLC/collector/internal/input/kafka"
  msgraph_input "github.com/ThoronicLLC/collector/internal/input/msgraph"
  pubsub_input "github.com/ThoronicLLC/collector/internal/input/pubsub"
  syslog_input "github.com/ThoronicLLC/collector/internal/input/syslog"

  cel_processor "github.com/ThoronicLLC/collector/internal/processor/cel"
  kv_processor "github.com/ThoronicLLC/collector/internal/processor/kv"
  syslog_processor "github.com/ThoronicLLC/collector/internal/processor/syslog"

  file_output "github.com/ThoronicLLC/collector/internal/output/file"
  gcs_output "github.com/ThoronicLLC/collector/internal/output/gcs"
  kafka_output "github.com/ThoronicLLC/collector/internal/output/kafka"
  log_analytics_output "github.com/ThoronicLLC/collector/internal/output/log_analytics"
  pubsub_output "github.com/ThoronicLLC/collector/internal/output/pubsub"
  s3_output "github.com/ThoronicLLC/collector/internal/output/s3"
  stdout_output "github.com/ThoronicLLC/collector/internal/output/stdout"
)

func AddInternalInputs() map[string]core.InputHandler {
  return map[string]core.InputHandler{
    file_input.InputName:    file_input.Handler(),
    kafka_input.InputName:   kafka_input.Handler(),
    pubsub_input.InputName:  pubsub_input.Handler(),
    syslog_input.InputName:  syslog_input.Handler(),
    msgraph_input.InputName: msgraph_input.Handler(),
  }
}

func AddInternalProcessors() map[string]core.ProcessHandler {
  return map[string]core.ProcessHandler{
    cel_processor.ProcessorName:    cel_processor.Handler(),
    syslog_processor.ProcessorName: syslog_processor.Handler(),
    kv_processor.ProcessorName:     kv_processor.Handler(),
  }
}

func AddInternalOutputs() map[string]core.OutputHandler {
  return map[string]core.OutputHandler{
    file_output.OutputName:          file_output.Handler(),
    kafka_output.OutputName:         kafka_output.Handler(),
    stdout_output.OutputName:        stdout_output.Handler(),
    s3_output.OutputName:            s3_output.Handler(),
    gcs_output.OutputName:           gcs_output.Handler(),
    log_analytics_output.OutputName: log_analytics_output.Handler(),
    pubsub_output.OutputName:        pubsub_output.Handler(),
  }
}
