package types

import (
	"errors"
)

type StubSilenceManager struct {
	GetSilencesError              error
	GetSilenceError               error
	GetSilenceMatchingAlertsError error
	CreateSilenceError            error

	Disabled bool

	Silences map[string]Silence
}

func NewStubSilenceManager() *StubSilenceManager {
	return &StubSilenceManager{
		Silences: make(map[string]Silence),
	}
}

func (m *StubSilenceManager) GetSilences() (Silences, error) {
	if m.GetSilencesError != nil {
		return nil, m.GetSilencesError
	}

	silences := make([]Silence, len(m.Silences))
	index := 0

	for _, silence := range m.Silences {
		silences[index] = silence
		index++
	}

	return silences, nil
}

func (m *StubSilenceManager) GetSilence(silenceID string) (Silence, error) {
	if m.GetSilenceError != nil {
		return Silence{}, m.GetSilenceError
	}

	if silence, ok := m.Silences[silenceID]; ok {
		return silence, nil
	}

	return Silence{}, errors.New("Silence was not found!")
}

func (m *StubSilenceManager) CreateSilence(silence Silence) (SilenceCreateResponse, error) {
	if m.CreateSilenceError != nil {
		return SilenceCreateResponse{}, m.CreateSilenceError
	}

	m.Silences[silence.ID] = silence

	return SilenceCreateResponse{SilenceID: silence.ID}, nil
}

func (m *StubSilenceManager) GetSilenceMatchingAlerts(silence Silence) ([]AlertmanagerAlert, error) {
	if m.GetSilenceMatchingAlertsError != nil {
		return nil, m.GetSilenceMatchingAlertsError
	}

	return []AlertmanagerAlert{}, nil
}

func (m *StubSilenceManager) DeleteSilence(silenceID string) error {
	if _, ok := m.Silences[silenceID]; ok {
		delete(m.Silences, silenceID)
		return nil
	} else {
		return errors.New("Silence was not found!")
	}
}

func (m *StubSilenceManager) GetSilencePrefix() string {
	return "stub_silence"
}

func (m *StubSilenceManager) GetUnsilencePrefix() string {
	return "stub_unsilence"
}

func (m *StubSilenceManager) Name() string {
	return "StubSilenceManager"
}

func (m *StubSilenceManager) Enabled() bool {
	return !m.Disabled
}

func (m *StubSilenceManager) GetMutesDurations() []string {
	return []string{}
}
