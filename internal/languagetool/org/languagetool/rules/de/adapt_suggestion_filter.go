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
// Java uses static GermanTagger.INSTANCE, GermanSynthesizer.INSTANCE, and
// SuggestionFilter(AgreementRule, German) inside acceptRuleMatch.
//
// GenderOf ports getNounGender (optional override; nil → GermanTagger.INSTANCE).
// Synthesize ports GermanSynthesizer.synthesize(..., true) (nil → INSTANCE).
// FilterSuggestions ports SuggestionFilter.filter (nil → AgreementRule wire).
type AdaptSuggestionFilter struct {
	GenderOf          func(word string) string
	Synthesize        func(lemma, postagRE string) []string
	FilterSuggestions func(suggs []string, template string) []string
}

func NewAdaptSuggestionFilter() *AdaptSuggestionFilter {
	return &AdaptSuggestionFilter{}
}

var (
	// Java private static final Patterns (String.matches = full-string).
	artProPattern     = regexp.MustCompile(`^(ART|PRO):.*$`)
	masFemNeuPattern  = regexp.MustCompile(`MAS|FEM|NEU`)
	proPosEtcPattern  = regexp.MustCompile(`PRO:POS:(NOM|AKK|GEN|DAT):(SIN|PLU):(MAS|FEM|NEU)`)
	stvBegPattern     = regexp.MustCompile(`:(STV|BEG).*`)
	artProBasePattern = regexp.MustCompile(`^(ART|PRO):.*$`)
	adjBasePattern    = regexp.MustCompile(`^ADJ:.*$`)
)

// DetReading is one determiner reading for unit tests of getAdaptedDet
// (Java AdaptSuggestionFilterTest.assertDet constructs a single AnalyzedToken).
type DetReading struct {
	Token string
	POS   string
	Lemma string
}

// AdaptedDet ports getAdaptedDet for a simplified single-reading determiner
// (Java assertDet).
func (f *AdaptSuggestionFilter) AdaptedDet(det DetReading, repl string) []string {
	if f == nil {
		return nil
	}
	return f.GetAdaptedDet(detReadingToATR(det), repl)
}

// GetAdaptedDet ports AdaptSuggestionFilter.getAdaptedDet (package-visible in Java).
func (f *AdaptSuggestionFilter) GetAdaptedDet(detToken *languagetool.AnalyzedTokenReadings, repl string) []string {
	return f.getAdaptedDet(detToken, repl)
}

// GetAdaptedDetAdj ports AdaptSuggestionFilter.getAdaptedDetAdj (package-visible in Java).
func (f *AdaptSuggestionFilter) GetAdaptedDetAdj(detToken, adjToken *languagetool.AnalyzedTokenReadings, repl string) []string {
	return f.getAdaptedDetAdj(detToken, adjToken, repl)
}

// AcceptRuleMatch ports AdaptSuggestionFilter.acceptRuleMatch.
func (f *AdaptSuggestionFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, patternTokenPos int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	var newMatch *rules.RuleMatch
	var newSugg []string
	if patternTokenPos > 0 && match.Sentence != nil {
		tokens := match.Sentence.GetTokensWithoutWhitespace()
		if patternTokenPos-1 < len(tokens) && patternTokenPos-1 >= 0 {
			prevToken := tokens[patternTokenPos-1]
			detNoun := prevToken != nil &&
				(prevToken.HasPosTagStartingWith("ART:") || prevToken.HasPosTagStartingWith("PRO:"))
			// Java: detAdjNoun && false — branch kept but disabled (needs more testing upstream).
			detAdjNoun := false && patternTokenPos > 1 &&
				prevToken != nil &&
				(tokens[patternTokenPos-2] != nil &&
					(tokens[patternTokenPos-2].HasPosTagStartingWith("ART:") ||
						tokens[patternTokenPos-2].HasPosTagStartingWith("PRO:"))) &&
				prevToken.HasPosTagStartingWith("ADJ:")
			if detNoun {
				newMatch = rules.NewRuleMatch(match.GetRule(), match.Sentence, prevToken.GetStartPos(), match.GetToPos(), match.GetMessage())
				newMatch.ShortMessage = match.ShortMessage
				for _, replacement := range match.GetSuggestedReplacements() {
					for _, adaptedDet := range f.getAdaptedDet(prevToken, replacement) {
						newRepl := adaptedDet + " " + replacement
						if !containsString(newSugg, newRepl) {
							newSugg = append(newSugg, newRepl)
						}
					}
				}
				newMatch.SetSuggestedReplacements(newSugg)
			} else if detAdjNoun {
				// Dead in Java (&& false); structure retained for 1:1 control flow.
				prevPrevToken := tokens[patternTokenPos-2]
				newMatch = rules.NewRuleMatch(match.GetRule(), match.Sentence, prevPrevToken.GetStartPos(), match.GetToPos(), match.GetMessage())
				newMatch.ShortMessage = match.ShortMessage
				for _, replacement := range match.GetSuggestedReplacements() {
					for _, adaptedDet := range f.getAdaptedDetAdj(prevPrevToken, prevToken, replacement) {
						newSugg = append(newSugg, adaptedDet+" "+replacement)
					}
				}
				newMatch.SetSuggestedReplacements(newSugg)
			}
		}
	}
	// Java: always construct AgreementRule + SuggestionFilter; filter only when newSugg non-empty.
	if len(newSugg) > 0 && newMatch != nil {
		filterFn := f.FilterSuggestions
		if filterFn == nil {
			filterFn = defaultAgreementSuggestionFilter()
		}
		if filterFn != nil {
			tpl := "Das ist {}."
			if tools.StartsWithUppercase(newSugg[0]) {
				tpl = "{} ist das."
			}
			newSugg = filterFn(newSugg, tpl)
			newMatch.SetSuggestedReplacements(newSugg)
		}
		return newMatch
	}
	return match
}

// getAdaptedDet ports AdaptSuggestionFilter.getAdaptedDet.
func (f *AdaptSuggestionFilter) getAdaptedDet(detToken *languagetool.AnalyzedTokenReadings, repl string) []string {
	if f == nil || detToken == nil {
		return nil
	}
	oldDetBaseform := getBaseform(detToken, artProBasePattern)
	replGender := f.genderOf(repl)
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
		// Java: replaceAll MAS|FEM|NEU → gender; BEG → (BEG|B/S); strip :STV
		newDetPos := masFemNeuPattern.ReplaceAllString(pos, replGender)
		newDetPos = strings.ReplaceAll(newDetPos, "BEG", "(BEG|B/S)")
		newDetPos = strings.ReplaceAll(newDetPos, ":STV", "")
		for _, s := range f.synthesize(oldDetBaseform, newDetPos) {
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
	return result
}

// getAdaptedDetAdj ports AdaptSuggestionFilter.getAdaptedDetAdj (Java package-visible;
// acceptRuleMatch detAdjNoun path is disabled with && false).
func (f *AdaptSuggestionFilter) getAdaptedDetAdj(detToken, adjToken *languagetool.AnalyzedTokenReadings, repl string) []string {
	if f == nil || detToken == nil || adjToken == nil {
		return nil
	}
	oldDetBaseform := getBaseform(detToken, artProBasePattern)
	oldAdjBaseform := getBaseform(adjToken, adjBasePattern)
	replGender := f.genderOf(repl)
	if replGender == "" || oldDetBaseform == "" || oldAdjBaseform == "" {
		return nil
	}
	var result []string
	detSurface := detToken.GetToken()
	firstRune, _ := utf8.DecodeRuneInString(detSurface)
	firstStr := string(firstRune)

	for _, reading := range detToken.GetReadings() {
		if reading == nil || reading.GetPOSTag() == nil {
			continue
		}
		pos := *reading.GetPOSTag()
		// Java: ART_PRO_PATTERN.matcher(pos).matches()
		if !artProPattern.MatchString(pos) {
			continue
		}
		// Java newDetPos: gender replace + BEG expand (no :STV strip on det)
		newDetPos := masFemNeuPattern.ReplaceAllString(pos, replGender)
		newDetPos = strings.ReplaceAll(newDetPos, "BEG", "(BEG|B/S)")
		var newAdjPos string
		if strings.HasPrefix(newDetPos, "ART:") {
			// Java uses original reading POS for adj path when ART
			newAdjPos = masFemNeuPattern.ReplaceAllString(pos, replGender)
			newAdjPos = strings.ReplaceAll(newAdjPos, "BEG", "(BEG|B/S)")
			newAdjPos = strings.ReplaceAll(newAdjPos, ":STV", "")
		} else if strings.HasPrefix(newDetPos, "PRO:") {
			// PRO:POS:(CASE):(NUM):(GENDER)… → ADJ:$1:$2:$3 then first :(STV|BEG).* → :GRU:IND
			newAdjPos = proPosEtcPattern.ReplaceAllString(newDetPos, "ADJ:$1:$2:$3")
			newAdjPos = stvBegPattern.ReplaceAllString(newAdjPos, ":GRU:IND")
		} else {
			// Java: throw new RuntimeException("Unexpected POS tag: " + newDetPos)
			panic("Unexpected POS tag: " + newDetPos)
		}
		replDet := f.synthesize(oldDetBaseform, newDetPos)
		replAdj := f.synthesize(oldAdjBaseform, newAdjPos)
		for _, det := range replDet {
			if firstStr != "" && !strings.HasPrefix(det, firstStr) {
				continue // mein/dein/sein share lemma; keep surface family
			}
			for _, adj := range replAdj {
				newDetAdj := det + " " + adj
				if !containsString(result, newDetAdj) {
					result = append(result, newDetAdj)
				}
			}
		}
	}
	return result
}

func (f *AdaptSuggestionFilter) genderOf(word string) string {
	if f != nil && f.GenderOf != nil {
		return f.GenderOf(word)
	}
	return nounGenderFromGermanTagger(word)
}

func (f *AdaptSuggestionFilter) synthesize(lemma, postagRE string) []string {
	if f != nil && f.Synthesize != nil {
		return f.Synthesize(lemma, postagRE)
	}
	return synthesizeGermanRE(lemma, postagRE)
}

// getBaseform ports AdaptSuggestionFilter.getBaseform(token, tagStartsWith regex).
// Java: reading.getPOSTag().matches(tagStartsWith) — full-string match.
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
			} else {
				baseform = ""
			}
			// Java: last match wins (TODO if more than one)
		}
	}
	return baseform
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

func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
