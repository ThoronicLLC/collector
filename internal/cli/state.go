package cli

import (
  "fmt"
  "github.com/ThoronicLLC/collector/pkg/core"
  log "github.com/sirupsen/logrus"
  "os"
  "path/filepath"
)

type fileState struct {
  dirPath string
}

func NewFileState(path string) *fileState {
  return &fileState{dirPath: path}
}

func (s *fileState) Save(id string, state core.State) error {
  statePath := filepath.Join(s.dirPath, fmt.Sprintf("%s.state", id))
  err := os.WriteFile(statePath, state, 0644)
  if err != nil {
    return fmt.Errorf("issue writing state file: %s", err)
  }
  return nil
}

func (s *fileState) Load(id string) core.State {
  statePath := filepath.Join(s.dirPath, fmt.Sprintf("%s.state", id))
  content, err := os.ReadFile(statePath)
  if err != nil {
    log.Errorf("issue reading state file: %s", err)
    return nil
  }

  return content
}

func defaultSaveStateFunc(stateManager *fileState) core.SaveStateFunc {
  return func(id string, state core.State) error {
    return stateManager.Save(id, state)
  }
}

func defaultLoadStateFunc(stateManager *fileState) core.LoadStateFunc {
  return func(id string) core.State {
    return stateManager.Load(id)
  }
}
