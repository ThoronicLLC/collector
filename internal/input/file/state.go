package file

import (
	"encoding/json"
	"github.com/ThoronicLLC/collector/pkg/core"
)

type fileState struct {
	Trackers []fileTracker `json:"trackers"`
}

type fileTracker struct {
	FilePath     string `json:"file_path"`
	FilePosition int64  `json:"file_position"`
}

func defaultState() fileState {
	return fileState{Trackers: make([]fileTracker, 0)}
}

func loadState(state core.State) fileState {
	if state == nil {
		return defaultState()
	}

	var loadedState fileState
	err := json.Unmarshal(state, &loadedState)
	if err != nil {
		return defaultState()
	}

	return loadedState
}

func getFilePastStatePosition(path string, state fileState) int64 {
	var position int64 = 0
	for _, v := range state.Trackers {
		if v.FilePath == path {
			return v.FilePosition
		}
	}
	return position
}

func updateFileState(path string, state fileState, position int64) fileState {
	newTrackers := make([]fileTracker, 0)

	// Add all the previous states
	for _, v := range state.Trackers {
		if v.FilePath != path {
			newTrackers = append(newTrackers, v)
		}
	}

	newTrackers = append(newTrackers, fileTracker{FilePath: path, FilePosition: position})

	return fileState{Trackers: newTrackers}
}
