package server

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// requestEvent ports RequestLimiter.RequestEvent.
// mode uses CheckMode from text_checker.go (JLanguageTool.Mode twin).
type requestEvent struct {
	ip          string
	date        time.Time
	sizeInBytes int
	fingerprint string
	mode        CheckMode
}

const requestQueueSize = 1000

// RequestLimiter ports org.languagetool.server.RequestLimiter.
type RequestLimiter struct {
	mu                        sync.Mutex
	requestLimit              int
	requestLimitInBytes       int
	requestLimitPeriodSeconds int
	ipFingerprintFactor       int
	ipRequestLimit            int
	ipRequestLimitInBytes     int
	requestEvents             []requestEvent
	// legacy simple IP request-count path (NewRequestLimiter / Allow)
	maxRequests int
	window      time.Duration
	byIP        map[string][]time.Time
}

// NewRequestLimiter is the legacy count-only constructor used by older twins.
func NewRequestLimiter(maxRequests int, windowSeconds int) *RequestLimiter {
	if maxRequests <= 0 {
		maxRequests = 100
	}
	if windowSeconds <= 0 {
		windowSeconds = 60
	}
	return &RequestLimiter{
		maxRequests: maxRequests,
		window:      time.Duration(windowSeconds) * time.Second,
		byIP:        map[string][]time.Time{},
	}
}

// NewRequestLimiterFull ports RequestLimiter(requestLimit, requestLimitInBytes, periodSeconds, ipFingerprintFactor).
func NewRequestLimiterFull(requestLimit, requestLimitInBytes, periodSeconds, ipFingerprintFactor int) *RequestLimiter {
	l := &RequestLimiter{
		requestLimit:              requestLimit,
		requestLimitInBytes:       requestLimitInBytes,
		requestLimitPeriodSeconds: periodSeconds,
		ipFingerprintFactor:       ipFingerprintFactor,
		byIP:                      map[string][]time.Time{},
		maxRequests:               requestLimit,
		window:                    time.Duration(periodSeconds) * time.Second,
	}
	if ipFingerprintFactor > 0 {
		l.ipRequestLimit = requestLimit * ipFingerprintFactor
		l.ipRequestLimitInBytes = requestLimitInBytes * ipFingerprintFactor
	} else {
		l.ipRequestLimit = requestLimit
		l.ipRequestLimitInBytes = requestLimitInBytes
	}
	return l
}

// ComputeFingerprint ports RequestLimiter.computeFingerprint.
func ComputeFingerprint(httpHeader map[string][]string, parameters map[string]string) string {
	get := func(k string) string {
		if httpHeader == nil {
			return ""
		}
		v := httpHeader[k]
		if len(v) == 0 {
			return ""
		}
		return strings.Join(v, "|")
	}
	session := ""
	if parameters != nil {
		session = parameters["textSessionId"]
	}
	return strings.Join([]string{
		get("User-Agent"),
		get("Accept-Language"),
		get("Referer"),
		session,
	}, "|")
}

func getRequestSize(params map[string]string) int {
	if params == nil {
		return 0
	}
	if t, ok := params["text"]; ok && t != "" {
		return len(t)
	}
	if d, ok := params["data"]; ok && d != "" {
		return len(d)
	}
	return 0
}

func modeFromParams(params map[string]string) CheckMode {
	if params != nil && strings.EqualFold(params["mode"], "textLevelOnly") {
		return CheckModeTextLevelOnly
	}
	return CheckModeAll
}

// CheckAccess ports checkAccess (records event then enforces limits).
// skipLimits true skips enforcement (UserLimits.getSkipLimits).
func (l *RequestLimiter) CheckAccess(ip string, params map[string]string, httpHeader map[string][]string, skipLimits bool) error {
	if l == nil || skipLimits {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	reqSize := getRequestSize(params)
	for len(l.requestEvents) > requestQueueSize {
		l.requestEvents = l.requestEvents[1:]
	}
	fp := ComputeFingerprint(httpHeader, params)
	l.requestEvents = append(l.requestEvents, requestEvent{
		ip:          ip,
		date:        time.Now(),
		sizeInBytes: reqSize,
		fingerprint: fp,
		mode:        modeFromParams(params),
	})
	return l.checkLimitLocked(ip, httpHeader, params)
}

func (l *RequestLimiter) checkLimitLocked(ip string, httpHeader map[string][]string, parameters map[string]string) error {
	requestsByIP := 0
	requestSizeByIP := 0.0
	requestsByFP := 0
	requestSizeByFP := 0.0
	threshold := time.Now().Add(-time.Duration(l.requestLimitPeriodSeconds) * time.Second)
	fingerprint := ComputeFingerprint(httpHeader, parameters)

	for _, event := range l.requestEvents {
		if event.ip != ip || !event.date.After(threshold) {
			continue
		}
		modeFactor := 1.0
		if event.mode == CheckModeTextLevelOnly {
			modeFactor = 0.1
		}
		requestsByIP++
		requestSizeByIP += float64(event.sizeInBytes) * modeFactor
		if event.fingerprint == fingerprint {
			requestsByFP++
			requestSizeByFP += float64(event.sizeInBytes) * modeFactor
		}
		if l.ipFingerprintFactor > 0 && l.requestLimit > 0 && requestsByFP > l.requestLimit {
			return NewTooManyRequestsError( fmt.Sprintf(
				"Client request limit of %d requests per %d seconds exceeded",
				l.requestLimit, l.requestLimitPeriodSeconds))
		}
		if l.requestLimit > 0 && requestsByIP > l.ipRequestLimit {
			return NewTooManyRequestsError( fmt.Sprintf(
				"IP request limit of %d requests per %d seconds exceeded",
				l.ipRequestLimit, l.requestLimitPeriodSeconds))
		}
		if event.mode == CheckModeTextLevelOnly {
			if l.ipFingerprintFactor > 0 && l.requestLimitInBytes > 0 && requestSizeByFP > float64(l.requestLimitInBytes) {
				return NewTooManyRequestsError( fmt.Sprintf(
					"Client request size limit of %d bytes per %d seconds exceeded in text-level checks",
					l.requestLimitInBytes, l.requestLimitPeriodSeconds))
			}
			if l.requestLimitInBytes > 0 && requestSizeByIP > float64(l.ipRequestLimitInBytes) {
				return NewTooManyRequestsError( fmt.Sprintf(
					"IP request size limit of %d bytes per %d seconds exceeded in text-level checks",
					l.ipRequestLimitInBytes, l.requestLimitPeriodSeconds))
			}
		} else {
			if l.ipFingerprintFactor > 0 && l.requestLimitInBytes > 0 && requestSizeByFP > float64(l.requestLimitInBytes) {
				return NewTooManyRequestsError( fmt.Sprintf(
					"Client request size limit of %d bytes per %d seconds exceeded",
					l.requestLimitInBytes, l.requestLimitPeriodSeconds))
			}
			if l.requestLimitInBytes > 0 && requestSizeByIP > float64(l.ipRequestLimitInBytes) {
				return NewTooManyRequestsError( fmt.Sprintf(
					"IP request size limit of %d bytes per %d seconds exceeded",
					l.ipRequestLimitInBytes, l.requestLimitPeriodSeconds))
			}
		}
	}
	return nil
}

// Allow records a request from ip and returns false if over limit (legacy count path).
func (l *RequestLimiter) Allow(ip string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-l.window)
	if l.byIP == nil {
		l.byIP = map[string][]time.Time{}
	}
	times := l.byIP[ip]
	i := 0
	for _, t := range times {
		if t.After(cutoff) {
			times[i] = t
			i++
		}
	}
	times = times[:i]
	if len(times) >= l.maxRequests {
		l.byIP[ip] = times
		return false
	}
	l.byIP[ip] = append(times, now)
	return true
}

// Count returns recent request count for ip (legacy).
func (l *RequestLimiter) Count(ip string) int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-l.window)
	n := 0
	for _, t := range l.byIP[ip] {
		if t.After(cutoff) {
			n++
		}
	}
	return n
}

// Record appends a request timestamp for ip without enforcing the limit
// (used by ErrorRequestLimiter.LogAccess).
func (l *RequestLimiter) Record(ip string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-l.window)
	if l.byIP == nil {
		l.byIP = map[string][]time.Time{}
	}
	times := l.byIP[ip]
	i := 0
	for _, t := range times {
		if t.After(cutoff) {
			times[i] = t
			i++
		}
	}
	l.byIP[ip] = append(times[:i], now)
}
