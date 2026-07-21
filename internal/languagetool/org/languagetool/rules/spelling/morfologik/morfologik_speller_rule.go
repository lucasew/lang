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

// pStartsWithNumbersBullets ports MorfologikSpellerRule.pStartsWithNumbersBullets:
// "^(\\d[\\.,\\d]*|\\P{L}+)(.*)$" — leading number (with .,) or non-letters + rest.
var pStartsWithNumbersBullets = regexp.MustCompile(`^(\d[\.,\d]*|\P{L}+)(.*)$`)

// pStartsWithNumbersBulletsExceptions ports pStartsWithNumbersBulletsExceptions:
// "^([\\p{C}\\-\\$%&]+)(.*)$" — control/other marks, $, %, & (do not strip).
var pStartsWithNumbersBulletsExceptions = regexp.MustCompile(`^([\p{C}\-\$%&]+)(.*)$`)

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
	// Multi is an alias for Speller1 (isMisspelled / frequency / legacy callers).
	// Java: protected MorfologikMultiSpeller speller1.
	Multi *MorfologikMultiSpeller
	// Speller1/2/3 port Java speller1 (edit 1), speller2 (edit 2), speller3 (edit 3).
	// calcSpellerSuggestions cascades 1→2→3; isMisspelled uses Speller1 only.
	Speller1 *MorfologikMultiSpeller
	Speller2 *MorfologikMultiSpeller
	Speller3 *MorfologikMultiSpeller
	// FullResults ports calcSpellerSuggestions(fullResults) — when true, always
	// pulls speller2/3 even if speller1 already returned suggestions.
	FullResults bool
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

// SetMultiSpellers ports initSpeller assignment of speller1/speller2/speller3.
// Multi and Speller are wired to speller1 (binary primary) for isMisspelled.
// Pass nil Multis to clear (map-inject tests).
func (r *MorfologikSpellerRule) SetMultiSpellers(s1, s2, s3 *MorfologikMultiSpeller) {
	if r == nil {
		return
	}
	r.Speller1 = s1
	r.Speller2 = s2
	r.Speller3 = s3
	r.Multi = s1
	if s1 != nil {
		// Prefer first non-user speller (binary) as Speller primary.
		for _, sp := range s1.DefaultDictSpellers {
			if sp != nil {
				r.Speller = sp
				break
			}
		}
		if r.Speller == nil && len(s1.Spellers) > 0 {
			r.Speller = s1.Spellers[0]
		}
		// Java: setConvertsCase(speller1.convertsCase()) — Multi field from binary.
		if r.SpellingCheckRule != nil {
			r.ConvertsCase = s1.ConvertsCase()
		}
	}
}

// ClearMultiSpellers disables Multi / Speller1–3 so map-inject Speller is used alone.
func (r *MorfologikSpellerRule) ClearMultiSpellers() {
	if r == nil {
		return
	}
	r.Speller1 = nil
	r.Speller2 = nil
	r.Speller3 = nil
	r.Multi = nil
}

// ApplyUserConfig ports SpellingCheckRule(userConfig) + Multi user-dict (premium only).
// acceptedWords always go to wordsToBeIgnored; user FSA Multis only when premiumUID != nil.
// When rebuilding Multis, binaryClasspath/plainRels/variant/prepareLine match initSpeller paths.
func (r *MorfologikSpellerRule) ApplyUserConfig(acceptedWords []string, premiumUID *int64, binaryClasspath string, plainTextRels []string, languageVariantRel string, prepareLine PrepareLineFn) {
	if r == nil {
		return
	}
	if r.SpellingCheckRule != nil {
		r.SpellingCheckRule.ApplyUserAcceptedWords(acceptedWords)
	}
	userWords := UserDictWordsForMulti(acceptedWords, premiumUID)
	if binaryClasspath == "" {
		binaryClasspath = r.FileName
	}
	if binaryClasspath == "" {
		return
	}
	// Rebuild Multis with same plain paths; edit distances 1/2/3.
	s1 := OpenMultiSpellerFromClasspathWithUser(binaryClasspath, plainTextRels, languageVariantRel, 1, prepareLine, userWords)
	s2 := OpenMultiSpellerFromClasspathWithUser(binaryClasspath, plainTextRels, languageVariantRel, 2, prepareLine, userWords)
	s3 := OpenMultiSpellerFromClasspathWithUser(binaryClasspath, plainTextRels, languageVariantRel, 3, prepareLine, userWords)
	r.SetMultiSpellers(s1, s2, s3)
}

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

		// Java getRuleMatches: word starting with numbers or bullets
		cleanWord, beforePrefix, preventFurther, spaceInsert := r.applyNumbersBulletsPrefix(w)
		if spaceInsert != "" {
			addSug(m, spaceInsert)
		}
		if !preventFurther {
			// Java appendLazySuggestions(cleanWord, beforeSuggestionStr, afterSuggestionStr)
			// afterStr is empty here (wrong-split paths already continued above).
			sug := r.collectSuggestions(cleanWord)
			for _, s := range sug {
				joined := strings.TrimSpace(beforePrefix + s)
				if joined != "" {
					addSug(m, joined)
				}
			}
		}
		out = append(out, m)
	}
	return out, nil
}

// applyNumbersBulletsPrefix ports getRuleMatches numbers/bullets block.
// Returns cleanWord (may strip leading digits/bullets), beforePrefix for suggestions
// ("firstPart "), preventFurther (true when firstPart+" "+secondPart is already added),
// and spaceInsert suggestion (firstPart+" "+secondPart) when the second part is OK.
func (r *MorfologikSpellerRule) applyNumbersBulletsPrefix(word string) (cleanWord, beforePrefix string, preventFurther bool, spaceInsert string) {
	cleanWord = word
	if word == "" || r == nil {
		return word, "", false, ""
	}
	m := pStartsWithNumbersBullets.FindStringSubmatch(word)
	if m == nil {
		return word, "", false, ""
	}
	// Java: if exception matches, skip the whole numbers/bullets arm
	if pStartsWithNumbersBulletsExceptions.MatchString(word) {
		return word, "", false, ""
	}
	firstPart, secondPart := m[1], m[2]
	if secondPart == "" {
		// Nothing to split; keep original word for speller suggestions
		return word, "", false, ""
	}

	// Java: language.getWordTokenizer().tokenize(secondPart)
	lang := "en"
	if r.SpellingCheckRule != nil && r.SpellingCheckRule.LanguageCode != "" {
		lang = r.SpellingCheckRule.LanguageCode
	}
	multitokenMisspelled := false
	if wt := languagetool.WordTokenizerForLanguage(lang); wt != nil {
		for _, t := range wt.Tokenize(secondPart) {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			// Java: anyMatch(str -> isMisspelled(speller1, str))
			if r.isMisspelledWord(t) {
				multitokenMisspelled = true
				break
			}
		}
	} else if r.isMisspelledWord(secondPart) {
		multitokenMisspelled = true
	}

	ignored := false
	if r.SpellingCheckRule != nil {
		ignored = r.SpellingCheckRule.IsIgnoredNoCase(secondPart)
	}
	prohibited := r.IsProhibited(secondPart)

	// Java: (!multitokenIsMisspeled || isIgnoredNoCase(secondPart)) && !isProhibited(secondPart)
	if (!multitokenMisspelled || ignored) && !prohibited {
		// Suggest inserting a space after the leading numbers/bullets
		return word, "", true, firstPart + " " + secondPart
	}
	// Otherwise spell-check the second part only; prefix suggestions with firstPart+" "
	return secondPart, firstPart + " ", false, ""
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
	// Multi-speller: accepted if any component accepts (Java MorfologikMultiSpeller.isMisspelled).
	if r.Multi != nil && len(r.Multi.Spellers) > 0 {
		if !r.Multi.IsMisspelled(word) {
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
		parts := re.Split(word, -1)
		for _, p := range parts {
			if p == "" {
				continue
			}
			if r.Multi.IsMisspelled(p) {
				return true
			}
		}
		return false
	}
	// Fail-closed empty dict (map inject empty and no binary FSA).
	if r.Speller == nil || !r.Speller.HasDictionary() {
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
// only-suggestions early return; speller1 then optional speller2/3 cascade;
// user vs default concat by word length; filterSuggestions / filterDupes.
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

	// Resolve speller1/2/3 (Multi aliases Speller1).
	s1 := r.Speller1
	if s1 == nil {
		s1 = r.Multi
	}
	s2 := r.Speller2
	s3 := r.Speller3
	fullResults := r.FullResults

	var defaultSug, userSug []string
	if s1 != nil {
		defaultSug = s1.GetSuggestionsFromDefaultDicts(word)
		userSug = s1.GetSuggestionsFromUserDicts(word)
	} else if r.Speller != nil {
		defaultSug = r.Speller.FindReplacements(word)
	}

	// Java: onlyCaseDiffers when first default suggestion is case-only change.
	onlyCaseDiffers := false
	if len(defaultSug) > 0 && strings.EqualFold(word, defaultSug[0]) {
		onlyCaseDiffers = true
	}
	// Java: word.length() >= 3 && (onlyCaseDiffers || fullResults || defaultSuggestions.isEmpty())
	// speller1 maxEdit=1 won't find "garentee", "greatful", etc. → pull speller2/3.
	wLen := UTF16Len(word)
	if wLen >= 3 && (onlyCaseDiffers || fullResults || len(defaultSug) == 0) {
		if s2 != nil {
			defaultSug = append(defaultSug, s2.GetSuggestionsFromDefaultDicts(word)...)
			userSug = append(userSug, s2.GetSuggestionsFromUserDicts(word)...)
		}
		// Java: word.length() >= 5 && (fullResults || defaultSuggestions.isEmpty())
		if wLen >= 5 && (fullResults || len(defaultSug) == 0) {
			if s3 != nil {
				defaultSug = append(defaultSug, s3.GetSuggestionsFromDefaultDicts(word)...)
				userSug = append(userSug, s3.GetSuggestionsFromUserDicts(word)...)
			}
		}
	}

	// Java: if no default/user sugs and word contains "-", addHyphenSuggestions first into top.
	var top []string
	if len(defaultSug) == 0 && len(userSug) == 0 && strings.Contains(word, "-") && r.AddHyphenSuggestionsFn != nil {
		// Java word.split("-") keeps empty segments for leading/trailing/double hyphens.
		parts := strings.Split(word, "-")
		top = append(top, r.AddHyphenSuggestionsFn(parts)...)
	}
	// Java: topSuggestions.addAll(getAdditionalTopSuggestions(defaultSuggestions, word))
	if r.GetAdditionalTopSuggestionsFn != nil {
		if langTop := r.GetAdditionalTopSuggestionsFn(defaultSug, word); len(langTop) > 0 {
			top = append(top, langTop...)
		} else if baseTop := spelling.AdditionalTopSuggestions(defaultSug, word); len(baseTop) > 0 {
			// EN returns early when curated non-empty; only use base when language top empty.
			top = append(top, baseTop...)
		}
	} else if baseTop := spelling.AdditionalTopSuggestions(defaultSug, word); len(baseTop) > 0 {
		top = append(top, baseTop...)
	}
	if len(top) > 0 {
		defaultSug = append(top, defaultSug...)
	}
	if len(defaultSug) == 0 && len(userSug) == 0 {
		return nil
	}
	// Java: defaultSuggestions = filterSuggestions(...); userSuggestions = filterDupes(...)
	if r.SpellingCheckRule != nil {
		defaultSug = r.SpellingCheckRule.FilterSuggestions(defaultSug)
	} else {
		defaultSug = filterStringDupes(defaultSug)
	}
	userSug = filterStringDupes(userSug)
	// Java orderSuggestions is identity on base rule.
	// Java: word.length()>4 → user then default; else default then user
	// (short user-dict hits usually hide best suggestions).
	var sug []string
	if wLen > 4 {
		sug = append(append([]string{}, userSug...), defaultSug...)
	} else {
		sug = append(append([]string{}, defaultSug...), userSug...)
	}
	// Final de-dupe preserving order (user may overlap default).
	return filterStringDupes(sug)
}

func filterStringDupes(in []string) []string {
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

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
