package suggestions

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

// SuggestionChangesExperiment ports
// org.languagetool.rules.spelling.suggestions.SuggestionChangesExperiment
// (name + concrete parameter map after grid expansion).
type SuggestionChangesExperiment struct {
	Name       string
	Parameters map[string]any
}

func (e *SuggestionChangesExperiment) String() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("name=%s,parameters=%v", e.Name, e.Parameters)
}

// Equal ports SuggestionChangesExperiment.equals (name + parameters).
func (e *SuggestionChangesExperiment) Equal(o *SuggestionChangesExperiment) bool {
	if e == o {
		return true
	}
	if e == nil || o == nil {
		return false
	}
	if e.Name != o.Name {
		return false
	}
	return paramsEqual(e.Parameters, o.Parameters)
}

func paramsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || fmt.Sprint(va) != fmt.Sprint(vb) {
			return false
		}
	}
	return true
}

// experimentKey is a map key equivalent to Java equals/hashCode on experiment.
type experimentKey struct {
	name string
	// sorted "k=v;" encoding of parameters
	params string
}

func keyOfExperiment(e *SuggestionChangesExperiment) experimentKey {
	if e == nil {
		return experimentKey{}
	}
	return experimentKey{name: e.Name, params: encodeParams(e.Parameters)}
}

func encodeParams(p map[string]any) string {
	if len(p) == 0 {
		return ""
	}
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprint(p[k]))
		b.WriteByte(';')
	}
	return b.String()
}

// SuggestionChangesExperimentRuns ports the YAML/config grid specification
// (SuggestionChangesExperimentRuns) with parameters as name → value lists.
type SuggestionChangesExperimentRuns struct {
	Name       string
	Parameters map[string][]any // nil/empty → one experiment with empty params
}

// SuggestionChangesDataset ports SuggestionChangesDataset.
type SuggestionChangesDataset struct {
	Name                  string
	Path                  string
	Type                  string // dump | artificial
	SampleRate            float32
	EnforceCorrect        bool
	EnforceAcceptLanguage bool
}

// SuggestionChangesTestConfig ports SuggestionChangesTestConfig.
type SuggestionChangesTestConfig struct {
	NgramLocation string
	Rule          string
	Language      string
	LogDir        string
	// ExperimentRuns ports config.experiments (grid specs).
	ExperimentRuns []SuggestionChangesExperimentRuns
	Datasets       []SuggestionChangesDataset
}

// SuggestionsChanges ports org.languagetool.rules.spelling.suggestions.SuggestionsChanges.
// nil instance (GetSuggestionsChanges()==nil) is the normal production case.
type SuggestionsChanges struct {
	mu sync.Mutex

	config      *SuggestionChangesTestConfig
	experiments []*SuggestionChangesExperiment

	correctSuggestions  map[experimentKey]int
	notFoundSuggestions map[experimentKey]int
	suggestionPosSum    map[experimentKey]int
	textSize            map[experimentKey]int
	computationTime     map[experimentKey]int64
	numSamples          map[experimentKey]int

	// dataset-scoped counters keyed by experimentKey + dataset name
	datasetCorrect          map[datasetKey]int
	datasetNotFound         map[datasetKey]int
	datasetSuggestionPosSum map[datasetKey]int
	datasetNumSamples       map[datasetKey]int
	datasetTextSize         map[datasetKey]int
	datasetComputationTime  map[datasetKey]int64

	currentExperiment *SuggestionChangesExperiment

	// reportWriter optional; when set, BuildReport can write (Java BufferedWriter).
	reportWriter io.Writer
}

type datasetKey struct {
	exp     experimentKey
	dataset string
}

var (
	suggestionsChangesMu sync.Mutex
	suggestionsChanges   *SuggestionsChanges
)

// GetSuggestionsChanges ports SuggestionsChanges.getInstance (nil if not configured).
func GetSuggestionsChanges() *SuggestionsChanges {
	suggestionsChangesMu.Lock()
	defer suggestionsChangesMu.Unlock()
	return suggestionsChanges
}

// InitSuggestionsChanges ports SuggestionsChanges.init(config, reportWriter).
// reportWriter may be nil (Java @Nullable BufferedWriter).
func InitSuggestionsChanges(cfg *SuggestionChangesTestConfig, reportWriter ...io.Writer) *SuggestionsChanges {
	suggestionsChangesMu.Lock()
	defer suggestionsChangesMu.Unlock()
	var w io.Writer
	if len(reportWriter) > 0 {
		w = reportWriter[0]
	}
	if cfg == nil {
		cfg = &SuggestionChangesTestConfig{}
	}
	s := &SuggestionsChanges{
		config:                  cfg,
		experiments:             generateExperiments(cfg.ExperimentRuns),
		correctSuggestions:      map[experimentKey]int{},
		notFoundSuggestions:     map[experimentKey]int{},
		suggestionPosSum:        map[experimentKey]int{},
		textSize:                map[experimentKey]int{},
		computationTime:         map[experimentKey]int64{},
		numSamples:              map[experimentKey]int{},
		datasetCorrect:          map[datasetKey]int{},
		datasetNotFound:         map[datasetKey]int{},
		datasetSuggestionPosSum: map[datasetKey]int{},
		datasetNumSamples:       map[datasetKey]int{},
		datasetTextSize:         map[datasetKey]int{},
		datasetComputationTime:  map[datasetKey]int64{},
		reportWriter:            w,
	}
	suggestionsChanges = s
	return s
}

// ResetSuggestionsChanges clears the singleton (tests; no Java twin).
func ResetSuggestionsChanges() {
	suggestionsChangesMu.Lock()
	defer suggestionsChangesMu.Unlock()
	suggestionsChanges = nil
}

// GetConfig ports getConfig.
func (s *SuggestionsChanges) GetConfig() *SuggestionChangesTestConfig {
	if s == nil {
		return nil
	}
	return s.config
}

// GetExperiments ports getExperiments (expanded grid).
func (s *SuggestionsChanges) GetExperiments() []*SuggestionChangesExperiment {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*SuggestionChangesExperiment, len(s.experiments))
	copy(out, s.experiments)
	return out
}

func (s *SuggestionsChanges) SetCurrentExperiment(e *SuggestionChangesExperiment) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentExperiment = e
}

func (s *SuggestionsChanges) GetCurrentExperiment() *SuggestionChangesExperiment {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentExperiment
}

// IsRunningExperiment ports SuggestionsChanges.isRunningExperiment.
func IsRunningExperiment(name string) bool {
	s := GetSuggestionsChanges()
	if s == nil {
		return false
	}
	e := s.GetCurrentExperiment()
	return e != nil && e.Name == name
}

// TrackExperimentResult ports trackExperimentResult(Pair<source>, position, textSize, time).
// position 0 → correct; -1 → not found; else accumulate pos sum.
func (s *SuggestionsChanges) TrackExperimentResult(
	experiment *SuggestionChangesExperiment,
	dataset *SuggestionChangesDataset,
	position int,
	resultTextSize int,
	resultComputationTime int64,
) {
	if s == nil || experiment == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	ek := keyOfExperiment(experiment)
	s.numSamples[ek]++
	s.textSize[ek] += resultTextSize
	s.computationTime[ek] += resultComputationTime

	var dk datasetKey
	if dataset != nil {
		dk = datasetKey{exp: ek, dataset: dataset.Name}
		s.datasetNumSamples[dk]++
		s.datasetTextSize[dk] += resultTextSize
		s.datasetComputationTime[dk] += resultComputationTime
	}

	if position == 0 {
		s.correctSuggestions[ek]++
		if dataset != nil {
			s.datasetCorrect[dk]++
		}
	}
	if position == -1 {
		s.notFoundSuggestions[ek]++
		if dataset != nil {
			s.datasetNotFound[dk]++
		}
	} else {
		// Java: else branch covers position >= 0 (including 0)
		s.suggestionPosSum[ek] += position
		if dataset != nil {
			s.datasetSuggestionPosSum[dk] += position
		}
	}
}

// gridsearch ports SuggestionsChanges.gridsearch (TreeMap lastKey peel).
// grid keys must be processed highest-first (Java SortedMap.lastKey).
func gridsearch(grid map[string][]any, current []map[string]any) []map[string]any {
	if len(grid) == 0 {
		return current
	}
	// lastKey = highest alphabetical (TreeMap)
	keys := make([]string, 0, len(grid))
	for k := range grid {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	name := keys[len(keys)-1]
	params := grid[name]

	var result []map[string]any
	if len(current) == 0 {
		for _, value := range params {
			result = append(result, map[string]any{name: value})
		}
	} else {
		for _, entry := range current {
			for _, value := range params {
				modified := make(map[string]any, len(entry)+1)
				for k, v := range entry {
					modified[k] = v
				}
				modified[name] = value
				result = append(result, modified)
			}
		}
	}
	// headMap(name) — all keys strictly less than name
	head := make(map[string][]any, len(grid)-1)
	for _, k := range keys[:len(keys)-1] {
		head[k] = grid[k]
	}
	return gridsearch(head, result)
}

// generateExperiments ports SuggestionsChanges.generateExperiments.
func generateExperiments(specs []SuggestionChangesExperimentRuns) []*SuggestionChangesExperiment {
	var experiments []*SuggestionChangesExperiment
	for _, spec := range specs {
		if len(spec.Parameters) == 0 {
			experiments = append(experiments, &SuggestionChangesExperiment{
				Name:       spec.Name,
				Parameters: map[string]any{},
			})
			continue
		}
		// Java: SortedMap = new TreeMap<>(spec.parameters)
		grid := make(map[string][]any, len(spec.Parameters))
		for k, v := range spec.Parameters {
			grid[k] = v
		}
		combinations := gridsearch(grid, nil)
		for _, settings := range combinations {
			// copy settings map
			params := make(map[string]any, len(settings))
			for k, v := range settings {
				params[k] = v
			}
			experiments = append(experiments, &SuggestionChangesExperiment{
				Name:       spec.Name,
				Parameters: params,
			})
		}
	}
	return experiments
}

// BuildReport ports SuggestionsChanges.Report.run text (overall + per-dataset).
// If a reportWriter was set at init, also writes there.
func (s *SuggestionsChanges) BuildReport() string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var report strings.Builder
	report.WriteString("Overall report:\n\n")

	var best *SuggestionChangesExperiment
	bestID := -1
	bestAccuracy := 0.0

	for experimentID, experiment := range s.experiments {
		id := experimentID + 1
		ek := keyOfExperiment(experiment)
		correct := s.correctSuggestions[ek]
		score := s.suggestionPosSum[ek]
		notFound := s.notFoundSuggestions[ek]
		total := s.numSamples[ek]
		accuracy := 0.0
		if total > 0 {
			accuracy = float64(correct) / float64(total) * 100.0
		}
		speed := 0.0
		if ct := s.computationTime[ek]; ct > 0 {
			speed = float64(s.textSize[ek]) / float64(ct) * 1000.0
		}
		if accuracy > bestAccuracy {
			best = experiment
			bestAccuracy = accuracy
			bestID = id
		}
		report.WriteString(fmt.Sprintf(
			"Experiment #%d (%s): %d / %d correct suggestions -> %f%% accuracy; score (less = better): %d; not found: %d; processed %f chars/second.\n",
			id, experiment, correct, total, accuracy, score, notFound, speed))
	}
	report.WriteString(fmt.Sprintf("\nBest experiment: #%d (%s) @ %f%% accuracy\n", bestID, best, bestAccuracy))

	if s.config != nil {
		for _, dataset := range s.config.Datasets {
			report.WriteString(fmt.Sprintf("\nReport for dataset: %s\n", dataset.Name))
			best = nil
			bestAccuracy = 0
			bestID = -1
			for experimentID, experiment := range s.experiments {
				id := experimentID + 1
				ek := keyOfExperiment(experiment)
				dk := datasetKey{exp: ek, dataset: dataset.Name}
				correct := s.datasetCorrect[dk]
				score := s.datasetSuggestionPosSum[dk]
				notFound := s.datasetNotFound[dk]
				total := s.datasetNumSamples[dk]
				accuracy := 0.0
				if total > 0 {
					accuracy = float64(correct) / float64(total) * 100.0
				}
				speed := 0.0
				if ct := s.datasetComputationTime[dk]; ct > 0 {
					speed = float64(s.datasetTextSize[dk]) / float64(ct) * 1000.0
				}
				if accuracy > bestAccuracy {
					best = experiment
					bestAccuracy = accuracy
					bestID = id
				}
				report.WriteString(fmt.Sprintf(
					"Experiment #%d (%s): %d / %d correct suggestions-> %f%% accuracy; score (less = better): %d; not found: %d; processed %f chars/second.\n",
					id, experiment, correct, total, accuracy, score, notFound, speed))
			}
			report.WriteString(fmt.Sprintf("\nBest experiment: #%d (%s) @ %f%% accuracy\n", bestID, best, bestAccuracy))
		}
	}

	out := report.String()
	if s.reportWriter != nil {
		_, _ = io.WriteString(s.reportWriter, out)
	}
	return out
}

// --- Convenience counters (test helpers; map to TrackExperimentResult semantics) ---

// RecordCorrect increments correct + samples for the current experiment (position 0).
func (s *SuggestionsChanges) RecordCorrect() {
	if s == nil {
		return
	}
	s.TrackExperimentResult(s.GetCurrentExperiment(), nil, 0, 0, 0)
}

// RecordNotFound increments not-found + samples (position -1).
func (s *SuggestionsChanges) RecordNotFound() {
	if s == nil {
		return
	}
	s.TrackExperimentResult(s.GetCurrentExperiment(), nil, -1, 0, 0)
}

// RecordSuggestionPos tracks a non-zero rank position (still increments samples).
func (s *SuggestionsChanges) RecordSuggestionPos(pos int) {
	if s == nil {
		return
	}
	if pos == 0 {
		s.RecordCorrect()
		return
	}
	if pos < 0 {
		s.RecordNotFound()
		return
	}
	s.TrackExperimentResult(s.GetCurrentExperiment(), nil, pos, 0, 0)
}

// CorrectCount returns correctSuggestions for the named experiment (any params).
// Prefer TrackExperimentResult + GetExperiments for multi-param grids.
func (s *SuggestionsChanges) CorrectCount(e *SuggestionChangesExperiment) int {
	if s == nil || e == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.correctSuggestions[keyOfExperiment(e)]
}

func (s *SuggestionsChanges) NotFoundCount(e *SuggestionChangesExperiment) int {
	if s == nil || e == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.notFoundSuggestions[keyOfExperiment(e)]
}

func (s *SuggestionsChanges) PosSum(e *SuggestionChangesExperiment) int {
	if s == nil || e == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.suggestionPosSum[keyOfExperiment(e)]
}

func (s *SuggestionsChanges) NumSamples(e *SuggestionChangesExperiment) int {
	if s == nil || e == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.numSamples[keyOfExperiment(e)]
}
