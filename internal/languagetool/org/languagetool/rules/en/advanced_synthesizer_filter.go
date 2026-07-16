package en

// AdvancedSynthesizerFilter ports the empty language subclass of AbstractAdvancedSynthesizerFilter.
// Full synthesizer integration is deferred; this type marks the filter extension point.
type AdvancedSynthesizerFilter struct{}

func NewAdvancedSynthesizerFilter() *AdvancedSynthesizerFilter {
	return &AdvancedSynthesizerFilter{}
}
