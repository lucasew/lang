package hunspell

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
	// Java SpellingCheckRule.init: ignore/spelling/prohibit word lists for language.
	ApplyDefaultSpellingWordLists(r.SpellingCheckRule)
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
	// Java HunspellRule.match: getSentenceWithImmunization(sentence).
	work := sentence
	if r.SpellingCheckRule != nil {
		work = r.SpellingCheckRule.SentenceWithImmunization(sentence)
		r.SpellingCheckRule.MarkMultiWordIgnoreSpelling(work)
	}
	tokens := work.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	for idx, tok := range tokens {
		// Java: immunized / ignoredBySpeller / isUrl / isEMail
		if spelling.CanBeIgnoredToken(tok) {
			continue
		}
		// Java ignoreToken → ignoreWord
		if r.SpellingCheckRule != nil && r.IgnoreToken(tokens, idx) {
			continue
		}
		w := tok.GetToken()
		// Sentence-end markers with no letters (e.g. bare ".") are skipped; the last
		// content word may still carry IsSentenceEnd in AnalyzePlain and must be checked.
		if w == "" || !hasLetter(w) {
			continue
		}
		// Java getSentenceTextWithoutImmunizedTokens: stringForSpeller strips emoji etc.
		check := tools.StringForSpeller(w)
		check = strings.TrimSpace(check)
		if check == "" || !hasLetter(check) {
			continue
		}
		if r.IgnoreTaggedWords && tok.IsTagged() {
			if r.SpellingCheckRule == nil || !r.IsProhibited(w) {
				continue
			}
		}
		// AcceptWord is true when the word should not be flagged.
		if r.AcceptWord(check) {
			continue
		}
		// Java HunspellRule.match: after isMisspelled, ignorePotentiallyMisspelledWord.
		if r.SpellingCheckRule != nil && r.IgnorePotentiallyMisspelledWord(check) {
			continue
		}
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
			"Possible spelling mistake found")
		if sug := r.Suggest(check); len(sug) > 0 {
			// Java: getAdditionalTopSuggestions then filterSuggestions
			if top := spelling.AdditionalTopSuggestions(sug, check); len(top) > 0 {
				sug = append(top, sug...)
			}
			if r.SpellingCheckRule != nil {
				sug = r.SpellingCheckRule.FilterSuggestions(sug)
			}
			if len(sug) > 0 {
				m.SetSuggestedReplacements(sug)
			}
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
