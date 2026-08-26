// Package speller implements Morfologik dictionary spelling checks.
package speller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/attic/finding"
	"github.com/lucasew/lang/internal/attic/messages"
	"github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/attic/pipeline"
)

// Speller checks words against a morfologik speller dict.
type Speller struct {
	dict   *morfologik.Dictionary
	ruleID string
	family string
}

// Open opens a speller dictionary with the given LT rule id.
func Open(dictPath, ruleID string) (*Speller, error) {
	d, err := morfologik.OpenDictionary(dictPath)
	if err != nil {
		return nil, err
	}
	return &Speller{dict: d, ruleID: ruleID}, nil
}

// OpenForLanguage finds a hunspell morfologik dict under the LT data tree.
// rule id: MORFOLOGIK_RULE_<LANGCODE_UPPER_UNDERSCORE> e.g. MORFOLOGIK_RULE_DE_DE
func OpenForLanguage(dataRoot, family, langCode string) (*Speller, error) {
	code := strings.ReplaceAll(langCode, "-", "_")
	if !strings.Contains(code, "_") && family != "" {
		// try common defaults
		switch family {
		case "de":
			code = "de_DE"
		case "en":
			code = "en_US"
		case "pt":
			code = "pt_BR"
		case "it":
			code = "it_IT"
		case "pl":
			code = "pl_PL"
		case "ru":
			code = "ru_RU"
		case "nl":
			code = "nl_NL"
		case "es":
			code = "es_ES"
		case "fr":
			code = "fr_FR"
		default:
			code = family + "_" + strings.ToUpper(family)
		}
	}
	candidates := []string{
		filepath.Join(dataRoot, "languagetool-language-modules", family, "src", "main", "resources",
			"org", "languagetool", "resource", family, "hunspell", code+".dict"),
		filepath.Join(dataRoot, "languagetool-language-modules", family, "src", "main", "resources",
			"org", "languagetool", "resource", family, "hunspell", strings.ToUpper(code[:1])+code[1:]+".dict"),
	}
	// English third_party
	if family == "en" {
		if p := findUp(filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", "en_US.dict")); p != "" {
			candidates = append([]string{p}, candidates...)
		}
		// also try lang code file
		if p := findUp(filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", code+".dict")); p != "" {
			candidates = append([]string{p}, candidates...)
		}
	}
	// scan hunspell dir for any .dict
	hunDir := filepath.Join(dataRoot, "languagetool-language-modules", family, "src", "main", "resources",
		"org", "languagetool", "resource", family, "hunspell")
	if ents, err := os.ReadDir(hunDir); err == nil {
		for _, e := range ents {
			if strings.HasSuffix(e.Name(), ".dict") {
				candidates = append(candidates, filepath.Join(hunDir, e.Name()))
			}
		}
	}

	var lastErr error
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			ruleID := "MORFOLOGIK_RULE_" + strings.ToUpper(strings.ReplaceAll(
				strings.TrimSuffix(filepath.Base(p), ".dict"), "-", "_"))
			sp, err := Open(p, ruleID)
			if err != nil {
				lastErr = err
				continue
			}
			sp.family = family
			return sp, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no speller dict for %s/%s", family, langCode)
}

// OpenEnglishUS loads en_US.dict (compat).
func OpenEnglishUS(dataRoot string) (*Speller, error) {
	return OpenForLanguage(dataRoot, "en", "en-US")
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

// RuleID returns the LanguageTool-style rule id for this speller.
func (s *Speller) RuleID() string {
	if s == nil {
		return ""
	}
	return s.ruleID
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
	var tokens []pipeline.Token
	if s.family == "en" || strings.HasPrefix(lang, "en") {
		tokens = pipeline.EnglishWordTokenize(text)
	} else {
		tokens = pipeline.WordTokenize(text)
	}
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
	lw := strings.ToLower(word)
	seen := map[string]bool{word: true, lw: true}
	var out []string
	add := func(c string) {
		if c == "" || seen[c] {
			return
		}
		if len([]rune(word)) >= 3 && len([]rune(c)) < 2 {
			return
		}
		if s.known(c) {
			seen[c] = true
			out = append(out, c)
		}
	}
	runes := []rune(lw)
	for i := 0; i+1 < len(runes); i++ {
		r := append([]rune{}, runes...)
		r[i], r[i+1] = r[i+1], r[i]
		add(string(r))
	}
	for i := range runes {
		orig := runes[i]
		for c := 'a'; c <= 'z'; c++ {
			if rune(c) == orig {
				continue
			}
			runes[i] = c
			add(string(runes))
			runes[i] = orig
			if len(out) >= limit {
				return out[:limit]
			}
		}
	}
	for i := range runes {
		add(string(append(append([]rune{}, runes[:i]...), runes[i+1:]...)))
	}
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
