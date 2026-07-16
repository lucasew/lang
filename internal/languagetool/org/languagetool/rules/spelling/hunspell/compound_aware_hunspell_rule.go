package hunspell

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const maxCompoundSuggestions = 20

// CompoundAwareHunspellRule ports
// org.languagetool.rules.spelling.hunspell.CompoundAwareHunspellRule —
// combines Hunspell with compound splitting + Morfologik multi-speller suggestions.
type CompoundAwareHunspellRule struct {
	*HunspellRule
	CompoundSplitter tokenizers.CompoundWordTokenizer
	MorfoSpeller     *morfologik.MorfologikMultiSpeller
	// FilterForLanguage optional language-specific suggestion filter.
	FilterForLanguage func(suggestions []string) []string
	// MaxSuggestions caps returned suggestions (default 20).
	MaxSuggestions int
}

func NewCompoundAwareHunspellRule(
	languageCode string,
	dict HunspellDictionary,
	splitter tokenizers.CompoundWordTokenizer,
	morfo *morfologik.MorfologikMultiSpeller,
) *CompoundAwareHunspellRule {
	if splitter == nil {
		splitter = tokenizers.NewSimpleCompoundWordTokenizer()
	}
	r := &CompoundAwareHunspellRule{
		HunspellRule:     NewHunspellRule(languageCode, dict),
		CompoundSplitter: splitter,
		MorfoSpeller:     morfo,
		MaxSuggestions:   maxCompoundSuggestions,
	}
	return r
}

// SpellingFilePaths ports getSpellingFilePaths.
func SpellingFilePaths(langCode string) []string {
	return []string{
		"/" + langCode + "/hunspell/spelling.txt",
		"/" + langCode + "/hunspell/spelling_custom.txt",
		"/" + langCode + "/multitoken-suggest.txt",
		"spelling_global.txt",
	}
}

// Suggest overrides HunspellRule.Suggest with compound-aware candidates.
func (r *CompoundAwareHunspellRule) Suggest(word string) []string {
	if r == nil {
		return nil
	}
	// start with dictionary suggestions
	var suggestions []string
	if r.Dict != nil {
		suggestions = append(suggestions, r.Dict.Suggest(word)...)
	}
	// compound parts via splitter
	var simple []string
	if r.CompoundSplitter != nil {
		parts := r.CompoundSplitter.Tokenize(word)
		if len(parts) > 1 {
			// join corrected parts if each is known
			ok := true
			for _, p := range parts {
				if r.Dict != nil && !r.Dict.Spell(p) {
					ok = false
					break
				}
			}
			if ok {
				simple = append(simple, strings.Join(parts, ""))
			}
			// also try morfo suggestions per part
			if r.MorfoSpeller != nil {
				for i, p := range parts {
					if r.Dict != nil && r.Dict.Spell(p) {
						continue
					}
					for _, s := range r.MorfoSpeller.GetSuggestions(p) {
						cp := append([]string{}, parts...)
						cp[i] = s
						simple = append(simple, strings.Join(cp, ""))
					}
				}
			}
		}
	}
	// morfologik whole-word
	var noSplit []string
	if r.MorfoSpeller != nil {
		noSplit = r.MorfoSpeller.GetSuggestions(word)
		if tools.StartsWithUppercase(word) && !tools.IsAllUppercase(word) {
			for _, s := range r.MorfoSpeller.GetSuggestions(strings.ToLower(word)) {
				noSplit = append(noSplit, tools.UppercaseFirstChar(s))
			}
		}
		// trailing punctuation
		for _, punct := range []string{".", "..."} {
			if strings.HasSuffix(word, punct) {
				base := strings.TrimSuffix(word, punct)
				for _, s := range r.MorfoSpeller.GetSuggestions(base) {
					noSplit = append(noSplit, s+punct)
				}
			}
		}
	}
	// interleave
	mixed := interleave(noSplit, simple)
	mixed = append(suggestions, mixed...)
	mixed = filterDupes(mixed)
	if r.FilterForLanguage != nil {
		mixed = r.FilterForLanguage(mixed)
	}
	max := r.MaxSuggestions
	if max <= 0 {
		max = maxCompoundSuggestions
	}
	if len(mixed) > max {
		mixed = mixed[:max]
	}
	return mixed
}

func interleave(lists ...[]string) []string {
	max := 0
	for _, l := range lists {
		if len(l) > max {
			max = len(l)
		}
	}
	var out []string
	for i := 0; i < max; i++ {
		for _, l := range lists {
			if i < len(l) {
				out = append(out, l[i])
			}
		}
	}
	return out
}

func filterDupes(in []string) []string {
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
