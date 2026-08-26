package rules

import "sync"

// RemoteRequestResult ports RemoteRuleMetrics.RequestResult.
type RemoteRequestResult string

const (
	RemoteResultSuccess     RemoteRequestResult = "SUCCESS"
	RemoteResultSkipped     RemoteRequestResult = "SKIPPED"
	RemoteResultTimeout     RemoteRequestResult = "TIMEOUT"
	RemoteResultInterrupted RemoteRequestResult = "INTERRUPTED"
	RemoteResultDown        RemoteRequestResult = "DOWN"
	RemoteResultError       RemoteRequestResult = "ERROR"
)

// RemoteRuleMetrics is a lightweight in-process recorder (no Prometheus dependency).
type RemoteRuleMetrics struct {
	mu sync.Mutex
	// RequestCounts[ruleID][result]++
	RequestCounts map[string]map[RemoteRequestResult]int
	// LatencySeconds samples
	LatencySeconds map[string][]float64
	// ThroughputChars samples
	ThroughputChars map[string][]float64
	// WaitSeconds by language
	WaitSeconds map[string][]float64
}

var defaultRemoteMetrics = NewRemoteRuleMetrics()

func NewRemoteRuleMetrics() *RemoteRuleMetrics {
	return &RemoteRuleMetrics{
		RequestCounts:   map[string]map[RemoteRequestResult]int{},
		LatencySeconds:  map[string][]float64{},
		ThroughputChars: map[string][]float64{},
		WaitSeconds:     map[string][]float64{},
	}
}

// DefaultRemoteRuleMetrics returns the process-wide recorder.
func DefaultRemoteRuleMetrics() *RemoteRuleMetrics { return defaultRemoteMetrics }

// RecordRequest ports RemoteRuleMetrics.request.
func (m *RemoteRuleMetrics) RecordRequest(rule string, latencySeconds float64, characters int64, result RemoteRequestResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.RequestCounts[rule] == nil {
		m.RequestCounts[rule] = map[RemoteRequestResult]int{}
	}
	m.RequestCounts[rule][result]++
	m.LatencySeconds[rule] = append(m.LatencySeconds[rule], latencySeconds)
	m.ThroughputChars[rule] = append(m.ThroughputChars[rule], float64(characters))
}

// RecordWait ports RemoteRuleMetrics.wait (milliseconds → seconds stored).
func (m *RemoteRuleMetrics) RecordWait(langCode string, milliseconds int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WaitSeconds[langCode] = append(m.WaitSeconds[langCode], float64(milliseconds)/1000.0)
}

// Request is the package-level convenience matching Java static RemoteRuleMetrics.request.
func RecordRemoteRuleRequest(rule string, startNanos int64, characters int64, result RemoteRequestResult, nowNanos int64) {
	delta := float64(nowNanos-startNanos) / 1e9
	DefaultRemoteRuleMetrics().RecordRequest(rule, delta, characters, result)
}
