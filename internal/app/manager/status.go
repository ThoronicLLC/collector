package manager

import "time"

type StatusHandler func(status Status)

func (manager *Manager) successfulStatus(count int) {
	manager.status = &Status{
		Running:                   manager.status.Running,
		Errors:                    make([]error, 0),
		LastSuccessfulRun:         time.Now(),
		LastSuccessfulResultCount: count,
		HasErrors:                 false,
		ErrorsSinceSuccessfulRun:  0,
	}
}

func (manager *Manager) failureStatus(err error) {
	manager.status = &Status{
		Running:                   manager.status.Running,
		Errors:                    append(manager.status.Errors, err),
		LastSuccessfulRun:         manager.status.LastSuccessfulRun,
		LastSuccessfulResultCount: manager.status.LastSuccessfulResultCount,
		HasErrors:                 true,
		ErrorsSinceSuccessfulRun:  manager.status.ErrorsSinceSuccessfulRun + 1,
	}
}
