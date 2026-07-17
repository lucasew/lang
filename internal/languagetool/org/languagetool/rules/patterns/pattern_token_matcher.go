package patterns

import (
	"fmt"
	"regexp"
	"strconv"
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
		re, err := regexp.Compile(flags + "^(?:" + softNormalizeJavaRegexp(pt.Token) + ")$")
		if err == nil {
			m.tokenRE = re
		}
	}
	return m
}

// softNormalizeJavaRegexp maps Java/PCRE unicode escapes used in LT XML
// (\uXXXX, \UXXXXXXXX) to Go RE2 \x{...} form. Leaves other escapes alone.
func softNormalizeJavaRegexp(pat string) string {
	if pat == "" || !strings.Contains(pat, `\u`) && !strings.Contains(pat, `\U`) {
		return pat
	}
	var b strings.Builder
	b.Grow(len(pat) + 8)
	for i := 0; i < len(pat); {
		if pat[i] == '\\' && i+1 < len(pat) {
			switch pat[i+1] {
			case 'u':
				// \uXXXX
				if i+6 <= len(pat) {
					hex := pat[i+2 : i+6]
					if _, err := strconv.ParseUint(hex, 16, 32); err == nil {
						fmt.Fprintf(&b, `\x{%s}`, strings.ToLower(hex))
						i += 6
						continue
					}
				}
			case 'U':
				// \UXXXXXXXX
				if i+10 <= len(pat) {
					hex := pat[i+2 : i+10]
					if _, err := strconv.ParseUint(hex, 16, 32); err == nil {
						// strip leading zeros for \x{}
						n, _ := strconv.ParseUint(hex, 16, 32)
						fmt.Fprintf(&b, `\x{%x}`, n)
						i += 10
						continue
					}
				}
			}
		}
		b.WriteByte(pat[i])
		i++
	}
	return b.String()
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
		// RE patterns with | alternatives (программный|аппаратный): try each alt.
		if !matched && pt.Regexp && strings.Contains(pt.Token, "|") {
			for _, alt := range softRegexpAlternatives(pt.Token) {
				if softInflectedSurfaceMatch(token.GetToken(), alt, pt.CaseSensitive) {
					matched = true
					break
				}
			}
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
				re, err := regexp.Compile("^(?:" + softNormalizeJavaRegexp(pt.Pos.PosTag) + ")$")
				if err == nil {
					posOK = re.MatchString(*pos)
				}
			} else {
				posOK = *pos == pt.Pos.PosTag
			}
		} else {
			// Soft path without a tagger: untagged tokens act as UNKNOWN.
			// Postag-only empty surface tokens accept letter words or punctuation
			// when the postag pattern looks like sentence-end / punct.
			// Surface+punct-tag (e.g. token="." postag="SENT_END") also soft-accepts
			// when the surface already matched and looks like punctuation.
			// Surface+word POS (e.g. TL ADJECTIVE-V with RE+postag): when the
			// surface already matched, accept non-negated POS without a tagger.
			tag := strings.ToUpper(pt.Pos.PosTag)
			if tag == "UNKNOWN" || strings.HasPrefix(tag, "UNKNOWN") {
				posOK = true
			} else if pt.Token == "" {
				tok := token.GetToken()
				if softLooksLikeWord(tok) {
					posOK = true
				} else if softLooksLikePunct(tok) && softPostagLooksLikePunct(tag) {
					posOK = true
				}
			} else if softLooksLikePunct(token.GetToken()) && softPostagLooksLikePunct(tag) {
				posOK = true
			} else if matched && !pt.Pos.Negate {
				// Dual surface+POS constraint: surface is the only soft signal.
				posOK = true
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
		re, err := regexp.Compile(flags + "^(?:" + softNormalizeJavaRegexp(pt.TokenException) + ")$")
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
	// Keep the raw surface for regexp matching against REs compiled with either form.
	rawSurface := surface
	surface = normalizeApostrophes(surface)
	want := normalizeApostrophes(pt.Token)
	if pt.Regexp {
		if m.tokenRE != nil {
			// Try raw and apostrophe-normalized surfaces (pattern may use ’ or ').
			if m.tokenRE.MatchString(rawSurface) || m.tokenRE.MatchString(surface) {
				return true
			}
			// Soft EO x-system (Ambaux) — only when digraphs are present, never lowercasing alone.
			if folded := softEsperantoUnicode(rawSurface); folded != rawSurface && m.tokenRE.MatchString(folded) {
				return true
			}
			// Inflected EO/regexp (biliardoj vs biliardo|…): try lemma-like candidates.
			if pt.MatchInflected {
				for _, cand := range softEsperantoLemmaCandidates(rawSurface) {
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
		return rawSurface == pt.Token || surface == want
	}
	if strings.EqualFold(surface, want) || strings.EqualFold(rawSurface, pt.Token) {
		return true
	}
	// Soft Esperanto: Ambaux/Ambau ↔ ambaŭ after x-system + diacritic fold.
	return softEsperantoFold(rawSurface) == softEsperantoFold(pt.Token)
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
		// Allow combining marks (Khmer coeng/vowels, Indic matras, etc.).
		if unicode.IsLetter(r) {
			letters++
			continue
		}
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r) {
			continue
		}
		return false
	}
	return letters > 0
}

func softLooksLikePunct(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func softPostagLooksLikePunct(tag string) bool {
	// SENT_END, PSN*, PUNCT*, PCT (EN), SENTENCE_END, etc.
	u := strings.ToUpper(tag)
	return strings.Contains(u, "SENT_END") ||
		strings.Contains(u, "SENTENCE_END") ||
		strings.Contains(u, "PSN") ||
		strings.Contains(u, "PUNC") ||
		strings.Contains(u, "PCT") ||
		strings.Contains(u, "SENT_START")
}

// softEnglishLemma maps common irregular EN surfaces to dictionary lemmas
// (be/have/do). Used only for soft MatchInflected without a tagger.
var softEnglishLemma = map[string]string{
	"am": "be", "is": "be", "are": "be", "was": "be", "were": "be", "been": "be", "being": "be",
	"has": "have", "had": "have", "having": "have",
	"does": "do", "did": "do", "done": "do", "doing": "do",
}

// softInflectedSurfaceMatch approximates lemma matching without a tagger:
// surface equals base, or base is a prefix of surface with a short suffix (s, es, n, en, …).
// Also allows a shared stem of length ≥4 with short residual suffixes (говорить/говорите).
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
	// English irregular auxiliaries (was→be, has→have, …).
	if lem, ok := softEnglishLemma[surface]; ok && lem == base {
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
	if softSharedStemMatch(surface, base) {
		return true
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

// softSharedStemMatch is true when surface and base share a long letter stem
// and differ only by short inflectional endings (говорить/говорите, храбрый/храбрая).
// Min stem is 4 for longer words; 3 is allowed for short bases (яйцо/яйца).
func softSharedStemMatch(a, b string) bool {
	ar, br := []rune(a), []rune(b)
	n := 0
	for n < len(ar) && n < len(br) && ar[n] == br[n] {
		n++
	}
	minStem := 4
	if len(ar) <= 5 || len(br) <= 5 {
		minStem = 3
	}
	if n < minStem {
		return false
	}
	sa, sb := string(ar[n:]), string(br[n:])
	ra, rb := []rune(sa), []rune(sb)
	if len(ra) > 5 || len(rb) > 5 {
		return false
	}
	// residual must be letters only (inflection), not a different stem
	for _, r := range sa {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	for _, r := range sb {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// softRegexpAlternatives splits a simple top-level a|b|c pattern into alts.
// Nested groups/character classes are not fully parsed — only plain | splits
// used by upstream soft packs (программный|аппаратный).
func softRegexpAlternatives(pat string) []string {
	if pat == "" || !strings.Contains(pat, "|") {
		if pat == "" {
			return nil
		}
		return []string{pat}
	}
	// Strip outer non-capturing group if present.
	p := strings.TrimSpace(pat)
	if strings.HasPrefix(p, "(?:") && strings.HasSuffix(p, ")") {
		p = p[3 : len(p)-1]
	}
	depth := 0
	start := 0
	var alts []string
	for i, r := range p {
		switch r {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '|':
			if depth == 0 {
				alts = append(alts, p[start:i])
				start = i + 1
			}
		}
	}
	alts = append(alts, p[start:])
	out := make([]string, 0, len(alts))
	for _, a := range alts {
		a = strings.TrimSpace(a)
		if a != "" {
			out = append(out, a)
		}
	}
	return out
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
