package server

import (
	"sync"
	"sync/atomic"
)

// RequestErrorType ports ServerMetricsCollector.RequestErrorType.
type RequestErrorType string

const (
	RequestErrorQueueFull      RequestErrorType = "queue_full"
	RequestErrorTooManyErrors  RequestErrorType = "too_many_errors"
	RequestErrorMaxCheckTime   RequestErrorType = "max_check_time"
	RequestErrorMaxTextSize    RequestErrorType = "max_text_size"
	RequestErrorInvalidRequest RequestErrorType = "invalid_request"
)

const MetricsUnknown = "unknown"

// ServerMetricsCollector ports metrics recording without Prometheus dependency.
// Counters are process-local and thread-safe.
type ServerMetricsCollector struct {
	checks           atomic.Int64
	matches          atomic.Int64
	characters       atomic.Int64
	httpRequests     atomic.Int64
	failedHealth     atomic.Int64
	computationNanos atomic.Int64

	mu             sync.Mutex
	responsesByCode map[int]int64
	errorsByReason  map[string]int64
	checksByLang    map[string]int64
}

var defaultMetrics = NewServerMetricsCollector()

func NewServerMetricsCollector() *ServerMetricsCollector {
	return &ServerMetricsCollector{
		responsesByCode: map[int]int64{},
		errorsByReason:  map[string]int64{},
		checksByLang:    map[string]int64{},
	}
}

func Metrics() *ServerMetricsCollector { return defaultMetrics }

func (m *ServerMetricsCollector) LogCheck(languageCode string, milliseconds int64, textSize, matchCount int, mode string) {
	if m == nil {
		return
	}
	if languageCode == "" {
		languageCode = MetricsUnknown
	}
	if mode == "" {
		mode = MetricsUnknown
	}
	m.checks.Add(1)
	m.matches.Add(int64(matchCount))
	m.characters.Add(int64(textSize))
	m.computationNanos.Add(milliseconds * 1_000_000)
	m.mu.Lock()
	m.checksByLang[languageCode]++
	m.mu.Unlock()
	_ = mode
}

func (m *ServerMetricsCollector) LogRequestError(t RequestErrorType) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.errorsByReason[string(t)]++
	m.mu.Unlock()
}

func (m *ServerMetricsCollector) LogRequest() {
	if m != nil {
		m.httpRequests.Add(1)
	}
}

func (m *ServerMetricsCollector) LogResponse(httpCode int) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.responsesByCode[httpCode]++
	m.mu.Unlock()
}

func (m *ServerMetricsCollector) LogFailedHealthcheck() {
	if m != nil {
		m.failedHealth.Add(1)
	}
}

func (m *ServerMetricsCollector) Checks() int64 {
	if m == nil {
		return 0
	}
	return m.checks.Load()
}

func (m *ServerMetricsCollector) Matches() int64 {
	if m == nil {
		return 0
	}
	return m.matches.Load()
}

func (m *ServerMetricsCollector) Characters() int64 {
	if m == nil {
		return 0
	}
	return m.characters.Load()
}

func (m *ServerMetricsCollector) HTTPRequests() int64 {
	if m == nil {
		return 0
	}
	return m.httpRequests.Load()
}

func (m *ServerMetricsCollector) ResponseCount(code int) int64 {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.responsesByCode[code]
}

func (m *ServerMetricsCollector) ErrorCount(t RequestErrorType) int64 {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errorsByReason[string(t)]
}

// MetricsSnapshot is a JSON-friendly view of process-local counters.
type MetricsSnapshot struct {
	Checks           int64            `json:"checks"`
	Matches          int64            `json:"matches"`
	Characters       int64            `json:"characters"`
	HTTPRequests     int64            `json:"httpRequests"`
	FailedHealth     int64            `json:"failedHealthchecks"`
	ComputationMs    int64            `json:"computationMs"`
	ResponsesByCode  map[string]int64 `json:"responsesByCode"`
	ErrorsByReason   map[string]int64 `json:"errorsByReason"`
	ChecksByLanguage map[string]int64 `json:"checksByLanguage"`
}

// Snapshot returns a copy of current metrics (soft /v2/metrics surface).
func (m *ServerMetricsCollector) Snapshot() MetricsSnapshot {
	if m == nil {
		return MetricsSnapshot{
			ResponsesByCode:  map[string]int64{},
			ErrorsByReason:   map[string]int64{},
			ChecksByLanguage: map[string]int64{},
		}
	}
	snap := MetricsSnapshot{
		Checks:           m.checks.Load(),
		Matches:          m.matches.Load(),
		Characters:       m.characters.Load(),
		HTTPRequests:     m.httpRequests.Load(),
		FailedHealth:     m.failedHealth.Load(),
		ComputationMs:    m.computationNanos.Load() / 1_000_000,
		ResponsesByCode:  map[string]int64{},
		ErrorsByReason:   map[string]int64{},
		ChecksByLanguage: map[string]int64{},
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for code, n := range m.responsesByCode {
		snap.ResponsesByCode[itoaCode(code)] = n
	}
	for k, n := range m.errorsByReason {
		snap.ErrorsByReason[k] = n
	}
	for k, n := range m.checksByLang {
		snap.ChecksByLanguage[k] = n
	}
	return snap
}

func itoaCode(code int) string {
	// small int to decimal without strconv import cycle concerns
	if code == 0 {
		return "0"
	}
	neg := code < 0
	if neg {
		code = -code
	}
	var b [12]byte
	i := len(b)
	for code > 0 {
		i--
		b[i] = byte('0' + code%10)
		code /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
