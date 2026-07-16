package hunspell

import (
	"regexp"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
)

// HunspellRuleID ports HunspellRule.RULE_ID.
const HunspellRuleID = "HUNSPELL_RULE"

// FileExtension ports HunspellRule.FILE_EXTENSION.
const FileExtension = ".dic"

var nonAlphabeticRE = regexp.MustCompile(`^[^\p{L}]+$`)

// HunspellRule ports org.languagetool.rules.spelling.hunspell.HunspellRule
// with a pluggable HunspellDictionary (native hunspell deferred).
type HunspellRule struct {
	*spelling.SpellingCheckRule
	Dict HunspellDictionary
	// IgnoreTaggedWords skips tokens that already have a real POS tag.
	IgnoreTaggedWords bool
}

func NewHunspellRule(languageCode string, dict HunspellDictionary) *HunspellRule {
	r := &HunspellRule{
		SpellingCheckRule: spelling.NewSpellingCheckRule(HunspellRuleID, "Possible spelling mistake", languageCode),
		Dict:              dict,
	}
	r.IsMisspelled = r.IsMisspelledWord
	return r
}

// IsMisspelledWord ports HunspellRule.isMisspelled.
func (r *HunspellRule) IsMisspelledWord(word string) bool {
	if r == nil || r.Dict == nil {
		return false
	}
	if nonAlphabeticRE.MatchString(word) {
		return false
	}
	return !r.Dict.Spell(word)
}

// Suggest ports dictionary suggestions (empty if dict has none).
func (r *HunspellRule) Suggest(word string) []string {
	if r == nil || r.Dict == nil {
		return nil
	}
	return r.Dict.Suggest(word)
}

// Match flags misspelled tokens in the analyzed sentence.
func (r *HunspellRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	var out []*rules.RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() || tok.IsSentenceEnd() {
			continue
		}
		w := tok.GetToken()
		if w == "" || !hasLetter(w) {
			continue
		}
		if r.IgnoreTaggedWords && tok.IsTagged() {
			continue
		}
		// AcceptWord is true when the word should not be flagged.
		if r.AcceptWord(w) {
			continue
		}
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
			"Possible spelling mistake found")
		if sug := r.Suggest(w); len(sug) > 0 {
			m.SetSuggestedReplacements(sug)
		}
		out = append(out, m)
	}
	return out, nil
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
