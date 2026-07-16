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
}

func NewAvsAnRule(messages map[string]string) *AvsAnRule {
	return &AvsAnRule{Messages: messages}
}

func (r *AvsAnRule) GetID() string { return "EN_A_VS_AN" }

var (
	cleanupPattern     = regexp.MustCompile(`[^αa-zA-Z0-9.;,:']`)
	// Include curly single quotes ’ (U+2019) used as openers/closers in tests.
	delimPattern = regexp.MustCompile(`^[-"“”'‘’()\[\]]+$`)
	dashQuotePattern   = regexp.MustCompile(`[-']`)
	anPrefixes         = regexp.MustCompile(`(?i)^(unidentif|uni[mn])[a-z]+$`)
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
		if (equalsA || equalsAn) && isArticleContext(token.GetToken()) &&
			!isApostropheContraction(tokens, i) {
			// Without a tagger, skip letter/variable uses like "a and b" where Java
			// does not tag "a" as DT.
			determiner := GetCorrectDeterminerFor(token)
			var msg string
			if equalsA && determiner == DeterminerAN {
				replacement := "an"
				if tools.StartsWithUppercase(prevTokenStr) {
					replacement = "An"
				}
				msg = "Use <suggestion>" + replacement + "</suggestion> instead of '" + prevTokenStr + "' if the following " +
					"word starts with a vowel sound, e.g. 'an article', 'an hour'."
			} else if equalsAn && determiner == DeterminerA {
				replacement := "a"
				if tools.StartsWithUppercase(prevTokenStr) {
					replacement = "A"
				}
				msg = "Use <suggestion>" + replacement + "</suggestion> instead of '" + prevTokenStr + "' if the following " +
					"word doesn't start with a vowel sound, e.g. 'a sentence', 'a university'."
			}
			if msg != "" {
				prev := tokens[prevTokenIndex]
				rm := rules.NewRuleMatch(r, sentence, prev.GetStartPos(), prev.GetEndPos(), msg)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		nextToken := ""
		if i+1 < len(tokens) {
			nextToken = tokens[i+1].GetToken()
		}
		// Without English tagger: treat surface a/an as DT (Java uses hasPosTag("DT")).
		if isDeterminerToken(tokens, i) {
			prevTokenIndex = i
		} else if len(nextToken) > 1 && delimPattern.MatchString(token.GetToken()) {
			// skip quotes etc.
		} else {
			prevTokenIndex = 0
		}
	}
	return ruleMatches
}

func isDeterminerToken(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	token := tokens[i]
	if token.HasPosTag("DT") {
		return true
	}
	t := token.GetToken()
	// Other DTs by surface
	if strings.EqualFold(t, "the") || strings.EqualFold(t, "this") ||
		strings.EqualFold(t, "that") || strings.EqualFold(t, "these") ||
		strings.EqualFold(t, "those") {
		return true
	}
	if !strings.EqualFold(t, "a") && !strings.EqualFold(t, "an") {
		return false
	}
	// a / an surface fallback without tagger.
	// Reject mid-word splits: Qur'an → an, Abbie an' (an then apostrophe).
	if i+1 < len(tokens) {
		n := tokens[i+1].GetToken()
		if (n == "'" || n == "’") && !tokens[i+1].IsWhitespaceBefore() {
			return false // an' contraction
		}
	}
	if token.IsWhitespaceBefore() || token.GetStartPos() == 0 {
		return true
	}
	// No whitespace before: OK after opening quotes/brackets ("a industry"),
	// not after letter+apostrophe (Qur' + an).
	if i > 0 {
		prev := tokens[i-1].GetToken()
		if prev == "'" || prev == "’" || prev == "‘" {
			// letter before apostrophe? e.g. Qur '
			if i > 1 {
				before := tokens[i-2].GetToken()
				if before != "" && unicode.IsLetter([]rune(before)[0]) {
					return false
				}
			}
			return false
		}
	}
	return true
}

// isArticleContext is false for closed-class words that cannot follow an indefinite
// article in the a/an sense (e.g. "a and b" as variables).
func isArticleContext(word string) bool {
	w := strings.ToLower(cleanupPattern.ReplaceAllString(word, ""))
	if w == "" {
		return true // let GetCorrectDeterminerFor decide UNKNOWN
	}
	switch w {
	case "and", "or", "but", "nor", "as", "vs", "equals", "equal", "plus", "minus":
		return false
	}
	return true
}

// isApostropheContraction detects A'ight etc. where a/an is glued via apostrophe.
// Opening curly quotes (‘…’) are NOT contractions.
func isApostropheContraction(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 1 || tokens[i].IsWhitespaceBefore() {
		return false
	}
	prev := tokens[i-1].GetToken()
	if prev != "'" && prev != "’" {
		return false
	}
	// true contraction only if something letter-like sits before the apostrophe
	if i >= 2 {
		before := tokens[i-2].GetToken()
		if before != "" {
			r := []rune(before)[0]
			if unicode.IsLetter(r) {
				return true // A ' ight
			}
		}
	}
	return false
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
