package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PossessiusRedundantsFilter ports suggestion assembly from
// org.languagetool.rules.ca.PossessiusRedundantsFilter.
type PossessiusRedundantsFilter struct{}

func NewPossessiusRedundantsFilter() *PossessiusRedundantsFilter {
	return &PossessiusRedundantsFilter{}
}

// PossessiveSuggestionInput holds pre-analyzed span info.
type PossessiveSuggestionInput struct {
	Persona, Number string // from PX postag (chars at 2 and 6)
	// PronounFound: matching dative/reflexive already present
	PronounFound, HasSomePronoun bool
	ApostropheNeeded             bool
	// Verb token at start of verb group
	VerbToken         string
	VerbIsInfOrGerund bool // VMN / VMG
	// Tokens between verb and possessive (exclusive of det before possessive)
	MiddleTokens []string
	// Det before possessive and following noun token
	DetToken, NounToken string
	// Full tokens for non-apostrophe rebuild (det + noun around possessive skip)
	AroundPossessive []string // det and following word(s) excluding possessive
	CasingModel      string
}

// Suggest builds the replacement string; empty means suppress.
func (f *PossessiusRedundantsFilter) Suggest(in PossessiveSuggestionInput) string {
	if in.PronounFound {
		if in.ApostropheNeeded {
			return "l'" + in.NounToken
		}
		return "" // delete possessive only — empty suggestion in Java
	}
	if !in.HasSomePronoun {
		var b strings.Builder
		if in.VerbIsInfOrGerund {
			pronounSugg := TransformDarrere(GetDativePronoun(in.Persona+in.Number), in.VerbToken)
			b.WriteString(in.VerbToken)
			b.WriteString(pronounSugg)
		} else {
			pronounSugg := TransformDavant(GetDativePronoun(in.Persona+in.Number), in.VerbToken)
			b.WriteString(tools.PreserveCase(pronounSugg, in.VerbToken))
			b.WriteString(strings.ToLower(in.VerbToken))
		}
		for _, tok := range in.MiddleTokens {
			b.WriteByte(' ')
			b.WriteString(strings.ToLower(tok))
		}
		if in.ApostropheNeeded {
			b.WriteString(" ")
			b.WriteString("l'" + in.NounToken)
		} else {
			for _, tok := range in.AroundPossessive {
				b.WriteByte(' ')
				b.WriteString(tok)
			}
		}
		s := b.String()
		if in.CasingModel != "" {
			s = tools.PreserveCase(s, in.CasingModel)
		}
		return strings.TrimSpace(s)
	}
	return "" // no suggestion
}

// PersonaNumberFromPX extracts persona+number from a PX… postag (Java substrings).
func PersonaNumberFromPX(postag string) (persona, number string) {
	if len(postag) < 7 {
		return "", ""
	}
	return string(postag[2]), string(postag[6])
}
