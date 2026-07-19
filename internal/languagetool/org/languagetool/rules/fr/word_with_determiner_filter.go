package fr

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WordWithDeterminerFilter ports
// org.languagetool.rules.fr.WordWithDeterminerFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports FrenchSynthesizer.synthesize(token, postagRE, true).
// ValidateSuggestion ports suggestionHasNoErrors (CAT_ELISION / CET_CE / …);
// when nil, all generated forms are kept (host must wire real LT check).
type WordWithDeterminerFilter struct {
	// Synthesize ports FrenchSynthesizer.INSTANCE.synthesize(token, postag, true).
	Synthesize func(tok *languagetool.AnalyzedToken, postagRE string) []string
	// ValidateSuggestion returns true if the det+word string has no elision errors.
	ValidateSuggestion func(suggestion string) bool
}

func NewWordWithDeterminerFilter() *WordWithDeterminerFilter {
	return &WordWithDeterminerFilter{}
}

// Java Pattern.compile strings (Matcher.matches = full string).
var (
	detPattern  = regexp.MustCompile(`(P.)?D .*|J .*|V.* ppa .*`)
	wordPattern = regexp.MustCompile(`[ZNJ] .*|V.* ppa .*`)
	// 0=MS, 1=FS, 2=MP, 3=FP
	genderNumber = []string{
		`([me]) (s|sp)`,
		`([fe]) (s|sp)`,
		`([me]) (p|sp)`,
		`([fe]) (p|sp)`,
	}
	determinerPrefix = `((P.)?D |J |V.* ppa )`
)

// ExceptionsDeterminer are irregular plural det forms that skip some rewrites.
var ExceptionsDeterminer = map[string]struct{}{
	"bels": {}, "fols": {}, "mols": {}, "nouvels": {},
}

// ElisionRulesToCheck are French rule IDs related to elision for post-filtering.
var ElisionRulesToCheck = []string{"CET_CE", "CE_CET", "MA_VOYELLE", "MON_NFS", "VIEUX"}

// CategoryToCheck is the category id used when validating elision.
const CategoryToCheck = "CAT_ELISION"

// Legacy aliases used by unit tests.
var (
	DetPOS               = detPattern
	WordPOS              = wordPattern
	GenderNumberPatterns = genderNumber
)

// IsExceptionDeterminer reports irregular determiner plurals.
func (f *WordWithDeterminerFilter) IsExceptionDeterminer(token string) bool {
	_, ok := ExceptionsDeterminer[token]
	return ok
}

// MatchesDetPOS / MatchesWordPOS check POS patterns (full match).
func (f *WordWithDeterminerFilter) MatchesDetPOS(pos string) bool {
	return wwdFullMatch(detPattern, pos)
}
func (f *WordWithDeterminerFilter) MatchesWordPOS(pos string) bool {
	return wwdFullMatch(wordPattern, pos)
}

// NounAdjPrefix returns the synthesizer prefix for noun-only / adj-only / both.
func (f *WordWithDeterminerFilter) NounAdjPrefix(isNoun, isAdjective bool) string {
	if isNoun && !isAdjective {
		return "[NZ] "
	}
	if !isNoun && isAdjective {
		return "J "
	}
	return "[ZNJ] "
}

func wwdFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

func getAnalyzedTokenWW(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) *languagetool.AnalyzedToken {
	if aToken == nil || pattern == nil {
		return nil
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if pt := analyzedToken.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if wwdFullMatch(pattern, posTag) {
			return analyzedToken
		}
	}
	return nil
}

// AcceptRuleMatch ports WordWithDeterminerFilter.acceptRuleMatch.
func (f *WordWithDeterminerFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = patternTokenPos
	_ = tokenPositions
	if f == nil || match == nil {
		return nil
	}
	if arguments == nil {
		panic("WordWithDeterminerFilter: undefined parameters wordFrom or determinerFrom")
	}
	wordFrom, ok1 := arguments["wordFrom"]
	determinerFrom, ok2 := arguments["determinerFrom"]
	if !ok1 || !ok2 || wordFrom == "" || determinerFrom == "" {
		panic("WordWithDeterminerFilter: undefined parameters wordFrom or determinerFrom in rule")
	}
	posWord, err1 := strconv.Atoi(wordFrom)
	posDeterminer, err2 := strconv.Atoi(determinerFrom)
	if err1 != nil || err2 != nil {
		panic("WordWithDeterminerFilter: invalid wordFrom/determinerFrom")
	}
	if posWord < 1 || posWord > len(patternTokens) {
		panic("WordWithDeterminerFilter: Index out of bounds, wordFrom: " + wordFrom)
	}
	if posDeterminer < 1 || posDeterminer > len(patternTokens) {
		panic("WordWithDeterminerFilter: Index out of bounds, posDeterminer: " + determinerFrom)
	}

	atrDeterminer := patternTokens[posDeterminer-1]
	atrWord := patternTokens[posWord-1]
	isDeterminerCapitalized := tools.IsCapitalizedWord(atrDeterminer.GetToken())
	isWordCapitalized := tools.IsCapitalizedWord(atrWord.GetToken())
	isDeterminerAllupper := tools.IsAllUppercase(atrDeterminer.GetToken()) &&
		!strings.EqualFold(atrDeterminer.GetToken(), "L'")
	isWordAllupper := tools.IsAllUppercase(atrWord.GetToken())

	atDeterminer := getAnalyzedTokenWW(atrDeterminer, detPattern)
	atWord := getAnalyzedTokenWW(atrWord, wordPattern)
	if atWord == nil || atDeterminer == nil {
		text := ""
		if match.Sentence != nil {
			text = match.Sentence.GetText()
		}
		panic("Error analyzing sentence: '" + text + "' with rule")
	}
	if atWord.GetPOSTag() == nil || atDeterminer.GetPOSTag() == nil {
		panic("Error analyzing sentence: missing POS")
	}
	wordPOS := *atWord.GetPOSTag()
	detPOS := *atDeterminer.GetPOSTag()
	isNoun := strings.HasPrefix(wordPOS, "N") || strings.HasPrefix(wordPOS, "Z")
	isAdjective := strings.HasPrefix(wordPOS, "J")

	prefix := f.NounAdjPrefix(isNoun, isAdjective)

	// Without synthesizer cannot invent forms — return match with existing suggestions only
	if f.Synthesize == nil {
		out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
		out.ShortMessage = match.GetShortMessage()
		if len(match.GetSuggestedReplacements()) > 0 {
			out.SetSuggestedReplacements(match.GetSuggestedReplacements())
		}
		return out
	}

	determinerForms := make([][]string, 4)
	wordForms := make([][]string, 4)
	for i := 0; i < 4; i++ {
		determinerForms[i] = f.Synthesize(atDeterminer, determinerPrefix+genderNumber[i])
		wordForms[i] = f.Synthesize(atWord, prefix+genderNumber[i])
		// if it cannot be synthesized, keep the original determiner when POS matches gender
		if len(determinerForms[i]) == 0 {
			gnRE := regexp.MustCompile(".+" + genderNumber[i])
			if wwdFullMatch(gnRE, detPOS) {
				determinerForms[i] = []string{atDeterminer.GetToken()}
			}
		}
		if len(wordForms[i]) == 0 {
			gnRE := regexp.MustCompile(".+" + genderNumber[i])
			if wwdFullMatch(gnRE, wordPOS) {
				wordForms[i] = []string{atWord.GetToken()}
			}
		}
	}

	var replacements []string
	for i := 0; i < 4; i++ {
		for _, word := range wordForms[i] {
			for _, detForm := range determinerForms[i] {
				if _, skip := ExceptionsDeterminer[detForm]; skip {
					continue
				}
				if detForm == "" || word == "" {
					continue
				}
				determiner := detForm
				w := word
				if isDeterminerCapitalized {
					determiner = tools.UppercaseFirstChar(determiner)
				}
				if isWordCapitalized {
					w = tools.UppercaseFirstChar(w)
				}
				if isDeterminerAllupper {
					determiner = strings.ToUpper(determiner)
				}
				if isWordAllupper {
					w = strings.ToUpper(w)
				}
				r := determiner + " " + w
				r = strings.ReplaceAll(r, "' ", "'")
				ok := true
				if f.ValidateSuggestion != nil {
					ok = f.ValidateSuggestion(r)
				}
				if ok && !containsStrWW(replacements, r) {
					if strings.HasSuffix(r, atWord.GetToken()) {
						// add at front
						replacements = append([]string{r}, replacements...)
					} else {
						replacements = append(replacements, r)
					}
				}
			}
		}
	}

	// add existing suggestion in the XML rule at front
	existing := match.GetSuggestedReplacements()
	if len(existing) > 0 {
		replacements = append(append([]string{}, existing...), replacements...)
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	if len(replacements) > 0 {
		out.SetSuggestedReplacements(replacements)
	}
	return out
}

func containsStrWW(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
