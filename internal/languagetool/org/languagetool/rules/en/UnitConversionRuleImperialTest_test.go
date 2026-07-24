package en

// Twin of languagetool-language-modules/en/.../UnitConversionRuleImperialTest.java
// Helper: UnitConversionRuleTestHelper.assertMatches (count + suggestion contains).

import (
	"testing"
)

func TestUnitConversionRuleImperial_Match(t *testing.T) {
	// Java: Language lang = Languages.getLanguageForShortCode("en-GB");
	//       JLanguageTool lt = new JLanguageTool(lang);
	//       AbstractUnitConversionRule rule = new UnitConversionRuleImperial(JLanguageTool.getMessageBundle(lang));
	rule := NewUnitConversionRuleImperial(nil)

	// Full Java UnitConversionRuleImperialTest.match() matrix
	assertUnitConversionMatches(t, rule, "I just drank 3 pints.", 1, "1.7 l")
	assertUnitConversionMatches(t, rule, "I am 6 feet (2.02 m) tall.", 1, "1.83 metres")
}
