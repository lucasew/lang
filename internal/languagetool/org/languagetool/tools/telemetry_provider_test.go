package tools

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTelemetryProvider(t *testing.T) {
	p := &TelemetryProvider{}
	var ended bool
	p.SetSpanFactory(func(name string, attrs map[string]any) Span {
		return &recordingSpan{name: name, onEnd: func() { ended = true }}
	})
	v, err := CreateSpanValue(p, "check", map[string]any{"lang": "en"}, WrappedValue[string](func() (string, error) {
		return "ok", nil
	}))
	require.NoError(t, err)
	require.Equal(t, "ok", v)
	require.True(t, ended)

	_, err = CreateSpanTraced(p, "fail", nil, TracedFunction[int](func(span Span) (int, error) {
		return 0, errors.New("boom")
	}))
	require.Error(t, err)
}

type recordingSpan struct {
	name  string
	onEnd func()
	err   error
}

func (s *recordingSpan) Name() string           { return s.name }
func (s *recordingSpan) SetAttribute(string, any) {}
func (s *recordingSpan) RecordError(err error)  { s.err = err }
func (s *recordingSpan) End() {
	if s.onEnd != nil {
		s.onEnd()
	}
}
