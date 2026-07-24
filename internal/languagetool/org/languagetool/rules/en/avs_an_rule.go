package en

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Determiner ports AvsAnRule.Determiner.
type Determiner int

const (
	DeterminerA Determiner = iota
	DeterminerAN
	DeterminerAOrAN
	DeterminerUnknown
)

// AvsAnRule ports org.languagetool.rules.en.AvsAnRule.
type AvsAnRule struct {
	Messages map[string]string
	// URL ports Rule.url (Java setUrl indefinite-articles insights post).
	URL string
	// Category ports Rule.category (Java Categories.MISC).
	Category *rules.Category
	// IssueType ports getLocQualityIssueType (Java Misspelling).
	IssueType rules.ITSIssueType
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewAvsAnRule(messages map[string]string) *AvsAnRule {
	r := &AvsAnRule{
		Messages:  messages,
		URL:       "https://languagetool.org/insights/post/indefinite-articles/",
		Category:  rules.CatMisc.GetCategory(messages),
		IssueType: rules.ITSMisspelling,
	}
	// Java: addExamplePair(Example.wrong(...), Example.fixed(...))
	r.AddExamplePair(
		rules.Wrong("The train arrived <marker>a hour</marker> ago."),
		rules.Fixed("The train arrived <marker>an hour</marker> ago."),
	)
	return r
}

func (r *AvsAnRule) GetID() string { return "EN_A_VS_AN" }

// GetDescription ports AvsAnRule.getDescription.
func (r *AvsAnRule) GetDescription() string { return "Use of 'a' vs. 'an'" }

// GetURL ports Rule.getUrl.
func (r *AvsAnRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *AvsAnRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// GetCategory ports Rule.getCategory.
func (r *AvsAnRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *AvsAnRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil {
		return rules.ITSUncategorized
	}
	return r.IssueType
}

// EstimateContextForSureMatch ports AvsAnRule.estimateContextForSureMatch (Java 1).
func (r *AvsAnRule) EstimateContextForSureMatch() int { return 1 }

// AddExamplePair ports Rule.addExamplePair.
func (r *AvsAnRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	// Reuse BaseRule helper via temporary store on a BaseRule.
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AvsAnRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AvsAnRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

var (
	cleanupPattern = regexp.MustCompile(`[^αa-zA-Z0-9.;,:']`)
	// Java: Pattern.compile("[-\"“'‘()\\[\\]]+") — no ”/’ (U+201D/U+2019).
	delimPattern = regexp.MustCompile(`^[-"“'‘()\[\]]+$`)
	dashQuotePattern    = regexp.MustCompile(`[-']`)
	anPrefixes          = regexp.MustCompile(`(?i)^(unidentif|uni[mn])[a-z]+$`)
	anExceptionPrefixes = regexp.MustCompile(`(?i)^(eu|one|uni|u[rst][aeiou])[a-z]*$`)
)

func (r *AvsAnRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	prevTokenIndex := 0
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		var prevTokenStr string
		if prevTokenIndex > 0 {
			prevTokenStr = tokens[prevTokenIndex].GetToken()
		}
		isSentenceStart := prevTokenIndex == 1
		var equalsA, equalsAn bool
		if !isSentenceStart {
			equalsA = prevTokenStr == "a"
			equalsAn = prevTokenStr == "an"
		} else {
			equalsA = strings.EqualFold(prevTokenStr, "a")
			equalsAn = strings.EqualFold(prevTokenStr, "an")
		}
		if equalsA || equalsAn {
			// Java: only when prev token was DT (hasPosTag("DT"))
			determiner := GetCorrectDeterminerFor(token)
			var msg, replacement string
			if equalsA && determiner == DeterminerAN {
				replacement = "an"
				if tools.StartsWithUppercase(prevTokenStr) {
					replacement = "An"
				}
				msg = "Use <suggestion>" + replacement + "</suggestion> instead of '" + prevTokenStr + "' if the following " +
					"word starts with a vowel sound, e.g. 'an article', 'an hour'."
			} else if equalsAn && determiner == DeterminerA {
				replacement = "a"
				if tools.StartsWithUppercase(prevTokenStr) {
					replacement = "A"
				}
				msg = "Use <suggestion>" + replacement + "</suggestion> instead of '" + prevTokenStr + "' if the following " +
					"word doesn't start with a vowel sound, e.g. 'a sentence', 'a university'."
			}
			if msg != "" {
				prev := tokens[prevTokenIndex]
				rm := rules.NewRuleMatch(r, sentence, prev.GetStartPos(), prev.GetEndPos(), msg)
				rm.ShortMessage = "Wrong article"
				rm.SetSuggestedReplacement(replacement)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		nextToken := ""
		if i+1 < len(tokens) {
			nextToken = tokens[i+1].GetToken()
		}
		// Java: if (token.hasPosTag("DT")) prevTokenIndex = i;
		// fail closed without DT (no surface invent of a/an/the)
		if token.HasPosTag("DT") {
			prevTokenIndex = i
		} else if len(nextToken) > 1 && delimPattern.MatchString(token.GetToken()) {
			// skip e.g. the quote in >>an "industry party"<<
		} else {
			prevTokenIndex = 0
		}
	}
	return ruleMatches
}

// SuggestAorAn ports AvsAnRule.suggestAorAn.
func (r *AvsAnRule) SuggestAorAn(origWord string) string {
	token := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(origWord, nil, nil))
	determiner := GetCorrectDeterminerFor(token)
	switch determiner {
	case DeterminerA, DeterminerAOrAN:
		return "a " + tools.LowercaseFirstCharIfCapitalized(origWord)
	case DeterminerAN:
		return "an " + tools.LowercaseFirstCharIfCapitalized(origWord)
	default:
		return origWord
	}
}

// GetCorrectDeterminerFor ports AvsAnRule.getCorrectDeterminerFor.
func GetCorrectDeterminerFor(token *languagetool.AnalyzedTokenReadings) Determiner {
	if token == nil {
		panic("null token")
	}
	word := token.GetToken()
	determiner := DeterminerUnknown
	parts := dashQuotePattern.Split(word, -1)
	if len(parts) >= 1 && !strings.EqualFold(parts[0], "a") {
		word = parts[0]
	}
	// whitespace before check: if not whitespace before and word is "-", keep?
	// Java: if (token.isWhitespaceBefore() || !"-".equals(word))
	if token.IsWhitespaceBefore() || word != "-" {
		word = cleanupPattern.ReplaceAllString(word, "")
		if tools.IsEmptyStr(word) {
			return DeterminerUnknown
		}
	}
	reqA := getWordsRequiringA()
	reqAn := getWordsRequiringAn()
	if reqA[strings.ToLower(word)] || reqA[word] {
		determiner = DeterminerA
	}
	if reqAn[strings.ToLower(word)] || reqAn[word] {
		if determiner == DeterminerA {
			determiner = DeterminerAOrAN
		} else {
			determiner = DeterminerAN
		}
	}
	if determiner == DeterminerUnknown {
		tokenFirstChar := []rune(word)[0]
		if tools.IsAllUppercase(word) || tools.IsMixedCase(word) {
			determiner = DeterminerUnknown
		} else if anPrefixes.MatchString(token.GetToken()) {
			// Java matches against full token, not cleaned word
			determiner = DeterminerAN
		} else if isVowel(tokenFirstChar) && !anExceptionPrefixes.MatchString(token.GetToken()) {
			determiner = DeterminerAN
		} else {
			determiner = DeterminerA
		}
	}
	return determiner
}

func isVowel(c rune) bool {
	lc := unicode.ToLower(c)
	return lc == 'a' || lc == 'e' || lc == 'i' || lc == 'o' || lc == 'u'
}
