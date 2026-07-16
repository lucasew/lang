package server

// ErrorRequestLimiter ports org.languagetool.server.ErrorRequestLimiter.
// Limits error-producing requests per IP (uses RequestLimiter under the hood).
type ErrorRequestLimiter struct {
	inner *RequestLimiter
}

func NewErrorRequestLimiter(requestLimit, periodSeconds int) *ErrorRequestLimiter {
	return &ErrorRequestLimiter{inner: NewRequestLimiter(requestLimit, periodSeconds)}
}

// WouldAccessBeOkay reports whether the client is under the error limit.
func (e *ErrorRequestLimiter) WouldAccessBeOkay(ip string) bool {
	if e == nil || e.inner == nil {
		return true
	}
	// Peek without consuming: check count only.
	return e.inner.Count(ip) < e.inner.maxRequests
}

// LogAccess records an error-causing request against the IP.
func (e *ErrorRequestLimiter) LogAccess(ip string) {
	if e == nil || e.inner == nil {
		return
	}
	_ = e.inner.Allow(ip)
}

// CheckLimit returns an error if the IP is over the error request limit.
func (e *ErrorRequestLimiter) CheckLimit(ip string) error {
	if e == nil || e.inner == nil {
		return nil
	}
	if !e.WouldAccessBeOkay(ip) {
		return NewTooManyRequestsError("Error request limit exceeded for " + ip)
	}
	return nil
}
