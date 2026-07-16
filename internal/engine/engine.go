package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/data"
	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/langid"
	"github.com/lucasew/lang/internal/messages"
	"github.com/lucasew/lang/internal/pattern"
	"github.com/lucasew/lang/internal/rules"
	"github.com/lucasew/lang/internal/srx"
)

// Options configures a check.
type Options struct {
	DataDir       string
	Language      string
	DisabledRules map[string]bool
	EnabledOnly   map[string]bool
}

// Result is the outcome of checking one text blob.
type Result struct {
	Language string
	Findings []finding.Finding
}

// Checker loads data once and runs checks.
type Checker struct {
	dataRoot string
	langs    []data.Language
	srxDoc   *srx.Document
	msgCache sync.Map // family -> messages.Bundle
	// pattern rules by language family
	rulesCache sync.Map // family -> []*pattern.Rule
}

// New resolves data and discovers languages.
func New(dataDirFlag string) (*Checker, error) {
	root, err := data.Resolve(dataDirFlag)
	if err != nil {
		return nil, err
	}
	langs, err := data.DiscoverLanguages(root)
	if err != nil {
		return nil, fmt.Errorf("discover languages: %w", err)
	}
	if len(langs) == 0 {
		return nil, fmt.Errorf("no languages found under %s", root)
	}
	srxPath := filepath.Join(root, "languagetool-core", "src", "main", "resources", "org", "languagetool", "resource", "segment.srx")
	srxDoc, err := srx.Load(srxPath)
	if err != nil {
		return nil, fmt.Errorf("load segment.srx: %w", err)
	}
	return &Checker{dataRoot: root, langs: langs, srxDoc: srxDoc}, nil
}

func (c *Checker) DataRoot() string           { return c.dataRoot }
func (c *Checker) Languages() []data.Language { return c.langs }

// Check analyzes text for the given file label.
func (c *Checker) Check(file, text string, opt Options) (*Result, error) {
	langCode := opt.Language
	if langCode == "" || langCode == "auto" {
		code, ok := langid.Detect(text, c.langs)
		if !ok {
			return nil, fmt.Errorf("language auto-detect failed")
		}
		langCode = code
	}

	lang, ok := data.Lookup(c.langs, langCode)
	if !ok {
		return nil, fmt.Errorf("unknown language %q (use lang languages)", langCode)
	}

	msg, err := c.messages(lang.Family)
	if err != nil {
		return nil, err
	}

	var findings []finding.Finding

	// Built-in Java-port rules
	if ruleEnabled(rules.RuleWhitespace, opt) {
		findings = append(findings, rules.MultipleWhitespace(text, file, lang.Code, msg)...)
	}
	if ruleEnabled(rules.RuleWordRepeat, opt) {
		findings = append(findings, rules.WordRepeat(text, file, lang.Code, msg)...)
	}

	// Pattern rules from official grammar XML
	prules, err := c.patternRules(lang)
	if err != nil {
		return nil, err
	}

	sentences := c.srxDoc.Split(text, lang.Family, "_two")
	if len(sentences) == 0 {
		sentences = []string{text}
	}

	offset := 0
	// Map rune offsets: walk text to find sentence starts
	fullRunes := []rune(text)
	cursor := 0
	for _, sent := range sentences {
		// Find sent in full text starting at cursor
		sentRunes := []rune(sent)
		// advance cursor to match
		pos := indexRunes(fullRunes[cursor:], sentRunes)
		if pos < 0 {
			pos = 0
		}
		base := cursor + pos
		ctx := pattern.NewMatchContext(file, lang.Code, sent, base)
		for _, r := range prules {
			if !ruleEnabled(r.ID, opt) && !ruleEnabled(r.FullID(), opt) {
				continue
			}
			if len(opt.EnabledOnly) > 0 {
				if !opt.EnabledOnly[r.ID] && !opt.EnabledOnly[r.FullID()] {
					continue
				}
			}
			findings = append(findings, pattern.MatchRule(r, ctx)...)
		}
		cursor = base + len(sentRunes)
		_ = offset
	}

	// Fix line/column for multi-sentence: recompute from full text using absolute offsets
	for i := range findings {
		line, col := offsetToLineColRunes(text, findings[i].Offset)
		endLine, endCol := offsetToLineColRunes(text, findings[i].EndOffset)
		findings[i].Line = line
		findings[i].Column = col
		findings[i].EndLine = endLine
		findings[i].EndColumn = endCol
	}

	return &Result{Language: lang.Code, Findings: findings}, nil
}

func (c *Checker) messages(family string) (messages.Bundle, error) {
	if v, ok := c.msgCache.Load(family); ok {
		return v.(messages.Bundle), nil
	}
	b, err := messages.Load(c.dataRoot, family)
	if err != nil {
		return nil, err
	}
	c.msgCache.Store(family, b)
	return b, nil
}

func (c *Checker) patternRules(lang data.Language) ([]*pattern.Rule, error) {
	if v, ok := c.rulesCache.Load(lang.Family); ok {
		return v.([]*pattern.Rule), nil
	}
	paths := grammarPaths(c.dataRoot, lang)
	var all []*pattern.Rule
	for _, p := range paths {
		rs, err := pattern.LoadFile(p)
		if err != nil {
			// missing optional files ok
			continue
		}
		all = append(all, rs...)
	}
	c.rulesCache.Store(lang.Family, all)
	return all, nil
}

func grammarPaths(dataRoot string, lang data.Language) []string {
	base := filepath.Join(data.LanguageModules(dataRoot), lang.Family, "src", "main", "resources", "org", "languagetool", "rules", lang.Family)
	var paths []string
	// All *.xml under rules/<family>/ (grammar, style, punctuation, variants…)
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".xml") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths
}

func ruleEnabled(id string, opt Options) bool {
	if opt.DisabledRules[id] {
		return false
	}
	if len(opt.EnabledOnly) > 0 {
		return opt.EnabledOnly[id]
	}
	return true
}

func indexRunes(haystack, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func offsetToLineColRunes(text string, runeOffset int) (line, col int) {
	line, col = 1, 1
	i := 0
	for _, r := range text {
		if i >= runeOffset {
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
