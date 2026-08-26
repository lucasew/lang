package uk

// CompoundDebugLogger ports tagging.uk.CompoundDebugLogger (optional verbose log).
type CompoundDebugLogger struct {
	Enabled bool
	// Lines collects messages when Enabled (tests / diagnostics).
	Lines []string
}

func NewCompoundDebugLogger(enabled bool) *CompoundDebugLogger {
	return &CompoundDebugLogger{Enabled: enabled}
}

func (l *CompoundDebugLogger) Log(kind, word string) {
	if l == nil || !l.Enabled {
		return
	}
	l.Lines = append(l.Lines, kind+":"+word)
}
