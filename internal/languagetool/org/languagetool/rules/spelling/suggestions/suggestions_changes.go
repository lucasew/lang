package suggestions

import (
	"sync"
)

// SuggestionChangesExperiment ports experiment identity + free-form parameters.
type SuggestionChangesExperiment struct {
	Name       string
	Parameters map[string]any
}

// SuggestionChangesDataset describes a corpus used in experiment runs.
type SuggestionChangesDataset struct {
	Name                  string
	Path                  string
	Type                  string // dump | artificial
	SampleRate            float32
	EnforceCorrect        bool
	EnforceAcceptLanguage bool
}

// SuggestionChangesTestConfig ports the test harness configuration surface.
type SuggestionChangesTestConfig struct {
	NgramLocation string
	Rule          string
	Language      string
	LogDir        string
	Experiments   []SuggestionChangesExperiment
	Datasets      []SuggestionChangesDataset
}

// SuggestionsChanges ports org.languagetool.rules.spelling.suggestions.SuggestionsChanges
// as a process-wide experiment tracker (nil when not configured).
type SuggestionsChanges struct {
	mu                sync.Mutex
	Config            *SuggestionChangesTestConfig
	CurrentExperiment *SuggestionChangesExperiment
	Correct           map[string]int
	NotFound          map[string]int
	PosSum            map[string]int
	NumSamples        map[string]int
}

var (
	suggestionsChangesMu sync.Mutex
	suggestionsChanges   *SuggestionsChanges
)

// GetSuggestionsChanges returns the singleton (nil if not initialized).
func GetSuggestionsChanges() *SuggestionsChanges {
	suggestionsChangesMu.Lock()
	defer suggestionsChangesMu.Unlock()
	return suggestionsChanges
}

// InitSuggestionsChanges installs a process-wide tracker.
func InitSuggestionsChanges(cfg *SuggestionChangesTestConfig) *SuggestionsChanges {
	suggestionsChangesMu.Lock()
	defer suggestionsChangesMu.Unlock()
	suggestionsChanges = &SuggestionsChanges{
		Config:     cfg,
		Correct:    map[string]int{},
		NotFound:   map[string]int{},
		PosSum:     map[string]int{},
		NumSamples: map[string]int{},
	}
	return suggestionsChanges
}

// ResetSuggestionsChanges clears the singleton (tests).
func ResetSuggestionsChanges() {
	suggestionsChangesMu.Lock()
	defer suggestionsChangesMu.Unlock()
	suggestionsChanges = nil
}

func (s *SuggestionsChanges) SetCurrentExperiment(e *SuggestionChangesExperiment) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentExperiment = e
}

func (s *SuggestionsChanges) GetCurrentExperiment() *SuggestionChangesExperiment {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.CurrentExperiment
}

// IsRunningExperiment reports whether the named experiment is current.
func IsRunningExperiment(name string) bool {
	s := GetSuggestionsChanges()
	if s == nil {
		return false
	}
	e := s.GetCurrentExperiment()
	return e != nil && e.Name == name
}

// RecordCorrect increments the correct-suggestion counter for the current experiment.
func (s *SuggestionsChanges) RecordCorrect() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentExperiment == nil {
		return
	}
	s.Correct[s.CurrentExperiment.Name]++
	s.NumSamples[s.CurrentExperiment.Name]++
}

// RecordNotFound increments the not-found counter for the current experiment.
func (s *SuggestionsChanges) RecordNotFound() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentExperiment == nil {
		return
	}
	s.NotFound[s.CurrentExperiment.Name]++
	s.NumSamples[s.CurrentExperiment.Name]++
}

// RecordSuggestionPos adds a ranked position (0-based) for the current experiment.
func (s *SuggestionsChanges) RecordSuggestionPos(pos int) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentExperiment == nil {
		return
	}
	s.PosSum[s.CurrentExperiment.Name] += pos
	s.NumSamples[s.CurrentExperiment.Name]++
}
