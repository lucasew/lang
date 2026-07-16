package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MatchState ports org.languagetool.rules.patterns.MatchState (partial; synthesizer deferred).
// Not safe for concurrent use (same as Java).
type MatchState struct {
	Match          *Match
	FormattedToken *languagetool.AnalyzedTokenReadings
	MatchedToken   *languagetool.AnalyzedTokenReadings
	SkippedTokens  string
}

func NewMatchState(match *Match) *MatchState {
	s := &MatchState{Match: match}
	if match != nil && !tools.IsEmptyStr(match.Lemma) {
		pos := match.PosTag
		var p *string
		if pos != "" {
			p = &pos
		}
		lemma := match.Lemma
		s.FormattedToken = languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(lemma, p, &lemma),
		)
	}
	return s
}

// SetToken sets the token to format.
func (s *MatchState) SetToken(token *languagetool.AnalyzedTokenReadings) {
	if s.Match != nil && s.Match.IsStaticLemma() {
		s.MatchedToken = token
	} else {
		s.FormattedToken = token
	}
}

// SetTokenRange sets the token and optional skipped tokens between index and next.
func (s *MatchState) SetTokenRange(tokens []*languagetool.AnalyzedTokenReadings, index, next int) {
	idx := index
	if index >= len(tokens) && len(tokens) > 0 {
		idx = len(tokens) - 1
	}
	if idx >= 0 && idx < len(tokens) {
		s.SetToken(tokens[idx])
	}
	includeSkipped := IncludeNone
	if s.Match != nil {
		includeSkipped = s.Match.GetIncludeSkipped()
	}
	if includeSkipped == IncludeFollowing {
		s.FormattedToken = nil
	}
	if next > 1 && includeSkipped != IncludeNone {
		var b strings.Builder
		for k := index + 1; k < index+next && k < len(tokens); k++ {
			if tokens[k].IsWhitespaceBefore() && !(k == index+1 && includeSkipped == IncludeFollowing) {
				b.WriteByte(' ')
			}
			b.WriteString(tokens[k].GetToken())
		}
		s.SkippedTokens = b.String()
	}
}

// ConvertCase ports MatchState.convertCase via CaseConversionHelper.
// langShortCode enables Dutch "ij" → "IJ" special case when non-empty.
func (s *MatchState) ConvertCase(str, sample, langShortCode string) string {
	if s == nil || s.Match == nil {
		return str
	}
	return ConvertCaseLang(s.Match.GetCaseConversionType(), str, sample, langShortCode)
}
