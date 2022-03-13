package manager

import "time"

func (manager *Manager) statusHandler(err error) {
	if err != nil {
		manager.status = failureStatus(manager, err)
	} else {
		manager.status = successfulStatus(manager)
	}
}

func successfulStatus(manager *Manager) *Status {
	return &Status{
		Running:                  manager.status.Running,
		Errors:                   make([]error, 0),
		LastSuccessfulRun:        time.Now(),
		HasErrors:                false,
		ErrorsSinceSuccessfulRun: 0,
	}
}

func failureStatus(manager *Manager, err error) *Status {
	return &Status{
		Running:                  manager.status.Running,
		Errors:                   append(manager.status.Errors, err),
		LastSuccessfulRun:        manager.status.LastSuccessfulRun,
		HasErrors:                true,
		ErrorsSinceSuccessfulRun: manager.status.ErrorsSinceSuccessfulRun + 1,
	}
}
