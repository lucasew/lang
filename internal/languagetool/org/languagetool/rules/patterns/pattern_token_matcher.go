package patterns

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// PatternTokenMatcher ports org.languagetool.rules.patterns.PatternTokenMatcher.
// Matching is Java-faithful: surface/regexp via StringMatcher + real POS tags only.
// Soft invent (closed-class surface lists, untagged open-class accept, soft lemmas)
// is not part of this twin — see docs/faithful-port-policy.md.
type PatternTokenMatcher struct {
	Base *PatternToken
	// patternToken is the working token after resolveReference (Java patternToken field).
	patternToken *PatternToken
	// andMatchers are AndGroup members after prepareAndGroup/resolveReference.
	andMatchers []*PatternTokenMatcher
	// textMatcher ports PatternToken.textMatcher (StringMatcher.create).
	textMatcher *StringMatcher
	// StrictPOS: untagged tokens only match postag=UNKNOWN (Java with a real tagger).
	// Default true for faithful matching; soft false is not used inside the wall.
	StrictPOS bool
}

func NewPatternTokenMatcher(pt *PatternToken) *PatternTokenMatcher {
	m := &PatternTokenMatcher{Base: pt, patternToken: pt, StrictPOS: true}
	m.recompileTextMatcher()
	return m
}

// recompileTextMatcher ports PatternToken.setTextMatcher /
// StringMatcher.create(normalizeTextPattern(token), regExp, caseSensitive).
func (m *PatternTokenMatcher) recompileTextMatcher() {
	m.textMatcher = nil
	pt := m.active()
	if pt == nil {
		return
	}
	pat := NormalizeTextPattern(pt.Token)
	if pt.Regexp {
		// Go RE2 mapping of Java \uXXXX / inline flags (not a semantic invent).
		pat = normalizeJavaRegexp(pat)
	}
	// Empty pattern: Java TEST_STRING_MASK off — still build matcher for getString parity.
	m.textMatcher = NewStringMatcher(pat, pt.Regexp, pt.CaseSensitive)
}

// recompileTokenRE is kept as an alias for call sites that still use the old name.
func (m *PatternTokenMatcher) recompileTokenRE() { m.recompileTextMatcher() }

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
// Ports PatternToken.isMatched — Java uses XOR for negation and posNegation with
// operator precedence (^ binds tighter than &&):
//
//	hasString: (textMatch ^ negation) && (posMatch ^ posNegation)
//	no string: !negation && (posMatch ^ posNegation)
func (m *PatternTokenMatcher) IsMatched(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.active() == nil || token == nil {
		return false
	}
	pt := m.active()
	// Java PatternToken.isMatched: spacebefore=yes/no must match token.isWhitespaceBefore().
	if pt.WhitespaceBefore != nil && token.IsWhitespaceBefore() != *pt.WhitespaceBefore {
		return false
	}
	// Java TEST_STRING_MASK: !StringTools.isEmpty(textMatcher.pattern) after normalize.
	hasSurface := NormalizeTextPattern(pt.Token) != ""
	surfaceMatch := false
	if hasSurface {
		// Java: textMatcher.matches(getTestToken(token)) — single string, not surface-then-lemma.
		surfaceMatch = m.matchSurface(getTestToken(pt, token))
	}
	posNegation := pt.Pos != nil && pt.Pos.Negate
	// Java: isPosTokenMatched(token) ^ posNegation
	posMatch := IsPosTokenMatched(pt.Pos, token)
	if hasSurface {
		return (surfaceMatch != pt.Negation) && (posMatch != posNegation)
	}
	return !pt.Negation && (posMatch != posNegation)
}

// getTestToken ports PatternToken.getTestToken — when inflected, lemma if non-null else surface.
func getTestToken(pt *PatternToken, token *languagetool.AnalyzedToken) string {
	if pt != nil && pt.MatchInflected {
		if lem := token.GetLemma(); lem != nil {
			return *lem
		}
	}
	return token.GetToken()
}

// IsMatchedByPreviousException ports PatternToken.isMatchedByPreviousException
// (AnalyzedTokenReadings overload): any reading matching any previous exception.
func (m *PatternTokenMatcher) IsMatchedByPreviousException(prev *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || prev == nil || !m.Base.HasPreviousException() {
		return false
	}
	if len(m.Base.PreviousExceptions) > 0 {
		for _, r := range prev.GetReadings() {
			if r == nil {
				continue
			}
			for _, ex := range m.Base.PreviousExceptions {
				if ex == nil {
					continue
				}
				// Java: !testException.hasNextException() — previous list never has next flag.
				if NewPatternTokenMatcher(ex).IsMatched(r) {
					return true
				}
			}
		}
		return false
	}
	// Legacy single surface field
	return matchExceptionOnReadings(prev, m.Base.PreviousException, m.Base.PreviousExceptionRE, m.Base.PreviousExceptionCaseSensitive)
}

// IsMatchedByNextException ports isMatchedByScopeNextException over readings:
// any reading matching any next-scope exception (Java often probes first reading only
// at call sites; scanning all readings is a safe superset for AnalyzedTokenReadings API).
func (m *PatternTokenMatcher) IsMatchedByNextException(next *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || next == nil || !m.Base.HasNextException() {
		return false
	}
	if len(m.Base.NextExceptions) > 0 {
		for _, r := range next.GetReadings() {
			if r == nil {
				continue
			}
			for _, ex := range m.Base.NextExceptions {
				if ex == nil {
					continue
				}
				if NewPatternTokenMatcher(ex).IsMatched(r) {
					return true
				}
			}
		}
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
	// Java exception PatternToken also uses StringMatcher.create(normalizeTextPattern(...)).
	pat := NormalizeTextPattern(exc)
	if isRE {
		pat = normalizeJavaRegexp(pat)
	}
	return NewStringMatcher(pat, isRE, caseSensitive).Matches(surface)
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

// IsMatchedReadings evaluates the pattern token against all readings.
// Ports AbstractPatternRulePerformer.testAllReadings order for one token:
// reading match → exceptions → chunk (XOR negation) → and-group chunks.
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

	anyMatched := false
	if len(pt.AndGroup) > 0 {
		anyMatched = m.matchAndGroupReadings(atr)
	} else {
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
			// Faithful: untagged surface — only UNKNOWN postag patterns can match.
			// Chunk-only / empty-surface tokens: IsMatched with no string is !negation
			// (and optional POS); chunk gate applied below (Java).
			probe := languagetool.NewAnalyzedToken(atr.GetToken(), nil, nil)
			probe.SetWhitespaceBefore(atr.IsWhitespaceBefore())
			if m.IsMatched(probe) {
				anyMatched = true
				// Exception check for untagged probe path (no readings matched).
				if m.isExceptionMatchedCompletely(probe) {
					return false
				}
			}
		}
	}

	if anyMatched {
		if m.anyReadingExceptionMatched(atr) {
			return false
		}
	}

	// Java: anyMatched &= chunkMatch ^ getNegation()
	if pt.ChunkTag != "" {
		chunkOK := chunkTagMatches(atr, pt.ChunkTag, pt.ChunkTagRegexp)
		anyMatched = anyMatched && (chunkOK != pt.Negation)
	}
	// Java and-group chunk tags (no XOR with negation).
	for _, andTok := range pt.AndGroup {
		if andTok == nil || andTok.ChunkTag == "" {
			continue
		}
		if !chunkTagMatches(atr, andTok.ChunkTag, andTok.ChunkTagRegexp) {
			anyMatched = false
			break
		}
	}
	return anyMatched
}

func (m *PatternTokenMatcher) anyReadingExceptionMatched(atr *languagetool.AnalyzedTokenReadings) bool {
	if m == nil || m.Base == nil || atr == nil {
		return false
	}
	// Current exceptions on this token or on AndGroup members (Java isExceptionMatchedCompletely).
	if !m.Base.HasCurrentException() && !m.hasAndGroupCurrentExceptions() {
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

func (m *PatternTokenMatcher) hasAndGroupCurrentExceptions() bool {
	if m == nil || m.Base == nil {
		return false
	}
	for _, andTok := range m.Base.AndGroup {
		if andTok != nil && andTok.HasCurrentException() {
			return true
		}
	}
	return false
}

// isExceptionMatchedCompletely ports PatternToken.isExceptionMatchedCompletely:
// isExceptionMatched (any current exception) || isAndExceptionGroupMatched.
func (m *PatternTokenMatcher) isExceptionMatchedCompletely(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.Base == nil || token == nil {
		return false
	}
	if m.isExceptionMatched(token) {
		return true
	}
	return m.isAndExceptionGroupMatched(token)
}

// isExceptionMatched ports PatternToken.isExceptionMatched: any current-scope
// exception PatternToken.isMatched (disjunction). Skips next-scope (separate list).
func (m *PatternTokenMatcher) isExceptionMatched(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.Base == nil || token == nil {
		return false
	}
	pt := m.Base
	if len(pt.CurrentExceptions) > 0 {
		for _, ex := range pt.CurrentExceptions {
			if ex == nil {
				continue
			}
			if NewPatternTokenMatcher(ex).IsMatched(token) {
				return true
			}
		}
		return false
	}
	// Legacy single TokenException* fields — build a PatternToken and use isMatched
	// (Java exceptions are full PatternToken instances).
	if !pt.HasCurrentException() {
		return false
	}
	ex := NewPatternToken(pt.TokenException, pt.TokenExceptionCaseSensitive, pt.TokenExceptionRE, false)
	ex.SetNegation(pt.TokenExceptionNegation)
	if pt.TokenExceptionPosTag != "" {
		ex.SetPosToken(PosToken{
			PosTag: pt.TokenExceptionPosTag,
			Regexp: pt.TokenExceptionPosRE,
			Negate: pt.TokenExceptionPosNegation,
		})
	}
	return NewPatternTokenMatcher(ex).IsMatched(token)
}

// isAndExceptionGroupMatched ports PatternToken.isAndExceptionGroupMatched:
// true if any AndGroup member has a current exception matching the token.
func (m *PatternTokenMatcher) isAndExceptionGroupMatched(token *languagetool.AnalyzedToken) bool {
	if m == nil || m.Base == nil || token == nil || len(m.Base.AndGroup) == 0 {
		return false
	}
	for _, andTok := range m.Base.AndGroup {
		if andTok == nil {
			continue
		}
		if NewPatternTokenMatcher(andTok).isExceptionMatched(token) {
			return true
		}
	}
	return false
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

// matchSurface ports textMatcher.matches(s) for a candidate surface/lemma string.
func (m *PatternTokenMatcher) matchSurface(surface string) bool {
	if m == nil {
		return false
	}
	if m.textMatcher == nil {
		m.recompileTextMatcher()
	}
	if m.textMatcher == nil {
		return false
	}
	// Empty pattern string always "matches" for TEST_STRING_MASK-off callers.
	if m.textMatcher.Pattern == "" && !m.textMatcher.IsRegExp {
		return true
	}
	return m.textMatcher.Matches(surface)
}
