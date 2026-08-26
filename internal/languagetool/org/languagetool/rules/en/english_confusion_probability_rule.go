package en

import (
	"os"
	"regexp"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

// EnglishConfusionProbabilityRule ports
// org.languagetool.rules.en.EnglishConfusionProbabilityRule.
// Java does not override getId() → inherits ConfusionProbabilityRule.RULE_ID ("CONFUSION_RULE").
type EnglishConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

// EnglishConfusionFilenames ports ConfusionProbabilityRule.getFilenames() default
// used by EnglishConfusionProbabilityRule (no override → confusion_sets.txt).
func EnglishConfusionFilenames() []string {
	return []string{"confusion_sets.txt"}
}

// Java EnglishConfusionProbabilityRule.CONTRACTION
// Note: Java String indices are UTF-16; Go GoogleToken uses UTF-8 byte offsets
// (same as GetGoogleTokens). For ASCII + typographic apostrophe windows this still
// matches the intended "'t …" / "’t …" contraction skip.
var enConfusionContraction = regexp.MustCompile(`['’` + "`" + `´‘]t .*`)

var (
	enConfusionPairsOnce sync.Once
	enConfusionPairs     map[string][]*rules.ConfusionPair
)

// loadENConfusionPairs ports ConfusionProbabilityRule constructor loop:
// for filename in getFilenames(): load "/" + language + "/" + filename via ConfusionSetLoader.
// Fail-closed: missing file or parse error → empty map (no invent pairs).
func loadENConfusionPairs() map[string][]*rules.ConfusionPair {
	enConfusionPairsOnce.Do(func() {
		merged := map[string][]*rules.ConfusionPair{}
		for _, fn := range EnglishConfusionFilenames() {
			part := loadCachedENConfusionPairs(fn)
			for k, v := range part {
				merged[k] = append(merged[k], v...)
			}
		}
		enConfusionPairs = merged
	})
	return enConfusionPairs
}

func loadCachedENConfusionPairs(filename string) map[string][]*rules.ConfusionPair {
	// Reuse L2 discover path: official /en/<filename> under resource dir.
	path := discoverENL2ConfusionSet(filename)
	if path == "" {
		return map[string][]*rules.ConfusionPair{}
	}
	f, err := os.Open(path)
	if err != nil {
		return map[string][]*rules.ConfusionPair{}
	}
	defer f.Close()
	// Java: new ConfusionSetLoader(language) — ShortDescriptionProvider optional; nil = no invent defs.
	m, err := rules.NewConfusionSetLoader(nil).LoadConfusionPairs(f)
	if err != nil || m == nil {
		return map[string][]*rules.ConfusionPair{}
	}
	return m
}

func NewEnglishConfusionProbabilityRule(lm ngrams.LanguageModel) *EnglishConfusionProbabilityRule {
	return NewEnglishConfusionProbabilityRuleGrams(lm, 3)
}

// NewEnglishConfusionProbabilityRuleGrams ports the grams constructor:
// super(messages, languageModel, language, grams, EXCEPTIONS, ANTI_PATTERNS)
// + example pair breaks/brakes.
func NewEnglishConfusionProbabilityRuleGrams(lm ngrams.LanguageModel, grams int) *EnglishConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, grams)
	// Java EnglishConfusionProbabilityRule does not override getId() → CONFUSION_RULE (not DE_CONFUSION_RULE-style).
	// Java: LanguageModelUtils.getGoogleStyleWordTokenizer(language) → EnglishWordTokenizer
	wt := entok.NewEnglishWordTokenizer()
	base.Tokenize = wt.Tokenize
	// Java: super(..., EXCEPTIONS, ANTI_PATTERNS)
	base.Exceptions = append([]string(nil), EnglishConfusionExceptions...)
	base.IsException = enConfusionIsException
	base.IsCoveredByAntiPattern = enConfusionIsCoveredByAntiPattern
	// Java CPR ctor: load getFilenames() → /en/confusion_sets.txt
	base.SetWordToPairs(loadENConfusionPairs())
	// Java: breaks → brakes
	base.AddExamplePair(
		rules.Wrong("Don't forget to put on the <marker>breaks</marker>."),
		rules.Fixed("Don't forget to put on the <marker>brakes</marker>."),
	)
	return &EnglishConfusionProbabilityRule{ConfusionProbabilityRule: base}
}

// enConfusionIsException ports EnglishConfusionProbabilityRule.isException (CONTRACTION).
// Java: if (startPos > 3) covered = sentence.substring(startPos-3, endPos)
// startPos/endPos are Java String / GoogleToken indices (UTF-16 code units).
func enConfusionIsException(sentence string, startPos, endPos int) bool {
	if startPos <= 3 || endPos < startPos {
		return false
	}
	from := startPos - 3
	covered := enJavaSubstring(sentence, from, endPos)
	return enConfusionContraction.MatchString(covered)
}

// enJavaSubstring ports Java String.substring(begin, end) over UTF-16 code units
// (GoogleToken startPos/endPos use the same indexing).
func enJavaSubstring(s string, begin, end int) string {
	u := utf16.Encode([]rune(s))
	n := len(u)
	if begin < 0 {
		begin = 0
	}
	if end > n {
		end = n
	}
	if begin >= end || begin >= n {
		return ""
	}
	return string(utf16.Decode(u[begin:end]))
}

// Match ports ConfusionProbabilityRule.match through the English wrapper.
func (r *EnglishConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.Match(sentence)
}

// GetID ports inherited ConfusionProbabilityRule.getId() → CONFUSION_RULE.
func (r *EnglishConfusionProbabilityRule) GetID() string {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return ngrams.ConfusionRuleID
	}
	return r.ConfusionProbabilityRule.GetID()
}

// GetFilenames ports ConfusionProbabilityRule.getFilenames for English (default list).
func (r *EnglishConfusionProbabilityRule) GetFilenames() []string {
	return EnglishConfusionFilenames()
}
