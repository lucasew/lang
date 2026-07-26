package uk

import (
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	// MorfologikUkrainianSpellerRuleID ports getId() → "MORFOLOGIK_RULE_UK_UA".
	MorfologikUkrainianSpellerRuleID = "MORFOLOGIK_RULE_UK_UA"
	// UkrainianSpellerDict ports RESOURCE_FILENAME / getFileName().
	UkrainianSpellerDict = "/uk/hunspell/uk_UA.dict"
	ukAbbreviationChar   = "."
)

var (
	// Java UKRAINIAN_LETTERS — substring (find), not full match.
	ukrainianLetters = regexp.MustCompile(`.*[а-яіїєґА-ЯІЇЄҐ].*`)
	// Java DO_NOT_SUGGEST_SPACED_PATTERN (Matcher.matches → full string).
	ukDoNotSuggestSpaced = regexp.MustCompile(
		`^(авіа|авто|анти|аудіо|відео|водо|гідро|екстра|квазі|кіно|лже|мета|моно|мото|псевдо|пост|радіо|стерео|супер|ультра|фото) .*`)
	// Java PATTERN for single-capital initials (Matcher.matches).
	ukSingleCapital    = regexp.MustCompile(`^[А-ЯІЇЄҐ]$`)
	ukDashPrefixKeyOK  = regexp.MustCompile(`^[а-яіїєґ]{3,}$`)
	ukDashPrefixBadVal = regexp.MustCompile(`^:(bad|alt|slang)$`)
	dashSpellerOnce    sync.Once
	// dashPrefixesSpeller ports dashPrefixes2019 key set after removeIf filters.
	dashPrefixesSpeller map[string]struct{}
)

// loadDashPrefixesSpeller ports static dashPrefixes2019:
// ExtraDictionaryLoader.loadMap("/uk/dash_prefixes.txt") then removeIf
// value.matches(":(bad|alt|slang)") || !key.matches("[а-яіїєґ]{3,}").
func loadDashPrefixesSpeller() map[string]struct{} {
	dashSpellerOnce.Do(func() {
		out := map[string]struct{}{}
		data, err := dashPrefixFS.ReadFile("data/dash_prefixes.txt")
		if err != nil {
			dashPrefixesSpeller = out
			return
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "#") {
				continue
			}
			line = tools.JavaStringTrim(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, " ")
			if len(parts) == 0 || parts[0] == "" {
				continue
			}
			key := parts[0]
			tag := ""
			if len(parts) > 1 {
				tag = parts[1]
			}
			// Java: removeIf value.matches(":(bad|alt|slang)") || !key.matches("[а-яіїєґ]{3,}")
			if tag != "" && ukDashPrefixBadVal.MatchString(tag) {
				continue
			}
			if !ukDashPrefixKeyOK.MatchString(key) {
				continue
			}
			out[key] = struct{}{}
		}
		dashPrefixesSpeller = out
	})
	return dashPrefixesSpeller
}

// MorfologikUkrainianSpellerRule ports rules.uk.MorfologikUkrainianSpellerRule.
type MorfologikUkrainianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

// NewMorfologikUkrainianSpellerRule ports
// MorfologikUkrainianSpellerRule(messages, language, userConfig, altLanguages) with null userConfig.
func NewMorfologikUkrainianSpellerRule() *MorfologikUkrainianSpellerRule {
	return NewMorfologikUkrainianSpellerRuleWithUser(nil)
}

// NewMorfologikUkrainianSpellerRuleWithUser ports the full Java ctor.
func NewMorfologikUkrainianSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikUkrainianSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikUkrainianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikUkrainianSpellerRuleID, "uk", UkrainianSpellerDict, nil),
	}
	// Java isLatinScript() = false
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
		r.ApplyUserAcceptedWords(accepted)
		// Java filterSuggestions override after super.
		r.FilterSuggestionsExtraFn = filterUKSuggestions
	}
	if userConfig != nil {
		r.SetUserConfig(userConfig)
	}
	r.IgnoreTokenFn = r.ukIgnoreToken
	// Java getAdditionalSuggestions — dash prefix compounds.
	r.GetAdditionalSuggestionsFn = r.ukGetAdditionalSuggestions
	// Java MorfologikSpellerRule.initSpeller: binary + spelling.txt (+ user FSA if premium).
	r.InitSpellersFromGetters(nil, userWords)
	// Java isMisspelled override (trailing hyphen) around base isMisspelledWord.
	inner := r.IsMisspelled
	r.IsMisspelled = func(word string) bool {
		return r.ukIsMisspelled(word, inner)
	}
	_ = loadDashPrefixesSpeller()
	return r
}

// GetFileName ports getFileName().
func (r *MorfologikUkrainianSpellerRule) GetFileName() string {
	if r == nil || r.MorfologikSpellerRule == nil {
		return UkrainianSpellerDict
	}
	return r.MorfologikSpellerRule.GetFileName()
}

// IsLatinScript ports isLatinScript() = false.
func (r *MorfologikUkrainianSpellerRule) IsLatinScript() bool { return false }

// GetSpeller1 ports package-private getSpeller1() (lazy init via isMisspelled).
func (r *MorfologikUkrainianSpellerRule) GetSpeller1() *morfologik.MorfologikMultiSpeller {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil
	}
	if r.Speller1 == nil {
		// Java: isMisspelled("1") forces initSpellers when speller1 null.
		_ = r.IsMisspelled != nil && r.IsMisspelled("1")
		if r.Speller1 == nil {
			r.InitSpellersFromGetters(nil, nil)
		}
	}
	return r.Speller1
}

// ukIsMisspelled ports isMisspelled(MorfologikMultiSpeller, word):
// if word.endsWith("-") return !word.startsWith("-"); else super.
func (r *MorfologikUkrainianSpellerRule) ukIsMisspelled(word string, inner func(string) bool) bool {
	if word == "" {
		return false
	}
	if strings.HasSuffix(word, "-") {
		return !strings.HasPrefix(word, "-")
	}
	if inner != nil {
		return inner(word)
	}
	return false
}

// ukIgnoreToken ports ignoreToken:
//  1. no Ukrainian letters → true
//  2. super.ignoreToken → ignoreWord
//  3. next token "." + (ignoreWord(word+".") || single capital)
//  4. hasGoodTag
func (r *MorfologikUkrainianSpellerRule) ukIgnoreToken(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool {
	if idx < 0 || idx >= len(tokens) || tokens[idx] == nil {
		return false
	}
	word := tokens[idx].GetToken()
	// don't check words that don't have Ukrainian letters
	if !ukrainianLetters.MatchString(word) {
		return true
	}
	// super.ignoreToken → ignoreWord(words, idx)
	if r.SpellingCheckRule != nil && r.IgnoreWord(word) {
		return true
	}
	if idx < len(tokens)-1 && tokens[idx+1] != nil && tokens[idx+1].GetToken() == ukAbbreviationChar {
		if r.SpellingCheckRule != nil && r.IgnoreWord(word+ukAbbreviationChar) {
			return true
		}
		// PATTERN.matcher(word).matches() — single Ukrainian capital (initials)
		if ukSingleCapital.MatchString(word) {
			return true
		}
	}
	return hasGoodTagUK(tokens[idx])
}

// hasGoodTagUK ports hasGoodTag: any POS except null / SENT_END / PARA_END.
func hasGoodTagUK(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, rd := range tok.GetReadings() {
		if rd == nil {
			continue
		}
		pos := rd.GetPOSTag()
		if pos == nil || *pos == "" {
			continue
		}
		switch *pos {
		case languagetool.SentenceEndTagName, languagetool.ParagraphEndTagName,
			"SENTENCE_END", "PARAGRAPH_END":
			continue
		}
		return true
	}
	return false
}

// Match ports MorfologikSpellerRule.match + UK getRuleMatches post-arm:
// when super returns empty and !hasGoodTag, add "Потенційна орфографічна помилка".
func (r *MorfologikUkrainianSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil || sentence == nil {
		return nil, nil
	}
	out, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil {
		return out, err
	}
	// Only when a real dict is active (fail-closed: empty dict must not invent flags).
	if !r.dictActive() {
		return out, nil
	}
	work := sentence
	if r.SpellingCheckRule != nil {
		work = r.SpellingCheckRule.SentenceWithImmunization(sentence)
		r.SpellingCheckRule.MarkMultiWordIgnoreSpelling(work)
	}
	tokens := work.GetTokensWithoutWhitespace()
	covered := map[int]bool{}
	for _, m := range out {
		if m == nil {
			continue
		}
		// mark from-pos covered by existing matches
		covered[m.GetFromPos()] = true
	}
	for idx, tok := range tokens {
		if spelling.CanBeIgnoredToken(tok) {
			continue
		}
		if r.SkipTokenFn != nil && r.SkipTokenFn(tok) {
			continue
		}
		if r.SpellingCheckRule != nil && r.IgnoreToken(tokens, idx) {
			// ignoreToken ends with hasGoodTag — good tags never reach this arm
			continue
		}
		// spell-check surface without ignored chars (soft hyphen / combining acute)
		w := spellCheckWordUK(tok)
		if w == "" || !hasLetterUK(w) {
			continue
		}
		if tools.IsEmoji(w) || tools.IsEmoji(tok.GetToken()) {
			continue
		}
		startPos := tok.GetStartPos()
		if covered[startPos] {
			continue
		}
		// Java: super.getRuleMatches empty (not misspelled / not prohibited) + !hasGoodTag
		if r.IsProhibited(w) {
			continue
		}
		// AcceptWord true ⇒ spelled OK / ignored list — we want spelled OK only
		if r.SpellingCheckRule != nil && r.IgnoreWord(w) {
			continue
		}
		if r.IsMisspelled != nil && r.IsMisspelled(w) {
			// misspelled tokens already handled by base Match
			continue
		}
		if hasGoodTagUK(tok) {
			continue
		}
		// disambig may remove v_kly leaving token with no good tags
		m := rules.NewRuleMatch(r, sentence, startPos, startPos+tokenizers.UTF16Len(w),
			"Потенційна орфографічна помилка")
		// Prefer token end when surface has hidden chars
		if tok.GetEndPos() > m.GetToPos() {
			m.SetOffsetPosition(startPos, tok.GetEndPos())
		}
		out = append(out, m)
		covered[startPos] = true
	}
	return out, nil
}

// dictActive reports whether misspell checks use a real dictionary (binary multi or map words).
func (r *MorfologikUkrainianSpellerRule) dictActive() bool {
	if r == nil || r.MorfologikSpellerRule == nil {
		return false
	}
	if r.Speller1 != nil && len(r.Speller1.Spellers) > 0 {
		for _, s := range r.Speller1.Spellers {
			if s != nil && s.HasDictionary() {
				return true
			}
		}
	}
	if r.Multi != nil && len(r.Multi.Spellers) > 0 {
		for _, s := range r.Multi.Spellers {
			if s != nil && s.HasDictionary() {
				return true
			}
		}
	}
	return r.Speller != nil && r.Speller.HasDictionary()
}

// ukGetAdditionalSuggestions ports getAdditionalSuggestions (dash prefixes).
// Java does not re-capitalize the suggestion when the original word was capitalized.
func (r *MorfologikUkrainianSpellerRule) ukGetAdditionalSuggestions(existing []string, word string) []string {
	return additionalDashPrefixSuggestions(word)
}

func additionalDashPrefixSuggestions(word string) []string {
	if word == "" {
		return nil
	}
	w := word
	if tools.IsCapitalizedWord(word) {
		// Java: word = word.toLowerCase() for matching only
		w = strings.ToLower(word)
	}
	prefs := loadDashPrefixesSpeller()
	var out []string
	for key := range prefs {
		if !strings.HasPrefix(w, key) {
			continue
		}
		// Java: word.length() > key.length() + 2 (UTF-16)
		if tokenizers.UTF16Len(w) <= tokenizers.UTF16Len(key)+2 {
			continue
		}
		// Java: word.charAt(key.length()) != '-'
		rest := []rune(w)
		kr := []rune(key)
		if len(rest) <= len(kr) || rest[len(kr)] == '-' {
			continue
		}
		// Java: suggestions.add(key + "-" + second) — second from lowercased word
		out = append(out, key+"-"+string(rest[len(kr):]))
	}
	return out
}

// filterUKSuggestions ports filterSuggestions after super:
// remove spaced DO_NOT_SUGGEST prefixes and "- " fragments.
func filterUKSuggestions(suggestions []string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	out := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		// Java: contains(" ") && PATTERN.matches() || contains("- ")
		if strings.Contains(s, "- ") {
			continue
		}
		if strings.Contains(s, " ") && ukDoNotSuggestSpaced.MatchString(s) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func spellCheckWordUK(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	if c := tok.GetCleanToken(); c != "" && c != tok.GetToken() {
		return c
	}
	if n := len(tok.GetReadings()); n > 0 {
		if at := tok.GetAnalyzedToken(0); at != nil {
			if t := at.GetToken(); t != "" {
				return t
			}
		}
	}
	return tok.GetToken()
}

func hasLetterUK(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// capitalizeFirstUK kept for tests/helpers; Java getAdditionalSuggestions does not capitalize.
func capitalizeFirstUK(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}
