package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	MorfologikUkrainianSpellerRuleID = "MORFOLOGIK_RULE_UK_UA"
	UkrainianSpellerDict             = "/uk/hunspell/uk_UA.dict"
	ukAbbreviationChar               = "."
)

var (
	ukrainianLetters = regexp.MustCompile(`.*[а-яіїєґА-ЯІЇЄҐ].*`)
	// Java DO_NOT_SUGGEST_SPACED_PATTERN (full-string via Matcher.matches → ^...$)
	ukDoNotSuggestSpaced = regexp.MustCompile(
		`^(авіа|авто|анти|аудіо|відео|водо|гідро|екстра|квазі|кіно|лже|мета|моно|мото|псевдо|пост|радіо|стерео|супер|ультра|фото) .*`)
	ukSingleCapital     = regexp.MustCompile(`^[А-ЯІЇЄҐ]$`)
	ukDashPrefixKeyOK   = regexp.MustCompile(`^[а-яіїєґ]{3,}$`)
	ukDashPrefixBadVal  = regexp.MustCompile(`^:(bad|alt|slang)$`)
	dashSpellerOnce     sync.Once
	dashPrefixesSpeller map[string]struct{}
)

// loadDashPrefixesSpeller ports dashPrefixes2019 for getAdditionalSuggestions.
func loadDashPrefixesSpeller() map[string]struct{} {
	dashSpellerOnce.Do(func() {
		out := map[string]struct{}{}
		data, err := dashPrefixFS.ReadFile("data/dash_prefixes.txt")
		if err != nil {
			dashPrefixesSpeller = out
			return
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line[0] == '#' {
				continue
			}
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}
			key := strings.ToLower(parts[0])
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

func NewMorfologikUkrainianSpellerRule() *MorfologikUkrainianSpellerRule {
	r := &MorfologikUkrainianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikUkrainianSpellerRuleID, "uk", UkrainianSpellerDict, nil),
	}
	// Java isLatinScript() = false
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
	}
	r.IgnoreTokenFn = r.ukIgnoreToken
	inner := r.IsMisspelled
	r.IsMisspelled = func(word string) bool {
		return r.ukIsMisspelled(word, inner)
	}
	_ = loadDashPrefixesSpeller()
	return r
}

func (r *MorfologikUkrainianSpellerRule) ukIsMisspelled(word string, inner func(string) bool) bool {
	if word == "" {
		return false
	}
	// Java: if word.endsWith("-") return !word.startsWith("-");
	if strings.HasSuffix(word, "-") {
		return !strings.HasPrefix(word, "-")
	}
	if inner != nil {
		return inner(word)
	}
	return false
}

func (r *MorfologikUkrainianSpellerRule) ukIgnoreToken(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool {
	if idx < 0 || idx >= len(tokens) || tokens[idx] == nil {
		return false
	}
	word := tokens[idx].GetToken()
	if !ukrainianLetters.MatchString(word) {
		return true
	}
	if r.SpellingCheckRule != nil && r.IgnoreWord(word) {
		return true
	}
	if idx < len(tokens)-1 && tokens[idx+1] != nil && tokens[idx+1].GetToken() == ukAbbreviationChar {
		if r.SpellingCheckRule != nil && r.IgnoreWord(word+ukAbbreviationChar) {
			return true
		}
		if ukSingleCapital.MatchString(word) {
			return true
		}
	}
	return hasGoodTagUK(tokens[idx])
}

// hasGoodTag ports hasGoodTag (any POS except null / SENT_END / PARA_END).
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
		case "SENT_END", "PARA_END", "SENTENCE_END", "PARAGRAPH_END":
			continue
		}
		return true
	}
	return false
}

// Match ports match + getRuleMatches !hasGoodTag + filterSuggestions + dash tops.
func (r *MorfologikUkrainianSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil || sentence == nil {
		return nil, nil
	}
	work := sentence
	if r.SpellingCheckRule != nil {
		work = r.SpellingCheckRule.SentenceWithImmunization(sentence)
		r.SpellingCheckRule.MarkMultiWordIgnoreSpelling(work)
	}
	tokens := work.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	for idx, tok := range tokens {
		if spelling.CanBeIgnoredToken(tok) {
			continue
		}
		if r.SkipTokenFn != nil && r.SkipTokenFn(tok) {
			continue
		}
		if r.SpellingCheckRule != nil && r.IgnoreToken(tokens, idx) {
			continue
		}
		w := tok.GetToken()
		if w == "" || !hasLetterUK(w) {
			continue
		}
		if r.AcceptWord(w) {
			// Java getRuleMatches: empty super matches + !hasGoodTag → still flag.
			// Only when a real dict is active (fail-closed empty Words must not invent flags).
			if r.dictActive() && !hasGoodTagUK(tok) {
				m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
					"Потенційна орфографічна помилка")
				out = append(out, m)
			}
			continue
		}
		if r.SpellingCheckRule != nil && r.IgnorePotentiallyMisspelledWord(w) {
			continue
		}
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
			"Possible spelling mistake found")
		if sugs := r.collectSuggestions(w); len(sugs) > 0 {
			m.SetSuggestedReplacements(sugs)
		}
		out = append(out, m)
	}
	return out, nil
}

// dictActive reports whether misspell checks use a real dictionary (map words or CFSA2 wire).
func (r *MorfologikUkrainianSpellerRule) dictActive() bool {
	if FilterDictAvailableUK() {
		return true
	}
	return r != nil && r.Speller != nil && len(r.Speller.Words) > 0
}

func (r *MorfologikUkrainianSpellerRule) collectSuggestions(word string) []string {
	var sugs []string
	if r.Speller != nil {
		sugs = append(sugs, r.Speller.FindReplacements(word)...)
	}
	if FilterDictAvailableUK() {
		sugs = append(sugs, FilterDictSuggestUK(word)...)
	}
	sugs = append(sugs, additionalDashPrefixSuggestions(word)...)
	return filterUKSuggestions(sugs)
}

func additionalDashPrefixSuggestions(word string) []string {
	if word == "" {
		return nil
	}
	w := word
	cap := tools.IsCapitalizedWord(word)
	if cap {
		w = strings.ToLower(word)
	}
	prefs := loadDashPrefixesSpeller()
	var out []string
	for key := range prefs {
		if !strings.HasPrefix(w, key) {
			continue
		}
		if tokenizers.UTF16Len(w) <= tokenizers.UTF16Len(key)+2 {
			continue
		}
		rest := []rune(w)
		kr := []rune(key)
		if len(rest) <= len(kr) || rest[len(kr)] == '-' {
			continue
		}
		sug := key + "-" + string(rest[len(kr):])
		if cap {
			sug = capitalizeFirstUK(sug)
		}
		out = append(out, sug)
	}
	return out
}

func filterUKSuggestions(suggestions []string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	out := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
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

func capitalizeFirstUK(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func hasLetterUK(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
