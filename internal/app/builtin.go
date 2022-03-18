package app

import (
	"github.com/ThoronicLLC/collector/internal/input/file"
	pubsubin "github.com/ThoronicLLC/collector/internal/input/pubsub"
	"github.com/ThoronicLLC/collector/internal/input/syslog"
	"github.com/ThoronicLLC/collector/internal/output/gcs"
	"github.com/ThoronicLLC/collector/internal/output/log_analytics"
	pubsubout "github.com/ThoronicLLC/collector/internal/output/pubsub"
	"github.com/ThoronicLLC/collector/internal/output/s3"
	"github.com/ThoronicLLC/collector/internal/output/stdout"
	"github.com/ThoronicLLC/collector/internal/processor/cel"
	"github.com/ThoronicLLC/collector/pkg/core"
)

func AddInternalInputs() map[string]core.InputHandler {
	return map[string]core.InputHandler{
		file.InputName:     file.Handler(),
		pubsubin.InputName: pubsubin.Handler(),
		syslog.InputName:   syslog.Handler(),
	}
}

func AddInternalProcessors() map[string]core.ProcessHandler {
	return map[string]core.ProcessHandler{
		cel.ProcessorName: cel.Handler(),
	}
}

func AddInternalOutputs() map[string]core.OutputHandler {
	return map[string]core.OutputHandler{
		stdout.OutputName:        stdout.Handler(),
		s3.OutputName:            s3.Handler(),
		gcs.OutputName:           gcs.Handler(),
		log_analytics.OutputName: log_analytics.Handler(),
		pubsubout.OutputName:     pubsubout.Handler(),
	}
}
