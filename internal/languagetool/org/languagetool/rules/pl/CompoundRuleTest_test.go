package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/CompoundRuleTest.java
//
// Java setUp:
//   lt = new JLanguageTool(new Polish());
//   rule = new CompoundRule(TestTools.getEnglishMessages(), new Polish(), null);
// Java testRule → AbstractCompoundRuleTest.check(expected, text, suggestions…)
// via rule.match(lt.getAnalyzedSentence(text)).
// Java testCompoundFile → MorfologikPolishSpellerRule + getFromResourceDirAsLines("/pl/compounds.txt")
// with +/* /plain spelling assertions (real PL analysis).

import (
	"bufio"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Metadata(t *testing.T) {
	r := NewCompoundRule(nil)
	// Java getId / getDescription
	require.Equal(t, "PL_COMPOUNDS", r.GetID())
	require.Equal(t, "Sprawdza wyrazy z łącznikiem, np. „łapu capu” zamiast „łapu-capu”", r.GetDescription())
	// Java AbstractCompoundRule ctor: Categories.MISC + ITSIssueType.Misspelling
	require.Equal(t, rules.CatMisc.GetCategory(nil).GetID(), r.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
	// Java 4-arg super shortMessage + PL messages
	require.Equal(t, "Brak łącznika lub zbędny łącznik", r.ShortDesc)
	require.Equal(t, "Ten wyraz pisze się z łącznikiem.", r.WithHyphenMessage)
	require.Equal(t, "Ten wyraz pisze się razem (bez spacji ani łącznika).", r.WithoutHyphenMessage)
	require.Equal(t, "Ten wyraz pisze się z łącznikiem lub bez niego.", r.WithOrWithoutHyphenMessage)
	// Java AbstractCompoundRule default / field true
	require.True(t, r.SentenceStartsWithUpperCase)
	// Java getCompoundRuleData singleton from /pl/compounds.txt
	require.NotNil(t, r.GetCompoundRuleData())
	require.Same(t, loadCompoundData(), r.GetCompoundRuleData())
	// Java example pair
	require.Equal(t, []string{"Rabce-Zdroju"}, r.GetIncorrectExamples()[0].GetCorrections())
}

func TestCompoundRule_Rule(t *testing.T) {
	// Java: lt = new JLanguageTool(new Polish()); rule = new CompoundRule(...);
	// Analysis path matches analyzePLSpeller (tagger + hybrid disambiguator when present).
	rule := NewCompoundRule(nil)

	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		// Java AbstractCompoundRuleTest.check → rule.match(lt.getAnalyzedSentence(text))
		sent := analyzePLSpeller(t, text)
		matches := rule.Match(sent)
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
		}
	}

	// correct sentences:
	check(0, "Nie róbmy nic na łapu-capu.")
	check(0, "Jedzmy kogel-mogel.")
	// incorrect sentences:
	check(1, "bim bom", "bim-bom")
}

// Port of CompoundRuleTest.testCompoundFile — full loop over official /pl/compounds.txt
// with real MorfologikPolishSpellerRule + real PL analysis (Java JLanguageTool(Polish)).
// Assertions match Java +/* /plain branches exactly. No soft load-only smoke.
func TestCompoundRule_CompoundFile(t *testing.T) {
	if morfologik.DiscoverLanguageDict(PolishSpellerDict) == "" {
		t.Skip("pl_PL.dict not discoverable (inspiration/.../pl/hunspell)")
	}
	// Java: new MorfologikPolishSpellerRule(TestTools.getMessages("pl"), new Polish(), null, Collections.emptyList())
	spellRule := NewMorfologikPolishSpellerRule()
	require.NotNil(t, spellRule)
	require.NotNil(t, spellRule.Speller1, "InitSpellersFromGetters must attach multi when dict present")

	// Java: JLanguageTool.getDataBroker().getFromResourceDirAsLines("/pl/compounds.txt")
	// Embed is byte-identical to official resource/pl/compounds.txt; path law /pl/compounds.txt.
	raw, err := compoundsFS.ReadFile("data/compounds.txt")
	require.NoError(t, err)

	// Java setUp also creates lt = new JLanguageTool(new Polish()) for getAnalyzedSentence.
	// analyzePLSpeller is the colocated twin of that path (MorfologikPolishSpellerRuleTest).

	sc := bufio.NewScanner(strings.NewReader(string(raw)))
	checked := 0
	for sc.Scan() {
		line := sc.Text()
		// Java: if (line.isEmpty() || line.charAt(0) == '#') continue;
		if line == "" || line[0] == '#' {
			continue
		}
		if strings.HasSuffix(line, "+") {
			// Java: removeLastCharacter; replace('-', ' '); spellRule.match → 0
			line = line[:len(line)-1]
			line = strings.ReplaceAll(line, "-", " ")
			ms, err := spellRule.Match(analyzePLSpeller(t, line))
			require.NoError(t, err)
			require.Equal(t, 0, len(ms), "The entry: %s is not found in the spelling dictionary!", line)
			checked++
		} else if strings.HasSuffix(line, "*") {
			// Java: removeLastCharacter; spellRule.match → 0
			line = line[:len(line)-1]
			ms, err := spellRule.Match(analyzePLSpeller(t, line))
			require.NoError(t, err)
			require.Equal(t, 0, len(ms), "The entry: %s is not found in the spelling dictionary!", line)
			checked++
		} else {
			// Java: hyphenated form + joined (no hyphen) form both in dict
			ms, err := spellRule.Match(analyzePLSpeller(t, line))
			require.NoError(t, err)
			require.Equal(t, 0, len(ms), "The entry: %s is not found in the spelling dictionary!", line)
			joined := strings.ReplaceAll(line, "-", "")
			ms2, err := spellRule.Match(analyzePLSpeller(t, joined))
			require.NoError(t, err)
			require.Equal(t, 0, len(ms2), "The entry: %s is not found in the spelling dictionary!", joined)
			checked++
		}
	}
	require.NoError(t, sc.Err())
	// non-empty official file — guard against accidental empty embed
	require.Greater(t, checked, 0, "expected non-empty compounds.txt entries")
	// CompoundRule loads the same official resource (Java rule uses same path)
	require.NotNil(t, NewCompoundRule(nil).GetCompoundRuleData())
}
