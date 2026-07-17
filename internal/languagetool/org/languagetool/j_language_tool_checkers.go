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

// SimplePhraseReplaceChecker flags exact phrase occurrences (ASCII/space phrases).
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
		var out []LocalMatch
		// find non-overlapping left-to-right
		used := make([]bool, len(text))
		for _, wrong := range keys {
			repl := phrases[wrong]
			start := 0
			for {
				idx := strings.Index(text[start:], wrong)
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
					out = append(out, LocalMatch{
						FromPos: idx, ToPos: idx + len(wrong),
						Message: "Phrase", RuleID: ruleID,
						Suggestions: []string{repl},
					})
				}
				start = idx + len(wrong)
			}
		}
		return out
	}
}

// CheckAnnotated runs Check on plain text extracted from AnnotatedText.
// Match offsets are in plain-text space (use AnnotatedText mapping to project).
func (lt *JLanguageTool) CheckAnnotated(at *markup.AnnotatedText) []LocalMatch {
	if lt == nil || at == nil {
		return nil
	}
	return lt.Check(at.GetPlainText())
}

// ProjectMatchesToOriginal maps plain-text LocalMatch offsets to original markup offsets.
func ProjectMatchesToOriginal(at *markup.AnnotatedText, matches []LocalMatch) []LocalMatch {
	if at == nil || len(matches) == 0 {
		return matches
	}
	out := make([]LocalMatch, len(matches))
	for i, m := range matches {
		out[i] = m
		out[i].FromPos = at.GetOriginalTextPositionFor(m.FromPos, false)
		out[i].ToPos = at.GetOriginalTextPositionFor(m.ToPos, true)
	}
	return out
}

// RegisterDemoEnglishCheckers installs a/an, word-repeat, multi-space, unpaired brackets,
// common phrase fixes, and optional map speller for homepage-style demos.
// Tiny demo lexicons do not use nearestKnownWords (that fights a/an and loops forever).
func (lt *JLanguageTool) RegisterDemoEnglishCheckers(known map[string]struct{}, spellSuggestions map[string][]string) {
	if lt == nil {
		return
	}
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("WHITESPACE_RULE", SimpleMultipleWhitespaceChecker())
	lt.AddRuleChecker("UNPAIRED_BRACKETS", SimpleUnpairedBracketsChecker())
	lt.AddRuleChecker("PHRASE_REPLACE", SimplePhraseReplaceChecker("PHRASE_REPLACE", map[string]string{
		"tot he": "to the",
	}))
	if known != nil {
		isKnown := func(w string) bool {
			if _, ok := known[w]; ok {
				return true
			}
			_, ok := known[strings.ToLower(w)]
			return ok
		}
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
