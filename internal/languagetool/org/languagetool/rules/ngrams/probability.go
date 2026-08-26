package ngrams

import (
	"fmt"
	"math"
)

// Probability ports org.languagetool.rules.ngrams.Probability.
// Does not enforce an upper bound of 1.0 (same as Java).
type Probability struct {
	prob        float64
	coverage    float32
	occurrences int64
}

func NewProbability(prob float64, coverage float32, occurrences int64) Probability {
	if prob < 0 {
		panic(fmt.Sprintf("Probability must be >= 0: %v", prob))
	}
	return Probability{prob: prob, coverage: coverage, occurrences: occurrences}
}

func NewProbabilitySimple(prob float64, coverage float32) Probability {
	return NewProbability(prob, coverage, -1)
}

func (p Probability) GetProb() float64      { return p.prob }
func (p Probability) GetLogProb() float64   { return math.Log(p.prob) }
func (p Probability) GetCoverage() float32  { return p.coverage }
func (p Probability) GetOccurrences() int64 { return p.occurrences }

func (p Probability) String() string {
	return fmt.Sprintf("%v, coverage=%v", p.prob, p.coverage)
}
