package rules

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConfusionPair ports org.languagetool.rules.ConfusionPair.
type ConfusionPair struct {
	term1         *ConfusionString
	term2         *ConfusionString
	factor        int64
	bidirectional bool
}

func NewConfusionPair(cs1, cs2 *ConfusionString, factor int64, bidirectional bool) *ConfusionPair {
	if cs1 == nil || cs2 == nil {
		panic("ConfusionPair: nil ConfusionString")
	}
	if factor < 1 {
		panic(fmt.Sprintf("factor must be >= 1: %d", factor))
	}
	return &ConfusionPair{term1: cs1, term2: cs2, factor: factor, bidirectional: bidirectional}
}

func NewConfusionPairTokens(token1, token2 string, factor int64, bidirectional bool) *ConfusionPair {
	return NewConfusionPair(
		NewConfusionString(token1, nil),
		NewConfusionString(token2, nil),
		factor, bidirectional,
	)
}

func (p *ConfusionPair) GetFactor() int64            { return p.factor }
func (p *ConfusionPair) IsBidirectional() bool        { return p.bidirectional }
func (p *ConfusionPair) GetTerm1() *ConfusionString   { return p.term1 }
func (p *ConfusionPair) GetTerm2() *ConfusionString   { return p.term2 }
func (p *ConfusionPair) GetTerms() []*ConfusionString { return []*ConfusionString{p.term1, p.term2} }

func (p *ConfusionPair) GetUppercaseFirstCharTerms() []*ConfusionString {
	return []*ConfusionString{
		NewConfusionString(tools.UppercaseFirstChar(p.term1.GetString()), p.term1.GetDescription()),
		NewConfusionString(tools.UppercaseFirstChar(p.term2.GetString()), p.term2.GetDescription()),
	}
}

func (p *ConfusionPair) String() string {
	sep := " -> "
	if p.bidirectional {
		sep = "; "
	}
	return p.term1.String() + sep + p.term2.String()
}

func (p *ConfusionPair) Equals(o *ConfusionPair) bool {
	if p == o {
		return true
	}
	if p == nil || o == nil {
		return false
	}
	return p.factor == o.factor &&
		p.bidirectional == o.bidirectional &&
		p.term1.Equal(o.term1) &&
		p.term2.Equal(o.term2)
}
