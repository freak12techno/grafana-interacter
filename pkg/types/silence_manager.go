package types

import (
	"fmt"
	"main/pkg/utils/generic"
	"sync"
	"time"
)

type SilenceManager interface {
	GetSilences() (Silences, error)
	GetSilenceMatchingAlerts(silence Silence) ([]AlertmanagerAlert, error)
}

func GetSilencesWithAlerts(manager SilenceManager) ([]SilenceWithAlerts, error) {
	silences, err := manager.GetSilences()
	if err != nil {
		return []SilenceWithAlerts{}, err
	}

	silences = generic.Filter(silences, func(s Silence) bool {
		return s.EndsAt.After(time.Now())
	})

	silencesWithAlerts := make([]SilenceWithAlerts, len(silences))

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errs []error

	for index, silence := range silences {
		wg.Add(1)
		go func(index int, silence Silence) {
			defer wg.Done()

			alerts, alertsErr := manager.GetSilenceMatchingAlerts(silence)
			if alertsErr != nil {
				mutex.Lock()
				errs = append(errs, alertsErr)
				mutex.Unlock()
				return
			}

			mutex.Lock()
			silencesWithAlerts[index] = SilenceWithAlerts{
				Silence:       silence,
				AlertsPresent: true,
				Alerts:        alerts,
			}
			mutex.Unlock()
		}(index, silence)
	}

	wg.Wait()

	if len(errs) > 0 {
		return []SilenceWithAlerts{}, fmt.Errorf("Error getting alerts for silence on %d silences!", len(errs))
	}

	return silencesWithAlerts, nil
}
