package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MultipleWhitespaceRule wraps the core MultipleWhitespaceRule for this language.
type MultipleWhitespaceRule struct {
	*rules.MultipleWhitespaceRule
}

func NewMultipleWhitespaceRule(messages map[string]string) *MultipleWhitespaceRule {
	return &MultipleWhitespaceRule{MultipleWhitespaceRule: rules.NewMultipleWhitespaceRule(messages)}
}

func (r *MultipleWhitespaceRule) Match(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.MultipleWhitespaceRule.Match(sentences)
}
