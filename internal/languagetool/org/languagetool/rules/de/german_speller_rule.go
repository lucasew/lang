package de

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanSpellerRule ports org.languagetool.rules.de.GermanSpellerRule at a partial
// faithfulness level: ID variants, language-specific ignore lists, prohibit lists
// (exact + prefix.* + .*suffix), and isMisspelled via WireGermanFilterSpeller
// (de/hunspell/de_DE.dict).
//
// Compound accept (processTwoPart/ThreePart) and suggestion stack are ported;
// TagPOS/LemmaOf/Synthesize need resources (added.txt wired; german.dict optional).
// Multi-token ignore phrases applied in Match/IgnoreWordAt.
// Match: token walk + wrong-split + UnknownWord type + high-confidence objects.
type GermanSpellerRule struct {
	Messages map[string]string
	// Category ports MorfologikSpellerRule/SpellingCheckRule (TYPOS).
	Category *rules.Category
	// IssueType ports setLocQualityIssueType(Misspelling).
	IssueType rules.ITSIssueType
	// LanguageVariant: "", "AT", "CH"
	LanguageVariant string
	// LanguageSpecific is optional plain-text spelling resource (Java languageSpecificPlainTextDict).
	LanguageSpecific string
	// IgnoreWords ports wordsToBeIgnored from spelling extras / addIgnoreWords.
	IgnoreWords map[string]struct{}
	// IgnorePhrases ports multi-token ignore lines (Java DisambiguationPatternRule
	// IGNORE_SPELLING antipatterns from SpellingCheckRule.addIgnoreWords).
	// Each phrase is a sequence of case-sensitive tokens.
	IgnorePhrases [][]string
	// IgnoredInCompounds ports wordsToBeIgnoredInCompounds (lines ending "-*").
	IgnoredInCompounds map[string]struct{}
	// Prohibited ports wordsToBeProhibited (exact forms after LineExpander).
	Prohibited map[string]struct{}
	// ProhibitedStarts ports GermanSpellerRule.wordStartsToBeProhibited (lines ending ".*").
	ProhibitedStarts map[string]struct{}
	// ProhibitedEnds ports GermanSpellerRule.wordEndingsToBeProhibited (lines starting ".*").
	ProhibitedEnds map[string]struct{}
	// IsMisspelledOverride optional; when set, used instead of FilterDict (tests).
	IsMisspelledOverride func(word string) bool
	// TagPOS optional POS tags for a surface form (GermanTagger.tag).
	// Used by non-hyphenated ignoreCompoundWithIgnoredWord (isNoun/isAdjective).
	// Nil → those POS checks fail-closed (do not invent noun/adj readings).
	TagPOS func(word string) []string
	// CompoundTokenize optional GermanCompoundTokenizer.tokenize for hanging-hyphen
	// isCompound (tokenize size > 1). Nil → only hyphen / SPECIAL_CASE_THIRD arms.
	CompoundTokenize func(word string) []string
	// WordsNeedingInfixS ports wordsNeedingInfixS (de/words_infix_s.txt).
	WordsNeedingInfixS map[string]struct{}
	// VerbStems ports verbStems (de/verb_stems.txt).
	VerbStems map[string]struct{}
	// OtherPrefixes ports otherPrefixes (de/other_prefixes.txt).
	OtherPrefixes map[string]struct{}
	// VerbPrefixes ports verbPrefixes (de/verb_prefixes.txt).
	VerbPrefixes map[string]struct{}
	// OldSpelling ports oldSpelling set from alt_neu.csv (old orthography forms).
	OldSpelling map[string]struct{}
	// LemmaOf optional noun/verb lemma for surface (findLemmaForNoun / baseForThirdPerson).
	// Nil → empty lemma (fail-closed).
	LemmaOf func(word string) string
	// Synthesize optional GermanSynthesizer.synthesize(token, postagRE, true).
	// Nil → past-tense/participle suggestion paths inactive.
	Synthesize func(lemma, postagRE string) []string
	// CompoundTokenizeNonStrict optional non-strict GermanCompoundTokenizer when
	// CompoundTokenize returns a single part (Java nonStrictCompoundTokenizer fallback).
	CompoundTokenizeNonStrict func(word string) []string
	// CompoundTokenizeAll optional jWordSplitter.getAllSplits stand-in: every
	// dictionary partition of word. Nil → GetCandidates uses single Tokenize only.
	CompoundTokenizeAll func(word string) [][]string
}

// GermanSpellingFile is Java SpellingCheckRule.getSpellingFileName for de:
// language.shortCode + "/hunspell/spelling.txt".
const GermanSpellingFile = "de/hunspell/spelling.txt"

// GermanSpellingFileResource is the resource-dir form of the base spelling extras.
const GermanSpellingFileResource = "/de/hunspell/spelling.txt"

// GermanIgnoreFile is Java SpellingCheckRule.getIgnoreFileName for de:
// language.shortCode + "/hunspell/ignore.txt".
const GermanIgnoreFile = "de/hunspell/ignore.txt"

// GermanIgnoreFileResource is the resource-dir form of ignore.txt.
const GermanIgnoreFileResource = "/de/hunspell/ignore.txt"

// GermanProhibitFile is Java SpellingCheckRule.getProhibitFileName for de:
// language.shortCode + "/hunspell/prohibit.txt".
const GermanProhibitFile = "de/hunspell/prohibit.txt"

// GermanProhibitFileResource is the resource-dir form of prohibit.txt.
const GermanProhibitFileResource = "/de/hunspell/prohibit.txt"

func NewGermanSpellerRule(messages map[string]string) *GermanSpellerRule {
	// Java MorfologikSpellerRule: TYPOS + Misspelling (GermanSpellerRule extends it).
	return &GermanSpellerRule{
		Messages:           messages,
		Category:           rules.CatTypos.GetCategory(messages),
		IssueType:          rules.ITSMisspelling,
		IgnoreWords:        map[string]struct{}{},
		IgnoredInCompounds: map[string]struct{}{},
		Prohibited:         map[string]struct{}{},
		ProhibitedStarts:   map[string]struct{}{},
		ProhibitedEnds:     map[string]struct{}{},
		WordsNeedingInfixS: map[string]struct{}{},
		VerbStems:          map[string]struct{}{},
		OtherPrefixes:      map[string]struct{}{},
		VerbPrefixes:       map[string]struct{}{},
		OldSpelling:        map[string]struct{}{},
	}
}

func (r *GermanSpellerRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *GermanSpellerRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

func (r *GermanSpellerRule) GetID() string {
	if r == nil {
		return "GERMAN_SPELLER_RULE"
	}
	switch r.LanguageVariant {
	case "AT":
		return "AUSTRIAN_GERMAN_SPELLER_RULE"
	case "CH":
		return "SWISS_GERMAN_SPELLER_RULE"
	default:
		return "GERMAN_SPELLER_RULE"
	}
}

// spellingShortMessage ports messages.getString("desc_spelling_short").
func (r *GermanSpellerRule) spellingShortMessage() string {
	if r != nil && r.Messages != nil {
		if s, ok := r.Messages["desc_spelling_short"]; ok && s != "" {
			return s
		}
		if s, ok := r.Messages["spelling_short"]; ok && s != "" {
			return s
		}
	}
	return "Rechtschreibfehler"
}

// GetMessage ports GermanSpellerRule.getMessage (ss→ß pedagogy) with a plain
// fallback when Messages has no "spelling" key.
func (r *GermanSpellerRule) GetMessage(word, suggestion string) string {
	if suggestion != "" {
		// Java StringUtils.replaceOnce(origWord, "ss", "ß").equals(suggestion)
		if replacedOnceSS(word) == suggestion {
			if idx := strings.Index(word, "ss"); idx >= 2 {
				prefix := []rune(word[:idx])
				if len(prefix) >= 2 {
					prevPrev := prefix[len(prefix)-2]
					prev := prefix[len(prefix)-1]
					if IsVowel(prevPrev) && IsVowel(prev) {
						return "Nach einer Silbe aus zwei Vokalen (hier: " + string(prevPrev) + string(prev) +
							") schreibt man 'ß' statt 'ss'."
					}
					return "Nach einer lang gesprochenen Silbe (hier: " + string(prev) +
						") schreibt man 'ß' statt 'ss'."
				}
			}
		}
	}
	if r != nil && r.Messages != nil {
		if s, ok := r.Messages["spelling"]; ok && s != "" {
			return s
		}
	}
	if suggestion == "" {
		return "Möglicher Rechtschreibfehler: " + word
	}
	return "Möglicher Rechtschreibfehler: " + word + " → " + suggestion
}

// replacedOnceSS ports StringUtils.replaceOnce(s, "ss", "ß").
func replacedOnceSS(s string) string {
	i := strings.Index(s, "ss")
	if i < 0 {
		return s
	}
	return s[:i] + "ß" + s[i+2:]
}

// Match ports HunspellRule.match for GermanSpellerRule:
// token walk with IgnoreWordAt / IgnoreWord / isProhibited / isMisspelled /
// ignorePotentiallyMisspelledWord + wrong-split + Suggest + Type.UnknownWord.
// Without a wired FilterDict (Java hunspell == null) returns empty (silent).
func (r *GermanSpellerRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	if !FilterDictAvailable() {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	if len(tokens) == 0 {
		return nil
	}
	words := make([]string, len(tokens))
	for i, tok := range tokens {
		if tok != nil {
			words[i] = tok.GetToken()
		}
	}
	var out []*rules.RuleMatch
	prevWord := ""
	prevFrom := -1
	for i, tok := range tokens {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		word := tok.GetToken()
		// Last content token may be marked sentence-end by AnalyzePlain; still spell-check it.
		if word == "" {
			continue
		}
		// URL / email / immunized / quoted compound / _english_ignore_
		if r.shouldSkipSpellToken(sentence, tokens, i) {
			prevWord = word
			prevFrom = tok.GetStartPos()
			continue
		}
		// Java: (ignoreWord(list,i) || ignoreWord(word)) && !isProhibited(cutOffDot(word))
		if (r.IgnoreWordAt(words, i) || r.IgnoreWord(word)) && !r.IsProhibited(cutOffDot(word)) {
			// ignored tokens still advance prev for wrong-split context? Java continues without updating prev for misspelled path
			// only misspelled path updates prevStartPos after processing; ignored skip updates prevStartPos = len
			prevWord = word
			prevFrom = tok.GetStartPos()
			continue
		}
		if !r.IsMisspelled(word) {
			prevWord = word
			prevFrom = tok.GetStartPos()
			continue
		}
		if r.IgnorePotentiallyMisspelledWord(word) {
			prevWord = word
			prevFrom = tok.GetStartPos()
			continue
		}
		cleanWord := word
		if strings.HasSuffix(word, ".") {
			cleanWord = word[:len(word)-1]
		}
		dashCorr := 0
		if strings.HasPrefix(word, "-") {
			rest := cleanWord
			if len(rest) > 0 {
				rest = rest[1:]
			}
			if !r.IsMisspelled(rest) {
				prevWord = word
				prevFrom = tok.GetStartPos()
				continue
			}
			dashCorr = 1
		}
		// wrong-split with previous token
		if prevFrom >= 0 && prevWord != "" {
			if ws := r.tryWrongSplitSuggestions(sentence, prevWord, prevFrom, word, tok.GetStartPos(), cleanWord); ws != nil {
				out = removeLastWrongSplitIfSamePrev(out, prevFrom)
				out = append(out, ws)
				prevWord = word
				prevFrom = tok.GetStartPos()
				continue
			}
		}
		from := tok.GetStartPos() + dashCorr
		to := tok.GetStartPos() + utf16LenDE(cleanWord)
		if to < from {
			to = tok.GetEndPos()
		}
		surf := cleanWord
		if dashCorr > 0 && len(cleanWord) > dashCorr {
			surf = cleanWord[dashCorr:]
		}
		sugs := r.Suggest(surf)
		msg := r.GetMessage(surf, "")
		if len(sugs) > 0 {
			msg = r.GetMessage(surf, sugs[0])
		}
		m := rules.NewRuleMatch(r, sentence, from, to, msg)
		// HunspellRule.match: Type.UnknownWord + messages "desc_spelling_short"
		m.SetType(rules.RuleMatchTypeUnknownWord)
		m.SetShortMessage(r.spellingShortMessage())
		if len(sugs) > 0 {
			objs := rules.ConvertSuggestions(sugs)
			// HunspellRule: isFirstItemHighConfidenceSuggestion → HIGH_CONFIDENCE 0.99
			if r.isFirstItemHighConfidenceSuggestion(surf, sugs) && len(objs) > 0 && objs[0] != nil {
				c := rules.SpellingHighConfidence
				objs[0].SetConfidence(&c)
			}
			m.SetSuggestedReplacementObjects(objs)
		}
		out = append(out, m)
		prevWord = word
		prevFrom = tok.GetStartPos()
	}
	return r.removeGenderCompoundMatches(sentence, out)
}

// utf16LenDE ports Java String.length for match spans (UTF-16 code units).
func utf16LenDE(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// AddIgnoreWords ports SpellingCheckRule.addIgnoreWords for single-token lines.
// Multi-token strings (spaces) are routed to AddIgnorePhrase (Java tokenize → antipattern).
func (r *GermanSpellerRule) AddIgnoreWords(words ...string) {
	if r == nil {
		return
	}
	if r.IgnoreWords == nil {
		r.IgnoreWords = map[string]struct{}{}
	}
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		if strings.Contains(w, " ") {
			parts := strings.Fields(w)
			if len(parts) > 1 {
				r.AddIgnorePhrase(parts...)
			} else if len(parts) == 1 {
				r.IgnoreWords[parts[0]] = struct{}{}
			}
			continue
		}
		r.IgnoreWords[w] = struct{}{}
	}
}

// AddIgnorePhrase ports multi-token spelling extras as case-sensitive phrase ignore
// (Java PatternToken case-sensitive, non-inflected + IGNORE_SPELLING).
func (r *GermanSpellerRule) AddIgnorePhrase(tokens ...string) {
	if r == nil || len(tokens) == 0 {
		return
	}
	var phrase []string
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		phrase = append(phrase, t)
	}
	if len(phrase) == 0 {
		return
	}
	if len(phrase) == 1 {
		if r.IgnoreWords == nil {
			r.IgnoreWords = map[string]struct{}{}
		}
		r.IgnoreWords[phrase[0]] = struct{}{}
		return
	}
	for _, existing := range r.IgnorePhrases {
		if len(existing) != len(phrase) {
			continue
		}
		same := true
		for i := range phrase {
			if existing[i] != phrase[i] {
				same = false
				break
			}
		}
		if same {
			return
		}
	}
	r.IgnorePhrases = append(r.IgnorePhrases, append([]string(nil), phrase...))
}

// isInIgnorePhrase reports whether words[idx] is covered by a multi-token ignore phrase
// (any phrase that fully matches a window containing idx). Case-sensitive token equality.
func (r *GermanSpellerRule) isInIgnorePhrase(words []string, idx int) bool {
	if r == nil || idx < 0 || idx >= len(words) || len(r.IgnorePhrases) == 0 {
		return false
	}
	for _, phrase := range r.IgnorePhrases {
		n := len(phrase)
		if n < 2 {
			continue
		}
		minStart := idx - n + 1
		if minStart < 0 {
			minStart = 0
		}
		for start := minStart; start <= idx; start++ {
			if start+n > len(words) {
				break
			}
			match := true
			for j := 0; j < n; j++ {
				if words[start+j] != phrase[j] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

// GermanIgnoreWordsWithLength ports GermanSpellerRule.init() setting
// super.ignoreWordsWithLength = 1 (accept single-letter tokens).
const GermanIgnoreWordsWithLength = 1

// GermanSpellerMaxTokenLength ports SpellingCheckRule.MAX_TOKEN_LENGTH / German MAX_TOKEN_LENGTH.
const GermanSpellerMaxTokenLength = 200

// isIgnoredNoCase ports GermanSpellerRule.isIgnoredNoCase (override of SpellingCheckRule):
// exact ignore set, FirstUpper→lower, or length ≤ ignoreWordsWithLength.
func (r *GermanSpellerRule) isIgnoredNoCase(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if _, ok := r.IgnoreWords[word]; ok {
		return true
	}
	// Java FIRST_UPPER_CASE && wordsToBeIgnored.contains(word.toLowerCase(locale))
	if isFirstUpper(word) {
		if _, ok := r.IgnoreWords[strings.ToLower(word)]; ok {
			return true
		}
	}
	if utf16LenDE(word) <= GermanIgnoreWordsWithLength {
		return true
	}
	return false
}

// isIgnored is the plain set membership path (no length-1 special); kept for call sites
// that only mean wordsToBeIgnored. Prefer IgnoreWord for SpellingCheckRule.ignoreWord.
func (r *GermanSpellerRule) isIgnored(word string) bool {
	if r == nil || word == "" || len(r.IgnoreWords) == 0 {
		return false
	}
	if _, ok := r.IgnoreWords[word]; ok {
		return true
	}
	if isFirstUpper(word) {
		if _, ok := r.IgnoreWords[strings.ToLower(word)]; ok {
			return true
		}
	}
	return false
}

func isFirstUpper(word string) bool {
	if word == "" {
		return false
	}
	r0 := []rune(word)[0]
	return unicode.IsUpper(r0)
}

func startsWithUppercase(word string) bool {
	return isFirstUpper(word)
}

// hasNoLatinLetter ports SpellingCheckRule pHasNoLetterLatin: true if no Latin-script letters.
func hasNoLatinLetter(word string) bool {
	for _, r := range word {
		if unicode.In(r, unicode.Latin) {
			return false
		}
	}
	return true
}

// IgnoreWord ports SpellingCheckRule.ignoreWord(String) for German (Latin script):
// max length, no-letter tokens, trailing ".", isIgnoredNoCase (incl. length-1).
// Does not include German list-form extras (hyphen compounds, bullet points).
func (r *GermanSpellerRule) IgnoreWord(word string) bool {
	if r == nil {
		return false
	}
	if utf16LenDE(word) > GermanSpellerMaxTokenLength {
		return true
	}
	if hasNoLatinLetter(word) {
		return true
	}
	if strings.HasSuffix(word, ".") && !r.isInIgnoredSet(word) {
		return r.isIgnoredNoCase(word[:len(word)-1])
	}
	return r.isIgnoredNoCase(word)
}

func (r *GermanSpellerRule) isInIgnoredSet(word string) bool {
	if r == nil || r.IgnoreWords == nil {
		return false
	}
	_, ok := r.IgnoreWords[word]
	return ok
}

// StartsWithIgnoredWord ports SpellingCheckRule.startsWithIgnoredWord:
// longest prefix of word that is in wordsToBeIgnored; word must be ≥ 4 runes; 0 if none.
// caseSensitive=true uses exact forms (German compound path).
func (r *GermanSpellerRule) StartsWithIgnoredWord(word string, caseSensitive bool) int {
	if r == nil || len(r.IgnoreWords) == 0 {
		return 0
	}
	runes := []rune(word)
	if len(runes) < 4 {
		return 0
	}
	best := 0
	for ign := range r.IgnoreWords {
		ir := []rune(ign)
		if len(ir) < 1 || len(ir) > len(runes) || len(ir) <= best {
			continue
		}
		if caseSensitive {
			if strings.HasPrefix(word, ign) {
				best = len(ir)
			}
		} else if strings.HasPrefix(strings.ToLower(word), strings.ToLower(ign)) {
			best = len(ir)
		}
	}
	return best
}

// reDirection ports GermanSpellerRule.DIRECTION (full match nord|ost|süd|west).
var reDirection = regexp.MustCompile(`^(?:nord|ost|süd|west)$`)

// reDirectionalIsch ports partialWord.matches(".+ische?[mnrs]?") for geo compounds.
var reDirectionalIsch = regexp.MustCompile(`.+ische?[mnrs]?$`)

// isNeedingFugenS ports GermanSpellerRule.isNeedingFugenS.
func isNeedingFugenS(word string) bool {
	for _, suf := range []string{"tum", "ling", "ion", "tät", "keit", "schaft", "sicht", "ung", "en"} {
		if strings.HasSuffix(word, suf) {
			return true
		}
	}
	return false
}

func uppercaseFirstChar(word string) string {
	if word == "" {
		return word
	}
	r := []rune(word)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func lowercaseFirstChar(word string) string {
	if word == "" {
		return word
	}
	r := []rune(word)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func isAllUpperCase(word string) bool {
	if word == "" {
		return false
	}
	hasLetter := false
	for _, r := range word {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func isAllLowerCase(word string) bool {
	if word == "" {
		return false
	}
	hasLetter := false
	for _, r := range word {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter
}

// isNoun ports GermanSpellerRule.isNoun via TagPOS (SUB:…). Nil TagPOS → false (fail-closed).
func (r *GermanSpellerRule) isNoun(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "SUB:") {
			return true
		}
	}
	return false
}

// isAdjective ports GermanSpellerRule.isAdjective via TagPOS (ADJ:…). Nil TagPOS → false.
func (r *GermanSpellerRule) isAdjective(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "ADJ:") {
			return true
		}
	}
	return false
}

func spellerDictAccepts(word string) bool {
	if word == "" || !FilterDictAvailable() {
		return false
	}
	return !FilterDictIsMisspelled(word)
}

// IgnoreElative ports GermanSpellerRule.ignoreElative (prefix intensifiers + valid remainder).
func (r *GermanSpellerRule) IgnoreElative(word string) bool {
	if r == nil || word == "" {
		return false
	}
	prefixes := []string{
		"bitter", "dunkel", "erz", "extra", "früh", "gemein", "hyper", "lau",
		"mega", "minder", "stock", "super", "tod", "ultra", "ur",
	}
	okStart := false
	for _, p := range prefixes {
		if strings.HasPrefix(word, p) {
			okStart = true
			break
		}
	}
	if !okStart {
		return false
	}
	// Java RegExUtils.removePattern(..., "^(bitter|…|grund|…|voll)")
	reStrip := regexp.MustCompile(`^(?:bitter|dunkel|erz|extra|früh|gemein|grund|hyper|lau|mega|minder|stock|super|tod|ultra|ur|voll)`)
	lastPart := reStrip.ReplaceAllString(word, "")
	return utf16LenDE(lastPart) >= 3 && !r.IsMisspelled(lastPart)
}

// ignoreCompoundNonHyphenated ports the non-hyphenated branch of ignoreCompoundWithIgnoredWord
// (e.g. "Feynmandiagramm", "westperuanische"). Needs TagPOS for isNoun/isAdjective.
func (r *GermanSpellerRule) ignoreCompoundNonHyphenated(word string) bool {
	end := r.StartsWithIgnoredWord(word, true)
	if end < 3 {
		// geo prefixes (case-sensitive as in Java)
		if strings.HasPrefix(word, "ost") || strings.HasPrefix(word, "süd") {
			end = 3
		} else if strings.HasPrefix(word, "west") || strings.HasPrefix(word, "nord") {
			end = 4
		} else {
			return false
		}
	}
	// work in runes for correct UTF-8 slicing of ignoredWord/partialWord
	runes := []rune(word)
	if end > len(runes) {
		return false
	}
	ignoredWord := string(runes[:end])
	partialWord := string(runes[end:])
	if strings.HasSuffix(partialWord, ".") {
		partialWord = partialWord[:len(partialWord)-1]
	}
	isN := r.isNoun(partialWord)
	isUppercaseNoun := false
	if !isN && !startsWithUppercase(partialWord) {
		isUppercaseNoun = r.isNoun(uppercaseFirstChar(partialWord))
	}
	isDirection := reDirection.MatchString(ignoredWord)
	isAdj := r.isAdjective(ignoredWord)
	isDirectionalAdjective := isDirection && (isAdj || reDirectionalIsch.MatchString(partialWord))
	isCandidate := (isDirectionalAdjective || isN || isUppercaseNoun) &&
		!isAllUpperCase(ignoredWord) &&
		(isAllLowerCase(partialWord) || strings.HasSuffix(ignoredWord, "-"))
	if !isCandidate || utf16LenDE(partialWord) <= 2 {
		return false
	}
	needFugenS := isNeedingFugenS(ignoredWord)
	if needFugenS {
		if strings.HasPrefix(partialWord, "s") {
			partialWord = partialWord[1:]
		}
	}
	return spellerDictAccepts(partialWord) || spellerDictAccepts(uppercaseFirstChar(partialWord))
}

// IgnoreCompoundWithIgnoredWord ports GermanSpellerRule.ignoreCompoundWithIgnoredWord:
// hyphenated compounds and non-hyphenated (tagger-gated) compounds that start with
// an ignored / geo prefix.
func (r *GermanSpellerRule) IgnoreCompoundWithIgnoredWord(word string) bool {
	if r == nil || word == "" {
		return false
	}
	// Java: if (!startsWithUppercase(word) && !startsWithAny nord/west/ost/süd) return false
	if !startsWithUppercase(word) &&
		!strings.HasPrefix(word, "nord") &&
		!strings.HasPrefix(word, "west") &&
		!strings.HasPrefix(word, "ost") &&
		!strings.HasPrefix(word, "süd") {
		return false
	}
	parts := strings.Split(word, "-")
	if len(parts) < 2 {
		return r.ignoreCompoundNonHyphenated(word)
	}
	// hyphenated compound
	hasIgnoredWord := false
	var toSpellCheck []string
	stripFirst := word[len(parts[0])+1:] // after first "-"
	stripLast := word[:len(word)-len(parts[len(parts)-1])-1]

	if r.IgnoreWord(stripFirst) || r.IsIgnoredInCompounds(stripFirst) {
		hasIgnoredWord = true
		if !r.IgnoreWord(parts[0]) {
			toSpellCheck = append(toSpellCheck, parts[0])
		}
	} else if r.IgnoreWord(stripLast) || r.IsIgnoredInCompounds(stripLast) {
		hasIgnoredWord = true
		if !r.IgnoreWord(parts[len(parts)-1]) {
			toSpellCheck = append(toSpellCheck, parts[len(parts)-1])
		}
	} else {
		for _, p := range parts {
			if r.IgnoreWord(p) || r.IsIgnoredInCompounds(p) {
				hasIgnoredWord = true
			} else {
				toSpellCheck = append(toSpellCheck, p)
			}
		}
	}
	if !hasIgnoredWord {
		return false
	}
	if len(toSpellCheck) == 0 {
		return true
	}
	// remaining parts must spell-check; without dict cannot confirm
	if !FilterDictAvailable() {
		return false
	}
	for _, w := range toSpellCheck {
		if FilterDictIsMisspelled(w) {
			return false
		}
	}
	return true
}

// cutOffDot ports HunspellRule.cutOffDot for isProhibited checks.
func cutOffDot(s string) string {
	if strings.HasSuffix(s, ".") {
		return s[:len(s)-1]
	}
	return s
}

// AddProhibitedWords ports GermanSpellerRule.addProhibitedWords:
// single word ending ".*" → wordStartsToBeProhibited (prefix);
// first word starting ".*" → each form's suffix after ".*" → wordEndingsToBeProhibited;
// otherwise exact wordsToBeProhibited (post LineExpander expansions).
func (r *GermanSpellerRule) AddProhibitedWords(words []string) {
	if r == nil || len(words) == 0 {
		return
	}
	if r.Prohibited == nil {
		r.Prohibited = map[string]struct{}{}
	}
	if r.ProhibitedStarts == nil {
		r.ProhibitedStarts = map[string]struct{}{}
	}
	if r.ProhibitedEnds == nil {
		r.ProhibitedEnds = map[string]struct{}{}
	}
	if len(words) == 1 && strings.HasSuffix(words[0], ".*") {
		prefix := words[0][:len(words[0])-2]
		if prefix != "" {
			r.ProhibitedStarts[prefix] = struct{}{}
		}
		return
	}
	if strings.HasPrefix(words[0], ".*") {
		for _, w := range words {
			if strings.HasPrefix(w, ".*") {
				end := w[2:]
				if end != "" {
					r.ProhibitedEnds[end] = struct{}{}
				}
			}
		}
		return
	}
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		r.Prohibited[w] = struct{}{}
	}
}

// IsProhibited ports GermanSpellerRule.isProhibited:
// exact set, or any wordStartsToBeProhibited prefix, or any wordEndingsToBeProhibited suffix.
func (r *GermanSpellerRule) IsProhibited(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if _, ok := r.Prohibited[word]; ok {
		return true
	}
	for start := range r.ProhibitedStarts {
		if start != "" && strings.HasPrefix(word, start) {
			return true
		}
	}
	for end := range r.ProhibitedEnds {
		if end != "" && strings.HasSuffix(word, end) {
			return true
		}
	}
	return false
}

// Java GermanSpellerRule.isMisspelled hard cases (before super.isMisspelled).
var (
	// Patterns use ^…$ because Java Pattern.matches() is full-string; Go MatchString is not.
	// START_WITH_SPIEL whitelist when word starts with "Spielzug".
	reStartWithSpiel = regexp.MustCompile(`^(?:Spielzugs?|Spielzugangs?|Spielzuganges|Spielzugbuchs?|Spielzugbüchern?|Spielzuges|Spielzugverluste?|Spielzugverluste[ns])$`)
	reEndWithSchafte = regexp.MustCompile(`^[A-ZÖÄÜ][a-zöäß-]+schafte$`)
	reSchafPattern   = regexp.MustCompile(`^.{3,}schaf(s|en)?$`)
	reSchafePattern  = regexp.MustCompile(`^(?:Alpenschaf|Berberschaf|Bergschaf|Blauschaf|Brillenschaf|Dachsteinschaf|Deichschaf|Dickhornschaf|Feinwollschaf|Fettschwanzschaf|Fleischschaf|Fuchsschaf|Glücksschaf|Hausschaf|Jungschaf|Karakulschaf|Klonschaf|Merinoschaf|Milchschaf|Mondschaf|Nutzschaf|Rhönschaf|Riesenwildschaf|Schaukelschaf|Schneeschaf|Steinschaf|Steppenschaf|Superschaf|Waldschaf|Weideschaf|Wildschaf|Wollschaf|Zackelschaf|Zuchtschaf|Zwergblauschaf)(s|en)?$`)
	reReplaceSchaf   = regexp.MustCompile(`schaf$`)
	reReplaceSchafs  = regexp.MustCompile(`schafs$`)
	reReplaceSchafen = regexp.MustCompile(`schafen$`)
)

// germanIsMisspelledHardCases ports the prefix of GermanSpellerRule.isMisspelled
// (SCHAF_/Spielzug/Standart/schafte) before super.isMisspelled.
// Returns (handled, result). handled=true means return result immediately.
func (r *GermanSpellerRule) germanIsMisspelledHardCases(word string) (handled bool, misspelled bool) {
	if reSchafPattern.MatchString(word) && !reSchafePattern.MatchString(word) {
		variant := reReplaceSchaf.ReplaceAllString(word, "schaft")
		variant = reReplaceSchafs.ReplaceAllString(variant, "schaft")
		variant = reReplaceSchafen.ReplaceAllString(variant, "schaften")
		if !r.IsMisspelled(variant) {
			return true, true
		}
	}
	if strings.HasPrefix(word, "Spielzug") && !reStartWithSpiel.MatchString(word) {
		return true, true
	}
	if strings.HasPrefix(word, "Standart") &&
		word != "Standarte" &&
		word != "Standarten" &&
		!strings.HasPrefix(word, "Standartenträger") &&
		!strings.HasPrefix(word, "Standartenführer") {
		return true, true
	}
	if strings.HasSuffix(word, "schafte") && reEndWithSchafte.MatchString(word) {
		return true, true
	}
	return false, false
}

// IsMisspelled ports GermanSpellerRule.isMisspelled then HunspellRule.isMisspelled:
// hard cases (SCHAF/Spielzug/Standart/schafte); isProhibited (after cutOffDot) forces true;
// ignore list accepts; else FilterDict when wired.
// Match walks tokens with IgnoreWordAt (incl. IgnoreCompoundWithIgnoredWord).
func (r *GermanSpellerRule) IsMisspelled(word string) bool {
	if r != nil && r.IsMisspelledOverride != nil {
		return r.IsMisspelledOverride(word)
	}
	if r != nil {
		if handled, miss := r.germanIsMisspelledHardCases(word); handled {
			return miss
		}
	}
	if r != nil && r.IsProhibited(cutOffDot(word)) {
		return true
	}
	// HunspellRule: … && !ignoreWord(word) || isProhibited
	if r != nil && r.IgnoreWord(word) {
		return false
	}
	if !FilterDictAvailable() {
		return false
	}
	// list-form German ignore extras (hyphen compounds) are Match-path only in Java;
	// expose via IgnoreCompoundWithIgnoredWord for callers / future Match — not folded here.
	return FilterDictIsMisspelled(word)
}

// LoadSpellingWordList ports CachingWordListLoader.loadWords for spelling extras:
// skip empty lines and # comments; one word (or multi-token line) per non-comment line.
// Multi-token antipatterns are returned as full lines (caller may special-case).
func LoadSpellingWordList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip trailing comments
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out, sc.Err()
}

// LoadIgnoreWordsFromFile ports GermanSpellerRule.addIgnoreWords over a spelling extras file:
// CH: ß→ss; lines ending "-*" → wordsToBeIgnoredInCompounds (IgnoredInCompounds);
// other lines expanded via LineExpander flags (/A /N /S …);
// multi-token expansions become IgnorePhrases (Java IGNORE_SPELLING antipatterns).
func (r *GermanSpellerRule) LoadIgnoreWordsFromFile(path string) error {
	words, err := LoadSpellingWordList(path)
	if err != nil {
		return err
	}
	exp := WireLineExpander()
	for _, origLine := range words {
		line := origLine
		// Java GermanSpellerRule.addIgnoreWords: Swiss replaces ß with ss
		if r != nil && r.LanguageVariant == "CH" {
			line = strings.ReplaceAll(line, "ß", "ss")
		}
		if strings.HasSuffix(origLine, "-*") {
			// Java: wordsToBeIgnoredInCompounds.add(line.substring(0, line.length()-2))
			base := line
			if len(base) >= 2 {
				base = base[:len(base)-2]
			}
			r.AddIgnoredInCompounds(base)
			continue
		}
		for _, w := range exp.ExpandLine(line) {
			// multi-token → IgnorePhrases; single-token → IgnoreWords (AddIgnoreWords routes both)
			r.AddIgnoreWords(w)
		}
	}
	return nil
}

// AddIgnoredInCompounds ports wordsToBeIgnoredInCompounds.add for "-*" lines.
func (r *GermanSpellerRule) AddIgnoredInCompounds(words ...string) {
	if r == nil {
		return
	}
	if r.IgnoredInCompounds == nil {
		r.IgnoredInCompounds = map[string]struct{}{}
	}
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		r.IgnoredInCompounds[w] = struct{}{}
	}
}

// IsIgnoredInCompounds reports membership in wordsToBeIgnoredInCompounds.
// Used by IgnoreWordAt / IgnoreCompoundWithIgnoredWord hyphen paths.
func (r *GermanSpellerRule) IsIgnoredInCompounds(word string) bool {
	if r == nil || word == "" || len(r.IgnoredInCompounds) == 0 {
		return false
	}
	_, ok := r.IgnoredInCompounds[word]
	return ok
}

// InitBaseSpellingIgnoreWords ports SpellingCheckRule.init() load of getSpellingFileName()
// (de/hunspell/spelling.txt) through GermanSpellerRule.addIgnoreWords / LineExpander.
func (r *GermanSpellerRule) InitBaseSpellingIgnoreWords(fsPath string) error {
	if r == nil {
		return nil
	}
	return r.LoadIgnoreWordsFromFile(fsPath)
}

// InitIgnoreFile ports SpellingCheckRule.init() load of getIgnoreFileName()
// (de/hunspell/ignore.txt) through addIgnoreWords / LineExpander.
// Words here are accepted but not used for suggestion generation (Java distinction).
func (r *GermanSpellerRule) InitIgnoreFile(fsPath string) error {
	if r == nil {
		return nil
	}
	return r.LoadIgnoreWordsFromFile(fsPath)
}

// LoadProhibitWordsFromFile ports SpellingCheckRule.init() load of getProhibitFileName():
// each line expandLine (LineExpander) then addProhibitedWords.
func (r *GermanSpellerRule) LoadProhibitWordsFromFile(path string) error {
	if r == nil {
		return nil
	}
	words, err := LoadSpellingWordList(path)
	if err != nil {
		return err
	}
	exp := WireLineExpander()
	for _, line := range words {
		expanded := exp.ExpandLine(line)
		if len(expanded) == 0 {
			continue
		}
		r.AddProhibitedWords(expanded)
	}
	return nil
}

// InitProhibitFile ports SpellingCheckRule.init() load of de/hunspell/prohibit.txt.
func (r *GermanSpellerRule) InitProhibitFile(fsPath string) error {
	return r.LoadProhibitWordsFromFile(fsPath)
}

