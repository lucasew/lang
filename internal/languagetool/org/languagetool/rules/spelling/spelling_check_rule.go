package spelling

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// Constants from org.languagetool.rules.spelling.SpellingCheckRule.
const (
	HighConfidence     = float32(0.99)
	LanguageTool       = "LanguageTool"
	LanguageTooler     = "LanguageTooler"
	MaxTokenLength     = 200
	SpellingIgnoreFile = "/hunspell/ignore.txt"
	SpellingFile       = "/hunspell/spelling.txt"
	CustomSpellingFile = "/hunspell/spelling_custom.txt"
	GlobalSpellingFile = "spelling_global.txt"
)

// SpellingCheckRule is a surface for spellcheck rules (full match deferred).
type SpellingCheckRule struct {
	ID           string
	Description  string
	LanguageCode string
	// IsMisspelled returns true if word is not in the dictionary.
	IsMisspelled func(word string) bool
	// IgnoreWords is a set of words to accept.
	IgnoreWords map[string]struct{}
	// ProhibitedWords ports wordsToBeProhibited (prohibit.txt): always flag as misspell
	// even when the dictionary would accept them (Java isProhibited).
	ProhibitedWords map[string]struct{}
	// MultiWordIgnore ports multi-token addIgnoreWords lines that become
	// DisambiguationPatternRule(IGNORE_SPELLING) antipatterns in Java.
	// Each entry is a case-sensitive token sequence from TokenizeIgnoreLine
	// (Java language.getWordTokenizer().tokenize, empty/whitespace tokens dropped).
	MultiWordIgnore [][]string
	// TokenizeIgnoreLine optional override for multi-word ignore splitting.
	// When nil, DefaultTokenizeIgnoreLine(LanguageCode, line) is used.
	TokenizeIgnoreLine func(line string) []string
	// AntiPatterns ports SpellingCheckRule.antiPatterns (getAntiPatterns).
	// Built as IGNORE_SPELLING DisambiguationPatternRule for multi-token ignore
	// and acceptPhrases (Java INTERNAL_ANTIPATTERN).
	AntiPatterns []*disambigrules.DisambiguationPatternRule
	// IgnoreWordsWithLength ports ignoreWordsWithLength: accept any word whose
	// length is ≤ this value when > 0 (Java EN/DE/GA spellers set 1).
	IgnoreWordsWithLength int
	// ConvertsCase ports convertsCase (set by MorfologikSpellerRule.setConvertsCase).
	// When true, isIgnoredNoCase also checks lowercased form if not mixed case.
	ConvertsCase bool
	// DisableConsiderIgnoreWords ports considerIgnoreWords=false (default is consider).
	DisableConsiderIgnoreWords bool
	// NonLatinScript ports isLatinScript()==false (Java default isLatinScript true).
	// When true, ignoreWord uses \p{L} (any letter); when false (Latin script langs),
	// ignoreWord ignores tokens with no Latin-script letters (Java pHasNoLetterLatin).
	NonLatinScript bool
	// DisableTokenizeNewWords ports tokenizeNewWords()==false (Java default true).
	// When true, multi-token spelling lines are stored as a single IgnoreWords key
	// (whole line), not MultiWordIgnore / IGNORE_SPELLING antipatterns.
	// Java CA/ES/EN: multi-token phrases belong in multiwords.txt for chunking.
	DisableTokenizeNewWords bool
	// TagPOS optional surface → POS tags for isProperNoun (NNP) in filterSuggestions.
	// Fail-closed when nil: " s" → "'s" rewrite skipped.
	TagPOS func(word string) []string
	// IsProperNounFn optional override for isProperNoun (tests).
	IsProperNounFn func(word string) bool
	// FilterNoSuggestWordsFn ports filterNoSuggestWords language overrides (EN/RU NOSUGGEST).
	FilterNoSuggestWordsFn func(suggestions []string) []string
	// FilterSuggestionsExtraFn ports language filterSuggestions after super (e.g. EN CONTAINS_TOKEN).
	FilterSuggestionsExtraFn func(suggestions []string) []string
	// IgnoreTokenFn optional language override for ignoreToken(tokens, idx).
	IgnoreTokenFn func(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool
	// IgnorePotentiallyMisspelledWordFn ports ignorePotentiallyMisspelledWord.
	// Called only after the dictionary already considers the word incorrect
	// (Java SpellingCheckRule default returns false; NL/DE override).
	IgnorePotentiallyMisspelledWordFn func(word string) bool
	// ExpandLineFn ports SpellingCheckRule.expandLine (default: singleton of the line).
	// Used when loading prohibit.txt / prohibit_custom.txt (Java init).
	// DE GermanSpellerRule overrides to LineExpander (/S /N /A /E /F, .*prefix, …).
	ExpandLineFn func(line string) []string
	// Cached sorted ignore lists for startsWithIgnoredWord (invalidated on AddIgnoreWords).
	ignoreDictSorted     []string
	ignoreDictSortedFold []string
}

// ExpandLine ports SpellingCheckRule.expandLine — default returns the line as-is.
func (r *SpellingCheckRule) ExpandLine(line string) []string {
	if r != nil && r.ExpandLineFn != nil {
		return r.ExpandLineFn(line)
	}
	return []string{line}
}

func NewSpellingCheckRule(id, description, languageCode string) *SpellingCheckRule {
	return &SpellingCheckRule{
		ID:              id,
		Description:     description,
		LanguageCode:    languageCode,
		IgnoreWords:     map[string]struct{}{},
		ProhibitedWords: map[string]struct{}{},
	}
}

// ApplyUserAcceptedWords ports SpellingCheckRule ctor:
// wordsToBeIgnored.addAll(userConfig.getAcceptedWords()) for all users (premium or free).
func (r *SpellingCheckRule) ApplyUserAcceptedWords(accepted []string) {
	if r == nil || len(accepted) == 0 {
		return
	}
	r.AddIgnoreWords(accepted...)
}

func (r *SpellingCheckRule) GetID() string          { return r.ID }
func (r *SpellingCheckRule) GetDescription() string { return r.Description }

// AcceptWord reports whether word should not be flagged (ignore list or not misspelled).
// Java SpellingCheckRule: prohibited wins over dictionary accept; ignoreWord wins over misspell.
func (r *SpellingCheckRule) AcceptWord(word string) bool {
	if r == nil {
		return false
	}
	// Java isProhibited: force misspell even if dict accepts.
	if r.IsProhibited(word) {
		return false
	}
	// Java ignoreWord / isIgnoredNoCase (includes MaxTokenLength, no-letter, ignore set, length).
	if r.IgnoreWord(word) {
		return true
	}
	if r.IsMisspelled == nil {
		return true
	}
	return !r.IsMisspelled(word)
}

// CanBeIgnoredToken ports MorfologikSpellerRule.canBeIgnored fixed checks:
// sentence start, immunized, ignored-by-speller, URL, email.
// ignoreToken (ignoreWord) is applied separately with token index context.
func CanBeIgnoredToken(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil || tok.IsSentenceStart() {
		return true
	}
	if tok.IsImmunized() || tok.IsIgnoredBySpeller() {
		return true
	}
	w := tok.GetToken()
	if IsUrl(w) || IsEMail(w) {
		return true
	}
	return false
}

// IsProhibited ports SpellingCheckRule.isProhibited (exact membership).
func (r *SpellingCheckRule) IsProhibited(word string) bool {
	if r == nil || word == "" || len(r.ProhibitedWords) == 0 {
		return false
	}
	_, ok := r.ProhibitedWords[word]
	return ok
}

// AddIgnoreWords ports addIgnoreWords for one or more lines.
// When tokenizeNewWords() is true (default): single-token → IgnoreWords;
// multi-token → MultiWordIgnore + IGNORE_SPELLING antipattern.
// When DisableTokenizeNewWords (tokenizeNewWords false): entire line → IgnoreWords
// (Java CA/ES/EN; multi-token phrases are not antipatterns).
func (r *SpellingCheckRule) AddIgnoreWords(words ...string) {
	if r == nil {
		return
	}
	if r.IgnoreWords == nil {
		r.IgnoreWords = map[string]struct{}{}
	}
	// Invalidate startsWithIgnoredWord caches.
	r.ignoreDictSorted = nil
	r.ignoreDictSortedFold = nil
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		// Java: if (!tokenizeNewWords()) { wordsToBeIgnored.add(line); }
		if r.DisableTokenizeNewWords {
			r.IgnoreWords[w] = struct{}{}
			continue
		}
		// Java: language.getWordTokenizer().tokenize(line); skip empty tokens.
		parts := r.tokenizeIgnoreLine(w)
		if len(parts) > 1 {
			r.MultiWordIgnore = append(r.MultiWordIgnore, append([]string(nil), parts...))
			// Java: antiPatterns.add(new DisambiguationPatternRule(..., IGNORE_SPELLING)).
			r.appendIgnoreSpellingAntiPattern(parts)
			continue
		}
		if len(parts) == 1 {
			r.IgnoreWords[parts[0]] = struct{}{}
			continue
		}
		// No tokens after filter — keep raw surface as single ignore (fail closed).
		r.IgnoreWords[w] = struct{}{}
	}
}

// tokenizeIgnoreLine uses TokenizeIgnoreLine or DefaultTokenizeIgnoreLine.
func (r *SpellingCheckRule) tokenizeIgnoreLine(line string) []string {
	if r != nil && r.TokenizeIgnoreLine != nil {
		return r.TokenizeIgnoreLine(line)
	}
	lang := ""
	if r != nil {
		lang = r.LanguageCode
	}
	return DefaultTokenizeIgnoreLine(lang, line)
}

// DefaultTokenizeIgnoreLine ports Java language.getWordTokenizer().tokenize for
// SpellingCheckRule.addIgnoreWords multi-token lines. Uses WordTokenizerForLanguage
// (same as JLanguageTool.Analyze) and drops empty/whitespace tokens.
func DefaultTokenizeIgnoreLine(langCode, line string) []string {
	wt := languagetool.WordTokenizerForLanguage(langCode)
	if wt == nil {
		// Should not happen; WordTokenizerForLanguage falls back to WordTokenizer.
		return strings.Fields(line)
	}
	raw := wt.Tokenize(line)
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if strings.TrimSpace(t) != "" {
			out = append(out, t)
		}
	}
	return out
}

// AddProhibitedWords ports addProhibitedWords (prohibit.txt / prohibit_custom.txt).
func (r *SpellingCheckRule) AddProhibitedWords(words ...string) {
	if r.ProhibitedWords == nil {
		r.ProhibitedWords = map[string]struct{}{}
	}
	for _, w := range words {
		if w != "" {
			r.ProhibitedWords[w] = struct{}{}
		}
	}
}

// MarkMultiWordIgnoreSpelling ports multi-token IGNORE_SPELLING antipatterns:
// when consecutive non-whitespace tokens match a MultiWordIgnore phrase
// (case-sensitive, Java PatternToken caseSensitive=true), mark each token
// IgnoreSpelling so Hunspell/Morfologik Match skip them.
func (r *SpellingCheckRule) MarkMultiWordIgnoreSpelling(sentence *languagetool.AnalyzedSentence) {
	if r == nil || sentence == nil || len(r.MultiWordIgnore) == 0 {
		return
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	if len(tokens) == 0 {
		return
	}
	for _, phrase := range r.MultiWordIgnore {
		n := len(phrase)
		if n == 0 || n > len(tokens) {
			continue
		}
		for i := 0; i+n <= len(tokens); i++ {
			ok := true
			for j := 0; j < n; j++ {
				tok := tokens[i+j]
				if tok == nil || tok.GetToken() != phrase[j] {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			for j := 0; j < n; j++ {
				if tokens[i+j] != nil {
					tokens[i+j].IgnoreSpelling()
				}
			}
		}
	}
}
