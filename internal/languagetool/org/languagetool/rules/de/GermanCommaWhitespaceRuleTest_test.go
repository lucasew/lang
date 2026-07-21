package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanCommaWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanCommaWhitespaceRule_Rule(t *testing.T) {
	rule := NewGermanCommaWhitespaceRule(nil)
	require.Equal(t, "COMMA_PARENTHESIS_WHITESPACE", rule.GetID())
	require.Contains(t, rule.GetURL(), "grammatik-leerzeichen")

	// Java: space before . for TLD/domain labels is exception (.de-Domains)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Es gibt 5 Millionen .de-Domains."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Es gibt 5 Millionen .com-Domain."))))

	// normal space-before-comma still flagged (Java example pair)
	ms := rule.Match(languagetool.AnalyzePlain("Die Partei , die die letzte Wahl gewann."))
	require.Equal(t, 1, len(ms))
	// DE wrapper sets Rule so URL is available on the match rule
	if ms[0].Rule != nil {
		if u, ok := ms[0].Rule.(interface{ GetURL() string }); ok {
			require.Contains(t, u.GetURL(), "leerzeichen")
		}
	}

	// non-domain space before period still flagged
	require.GreaterOrEqual(t, len(rule.Match(languagetool.AnalyzePlain("Ende . Nächster Satz."))), 1)
}

// Twin of Java isException domain regex [a-z]{2,10}-Domains?
func TestGermanCommaWhitespaceRule_DomainExceptionRE(t *testing.T) {
	require.True(t, deDomainLabel.MatchString("de-Domains"))
	require.True(t, deDomainLabel.MatchString("com-Domain"))
	require.True(t, deDomainLabel.MatchString("info-Domains"))
	require.False(t, deDomainLabel.MatchString("DE-Domains")) // case-sensitive like Java
	require.False(t, deDomainLabel.MatchString("toolongname-Domain"))
	require.False(t, deDomainLabel.MatchString("x-Domain")) // need ≥2 letters
}
