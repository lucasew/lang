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

// PatternTokenMatcher ports org.languagetool.rules.patterns.PatternTokenMatcher.
// Matching is Java-faithful: surface/regexp + real POS tags only.
// Soft invent (closed-class surface lists, untagged open-class accept, soft lemmas)
// is not part of this twin — see docs/faithful-port-policy.md.
type PatternTokenMatcher struct {
	Base *PatternToken
	// patternToken is the working token after resolveReference (Java patternToken field).
	patternToken *PatternToken
	// andMatchers are AndGroup members after prepareAndGroup/resolveReference.
	andMatchers []*PatternTokenMatcher
	// compiled RE for Token when Regexp is set
	tokenRE *regexp.Regexp
	// StrictPOS: untagged tokens only match postag=UNKNOWN (Java with a real tagger).
	// Default true for faithful matching; soft false is not used inside the wall.
	StrictPOS bool
}

func NewPatternTokenMatcher(pt *PatternToken) *PatternTokenMatcher {
	m := &PatternTokenMatcher{Base: pt, patternToken: pt, StrictPOS: true}
	m.recompileTokenRE()
	return m
}

func (m *PatternTokenMatcher) recompileTokenRE() {
	m.tokenRE = nil
	pt := m.active()
	if pt != nil && pt.Regexp && pt.Token != "" {
		flags := ""
		pat := normalizeJavaRegexp(pt.Token)
		if !pt.CaseSensitive && !strings.Contains(pat, `\p{`) {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + "^(?:" + pat + ")$")
		if err == nil {
			m.tokenRE = re
		}
	}
}

// active returns the working PatternToken (compiled reference or base).
func (m *PatternTokenMatcher) active() *PatternToken {
	if m == nil {
		return nil
	}
	if m.patternToken != nil {
		return m.patternToken
	}
	return m.Base
}

// ResolveReference ports PatternTokenMatcher.resolveReference.
// firstMatchToken is the absolute index of the first matched pattern token (-1 if none yet).
// langCode selects LanguageSynthesizer for setpos target tags.
func (m *PatternTokenMatcher) ResolveReference(firstMatchToken int, tokens []*languagetool.AnalyzedTokenReadings, langCode string) {
	if m == nil || m.Base == nil {
		return
	}
	m.patternToken = m.Base
	m.recompileTokenRE()
	if !m.Base.IsReferenceElement() {
		return
	}
	// Java: refPos = firstMatchToken + getTokenRef() (TokenRef is raw no=, not 1-based).
	refPos := firstMatchToken + m.Base.TokenMatch.GetTokenRef()
	if firstMatchToken < 0 || refPos < 0 || refPos >= len(tokens) || tokens[refPos] == nil {
		return
	}
	synth := LanguageSynthesizer(langCode)
	m.patternToken = m.Base.CompileFromReference(tokens[refPos], synth)
	m.recompileTokenRE()
}

// PrepareAndGroup ports PatternTokenMatcher.prepareAndGroup:
// resolve references on each AndGroup member before and-group checks.
func (m *PatternTokenMatcher) PrepareAndGroup(firstMatchToken int, tokens []*languagetool.AnalyzedTokenReadings, langCode string) {
	if m == nil || m.Base == nil || len(m.Base.AndGroup) == 0 {
		m.andMatchers = nil
		return
	}
	m.andMatchers = make([]*PatternTokenMatcher, 0, len(m.Base.AndGroup))
	for _, andPt := range m.Base.AndGroup {
		if andPt == nil {
			continue
		}
		am := NewPatternTokenMatcher(andPt)
		am.StrictPOS = m.StrictPOS
		am.ResolveReference(firstMatchToken, tokens, langCode)
		m.andMatchers = append(m.andMatchers, am)
	}
}

// normalizeJavaRegexp maps Java/PCRE constructs used in LT XML to Go RE2.
func normalizeJavaRegexp(pat string) string {
	if pat == "" {
		return pat
	}
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
				if i+6 <= len(pat) {
					hex := pat[i+2 : i+6]
					if _, err := strconv.ParseUint(hex, 16, 32); err == nil {
						fmt.Fprintf(&b, `\x{%s}`, strings.ToLower(hex))
						i += 6
						continue
					}
				}
			case 'U':
				if i+10 <= len(pat) {
					hex := pat[i+2 : i+10]
					if _, err := strconv.ParseUint(hex, 16, 32); err == nil {
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
	return m.active()
}

// IsMatched checks whether a single AnalyzedToken matches the pattern token.
// Ports PatternToken.isMatched / PatternTokenMatcher.isMatched (string + POS only).
func (m *PatternTokenMatcher) IsMatched(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.active() == nil || token == nil {
		return false
	}
	pt := m.active()
	// Java PatternToken.isMatched: spacebefore=yes/no must match token.isWhitespaceBefore().
	if pt.WhitespaceBefore != nil && token.IsWhitespaceBefore() != *pt.WhitespaceBefore {
		return false
	}
	matched := m.matchSurface(token.GetToken())
	if pt.MatchInflected && !matched {
		// Java: also try lemma via dictionary readings — not soft morphological invent.
		if lem := token.GetLemma(); lem != nil && *lem != "" {
			matched = m.matchSurface(*lem)
		}
	}
	if pt.Pos != nil && pt.Pos.PosTag != "" {
		pos := token.GetPOSTag()
		posOK := false
		if pos != nil && *pos != "" {
			if pt.Pos.Regexp {
				re, err := regexp.Compile("^(?:" + normalizeJavaRegexp(pt.Pos.PosTag) + ")$")
				if err == nil {
					posOK = re.MatchString(*pos)
				}
			} else {
				posOK = *pos == pt.Pos.PosTag
			}
		} else {
			// Untagged: Java UNKNOWN only (faithful StrictPOS).
			tag := strings.ToUpper(pt.Pos.PosTag)
			posOK = tag == "UNKNOWN" || strings.HasPrefix(tag, "UNKNOWN")
		}
		if pt.Pos.Negate {
			posOK = !posOK
		}
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

// IsMatchedByPreviousException ports PatternToken.isMatchedByPreviousException.
func (m *PatternTokenMatcher) IsMatchedByPreviousException(prev *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || prev == nil || m.Base.PreviousException == "" {
		return false
	}
	return matchExceptionOnReadings(prev, m.Base.PreviousException, m.Base.PreviousExceptionRE, m.Base.PreviousExceptionCaseSensitive)
}

// IsMatchedByNextException ports PatternToken next-scope exception (surface).
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
	if matchExceptionSurface(tok.GetToken(), exc, isRE, caseSensitive) {
		return true
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		if matchExceptionSurface(r.GetToken(), exc, isRE, caseSensitive) {
			return true
		}
		if lem := r.GetLemma(); lem != nil && *lem != "" {
			if matchExceptionSurface(*lem, exc, isRE, caseSensitive) {
				return true
			}
		}
	}
	return false
}

func matchExceptionSurface(surface, exc string, isRE, caseSensitive bool) bool {
	if exc == "" {
		return false
	}
	surface = normalizeApostrophes(surface)
	exc = normalizeApostrophes(exc)
	if isRE {
		flags := ""
		pat := normalizeJavaRegexp(exc)
		if !caseSensitive && !strings.Contains(pat, `\p{`) {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + "^(?:" + pat + ")$")
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

// CollectMatchedReadings returns readings that satisfy IsMatched (for Unifier).
// Ports the readingsToUnify collection in AbstractPatternRulePerformer.testAllReadings.
func (m *PatternTokenMatcher) CollectMatchedReadings(atr *languagetool.AnalyzedTokenReadings) []*languagetool.AnalyzedToken {
	if m == nil || atr == nil {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, r := range atr.GetReadings() {
		if r != nil && m.IsMatched(r) {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		// Untagged surface path used by IsMatchedReadings.
		probe := languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil)
		probe.SetWhitespaceBefore(atr.IsWhitespaceBefore())
		if m.IsMatched(probe) {
			out = append(out, probe)
		}
	}
	return out
}

// IsMatchedReadings evaluates the pattern token against all readings (Java-style).
func (m *PatternTokenMatcher) IsMatchedReadings(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	pt := m.active()
	if pt == nil {
		return false
	}
	if pt.WhitespaceBefore != nil {
		if atr.IsWhitespaceBefore() != *pt.WhitespaceBefore {
			return false
		}
	}
	if pt.ChunkTag != "" {
		if !chunkTagMatches(atr, pt.ChunkTag, pt.ChunkTagRegexp) {
			return false
		}
	}
	if len(pt.AndGroup) > 0 {
		if !m.matchAndGroupReadings(atr) {
			return false
		}
		return !m.anyReadingExceptionMatched(atr)
	}
	anyMatched := false
	for _, r := range atr.GetReadings() {
		if r == nil {
			continue
		}
		if m.IsMatched(r) {
			anyMatched = true
			break
		}
	}
	if !anyMatched {
		// Chunk-only pattern tokens (empty surface/POS): chunk already matched above.
		if pt.Token == "" && (pt.Pos == nil || pt.Pos.PosTag == "") && pt.ChunkTag != "" {
			return !m.anyReadingExceptionMatched(atr)
		}
		// Faithful: untagged surface — only UNKNOWN postag patterns can match.
		probe := languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil)
		probe.SetWhitespaceBefore(atr.IsWhitespaceBefore())
		if !m.IsMatched(probe) {
			return false
		}
		return !m.isExceptionMatchedCompletely(probe)
	}
	return !m.anyReadingExceptionMatched(atr)
}

func (m *PatternTokenMatcher) anyReadingExceptionMatched(atr *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || !m.Base.HasCurrentException() || atr == nil {
		return false
	}
	for _, r := range atr.GetReadings() {
		if r != nil && m.isExceptionMatchedCompletely(r) {
			return true
		}
	}
	if len(atr.GetReadings()) == 0 {
		probe := languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil)
		return m.isExceptionMatchedCompletely(probe)
	}
	return false
}

// isExceptionMatchedCompletely ports PatternToken.isExceptionMatchedCompletely
// via isExceptionMatched → exception PatternToken.isMatched (with negation XOR).
func (m *PatternTokenMatcher) isExceptionMatchedCompletely(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.Base == nil || token == nil || !m.Base.HasCurrentException() {
		return false
	}
	pt := m.Base
	// Surface match (Java textMatcher / getNegation XOR).
	hasSurface := pt.TokenException != ""
	surfaceMatch := false
	if hasSurface {
		surfaceMatch = matchExceptionSurface(token.GetToken(), pt.TokenException, pt.TokenExceptionRE, pt.TokenExceptionCaseSensitive)
	}
	// POS match (Java isPosTokenMatched; null posToken → true).
	posMatch := true
	if pt.TokenExceptionPosTag != "" {
		pos := token.GetPOSTag()
		if pos == nil || *pos == "" {
			// Java UNKNOWN_TAG matches null POS; otherwise false
			if strings.Contains(pt.TokenExceptionPosTag, "UNKNOWN") {
				posMatch = true
			} else {
				posMatch = false
			}
		} else if pt.TokenExceptionPosRE {
			re, err := regexp.Compile("^(?:" + normalizeJavaRegexp(pt.TokenExceptionPosTag) + ")$")
			posMatch = err == nil && re.MatchString(*pos)
		} else {
			posMatch = *pos == pt.TokenExceptionPosTag
		}
	}
	// Java isMatched:
	//   with surface: (textMatch ^ neg) && (posMatch ^ posNeg)
	//   without:      !neg && (posMatch ^ posNeg)
	if hasSurface {
		return (surfaceMatch != pt.TokenExceptionNegation) && (posMatch != pt.TokenExceptionPosNegation)
	}
	return !pt.TokenExceptionNegation && (posMatch != pt.TokenExceptionPosNegation)
}

// matchAndGroupReadings ports Java and-group accumulation over all readings.
// Uses prepareAndGroup-resolved andMatchers when present (refs/setpos on and-members).
func (m *PatternTokenMatcher) matchAndGroupReadings(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	// Base element without AndGroup list (Java isMatched on base patternToken).
	base := m.active()
	if base == nil {
		return false
	}
	baseBare := *base
	baseBare.AndGroup = nil
	baseM := NewPatternTokenMatcher(&baseBare)
	baseM.patternToken = &baseBare
	baseM.StrictPOS = m.StrictPOS
	baseM.recompileTokenRE()

	andMatchers := m.andMatchers
	if len(andMatchers) == 0 && m.Base != nil {
		// Fallback without prepareAndGroup (tests).
		for _, andPt := range m.Base.AndGroup {
			am := NewPatternTokenMatcher(andPt)
			am.StrictPOS = m.StrictPOS
			andMatchers = append(andMatchers, am)
		}
	}
	baseOK := false
	andOK := make([]bool, len(andMatchers))
	for _, r := range atr.GetReadings() {
		if r == nil {
			continue
		}
		if baseM.IsMatched(r) {
			baseOK = true
		}
		for i, am := range andMatchers {
			if am != nil && am.IsMatched(r) {
				andOK[i] = true
			}
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

func chunkTagMatches(atr *languagetool.AnalyzedTokenReadings, want string, isRegexp bool) bool {
	if atr == nil || want == "" {
		return false
	}
	if isRegexp {
		return atr.MatchesChunkRegex(want)
	}
	for _, t := range atr.GetChunkTags() {
		if t == want {
			return true
		}
	}
	return false
}

// matchSurface ports Java string/regexp surface match (no soft morphology invent).
func (m *PatternTokenMatcher) matchSurface(surface string) bool {
	pt := m.active()
	if pt == nil {
		return false
	}
	if pt.Token == "" {
		return true
	}
	rawSurface := surface
	surface = normalizeApostrophes(surface)
	want := normalizeApostrophes(pt.Token)
	// Arabic: equality may compare undiacritized forms (tagger strips tashkeel for lookup).
	// Do not strip before regexp — patterns like .*اً$ need tanwin.
	surfaceNT, wantNT := surface, want
	if hasArabic(surface) || hasArabic(pt.Token) {
		surfaceNT = tools.RemoveTashkeel(surface)
		wantNT = tools.RemoveTashkeel(want)
	}
	if pt.Regexp {
		if m.tokenRE == nil {
			return false
		}
		return m.tokenRE.MatchString(rawSurface) || m.tokenRE.MatchString(surface)
	}
	if pt.CaseSensitive {
		return rawSurface == pt.Token || surface == want ||
			(surfaceNT != surface && surfaceNT == wantNT)
	}
	if strings.EqualFold(surface, want) || strings.EqualFold(rawSurface, pt.Token) {
		return true
	}
	if surfaceNT != surface || wantNT != want {
		if strings.EqualFold(surfaceNT, wantNT) {
			return true
		}
	}
	return false
}

func hasArabic(s string) bool {
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
	s = strings.ReplaceAll(s, "\u2019", "'")
	s = strings.ReplaceAll(s, "\u02BC", "'")
	s = strings.ReplaceAll(s, "\u2018", "'")
	return s
}
