package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of Java PatternRule: message keeps <suggestion> tags through registration.
// Soft strip→SuggestionTemplates path removed (broke suppress_misspelled + multi-synth).
func TestRegisterGrammarXML_KeepsSuggestionTags(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<rules lang="en">
  <category id="CAT" name="Cat">
    <rule id="KEEP_SUG" name="keep">
      <pattern>
        <token>foo</token>
      </pattern>
      <message>Use <suggestion>bar</suggestion> instead</message>
    </rule>
  </category>
</rules>`
	n, err := RegisterGrammarXML(lt, xml, "t.xml", "en")
	require.NoError(t, err)
	require.Equal(t, 1, n)

	// Loader keeps <suggestion> tags (Java PatternRule message); registration must not strip.
	loader := NewPatternRuleLoader()
	ars, err := loader.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	require.Contains(t, ars[0].Message, "<suggestion>")
	require.Contains(t, ars[0].Message, "bar")
	require.Contains(t, ars[0].Message, "</suggestion>")
}

// End-to-end: FormatMatches + createRuleMatch extract suggestions from tags (Java RuleMatch ctor).
func TestCreateRuleMatch_SuggestionTagsNotStripped(t *testing.T) {
	raw := `Use <suggestion><match no="1" case_conversion="allupper"/></suggestion>`
	msg, matches := ProcessRuleMessage(raw)
	require.Contains(t, msg, "<suggestion>")
	pr := NewPatternRule("T", "en", []*PatternToken{Token("hello")}, "d", msg, "")
	pr.SuggestionMatches = matches
	sent := testSentence(atr("hello", 0))
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, []string{"HELLO"}, ms[0].GetSuggestedReplacements())
	// Message still carries suggestion markup like Java RuleMatch.message
	require.Contains(t, ms[0].Message, "HELLO")
}

// Twin of createRuleMatch suppress_misspelled with no remaining suggestions → no match.
func TestCreateRuleMatch_SuppressMisspelledDropsParenSynth(t *testing.T) {
	// ProcessRuleMessage injects <pleasespellme/> while keeping suggestion open tag (attrs OK).
	raw := `Bad <suggestion suppress_misspelled="yes"><match no="1"/></suggestion>`
	msg, matches := ProcessRuleMessage(raw)
	require.Contains(t, msg, PleaseSpellMe)
	require.Contains(t, msg, "<suggestion")
	require.Contains(t, msg, "</suggestion>")
	require.NotEmpty(t, matches)

	// PLEASE_SPELL_ME without any <suggestion> after format → no match (Java createRuleMatch).
	pr := NewPatternRule("SPELL", "en", []*PatternToken{Token("xyzzy")}, "d",
		"bad "+PleaseSpellMe+" only", "")
	sent := testSentence(atr("xyzzy", 0))
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms)

	// With tags + PleaseSpellMe + MistakeMarker body, removeSuppressMisspelled drops suggestion.
	msg2 := `fix <suggestion>` + PleaseSpellMe + `foo` + MistakeMarker + `bar</suggestion> end`
	out := removeSuppressMisspelled(msg2)
	require.NotContains(t, out, PleaseSpellMe)
	require.NotContains(t, out, "<suggestion>")
	require.Equal(t, "fix  end", out)
}
