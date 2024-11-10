package silence_manager

import (
	"fmt"
	"main/pkg/types"
	"main/pkg/utils/generic"
	"sync"
	"time"
)

type Prefixes struct {
	PaginatedSilencesList string
	Silence               string
	PrepareSilence        string
	Unsilence             string
	ListSilencesCommand   string
}

type SilenceManager interface {
	GetSilences() (types.Silences, error)
	GetSilence(silenceID string) (types.Silence, error)
	CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error)
	GetSilenceMatchingAlerts(silence types.Silence) ([]types.AlertmanagerAlert, error)
	DeleteSilence(silenceID string) error
	Prefixes() Prefixes
	Name() string
	Enabled() bool
	GetMutesDurations() []string
}

func GetSilencesWithAlerts(manager SilenceManager) ([]types.SilenceWithAlerts, error) {
	silences, err := manager.GetSilences()
	if err != nil {
		return []types.SilenceWithAlerts{}, err
	}

	silences = generic.Filter(silences, func(s types.Silence) bool {
		return s.EndsAt.After(time.Now())
	})

	silencesWithAlerts := make([]types.SilenceWithAlerts, len(silences))

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errs []error

	for index, silence := range silences {
		wg.Add(1)
		go func(index int, silence types.Silence) {
			defer wg.Done()

			alerts, alertsErr := manager.GetSilenceMatchingAlerts(silence)
			if alertsErr != nil {
				mutex.Lock()
				errs = append(errs, alertsErr)
				mutex.Unlock()
				return
			}

			mutex.Lock()
			silencesWithAlerts[index] = types.SilenceWithAlerts{
				Silence:       silence,
				AlertsPresent: true,
				Alerts:        alerts,
			}
			mutex.Unlock()
		}(index, silence)
	}

	wg.Wait()

	if len(errs) > 0 {
		return []types.SilenceWithAlerts{}, fmt.Errorf("Error getting alerts for silence on %d silences!", len(errs))
	}

	return silencesWithAlerts, nil
}
