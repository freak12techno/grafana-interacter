package silence_manager

import (
	"errors"
	"main/pkg/types"
)

type StubSilenceManager struct {
	GetSilencesError              error
	GetSilenceError               error
	GetSilenceMatchingAlertsError error
	CreateSilenceError            error

	Disabled bool

	Silences map[string]types.Silence
}

func NewStubSilenceManager() *StubSilenceManager {
	return &StubSilenceManager{
		Silences: make(map[string]types.Silence),
	}
}

func (m *StubSilenceManager) GetSilences() (types.Silences, error) {
	if m.GetSilencesError != nil {
		return nil, m.GetSilencesError
	}

	silences := make([]types.Silence, len(m.Silences))
	index := 0

	for _, silence := range m.Silences {
		silences[index] = silence
		index++
	}

	return silences, nil
}

func (m *StubSilenceManager) GetSilence(silenceID string) (types.Silence, error) {
	if m.GetSilenceError != nil {
		return types.Silence{}, m.GetSilenceError
	}

	if silence, ok := m.Silences[silenceID]; ok {
		return silence, nil
	}

	return types.Silence{}, errors.New("Silence was not found!")
}

func (m *StubSilenceManager) CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error) {
	if m.CreateSilenceError != nil {
		return types.SilenceCreateResponse{}, m.CreateSilenceError
	}

	m.Silences[silence.ID] = silence

	return types.SilenceCreateResponse{SilenceID: silence.ID}, nil
}

func (m *StubSilenceManager) GetSilenceMatchingAlerts(silence types.Silence) ([]types.AlertmanagerAlert, error) {
	if m.GetSilenceMatchingAlertsError != nil {
		return nil, m.GetSilenceMatchingAlertsError
	}

	return []types.AlertmanagerAlert{}, nil
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

func (m *StubSilenceManager) GetPaginatedSilencesListPrefix() string {
	return "stub_paginate_silences"
}

func (m *StubSilenceManager) GetPrepareSilencePrefix() string {
	return "stub_prepare_silence"
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
