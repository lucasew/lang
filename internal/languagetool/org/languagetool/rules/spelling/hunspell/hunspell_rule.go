package hunspell

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// HunspellRuleID ports HunspellRule.RULE_ID.
const HunspellRuleID = "HUNSPELL_RULE"

// FileExtension ports HunspellRule.FILE_EXTENSION.
const FileExtension = ".dic"

// tooManyErrorsMsg ports MessagesBundle too_many_errors (en default).
const tooManyErrorsMsg = "(suggestion limit reached)"

var (
	nonAlphabeticRE             = regexp.MustCompile(`^[^\p{L}]+$`)
	minusPlusRE                 = regexp.MustCompile(`^-+$`)
	startsWithTwoUppercaseChars = regexp.MustCompile(`^[A-Z][A-Z]\p{Ll}+`)
)

// HunspellRule ports org.languagetool.rules.spelling.hunspell.HunspellRule
// with a pluggable HunspellDictionary (native hunspell deferred).
type HunspellRule struct {
	*spelling.SpellingCheckRule
	Dict HunspellDictionary
	// IgnoreTaggedWords skips tokens that already have a real POS tag.
	IgnoreTaggedWords bool
	// UserConfig ports HunspellRule.userConfig (suggestionsEnabled / maxSpellingSuggestions /
	// preferredLanguages for ForeignLanguageChecker).
	UserConfig *languagetool.UserConfig
	// ForeignDetect ports ForeignLanguageChecker language-id hook.
	// When nil, foreign-language scoring is inactive (Java langIdent == null → empty).
	ForeignDetect spelling.DetectScoresFunc
	// GetOnlySuggestionsFn ports getOnlySuggestions: when non-empty, replaces dict sugs.
	GetOnlySuggestionsFn func(word string) []string
	// GetAdditionalTopSuggestionsFn ports getAdditionalTopSuggestions language overrides.
	GetAdditionalTopSuggestionsFn func(existing []string, word string) []string
	// AcceptSuggestionFn ports acceptSuggestion (default true).
	AcceptSuggestionFn func(suggestion string) bool
	// SuggestFn optional override for getSuggestions (CompoundAwareHunspellRule wires this
	// so Match/calcSuggestions dispatch to compound-aware logic — Go embedding does not
	// virtualize Suggest like Java overrides).
	SuggestFn func(word string) []string
	// NonWordSplitter ports HunspellRule.nonWordPattern (from .aff WORDCHARS or NON_ALPHABETIC).
	// Used by TokenizeText / getCorrectWords. Zero value → letters-only (default).
	NonWordSplitter NonWordSplitter
	// IsQuotedCompoundFn ports isQuotedCompound override (German). Base returns false.
	IsQuotedCompoundFn func(sentence *languagetool.AnalyzedSentence, idx int, token string) bool
}

func NewHunspellRule(languageCode string, dict HunspellDictionary) *HunspellRule {
	r := &HunspellRule{
		SpellingCheckRule: spelling.NewSpellingCheckRule(HunspellRuleID, spelling.DescSpelling, languageCode),
		Dict:              dict,
	}
	r.IsMisspelled = r.IsMisspelledWord
	// Java SpellingCheckRule.init: ignore/spelling/prohibit word lists for language.
	ApplyDefaultSpellingWordLists(r.SpellingCheckRule)
	return r
}

// SetUserConfig stores UserConfig for Match suggestion / foreign-language gates.
func (r *HunspellRule) SetUserConfig(uc *languagetool.UserConfig) {
	if r != nil {
		r.UserConfig = uc
	}
}

// IsMisspelledWord ports HunspellRule.isMisspelled (minus ignoreWord — applied in Match
// via AcceptWord / IgnoreToken; prohibited always flags).
func (r *HunspellRule) IsMisspelledWord(word string) bool {
	if r == nil {
		return false
	}
	// Java: isProhibited(cutOffDot(word)) forces misspell even when dict accepts.
	if r.SpellingCheckRule != nil && r.IsProhibited(cutOffDotHun(word)) {
		return true
	}
	if r.Dict == nil {
		return false
	}
	if word == "--" {
		return false
	}
	// Java: length==1 → only alphabetic punctuation check; else treat as alphabetic path.
	if utf16LenHun(word) == 1 {
		rns := []rune(word)
		if len(rns) == 1 && !unicode.IsLetter(rns[0]) {
			return false
		}
	}
	if nonAlphabeticRE.MatchString(word) {
		return false
	}
	// Java: (hunspell != null && !hunspell.spell(word)) && !ignoreWord(word)
	if r.SpellingCheckRule != nil && r.IgnoreWord(word) {
		return false
	}
	return !r.Dict.Spell(word)
}

// Suggest ports HunspellRule.getSuggestions (dictionary only).
// When SuggestFn is set (e.g. CompoundAware), that override is used.
func (r *HunspellRule) Suggest(word string) []string {
	if r == nil {
		return nil
	}
	if r.SuggestFn != nil {
		return r.SuggestFn(word)
	}
	if r.Dict == nil {
		return nil
	}
	return r.Dict.Suggest(word)
}

// Match flags misspelled tokens in the analyzed sentence.
// Ports HunspellRule.match: wrong-split, leading-dash, UserConfig sug gates,
// ForeignLanguageChecker, high-confidence DE case, Type.UnknownWord.
func (r *HunspellRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	// Java: if (hunspell == null) return empty
	if r.Dict == nil {
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

	// ForeignLanguageChecker when preferredLanguages ≥ 2 (Java).
	var foreignChecker *spelling.ForeignLanguageChecker
	gotForeignResults := false
	if r.UserConfig != nil {
		pref := r.UserConfig.GetPreferredLanguages()
		if preferredLanguagesActiveHun(pref) {
			langCode := ""
			if r.SpellingCheckRule != nil {
				langCode = r.SpellingCheckRule.LanguageCode
			}
			if i := strings.IndexAny(langCode, "-_"); i > 0 {
				langCode = langCode[:i]
			}
			text := ""
			if sentence != nil {
				text = sentence.GetText()
			}
			foreignChecker = spelling.NewForeignLanguageChecker(langCode, text, foreignSentenceLengthHun(sentence), pref)
			if r.ForeignDetect != nil {
				foreignChecker.Detect = r.ForeignDetect
			}
		}
	}

	var prevTok *languagetool.AnalyzedTokenReadings
	for idx, tok := range tokens {
		// Java canBeIgnored-like: immunized / ignoredBySpeller / isUrl / isEMail
		if spelling.CanBeIgnoredToken(tok) {
			prevTok = tok
			continue
		}
		// Java ignoreToken → ignoreWord (via IgnoreToken)
		if r.SpellingCheckRule != nil && r.IgnoreToken(tokens, idx) {
			prevTok = tok
			continue
		}
		w := tok.GetToken()
		if w == "" || !hasLetter(w) {
			prevTok = tok
			continue
		}
		// Java getSentenceTextWithoutUrlsAndImmunizedTokens: stringForSpeller
		// (no String.trim on token; split path yields non-empty segments).
		// After emoji→space replacement, ASCII spaces may pad the surface —
		// Java nonWordPattern split drops them; JavaStringTrim matches that for
		// pure ASCII padding without invent Unicode TrimSpace.
		check := tools.StringForSpeller(w)
		check = tools.JavaStringTrim(check)
		if check == "" || !hasLetter(check) {
			prevTok = tok
			continue
		}
		if r.IgnoreTaggedWords && tok.IsTagged() {
			if r.SpellingCheckRule == nil || !r.IsProhibited(w) {
				prevTok = tok
				continue
			}
		}
		// Java: (ignoreWord(...) || ignoreWord(word)) && !isProhibited(cutOffDot(word))
		// AcceptWord already folds ignore + prohibited + dict.
		if r.AcceptWord(check) {
			prevTok = tok
			continue
		}
		// Java: after isMisspelled, ignorePotentiallyMisspelledWord
		if r.SpellingCheckRule != nil && r.IgnorePotentiallyMisspelledWord(check) {
			prevTok = tok
			continue
		}

		cleanWord := cutOffDotHun(check)
		dashCorr := 0
		// Java: word.startsWith("-")
		if strings.HasPrefix(check, "-") {
			rest := cleanWord
			if len(rest) > 0 {
				// UTF-16 first unit may be multi-byte; use rune-aware drop of first '-'
				rest = strings.TrimPrefix(rest, "-")
			}
			if !r.IsMisspelledWord(rest) || minusPlusRE.MatchString(cleanWord) {
				prevTok = tok
				continue
			}
			dashCorr = 1
		}

		// Java wrong-split: may ADD matches (does not skip the per-word match below).
		if prevTok != nil {
			prevWord := tools.StringForSpeller(prevTok.GetToken())
			prevWord = tools.JavaStringTrim(prevWord)
			if prevWord != "" && !r.ignoreWrongSplit(prevWord, check) {
				r.addWrongSplits(sentence, &out, prevWord, prevTok.GetStartPos(), check, tok.GetStartPos(), cleanWord)
			}
		}

		// Java: RuleMatch from len+dashCorr .. len+cleanWord.length()
		// Token-based twin uses ATR positions with dashCorr on fromPos.
		from := tok.GetStartPos() + dashCorr
		to := tok.GetStartPos() + utf16LenHun(cleanWord)
		// Prefer token end when no dash correction and end is known.
		if dashCorr == 0 && tok.GetEndPos() > to {
			to = tok.GetEndPos()
		}
		m := spelling.NewSpellingRuleMatch(r, sentence, from, to)

		cleanWord2 := cleanWord
		if dashCorr > 0 && utf16LenHun(cleanWord) > dashCorr {
			// Java: cleanWord.substring(dashCorr)
			u := utf16.Encode([]rune(cleanWord))
			if dashCorr < len(u) {
				cleanWord2 = string(utf16.Decode(u[dashCorr:]))
			}
		}

		// Java: userConfig suggestions gates
		if r.UserConfig != nil && !r.UserConfig.IsSuggestionsEnabled() {
			m.SetSuggestedReplacements(nil)
		} else if r.allowMoreSpellingSuggestions(len(out)) {
			sug := r.calcSuggestions(check, cleanWord2)
			if r.isFirstItemHighConfidenceSuggestion(check, sug) && len(sug) > 0 {
				// Attach high confidence on SuggestedReplacementObjects path.
				objs := make([]*rules.SuggestedReplacement, 0, len(sug))
				for i, s := range sug {
					sr := rules.NewSuggestedReplacement(s)
					if i == 0 {
						c := rules.SpellingHighConfidence
						sr.SetConfidence(&c)
					}
					objs = append(objs, sr)
				}
				m.SetSuggestedReplacementObjects(objs)
			} else if len(sug) > 0 {
				m.SetSuggestedReplacements(sug)
			}
		} else {
			m.SetSuggestedReplacement(tooManyErrorsMsg)
		}
		out = append(out, m)

		if foreignChecker != nil && !gotForeignResults {
			scores := foreignChecker.Check(len(out))
			if len(scores) > 0 {
				if _, noForeign := scores[spelling.NoForeignLangDetected]; !noForeign && out[0] != nil {
					out[0].SetNewLanguageMatches(scores)
				}
				gotForeignResults = true
			}
		}
		prevTok = tok
	}
	return out, nil
}

func (r *HunspellRule) allowMoreSpellingSuggestions(ruleMatchesSoFar int) bool {
	if r == nil || r.UserConfig == nil {
		return true
	}
	max := r.UserConfig.GetMaxSpellingSuggestions()
	if max == 0 {
		return true
	}
	return ruleMatchesSoFar <= max
}

// ignoreWrongSplit ports PT/DE common-word skip for wrong-split.
func (r *HunspellRule) ignoreWrongSplit(prevWord, word string) bool {
	if r == nil || r.SpellingCheckRule == nil {
		return false
	}
	code := strings.ToLower(r.SpellingCheckRule.LanguageCode)
	if i := strings.IndexAny(code, "-_"); i > 0 {
		code = code[:i]
	}
	pl := strings.ToLower(prevWord)
	wl := strings.ToLower(word)
	switch code {
	case "pt":
		_, a := commonPortugueseWords[pl]
		_, b := commonPortugueseWords[wl]
		return a || b
	case "de":
		_, a := commonGermanWords[pl]
		_, b := commonGermanWords[wl]
		return a || b
	}
	return false
}

// addWrongSplits ports both wrong-split arms (may add 0–2 matches; second may replace first).
func (r *HunspellRule) addWrongSplits(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	prevWord string,
	prevFrom int,
	word string,
	wordFrom int,
	cleanWord string,
) {
	// "thanky ou" → "thank you"
	if pu := utf16.Encode([]rune(prevWord)); len(pu) >= 1 {
		sugg1a := string(utf16.Decode(pu[:len(pu)-1]))
		sugg1b := cutOffDotHun(string(utf16.Decode(pu[len(pu)-1:])) + word)
		// Java: acceptSuggestion(sugg1a + " " + sugg1b) — no trim
		joined := sugg1a + " " + sugg1b
		if sugg1a != "" && sugg1b != "" &&
			!r.IsMisspelledWord(sugg1a) && !r.IsMisspelledWord(sugg1b) &&
			r.acceptSuggestion(joined) {
			if rm := r.createWrongSplitMatch(sentence, ruleMatches, wordFrom, cleanWord, sugg1a, sugg1b, prevFrom); rm != nil {
				*ruleMatches = append(*ruleMatches, rm)
			}
		}
	}
	// "than kyou" → "thank you"
	if wu := utf16.Encode([]rune(word)); len(wu) > 1 {
		sugg2a := prevWord + string(utf16.Decode(wu[:1]))
		sugg2b := cutOffDotHun(string(utf16.Decode(wu[1:])))
		// Java: acceptSuggestion(sugg2a + " " + sugg2b) — no trim
		joined := sugg2a + " " + sugg2b
		if sugg2a != "" && sugg2b != "" &&
			!r.IsMisspelledWord(sugg2a) && !r.IsMisspelledWord(sugg2b) &&
			r.acceptSuggestion(joined) {
			if rm := r.createWrongSplitMatch(sentence, ruleMatches, wordFrom, cleanWord, sugg2a, sugg2b, prevFrom); rm != nil {
				*ruleMatches = append(*ruleMatches, rm)
			}
		}
	}
}

func (r *HunspellRule) acceptSuggestion(s string) bool {
	if r != nil && r.AcceptSuggestionFn != nil {
		return r.AcceptSuggestionFn(s)
	}
	return true
}

// createWrongSplitMatch ports SpellingCheckRule.createWrongSplitMatch via shared twin.
func (r *HunspellRule) createWrongSplitMatch(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	pos int,
	coveredWord, suggestion1, suggestion2 string,
	prevPos int,
) *rules.RuleMatch {
	return spelling.CreateWrongSplitMatch(r, sentence, ruleMatches, pos, coveredWord, suggestion1, suggestion2, prevPos)
}

// calcSuggestions ports HunspellRule.calcSuggestions (core arms).
func (r *HunspellRule) calcSuggestions(word, cleanWord string) []string {
	if r == nil {
		return nil
	}
	if r.GetOnlySuggestionsFn != nil {
		if only := r.GetOnlySuggestionsFn(cleanWord); len(only) > 0 {
			return r.filterSugs(only)
		}
	}
	suggestions := r.Suggest(cleanWord)
	// Java: if word.endsWith("."), interleave suggestions for word with trailing dot stripped
	if strings.HasSuffix(word, ".") {
		pos := 1
		for _, s := range r.Suggest(word) {
			// insert stripped-dot form at pos (mixing lists)
			stripped := s
			if strings.HasSuffix(s, ".") {
				stripped = s[:len(s)-1]
			}
			if !containsStr(suggestions, stripped) {
				if pos > len(suggestions) {
					pos = len(suggestions)
				}
				suggestions = insertAt(suggestions, pos, stripped)
				pos += 2
			}
		}
	}
	// additional top suggestions
	var top []string
	if r.GetAdditionalTopSuggestionsFn != nil {
		top = r.GetAdditionalTopSuggestionsFn(suggestions, cleanWord)
	}
	if len(top) == 0 {
		top = spelling.AdditionalTopSuggestions(suggestions, cleanWord)
	}
	if len(top) == 0 && strings.HasSuffix(word, ".") {
		if r.GetAdditionalTopSuggestionsFn != nil {
			top = r.GetAdditionalTopSuggestionsFn(suggestions, word)
		} else {
			top = spelling.AdditionalTopSuggestions(suggestions, word)
		}
		// Java: append "." when top does not end with "."
		for i, t := range top {
			if !strings.HasSuffix(t, ".") {
				top[i] = t + "."
			}
		}
	}
	// Java Collections.reverse(additionalTopSuggestions) then add(0, ...)
	for i := len(top) - 1; i >= 0; i-- {
		t := top[i]
		if t != cleanWord {
			suggestions = append([]string{t}, suggestions...)
		}
	}
	// filter acceptSuggestion + filterSuggestions + filterDupes
	filtered := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		if r.acceptSuggestion(s) {
			filtered = append(filtered, s)
		}
	}
	return r.filterSugs(filtered)
}

func (r *HunspellRule) filterSugs(sug []string) []string {
	if r.SpellingCheckRule != nil {
		sug = r.SpellingCheckRule.FilterSuggestions(sug)
	}
	return filterDupesHun(sug)
}

// isFirstItemHighConfidenceSuggestion ports HunspellRule.isFirstItemHighConfidenceSuggestion (DE).
func (r *HunspellRule) isFirstItemHighConfidenceSuggestion(word string, sug []string) bool {
	if len(sug) == 0 || word == "IPs" {
		return false
	}
	if !strings.EqualFold(word, sug[0]) {
		return false
	}
	if !startsWithTwoUppercaseChars.MatchString(word) {
		return false
	}
	code := ""
	if r != nil && r.SpellingCheckRule != nil {
		code = strings.ToLower(r.SpellingCheckRule.LanguageCode)
	}
	if i := strings.IndexAny(code, "-_"); i > 0 {
		code = code[:i]
	}
	if code != "de" {
		return false
	}
	// Java: word.endsWith("s") && StringUtils.isAllUpperCase(sugg.get(0)) → false
	if strings.HasSuffix(word, "s") && isAllUpperCase(sug[0]) {
		return false
	}
	return true
}

func isAllUpperCase(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func cutOffDotHun(s string) string {
	if strings.HasSuffix(s, ".") {
		return s[:len(s)-1]
	}
	return s
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func utf16LenHun(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func filterDupesHun(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func containsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func insertAt(ss []string, i int, s string) []string {
	if i < 0 {
		i = 0
	}
	if i >= len(ss) {
		return append(ss, s)
	}
	out := make([]string, 0, len(ss)+1)
	out = append(out, ss[:i]...)
	out = append(out, s)
	out = append(out, ss[i:]...)
	return out
}

func preferredLanguagesActiveHun(pref []string) bool {
	n := 0
	for _, p := range pref {
		if strings.TrimSpace(p) != "" {
			n++
		}
	}
	return n >= 2
}

func foreignSentenceLengthHun(sentence *languagetool.AnalyzedSentence) int64 {
	if sentence == nil {
		return 0
	}
	var n int64
	for _, t := range sentence.GetTokensWithoutWhitespace() {
		if t != nil && !t.IsNonWord() {
			n++
		}
	}
	return n - 1
}
