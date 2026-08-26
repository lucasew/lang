package tools

// Span is a minimal OpenTelemetry span surface (no OTEL dependency).
type Span interface {
	// Name returns the span name.
	Name() string
	// SetAttribute records a key/value attribute.
	SetAttribute(key string, value any)
	// RecordError attaches an error to the span.
	RecordError(err error)
	// End finishes the span.
	End()
}

// noopSpan is used when no real tracer is configured.
type noopSpan struct{ name string }

func (s noopSpan) Name() string                 { return s.name }
func (noopSpan) SetAttribute(string, any)       {}
func (noopSpan) RecordError(error)              {}
func (noopSpan) End()                           {}

// TracedFunction ports org.languagetool.tools.TracedFunction.
type TracedFunction[T any] func(span Span) (T, error)
