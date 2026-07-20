package languagetool

import "fmt"

// ExtendedSentenceRange ports org.languagetool.ExtendedSentenceRange.
type ExtendedSentenceRange struct {
	FromPos                 int
	ToPos                   int
	LanguageConfidenceRates map[string]float32 // language code → 0..1
}

func NewExtendedSentenceRange(fromPos, toPos int, languageCode string) ExtendedSentenceRange {
	return NewExtendedSentenceRangeWithRates(fromPos, toPos, map[string]float32{languageCode: 1.0})
}

func NewExtendedSentenceRangeWithRates(fromPos, toPos int, rates map[string]float32) ExtendedSentenceRange {
	cp := make(map[string]float32, len(rates))
	for k, v := range rates {
		cp[k] = v
	}
	return ExtendedSentenceRange{
		FromPos:                 fromPos,
		ToPos:                   toPos,
		LanguageConfidenceRates: cp,
	}
}

func (r ExtendedSentenceRange) GetFromPos() int { return r.FromPos }
func (r ExtendedSentenceRange) GetToPos() int   { return r.ToPos }
func (r ExtendedSentenceRange) GetLanguageConfidenceRates() map[string]float32 {
	return r.LanguageConfidenceRates
}

func (r *ExtendedSentenceRange) UpdateLanguageConfidenceRates(rates map[string]float32) {
	r.LanguageConfidenceRates = make(map[string]float32, len(rates))
	for k, v := range rates {
		r.LanguageConfidenceRates[k] = v
	}
}

func (r ExtendedSentenceRange) Equal(o ExtendedSentenceRange) bool {
	return r.FromPos == o.FromPos && r.ToPos == o.ToPos
}

// HashCode ports ExtendedSentenceRange.hashCode (fromPos, toPos only).
func (r ExtendedSentenceRange) HashCode() int {
	return 31*r.FromPos + r.ToPos
}

func (r ExtendedSentenceRange) Less(o ExtendedSentenceRange) bool {
	return r.FromPos < o.FromPos
}

func (r ExtendedSentenceRange) String() string {
	return fmt.Sprintf("%d-%d:%v", r.FromPos, r.ToPos, r.LanguageConfidenceRates)
}
