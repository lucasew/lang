package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanCommaWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// TestGermanCommaWhitespaceRule_Rule ports GermanCommaWhitespaceRuleTest.testRule.
//
// Java:
//
//	rule = new GermanCommaWhitespaceRule(TestTools.getEnglishMessages(),
//	  Example.wrong("Die Partei<marker> ,</marker> die die letzte Wahl gewann."),
//	  Example.fixed("Die Partei<marker>,</marker> die die letzte Wahl gewann."),
//	  Tools.getUrl("http://fake"));
//	lt = new JLanguageTool(Languages.getLanguageForShortCode("de"));
//	assertThat(rule.match(lt.getAnalyzedSentence("Es gibt 5 Millionen .de-Domains.")).length, is(0));
func TestGermanCommaWhitespaceRule_Rule(t *testing.T) {
	rule := NewGermanCommaWhitespaceRule(
		nil, // TestTools.getEnglishMessages — not asserted by Java testRule
		rules.Wrong("Die Partei<marker> ,</marker> die die letzte Wahl gewann."),
		rules.Fixed("Die Partei<marker>,</marker> die die letzte Wahl gewann."),
		"http://fake",
	)
	lt := languagetool.NewJLanguageTool("de")
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Es gibt 5 Millionen .de-Domains."))))
}

// Twin of Java isException domain regex [a-z]{2,10}-Domains? with Matcher.matches().
func TestGermanCommaWhitespaceRule_DomainExceptionRE(t *testing.T) {
	require.True(t, domains.MatchString("de-Domains"))
	require.True(t, domains.MatchString("com-Domain"))
	require.True(t, domains.MatchString("info-Domains"))
	require.False(t, domains.MatchString("DE-Domains"))         // case-sensitive like Java
	require.False(t, domains.MatchString("toolongname-Domain")) // >10 letters before -
	require.False(t, domains.MatchString("x-Domain"))           // need ≥2 letters
	require.False(t, domains.MatchString("de-DomainsX"))        // matches() full string only
	require.True(t, domains.MatchString("xde-Domains"))         // [a-z]{2,10} accepts "xde"
}
