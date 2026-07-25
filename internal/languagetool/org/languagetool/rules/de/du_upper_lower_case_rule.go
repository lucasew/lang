package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DuUpperLowerCaseRule ports org.languagetool.rules.de.DuUpperLowerCaseRule.
// Coherent use of Du/du, Dich/dich etc.: first use sets casing; later mismatches flag.
// Java: TextLevelRule, Categories.CASING, setUrl, minToCheckParagraph=-1.
type DuUpperLowerCaseRule struct {
	Messages map[string]string
	Category *rules.Category
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// Java: private static final Set<String> lowerWords = new HashSet<>(Arrays.asList(...))
var duLowerWords = map[string]struct{}{
	"du": {}, "dir": {}, "dich": {}, "dein": {}, "deine": {}, "deines": {}, "deins": {},
	"deiner": {}, "deinen": {}, "deinem": {},
	"euch": {}, "euer": {}, "eure": {}, "euere": {}, "euren": {}, "eueren": {}, "euern": {},
	"eurer": {}, "euerer": {}, "eurem": {}, "euerem": {}, "eures": {}, "eueres": {},
}

// Java StringUtils.equalsAny skip tokens after which the next token is ignored.
// The three "\\uXXXX" source literals compile to the six-char strings \uXXXX (not U+XXXX).
var duSkipPrev = map[string]struct{}{
	"\"": {}, "„": {}, "‚": {}, ":": {}, "»": {}, "«": {}, "“": {}, "-": {}, "–": {},
	"*": {}, "•": {},
	`\u2063`: {}, `\u25E6`: {}, `\u00B7`: {},
}

func NewDuUpperLowerCaseRule(messages map[string]string) *DuUpperLowerCaseRule {
	r := &DuUpperLowerCaseRule{
		Messages: messages,
		Category: rules.CatCasing.GetCategory(messages),
	}
	// Java: addExamplePair(Example.wrong(...), Example.fixed(...))
	r.AddExamplePair(
		rules.Wrong("Wie geht es Dir? Bist <marker>du</marker> wieder gesund?"),
		rules.Fixed("Wie geht es Dir? Bist <marker>Du</marker> wieder gesund?"),
	)
	// Java: setUrl(Tools.getUrl("https://languagetool.org/insights/de/beitrag/duzen-grossgeschrieben/"))
	return r
}

func (r *DuUpperLowerCaseRule) GetID() string { return "DE_DU_UPPER_LOWER" }

func (r *DuUpperLowerCaseRule) GetDescription() string {
	return "Einheitliche Verwendung von Du/du, Dir/dir etc."
}

// GetURL ports DuUpperLowerCaseRule constructor setUrl.
func (r *DuUpperLowerCaseRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/duzen-grossgeschrieben/"
}

func (r *DuUpperLowerCaseRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// AddExamplePair ports Rule.addExamplePair.
func (r *DuUpperLowerCaseRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *DuUpperLowerCaseRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *DuUpperLowerCaseRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// MinToCheckParagraph ports TextLevelRule.minToCheckParagraph (Java returns -1).
func (r *DuUpperLowerCaseRule) MinToCheckParagraph() int { return -1 }

// EstimateContextForSureMatch ports TextLevelRule.estimateContextForSureMatch → -1.
func (r *DuUpperLowerCaseRule) EstimateContextForSureMatch() int { return -1 }

// MatchList ports DuUpperLowerCaseRule.match(List<AnalyzedSentence>).
func (r *DuUpperLowerCaseRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var firstUse string // Java: String firstUse = null; empty means unset
	firstUseSet := false
	var ruleMatches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		for i := 0; i < len(tokens); i++ {
			// Java: if (i > 0 && (tokens[i-1].isSentenceStart() || StringUtils.equalsAny(...))) continue;
			if i > 0 {
				prev := tokens[i-1]
				if prev != nil && prev.IsSentenceStart() {
					continue
				}
				if prev != nil {
					if _, ok := duSkipPrev[prev.GetToken()]; ok {
						continue
					}
				}
			}
			if tokens[i] == nil {
				continue
			}
			word := tokens[i].GetToken()
			// Java: word.toLowerCase() (default locale; lowerWords are plain Latin)
			lcWord := strings.ToLower(word)
			if _, ok := duLowerWords[lcWord]; !ok {
				continue
			}
			if !firstUseSet {
				firstUse = word
				firstUseSet = true
				continue
			}
			firstUseIsUpper := tools.StartsWithUppercase(firstUse)
			var msg, replacement string
			if firstUseIsUpper && !tools.StartsWithUppercase(word) {
				replacement = tools.UppercaseFirstChar(word)
				msg = "Vorher wurde bereits '" + firstUse + "' großgeschrieben. " +
					"Aus Gründen der Einheitlichkeit '" + replacement + "' hier auch großschreiben?"
			} else if !firstUseIsUpper && tools.StartsWithUppercase(word) && !stringUtilsIsAllUpperCase(word) {
				// Java: StringUtils.isAllUpperCase(word) — Apache Commons, not StringTools
				replacement = tools.LowercaseFirstChar(word)
				msg = "Vorher wurde bereits '" + firstUse + "' kleingeschrieben. " +
					"Aus Gründen der Einheitlichkeit '" + replacement + "' hier auch kleinschreiben?"
			}
			if msg != "" {
				rm := rules.NewRuleMatch(r, sentence, pos+tokens[i].GetStartPos(), pos+tokens[i].GetEndPos(), msg)
				rm.SetSuggestedReplacement(replacement)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}

// stringUtilsIsAllUpperCase ports Apache Commons Lang3 StringUtils.isAllUpperCase:
// empty → false; every UTF-16 code unit must be Character.isUpperCase.
// Java DuUpperLowerCaseRule uses this (not StringTools.isAllUppercase).
func stringUtilsIsAllUpperCase(cs string) bool {
	if cs == "" {
		return false
	}
	for _, r := range cs {
		// Java charAt loop; BMP-only pronouns here. Character.isUpperCase(code unit).
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
