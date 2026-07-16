package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SynthesizeWithAnyDeterminerFilter ports multi-GN determiner+noun suggestion assembly.
// Form synthesis is pluggable; this builds determiner prefixes for MS/FS/MP/FP.
type SynthesizeWithAnyDeterminerFilter struct{}

func NewSynthesizeWithAnyDeterminerFilter() *SynthesizeWithAnyDeterminerFilter {
	return &SynthesizeWithAnyDeterminerFilter{}
}

// GenderNumberList is MS/FS/MP/FP.
var GenderNumberList = []string{"MS", "FS", "MP", "FP"}

// Prepositions recognized before determiners.
var PrepositionsList = []string{"a", "de", "per", "pe"}

// SuggestAll builds det+form (and prep+det+form) for each gender/number of form/POS pairs.
// preposition is "", "a", "de", "per", or "pe".
// preferGN (e.g. "MS") is listed first when non-empty.
func (f *SynthesizeWithAnyDeterminerFilter) SuggestAll(forms []struct{ Form, POS string }, preposition, preferGN, casingModel string) []string {
	// order gender numbers with prefer first
	gns := make([]string, 0, 4)
	if preferGN != "" {
		gns = append(gns, preferGN)
	}
	for _, gn := range GenderNumberList {
		if gn != preferGN {
			gns = append(gns, gn)
		}
	}
	var out []string
	seen := map[string]struct{}{}
	for _, gn := range gns {
		for _, fr := range forms {
			if fr.POS != "" && GenderNumberFromPOS(fr.POS) != "" && GenderNumberFromPOS(fr.POS) != gn {
				// if POS implies a different GN, skip unless empty
				if GenderNumberFromPOS(fr.POS) != gn {
					continue
				}
			}
			s := GetPrepositionAndDeterminer(fr.Form, gn, preposition) + fr.Form
			if casingModel != "" {
				s = tools.PreserveCase(s, casingModel)
			}
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

// PrepositionKey maps a full preposition token to the first-letter key used by ApostophationHelper.
func PrepositionKey(prep string) string {
	prep = strings.ToLower(strings.TrimSpace(prep))
	if prep == "" {
		return ""
	}
	return string([]rune(prep)[0])
}

// IsPreposition reports whether token is in PrepositionsList.
func IsPreposition(token string) bool {
	t := strings.ToLower(token)
	for _, p := range PrepositionsList {
		if t == p {
			return true
		}
	}
	return false
}
