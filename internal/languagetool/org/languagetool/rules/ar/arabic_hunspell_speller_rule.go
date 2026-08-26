package ar

import (
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Java ArabicHunspellSpellerRule.RULE_ID / RESOURCE_FILENAME.
const (
	ArabicHunspellRuleID   = "HUNSPELL_RULE_AR"
	ArabicHunspellDictPath = "/ar/hunspell/ar.dic"
)

// arabicTashkeelWordChars ports tokenizeText pattern: letters + TASHKEEL_CHARS are word chars
// (Java Pattern "[^\\p{L}" + TASHKEEL_CHARS + "]").
func arabicTashkeelWordChars() map[rune]struct{} {
	m := make(map[rune]struct{}, len([]rune(tools.TashkeelChars)))
	for _, r := range tools.TashkeelChars {
		m[r] = struct{}{}
	}
	return m
}

// ArabicHunspellSpellerRule ports org.languagetool.rules.ar.ArabicHunspellSpellerRule.
type ArabicHunspellSpellerRule struct {
	*hunspell.HunspellRule
}

// NewArabicHunspellSpellerRule constructs the AR hunspell speller with the given dictionary.
// Production registration loads official /ar/hunspell/ar.dic via TryOpenFromClasspath.
// Full Java match parity needs DumontsHunspellDictionary (native hunspell.spell with .aff);
// pure-Go FileHunspellDictionary is stems-only (no affix expansion).
func NewArabicHunspellSpellerRule(dict hunspell.HunspellDictionary) *ArabicHunspellSpellerRule {
	base := hunspell.NewHunspellRule("ar", dict)
	base.ID = ArabicHunspellRuleID
	base.Description = "Possible spelling mistake"
	// Java isLatinScript() = false → NonLatinScript.
	if base.SpellingCheckRule != nil {
		base.NonLatinScript = true
		// Java ignoreWord: strip tashkeel then super.ignoreWord.
		scr := base.SpellingCheckRule
		scr.IgnoreWordFn = func(word string) bool {
			return scr.IgnoreWordDefault(tools.RemoveTashkeel(word))
		}
	}
	// Java tokenizeText: non-word = not letter and not tashkeel.
	base.NonWordSplitter = hunspell.NonWordSplitter{WordChars: arabicTashkeelWordChars()}
	r := &ArabicHunspellSpellerRule{HunspellRule: base}
	// Java isMisspelled: strip tashkeel then super.isMisspelled.
	base.IsMisspelled = r.IsMisspelledStripped
	return r
}

// GetID ports ArabicHunspellSpellerRule.getId → HUNSPELL_RULE_AR.
func (r *ArabicHunspellSpellerRule) GetID() string { return ArabicHunspellRuleID }

// GetDictFilenameInResources ports getDictFilenameInResources → /ar/hunspell/ar.dic.
func (r *ArabicHunspellSpellerRule) GetDictFilenameInResources(_ string) string {
	return ArabicHunspellDictPath
}

// IsLatinScript ports isLatinScript → false.
func (r *ArabicHunspellSpellerRule) IsLatinScript() bool { return false }

// TokenizeText ports ArabicHunspellSpellerRule.tokenizeText.
func (r *ArabicHunspellSpellerRule) TokenizeText(sentence string) []string {
	return TokenizeArabicSpellText(sentence)
}

// TokenizeArabicSpellText ports tokenizeText:
// Pattern.compile("[^\\p{L}" + TASHKEEL_CHARS + "]").split(sentence).
func TokenizeArabicSpellText(sentence string) []string {
	return hunspell.NonWordSplitter{WordChars: arabicTashkeelWordChars()}.Split(sentence)
}

// IsMisspelledStripped ports isMisspelled: removeTashkeel then super.isMisspelled.
func (r *ArabicHunspellSpellerRule) IsMisspelledStripped(word string) bool {
	if r == nil || r.HunspellRule == nil {
		return false
	}
	return r.HunspellRule.IsMisspelledWord(tools.RemoveTashkeel(word))
}

// IgnoreWord ports ignoreWord: removeTashkeel then super.ignoreWord.
func (r *ArabicHunspellSpellerRule) IgnoreWord(word string) bool {
	if r == nil || r.HunspellRule == nil || r.SpellingCheckRule == nil {
		return false
	}
	return r.SpellingCheckRule.IgnoreWordDefault(tools.RemoveTashkeel(word))
}

// Match ports Rule.match via HunspellRule with AR isMisspelled hook.
func (r *ArabicHunspellSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil {
		return nil, nil
	}
	r.HunspellRule.IsMisspelled = r.IsMisspelledStripped
	return r.HunspellRule.Match(sentence)
}

// HasArabicLetter reports whether s contains any Unicode letter (token gate helper).
func HasArabicLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
