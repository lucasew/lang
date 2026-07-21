package morfologik

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// defaultCompoundRegex ports MorfologikSpellerRule.compoundRegex default "-".
var defaultCompoundRegex = regexp.MustCompile(`-`)

// MorfologikSpellerRule ports org.languagetool.rules.spelling.morfologik.MorfologikSpellerRule
// (map/dict-backed; binary morfologik deferred).
type MorfologikSpellerRule struct {
	*spelling.SpellingCheckRule
	Speller           *MorfologikSpeller
	IgnoreTaggedWords bool
	// FileName is the dictionary path from getFileName().
	FileName string
	// CheckCompound ports checkCompound (setCheckCompound): if true, a misspelled
	// whole word is accepted when every compoundRegex part is accepted.
	CheckCompound bool
	// CompoundRegex ports compoundRegex (default "-"). Nil → defaultCompoundRegex.
	CompoundRegex *regexp.Regexp
	// SkipTokenFn ports language-specific getRuleMatches early exits
	// (e.g. NL/PT: tokens[idx].hasPosTag("_english_ignore_")).
	SkipTokenFn func(tok *languagetool.AnalyzedTokenReadings) bool
	// GetOnlySuggestionsFn ports getOnlySuggestions: when non-empty, replaces all
	// other speller suggestions (Java calcSpellerSuggestions early return).
	GetOnlySuggestionsFn func(word string) []string
	// GetAdditionalTopSuggestionsFn ports getAdditionalTopSuggestions language overrides
	// (e.g. EN curated maps). Prepended before dict suggestions; empty → base LanguageTool tops.
	GetAdditionalTopSuggestionsFn func(existing []string, word string) []string
	// AddHyphenSuggestionsFn ports addHyphenSuggestions: when dict sugs empty and word
	// contains '-', rebuild hyphenated forms by fixing one misspelled part (EN).
	AddHyphenSuggestionsFn func(parts []string) []string
}

func NewMorfologikSpellerRule(id, languageCode, fileName string, speller *MorfologikSpeller) *MorfologikSpellerRule {
	if speller == nil {
		speller = NewMorfologikSpeller(fileName, 1)
	}
	r := &MorfologikSpellerRule{
		SpellingCheckRule: spelling.NewSpellingCheckRule(id, "Possible spelling mistake", languageCode),
		Speller:           speller,
		FileName:          fileName,
	}
	// Binary .dict load is deferred: empty Words means "dict not loaded".
	// Fail closed (do not invent misspell flags) — same policy as HunspellRule with nil dict.
	// When Words are map-injected for tests/partial dicts, Speller.IsMisspelled applies.
	// Compound-aware path ports isMisspelled(MorfologikMultiSpeller, word) checkCompound arm.
	r.IsMisspelled = func(word string) bool {
		return r.isMisspelledWord(word)
	}
	// Java setConvertsCase(speller.convertsCase()) after multi-speller init.
	if r.SpellingCheckRule != nil && r.Speller != nil {
		r.ConvertsCase = r.Speller.ConvertsCase()
	}
	// Java SpellingCheckRule.init: ignore/spelling/prohibit for language short code.
	spelling.ApplyDefaultSpellingWordLists(r.SpellingCheckRule)
	return r
}

func (r *MorfologikSpellerRule) GetFileName() string { return r.FileName }

// Match flags misspelled tokens.
func (r *MorfologikSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	// Java MorfologikSpellerRule.match: tokens = getSentenceWithImmunization(sentence)...
	work := sentence
	if r.SpellingCheckRule != nil {
		work = r.SpellingCheckRule.SentenceWithImmunization(sentence)
		// Also mark multi-word IGNORE_SPELLING phrases (redundant with antipatterns when
		// Replace matches; kept for Match when pattern matcher misses).
		r.SpellingCheckRule.MarkMultiWordIgnoreSpelling(work)
	}
	tokens := work.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	for idx, tok := range tokens {
		// Java canBeIgnored: SENT_START, immunized, ignore-spelling, URL, email.
		if spelling.CanBeIgnoredToken(tok) {
			continue
		}
		// Java language getRuleMatches early exit (e.g. _english_ignore_).
		if r.SkipTokenFn != nil && r.SkipTokenFn(tok) {
			continue
		}
		// Java canBeIgnored: ignoreToken(tokens, idx) → ignoreWord.
		if r.SpellingCheckRule != nil && r.IgnoreToken(tokens, idx) {
			continue
		}
		w := tok.GetToken()
		if w == "" || !hasLetter(w) {
			continue
		}
		// Java MorfologikSpellerRule.ignoreWord: super.ignoreWord || StringTools.isEmoji
		if tools.IsEmoji(w) {
			continue
		}
		if r.IgnoreTaggedWords && tok.IsTagged() {
			// Java: ignoreTaggedWords && isTagged && !isProhibited
			if r.SpellingCheckRule == nil || !r.IsProhibited(w) {
				continue
			}
		}
		// Dictionary misspell (ignore set already handled by IgnoreToken/IgnoreWord).
		if r.AcceptWord(w) {
			continue
		}
		// Java getRuleMatches: after isMisspelled, ignorePotentiallyMisspelledWord.
		if r.SpellingCheckRule != nil && r.IgnorePotentiallyMisspelledWord(w) {
			continue
		}
		startPos := tok.GetStartPos()
		// Java: previous match already covers this token (wrong-split span).
		if len(out) > 0 && out[len(out)-1] != nil && out[len(out)-1].GetToPos() > startPos {
			continue
		}

		// Java getRuleMatches wrong-split with previous / next word.
		ruleMatch, beforeStr, early := r.tryWrongSplitPrev(sentence, &out, idx, tokens, w, startPos)
		if early && ruleMatch != nil {
			out = append(out, ruleMatch)
			continue
		}
		if ruleMatch == nil {
			var afterStr string
			ruleMatch, afterStr, early = r.tryWrongSplitNext(sentence, &out, idx, tokens, w, startPos)
			if early && ruleMatch != nil {
				out = append(out, ruleMatch)
				continue
			}
			_ = afterStr
			if ruleMatch != nil {
				// wrong-split with correctly spelled neighbor: keep span, still append dict sugs
				sug := r.collectSuggestions(w)
				for _, s := range sug {
					joined := strings.TrimSpace(beforeStr + s)
					if afterStr != "" {
						joined = strings.TrimSpace(beforeStr + s + afterStr)
					}
					addSug(ruleMatch, joined)
				}
				out = append(out, ruleMatch)
				continue
			}
		} else {
			// prev wrong-split but prev not misspelled: keep span + dict sugs with beforeStr
			sug := r.collectSuggestions(w)
			for _, s := range sug {
				addSug(ruleMatch, strings.TrimSpace(beforeStr+s))
			}
			out = append(out, ruleMatch)
			continue
		}

		m := rules.NewRuleMatch(r, sentence, startPos, tok.GetEndPos(),
			"Possible spelling mistake found")
		m.SetType(rules.RuleMatchTypeUnknownWord)
		sug := r.collectSuggestions(w)
		if len(sug) > 0 {
			m.SetSuggestedReplacements(sug)
		}
		out = append(out, m)
	}
	return out, nil
}

// SetCheckCompound ports setCheckCompound (EN enables true).
func (r *MorfologikSpellerRule) SetCheckCompound(on bool) {
	if r != nil {
		r.CheckCompound = on
	}
}

// SetCompoundRegex ports setCompoundRegex.
func (r *MorfologikSpellerRule) SetCompoundRegex(pattern string) {
	if r == nil {
		return
	}
	if pattern == "" {
		r.CompoundRegex = defaultCompoundRegex
		return
	}
	r.CompoundRegex = regexp.MustCompile(pattern)
}

// isMisspelledWord ports isMisspelled(speller, word) including checkCompound.
func (r *MorfologikSpellerRule) isMisspelledWord(word string) bool {
	if r == nil || word == "" {
		return false
	}
	// Fail-closed empty dict (same as prior IsMisspelled hook).
	if r.Speller == nil || len(r.Speller.Words) == 0 {
		return false
	}
	if !r.Speller.IsMisspelled(word) {
		return false
	}
	if !r.CheckCompound {
		return true
	}
	re := r.CompoundRegex
	if re == nil {
		re = defaultCompoundRegex
	}
	if !re.MatchString(word) {
		return true
	}
	// Java: split and require every part accepted
	parts := re.Split(word, -1)
	for _, p := range parts {
		if p == "" {
			continue
		}
		if r.Speller.IsMisspelled(p) {
			return true
		}
	}
	return false
}

// collectSuggestions ports calcSpellerSuggestions:
// only-suggestions early return; else dict + getAdditionalTopSuggestions; then filterSuggestions.
func (r *MorfologikSpellerRule) collectSuggestions(word string) []string {
	if r == nil {
		return nil
	}
	// Java: getOnlySuggestions non-empty → return those only (still filtered below for parity).
	if r.GetOnlySuggestionsFn != nil {
		if only := r.GetOnlySuggestionsFn(word); len(only) > 0 {
			if r.SpellingCheckRule != nil {
				return r.SpellingCheckRule.FilterSuggestions(only)
			}
			return only
		}
	}
	var sug []string
	if r.Speller != nil {
		sug = r.Speller.FindReplacements(word)
	}
	// Java: if no default/user sugs and word contains "-", addHyphenSuggestions first into top.
	var top []string
	if len(sug) == 0 && strings.Contains(word, "-") && r.AddHyphenSuggestionsFn != nil {
		// Java word.split("-") keeps empty segments for leading/trailing/double hyphens.
		parts := strings.Split(word, "-")
		top = append(top, r.AddHyphenSuggestionsFn(parts)...)
	}
	// Java: topSuggestions.addAll(getAdditionalTopSuggestions(...))
	if r.GetAdditionalTopSuggestionsFn != nil {
		if langTop := r.GetAdditionalTopSuggestionsFn(sug, word); len(langTop) > 0 {
			top = append(top, langTop...)
		} else if baseTop := spelling.AdditionalTopSuggestions(sug, word); len(baseTop) > 0 {
			// EN returns early when curated non-empty; only use base when language top empty.
			top = append(top, baseTop...)
		}
	} else if baseTop := spelling.AdditionalTopSuggestions(sug, word); len(baseTop) > 0 {
		top = append(top, baseTop...)
	}
	if len(top) > 0 {
		sug = append(top, sug...)
	}
	if len(sug) == 0 {
		return nil
	}
	// Java: filterSuggestions (prohibit, " s"→"'s", no-suggest).
	if r.SpellingCheckRule != nil {
		sug = r.SpellingCheckRule.FilterSuggestions(sug)
	}
	return sug
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
