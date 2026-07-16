package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortarTempsSuggestionsFilter ports PORTA_UNA_HORA style "fa … que" rewrites.
type PortarTempsSuggestionsFilter struct {
	// SynthFer synthesizes "fer" with postag pattern (e.g. VMI[30][S0].0).
	SynthFer func(postagPattern string) string
	// SynthInfinitiveToFinite turns gerund/infinitive lemma into finite form.
	SynthInfinitiveToFinite func(lemma, finitePostag string) string
	// SynthEstar synthesizes "estar" finite forms.
	SynthEstar func(finitePostag string) string
}

func NewPortarTempsSuggestionsFilter() *PortarTempsSuggestionsFilter {
	return &PortarTempsSuggestionsFilter{}
}

// PortarTempsKind classifies the token after the time span.
type PortarTempsKind int

const (
	PortarTempsQue PortarTempsKind = iota
	PortarTempsGerund
	PortarTempsSenseInf
	PortarTempsEstarPred
)

// PortarTempsInput is the surface input for Suggest.
type PortarTempsInput struct {
	PortarPostag string
	// TimeTokens are the PTime chunk tokens (e.g. "una", "hora")
	TimeTokens []string
	// Kind selects the continuation after the time span.
	Kind PortarTempsKind
	// For gerund/sense: lemma of the following verb and optional pronouns after it.
	NextLemma, PronounsAfter string
	CasingModel              string
}

// Suggest builds "fa una hora que …" replacements.
func (f *PortarTempsSuggestionsFilter) Suggest(in PortarTempsInput) string {
	if f.SynthFer == nil || len(in.PortarPostag) < 8 {
		return ""
	}
	// Java: verbPostag.substring(0,4)+"[30][S0]."+verbPostag.substring(7,8)
	pattern := in.PortarPostag[:4] + "[30][S0]." + string(in.PortarPostag[7])
	fer := f.SynthFer(pattern)
	if fer == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(fer)
	for _, t := range in.TimeTokens {
		b.WriteByte(' ')
		b.WriteString(t)
	}
	switch in.Kind {
	case PortarTempsQue:
		b.WriteString(" que")
	case PortarTempsGerund:
		if f.SynthInfinitiveToFinite == nil {
			return ""
		}
		// Java: "V.I"+verbPostag.substring(3,8)
		finiteTag := "V.I" + in.PortarPostag[3:8]
		fin := f.SynthInfinitiveToFinite(in.NextLemma, finiteTag)
		if fin == "" {
			return ""
		}
		b.WriteString(" que ")
		if in.PronounsAfter != "" {
			b.WriteString(TransformDavant(in.PronounsAfter, fin))
		}
		b.WriteString(fin)
	case PortarTempsSenseInf:
		if f.SynthInfinitiveToFinite == nil {
			return ""
		}
		finiteTag := "V.I" + in.PortarPostag[3:8]
		fin := f.SynthInfinitiveToFinite(in.NextLemma, finiteTag)
		if fin == "" {
			return ""
		}
		b.WriteString(" que no ")
		if in.PronounsAfter != "" {
			b.WriteString(TransformDavant(in.PronounsAfter, fin))
		}
		b.WriteString(fin)
	case PortarTempsEstarPred:
		if f.SynthEstar == nil {
			return ""
		}
		finiteTag := "V.I" + in.PortarPostag[3:8]
		estar := f.SynthEstar(finiteTag)
		if estar == "" {
			return ""
		}
		b.WriteString(" que ")
		b.WriteString(estar)
	default:
		return ""
	}
	s := b.String()
	if in.CasingModel != "" {
		s = tools.PreserveCase(s, in.CasingModel)
	}
	return s
}
