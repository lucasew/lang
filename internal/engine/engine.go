package engine

import (
	"fmt"

	"github.com/lucasew/lang/internal/data"
	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/langid"
	"github.com/lucasew/lang/internal/messages"
	"github.com/lucasew/lang/internal/rules"
)

// Options configures a check.
type Options struct {
	DataDir  string
	Language string // "auto" or code
	// DisabledRules lists rule IDs to skip.
	DisabledRules map[string]bool
	// EnabledOnly if non-empty, only these rule IDs run.
	EnabledOnly map[string]bool
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
	return &Checker{dataRoot: root, langs: langs}, nil
}

func (c *Checker) DataRoot() string       { return c.dataRoot }
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

	msg, err := messages.Load(c.dataRoot, lang.Family)
	if err != nil {
		return nil, err
	}

	var findings []finding.Finding
	// Pipeline stage: rules (initial set).
	// WHITESPACE_RULE — port of MultipleWhitespaceRule.
	if ruleEnabled(rules.RuleWhitespace, opt) {
		findings = append(findings, rules.MultipleWhitespace(text, file, lang.Code, msg)...)
	}

	return &Result{Language: lang.Code, Findings: findings}, nil
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
