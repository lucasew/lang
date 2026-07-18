package patterns

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PatternTokenMatcher ports org.languagetool.rules.patterns.PatternTokenMatcher
// for basic string/regex/POS matching (full exception/and-group later).
type PatternTokenMatcher struct {
	Base *PatternToken
	// compiled RE for Token when Regexp is set
	tokenRE *regexp.Regexp
	// StrictPOS when true: untagged tokens only match postag=UNKNOWN (Java
	// disambiguation with a real tagger). Soft grammar without a tagger keeps
	// StrictPOS false so open-class postags can soft-accept letter words.
	StrictPOS bool
}

func NewPatternTokenMatcher(pt *PatternToken) *PatternTokenMatcher {
	m := &PatternTokenMatcher{Base: pt}
	if pt != nil && pt.Regexp && pt.Token != "" {
		flags := ""
		// Go RE2 (?i) folds \p{Lu}/\p{Ll} so "run" matches \p{Lu}\p{Ll}+.
		// Java CASE_INSENSITIVE does not fold Unicode property classes the same way;
		// LT rules like NN_NNP_JJR rely on that (titlecase-only second token).
		pat := softNormalizeJavaRegexp(pt.Token)
		if !pt.CaseSensitive && !strings.Contains(pat, `\p{`) {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + "^(?:" + pat + ")$")
		if err == nil {
			m.tokenRE = re
		}
	}
	return m
}

// softNormalizeJavaRegexp maps Java/PCRE constructs used in LT XML to Go RE2:
//   - \uXXXX / \UXXXXXXXX → \x{...}
//   - inline (?iu)/(?i)/(?u) flags stripped (case handled via PatternToken.CaseSensitive)
func softNormalizeJavaRegexp(pat string) string {
	if pat == "" {
		return pat
	}
	// Strip Java inline flags RE2 rejects ((?iu) is common in DE soft packs).
	for _, flag := range []string{"(?iu)", "(?ui)", "(?i)", "(?u)", "(?m)", "(?s)"} {
		pat = strings.ReplaceAll(pat, flag, "")
	}
	if !strings.Contains(pat, `\u`) && !strings.Contains(pat, `\U`) {
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
	// Java PatternToken.isMatched: spacebefore=yes/no must match token.isWhitespaceBefore().
	if pt.WhitespaceBefore != nil && token.IsWhitespaceBefore() != *pt.WhitespaceBefore {
		return false
	}
	// Exceptions are checked after any-reading match (Java isExceptionMatchedCompletely
	// in AbstractPatternRulePerformer.testAllReadings), not per-reading here — so a
	// POS-only exception on reading NNS can reject a token whose VBZ reading matched.
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
		// Soft irregular lemmas against RE bases (põe→pôr for p[ôo]r, etc.).
		if !matched && pt.Regexp {
			if lems, ok := softIrregularLemma[strings.ToLower(token.GetToken())]; ok {
				for _, lem := range lems {
					if m.matchSurface(lem) {
						matched = true
						break
					}
				}
			}
		}
		// Inflected non-RE: German adj stems (lateinischen→lateinisch). Gate on
		// DE-looking surfaces so CA/Romance packs do not pay this cost every match.
		if !matched && !pt.Regexp && softLooksGermanAdjSurface(token.GetToken()) {
			for _, cand := range softGermanAdjCandidates(token.GetToken()) {
				if softInflectedSurfaceMatch(cand, pt.Token, pt.CaseSensitive) ||
					(!pt.CaseSensitive && strings.EqualFold(cand, pt.Token)) ||
					(pt.CaseSensitive && cand == pt.Token) {
					matched = true
					break
				}
			}
		}
		// Esperanto: x-system/diacritic fold and -o/-oj/-ojn / -as→-i stems.
		// Cheap candidate list (suffix strips); still needed for pure-ASCII EO
		// (darfas→darfi) with no diacritics.
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
			// Soft without portuguese.dict: FreeLing disambig may only add Z0MS0
			// to "um" (numeral) while grammar patterns require D.+ / D[AI].+|NC.+.
			// Java Morfologik also has DA0MS0 / DP… / DI…; soft-accept FreeLing D*
			// (and PD* pronouns) when the surface is a known closed-class form.
			if !posOK && !m.StrictPOS &&
				softFreeLingClosedSurfaceMatch(pt.Pos.PosTag, token.GetToken()) {
				posOK = true
			}
			// Soft multiword/disambig may only attach underscore context tags
			// (_CONTEXTO_…); treat those as non-morph so open-class postags still
			// soft-match letter surfaces (PT PHRASAL_VERB_PREFERIR "carne").
			if !posOK && !m.StrictPOS && softPOSIsContextTag(pos) &&
				softLooksLikeWord(token.GetToken()) && !softPostagIsClosedClassOnly(pt.Pos.PosTag) &&
				!softPostagIsNumberOnly(pt.Pos.PosTag) && !softPostagIsSentenceBoundary(strings.ToUpper(pt.Pos.PosTag)) {
				posOK = true
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
			// StrictPOS (disambiguation with a tagger): nil POS is UNKNOWN only.
			// Do not soft-promote America1s→VBN so OF_VBN_JJ invents JJ tags.
			if m.StrictPOS {
				posOK = tag == "UNKNOWN" || strings.HasPrefix(tag, "UNKNOWN")
			} else if tag == "UNKNOWN" || strings.HasPrefix(tag, "UNKNOWN") {
				posOK = true
			} else if pt.Token == "" {
				tok := token.GetToken()
				// SENT_START/SENT_END must not soft-match ordinary words (would
				// make boundary tokens match every letter token).
				if softPostagIsSentenceBoundary(tag) {
					if tok == "" {
						posOK = true
					} else if softLooksLikePunct(tok) && softPostagLooksLikePunct(tag) {
						posOK = true
					}
				} else if softPostagIsClosedClassOnly(tag) {
					// DT/PRP/IN/… without a tagger: only known closed-class surfaces
					// (avoids DT_PRP soft-matching "a man", "the search", …).
					posOK = softClosedClassSurfaceMatch(tag, tok)
				} else if softPostagIsNumberOnly(tag) {
					// Java Z/CD/CARD = numeral readings. Soft untagged must not treat
					// ordinary words as Z.+ (PT A_RANGE was replace-tagging every "a"
					// between two letter tokens with SP000 and breaking grammar POS).
					posOK = softLooksLikeNumber(tok)
				} else if softLooksLikeWord(tok) {
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

// IsMatchedByPreviousException ports PatternToken.isMatchedByPreviousException
// for soft surface/regexp scope="previous" exceptions.
func (m *PatternTokenMatcher) IsMatchedByPreviousException(prev *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || prev == nil || m.Base.PreviousException == "" {
		return false
	}
	return matchExceptionOnReadings(prev, m.Base.PreviousException, m.Base.PreviousExceptionRE, m.Base.PreviousExceptionCaseSensitive)
}

// IsMatchedByNextException ports PatternToken scope="next" exception matching.
func (m *PatternTokenMatcher) IsMatchedByNextException(next *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || next == nil || m.Base.NextException == "" {
		return false
	}
	return matchExceptionOnReadings(next, m.Base.NextException, m.Base.NextExceptionRE, m.Base.NextExceptionCaseSensitive)
}

func matchExceptionOnReadings(tok *languagetool.AnalyzedTokenReadings, exc string, isRE, caseSensitive bool) bool {
	if tok == nil || exc == "" {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		if matchExceptionSurface(r.GetToken(), exc, isRE, caseSensitive) {
			return true
		}
	}
	return matchExceptionSurface(tok.GetToken(), exc, isRE, caseSensitive)
}

func matchExceptionSurface(surface, exc string, isRE, caseSensitive bool) bool {
	if exc == "" {
		return false
	}
	if isRE {
		flags := ""
		excPat := softNormalizeJavaRegexp(exc)
		if !caseSensitive && !strings.Contains(excPat, `\p{`) {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + "^(?:" + excPat + ")$")
		if err != nil {
			return false
		}
		return re.MatchString(surface)
	}
	if caseSensitive {
		return surface == exc
	}
	return strings.EqualFold(surface, exc)
}

// IsMatchedReadings is true if any reading of atr matches.
// Java PatternTokenMatcher only tests real readings. Soft fallback to an
// untagged surface is only used when the token has no real POS (no tagger path).
// When POS tags exist, requiring postag must not soft-accept via a nil-POS probe
// (that previously made dual surface+POS constraints match every surface hit).
//
// Java and-groups: base + each and-member must match *some* reading of the token
// (PatternTokenMatcher.addMemberAndGroup / checkAndGroup across all readings).
func (m *PatternTokenMatcher) IsMatchedReadings(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	// Space-before is a property of the token readings, not each POS reading.
	if m.Base != nil && m.Base.WhitespaceBefore != nil {
		if atr.IsWhitespaceBefore() != *m.Base.WhitespaceBefore {
			return false
		}
	}
	// Java AbstractPatternRulePerformer.testAllReadings: chunk tags are AND-ed
	// with the reading match (exact contains or regexp on chunk name).
	if m.Base != nil && m.Base.ChunkTag != "" {
		if !chunkTagMatches(atr, m.Base.ChunkTag, m.Base.ChunkTagRegexp) {
			return false
		}
	}
	// And-group: evaluate base and each and-member across all readings.
	if m.Base != nil && len(m.Base.AndGroup) > 0 {
		if !m.matchAndGroupReadings(atr) {
			return false
		}
		return !m.anyReadingExceptionMatched(atr)
	}
	hasRealPOS := false
	anyMatched := false
	for _, r := range atr.GetReadings() {
		if r == nil {
			continue
		}
		if p := r.GetPOSTag(); p != nil && *p != "" && *p != languagetool.SentenceStartTagName &&
			*p != languagetool.SentenceEndTagName && *p != languagetool.ParagraphEndTagName {
			hasRealPOS = true
		}
		if m.IsMatched(r) {
			anyMatched = true
		}
	}
	if !anyMatched {
		if hasRealPOS {
			// Java: no match if no reading satisfied string+POS constraints.
			return false
		}
		// Chunk-only pattern tokens (empty surface/POS): chunk already matched above.
		if m.Base != nil && m.Base.Token == "" && (m.Base.Pos == nil || m.Base.Pos.PosTag == "") && m.Base.ChunkTag != "" {
			return !m.anyReadingExceptionMatched(atr)
		}
		// Disambiguation with a tagger (StrictPOS): untagged surfaces are UNKNOWN only.
		// Do not soft-promote America1s→VBN so OF_VBN_JJ does not invent JJ tags.
		if m.StrictPOS {
			if m.Base != nil && m.Base.Pos != nil && m.Base.Pos.PosTag != "" {
				tag := strings.ToUpper(m.Base.Pos.PosTag)
				if tag != "UNKNOWN" && !strings.HasPrefix(tag, "UNKNOWN") {
					return false
				}
			}
		}
		// Soft path without a tagger: untagged tokens only.
		// Propagate whitespace-before so spacebefore= constraints still apply.
		probe := languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil)
		probe.SetWhitespaceBefore(atr.IsWhitespaceBefore())
		if !m.IsMatched(probe) {
			return false
		}
		return !m.isExceptionMatchedCompletely(probe)
	}
	// Java AbstractPatternRulePerformer: after anyMatched, reject if any reading
	// isExceptionMatchedCompletely (surface and/or POS exception).
	return !m.anyReadingExceptionMatched(atr)
}

// anyReadingExceptionMatched is true when any reading matches the current exception.
func (m *PatternTokenMatcher) anyReadingExceptionMatched(atr *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || !m.Base.HasCurrentException() || atr == nil {
		return false
	}
	for _, r := range atr.GetReadings() {
		if r != nil && m.isExceptionMatchedCompletely(r) {
			return true
		}
	}
	// Surface exception also checks the token surface when readings are empty.
	if len(atr.GetReadings()) == 0 {
		probe := languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil)
		return m.isExceptionMatchedCompletely(probe)
	}
	return false
}

// isExceptionMatchedCompletely ports PatternToken.isExceptionMatchedCompletely for
// the soft current-scope exception (surface and/or POS). Both gates AND when set.
func (m *PatternTokenMatcher) isExceptionMatchedCompletely(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.Base == nil || token == nil || !m.Base.HasCurrentException() {
		return false
	}
	pt := m.Base
	// Java PatternToken exception without inflected="yes" matches surface only
	// (DOUBLE_AUX <exception>be</exception> must not reject lemma-be "were").
	if pt.TokenException != "" {
		if !matchExceptionSurface(token.GetToken(), pt.TokenException, pt.TokenExceptionRE, pt.TokenExceptionCaseSensitive) {
			return false
		}
	}
	if pt.TokenExceptionPosTag != "" {
		pos := token.GetPOSTag()
		if pos == nil || *pos == "" {
			return false
		}
		if pt.TokenExceptionPosRE {
			re, err := regexp.Compile("^(?:" + softNormalizeJavaRegexp(pt.TokenExceptionPosTag) + ")$")
			if err != nil || !re.MatchString(*pos) {
				return false
			}
		} else if *pos != pt.TokenExceptionPosTag {
			return false
		}
	}
	return true
}

// matchAndGroupReadings ports Java and-group accumulation over all readings.
func (m *PatternTokenMatcher) matchAndGroupReadings(atr *languagetool.AnalyzedTokenReadings) bool {
	base := m.Base
	// Match base without re-entering and-group logic.
	baseBare := *base
	baseBare.AndGroup = nil
	baseM := NewPatternTokenMatcher(&baseBare)
	baseM.StrictPOS = m.StrictPOS
	baseOK := false
	andOK := make([]bool, len(base.AndGroup))
	for _, r := range atr.GetReadings() {
		if r == nil {
			continue
		}
		if baseM.IsMatched(r) {
			baseOK = true
		}
		for i, andPt := range base.AndGroup {
			if andPt == nil {
				continue
			}
			am := NewPatternTokenMatcher(andPt)
			am.StrictPOS = m.StrictPOS
			if am.IsMatched(r) {
				andOK[i] = true
			}
		}
	}
	if !baseOK {
		// Soft untagged probe for base
		if !m.StrictPOS {
			probe := languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil)
			probe.SetWhitespaceBefore(atr.IsWhitespaceBefore())
			baseOK = baseM.IsMatched(probe)
		}
	}
	if !baseOK {
		return false
	}
	for _, ok := range andOK {
		if !ok {
			return false
		}
	}
	return true
}

// chunkTagMatches ports Java ChunkTag exact/regexp check on readings' chunk tags.
func chunkTagMatches(atr *languagetool.AnalyzedTokenReadings, want string, isRegexp bool) bool {
	if atr == nil || want == "" {
		return false
	}
	tags := atr.GetChunkTags()
	if isRegexp {
		return atr.MatchesChunkRegex(want)
	}
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
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
	// Arabic: Java tagger strips tashkeel for lookup; lemmas often keep diacritics
	// (هَذَا) while pattern tokens are undiacritized (هذا). Do NOT strip before
	// regexp match — patterns like .*اً$ need the tanwin character.
	surfaceNT, wantNT, patNT := surface, want, pt.Token
	if softHasArabic(surface) || softHasArabic(pt.Token) {
		surfaceNT = tools.RemoveTashkeel(surface)
		wantNT = tools.RemoveTashkeel(want)
		patNT = tools.RemoveTashkeel(pt.Token)
	}
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
				// German adjective/participle endings (Steigende→steigend for RE steigend?).
				for _, cand := range softGermanAdjCandidates(rawSurface) {
					if m.tokenRE.MatchString(cand) {
						return true
					}
				}
				// French -er participles (désactivé → désactiver for RE .*er / ETRE_DE_VERBE).
				for _, cand := range softFrenchErLemmaCandidates(rawSurface) {
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
		return rawSurface == pt.Token || surface == want ||
			(surfaceNT != surface && (surfaceNT == wantNT || surfaceNT == patNT))
	}
	if strings.EqualFold(surface, want) || strings.EqualFold(rawSurface, pt.Token) {
		return true
	}
	// Undiacritized Arabic equality (هَذَا ↔ هذا).
	if surfaceNT != surface || wantNT != want {
		if strings.EqualFold(surfaceNT, wantNT) || strings.EqualFold(surfaceNT, patNT) {
			return true
		}
	}
	// French elision: d'/d’ ↔ de, l’ ↔ le/la, qu’ ↔ que (FrenchWordTokenizer splits).
	if softFrenchElisionMatch(surface, want) || softFrenchElisionMatch(rawSurface, want) {
		return true
	}
	// Catalan proclitics (CatalanWordTokenizer): l'/d' ↔ el/la/de.
	if softCatalanElisionMatch(surface, want) || softCatalanElisionMatch(rawSurface, want) {
		return true
	}
	// Soft Esperanto: Ambaux/Ambau ↔ ambaŭ after x-system + diacritic fold.
	return softEsperantoFold(rawSurface) == softEsperantoFold(pt.Token)
}

// softCatalanElisionMatch mirrors French elision for Catalan articles/preps
// and fused contractions from CatalanWordTokenizer (pel→pe+l, al→a+l).
func softCatalanElisionMatch(surface, base string) bool {
	s := strings.ToLower(normalizeApostrophes(surface))
	b := strings.ToLower(normalizeApostrophes(base))
	if s == b {
		return true
	}
	switch b {
	case "el", "la", "els", "les":
		// article forms + proclitic l' (and gender swap el↔la for soft packs)
		return s == "l'" || s == "l" || s == "el" || s == "la" || s == "els" || s == "les"
	case "de":
		return s == "d'" || s == "d"
	case "per":
		// pel → pe + l
		return s == "pe"
	case "a":
		// al → a + l already leaves "a"
		return s == "a"
	default:
		return false
	}
}

// softHasArabic reports if s contains an Arabic-script letter.
func softHasArabic(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Arabic) {
			return true
		}
	}
	return false
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
	digits := 0
	for _, r := range s {
		if r == '-' || r == '\'' || r == '’' || r == ',' || r == '.' {
			// allow 1,000 / 3.14 style numbers as soft "words" for CD tags
			continue
		}
		// Catalan ela geminada (l·l): Java CatalanWordTokenizer treats · as a
		// word character (see wordCharacters). Soft POS open-class needs the same.
		if r == '·' || r == '\u00B7' || r == '\u2022' || r == '\u22C5' || r == '\u2219' {
			continue
		}
		// Allow combining marks (Khmer coeng/vowels, Indic matras, etc.).
		if unicode.IsLetter(r) {
			letters++
			continue
		}
		if unicode.IsDigit(r) {
			digits++
			continue
		}
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r) {
			continue
		}
		return false
	}
	return letters > 0 || digits > 0
}

// softLooksLikeNumber is true for digit-only surfaces (optional grouping/decimal).
// Used when soft postag is number-only (Java Z/CD/CARD) without a real tagger.
func softLooksLikeNumber(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	hasDigit := false
	for _, r := range s {
		if unicode.IsDigit(r) {
			hasDigit = true
			continue
		}
		// thousands / decimal / thin space separators common in FreeLing Z tokens
		if r == '.' || r == ',' || r == ' ' || r == '\u00A0' || r == '\u202F' || r == '-' || r == '−' {
			continue
		}
		return false
	}
	return hasDigit
}

// softPostagIsNumberOnly is true when every | alternative is a number POS
// (PT/ES FreeLing Z, EN CD, DE CARD, …). Mixed Z|N patterns stay open-class.
func softPostagIsNumberOnly(tag string) bool {
	u := strings.ToUpper(strings.TrimSpace(tag))
	if u == "" {
		return false
	}
	any := false
	for _, part := range strings.Split(u, "|") {
		p := softNormalizePostagPart(part)
		if p == "" {
			continue
		}
		if !softPostagPartIsNumber(p) {
			return false
		}
		any = true
	}
	return any
}

func softPostagPartIsNumber(p string) bool {
	// FreeLing (PT/ES/CA): Z, Z.*, Z.+
	if p == "Z" || strings.HasPrefix(p, "Z.") || strings.HasPrefix(p, "Z ") ||
		strings.HasPrefix(p, "Z+") || strings.HasPrefix(p, "Z*") {
		return true
	}
	// Penn CD; DE CARD. Do NOT treat Irish "Num:Ord" / "Num:Dig:Ord" as digit-only:
	// those are morphological ordinal tags for spelled-out words (tríú), not Z/CD.
	if p == "CD" || strings.HasPrefix(p, "CD.") || strings.HasPrefix(p, "CD ") ||
		strings.HasPrefix(p, "CD+") || strings.HasPrefix(p, "CD*") {
		return true
	}
	if strings.Contains(p, "CARD") {
		return true
	}
	// Bare NUM / NUM.* digit tags only (not Num:Ord, Num:Dig:Ord, …).
	if p == "NUM" || strings.HasPrefix(p, "NUM.") || strings.HasPrefix(p, "NUM ") ||
		strings.HasPrefix(p, "NUM+") || strings.HasPrefix(p, "NUM*") {
		return true
	}
	return false
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
	// SENT_END, PSN*, PUNCT*, PCT (EN), PKT (DE STTS), M (FR), SENTENCE_END, etc.
	u := strings.ToUpper(tag)
	if strings.Contains(u, "SENT_END") ||
		strings.Contains(u, "SENTENCE_END") ||
		strings.Contains(u, "PSN") ||
		strings.Contains(u, "PUNC") ||
		strings.Contains(u, "PCT") ||
		strings.Contains(u, "PKT") ||
		strings.Contains(u, "SENT_START") {
		return true
	}
	// French FreeLing-style: M / M.* / M punc (not MD modal, not MD.*)
	for _, part := range strings.Split(u, "|") {
		p := softNormalizePostagPart(part)
		if p == "M" || strings.HasPrefix(p, "M.") || strings.HasPrefix(p, "M ") ||
			(strings.HasPrefix(p, "M") && !strings.HasPrefix(p, "MD") && !strings.HasPrefix(p, "MD.")) {
			// M, M.*, M punc… but not MD (English modal) or MM…
			if p == "M" || strings.HasPrefix(p, "M.") || strings.HasPrefix(p, "M.*") || strings.HasPrefix(p, "M.+") {
				return true
			}
		}
	}
	return false
}

func softPostagIsSentenceBoundary(tag string) bool {
	// Pure boundary tags only. Alternatives like SENT_END|VB.* must soft-match
	// words as well as punctuation (SEVERAL_OTHER, WHAT_IT_HAPPENING).
	u := strings.ToUpper(strings.TrimSpace(tag))
	if u == "" {
		return false
	}
	for _, part := range strings.Split(u, "|") {
		p := strings.TrimSpace(part)
		p = strings.TrimPrefix(p, "^")
		p = strings.TrimSuffix(p, "$")
		if p == "" {
			continue
		}
		if !(strings.Contains(p, "SENT_START") ||
			strings.Contains(p, "SENT_END") ||
			strings.Contains(p, "SENTENCE_END") ||
			strings.Contains(p, "SENTENCE_START")) {
			return false
		}
	}
	return strings.Contains(u, "SENT")
}

// softPostagIsClosedClassOnly is true when every | alternative is a closed-class
// Penn tag family (DT, PRP, IN, …), not open classes (NN, VB, JJ, RB, CD, …).
func softPostagIsClosedClassOnly(tag string) bool {
	u := strings.ToUpper(strings.TrimSpace(tag))
	if u == "" || softPostagIsSentenceBoundary(u) {
		return false
	}
	// If any open-class family appears, treat as open (soft word OK).
	for _, open := range []string{"NN", "VB", "JJ", "RB", "CD", "FW", "UH", "SYM", "LS", "UNKNOWN"} {
		// avoid DT matching inside "UNKNOWN" etc. via careful checks
		if open == "UNKNOWN" && (u == "UNKNOWN" || strings.HasPrefix(u, "UNKNOWN")) {
			return false
		}
	}
	for _, part := range strings.Split(u, "|") {
		// FreeLing (SPS00:)?PD.+ must classify as pronoun PD, not adposition SPS.
		p := softNormalizeFreeLingPostagPart(part)
		if p == "" {
			continue
		}
		if softPostagPartIsOpen(p) {
			return false
		}
		if !softPostagPartIsClosed(p) {
			return false
		}
	}
	return true
}

func softNormalizePostagPart(p string) string {
	p = strings.ToUpper(strings.TrimSpace(p))
	p = strings.TrimPrefix(p, "^")
	p = strings.TrimSuffix(p, "$")
	p = strings.TrimPrefix(p, "(?:")
	p = strings.TrimPrefix(p, "(")
	p = strings.TrimSuffix(p, ")")
	return p
}

func softPostagPartIsOpen(p string) bool {
	// German STTS open classes (often COLON-separated: SUB:NOM:SIN:…)
	if strings.Contains(p, ":") {
		for _, open := range []string{"SUB", "EIG", "ADJ", "ADV", "PA1", "PA2", "VER", "ZUS", "TRUNC", "FM"} {
			if strings.HasPrefix(p, open) || strings.Contains(p, ":"+open) || strings.HasPrefix(p, open+":") {
				return true
			}
			// ADJ:PRD, VER:INF, etc.
			if strings.HasPrefix(p, open) {
				return true
			}
		}
		// patterns like (ADV:|ADJ:PRD:GRU).*
		for _, open := range []string{"ADJ", "ADV", "SUB", "VER", "PA1", "PA2", "EIG"} {
			if strings.Contains(p, open+":") || strings.Contains(p, open+".") {
				return true
			}
		}
	}
	// FreeLing Romance open classes (PT/ES/CA/GL tagset_PT): N*, A* (adj), R* (adv), V*
	// before English Penn — FreeLing VMP00 / NCMS000 / AQ0FS0 / RG, not Penn VB/NN alone.
	if len(p) > 0 {
		switch p[0] {
		case 'N':
			// NC / NP / N.+ / N[CP] — not German "NEG" etc. without FreeLing shape
			if strings.HasPrefix(p, "NC") || strings.HasPrefix(p, "NP") ||
				strings.HasPrefix(p, "N.") || strings.HasPrefix(p, "N[") || p == "N" {
				return true
			}
		case 'A':
			// AQ / AO / AP / A.+ — adjectives (not ART)
			if strings.HasPrefix(p, "AQ") || strings.HasPrefix(p, "AO") ||
				strings.HasPrefix(p, "AP") || strings.HasPrefix(p, "A.") ||
				strings.HasPrefix(p, "A[") || p == "A" {
				return true
			}
		case 'R':
			// RG / RN / RM / R. — FreeLing adverbs (not English RB already below)
			if strings.HasPrefix(p, "RG") || strings.HasPrefix(p, "RN") ||
				strings.HasPrefix(p, "RM") || strings.HasPrefix(p, "R.") ||
				strings.HasPrefix(p, "R[") || p == "R" {
				return true
			}
		case 'V':
			// VM* / VA* / VS* / VMP00 / V.+ — FreeLing verbs/participles
			if strings.HasPrefix(p, "VM") || strings.HasPrefix(p, "VA") ||
				strings.HasPrefix(p, "VS") || strings.HasPrefix(p, "V.") ||
				strings.HasPrefix(p, "V[") || p == "V" {
				return true
			}
		}
	}
	for _, open := range []string{"NN", "VB", "JJ", "RB", "CD", "FW", "UH", "SYM", "LS"} {
		if strings.HasPrefix(p, open) {
			return true
		}
	}
	return false
}

func softPostagPartIsClosed(p string) bool {
	// FreeLing Romance determiners (D.+, DA0MS0, DI0FS0, DP…, …) — PT/ES/CA/GL soft packs.
	if softFreeLingPartIsDeterminer(p) {
		return true
	}
	// FreeLing pronouns (PD / PI / PP / PR / PX / PT) and adpositions/conjunctions.
	if softFreeLingPartIsPronoun(p) || softFreeLingPartIsAdposition(p) || softFreeLingPartIsConjunction(p) {
		return true
	}
	// German STTS closed: ART, PRP (preposition!), PRO, KON, APPR, APPO, APZR, …
	if strings.Contains(p, ":") || strings.HasPrefix(p, "PRP") || strings.HasPrefix(p, "ART") ||
		strings.HasPrefix(p, "PRO") || strings.HasPrefix(p, "KON") || strings.HasPrefix(p, "APPR") ||
		strings.HasPrefix(p, "APPO") || strings.HasPrefix(p, "APZR") || strings.HasPrefix(p, "KOUI") ||
		strings.HasPrefix(p, "KOUS") || strings.HasPrefix(p, "KOKOM") {
		// German PRP is preposition, not pronoun — still closed-class
		if strings.HasPrefix(p, "PRP") || strings.Contains(p, "PRP:") || strings.Contains(p, "PRP.") {
			return true
		}
		for _, c := range []string{"ART", "PRO", "KON", "APPR", "APPO", "APZR", "KOUI", "KOUS", "KOKOM", "PTKZU", "PTKNEG", "PTKVZ", "PTKANT"} {
			if strings.HasPrefix(p, c) || strings.Contains(p, c+":") || strings.Contains(p, c+".") {
				return true
			}
		}
	}
	// English Penn: PRP$ before PRP; WDT/WP/WRB before W*
	for _, c := range []string{"PRP$", "PRP", "WDT", "WP$", "WP", "WRB", "PDT", "POS", "DT", "IN", "CC", "MD", "TO", "EX", "RP"} {
		if strings.HasPrefix(p, c) {
			return true
		}
	}
	return false
}

func softClosedClassSurfaceMatch(tag, surface string) bool {
	s := strings.ToLower(strings.TrimSpace(surface))
	if s == "" {
		return false
	}
	// Match if surface fits any closed-class alternative in the tag pattern.
	u := strings.ToUpper(tag)
	// German STTS often uses .* wildcards without | — treat whole tag as one part.
	parts := strings.Split(u, "|")
	for _, part := range parts {
		// Strip FreeLing optional contraction prefixes so (SPS00:)?PD.+ → PD.+
		p := softNormalizeFreeLingPostagPart(part)
		if p == "" {
			continue
		}
		if softPostagPartIsOpen(p) {
			continue
		}
		if softClosedPartSurface(p, s) {
			return true
		}
	}
	return false
}

func softClosedPartSurface(part, s string) bool {
	// FreeLing Romance D*/P*/S*/C* (PT/ES/CA soft packs without full dict).
	if softFreeLingPartIsDeterminer(part) {
		return softFreeLingDetSurface(part, s)
	}
	if softFreeLingPartIsPronoun(part) {
		return softFreeLingPronSurface(part, s)
	}
	if softFreeLingPartIsAdposition(part) {
		return softIsRomancePrepSurface(s)
	}
	if softFreeLingPartIsConjunction(part) {
		return softIsCC(s)
	}
	// --- German STTS (colon tags / case-tagged PRP / ART / PRO / APPR / KON) ---
	// Do NOT treat English Penn "PRP.*" (pronoun regex) as STTS: "PRP." matches
	// the start of "PRP.*" and would route "you"/"we" through the prep list.
	if softIsSTTSClosedTag(part) {
		// Prepositions (STTS PRP:… / PRP.*DAT.* / APPR) — "aus", "in", "im", "von", …
		if softIsSTTSPrepositionTag(part) || strings.Contains(part, "APPR") ||
			strings.Contains(part, "APPO") || strings.Contains(part, "APZR") {
			return softIsPreposition(s) || softIsGermanPrep(s)
		}
		if strings.Contains(part, "ART") {
			return softIsGermanArticle(s)
		}
		// STTS PRO (personal/demonstrative/relative); many DE rules use PRO:.+ where
		// the surface is also a definite article reading (der/die/das…), as in WEHREND.
		if softIsPronounTag(part) {
			return softIsPronoun(s) || softIsGermanArticle(s)
		}
		if strings.Contains(part, "KON") || strings.Contains(part, "KOU") || strings.Contains(part, "KOKOM") {
			return softIsCC(s) || softIsGermanConj(s)
		}
		if strings.Contains(part, "PTK") {
			return softLooksLikeWord(s)
		}
	}
	// --- English Penn ---
	switch {
	case strings.HasPrefix(part, "PRP"):
		// PRP / PRP$ / PRP.* / PRP.+ — pronouns (not STTS prep)
		return softIsPronoun(s)
	case strings.HasPrefix(part, "DT") || strings.HasPrefix(part, "PDT"):
		return softIsDeterminer(s)
	case strings.HasPrefix(part, "IN"):
		return softIsPreposition(s)
	case strings.HasPrefix(part, "MD"):
		return softIsModal(s)
	case strings.HasPrefix(part, "CC"):
		return softIsCC(s)
	case strings.HasPrefix(part, "TO"):
		return s == "to"
	case strings.HasPrefix(part, "EX"):
		return s == "there"
	case strings.HasPrefix(part, "WDT") || strings.HasPrefix(part, "WP") || strings.HasPrefix(part, "WRB"):
		return softIsWh(s)
	case strings.HasPrefix(part, "RP") || strings.HasPrefix(part, "POS"):
		// particles / possessive clitics: allow short words and 's
		return softLooksLikeWord(s) || s == "'s" || s == "\u2019s"
	case strings.Contains(part, "PCT") || strings.Contains(part, "PUNC") || strings.Contains(part, "PKT"):
		return softLooksLikePunct(s)
	default:
		return false
	}
}

// softIsSTTSClosedTag is true for German STTS closed-class tag patterns, not
// English Penn PRP/PRP$ / PRP.* pronoun tags.
func softIsSTTSClosedTag(part string) bool {
	if strings.Contains(part, "ART") || strings.Contains(part, "APPR") ||
		strings.Contains(part, "APPO") || strings.Contains(part, "APZR") ||
		strings.Contains(part, "KON") || strings.Contains(part, "KOU") ||
		strings.Contains(part, "KOKOM") || strings.Contains(part, "PTK") {
		return true
	}
	if softIsPronounTag(part) {
		return true
	}
	return softIsSTTSPrepositionTag(part)
}

// softIsPronounTag: PRO/PRON families (not English/STTS PRP*).
func softIsPronounTag(part string) bool {
	if strings.HasPrefix(part, "PRP") {
		return false
	}
	// PRON… / PRO… / PRO:… / PRO.… (includes DA pron:.* uppercased to PRON:.*)
	return strings.HasPrefix(part, "PRON") || strings.HasPrefix(part, "PRO") ||
		strings.Contains(part, "PRO:") || strings.Contains(part, "PRO.")
}

// softIsSTTSPrepositionTag: German STTS preposition patterns on PRP.
// English Penn uses PRP / PRP$ / PRP.* for pronouns — those must NOT match here.
// STTS prepositions look like PRP:DAT:…, PRP.*DAT.*, PRP.AKK, …
func softIsSTTSPrepositionTag(part string) bool {
	if !strings.HasPrefix(part, "PRP") {
		return false
	}
	// English: PRP, PRP$, PRP$?, PRP.*, PRP.+
	if part == "PRP" || strings.HasPrefix(part, "PRP$") {
		return false
	}
	if strings.HasPrefix(part, "PRP.*") || strings.HasPrefix(part, "PRP.+") {
		// Only STTS if a case feature is present (PRP.*DAT.*)
		return strings.Contains(part, "DAT") || strings.Contains(part, "AKK") ||
			strings.Contains(part, "GEN") || strings.Contains(part, "NOM") ||
			strings.Contains(part, ":")
	}
	// STTS: PRP:DAT:SIN, PRP.DAT, …
	return strings.Contains(part, ":") ||
		strings.Contains(part, "DAT") || strings.Contains(part, "AKK") ||
		strings.Contains(part, "GEN") || strings.Contains(part, "NOM") ||
		// PRP.<letter> case (not PRP.* quantifier)
		(strings.Contains(part, "PRP.") && !strings.Contains(part, "PRP.*") && !strings.Contains(part, "PRP.+"))
}

func softIsGermanPrep(s string) bool {
	switch s {
	case "aus", "außer", "bei", "beim", "bis", "durch", "entlang", "für", "gegen",
		"gegenüber", "ohne", "um", "wider", "an", "am", "auf", "hinter", "in", "im", "ins",
		"neben", "über", "unter", "vor", "vom", "zwischen", "zu", "zum", "zur", "von",
		"nach", "mit", "seit", "während", "wegen", "trotz", "dank", "laut", "gemäß",
		"binnen", "entgegen", "entsprechend", "nahe", "nebst", "samt", "per", "pro",
		"via", "inklusive", "exklusive", "betreffs", "mangels", "mittels", "zwecks",
		"diesseits", "jenseits", "abseits", "außerhalb", "innerhalb", "oberhalb", "unterhalb":
		return true
	default:
		return false
	}
}

// softPOSIsContextTag: underscore-prefixed LT context tags (_CONTEXTO_…, _PUNCT…).
func softPOSIsContextTag(pos *string) bool {
	return pos != nil && strings.HasPrefix(*pos, "_")
}

// softPostagIsFreeLingDeterminer: FreeLing D.+ / DA0MS0 / DI… / mixed D[AI].+|NC.+.
// Mirrors FreeLing determiner category (tagset_PT: D + type A/D/E/I/T/N/P).
func softPostagIsFreeLingDeterminer(tag string) bool {
	u := strings.ToUpper(strings.TrimSpace(tag))
	if u == "" {
		return false
	}
	for _, part := range strings.Split(u, "|") {
		if softFreeLingPartIsDeterminer(softNormalizeFreeLingPostagPart(part)) {
			return true
		}
	}
	// Whole-pattern scan for (SPS00:)?D[AI].+ style wrappers.
	return softFreeLingPatternHasDeterminer(u)
}

// softFreeLingClosedSurfaceMatch: soft-accept FreeLing closed-class POS patterns
// (D*, PD*, PP*, …) when the surface is a known FreeLing closed-class form.
// Used both for untagged soft probes and when the only real reading is Z0MS0
// (numeral disambig) while Java also has DA0MS0/DP…/DI… from portuguese.dict.
func softFreeLingClosedSurfaceMatch(pattern, surface string) bool {
	u := strings.ToUpper(strings.TrimSpace(pattern))
	if u == "" || surface == "" {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(surface))
	// Prefer type-specific lists when the pattern is a single FreeLing family.
	for _, part := range strings.Split(u, "|") {
		p := softNormalizeFreeLingPostagPart(part)
		if softFreeLingPartIsDeterminer(p) && softFreeLingDetSurface(p, s) {
			return true
		}
		if softFreeLingPartIsPronoun(p) && softFreeLingPronSurface(p, s) {
			return true
		}
		if softFreeLingPartIsAdposition(p) && softIsRomancePrepSurface(s) {
			return true
		}
		if softFreeLingPartIsConjunction(p) && softIsCC(s) {
			return true
		}
	}
	// Mixed patterns like (SPS00:)?D[AI].+|NC.+|AQ.+ — accept D-family surfaces.
	if softFreeLingPatternHasDeterminer(u) && softIsRomanceDeterminerSurface(s) {
		return true
	}
	if softFreeLingPatternHasPronoun(u) && softIsRomancePronounSurface(s) {
		return true
	}
	return false
}

// softNormalizeFreeLingPostagPart strips optional FreeLing contraction wrappers
// used in PT/ES/CA grammar: (SPS00:)?D[AI].+ → D[AI].+
func softNormalizeFreeLingPostagPart(p string) string {
	p = softNormalizePostagPart(p)
	// After softNormalizePostagPart, "(SPS00:)?" becomes "SPS00:)?"
	for _, pref := range []string{
		"SPS00:)?", "SPS00:?", "SPS00:",
		"SP.+:)?", "SP.+:?", "SP.+:",
		"SP.:)?", "SP.:?",
	} {
		if strings.HasPrefix(p, pref) {
			p = p[len(pref):]
		}
	}
	// Leading "?:" leftovers from non-capturing groups
	p = strings.TrimPrefix(p, "?:")
	p = strings.TrimPrefix(p, "?")
	return p
}

func softFreeLingPartIsDeterminer(p string) bool {
	if p == "" {
		return false
	}
	// Pattern fragments D.+ / D[AI].+ / D
	if p == "D" || strings.HasPrefix(p, "D.") || strings.HasPrefix(p, "D[") || strings.HasPrefix(p, "D.+") {
		return true
	}
	// Concrete FreeLing: DA0MS0, DI0FS0, DD0CP0, DE0CN0, DP…, DT…, DN…
	// Not STTS DAT… (colon case tags handled elsewhere).
	if len(p) >= 2 && p[0] == 'D' {
		switch p[1] {
		case 'A', 'I', 'D', 'E', 'P', 'T', 'N':
			// Avoid bare "DAT" as FreeLing — DAT is STTS dative
			if strings.HasPrefix(p, "DAT") && (len(p) == 3 || p[3] == ':' || p[3] == '.') {
				return false
			}
			return true
		}
	}
	return false
}

func softFreeLingPartIsPronoun(p string) bool {
	if p == "" {
		return false
	}
	// Not English Penn PDT / POS / PRP / PRP$ (those use softIsDeterminer / softIsPronoun).
	if strings.HasPrefix(p, "PRP") || strings.HasPrefix(p, "PDT") || strings.HasPrefix(p, "POS") {
		return false
	}
	// FreeLing P + type: PD PI PP PR PX PT PE PN (tagset_PT pronoun)
	if len(p) >= 2 && p[0] == 'P' {
		switch p[1] {
		case 'D', 'I', 'P', 'R', 'X', 'T', 'E', 'N', '.', '[', '+', '*':
			return true
		}
	}
	return p == "P"
}

func softFreeLingPartIsAdposition(p string) bool {
	// FreeLing SPS00 / SP.* (prepositions); not English.
	return strings.HasPrefix(p, "SPS") || strings.HasPrefix(p, "SP.") ||
		strings.HasPrefix(p, "SP[") || p == "SP" || p == "SPS00"
}

func softFreeLingPartIsConjunction(p string) bool {
	// FreeLing CC / CS (not English CC alone — still fine: softIsCC handles EN+Romance)
	return p == "CC" || p == "CS" || strings.HasPrefix(p, "CC.") || strings.HasPrefix(p, "CS.") ||
		strings.HasPrefix(p, "C[") || (p == "C")
}

// softFreeLingPatternHasDeterminer scans full postag patterns for a D-family segment
// (including mixed | and optional (SPS00:)? prefixes).
func softFreeLingPatternHasDeterminer(u string) bool {
	for i := 0; i < len(u); i++ {
		if u[i] != 'D' {
			continue
		}
		if i > 0 {
			prev := u[i-1]
			// Segment start after | ( : ? ) 
			if prev != '|' && prev != '(' && prev != ':' && prev != '?' {
				continue
			}
		}
		if i+1 >= len(u) {
			return true
		}
		switch u[i+1] {
		case 'A', 'I', 'D', 'E', 'P', 'T', 'N', '.', '[', '*', '+', '?':
			return true
		}
	}
	return false
}

func softFreeLingPatternHasPronoun(u string) bool {
	for i := 0; i < len(u); i++ {
		if u[i] != 'P' {
			continue
		}
		if i > 0 {
			prev := u[i-1]
			if prev != '|' && prev != '(' && prev != ':' && prev != '?' {
				continue
			}
		}
		if i+1 >= len(u) {
			return true
		}
		switch u[i+1] {
		case 'D', 'I', 'P', 'R', 'X', 'T', 'E', 'N', '.', '[', '*', '+', '?':
			// Avoid matching Penn PDT / POS / PRP as FreeLing P-only via PDT prefix —
			// PDT is determiner (already handled); PRP is English pronoun (other path).
			if strings.HasPrefix(u[i:], "PRP") || strings.HasPrefix(u[i:], "PDT") ||
				strings.HasPrefix(u[i:], "POS") {
				continue
			}
			return true
		}
	}
	return false
}

// softFreeLingDetSurface matches FreeLing determiner type to surface lists
// (tagset_PT type: A article, D demonstrative, I indefinite, P possessive, …).
func softFreeLingDetSurface(part, s string) bool {
	// Type-specific when pattern pins a subtype; broad D.+ accepts any.
	switch {
	case strings.HasPrefix(part, "DP"):
		return softIsRomancePossessiveSurface(s)
	case strings.HasPrefix(part, "DI"):
		return softIsRomanceIndefiniteDetSurface(s)
	case strings.HasPrefix(part, "DD"):
		return softIsRomanceDemonstrativeSurface(s)
	case strings.HasPrefix(part, "DA"):
		return softIsRomanceArticleSurface(s)
	default:
		// D.+ / D[AI].+ / D[AIP].+ / mixed — any FreeLing determiner surface
		return softIsRomanceDeterminerSurface(s)
	}
}

func softFreeLingPronSurface(part, s string) bool {
	switch {
	case strings.HasPrefix(part, "PD"):
		return softIsRomanceDemonstrativeSurface(s)
	case strings.HasPrefix(part, "PI"):
		return softIsRomanceIndefiniteDetSurface(s) || softIsRomanceIndefinitePronSurface(s)
	case strings.HasPrefix(part, "PP"):
		return softIsPronoun(s)
	default:
		return softIsRomancePronounSurface(s)
	}
}

// softIsRomanceDeterminerSurface: any FreeLing D* surface (article/demo/indef/poss).
func softIsRomanceDeterminerSurface(s string) bool {
	return softIsRomanceArticleSurface(s) ||
		softIsRomanceDemonstrativeSurface(s) ||
		softIsRomanceIndefiniteDetSurface(s) ||
		softIsRomancePossessiveSurface(s) ||
		softIsDeterminer(s)
}

func softIsRomanceArticleSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "o", "a", "os", "as", "um", "uma", "uns", "umas",
		"el", "la", "los", "las", "un", "una", "unos", "unas",
		"lo", "les", "els", "na", "nas", "ao", "à", "aos", "às",
		// FreeLing fused prep+article (SPS00:DA…): do/da/no/na/pelo…
		"do", "da", "dos", "das", "no", "nos",
		"pelo", "pela", "pelos", "pelas",
		"dum", "duma", "duns", "dumas", "num", "numa", "nuns", "numas":
		return true
	default:
		return false
	}
}

func softIsRomanceDemonstrativeSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	// PT/ES/GL demonstratives (DD / PD)
	case "este", "esta", "estes", "estas", "esse", "essa", "esses", "essas",
		"aquele", "aquela", "aqueles", "aquelas", "isto", "isso", "aquilo",
		"estos", "ese", "esa", "esos", "esas",
		"aquel", "aquella", "aquellos", "aquellas",
		// FreeLing fused prep+demonstrative (SPS00:DD…): dessa/nesta/…
		// Java portuguese.dict tags "dessa" as SPS00:DD0FS0, "nesta" as SPS00:DD0FS0.
		"deste", "desta", "destes", "destas", "desse", "dessa", "desses", "dessas",
		"daquele", "daquela", "daqueles", "daquelas",
		"neste", "nesta", "nestes", "nestas", "nesse", "nessa", "nesses", "nessas",
		"naquele", "naquela", "naqueles", "naquelas",
		"àquele", "àquela", "àqueles", "àquelas", "aeste", "aesta",
		// Catalan
		"aquest", "aquesta", "aquestos", "aquestes", "aquell", "aquells", "aquelles",
		"això", "allò", "açò":
		return true
	default:
		return false
	}
}

// softIsRomancePossessiveSurface: FreeLing DP (meu/teu/seu/…); Java portuguese.dict.
func softIsRomancePossessiveSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "meu", "minha", "meus", "minhas",
		"teu", "tua", "teus", "tuas",
		"seu", "sua", "seus", "suas",
		"nosso", "nossa", "nossos", "nossas",
		"vosso", "vossa", "vossos", "vossas",
		// Spanish possessives
		"mi", "mis", "tu", "tus", "su", "sus",
		"nuestro", "nuestra", "nuestros", "nuestras",
		"vuestro", "vuestra", "vuestros", "vuestras",
		// Catalan possessives (meua/seua; meu/teu/seu shared with PT)
		"meua", "meues", "teua", "teues", "seua", "seues",
		"nostre", "nostra", "nostres", "vostra", "vostre", "vostres":
		return true
	default:
		return false
	}
}

// softIsRomanceIndefiniteDetSurface: FreeLing DI (qualquer, algum, …).
func softIsRomanceIndefiniteDetSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "qualquer", "quaisquer",
		"algum", "alguma", "alguns", "algumas",
		"nenhum", "nenhuma", "nenhuns", "nenhumas",
		"todo", "toda", "todos", "todas",
		"outro", "outra", "outros", "outras",
		"cada", "certo", "certa", "certos", "certas",
		"vário", "vária", "vários", "várias",
		"muito", "muita", "muitos", "muitas",
		"pouco", "pouca", "poucos", "poucas",
		"tanto", "tanta", "tantos", "tantas",
		"quanto", "quanta", "quantos", "quantas",
		"sendos", "sendas",
		// Spanish/Catalan common DI
		"algún", "alguna", "algunos", "algunas",
		"ningún", "ninguna", "ningunos", "ningunas",
		"otro", "otra", "otros", "otras",
		"varios", "varias",
		"cualquier", "cualquiera", "cualesquiera",
		"tot", "tota", "tots", "totes", "altre", "altra", "altres":
		return true
	default:
		return false
	}
}

func softIsRomanceIndefinitePronSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "alguém", "ninguém", "tudo", "nada", "algo", "outrem",
		"alguien", "nadie",
		"algú", "ningú", "res":
		return true
	default:
		return false
	}
}

func softIsRomancePronounSurface(s string) bool {
	return softIsPronoun(s) ||
		softIsRomanceDemonstrativeSurface(s) ||
		softIsRomanceIndefinitePronSurface(s) ||
		softIsRomanceIndefiniteDetSurface(s)
}

func softIsRomancePrepSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "a", "ante", "após", "até", "com", "contra", "de", "desde", "em", "entre",
		"para", "perante", "por", "sem", "sob", "sobre", "trás",
		"ao", "à", "aos", "às", "do", "da", "dos", "das", "no", "na", "nos", "nas",
		"pelo", "pela", "pelos", "pelas", "dum", "duma", "duns", "dumas",
		"num", "numa", "nuns", "numas", "pra", "pro",
		// Spanish/Catalan
		"en", "con", "sin", "hacia", "según", "durante", "mediante", "versus",
		"amb", "sense", "cap", "des", "dins", "fora", "per", "segons":
		return true
	default:
		return softIsPreposition(s)
	}
}

func softIsGermanArticle(s string) bool {
	switch s {
	case "der", "die", "das", "den", "dem", "des",
		"ein", "eine", "einen", "einem", "einer", "eines",
		"kein", "keine", "keinen", "keinem", "keiner", "keines":
		return true
	default:
		return softIsDeterminer(s)
	}
}

func softIsGermanConj(s string) bool {
	switch s {
	case "und", "oder", "aber", "denn", "sondern", "doch", "sowie",
		"weil", "dass", "daß", "ob", "wenn", "als", "wie", "indem",
		"während", "obwohl", "bevor", "nachdem", "seit", "seitdem",
		"sobald", "solange", "falls", "sofern", "damit", "sodass", "so daß":
		return true
	default:
		return false
	}
}

func softIsPronoun(s string) bool {
	switch s {
	// English
	case "i", "me", "my", "mine", "myself",
		"you", "your", "yours", "yourself", "yourselves",
		"he", "him", "his", "himself",
		"she", "her", "hers", "herself",
		"it", "its", "itself",
		"we", "us", "our", "ours", "ourselves",
		"they", "them", "their", "theirs", "themselves",
		"thou", "thee", "thy", "thine", "ye",
		// Portuguese personal (CONFUSÃO_TER_ESTAR PP.+)
		"eu", "tu", "ele", "ela", "nós", "vós", "eles", "elas",
		"mim", "ti", "si", "lhe", "lhes", "nos", "vos",
		"meu", "minha", "meus", "minhas", "teu", "tua", "teus", "tuas",
		"seu", "sua", "seus", "suas", "nosso", "nossa", "nossos", "nossas",
		// German (PRO:REF / personal)
		"ich", "mich", "mir", "mein", "meine", "meiner", "meinem", "meinen",
		"du", "dich", "dir", "dein", "deine", "deiner", "deinem", "deinen",
		"er", "ihn", "ihm", "sein", "seine", "seiner", "seinem", "seinen",
		"sie", "ihr", "ihre", "ihrer", "ihrem", "ihren", "ihnen",
		"es", "wir", "uns", "unser", "unsere", "unserer", "unserem", "unseren",
		"euch", "euer", "eure", "sich",
		// Irish (Pron.* — BHEAS_P etc.)
		"mé", "tú", "é", "í", "sinn", "sibh", "siad", "muid",
		"mise", "tusa", "eisean", "ise", "sinne", "sibhse", "siadsan":
		return true
	default:
		return false
	}
}

func softIsDeterminer(s string) bool {
	switch s {
	case "a", "an", "the", "this", "that", "these", "those",
		"some", "any", "no", "every", "each", "either", "neither",
		"all", "both", "half", "many", "much", "few", "several",
		"another", "other", "such":
		return true
	default:
		return false
	}
}

func softIsPreposition(s string) bool {
	switch s {
	case "in", "on", "at", "to", "for", "of", "with", "by", "from", "as",
		"into", "onto", "upon", "about", "over", "under", "after", "before",
		"between", "among", "through", "during", "without", "within", "against",
		"across", "behind", "beyond", "despite", "except", "inside", "outside",
		"toward", "towards", "until", "via", "per", "than", "like", "near",
		"off", "out", "up", "down", "around", "along", "since", "if", "whether",
		"while", "because", "although", "though", "unless", "whereas",
		"regarding", "concerning", "including", "excluding", "following",
		"according", "depending", "considering", "given", "versus", "amid", "amidst":
		return true
	default:
		return false
	}
}

func softIsModal(s string) bool {
	switch s {
	case "can", "could", "may", "might", "must", "shall", "should", "will", "would":
		return true
	default:
		return false
	}
}

func softIsCC(s string) bool {
	switch s {
	case "and", "or", "but", "nor", "yet", "so", "plus", "&",
		// Catalan / Spanish coordinating conjunctions (FreeLing CC)
		"i", "o", "ni", "però", "pero", "sinó", "sino", "y", "e", "u":
		return true
	default:
		return false
	}
}

func softIsWh(s string) bool {
	switch s {
	case "what", "which", "who", "whom", "whose", "where", "when", "why", "how",
		"whatever", "whichever", "whoever", "whomever",
		"wherever", "whenever", "however", "whyever":
		return true
	default:
		return false
	}
}

// softLooksGermanAdjSurface is true when surface likely needs DE adj stem stripping
// (umlaut/ß or long -isch/-lich endings). Avoids O(n) stem work on Romance soft packs.
func softLooksGermanAdjSurface(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch r {
		case 'ä', 'ö', 'ü', 'ß', 'Ä', 'Ö', 'Ü':
			return true
		}
	}
	low := strings.ToLower(s)
	for _, suf := range []string{"ischen", "lichem", "licher", "liches", "liche", "isch"} {
		if strings.HasSuffix(low, suf) && len(low) > len(suf)+3 {
			return true
		}
	}
	return false
}

// softGermanAdjCandidates yields lemma-like forms by stripping German adj endings.
// Used for inflected+regexp soft tokens (Steigende ↔ steigend?).
func softGermanAdjCandidates(surface string) []string {
	if surface == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(surface)
	low := strings.ToLower(surface)
	add(low)
	cur := low
	// longest endings first
	for _, suf := range []string{"em", "en", "er", "es", "e", "n", "s", "d"} {
		if strings.HasSuffix(cur, suf) && len([]rune(cur)) > len([]rune(suf))+3 {
			cur = cur[:len(cur)-len(suf)]
			add(cur)
		}
	}
	// one more pass from full lower (e.g. lateinischen → lateinisch)
	cur = low
	for _, suf := range []string{"ischen", "lichem", "licher", "liches", "liche", "isch", "em", "en", "er", "es", "e"} {
		if strings.HasSuffix(cur, suf) && len(cur) > len(suf)+3 {
			add(cur[:len(cur)-len(suf)])
			if strings.HasSuffix(suf, "en") || strings.HasSuffix(suf, "e") {
				// lateinisch from lateinischen
				stem := cur[:len(cur)-len(suf)]
				if !strings.HasSuffix(stem, "isch") && strings.Contains(low, "isch") {
					// already handled by ischen
				}
				add(stem + "isch")
			}
		}
	}
	return out
}

// softFrenchErInflected maps common -er verb surfaces to infinitive without a tagger.
// Examples: placé/placer, places/placer, rencontré/rencontrer, opère/opérer (é/è fold).
func softFrenchErInflected(surface, base string) bool {
	s, b := strings.ToLower(surface), strings.ToLower(base)
	if !strings.HasSuffix(b, "er") || len(b) < 4 {
		return false
	}
	stem := b[:len(b)-2] // placer → plac
	if len(stem) < 3 {
		return false
	}
	sf := softFrenchAccentFold(s)
	// present / participle endings on the stem
	for _, suf := range []string{"é", "ée", "és", "ées", "è", "ès", "e", "es", "ent", "ons", "ez", "ant", "ais", "ait", "aient", "ai", "as", "a", "âmes", "âtes", "èrent"} {
		if s == stem+suf || sf == softFrenchAccentFold(stem+suf) {
			return true
		}
	}
	// ç- variants (plaçons)
	if strings.HasSuffix(stem, "c") {
		ced := stem[:len(stem)-1] + "ç"
		for _, suf := range []string{"ons", "ait", "ais", "aient"} {
			if s == ced+suf || sf == softFrenchAccentFold(ced+suf) {
				return true
			}
		}
	}
	return false
}

// softFrenchAccentFold maps accented vowels to ASCII for soft stem compares (opère/opérer).
var softFrenchAccentReplacer = strings.NewReplacer(
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"à", "a", "â", "a", "ä", "a",
	"ù", "u", "û", "u", "ü", "u",
	"ô", "o", "ö", "o",
	"î", "i", "ï", "i",
	"ç", "c",
	"É", "e", "È", "e", "Ê", "e",
)

func softFrenchAccentFold(s string) string {
	return softFrenchAccentReplacer.Replace(strings.ToLower(s))
}

// softPortugueseStripEnclitic removes PT ênclise clitics after hyphen or fused
// (Esqueceu-se → esqueceu, distrai-los kept whole by tokenizer then stripped).
func softPortugueseStripEnclitic(s string) string {
	s = strings.ToLower(s)
	// hyphenated clitics: esqueceu-se, puxa-las
	if i := strings.LastIndex(s, "-"); i > 0 {
		cl := s[i+1:]
		switch cl {
		case "se", "me", "te", "lhe", "lhes", "nos", "vos",
			"lo", "la", "los", "las", "o", "a", "os", "as", "mo", "to", "lho":
			return s[:i]
		}
	}
	return s
}

// softSpanishStripEnclitic removes common Spanish object clitics from the end
// of gerunds/infinitives (asiéndolos → asiendo, hacerlo → hacer).
func softSpanishStripEnclitic(s string) string {
	s = strings.ToLower(s)
	for _, suf := range []string{
		"noslos", "noslas", "selos", "selas", "melos", "melas", "telos", "telas",
		"noslo", "nosla", "selo", "sela", "melo", "mela", "telo", "tela",
		"los", "las", "les", "nos", "me", "te", "se", "lo", "la", "le", "os",
	} {
		if strings.HasSuffix(s, suf) {
			core := s[:len(s)-len(suf)]
			if len([]rune(core)) >= 4 {
				return core
			}
		}
	}
	return s
}

// softFrenchElisionMatch is true when surface is a French elided form of base
// (or equal), as produced by FrenchWordTokenizer (d’électricité → d’ + électricité).
func softFrenchElisionMatch(surface, base string) bool {
	s := strings.ToLower(normalizeApostrophes(surface))
	b := strings.ToLower(normalizeApostrophes(base))
	if s == b {
		return true
	}
	switch b {
	case "de":
		return s == "d'"
	case "le", "la":
		return s == "l'"
	case "je":
		return s == "j'"
	case "me":
		return s == "m'"
	case "te":
		return s == "t'"
	case "se":
		return s == "s'"
	case "ne":
		return s == "n'"
	case "que":
		return s == "qu'"
	case "ce":
		return s == "c'" || s == "ç'"
	default:
		return false
	}
}

// softFrenchErLemmaCandidates yields likely -er infinitives for a surface form
// (désactivé → désactiver) so RE patterns like .*er match without a tagger.
func softFrenchErLemmaCandidates(surface string) []string {
	s := strings.ToLower(strings.TrimSpace(surface))
	if s == "" {
		return nil
	}
	for _, suf := range []string{"ées", "és", "ée", "é"} {
		if strings.HasSuffix(s, suf) && len([]rune(s)) > len([]rune(suf))+2 {
			stem := s[:len(s)-len(suf)]
			return []string{stem + "er"}
		}
	}
	return nil
}

// softGermanGeParticiple approximates ge- + stem + (e)t/en ↔ infinitive …en.
// Examples: gemacht/machen, gelernt/lernen, genommen/nehmen (strong ge-…en).
// Separable prefixes: ausgelost/auslosen, angefangen/anfangen.
func softGermanGeParticiple(surface, base string) bool {
	s, b := strings.ToLower(surface), strings.ToLower(base)
	if softGermanGeParticipleCore(s, b) {
		return true
	}
	// Separable verb prefixes (common set used in LT DE soft packs).
	for _, pref := range []string{
		"aus", "ein", "an", "auf", "ab", "zu", "mit", "vor", "nach", "bei",
		"her", "hin", "weg", "fest", "klar", "los", "weiter", "zurück", "zusammen",
		"durch", "über", "unter", "um", "wider", "fort", "dar", "entgegen",
	} {
		if strings.HasPrefix(s, pref) && strings.HasPrefix(b, pref) && len(s) > len(pref)+3 && len(b) > len(pref)+2 {
			if softGermanGeParticipleCore(s[len(pref):], b[len(pref):]) {
				return true
			}
		}
	}
	return false
}

func softGermanGeParticipleCore(s, b string) bool {
	if !strings.HasPrefix(s, "ge") || len(s) < 5 || len(b) < 3 {
		return false
	}
	core := s[2:]
	// strip participle endings
	for _, suf := range []string{"en", "et", "t", "n"} {
		if strings.HasSuffix(core, suf) && len(core) > len(suf)+2 {
			core = core[:len(core)-len(suf)]
			break
		}
	}
	bcore := b
	for _, suf := range []string{"en", "n", "ern", "eln"} {
		if strings.HasSuffix(bcore, suf) && len(bcore) > len(suf)+2 {
			bcore = bcore[:len(bcore)-len(suf)]
			break
		}
	}
	if len(core) < 3 || len(bcore) < 3 {
		return false
	}
	if core == bcore {
		return true
	}
	// nehmen → nomm in genommen (consonant change) — shared stem ≥3
	return softSharedStemMatch(core, bcore)
}

// softIrregularLemma maps common irregular surfaces → possible dictionary lemmas
// for soft MatchInflected without a tagger. Values are multi-lemma because the
// same surface can map to different lemmas across languages (va→aller|ir|dir).
var softIrregularLemma = map[string][]string{
	// English
	"am": {"be"}, "is": {"be"}, "are": {"be"}, "was": {"be"}, "were": {"be"}, "been": {"be"}, "being": {"be"},
	"has": {"have", "haber", "haver"}, "had": {"have"}, "having": {"have"},
	"does": {"do"}, "did": {"do"}, "done": {"do"}, "doing": {"do"},
	// English clitics (it's / he's …) for HAD_HARD etc.
	"'s": {"be", "have"}, "’s": {"be", "have"},
	// Common EN irregulars used by upstream soft packs (came/come, going/go, …)
	"came": {"come"}, "comes": {"come"}, "coming": {"come"},
	"went": {"go"}, "goes": {"go"}, "going": {"go"}, "gone": {"go"},
	"ate": {"eat"}, "eats": {"eat"}, "eating": {"eat"},
	"took": {"take"}, "takes": {"take"}, "taking": {"take"},
	"made": {"make"}, "makes": {"make"}, "making": {"make"},
	"drove": {"drive"}, "drives": {"drive"}, "driving": {"drive"},
	"applied": {"apply"}, "applies": {"apply"}, "applying": {"apply"},
	"participated": {"participate"}, "participates": {"participate"}, "participating": {"participate"},
	"threw": {"throw"}, "throws": {"throw"}, "throwing": {"throw"},
	// Modal stems (could←can, would←will) for NIT_NOT / BE_WILL soft packs
	"could": {"can"}, "would": {"will"},
	"ging": {"gehen"}, "ginge": {"gehen"}, "gegangen": {"gehen"},
	"gingst": {"gehen"}, "gingen": {"gehen"},
	// French être / avoir / aller / faire
	"suis": {"être"}, "es": {"être"}, "est": {"être"}, "sommes": {"être"}, "êtes": {"être"}, "sont": {"être"},
	"étais": {"être"}, "était": {"être"}, "étions": {"être"}, "étiez": {"être"}, "étaient": {"être"},
	"été": {"être"}, "étant": {"être"}, "sera": {"être"}, "serai": {"être"}, "seras": {"être"}, "seront": {"être"},
	// "a" = 3sg present (y-a-t-il); keep with other avoir forms
	"a": {"avoir"}, "ai": {"avoir"}, "as": {"avoir"}, "avons": {"avoir"}, "avez": {"avoir"}, "ont": {"avoir"},
	"avais": {"avoir"}, "avait": {"avoir"}, "avaient": {"avoir"}, "eu": {"avoir"}, "ayant": {"avoir"},
	// venir / venir forms (CONFUSION_OU viennent)

	// valoir / falloir / pouvoir / devoir (FR soft residuals)
	"vaut": {"valoir"}, "valu": {"valoir"}, "valait": {"valoir"}, "valaient": {"valoir"}, "valons": {"valoir"}, "valez": {"valoir"}, "valent": {"valoir"},
	"faut": {"falloir"}, "fallait": {"falloir"}, "faudra": {"falloir"}, "fallu": {"falloir"},
	"peux": {"pouvoir"}, "peut": {"pouvoir"}, "pouvons": {"pouvoir"}, "pouvez": {"pouvoir"}, "peuvent": {"pouvoir"},
	"pouvais": {"pouvoir"}, "pouvait": {"pouvoir"}, "pouvaient": {"pouvoir"}, "pu": {"pouvoir"},
	"dois": {"devoir"}, "doit": {"devoir"}, "devons": {"devoir"}, "devez": {"devoir"}, "doivent": {"devoir"},
	"devais": {"devoir"}, "devait": {"devoir"}, "devaient": {"devoir"}, "devrais": {"devoir"}, "devrait": {"devoir"}, "devrions": {"devoir"}, "devriez": {"devoir"}, "devraient": {"devoir"},
	"dû": {"devoir"}, "due": {"devoir"}, "dus": {"devoir"}, "dues": {"devoir"},
	// aller future/conditional (ira, irai…)
	"irai": {"aller"}, "iras": {"aller"}, "ira": {"aller"}, "irons": {"aller"}, "irez": {"aller"}, "iront": {"aller"},
	"irais": {"aller"}, "irait": {"aller"}, "irions": {"aller"}, "iriez": {"aller"}, "iraient": {"aller"},
	"viens": {"venir"},
	"vient": {"venir"}, "venons": {"venir"}, "venez": {"venir"}, "viennent": {"venir"},
	"venait": {"venir"}, "venaient": {"venir"}, "venu": {"venir"}, "venue": {"venir"}, "venus": {"venir"}, "venues": {"venir"},
	"allons": {"aller"}, "allez": {"aller"}, "vont": {"aller"},
	"allait": {"aller"}, "allaient": {"aller"}, "allé": {"aller"}, "allée": {"aller"}, "allés": {"aller"},
	"fais": {"faire"}, "fait": {"faire"}, "faisons": {"faire"}, "faites": {"faire"}, "font": {"faire"},
	"faisait": {"faire"}, "faisaient": {"faire"},
	"fit": {"faire"}, "firent": {"faire"}, "faisant": {"faire"},
	// French mettre / prendre / partir / passer (+ common -er past forms)
	"mets": {"mettre"}, "met": {"mettre"}, "mettons": {"mettre"}, "mettez": {"mettre"}, "mettent": {"mettre"},
	"mis": {"mettre"}, "mise": {"mettre"}, "mises": {"mettre"}, "mettant": {"mettre"},
	"prends": {"prendre"}, "prend": {"prendre"}, "prenons": {"prendre"}, "prenez": {"prendre"}, "prennent": {"prendre"},
	"pris": {"prendre"}, "prise": {"prendre"}, "prises": {"prendre"}, "prenant": {"prendre"},
	"pars": {"partir"}, "part": {"partir"}, "partons": {"partir"}, "partez": {"partir"}, "partent": {"partir"},
	"parti": {"partir"}, "partie": {"partir"}, "partis": {"partir"}, "partant": {"partir"},
	"passe": {"passer"}, "passes": {"passer"}, "passons": {"passer"}, "passez": {"passer"}, "passent": {"passer"},
	"passé": {"passer"}, "passée": {"passer"}, "passés": {"passer"}, "passant": {"passer"},
	"place": {"placer"}, "places": {"placer"}, "plaçons": {"placer"}, "placé": {"placer"}, "placée": {"placer"},
	"tire": {"tirer"}, "tires": {"tirer"}, "tirons": {"tirer"}, "tiré": {"tirer"}, "tirée": {"tirer"},
	"garde": {"garder"}, "gardes": {"garder"}, "gardons": {"garder"}, "gardé": {"garder"}, "gardée": {"garder"},
	"loge": {"loger"}, "loges": {"loger"}, "logé": {"loger"}, "logée": {"loger"},
	"remplis": {"remplir"}, "remplit": {"remplir"}, "remplissons": {"remplir"}, "rempli": {"remplir"}, "remplie": {"remplir"},
	"achète": {"acheter"}, "achètes": {"acheter"}, "achetons": {"acheter"}, "acheté": {"acheter"}, "achetée": {"acheter"},
	"ajoute": {"ajouter"}, "ajoutes": {"ajouter"}, "ajoutons": {"ajouter"}, "ajouté": {"ajouter"}, "ajoutée": {"ajouter"},
	// German sein / haben + common strong/weak forms used in soft packs
	"bin": {"sein"}, "bist": {"sein"}, "ist": {"sein"}, "sind": {"sein"}, "seid": {"sein"},
	"war": {"sein"}, "warst": {"sein"}, "waren": {"sein"}, "wart": {"sein"}, "gewesen": {"sein"}, "sei": {"sein"},
	"habe": {"haben"}, "hast": {"haben"}, "hat": {"haben"}, "habt": {"haben"},
	"hatte": {"haben"}, "hattest": {"haben"}, "hatten": {"haben"}, "gehabt": {"haben"},
	"mache": {"machen"}, "machst": {"machen"}, "macht": {"machen"}, "machen": {"machen"},
	"machte": {"machen"}, "gemacht": {"machen"},
	"nehme": {"nehmen"}, "nimmst": {"nehmen"}, "nimmt": {"nehmen"}, "nehmen": {"nehmen"},
	"nahm": {"nehmen"}, "nahmen": {"nehmen"}, "genommen": {"nehmen"},
	"bringe": {"bringen"}, "bringst": {"bringen"}, "bringt": {"bringen"},
	"brachte": {"bringen"}, "brachten": {"bringen"}, "gebracht": {"bringen"},
	"lasse": {"lassen"}, "lässt": {"lassen"}, "lasst": {"lassen"},
	"ließ": {"lassen"}, "liessen": {"lassen"}, "ließen": {"lassen"}, "gelassen": {"lassen"},
	"stehe": {"stehen"}, "stehst": {"stehen"}, "steht": {"stehen"},
	"stand": {"stehen"}, "standen": {"stehen"}, "gestanden": {"stehen"},
	"sehe": {"sehen"}, "siehst": {"sehen"}, "sieht": {"sehen"},
	"sah": {"sehen"}, "sahen": {"sehen"}, "gesehen": {"sehen"},
	"greife": {"greifen"}, "greifst": {"greifen"}, "greift": {"greifen"},
	"griff": {"greifen"}, "griffen": {"greifen"}, "gegriffen": {"greifen"},
	"treibe": {"treiben"}, "treibst": {"treiben"}, "treibt": {"treiben"},
	"trieb": {"treiben"}, "trieben": {"treiben"}, "getrieben": {"treiben"},
	"lerne": {"lernen"}, "lernst": {"lernen"}, "lernt": {"lernen"}, "lernte": {"lernen"}, "gelernt": {"lernen"},
	"tue": {"tun"}, "tust": {"tun"}, "tut": {"tun"}, "tat": {"tun"}, "taten": {"tun"}, "getan": {"tun"},
	"drücke": {"drücken"}, "drückst": {"drücken"}, "drückt": {"drücken"}, "gedrückt": {"drücken"},
	"ausdrücke": {"ausdrücken"}, "ausdrückt": {"ausdrücken"}, "ausgedrückt": {"ausdrücken"},
	// typo soft target for SICH_AUSDRUCKEN (ausgedruckt ← ausdrucken)
	"ausgedruckt": {"ausdrucken"},
	// Portuguese ser / estar / ter / fazer / dar
	"sou": {"ser"}, "és": {"ser"}, "é": {"ser"}, "somos": {"ser"}, "são": {"ser"}, "sóc": {"ser"}, "ets": {"ser"}, "som": {"ser"}, "són": {"ser"},
	"era": {"ser"}, "eram": {"ser"}, "eres": {"ser"}, "érem": {"ser"}, "éreu": {"ser"}, "eren": {"ser"},
	"foi": {"ser", "dir"}, "foram": {"ser"}, "sido": {"ser"},
	"estou": {"estar"}, "está": {"estar"}, "estamos": {"estar"}, "estão": {"estar"},
	"estava": {"estar"}, "estavam": {"estar"}, "estaven": {"estar"}, "estado": {"estar"},
	"tenho": {"ter"}, "tens": {"ter", "tenir"}, "tem": {"ter"}, "temos": {"ter"}, "têm": {"ter"},
	"tinha": {"ter"}, "tinham": {"ter"}, "tido": {"ter"},
	"tive": {"ter"}, "tiveste": {"ter"}, "teve": {"ter"}, "tivemos": {"ter"}, "tiveram": {"ter"},
	"faço": {"fazer"}, "fazes": {"fazer"}, "faz": {"fazer"}, "fazemos": {"fazer"}, "fazem": {"fazer"},
	"fez": {"fazer"}, "fizeram": {"fazer"}, "feito": {"fazer"},
	"fiz": {"fazer"}, "fizeste": {"fazer"}, "fizemos": {"fazer"}, "fizessem": {"fazer"}, "fizesse": {"fazer"},
	"fará": {"fazer"}, "farão": {"fazer"},
	"põe": {"pôr"}, "pões": {"pôr"}, "pomos": {"pôr"}, "põem": {"pôr"}, "pôs": {"pôr"}, "pus": {"pôr"},
	"ponho": {"pôr"}, "pondes": {"pôr"},
	// preferir (PHRASAL_VERB_PREFERIR soft without portuguese.dict)
	"prefiro": {"preferir"}, "preferes": {"preferir"}, "prefere": {"preferir"},
	"preferimos": {"preferir"}, "preferem": {"preferir"}, "preferiu": {"preferir"},
	// ir present extras; vou/vão/vais already under Romance go-verbs
	"ides": {"ir"}, "fui": {"ir"},
	"dei": {"dar"}, "deste": {"dar"}, "demos": {"dar"},
	"posso": {"poder"}, "podes": {"poder"}, "pode": {"poder"}, "podemos": {"poder"}, "podem": {"poder"}, "pôde": {"poder"},
	"puc": {"poder"}, "pots": {"poder"}, "pot": {"poder"}, "podeu": {"poder"}, "poden": {"poder"},
	"podia": {"poder"}, "podien": {"poder"}, "podré": {"poder"}, "podrà": {"poder"},
	"ouço": {"ouvir"}, "ouves": {"ouvir"}, "ouve": {"ouvir"}, "ouvimos": {"ouvir"}, "ouvem": {"ouvir"}, "ouvi": {"ouvir"},
	"escuto": {"escutar"}, "escutas": {"escutar"}, "escuta": {"escutar"}, "escutamos": {"escutar"}, "escutam": {"escutar"},
	"usou": {"usar"}, "usei": {"usar"}, "usamos": {"usar"}, "usaram": {"usar"}, "usando": {"usar"},
	"referimos": {"referir"}, "referirei": {"referir"}, "referiremos": {"referir"}, "refiro": {"referir"}, "refere": {"referir"},
	// Portuguese haver (há uns minutos) / cobrir / vir
	"há": {"haver"}, "houve": {"haver"}, "haverá": {"haver"},
	"coberto": {"cobrir"}, "coberta": {"cobrir"}, "cobertos": {"cobrir"}, "cobertas": {"cobrir"},
	"cobre": {"cobrir"}, "cobrem": {"cobrir"}, "cobria": {"cobrir"},
	"veio": {"vir"}, "vieram": {"vir"}, "viria": {"vir"}, "viriam": {"vir"},
	"venha": {"vir"}, "venham": {"vir"}, "vinha": {"vir"}, "vinham": {"vir"},
	"estive": {"estar"}, "esteve": {"estar"}, "estivemos": {"estar"}, "estiveram": {"estar"},
	"dou": {"dar"}, "dás": {"dar"}, "dá": {"dar"}, "damos": {"dar"}, "dão": {"dar"},
	"deu": {"dar", "deure"}, "deram": {"dar"}, "dado": {"dar"},
	// Shared Romance "go" present (FR aller / ES ir / AST dir)
	"vais": {"aller", "ir", "dir"},
	"vas": {"aller", "ir", "dir", "anar"},
	"va": {"aller", "ir", "dir", "anar", "haver", "valer"},
	"vamos": {"ir", "dir"},
	"van": {"ir", "dir", "anar", "haver"},
	"vaig": {"anar", "haver"}, "vares": {"anar"},
	"vam": {"anar"}, "vàrem": {"anar"}, "vau": {"anar"}, "vàreu": {"anar"}, "varen": {"anar"},
	"anava": {"anar"}, "anaves": {"anar"}, "anàvem": {"anar"}, "anàveu": {"anar"}, "anaven": {"anar"},
	"aniré": {"anar"}, "anirà": {"anar"}, "anirem": {"anar"}, "aniran": {"anar"},
	// Portuguese ir (vão fazer tempo) / esquecer
	"vou": {"ir"}, "vão": {"ir"}, "ia": {"ir"}, "iam": {"ir"},
	"esqueceu": {"esquecer"}, "esqueceste": {"esquecer"}, "esqueci": {"esquecer"},
	"esquecemos": {"esquecer"}, "esqueceram": {"esquecer"},
	// Spanish ir / dar / haber / asir / revertir
	"voy": {"ir"}, "iba": {"ir"}, "iban": {"ir"}, "fue": {"ir"}, "fueron": {"ir"}, "ido": {"ir"},
	"doy": {"dar"}, "das": {"dar"}, "da": {"dar"}, "dais": {"dar"}, "dan": {"dar"},
	"daba": {"dar"}, "dabas": {"dar"}, "dábamos": {"dar"}, "daban": {"dar"},
	"di": {"dar", "do"}, "diste": {"dar"}, "dio": {"dar"}, "dimos": {"dar"}, "dieron": {"dar"},
	"he": {"haber", "haver"}, "ha": {"haber", "haver"}, "hemos": {"haber"}, "habéis": {"haber"}, "han": {"haber", "haver"},
	"hem": {"haver"}, "heu": {"haver"},
	"havia": {"haver"}, "havien": {"haver"}, "hagut": {"haver"},
	"hauré": {"haver"}, "haurà": {"haver"}, "haurem": {"haver"}, "haureu": {"haver"}, "hauran": {"haver"},
	"había": {"haber"}, "habías": {"haber"}, "habíamos": {"haber"}, "habían": {"haber"}, "hubo": {"haber"},
	"asiendo": {"asir"}, "asiéndo": {"asir"},
	"revierte": {"revertir"}, "revierten": {"revertir"}, "revertía": {"revertir"},
	// Asturian dir
	"voi": {"dir"}, "foron": {"dir"},
	// Catalan (soft until catalan.dict is vendored; FreeLing-style lemmas)
	"faig": {"fer"}, "fas": {"fer"}, "fa": {"fer"}, "fem": {"fer"}, "feu": {"fer"}, "fan": {"fer"},
	"feia": {"fer"}, "feies": {"fer"}, "fèiem": {"fer"}, "feien": {"fer"},
	"farà": {"fer"}, "faré": {"fer"}, "faran": {"fer"}, "faréu": {"fer"}, "farem": {"fer"},
	"fet": {"fer"}, "feta": {"fer"}, "fets": {"fer"}, "fetes": {"fer"},
	"dóna": {"donar"}, "dona": {"donar"}, "dónes": {"donar"}, "dones": {"donar"},
	"donem": {"donar"}, "doneu": {"donar"}, "donen": {"donar"}, "donava": {"donar"}, "donaven": {"donar"},
	"estat": {"ser", "estar"}, "estats": {"ser", "estar"}, "estada": {"ser", "estar"}, "estades": {"ser", "estar"},
	"estic": {"estar"}, "estàs": {"estar"}, "està": {"estar"}, "estem": {"estar"}, "esteu": {"estar"}, "estan": {"estar"},
	"vull": {"voler"}, "vols": {"voler"}, "vol": {"voler"}, "volem": {"voler"}, "voleu": {"voler"}, "volen": {"voler"},
	"volia": {"voler"}, "volien": {"voler"}, "voldria": {"voler"},
	"deig": {"deure"}, "deus": {"deure"}, "devem": {"deure"}, "deuen": {"deure"},
	"duia": {"dur"}, "duies": {"dur"}, "dúiem": {"dur"}, "duien": {"dur"}, "du": {"dur"}, "duen": {"dur"},
	"portava": {"portar"}, "portaven": {"portar"},
	"coneixia": {"conèixer"}, "coneixien": {"conèixer"}, "conec": {"conèixer"}, "coneix": {"conèixer"},
	"queia": {"caure"}, "queien": {"caure"}, "cau": {"caure"}, "cauen": {"caure"},
	"correran": {"córrer"}, "correré": {"córrer"}, "correrà": {"córrer"}, "corre": {"córrer"}, "corren": {"córrer"},
	"permet": {"permetre"}, "permeten": {"permetre"}, "permetia": {"permetre"},
	"permetéssim": {"permetre"}, "permeté": {"permetre"}, "permetre": {"permetre"},
	"mereix": {"merèixer"}, "mereixen": {"merèixer"}, "mereixia": {"merèixer"}, "mereixerà": {"merèixer"},
	"riu": {"riure"}, "riuen": {"riure"}, "reia": {"riure"},
	"tinc": {"tenir"}, "té": {"tenir"}, "tenim": {"tenir"}, "teniu": {"tenir"}, "tenen": {"tenir"},
	"tenia": {"tenir"}, "tenien": {"tenir"}, "tindrà": {"tenir"}, "tindré": {"tenir"}, "tindran": {"tenir"},
	"tingueu": {"tenir"}, "tingut": {"tenir"}, "tindreu": {"tenir"}, "tingues": {"tenir"},
	"servia": {"servir"}, "serveix": {"servir"}, "seiem": {"seure"},
	"sortia": {"sortir"}, "surto": {"sortir"}, "surten": {"sortir"}, "eixit": {"eixir"}, "eixien": {"eixir"},
	"declarat": {"declarar"}, "declarada": {"declarar"},
	"autoimmolà": {"autoimmolar"}, "autoinculpà": {"autoinculpar"},
	"autoimmola": {"autoimmolar"}, "autoinculpa": {"autoinculpar"},
	"facis": {"fer"}, "faci": {"fer"}, "facin": {"fer"}, "fent": {"fer"}, "féu": {"fer"},
	"hauria": {"haver"}, "hauries": {"haver"}, "hauríem": {"haver"}, "haurieu": {"haver"}, "haurien": {"haver"},
	"hagués": {"haver"}, "haguéssiu": {"haver"}, "hagéssiu": {"hajar"}, "hagés": {"hajar"},
	"fou": {"ser"}, "cessat": {"cessar"}, "cessada": {"cessar"},
	"mútua": {"mutu"}, "mútues": {"mutu"}, "mutus": {"mutu"},
	"vingué": {"venir"}, "vingueren": {"venir"}, "venia": {"venir"}, "venien": {"venir"}, "vingui": {"venir"},
	"vivia": {"viure"}, "vivien": {"viure"}, "viu": {"viure"}, "viuen": {"viure"},
	"acompliran": {"acomplir", "complir"}, "acompleix": {"acomplir", "complir"}, "compleix": {"complir"},
	"vés": {"anar", "enviar"}, "ves": {"anar", "enviar"},
	"posis": {"posar"}, "posi": {"posar"}, "posa": {"posar"}, "posen": {"posar"},
	"sap": {"saber"}, "saps": {"saber"}, "sabem": {"saber"}, "saben": {"saber"},
	"tragueren": {"treure", "traure"}, "tragué": {"treure", "traure"}, "treu": {"treure"},
	"trepava": {"trepar"}, "trepa": {"trepar"}, "trepen": {"trepar"},
	"vulgueu": {"voler"}, "vulguis": {"voler"}, "vulgui": {"voler"},
	"sorts": {"sord"}, "sords": {"sord"}, "sordes": {"sord"},
	// Catalan soft residuals (FreeLing-style lemmas; no catalan.dict in tree)
	"calien": {"caldre"}, "cal": {"caldre"}, "calen": {"caldre"}, "calia": {"caldre"},
	"prengueren": {"prendre"}, "prengué": {"prendre"}, "prengui": {"prendre"}, "prenguin": {"prendre"},
	"conclogueren": {"concloure"}, "conclogué": {"concloure"}, "conclou": {"concloure"},
	"donades": {"donar"}, "donada": {"donar"}, "donats": {"donar"}, "donat": {"donar"},
	"cancerígenes": {"cancerigen"}, "cancerígena": {"cancerigen"}, "cancerígens": {"cancerigen"},
	"desinteressada": {"desinteressat"}, "desinteressades": {"desinteressat"},
	"interessant": {"interessant"}, "interessants": {"interessant"},
	// FreeLing multiword/special lemmas used by Catalan grammar.xml (Java tagger)
	"discapacitats": {"discapacitat2", "discapacitat"},
	"discapacitades": {"discapacitat2", "discapacitat"},
	"discapacitat": {"discapacitat2", "discapacitat"},
	"discapacitada": {"discapacitat2", "discapacitat"},
	// Catalan noun plurals (dies ← dia for TOTS_ELS_DIES etc.)
	"dies": {"dia"}, "anys": {"any"}, "mesos": {"mes"}, "hores": {"hora"}, "setmanes": {"setmana"},
	"coes": {"coa"}, "oïdes": {"oïda"}, "col·laboració": {"col·laboració"},
	"cèl·lules": {"cèl·lula"}, "cèl·lula": {"cèl·lula"},
	// Irish prepositional pronouns (orm ← ar + mé, liom ← le + mé; Java tagger lemmas)
	"orm": {"ar"}, "ort": {"ar"}, "air": {"ar"}, "uirthi": {"ar"}, "orainn": {"ar"}, "oraibh": {"ar"}, "orthu": {"ar"},
	"liom": {"le"}, "leat": {"le"}, "leis": {"le"}, "léi": {"le"}, "linn": {"le"}, "libh": {"le"}, "leo": {"le"},
	"agam": {"ag"}, "agat": {"ag"}, "aige": {"ag"}, "aici": {"ag"}, "againn": {"ag"}, "agaibh": {"ag"}, "acu": {"ag"},
	"dom": {"do"}, "duit": {"do"}, "dó": {"do"}, "dúinn": {"do"}, "daoibh": {"do"}, "dóibh": {"do"},
	// Polish suppletive comparative (dobry→lepszy) and zło case forms — LEPSZE_ZLO.
	"lepszy": {"dobry"}, "lepsza": {"dobry"}, "lepsze": {"dobry"},
	"lepszego": {"dobry"}, "lepszej": {"dobry"}, "lepszemu": {"dobry"},
	"lepszą": {"dobry"}, "lepszym": {"dobry"}, "lepszych": {"dobry"},
	"lepszymi": {"dobry"}, "lepsi": {"dobry"},
	"najlepszy": {"dobry"}, "najlepsza": {"dobry"}, "najlepsze": {"dobry"},
	"najlepszym": {"dobry"}, "najlepszej": {"dobry"}, "najlepszych": {"dobry"},
	"gorszy": {"zły"}, "gorsza": {"zły"}, "gorsze": {"zły"},
	"gorszym": {"zły"}, "gorszej": {"zły"}, "gorszych": {"zły"},
	"mniejszy": {"mały"}, "mniejsza": {"mały"}, "mniejsze": {"mały"},
	"mniejszego": {"mały"}, "mniejszej": {"mały"}, "mniejszemu": {"mały"},
	"mniejszą": {"mały"}, "mniejszym": {"mały"}, "mniejszych": {"mały"},
	"mniejszymi": {"mały"}, "mniejsi": {"mały"},
	"zła": {"zło"}, "złe": {"zło"}, "złego": {"zło"}, "złemu": {"zło"},
	"złym": {"zło"}, "złem": {"zło"}, "złych": {"zło"}, "złymi": {"zło"},
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
	// French elision before other checks (d’ ↔ de with inflected="yes").
	if softFrenchElisionMatch(surface, base) {
		return true
	}
	if softCatalanElisionMatch(surface, base) {
		return true
	}
	// EO x-system / diacritic fold only when digraphs/diacritics present.
	if softMayNeedEOFold(surface) || softMayNeedEOFold(base) {
		if softEsperantoFold(surface) == softEsperantoFold(base) {
			return true
		}
	}
	if surface == base {
		return true
	}
	// Irregular auxiliaries / go-verbs (was→be, est→être, va→dir, …).
	if lems, ok := softIrregularLemma[surface]; ok {
		for _, lem := range lems {
			if lem == base {
				return true
			}
		}
	}
	// Spanish enclitics on gerunds/infinitives (asiéndolos → asiendo → asir).
	if core := softSpanishStripEnclitic(surface); core != surface {
		if softInflectedSurfaceMatch(core, base, true) {
			return true
		}
		if lems, ok := softIrregularLemma[core]; ok {
			for _, lem := range lems {
				if lem == base {
					return true
				}
			}
		}
	}
	// Portuguese ênclise (Esqueceu-se → esqueceu → esquecer; cobrir-se → cobrir).
	if core := softPortugueseStripEnclitic(surface); core != surface {
		if softInflectedSurfaceMatch(core, base, true) {
			return true
		}
		if lems, ok := softIrregularLemma[core]; ok {
			for _, lem := range lems {
				if lem == base {
					return true
				}
			}
		}
	}
	// German ge- participles (gemacht←machen) when not listed above.
	if softGermanGeParticiple(surface, base) {
		return true
	}
	// German adj stem alternation: hoch → hohe/hohen/… (ch→h before vowel ending).
	if softGermanAdjStemAlt(surface, base) {
		return true
	}
	// French -er past participle / present (placé←placer, place←placer).
	if softFrenchErInflected(surface, base) {
		return true
	}
	// Prefix check on EO-folded forms (ambaŭ / Ambaux) when EO marks present.
	if softMayNeedEOFold(surface) || softMayNeedEOFold(base) {
		sf, bf := softEsperantoFold(surface), softEsperantoFold(base)
		if strings.HasPrefix(sf, bf) {
			suf := sf[len(bf):]
			if len(suf) > 0 && len(suf) <= 4 {
				switch suf {
				case "s", "es", "n", "en", "er", "e", "a", "os", "as", "is", "ns", "j", "jn", "oj", "ojn", "an", "on",
					"ing", "ed", "ied", "ies", "d":
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
	case "s", "es", "n", "en", "er", "e", "a", "os", "as", "is", "ns", "aren", "eren", "j", "jn", "oj", "ojn",
		"ing", "ed", "ied", "ies", "d":
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

// softGermanUmlautFold maps äöüß to ascii for soft stem compares (Tänze/Tanz).
var softGermanUmlautReplacer = strings.NewReplacer(
	"ä", "a", "ö", "o", "ü", "u", "ß", "ss",
	"Ä", "a", "Ö", "o", "Ü", "u",
)

func softGermanUmlautFold(s string) string {
	return softGermanUmlautReplacer.Replace(strings.ToLower(s))
}

// softGermanAdjStemAlt handles irregular adjective stems (Java tagger lemma hoch
// for surface hohe). Pattern: base ends in "ch", declined forms use stem+"h"+ending
// (hoch→hohe/hohen/hoher/hohes/hohem). Used when no DE Morfologik dict is available.
func softGermanAdjStemAlt(surface, base string) bool {
	if surface == "" || base == "" {
		return false
	}
	s := strings.ToLower(surface)
	b := strings.ToLower(base)
	if !strings.HasSuffix(b, "ch") || len(b) < 3 {
		return false
	}
	// hoch → hoh
	stem := b[:len(b)-2] + "h"
	if !strings.HasPrefix(s, stem) {
		return false
	}
	switch s[len(stem):] {
	case "", "e", "en", "er", "es", "em":
		return true
	default:
		return false
	}
}

// softSharedStemMatch is true when surface and base share a long letter stem
// and differ only by short inflectional endings (говорить/говорите, храбрый/храбрая).
// Min stem is 4 for longer words; 3 is allowed for short bases (яйцо/яйца).
func softSharedStemMatch(a, b string) bool {
	// Try umlaut-folded compare for German plurals (Tänze/Tanz).
	if softSharedStemMatchRunes([]rune(a), []rune(b)) {
		return true
	}
	fa, fb := softGermanUmlautFold(a), softGermanUmlautFold(b)
	if fa != strings.ToLower(a) || fb != strings.ToLower(b) {
		return softSharedStemMatchRunes([]rune(fa), []rune(fb))
	}
	return false
}

func softSharedStemMatchRunes(ar, br []rune) bool {
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
	if pat == "" {
		return nil
	}
	if !strings.Contains(pat, "|") {
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
		if a == "" {
			continue
		}
		// Flatten nested (?:ser|estar|ter) so inflected soft can match lemmas.
		if strings.HasPrefix(a, "(?:") && strings.HasSuffix(a, ")") && strings.Contains(a, "|") {
			out = append(out, softRegexpAlternatives(a)...)
			continue
		}
		out = append(out, a)
	}
	return out
}

// softMayNeedEOFold is true when s may contain Esperanto x-system digraphs or
// ĉĝĥĵŝŭ. Hot soft match paths skip EO fold work when this is false.
func softMayNeedEOFold(s string) bool {
	for _, r := range s {
		switch r {
		case 'x', 'X',
			'ĉ', 'ĝ', 'ĥ', 'ĵ', 'ŝ', 'ŭ',
			'Ĉ', 'Ĝ', 'Ĥ', 'Ĵ', 'Ŝ', 'Ŭ':
			return true
		}
	}
	return false
}

// softEsperantoUnicode converts x-system digraphs to Unicode diacritics (cx→ĉ).
func softEsperantoUnicode(s string) string {
	if s == "" || !softMayNeedEOFold(s) {
		return s
	}
	// Process lowercase digraphs; matching only needs a lowered canonical form.
	low := strings.ToLower(s)
	if !strings.Contains(low, "x") {
		return low
	}
	repl := []struct{ from, to string }{
		{"cx", "ĉ"}, {"gx", "ĝ"}, {"hx", "ĥ"}, {"jx", "ĵ"}, {"sx", "ŝ"}, {"ux", "ŭ"},
	}
	for _, r := range repl {
		low = strings.ReplaceAll(low, r.from, r.to)
	}
	return low
}

// softEOFoldReplacer strips EO diacritics to ASCII (cached; do not NewReplacer per call).
var softEOFoldReplacer = strings.NewReplacer(
	"ĉ", "c", "ĝ", "g", "ĥ", "h", "ĵ", "j", "ŝ", "s", "ŭ", "u",
)

// softEsperantoFold maps x-system and EO diacritics to plain ASCII letters for soft compare.
func softEsperantoFold(s string) string {
	if s == "" {
		return s
	}
	if !softMayNeedEOFold(s) {
		return strings.ToLower(s)
	}
	s = softEsperantoUnicode(strings.ToLower(s))
	return softEOFoldReplacer.Replace(s)
}

// softEsperantoLemmaCandidates yields likely dictionary forms for EO surfaces (biliardoj→biliardo).
// Pure-ASCII EO verbs (darfas→darfi) have no diacritics; always allow suffix strips.
func softEsperantoLemmaCandidates(surface string) []string {
	if surface == "" {
		return nil
	}
	low := strings.ToLower(surface)
	u := softEsperantoUnicode(low)
	if u == low && !softMayNeedEOFold(low) {
		// No digraph/diacritic rewrite; still try morphology on lowered surface.
		u = low
	}
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
	// Verb finite → infinitive -i (preferas→preferi, darfas→darfi).
	for _, st := range []strip{
		{"as", "i"}, {"is", "i"}, {"os", "i"}, {"us", "i"}, {"u", "i"},
	} {
		if strings.HasSuffix(u, st.suf) && len(u) > len(st.suf)+1 {
			stem := u[:len(u)-len(st.suf)] + st.base
			if stem != u {
				out = append(out, stem)
			}
		}
	}
	return out
}
