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
	// TokenizingPattern ports tokenizingPattern() (null by default).
	// When set (e.g. BR/GA "-"), Match splits the clean word on pattern matches and
	// runs getRuleMatches on each segment (Java match loop).
	TokenizingPattern *regexp.Regexp
	// UserConfig ports MorfologikSpellerRule.userConfig (suggestionsEnabled / maxSpellingSuggestions).
	UserConfig *languagetool.UserConfig
	// ForeignDetect ports ForeignLanguageChecker language-id hook (LanguageIdentifierService).
	// When nil, foreign-language scoring is inactive (Java langIdent == null → empty).
	ForeignDetect spelling.DetectScoresFunc
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

// SetUserConfig stores UserConfig for getRuleMatches suggestion gates (Java field).
func (r *MorfologikSpellerRule) SetUserConfig(uc *languagetool.UserConfig) {
	if r != nil {
		r.UserConfig = uc
	}
}

// GetSpellingSuggestions ports MorfologikSpellerRule.getSpellingSuggestions (since 5.6):
// wrap word in a one-token sentence, Match, return first match's replacements.
func (r *MorfologikSpellerRule) GetSpellingSuggestions(w string) []string {
	if r == nil || w == "" {
		return nil
	}
	at := languagetool.NewAnalyzedToken(w, nil, nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(at, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{atr})
	ms, err := r.Match(sent)
	if err != nil || len(ms) == 0 || ms[0] == nil {
		return nil
	}
	return ms[0].GetSuggestedReplacements()
}

// suggestionsEnabled ports userConfig == null || userConfig.isSuggestionsEnabled().
func (r *MorfologikSpellerRule) suggestionsEnabled() bool {
	if r == nil || r.UserConfig == nil {
		return true
	}
	return r.UserConfig.IsSuggestionsEnabled()
}

// allowMoreSpellingSuggestions ports:
// userConfig == null || maxSpellingSuggestions == 0 || ruleMatchesSoFar.size() <= max.
func (r *MorfologikSpellerRule) allowMoreSpellingSuggestions(ruleMatchesSoFar int) bool {
	if r == nil || r.UserConfig == nil {
		return true
	}
	max := r.UserConfig.GetMaxSpellingSuggestions()
	if max == 0 {
		return true
	}
	return ruleMatchesSoFar <= max
}

// Match flags misspelled tokens.
// Ports MorfologikSpellerRule.match including isFirstWord capitalization of suggestions
// and ForeignLanguageChecker newLanguageMatches when preferredLanguages ≥ 2.
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
	// Java: boolean isFirstWord = true; capitalize lower-case sugs on first real word.
	isFirstWord := true
	// Java: ForeignLanguageChecker when userConfig preferredLanguages size >= 2.
	var foreignChecker *spelling.ForeignLanguageChecker
	gotForeignResults := false
	if r.UserConfig != nil {
		pref := r.UserConfig.GetPreferredLanguages()
		// Java: !getPreferredLanguages().isEmpty() && size() >= 2
		// Go empty config splits to [""] (len 1) — not ≥ 2.
		if preferredLanguagesActive(pref) {
			// sentenceLength = non-word-filtered tokens count - 1 (Java stream).
			sentenceLength := foreignSentenceLength(sentence)
			langCode := ""
			if r.SpellingCheckRule != nil {
				langCode = r.SpellingCheckRule.LanguageCode
			}
			// Prefer short code (en from en-US) like language.getShortCode().
			if i := strings.IndexAny(langCode, "-_"); i > 0 {
				langCode = langCode[:i]
			}
			text := ""
			if sentence != nil {
				text = sentence.GetText()
			}
			foreignChecker = spelling.NewForeignLanguageChecker(langCode, text, sentenceLength, pref)
			// Optional Detect hook: set ForeignDetect on the rule when identifier is wired.
			if r.ForeignDetect != nil {
				foreignChecker.Detect = r.ForeignDetect
			}
		}
	}
	// applyForeign ports end-of-loop foreignLanguageChecker.check (once until result).
	applyForeign := func() {
		if foreignChecker == nil || gotForeignResults {
			return
		}
		scores := foreignChecker.Check(len(out))
		if len(scores) == 0 {
			return
		}
		if _, noForeign := scores[spelling.NoForeignLangDetected]; !noForeign && len(out) > 0 && out[0] != nil {
			out[0].SetNewLanguageMatches(scores)
		}
		gotForeignResults = true
	}

	for idx, tok := range tokens {
		// Java canBeIgnored continue arm also clears isFirstWord when idx>0 and non-punct.
		// Foreign checker is not invoked on canBeIgnored continues (Java continue).
		if spelling.CanBeIgnoredToken(tok) {
			maybeClearIsFirstWord(&isFirstWord, idx, tok)
			continue
		}
		// Java language getRuleMatches early exit (e.g. _english_ignore_).
		if r.SkipTokenFn != nil && r.SkipTokenFn(tok) {
			maybeClearIsFirstWord(&isFirstWord, idx, tok)
			continue
		}
		// Java canBeIgnored: ignoreToken(tokens, idx) → ignoreWord.
		if r.SpellingCheckRule != nil && r.IgnoreToken(tokens, idx) {
			maybeClearIsFirstWord(&isFirstWord, idx, tok)
			continue
		}
		// Java: word = token.getAnalyzedToken(0).getToken() — without ignored chars
		// (soft hyphen etc.). ATR.GetToken() keeps orig surface; GetCleanToken is clean.
		// When cleanToken metadata is set (replaceSoftHyphens), use it so the speller
		// does not choke on U+00AD. Fallback GetCleanToken() == GetToken() when unset.
		surface := tok.GetToken()
		w := spellCheckWord(tok)
		if w == "" || !hasLetter(w) {
			maybeClearIsFirstWord(&isFirstWord, idx, tok)
			continue
		}
		// Java MorfologikSpellerRule.ignoreWord: super.ignoreWord || StringTools.isEmoji
		// Emoji check uses surface token (Java ignoreWord(String word) on rule path).
		if tools.IsEmoji(w) || tools.IsEmoji(surface) {
			maybeClearIsFirstWord(&isFirstWord, idx, tok)
			continue
		}
		if r.IgnoreTaggedWords && tok.IsTagged() {
			// Java: ignoreTaggedWords && isTagged && !isProhibited
			if r.SpellingCheckRule == nil || !r.IsProhibited(w) {
				maybeClearIsFirstWord(&isFirstWord, idx, tok)
				continue
			}
		}
		startPos := tok.GetStartPos()
		// Java: previous match already covers this token → getRuleMatches returns empty,
		// but the loop still runs first-word clear + foreign check.
		if len(out) > 0 && out[len(out)-1] != nil && out[len(out)-1].GetToPos() > startPos {
			clearIsFirstWordAfterToken(&isFirstWord, idx, tok)
			applyForeign()
			continue
		}

		newRuleIdx := len(out)

		// Java: Pattern pattern = tokenizingPattern(); split word and getRuleMatches each segment.
		if r.TokenizingPattern != nil {
			segs := tokenizingSegments(w, r.TokenizingPattern)
			if len(segs) > 1 {
				for _, seg := range segs {
					if seg.word == "" {
						continue
					}
					// getRuleMatches per segment (includes isMisspelled / ignore gates).
					r.appendGetRuleMatches(sentence, tokens, idx, tok, surface, w, seg.word, startPos+seg.utf16Off, &out)
				}
				adjustHiddenCharOffsets(tok, surface, w, out, newRuleIdx)
				capitalizeFirstWordSuggestions(isFirstWord, idx, tokens, out)
				clearIsFirstWordAfterToken(&isFirstWord, idx, tok)
				applyForeign()
				continue
			}
		}

		// No tokenizing pattern (or pattern not found): Java getRuleMatches(word, startPos).
		// getRuleMatches early-returns when not misspelled; foreign still runs after.
		if r.AcceptWord(w) {
			clearIsFirstWordAfterToken(&isFirstWord, idx, tok)
			applyForeign()
			continue
		}
		if r.SpellingCheckRule != nil && r.IgnorePotentiallyMisspelledWord(w) {
			// Java getRuleMatches: ignorePotentiallyMisspelledWord → empty; foreign still runs.
			clearIsFirstWordAfterToken(&isFirstWord, idx, tok)
			applyForeign()
			continue
		}
		r.appendGetRuleMatches(sentence, tokens, idx, tok, surface, w, w, startPos, &out)
		adjustHiddenCharOffsets(tok, surface, w, out, newRuleIdx)
		capitalizeFirstWordSuggestions(isFirstWord, idx, tokens, out)
		clearIsFirstWordAfterToken(&isFirstWord, idx, tok)
		applyForeign()
	}
	return out, nil
}

// preferredLanguagesActive ports Java preferredLanguages size >= 2 with non-empty entries.
func preferredLanguagesActive(pref []string) bool {
	n := 0
	for _, p := range pref {
		if strings.TrimSpace(p) != "" {
			n++
		}
	}
	return n >= 2
}

// foreignSentenceLength ports match's sentenceLength:
// count of non-isNonWord tokens without whitespace, minus 1.
func foreignSentenceLength(sentence *languagetool.AnalyzedSentence) int64 {
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

// tokSeg is one segment from tokenizingPattern split (Java match loop).
type tokSeg struct {
	word     string
	utf16Off int // UTF-16 offset of segment start within parent word
}

// tokenizingSegments ports match's tokenizingPattern split of word.
// Pattern null / no match → single whole-word segment.
func tokenizingSegments(word string, pat *regexp.Regexp) []tokSeg {
	if pat == nil || word == "" {
		return []tokSeg{{word: word, utf16Off: 0}}
	}
	matches := pat.FindAllStringIndex(word, -1)
	if len(matches) == 0 {
		// Java: index == 0 after loop → whole word
		return []tokSeg{{word: word, utf16Off: 0}}
	}
	var segs []tokSeg
	index := 0 // byte index in word
	for _, m := range matches {
		// Java: word.subSequence(index, m.start())
		if m[0] > index {
			part := word[index:m[0]]
			segs = append(segs, tokSeg{word: part, utf16Off: byteIndexToUTF16(word, index)})
		}
		index = m[1]
	}
	// Java: remainder from index to end
	if index < len(word) {
		segs = append(segs, tokSeg{word: word[index:], utf16Off: byteIndexToUTF16(word, index)})
	}
	if len(segs) == 0 {
		return []tokSeg{{word: word, utf16Off: 0}}
	}
	return segs
}

// byteIndexToUTF16 converts a UTF-8 byte index into a Java String UTF-16 length prefix.
func byteIndexToUTF16(s string, byteIdx int) int {
	if byteIdx <= 0 {
		return 0
	}
	if byteIdx >= len(s) {
		return utf16LenMF(s)
	}
	return utf16LenMF(s[:byteIdx])
}

// appendGetRuleMatches ports getRuleMatches for one word/segment at startPos.
// wholeWord is the token clean surface (for hidden-char / multi-token gates);
// word is the segment under check; startPos is absolute UTF-16 start of the segment.
//
// Control flow matches Java MorfologikSpellerRule.getRuleMatches:
//  1. misspell / prohibit / ignorePotentially / covered-by-prev gates
//  2. wrong-split with prev (early return only when prev is also misspelled)
//  3. wrong-split with next when no match yet (early when next is misspelled)
//  4. create default match if still null
//  5. suggestionsEnabled gate
//  6. numbers/bullets prefix (may rewrite beforeStr / cleanWord / preventFurther)
//  7. maxSpellingSuggestions + joinBeforeAfterSuggestions(calcSpellerSuggestions)
func (r *MorfologikSpellerRule) appendGetRuleMatches(
	sentence *languagetool.AnalyzedSentence,
	tokens []*languagetool.AnalyzedTokenReadings,
	idx int,
	tok *languagetool.AnalyzedTokenReadings,
	surface, wholeWord, word string,
	startPos int,
	out *[]*rules.RuleMatch,
) {
	if r == nil || word == "" || out == nil {
		return
	}
	// Java getRuleMatches: if (!isMisspelled(speller1, word) && !isProhibited(word)) return
	if !r.IsProhibited(word) && !r.segmentIsMisspelled(word) {
		return
	}
	if r.SpellingCheckRule != nil && r.IgnorePotentiallyMisspelledWord(word) {
		return
	}
	// Covered by previous match (wrong-split / longer span)
	if len(*out) > 0 && (*out)[len(*out)-1] != nil && (*out)[len(*out)-1].GetToPos() > startPos {
		return
	}

	// Wrong-split uses neighboring sentence tokens; Java still runs it on segments.
	beforeStr := ""
	afterStr := ""
	ruleMatch, beforeStr, early := r.tryWrongSplitPrev(sentence, out, idx, tokens, word, startPos)
	if early && ruleMatch != nil {
		// Java: ruleMatches.add(ruleMatch); return (no further dict sugs).
		*out = append(*out, ruleMatch)
		return
	}
	// Java: if (ruleMatch == null && idx < tokens.length - 1 && ...)
	if ruleMatch == nil {
		var after string
		ruleMatch, after, early = r.tryWrongSplitNext(sentence, out, idx, tokens, word, startPos)
		afterStr = after
		if early && ruleMatch != nil {
			*out = append(*out, ruleMatch)
			return
		}
	}

	// Java: if (ruleMatch == null) { new RuleMatch(... UnknownWord); }
	if ruleMatch == nil {
		toPos := startPos + utf16LenMF(word)
		if word == wholeWord && tok != nil && tok.GetEndPos() > toPos {
			toPos = tok.GetEndPos()
		}
		ruleMatch = rules.NewRuleMatch(r, sentence, startPos, toPos, "Possible spelling mistake found")
		ruleMatch.SetType(rules.RuleMatchTypeUnknownWord)
	}

	// Java: if (userConfig != null && !userConfig.isSuggestionsEnabled()) { add; return; }
	if !r.suggestionsEnabled() {
		*out = append(*out, ruleMatch)
		return
	}

	// Numbers/bullets may rewrite beforeStr / cleanWord (Java assigns beforeSuggestionStr).
	cleanWord, beforePrefix, preventFurther, spaceInsert := r.applyNumbersBulletsPrefix(word)
	if beforePrefix != "" {
		// Java: beforeSuggestionStr = firstPart + " "; (replaces prior before when set)
		beforeStr = beforePrefix
	}
	if spaceInsert != "" {
		addSug(ruleMatch, spaceInsert)
	}

	// Java: maxSpellingSuggestions gate around appendLazySuggestions
	if r.allowMoreSpellingSuggestions(len(*out)) {
		if !preventFurther {
			// appendLazySuggestions: prev (already on match) + joinBeforeAfter(calcSpellerSuggestions)
			for _, s := range r.collectSuggestions(cleanWord) {
				// Java joinBeforeAfterSuggestions: before + str + after (no trim)
				addSug(ruleMatch, joinBeforeAfterSuggestion(s, beforeStr, afterStr))
			}
		}
	} else {
		// messages.getString("too_many_errors")
		ruleMatch.SetSuggestedReplacement(tooManyErrorsMsg)
	}
	*out = append(*out, ruleMatch)
}

// joinBeforeAfterSuggestion ports joinBeforeAfterSuggestions for one replacement:
// beforeSuggestionStr + str + afterSuggestionStr (Java SuggestedReplacement copy).
func joinBeforeAfterSuggestion(str, before, after string) string {
	return before + str + after
}

// tooManyErrorsMsg ports MessagesBundle too_many_errors (en default).
const tooManyErrorsMsg = "(suggestion limit reached)"

// spellCheckWord ports match's word = token.getAnalyzedToken(0).getToken() intent:
// surface without ignored characters so the speller does not choke (soft hyphen).
// Prefers GetCleanToken when set by replaceSoftHyphens; else ATR surface / AT0.
func spellCheckWord(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	// Clean token from soft-hyphen metadata (always preferred when present).
	if c := tok.GetCleanToken(); c != "" && c != tok.GetToken() {
		return c
	}
	// Java getAnalyzedToken(0).getToken() when readings keep the cleaned form.
	if n := len(tok.GetReadings()); n > 0 {
		if at := tok.GetAnalyzedToken(0); at != nil {
			if t := at.GetToken(); t != "" {
				// Prefer reading shorter than dirty surface (clean form left as AT0
				// when POS readings were kept and dirty was only added via addReading).
				if utf16LenMF(t) < utf16LenMF(tok.GetToken()) {
					return t
				}
			}
		}
	}
	return tok.GetToken()
}

// adjustHiddenCharOffsets ports match's hiddenCharOffset adjustment:
// token.getToken().length() - word.length() added to toPos when the match
// does not already extend past the token (multi-token wrong-split).
func adjustHiddenCharOffsets(tok *languagetool.AnalyzedTokenReadings, surface, word string, out []*rules.RuleMatch, newRuleIdx int) {
	if tok == nil || len(out) == 0 {
		return
	}
	// Java String.length() = UTF-16 units
	offset := utf16LenMF(surface) - utf16LenMF(word)
	if offset <= 0 {
		return
	}
	endPos := tok.GetEndPos()
	for i := newRuleIdx; i < len(out); i++ {
		m := out[i]
		if m == nil {
			continue
		}
		// Java: if (token.getEndPos() < ruleMatch.getToPos()) multi-token — skip
		if endPos < m.GetToPos() {
			continue
		}
		// setOffsetPosition(from, to+hiddenCharOffset)
		// When match was built with GetEndPos already, to already covers surface —
		// only extend when to was based on clean word length (to < endPos).
		if m.GetToPos() < endPos {
			// Prefer clamping to token end rather than overshooting.
			// Java adds offset to clean-based toPos → equals surface end when
			// toPos was start+cleanLen and endPos is start+surfaceLen.
			m.SetOffsetPosition(m.GetFromPos(), m.GetToPos()+offset)
		}
	}
}

// maybeClearIsFirstWord ports canBeIgnored branch:
// if (idx > 0 && isFirstWord && !StringTools.isPunctuationMark(token.getToken())) isFirstWord=false.
func maybeClearIsFirstWord(isFirstWord *bool, idx int, tok *languagetool.AnalyzedTokenReadings) {
	if isFirstWord == nil || !*isFirstWord {
		return
	}
	if idx > 0 && tok != nil && !tools.IsPunctuationMark(tok.GetToken()) {
		*isFirstWord = false
	}
}

// clearIsFirstWordAfterToken ports end-of-loop body isFirstWord clear (non-ignored tokens).
func clearIsFirstWordAfterToken(isFirstWord *bool, idx int, tok *languagetool.AnalyzedTokenReadings) {
	maybeClearIsFirstWord(isFirstWord, idx, tok)
}

// capitalizeFirstWordSuggestions ports match's first-word capitalization of suggestions:
// if isFirstWord && ruleMatches.nonEmpty && idx < tokens.length-1, uppercaseFirstChar
// each all-lower replacement on ruleMatches.get(0).
func capitalizeFirstWordSuggestions(isFirstWord bool, idx int, tokens []*languagetool.AnalyzedTokenReadings, out []*rules.RuleMatch) {
	if !isFirstWord || len(out) == 0 || idx >= len(tokens)-1 {
		return
	}
	// Java: RuleMatch ruleMatch = ruleMatches.get(0);
	m := out[0]
	if m == nil {
		return
	}
	sugs := m.GetSuggestedReplacements()
	if len(sugs) == 0 {
		return
	}
	var newSugs []string
	seen := map[string]struct{}{}
	for _, replacement := range sugs {
		// only if the replacement is all lower case
		var next string
		if replacement == strings.ToLower(replacement) {
			next = tools.UppercaseFirstChar(replacement)
		} else {
			next = replacement
		}
		if _, ok := seen[next]; ok {
			continue
		}
		seen[next] = struct{}{}
		newSugs = append(newSugs, next)
	}
	m.SetSuggestedReplacements(newSugs)
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

// segmentIsMisspelled is getRuleMatches' isMisspelled check for a word/segment.
// Prefer Multi/Speller compound path; when no dict is loaded, fall back to
// AcceptWord invert so tests/hooks that set SpellingCheckRule.IsMisspelled still work
// (without recursing through the default IsMisspelled → isMisspelledWord wrapper).
func (r *MorfologikSpellerRule) segmentIsMisspelled(word string) bool {
	if r == nil || word == "" {
		return false
	}
	hasMulti := r.Multi != nil && len(r.Multi.Spellers) > 0
	hasDict := r.Speller != nil && r.Speller.HasDictionary()
	if hasMulti || hasDict {
		return r.isMisspelledWord(word)
	}
	// No dict: AcceptWord encodes IsMisspelled hook + ignore/prohibit
	return !r.AcceptWord(word)
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
