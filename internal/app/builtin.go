package app

import (
	"github.com/ThoronicLLC/collector/internal/input/file"
	"github.com/ThoronicLLC/collector/internal/output/stdout"
	"github.com/ThoronicLLC/collector/pkg/core"
)

func AddInternalInputs() map[string]core.InputHandler {
	return map[string]core.InputHandler{
		"file": file.Handler(),
	}
}

func AddInternalProcessors() map[string]core.ProcessHandler {
	return make(map[string]core.ProcessHandler, 0)
}

func AddInternalOutputs() map[string]core.OutputHandler {
	return map[string]core.OutputHandler{
		"stdout": stdout.Handler(),
	}
}
