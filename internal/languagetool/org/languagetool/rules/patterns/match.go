package patterns

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Match ports org.languagetool.rules.patterns.Match — configuration of a <match/> element.
type Match struct {
	PosTag             string
	SuppressMisspelled bool
	RegexReplace       string
	// RegexReplacePresent is true when Java regexReplace != null (attr present).
	// filterReadings only replaces when both regexp_match and regexp_replace are non-null.
	RegexReplacePresent bool
	PosTagReplace string
	// PosTagReplacePresent is true when Java posTagReplace != null (postag_replace attr).
	PosTagReplacePresent bool
	CaseConversionType CaseConversion
	IncludeSkipped     IncludeRange
	RegexMatch         string // raw pattern; compiled lazily
	SetPos             bool
	PostagRegexp       bool
	StaticLemma        bool
	Lemma              string
	TokenRef           int
	InMessageOnly      bool
	regexCompiled      *regexp.Regexp
	posRegexCompiled   *regexp.Regexp
	// javaRE engines when RE2 cannot compile lookaround (Java Pattern).
	regexJavaRE *javaRegexp
	posJavaRE   *javaRegexp
}

// compileMatchPattern compiles a Java-oriented regex for Match attributes.
// Prefer RE2; on lookaround syntax fall back to javaRegexp (full-string match).
// Returns (re, javaRE) — at most one non-nil on success; both nil if uncompilable.
func compileMatchPattern(s string) (*regexp.Regexp, *javaRegexp) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	// Go RE2 has no inline u flag; case-insensitive is (?i).
	s = strings.ReplaceAll(s, "(?iu)", "(?i)")
	s = strings.ReplaceAll(s, "(?ui)", "(?i)")
	s = strings.ReplaceAll(s, "(?u)", "")
	s = normalizeJavaRegexp(s)
	re, err := regexp.Compile(s)
	if err == nil {
		return re, nil
	}
	if needsJavaRegexp(s) {
		if jr, jerr := compileJavaRegexp(s, true); jerr == nil {
			return nil, jr
		}
	}
	return nil, nil
}

// compileMatchRE compiles RE2 only (legacy helpers / tests that need *regexp.Regexp).
// Lookaround patterns return nil — use compileMatchPattern / PosFullMatch instead.
func compileMatchRE(s string) *regexp.Regexp {
	re, _ := compileMatchPattern(s)
	return re
}

// NewMatch constructs a Match (ports the Java constructor).
func NewMatch(
	posTag, posTagReplace string,
	postagRegexp bool,
	regexMatch, regexReplace string,
	caseConversionType CaseConversion,
	setPOS, suppressMisspelled bool,
	includeSkipped IncludeRange,
) *Match {
	m := &Match{
		PosTag:             posTag,
		PosTagReplace:      posTagReplace,
		PostagRegexp:       postagRegexp,
		RegexMatch:         regexMatch,
		RegexReplace:       regexReplace,
		// Non-empty replace implies attr present; empty replace with match-only → Present false
		// (callers that need empty-string replace set *Present after NewMatch).
		RegexReplacePresent:  regexReplace != "",
		PosTagReplacePresent: posTagReplace != "",
		CaseConversionType:   caseConversionType,
		SetPos:               setPOS,
		SuppressMisspelled:   suppressMisspelled,
		IncludeSkipped:       includeSkipped,
	}
	if regexMatch != "" {
		m.regexCompiled, m.regexJavaRE = compileMatchPattern(regexMatch)
	}
	if postagRegexp && posTag != "" {
		m.posRegexCompiled, m.posJavaRE = compileMatchPattern(posTag)
	}
	return m
}

func (m *Match) SetsPos() bool                         { return m != nil && m.SetPos }
func (m *Match) PosRegExp() bool                       { return m != nil && m.PostagRegexp }
func (m *Match) ChecksSpelling() bool                  { return m != nil && m.SuppressMisspelled }
func (m *Match) ConvertsCase() bool                    { return m != nil && m.CaseConversionType != CaseNone }
func (m *Match) GetCaseConversionType() CaseConversion { return m.CaseConversionType }
func (m *Match) GetLemma() string                      { return m.Lemma }
func (m *Match) IsStaticLemma() bool                   { return m.StaticLemma }
func (m *Match) GetTokenRef() int                      { return m.TokenRef }
func (m *Match) SetTokenRef(i int)                     { m.TokenRef = i }
func (m *Match) SetInMessageOnly(v bool)               { m.InMessageOnly = v }
func (m *Match) IsInMessageOnly() bool                 { return m.InMessageOnly }
func (m *Match) GetPosTag() string                     { return m.PosTag }
func (m *Match) GetRegexReplace() string               { return m.RegexReplace }
func (m *Match) GetPosTagReplace() string              { return m.PosTagReplace }
func (m *Match) GetIncludeSkipped() IncludeRange       { return m.IncludeSkipped }
func (m *Match) IsPostagRegexp() bool                  { return m.PostagRegexp }
func (m *Match) GetRegexMatch() *regexp.Regexp         { return m.regexCompiled }
func (m *Match) GetPosRegexMatch() *regexp.Regexp      { return m.posRegexCompiled }

// HasPosRegexp reports whether a POS regex (RE2 or lookaround) is available.
func (m *Match) HasPosRegexp() bool {
	return m != nil && (m.posRegexCompiled != nil || m.posJavaRE != nil)
}

// HasSurfaceRegexp reports whether a surface regex (RE2 or lookaround) is available.
func (m *Match) HasSurfaceRegexp() bool {
	return m != nil && (m.regexCompiled != nil || m.regexJavaRE != nil)
}

// HasSurfaceReplace ports filterReadings gate: regexMatch != null && regexReplace != null.
func (m *Match) HasSurfaceReplace() bool {
	return m.HasSurfaceRegexp() && m.RegexReplacePresent
}

// PosFullMatch ports Java Matcher.matches() against the POS pattern.
func (m *Match) PosFullMatch(s string) bool {
	if m == nil {
		return false
	}
	if m.posRegexCompiled != nil {
		return reFullMatch(m.posRegexCompiled, s)
	}
	if m.posJavaRE != nil {
		return m.posJavaRE.fullMatch(s)
	}
	return false
}

// SurfaceReplace applies regexp_match/replace when RE2-backed and replace is present.
// Java filterReadings requires both non-null; toFinalString only checks match != null
// (null replace would NPE — Present false skips invent empty-replace wipe).
func (m *Match) SurfaceReplace(s string) string {
	if m == nil || !m.RegexReplacePresent {
		return s
	}
	if m.regexCompiled != nil {
		return m.regexCompiled.ReplaceAllString(s, m.RegexReplace)
	}
	// javaRE has no replace — return original (Java would still replace; rare in LT).
	return s
}

// SetLemmaString ports Match.setLemmaString.
func (m *Match) SetLemmaString(lemmaString string) {
	if lemmaString == "" {
		return
	}
	m.Lemma = lemmaString
	m.StaticLemma = true
	m.PostagRegexp = true
	if m.PosTag != "" {
		m.posRegexCompiled, m.posJavaRE = compileMatchPattern(m.PosTag)
	}
}

// CreateState ports Match.createState(synthesizer, token) with nil synthesizer.
func (m *Match) CreateState() *MatchState {
	return NewMatchState(m)
}

// CreateStateWithSynth ports Match.createState(Synthesizer, AnalyzedTokenReadings).
func (m *Match) CreateStateWithSynth(synth synthesis.Synthesizer, token *languagetool.AnalyzedTokenReadings) *MatchState {
	st := NewMatchStateWithSynth(m, synth)
	if token != nil {
		st.SetToken(token)
	}
	return st
}

// CreateStateRange ports Match.createState(Synthesizer, AnalyzedTokenReadings[], index, next).
func (m *Match) CreateStateRange(synth synthesis.Synthesizer, tokens []*languagetool.AnalyzedTokenReadings, index, next int) *MatchState {
	st := NewMatchStateWithSynth(m, synth)
	st.SetTokenRange(tokens, index, next)
	return st
}
