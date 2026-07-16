package patterns

import "regexp"

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
		m.regexCompiled = regexp.MustCompile(regexMatch)
	}
	if postagRegexp && posTag != "" {
		m.posRegexCompiled = regexp.MustCompile(posTag)
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
		m.posRegexCompiled = regexp.MustCompile(m.PosTag)
	}
}

// CreateState ports Match.createState for a single token.
func (m *Match) CreateState(token interface { /* synthesizer deferred */
}) *MatchState {
	return NewMatchState(m)
}
