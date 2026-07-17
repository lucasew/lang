package patterns

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// PatternTokenMatcher ports org.languagetool.rules.patterns.PatternTokenMatcher
// for basic string/regex/POS matching (full exception/and-group later).
type PatternTokenMatcher struct {
	Base *PatternToken
	// compiled RE for Token when Regexp is set
	tokenRE *regexp.Regexp
}

func NewPatternTokenMatcher(pt *PatternToken) *PatternTokenMatcher {
	m := &PatternTokenMatcher{Base: pt}
	if pt != nil && pt.Regexp && pt.Token != "" {
		flags := ""
		if !pt.CaseSensitive {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + "^(?:" + pt.Token + ")$")
		if err == nil {
			m.tokenRE = re
		}
	}
	return m
}

func (m *PatternTokenMatcher) GetPatternToken() *PatternToken {
	if m == nil {
		return nil
	}
	return m.Base
}

// IsMatched checks whether a single AnalyzedToken matches the pattern token.
func (m *PatternTokenMatcher) IsMatched(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.Base == nil || token == nil {
		return false
	}
	pt := m.Base
	// Positive string exception: matching surface/lemma means "do not match this pattern token".
	if pt.TokenException != "" && m.matchesException(token) {
		if pt.Negation {
			return true
		}
		return false
	}
	matched := m.matchSurface(token.GetToken())
	if pt.MatchInflected && !matched {
		if lem := token.GetLemma(); lem != nil && *lem != "" {
			matched = m.matchSurface(*lem)
		}
		// Soft path without a tagger: accept simple morphological extensions
		// of the base form (abono→abonos) so inflected soft rules still fire.
		if !matched {
			matched = softInflectedSurfaceMatch(token.GetToken(), pt.Token, pt.CaseSensitive)
		}
		// Esperanto: try x-system/diacritic fold and common -o/-oj/-ojn stems.
		if !matched {
			for _, cand := range softEsperantoLemmaCandidates(token.GetToken()) {
				if m.matchSurface(cand) {
					matched = true
					break
				}
			}
		}
	}
	if pt.Pos != nil && pt.Pos.PosTag != "" {
		pos := token.GetPOSTag()
		posOK := false
		if pos != nil {
			if pt.Pos.Regexp {
				re, err := regexp.Compile("^(?:" + pt.Pos.PosTag + ")$")
				if err == nil {
					posOK = re.MatchString(*pos)
				}
			} else {
				posOK = *pos == pt.Pos.PosTag
			}
		} else {
			// Soft path without a tagger: untagged tokens act as UNKNOWN.
			// Postag-only empty surface tokens also accept any letter word.
			tag := strings.ToUpper(pt.Pos.PosTag)
			if tag == "UNKNOWN" || strings.HasPrefix(tag, "UNKNOWN") {
				posOK = true
			} else if pt.Token == "" {
				posOK = softLooksLikeWord(token.GetToken())
			}
		}
		if pt.Pos.Negate {
			posOK = !posOK
		}
		// if only POS is set (empty token), POS decides
		if pt.Token == "" {
			matched = posOK
		} else {
			matched = matched && posOK
		}
	}
	if pt.Negation {
		return !matched
	}
	return matched
}

func (m *PatternTokenMatcher) matchesException(token *languagetool.AnalyzedToken) bool {
	pt := m.Base
	if pt == nil || pt.TokenException == "" || token == nil {
		return false
	}
	surface := token.GetToken()
	// Exception case sensitivity is independent of the pattern token (LT).
	excCS := pt.TokenExceptionCaseSensitive
	if pt.TokenExceptionRE {
		flags := ""
		if !excCS {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + "^(?:" + pt.TokenException + ")$")
		if err != nil {
			return false
		}
		if re.MatchString(surface) {
			return true
		}
		if lem := token.GetLemma(); lem != nil {
			return re.MatchString(*lem)
		}
		return false
	}
	if excCS {
		if surface == pt.TokenException {
			return true
		}
	} else if strings.EqualFold(surface, pt.TokenException) {
		return true
	}
	if lem := token.GetLemma(); lem != nil {
		if excCS {
			return *lem == pt.TokenException
		}
		return strings.EqualFold(*lem, pt.TokenException)
	}
	return false
}

// IsMatchedReadings is true if any reading of atr matches.
func (m *PatternTokenMatcher) IsMatchedReadings(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	for _, r := range atr.GetReadings() {
		if m.IsMatched(r) {
			return true
		}
	}
	// also allow surface-only match against token string when untagged
	return m.IsMatched(languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil))
}

func (m *PatternTokenMatcher) matchSurface(surface string) bool {
	pt := m.Base
	if pt.Token == "" {
		return true
	}
	// Soft: treat ASCII and typographic apostrophes as equivalent so
	// French soft packs (often ASCII d'/l') match FrenchWordTokenizer (often ’).
	surface = normalizeApostrophes(surface)
	want := normalizeApostrophes(pt.Token)
	if pt.Regexp {
		if m.tokenRE != nil {
			if m.tokenRE.MatchString(surface) {
				return true
			}
			// Soft EO x-system (Ambaux) — only when digraphs are present, never lowercasing alone.
			if folded := softEsperantoUnicode(surface); folded != surface && m.tokenRE.MatchString(folded) {
				return true
			}
			// Inflected EO/regexp (biliardoj vs biliardo|…): try lemma-like candidates.
			if pt.MatchInflected {
				for _, cand := range softEsperantoLemmaCandidates(surface) {
					if m.tokenRE.MatchString(cand) {
						return true
					}
				}
			}
			return false
		}
		return false
	}
	if pt.CaseSensitive {
		// Exact only — do not EO-fold (would ignore case via ToLower).
		return surface == want
	}
	if strings.EqualFold(surface, want) {
		return true
	}
	// Soft Esperanto: Ambaux/Ambau ↔ ambaŭ after x-system + diacritic fold.
	return softEsperantoFold(surface) == softEsperantoFold(want)
}

func normalizeApostrophes(s string) string {
	if s == "" {
		return s
	}
	// U+2019 right single quotation mark, U+02BC modifier letter apostrophe, U+2018 left.
	s = strings.ReplaceAll(s, "\u2019", "'")
	s = strings.ReplaceAll(s, "\u02BC", "'")
	s = strings.ReplaceAll(s, "\u2018", "'")
	return s
}

func softLooksLikeWord(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	letters := 0
	for _, r := range s {
		if r == '-' || r == '\'' || r == '’' {
			continue
		}
		if !unicode.IsLetter(r) {
			return false
		}
		letters++
	}
	return letters > 0
}

// softInflectedSurfaceMatch approximates lemma matching without a tagger:
// surface equals base, or base is a prefix of surface with a short suffix (s, es, n, en, …).
func softInflectedSurfaceMatch(surface, base string, caseSensitive bool) bool {
	if surface == "" || base == "" {
		return false
	}
	if !caseSensitive {
		surface = strings.ToLower(surface)
		base = strings.ToLower(base)
	}
	// EO x-system / diacritic fold before prefix checks.
	if softEsperantoFold(surface) == softEsperantoFold(base) {
		return true
	}
	if surface == base {
		return true
	}
	// Prefix check on folded forms (ambaŭ / Ambaux).
	sf, bf := softEsperantoFold(surface), softEsperantoFold(base)
	if strings.HasPrefix(sf, bf) {
		suf := sf[len(bf):]
		if len(suf) > 0 && len(suf) <= 4 {
			switch suf {
			case "s", "es", "n", "en", "er", "e", "a", "os", "as", "is", "ns", "j", "jn", "oj", "ojn", "an", "on":
				return true
			default:
				ok := true
				for _, r := range suf {
					if !unicode.IsLetter(r) {
						ok = false
						break
					}
				}
				if ok && len(suf) <= 2 {
					return true
				}
			}
		}
	}
	if !strings.HasPrefix(surface, base) {
		return false
	}
	suf := surface[len(base):]
	if len(suf) == 0 || len(suf) > 4 {
		return false
	}
	// Common short inflection suffixes across LT languages (not full morphology).
	switch suf {
	case "s", "es", "n", "en", "er", "e", "a", "os", "as", "is", "ns", "aren", "eren", "j", "jn", "oj", "ojn":
		return true
	default:
		// all-letter short suffix only
		for _, r := range suf {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return len(suf) <= 2
	}
}

// softEsperantoUnicode converts x-system digraphs to Unicode diacritics (cx→ĉ).
func softEsperantoUnicode(s string) string {
	if s == "" || !strings.ContainsAny(strings.ToLower(s), "x") {
		return s
	}
	// Process lowercase digraphs in a case-preserving way via lower map then restore is hard;
	// apply case-insensitive sequential replaces on a lowered copy for matching only.
	low := strings.ToLower(s)
	repl := []struct{ from, to string }{
		{"cx", "ĉ"}, {"gx", "ĝ"}, {"hx", "ĥ"}, {"jx", "ĵ"}, {"sx", "ŝ"}, {"ux", "ŭ"},
	}
	for _, r := range repl {
		low = strings.ReplaceAll(low, r.from, r.to)
	}
	return low
}

// softEsperantoFold maps x-system and EO diacritics to plain ASCII letters for soft compare.
func softEsperantoFold(s string) string {
	s = softEsperantoUnicode(strings.ToLower(s))
	return strings.NewReplacer(
		"ĉ", "c", "ĝ", "g", "ĥ", "h", "ĵ", "j", "ŝ", "s", "ŭ", "u",
	).Replace(s)
}

// softEsperantoLemmaCandidates yields likely dictionary forms for EO surfaces (biliardoj→biliardo).
func softEsperantoLemmaCandidates(surface string) []string {
	if surface == "" {
		return nil
	}
	u := softEsperantoUnicode(strings.ToLower(surface))
	out := []string{u}
	// Strip accusative/plural endings common in EO.
	type strip struct{ suf, base string }
	for _, st := range []strip{
		{"ojn", "o"}, {"oj", "o"}, {"on", "o"}, {"an", "a"}, {"en", "e"},
		{"ajn", "a"}, {"ojn", "o"}, {"n", ""}, {"j", ""},
	} {
		if strings.HasSuffix(u, st.suf) {
			stem := u[:len(u)-len(st.suf)] + st.base
			if stem != u && stem != "" {
				out = append(out, stem)
			}
		}
	}
	return out
}
