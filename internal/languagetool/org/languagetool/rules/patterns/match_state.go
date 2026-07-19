package patterns

import (
	"regexp"
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MatchState ports org.languagetool.rules.patterns.MatchState.
// Not safe for concurrent use (same as Java).
type MatchState struct {
	Match          *Match
	Synthesizer    synthesis.Synthesizer
	FormattedToken *languagetool.AnalyzedTokenReadings
	MatchedToken   *languagetool.AnalyzedTokenReadings
	SkippedTokens  string
}

func NewMatchState(match *Match) *MatchState {
	return NewMatchStateWithSynth(match, nil)
}

// NewMatchStateWithSynth ports MatchState(Match, Synthesizer).
func NewMatchStateWithSynth(match *Match, synth synthesis.Synthesizer) *MatchState {
	s := &MatchState{Match: match, Synthesizer: synth}
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
func (s *MatchState) ConvertCase(str, sample, langShortCode string) string {
	if s == nil || s.Match == nil {
		return str
	}
	return ConvertCaseLang(s.Match.GetCaseConversionType(), str, sample, langShortCode)
}

// ToFinalString ports MatchState.toFinalString.
// When Synthesizer is nil and postag is set, returns surface after regex (Java synthesizer==null).
// Empty synthesis yields "(token)" like Java — not an invent form.
func (s *MatchState) ToFinalString(langCode string) []string {
	if s == nil {
		return []string{""}
	}
	formatted := []string{""}
	if s.FormattedToken != nil {
		surface := s.FormattedToken.GetToken()
		if s.Match != nil {
			if re := s.Match.GetRegexMatch(); re != nil {
				if langCode == "ar" {
					surface = tools.RemoveTashkeel(surface)
				}
				surface = re.ReplaceAllString(surface, s.Match.GetRegexReplace())
			}
			posTag := s.Match.GetPosTag()
			if posTag != "" {
				if s.Synthesizer == nil {
					// Java: synthesizer == null → original token (before regex is overwritten)
					surface = s.FormattedToken.GetToken()
					formatted = []string{surface}
				} else if s.Match.IsPostagRegexp() {
					formatted = s.synthesizeRegexpPOS(posTag)
				} else {
					formatted = s.synthesizeExactPOS(posTag)
				}
			} else {
				formatted = []string{surface}
			}
		} else {
			formatted = []string{surface}
		}
	}

	original := ""
	if s.Match != nil && s.Match.IsStaticLemma() {
		if s.MatchedToken != nil {
			original = s.MatchedToken.GetToken()
		}
	} else if s.FormattedToken != nil {
		original = s.FormattedToken.GetToken()
	}
	baseLang := langCode
	if i := strings.IndexByte(baseLang, '-'); i > 0 {
		baseLang = baseLang[:i]
	}
	for i := range formatted {
		if formatted[i] == "" && formatted[i] != " " {
			// keep empty
		}
		formatted[i] = s.ConvertCase(formatted[i], original, baseLang)
	}
	if s.Match != nil && s.Match.GetIncludeSkipped() != IncludeNone && s.SkippedTokens != "" {
		for i := range formatted {
			formatted[i] = formatted[i] + s.SkippedTokens
		}
	}
	// Java: match.checksSpelling() && tagger finds no lemma/tag → MISTAKE marker.
	if s.Match != nil && s.Match.ChecksSpelling() {
		for i := range formatted {
			if IsUnknownToTagger(baseLang, formatted[i]) {
				formatted[i] = MistakeMarker
			}
		}
	}
	return formatted
}

func (s *MatchState) synthesizeExactPOS(posTag string) []string {
	wordForms := map[string]struct{}{}
	readings := s.FormattedToken.GetReadings()
	for _, r := range readings {
		if r == nil {
			continue
		}
		forms, err := s.Synthesizer.Synthesize(r, posTag)
		if err != nil || len(forms) == 0 {
			continue
		}
		for _, f := range forms {
			if f != "" {
				wordForms[f] = struct{}{}
			}
		}
	}
	return sortedFormsOrParen(wordForms, s.FormattedToken.GetToken())
}

func (s *MatchState) synthesizeRegexpPOS(posTag string) []string {
	wordForms := map[string]struct{}{}
	readings := s.FormattedToken.GetReadings()
	oneForm := false
	for _, r := range readings {
		if r == nil {
			continue
		}
		if r.GetLemma() == nil {
			posUnique := ""
			if p := r.GetPOSTag(); p != nil {
				posUnique = *p
			}
			if posUnique == "" {
				wordForms[s.FormattedToken.GetToken()] = struct{}{}
				oneForm = true
			} else if posUnique == languagetool.SentenceStartTagName ||
				posUnique == languagetool.SentenceEndTagName ||
				posUnique == languagetool.ParagraphEndTagName {
				if !oneForm {
					wordForms[s.FormattedToken.GetToken()] = struct{}{}
				}
				oneForm = true
			} else {
				oneForm = false
			}
		}
	}
	targetPosTag := s.GetTargetPosTag()
	if !oneForm {
		for _, r := range readings {
			if r == nil {
				continue
			}
			forms, err := s.Synthesizer.SynthesizeRE(r, targetPosTag, true)
			if err != nil || len(forms) == 0 {
				continue
			}
			for _, f := range forms {
				if f != "" {
					wordForms[f] = struct{}{}
				}
			}
		}
	}
	return sortedFormsOrParen(wordForms, s.FormattedToken.GetToken())
}

func sortedFormsOrParen(wordForms map[string]struct{}, token string) []string {
	if len(wordForms) == 0 {
		// Java: "(" + token + ")" when synthesis finds nothing
		return []string{"(" + token + ")"}
	}
	out := make([]string, 0, len(wordForms))
	for f := range wordForms {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// reFullMatch ports Java Matcher.matches() / String.matches (entire string).
func reFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

// GetTargetPosTag ports MatchState.getTargetPosTag including getPosTagCorrection when setpos.
func (s *MatchState) GetTargetPosTag() string {
	if s == nil || s.Match == nil {
		return ""
	}
	targetPosTag := s.Match.GetPosTag()
	pPos := s.Match.GetPosRegexMatch()
	posTagReplace := s.Match.GetPosTagReplace()
	var posTags []string

	source := s.FormattedToken
	if s.Match.IsStaticLemma() {
		source = s.MatchedToken
	}
	if source != nil && pPos != nil {
		for _, r := range source.GetReadings() {
			if r == nil {
				continue
			}
			tst := ""
			if p := r.GetPOSTag(); p != nil {
				tst = *p
			}
			// Java: pPosRegexMatch.matcher(tst).matches()
			if tst != "" && reFullMatch(pPos, tst) {
				posTags = append(posTags, tst)
			}
		}
	}
	// language-specific pick if synthesizer supports it
	if bs, ok := s.Synthesizer.(interface {
		GetTargetPosTag([]string, string) string
	}); ok {
		targetPosTag = bs.GetTargetPosTag(posTags, targetPosTag)
	} else if len(posTags) > 0 {
		targetPosTag = posTags[0]
	}
	if pPos != nil && posTagReplace != "" {
		if s.Match.IsStaticLemma() {
			if len(posTags) > 0 {
				targetPosTag = pPos.ReplaceAllString(targetPosTag, posTagReplace)
			}
		} else {
			if len(posTags) == 0 {
				posTags = append(posTags, targetPosTag)
			}
			var parts []string
			for _, lPosTag := range posTags {
				lPosTag = pPos.ReplaceAllString(lPosTag, posTagReplace)
				// Java: if match.setsPos() → synthesizer.getPosTagCorrection(lPosTag)
				if s.Match.SetsPos() {
					if corr, ok := s.Synthesizer.(interface {
						GetPosTagCorrection(string) string
					}); ok {
						lPosTag = corr.GetPosTagCorrection(lPosTag)
					}
				}
				parts = append(parts, lPosTag)
			}
			targetPosTag = strings.Join(parts, "|")
		}
	}
	return targetPosTag
}
