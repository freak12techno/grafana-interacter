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
	return types.Silence{}, errors.New("Silence was not found!")
}

func (m *StubSilenceManager) CreateSilence(silence types.Silence) (types.SilenceCreateResponse, error) {
	m.Silences[silence.ID] = silence
	return types.SilenceCreateResponse{SilenceID: silence.ID}, nil
}

func (m *StubSilenceManager) GetMatchingAlerts(matchers types.SilenceMatchers) ([]types.AlertmanagerAlert, error) {
	if m.GetSilenceMatchingAlertsError != nil {
		return nil, m.GetSilenceMatchingAlertsError
	}

	return []types.AlertmanagerAlert{}, nil
}

func (m *StubSilenceManager) DeleteSilence(silenceID string) error {
	return nil
}

func (m *StubSilenceManager) Prefixes() Prefixes {
	return Prefixes{
		Silence:               "stub_silence",
		Unsilence:             "stub_unsilence",
		PaginatedSilencesList: "stub_paginated_silences_list",
		PrepareSilence:        "stub_prepare_silence",
	}
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
