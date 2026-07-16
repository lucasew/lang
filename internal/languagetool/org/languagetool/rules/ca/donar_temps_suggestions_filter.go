package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DonarTempsSuggestionsFilter ports DONAR_TEMPS suggestion assembly.
type DonarTempsSuggestionsFilter struct {
	// SynthHaver synthesizes "haver" with VA+suffix (e.g. "ha").
	SynthHaver func(verbPostagSuffix string) string
	// SynthTenir synthesizes "tenir" for the given postag pattern.
	SynthTenir func(postag string) string
}

func NewDonarTempsSuggestionsFilter() *DonarTempsSuggestionsFilter {
	return &DonarTempsSuggestionsFilter{}
}

// DonarTempsInput holds pre-analyzed span pieces.
type DonarTempsInput struct {
	// PronomGenderNumber e.g. "1S" from P postag chars
	PronomGenderNumber string
	// AuxTokens between first verb and "donar" (exclusive of donar)
	AuxTokens []string
	// VerbPostag of "donar"
	VerbPostag string
	// CasingModel for PreserveCase (usually the pronoun token)
	CasingModel string
}

// Suggest returns "hi ha temps" / "tinc temps" style replacements.
func (f *DonarTempsSuggestionsFilter) Suggest(in DonarTempsInput) []string {
	if len(in.VerbPostag) < 8 {
		return nil
	}
	var out []string
	// haver-hi temps
	if f.SynthHaver != nil {
		suffix := in.VerbPostag[2:8]
		haver := f.SynthHaver(suffix)
		if haver != "" {
			var b strings.Builder
			b.WriteString("hi")
			for _, tok := range in.AuxTokens {
				b.WriteByte(' ')
				b.WriteString(tok)
			}
			b.WriteString(" ")
			b.WriteString(haver)
			b.WriteString(" temps")
			s := strings.ReplaceAll(b.String(), "de haver", "d'haver")
			if in.CasingModel != "" {
				s = tools.PreserveCase(s, in.CasingModel)
			}
			out = append(out, s)
		}
	}
	// tenir temps
	if f.SynthTenir != nil {
		var s string
		if len(in.AuxTokens) == 0 {
			// direct: tenir with person from pronoun
			postag := in.VerbPostag[:4] + in.PronomGenderNumber + in.VerbPostag[6:8]
			tenir := f.SynthTenir(postag)
			if tenir != "" {
				s = tenir + " temps"
			}
		} else {
			// keep aux then tenir for main verb postag
			tenir := f.SynthTenir(in.VerbPostag)
			if tenir != "" {
				var b strings.Builder
				for i, tok := range in.AuxTokens {
					if i > 0 {
						b.WriteByte(' ')
					}
					b.WriteString(tok)
				}
				b.WriteString(" ")
				b.WriteString(tenir)
				b.WriteString(" temps")
				s = b.String()
			}
		}
		if s != "" {
			if in.CasingModel != "" {
				s = tools.PreserveCase(s, in.CasingModel)
			}
			out = append(out, s)
		}
	}
	return out
}

// PronomGenderNumberFromP extracts person+number from a P… postag.
func PronomGenderNumberFromP(postag string) string {
	if len(postag) < 5 {
		return ""
	}
	return string(postag[2]) + string(postag[4])
}
