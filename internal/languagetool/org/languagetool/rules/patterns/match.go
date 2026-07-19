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
	PosTagReplace      string
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
}

// compileMatchRE compiles a Java-oriented regex for Match attributes.
// Invalid patterns yield nil (no invent rewrite); Java (?iu)/(?ui) → Go (?i).
func compileMatchRE(s string) *regexp.Regexp {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Go RE2 has no inline u flag; case-insensitive is (?i).
	s = strings.ReplaceAll(s, "(?iu)", "(?i)")
	s = strings.ReplaceAll(s, "(?ui)", "(?i)")
	s = strings.ReplaceAll(s, "(?u)", "")
	re, err := regexp.Compile(s)
	if err != nil {
		return nil
	}
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
		CaseConversionType: caseConversionType,
		SetPos:             setPOS,
		SuppressMisspelled: suppressMisspelled,
		IncludeSkipped:     includeSkipped,
	}
	if regexMatch != "" {
		m.regexCompiled = compileMatchRE(regexMatch)
	}
	if postagRegexp && posTag != "" {
		m.posRegexCompiled = compileMatchRE(posTag)
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

// SetLemmaString ports Match.setLemmaString.
func (m *Match) SetLemmaString(lemmaString string) {
	if lemmaString == "" {
		return
	}
	m.Lemma = lemmaString
	m.StaticLemma = true
	m.PostagRegexp = true
	if m.PosTag != "" {
		m.posRegexCompiled = compileMatchRE(m.PosTag)
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
