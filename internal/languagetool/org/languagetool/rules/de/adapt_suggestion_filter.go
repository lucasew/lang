package de

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AdaptSuggestionFilter ports org.languagetool.rules.de.AdaptSuggestionFilter.
// Java uses GermanTagger + GermanSynthesizer + AgreementRule SuggestionFilter.
//
// GenderOf ports getNounGender (tag noun → MAS|FEM|NEU). Nil → fail-closed empty adapts.
// Synthesize ports GermanSynthesizer.synthesize(token, postagRE, true). Nil → fail-closed.
// FilterSuggestions ports SuggestionFilter.filter (optional; nil keeps adapted suggestions as-is).
type AdaptSuggestionFilter struct {
	GenderOf          func(word string) string
	Synthesize        func(lemma, postagRE string) []string
	FilterSuggestions func(suggs []string, template string) []string
}

func NewAdaptSuggestionFilter() *AdaptSuggestionFilter {
	return &AdaptSuggestionFilter{}
}

var (
	masFemNeuPattern  = regexp.MustCompile(`MAS|FEM|NEU`)
	artProBasePattern = regexp.MustCompile(`^(ART|PRO):`)
)

// DetReading is one determiner reading for unit tests of getAdaptedDet.
type DetReading struct {
	Token string
	POS   string // e.g. ART:DEF:NOM:SIN:FEM
	Lemma string
}

// AdaptedDet ports getAdaptedDet for a simplified single-reading determiner.
func (f *AdaptSuggestionFilter) AdaptedDet(det DetReading, repl string) []string {
	if f == nil {
		return nil
	}
	return f.getAdaptedDet(detReadingToATR(det), repl)
}

// SuggestWithDet rewrites "det + noun" suggestions when the previous token is a det.
func (f *AdaptSuggestionFilter) SuggestWithDet(prevToken, prevPOS, prevLemma string, replacements []string) []string {
	if f == nil {
		return nil
	}
	return f.SuggestWithDetFromATR(detReadingToATR(DetReading{Token: prevToken, POS: prevPOS, Lemma: prevLemma}), replacements)
}

// SuggestWithDetFromATR ports the detNoun branch suggestion building.
func (f *AdaptSuggestionFilter) SuggestWithDetFromATR(prev *languagetool.AnalyzedTokenReadings, replacements []string) []string {
	if f == nil || prev == nil {
		return nil
	}
	var out []string
	seen := map[string]struct{}{}
	for _, repl := range replacements {
		for _, ad := range f.getAdaptedDet(prev, repl) {
			s := ad + " " + repl
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

// AcceptRuleMatch ports AdaptSuggestionFilter.acceptRuleMatch (active det+noun path only).
// Java detAdjNoun branch is disabled (&& false) — not ported (would be invent).
func (f *AdaptSuggestionFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, patternTokenPos int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	var newSugg []string
	var newMatch *rules.RuleMatch
	if patternTokenPos > 0 && match.Sentence != nil {
		tokens := match.Sentence.GetTokensWithoutWhitespace()
		if patternTokenPos-1 < len(tokens) {
			prevToken := tokens[patternTokenPos-1]
			if prevToken != nil && (prevToken.HasPosTagStartingWith("ART:") || prevToken.HasPosTagStartingWith("PRO:")) {
				newMatch = rules.NewRuleMatch(match.GetRule(), match.Sentence, prevToken.GetStartPos(), match.GetToPos(), match.GetMessage())
				newMatch.ShortMessage = match.ShortMessage
				newSugg = f.SuggestWithDetFromATR(prevToken, match.GetSuggestedReplacements())
				newMatch.SetSuggestedReplacements(newSugg)
			}
		}
	}
	// Java: SuggestionFilter via AgreementRule when newSugg non-empty.
	if len(newSugg) > 0 && newMatch != nil {
		if f.FilterSuggestions != nil {
			tpl := "Das ist {}."
			if tools.StartsWithUppercase(newSugg[0]) {
				tpl = "{} ist das."
			}
			newSugg = f.FilterSuggestions(newSugg, tpl)
			newMatch.SetSuggestedReplacements(newSugg)
		}
		return newMatch
	}
	return match
}

// getAdaptedDet ports AdaptSuggestionFilter.getAdaptedDet.
func (f *AdaptSuggestionFilter) getAdaptedDet(detToken *languagetool.AnalyzedTokenReadings, repl string) []string {
	if f == nil || detToken == nil || f.GenderOf == nil || f.Synthesize == nil {
		return nil
	}
	oldDetBaseform := getBaseform(detToken, artProBasePattern)
	replGender := f.GenderOf(repl)
	if replGender == "" || oldDetBaseform == "" {
		return nil
	}
	var result []string
	detSurface := detToken.GetToken()
	isTitle := tools.StartsWithUppercase(detSurface)
	firstRune, _ := utf8.DecodeRuneInString(detSurface)
	firstStr := string(firstRune)
	firstLower := strings.ToLower(firstStr)

	for _, reading := range detToken.GetReadings() {
		if reading == nil || reading.GetPOSTag() == nil {
			continue
		}
		pos := *reading.GetPOSTag()
		if !strings.HasPrefix(pos, "ART:") && !strings.HasPrefix(pos, "PRO:") {
			continue
		}
		// Java: replaceAll MAS|FEM|NEU with gender; BEG → (BEG|B/S); strip :STV
		newDetPos := masFemNeuPattern.ReplaceAllString(pos, replGender)
		newDetPos = strings.ReplaceAll(newDetPos, "BEG", "(BEG|B/S)")
		newDetPos = strings.ReplaceAll(newDetPos, ":STV", "")
		for _, s := range f.Synthesize(oldDetBaseform, newDetPos) {
			if s == "" {
				continue
			}
			if isTitle {
				// Java: s.toLowerCase().startsWith(det first char lower)
				if !strings.HasPrefix(strings.ToLower(s), firstLower) {
					continue
				}
				result = append(result, tools.UppercaseFirstChar(s))
			} else {
				// Java: s.startsWith(detToken.getToken().substring(0, 1))
				if firstStr != "" && !strings.HasPrefix(s, firstStr) {
					continue
				}
				result = append(result, s)
			}
		}
	}
	return uniqueStrings(result)
}

func detReadingToATR(det DetReading) *languagetool.AnalyzedTokenReadings {
	var pos, lemma *string
	if det.POS != "" {
		p := det.POS
		pos = &p
	}
	if det.Lemma != "" {
		l := det.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(det.Token, pos, lemma), 0)
}

// getBaseform ports AdaptSuggestionFilter.getBaseform(token, tagStartsWith regex).
func getBaseform(token *languagetool.AnalyzedTokenReadings, tagStartsWith *regexp.Regexp) string {
	if token == nil {
		return ""
	}
	var baseform string
	for _, reading := range token.GetReadings() {
		if reading == nil || reading.GetPOSTag() == nil {
			continue
		}
		pos := *reading.GetPOSTag()
		if tagStartsWith.MatchString(pos) {
			if reading.GetLemma() != nil {
				baseform = *reading.GetLemma()
			}
		}
	}
	return baseform
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
