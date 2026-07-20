package languagetool

import (
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
)

// SimpleMultipleWhitespaceChecker flags two or more consecutive regular spaces.
func SimpleMultipleWhitespaceChecker() SentenceChecker {
	return func(sentence *AnalyzedSentence) []LocalMatch {
		if sentence == nil {
			return nil
		}
		text := sentence.GetText()
		var out []LocalMatch
		runes := []rune(text)
		i := 0
		for i < len(runes) {
			if runes[i] != ' ' {
				i++
				continue
			}
			j := i
			for j < len(runes) && runes[j] == ' ' {
				j++
			}
			if j-i >= 2 {
				// byte offsets: for ASCII spaces, rune index == byte index if pure ASCII;
				// compute carefully via rune→byte
				from := runeOffsetToByte(text, i)
				to := runeOffsetToByte(text, j)
				out = append(out, LocalMatch{
					FromPos: from, ToPos: to,
					Message:     "Multiple whitespace",
					RuleID:      "WHITESPACE_RULE",
					Suggestions: []string{" "},
				})
			}
			i = j
		}
		return out
	}
}

// SimpleUnpairedBracketsChecker flags unmatched () [] {} in sentence text.
func SimpleUnpairedBracketsChecker() SentenceChecker {
	pairs := map[rune]rune{')': '(', ']': '[', '}': '{'}
	openers := map[rune]bool{'(': true, '[': true, '{': true}
	return func(sentence *AnalyzedSentence) []LocalMatch {
		if sentence == nil {
			return nil
		}
		text := sentence.GetText()
		type frame struct {
			ch  rune
			pos int // byte offset
		}
		var stack []frame
		var out []LocalMatch
		bytePos := 0
		for _, r := range text {
			w := utf8.RuneLen(r)
			if openers[r] {
				stack = append(stack, frame{r, bytePos})
			} else if want, ok := pairs[r]; ok {
				if len(stack) == 0 || stack[len(stack)-1].ch != want {
					out = append(out, LocalMatch{
						FromPos: bytePos, ToPos: bytePos + w,
						Message: "Unpaired bracket", RuleID: "UNPAIRED_BRACKETS",
					})
				} else {
					stack = stack[:len(stack)-1]
				}
			}
			bytePos += w
		}
		for _, f := range stack {
			out = append(out, LocalMatch{
				FromPos: f.pos, ToPos: f.pos + utf8.RuneLen(f.ch),
				Message: "Unpaired bracket", RuleID: "UNPAIRED_BRACKETS",
			})
		}
		return out
	}
}

// SimplePhraseReplaceChecker flags phrase occurrences from an explicit map
// (test helper / injected data only — not an invent soft phrase pack).
// Matching is case-insensitive; suggestions inherit ALL-CAPS or leading capital
// from the matched span (Java StringTools-style case preservation).
// phrases maps wrong phrase → suggested replacement.
func SimplePhraseReplaceChecker(ruleID string, phrases map[string]string) SentenceChecker {
	if ruleID == "" {
		ruleID = "PHRASE_REPLACE"
	}
	// longer phrases first
	keys := make([]string, 0, len(phrases))
	for k := range phrases {
		if k != "" {
			keys = append(keys, k)
		}
	}
	// simple sort by length desc
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if len(keys[j]) > len(keys[i]) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return func(sentence *AnalyzedSentence) []LocalMatch {
		if sentence == nil {
			return nil
		}
		text := sentence.GetText()
		// ASCII-oriented case fold (soft EN/phrase pack is Latin-1 friendly).
		lowText := strings.ToLower(text)
		var out []LocalMatch
		// find non-overlapping left-to-right
		used := make([]bool, len(text))
		for _, wrong := range keys {
			repl := phrases[wrong]
			lowWrong := strings.ToLower(wrong)
			start := 0
			for {
				idx := strings.Index(lowText[start:], lowWrong)
				if idx < 0 {
					break
				}
				idx += start
				// skip if overlaps used
				overlap := false
				for b := idx; b < idx+len(wrong); b++ {
					if b < len(used) && used[b] {
						overlap = true
						break
					}
				}
				if !overlap {
					for b := idx; b < idx+len(wrong); b++ {
						if b < len(used) {
							used[b] = true
						}
					}
					matched := text[idx : idx+len(wrong)]
					sug := PreserveCase(matched, repl)
					out = append(out, LocalMatch{
						FromPos: idx, ToPos: idx + len(wrong),
						Message: "Phrase", RuleID: ruleID,
						Suggestions: []string{sug},
					})
				}
				start = idx + len(wrong)
			}
		}
		return out
	}
}

// PreserveCase maps suggestion casing from the matched original span
// (Java StringTools / suggestion case adjustment — not a soft invent path).
// ALL-CAPS match → ALL-CAPS suggestion; leading capital → capitalize suggestion
// when the suggestion replaces the whole span (same or more word tokens).
// Shorter single-token fixes for multi-word matches (e.g. "They is" → "are")
// keep suggestion casing so apply does not yield "Are …".
func PreserveCase(matched, suggestion string) string {
	if matched == "" || suggestion == "" {
		return suggestion
	}
	hasLetter := false
	allUpper := true
	for _, r := range matched {
		if r >= 'a' && r <= 'z' {
			hasLetter = true
			allUpper = false
			break
		}
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
		}
	}
	if hasLetter && allUpper {
		return strings.ToUpper(suggestion)
	}
	// leading capital (sentence/title start) — only when suggestion covers the span
	mWords := len(strings.Fields(matched))
	sWords := len(strings.Fields(suggestion))
	if mWords > 1 && sWords < mWords {
		return suggestion
	}
	for _, r := range matched {
		if r >= 'A' && r <= 'Z' {
			rs := []rune(suggestion)
			if len(rs) == 0 {
				return suggestion
			}
			if rs[0] >= 'a' && rs[0] <= 'z' {
				rs[0] = rs[0] - ('a' - 'A')
			}
			return string(rs)
		}
		if r >= 'a' && r <= 'z' {
			break
		}
	}
	return suggestion
}

// CheckAnnotated ports check(AnnotatedText): analyze plain text, then map match
// offsets to original markup positions via getOriginalTextPositionFor (same as
// TextCheckCallable adjustRuleMatchPos / getTextLevelRuleMatches).
func (lt *JLanguageTool) CheckAnnotated(at *markup.AnnotatedText) []LocalMatch {
	if lt == nil || at == nil {
		return nil
	}
	plain := at.GetPlainText()
	mapper := func(pos int, isToPos bool) int {
		return at.GetOriginalTextPositionFor(pos, isToPos)
	}
	return lt.checkInternal(plain, mapper)
}

// CheckAnnotatedWithResults ports CheckWithResults for AnnotatedText (original offsets).
func (lt *JLanguageTool) CheckAnnotatedWithResults(at *markup.AnnotatedText) (*CheckResults, error) {
	if lt == nil || at == nil {
		return NewCheckResults(nil, nil), nil
	}
	plain := at.GetPlainText()
	mapper := func(pos int, isToPos bool) int {
		return at.GetOriginalTextPositionFor(pos, isToPos)
	}
	matches := lt.checkInternal(plain, mapper)
	return lt.checkResultsFromMatches(plain, matches)
}

// ProjectMatchesToOriginal maps plain-text LocalMatch offsets to original markup offsets.
// Java adjustRuleMatchPos / text-level path:
//
//	fromPos = getOriginalTextPositionFor(fromPos, false)
//	toPos   = getOriginalTextPositionFor(toPos - 1, true) + 1
func ProjectMatchesToOriginal(at *markup.AnnotatedText, matches []LocalMatch) []LocalMatch {
	if at == nil || len(matches) == 0 {
		return matches
	}
	out := make([]LocalMatch, len(matches))
	for i, m := range matches {
		out[i] = m
		from := m.FromPos
		to := m.ToPos
		if from < 0 {
			from = 0
		}
		out[i].FromPos = at.GetOriginalTextPositionFor(from, false)
		if to <= 0 {
			out[i].ToPos = out[i].FromPos
		} else {
			out[i].ToPos = at.GetOriginalTextPositionFor(to-1, true) + 1
		}
	}
	return out
}

// CheckWithResults ports TextCheckCallable.call() surface for plain text:
// runs Check (mode-filtered matches), builds ExtendedSentenceRanges via
// computeSentenceData + whitespace-fix ranges, applies maxErrorsPerWordRate,
// and collects ignore Ranges from LocalMatch.NewLanguageMatches.
func (lt *JLanguageTool) CheckWithResults(text string) (*CheckResults, error) {
	if lt == nil {
		return NewCheckResults(nil, nil), nil
	}
	return lt.checkResultsFromMatches(text, lt.Check(text))
}

// checkResultsFromMatches builds CheckResults (extended ranges, ignore ranges, error rate).
func (lt *JLanguageTool) checkResultsFromMatches(text string, matches []LocalMatch) (*CheckResults, error) {
	sents := lt.Analyze(text)
	data := sentenceDataFromAnalyzed(sents)
	lang := lt.LanguageCode
	if lang == "" {
		lang = "?"
	}
	short := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		short = lang[:i]
	}
	ext := make([]ExtendedSentenceRange, 0, len(data))
	wordCounter := 0
	for _, sd := range data {
		ext = append(ext, BuildExtendedSentenceRange(sd, short))
		wordCounter += sd.WordCount
	}
	// getOtherRuleMatches: NewLanguageMatches → ignore range + confidence update.
	// Prefer sentence-local positions when FromPos is already in original/markup space.
	var ignore []Range
	for _, m := range matches {
		if len(m.NewLanguageMatches) == 0 || len(data) == 0 {
			continue
		}
		pos := m.FromPos
		if m.FromPosSentence >= 0 && m.SentenceText != "" {
			// locate sentence by text when possible
			for _, sd := range data {
				if sd.Text == m.SentenceText {
					pos = sd.StartOffset
					break
				}
			}
		}
		sd := findSentenceContaining(data, pos)
		from := sd.StartOffset
		to := sd.StartOffset + utf16Len(sd.Text)
		var extPtr *ExtendedSentenceRange
		for i := range ext {
			if ext[i].FromPos == from || (ext[i].FromPos <= from && from < ext[i].ToPos) {
				extPtr = &ext[i]
				break
			}
		}
		ignore = ApplyNewLanguageMatchesToSentence(ignore, extPtr, from, to, m.NewLanguageMatches)
	}
	name := lt.LanguageName
	if name == "" {
		name = lang
	}
	if err := CheckErrorRate(len(matches), wordCounter, lt.MaxErrorsPerWordRate, name, utf16Len(text)); err != nil {
		return nil, err
	}
	anyMatches := make([]any, len(matches))
	for i := range matches {
		anyMatches[i] = matches[i]
	}
	return NewCheckResultsFull(anyMatches, ignore, ext), nil
}

// LocalMatchesFromCheckResults unpacks LocalMatch values from CheckResults.RuleMatches.
func LocalMatchesFromCheckResults(cr *CheckResults) []LocalMatch {
	if cr == nil {
		return nil
	}
	out := make([]LocalMatch, 0, len(cr.RuleMatches))
	for _, m := range cr.RuleMatches {
		if lm, ok := m.(LocalMatch); ok {
			out = append(out, lm)
		}
	}
	return out
}

// RegisterDemoEnglishCheckers installs faithful a/an, word-repeat, multi-space,
// unpaired brackets, and optional map speller for homepage-style demos.
// Soft invent PHRASE_REPLACE packs ("tot he" etc.) are not registered — use grammar.xml.
// Tiny demo lexicons do not use nearestKnownWords (that fights a/an and loops forever).
func (lt *JLanguageTool) RegisterDemoEnglishCheckers(known map[string]struct{}, spellSuggestions map[string][]string) {
	if lt == nil {
		return
	}
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("WHITESPACE_RULE", SimpleMultipleWhitespaceChecker())
	lt.AddRuleChecker("UNPAIRED_BRACKETS", SimpleUnpairedBracketsChecker())
	if known != nil {
		// Exact surface only — no soft lowercase invent (DemoEnglishKnownWords dual-cases).
		isKnown := func(w string) bool {
			_, ok := known[w]
			return ok
		}
		// Explicit suggestion map only — no soft edit-distance invent when map miss.
		lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", SimplePredicateSpellerChecker(
			"MORFOLOGIK_RULE_EN_US", isKnown, spellSuggestions, nil, nil,
		))
	}
}

func runeOffsetToByte(s string, runeIndex int) int {
	if runeIndex <= 0 {
		return 0
	}
	i := 0
	for pos := range s {
		if i == runeIndex {
			return pos
		}
		i++
	}
	return len(s)
}
