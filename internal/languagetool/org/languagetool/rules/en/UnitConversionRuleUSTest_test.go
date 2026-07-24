package en

// Twin of languagetool-language-modules/en/.../UnitConversionRuleUSTest.java
// Helper: UnitConversionRuleTestHelper.assertMatches (count + suggestion contains).

import (
	"testing"
)

func TestUnitConversionRuleUS_Match(t *testing.T) {
	// Java: Language lang = Languages.getLanguageForShortCode("en-US");
	//       JLanguageTool lt = new JLanguageTool(lang);
	//       AbstractUnitConversionRule rule = new UnitConversionRuleUS(JLanguageTool.getMessageBundle(lang));
	rule := NewUnitConversionRuleUS(nil)

	// Full Java UnitConversionRuleUSTest.match() matrix
	assertUnitConversionMatches(t, rule, "I just drank 3 pints.", 1, "1.42 l")
	assertUnitConversionMatches(t, rule, "I am 6 feet (2.02 m) tall.", 1, "1.83 meters")
	assertUnitConversionMatches(t, rule, "16 kilometers (10 miles)", 1, "")
	assertUnitConversionMatches(t, rule, "16 kilometers (ca. 10 miles)", 0, "")
	assertUnitConversionMatches(t, rule, "16 kilometers (9.94 mi)", 0, "")
	// matched by antipattern;
	assertUnitConversionMatches(t, rule, "Use 3-5 pounds of butter.", 0, "")
	assertUnitConversionMatches(t, rule, "It's 1/4 mile away", 0, "")
	assertUnitConversionMatches(t, rule, "A 2:1 cup", 0, "")
	assertUnitConversionMatches(t, rule, "1⁄4 cup", 0, "")
	// matched by special case in the case as long as "x ft y inch" is not parsed properly:
	assertUnitConversionMatches(t, rule, "It is 0.76 metres (2 ft 6 inch)", 0, "")
}
