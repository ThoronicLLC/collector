package app

import (
	"github.com/ThoronicLLC/collector/internal/input/file"
	"github.com/ThoronicLLC/collector/internal/output/stdout"
	"github.com/ThoronicLLC/collector/internal/processor/cel"
	"github.com/ThoronicLLC/collector/pkg/core"
)

func AddInternalInputs() map[string]core.InputHandler {
	return map[string]core.InputHandler{
		"file": file.Handler(),
	}
}

func AddInternalProcessors() map[string]core.ProcessHandler {
	return map[string]core.ProcessHandler{
		"cel": cel.Handler(),
	}
}

func AddInternalOutputs() map[string]core.OutputHandler {
	return map[string]core.OutputHandler{
		"stdout": stdout.Handler(),
	}
}
