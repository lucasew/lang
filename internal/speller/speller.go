// Package speller implements Morfologik dictionary spelling checks.
package speller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/messages"
	"github.com/lucasew/lang/internal/morfologik"
	"github.com/lucasew/lang/internal/pipeline"
)

const (
	RuleEnUS = "MORFOLOGIK_RULE_EN_US"
	RuleEnGB = "MORFOLOGIK_RULE_EN_GB"
)

// Speller checks words against a morfologik speller dict.
type Speller struct {
	dict   *morfologik.Dictionary
	ruleID string
}

// Open opens a speller dictionary.
func Open(dictPath, ruleID string) (*Speller, error) {
	d, err := morfologik.OpenDictionary(dictPath)
	if err != nil {
		return nil, err
	}
	return &Speller{dict: d, ruleID: ruleID}, nil
}

// OpenEnglishUS loads en_US.dict.
func OpenEnglishUS(dataRoot string) (*Speller, error) {
	candidates := []string{
		filepath.Join(dataRoot, "languagetool-language-modules", "en", "src", "main", "resources",
			"org", "languagetool", "resource", "en", "hunspell", "en_US.dict"),
	}
	if p := findUp(filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", "en_US.dict")); p != "" {
		candidates = append([]string{p}, candidates...)
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return Open(p, RuleEnUS)
		}
	}
	return nil, fmt.Errorf("en_US.dict not found (run scripts/fetch-english-dicts.sh)")
}

func findUp(rel string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Check finds unknown words (misspellings) in text.
func (s *Speller) Check(text, file, lang string, msg messages.Bundle) []finding.Finding {
	if s == nil || s.dict == nil {
		return nil
	}
	message := msg.Get("spelling")
	if message == "spelling" || message == "" {
		message = "Possible spelling mistake found."
	}
	tokens := pipeline.EnglishWordTokenize(text)
	var out []finding.Finding
	for _, tok := range tokens {
		if tok.Whitespace || !isSpellable(tok.Text) {
			continue
		}
		w := strings.ReplaceAll(tok.Text, "’", "'")
		if s.known(w) {
			continue
		}
		if isCapitalized(w) && s.known(strings.ToLower(w)) {
			continue
		}
		if isAllUpper(w) && s.known(strings.ToLower(w)) {
			continue
		}
		line, col := offsetLineCol(text, tok.Start)
		endLine, endCol := offsetLineCol(text, tok.End)
		typ, sev := finding.WithType("misspelling")
		out = append(out, finding.Finding{
			File:        file,
			Line:        line,
			Column:      col,
			EndLine:     endLine,
			EndColumn:   endCol,
			Offset:      tok.Start,
			EndOffset:   tok.End,
			Rule:        s.ruleID,
			Type:        typ,
			Severity:    sev,
			Message:     message,
			Suggestions: s.suggest(w, 5),
			Language:    lang,
		})
	}
	return out
}

func (s *Speller) known(w string) bool {
	if s.dict.Contains(w) {
		return true
	}
	forms, _ := s.dict.Lookup(w)
	return len(forms) > 0
}

// suggest generates simple edit-distance candidates present in the dictionary.
func (s *Speller) suggest(word string, limit int) []string {
	if limit <= 0 {
		limit = 5
	}
	seen := map[string]bool{word: true}
	var out []string
	add := func(c string) {
		if c == "" || seen[c] {
			return
		}
		if s.known(c) {
			seen[c] = true
			out = append(out, c)
		}
	}
	// lowercase variants
	lw := strings.ToLower(word)
	add(lw)
	if len(lw) > 0 {
		add(strings.ToUpper(lw[:1]) + lw[1:])
	}
	runes := []rune(lw)
	// deletes
	for i := range runes {
		add(string(append(append([]rune{}, runes[:i]...), runes[i+1:]...)))
	}
	// transposes
	for i := 0; i+1 < len(runes); i++ {
		r := append([]rune{}, runes...)
		r[i], r[i+1] = r[i+1], r[i]
		add(string(r))
	}
	// replaces with a-z
	for i := range runes {
		for c := 'a'; c <= 'z'; c++ {
			if rune(c) == runes[i] {
				continue
			}
			r := append([]rune{}, runes...)
			r[i] = c
			add(string(r))
			if len(out) >= limit {
				return out[:limit]
			}
		}
	}
	// inserts a-z
	for i := 0; i <= len(runes); i++ {
		for c := 'a'; c <= 'z'; c++ {
			r := make([]rune, 0, len(runes)+1)
			r = append(r, runes[:i]...)
			r = append(r, c)
			r = append(r, runes[i:]...)
			add(string(r))
			if len(out) >= limit {
				return out[:limit]
			}
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func isSpellable(s string) bool {
	if s == "" || s == "SENT_START" {
		return false
	}
	has := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			has = true
			break
		}
	}
	if !has {
		return false
	}
	if strings.Contains(s, "://") || strings.Contains(s, "@") {
		return false
	}
	return true
}

func isCapitalized(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)
	if !unicode.IsUpper(r[0]) {
		return false
	}
	for i := 1; i < len(r); i++ {
		if unicode.IsLetter(r[i]) && !unicode.IsLower(r[i]) {
			return false
		}
	}
	return true
}

func isAllUpper(s string) bool {
	has := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			has = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return has
}

func offsetLineCol(text string, runeOff int) (line, col int) {
	line, col = 1, 1
	i := 0
	for _, r := range text {
		if i >= runeOff {
			break
		}
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		i++
	}
	return line, col
}
