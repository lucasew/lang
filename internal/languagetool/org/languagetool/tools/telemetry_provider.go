package tools

import "sync"

// TelemetryProvider ports org.languagetool.tools.TelemetryProvider without OpenTelemetry.
// Spans are no-ops unless a custom factory is installed.
type TelemetryProvider struct {
	mu      sync.Mutex
	factory func(name string, attrs map[string]any) Span
}

// DefaultTelemetryProvider is the process-wide INSTANCE.
var DefaultTelemetryProvider = &TelemetryProvider{}

// SetSpanFactory overrides span creation (tests / real OTEL wiring).
func (p *TelemetryProvider) SetSpanFactory(f func(name string, attrs map[string]any) Span) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.factory = f
}

func (p *TelemetryProvider) newSpan(name string, attrs map[string]any) Span {
	p.mu.Lock()
	f := p.factory
	p.mu.Unlock()
	if f != nil {
		return f(name, attrs)
	}
	return noopSpan{name: name}
}

// CreateSpan runs fn inside a span and returns its result.
func (p *TelemetryProvider) CreateSpan(name string, attrs map[string]any, fn TracedFunction[any]) (any, error) {
	span := p.newSpan(name, attrs)
	defer span.End()
	v, err := fn(span)
	if err != nil {
		span.RecordError(err)
		return v, err
	}
	return v, nil
}

// CreateSpanValue runs a WrappedValue inside a span.
func CreateSpanValue[T any](p *TelemetryProvider, name string, attrs map[string]any, fn WrappedValue[T]) (T, error) {
	if p == nil {
		p = DefaultTelemetryProvider
	}
	span := p.newSpan(name, attrs)
	defer span.End()
	v, err := fn.Call()
	if err != nil {
		span.RecordError(err)
	}
	return v, err
}

// CreateSpanTraced runs a TracedFunction inside a span.
func CreateSpanTraced[T any](p *TelemetryProvider, name string, attrs map[string]any, fn TracedFunction[T]) (T, error) {
	if p == nil {
		p = DefaultTelemetryProvider
	}
	span := p.newSpan(name, attrs)
	defer span.End()
	v, err := fn(span)
	if err != nil {
		span.RecordError(err)
	}
	return v, err
}
