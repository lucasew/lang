package ga

// Outcome twin of org.languagetool.rules.ga.EnglishHomophoneRule
// (no dedicated Java *RuleTest class; bar matches People / Spaces / Logainm ACCEPT:
// full metadata + Match outcomes + exact getMessage text).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// Official /ga/english-homophones.txt keys are English surfaces (e.g. sushi, well).
// Java does not call setIgnoreTaggedWords (false).

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official english-homophones.txt rows.
func TestEnglishHomophoneRule_Rule(t *testing.T) {
	// Java: new EnglishHomophoneRule(messages, language)
	rule := NewEnglishHomophoneRule(nil)

	// correct sentences (Irish forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("An bhialann súisí sin ba chúis leis."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("bhuel anseo."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair eochair-nil."))))

	// Java example: sushi → súisí
	checkEnglishHomophone(t, rule, "An bhialann sushi sin ba chúis leis.", "sushi", "súisí")

	// Official english-homophones.txt rows
	checkEnglishHomophone(t, rule, "well anseo.", "well", "bhuel")
	checkEnglishHomophone(t, rule, "tofu anseo.", "tofu", "tófú")
	checkEnglishHomophone(t, rule, "yoga anseo.", "yoga", "ióga")
	checkEnglishHomophone(t, rule, "bear anseo.", "bear", "béar")
	checkEnglishHomophone(t, rule, "bingo anseo.", "bingo", "biongó")
	checkEnglishHomophone(t, rule, "wow anseo.", "wow", "bhabh")
	checkEnglishHomophone(t, rule, "whoa anseo.", "whoa", "bhó")
	checkEnglishHomophone(t, rule, "pumpkin anseo.", "pumpkin", "puimcín")
	checkEnglishHomophone(t, rule, "turmeric anseo.", "turmeric", "tuirmiric")
	checkEnglishHomophone(t, rule, "homophone anseo.", "homophone", "homafón")
	checkEnglishHomophone(t, rule, "impression anseo.", "impression", "imprisean")
	checkEnglishHomophone(t, rule, "trampoline anseo.", "trampoline", "trampailín")

	// multi-key same replacement: miaow|meow=meamh
	checkEnglishHomophone(t, rule, "miaow anseo.", "miaow", "meamh")
	checkEnglishHomophone(t, rule, "meow anseo.", "meow", "meamh")

	// multi-key same replacement: hummus|humus=húmas
	checkEnglishHomophone(t, rule, "hummus anseo.", "hummus", "húmas")
	checkEnglishHomophone(t, rule, "humus anseo.", "humus", "húmas")

	// Java-visible createRuleMatch message/short for example-pair hit
	// (message uses both tokenStr and replacements; no trailing period)
	matches := rule.Match(languagetool.AnalyzePlain("An bhialann sushi sin ba chúis leis."))
	require.Len(t, matches, 1)
	require.Equal(t, "Homafón Béarla.", matches[0].GetShortMessage())
	require.Equal(t,
		"Is é súisí an litriú Gaelach ar \"sushi\"",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, EnglishHomophoneRuleID, base.GetID())
}

// checkEnglishHomophone asserts one match with the given first replacement and exact message.
// tokenStr is the surface form Java passes to getMessage.
func checkEnglishHomophone(t *testing.T, rule *EnglishHomophoneRule, sentence, tokenStr, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	// Java: "Is é " + join(replacements) + " an litriú Gaelach ar \"" + tokenStr + "\""
	// Message is built before uppercaseFirstChar on suggestions; single-rep keys use word as join.
	require.Equal(t,
		"Is é "+word+" an litriú Gaelach ar \""+tokenStr+"\"",
		matches[0].GetMessage(),
		"message for %q", sentence)
	require.Equal(t, "Homafón Béarla.", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, EnglishHomophoneRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of EnglishHomophoneRule constructor + overrides.
func TestEnglishHomophoneRule_Metadata(t *testing.T) {
	rule := NewEnglishHomophoneRule(nil)

	require.Equal(t, EnglishHomophoneRuleID, rule.GetID())
	require.Equal(t, "GA_ENGLISH_HOMOPHONE", rule.GetID())
	require.Equal(t, "Is homafóin iad na focail, m.sh., \"well\" agus \"bhuel\"", rule.GetDescription())
	require.Equal(t, "Homafón Béarla.", rule.GetShort())
	// MessageFn with empty token + empty replacements (static surface for metadata probes)
	// Java: "Is é " + join([]) + " an litriú Gaelach ar \"" + "" + "\""
	// → "Is é  an litriú Gaelach ar \"\""  (double space after é)
	require.Equal(t, "Is é  an litriú Gaelach ar \"\"", rule.GetMessage())

	// Java getLocale() → Locale("GA")
	require.Equal(t, "GA", rule.GetLocaleCode())
	require.Equal(t, "GA", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords) // Java does not call setIgnoreTaggedWords
	require.Equal(t, "", rule.GetURL())

	// Java: Categories.TYPOS + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /ga/english-homophones.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"súisí"}, ww["sushi"])
	require.Equal(t, []string{"bhuel"}, ww["well"])
	require.Equal(t, []string{"meamh"}, ww["miaow"])
	require.Equal(t, []string{"meamh"}, ww["meow"])
	require.Equal(t, []string{"húmas"}, ww["hummus"])
	require.Equal(t, []string{"húmas"}, ww["humus"])
	require.Equal(t, []string{"tófú"}, ww["tofu"])
	require.Equal(t, []string{"ióga"}, ww["yoga"])
	require.Equal(t, []string{"béar"}, ww["bear"])
	require.Equal(t, []string{"homafón"}, ww["homophone"])
	require.Equal(t, EnglishHomophoneFilePath, "/ga/english-homophones.txt")
	// 17 data lines → 19 keys (miaow|meow and hummus|humus each add an extra key)
	require.Equal(t, 19, len(ww), "all english-homophones.txt keys must load")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewEnglishHomophoneRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Is homafóin iad na focail, m.sh., \"well\" agus \"bhuel\"", ruleMsgs.GetDescription())
	require.Equal(t, "Homafón Béarla.", ruleMsgs.GetShort())

	// Java addExamplePair(sushi → súisí)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "An bhialann <marker>sushi</marker> sin ba chúis leis.", inc[0].GetExample())
	require.Equal(t, []string{"súisí"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "An bhialann <marker>súisí</marker> sin ba chúis leis.", corr[0].GetExample())
}

// Load does not panic; resource parses.
func TestEnglishHomophoneRule_Loads(t *testing.T) {
	rule := NewEnglishHomophoneRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
}

func TestEnglishHomophoneRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := englishHomophonesFS.ReadFile("data/english-homophones.txt")
	require.NoError(t, err)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// this file: .../internal/languagetool/org/languagetool/rules/ga/*.go
	// seven Dir() steps reach repo root (ga→rules→languagetool→org→languagetool→internal→repo)
	repoRoot := thisFile
	for i := 0; i < 7; i++ {
		repoRoot = filepath.Dir(repoRoot)
	}
	official := filepath.Join(repoRoot,
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/english-homophones.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/english-homophones.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/english-homophones.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/english-homophones.txt must be byte-identical to official Java resource")

	m, err := rules.LoadSimpleReplaceWords(bytes.NewReader(embedded))
	require.NoError(t, err)
	require.Equal(t, []string{"súisí"}, m["sushi"])
	require.Equal(t, []string{"bhuel"}, m["well"])
	require.Equal(t, []string{"meamh"}, m["miaow"])
	require.Equal(t, []string{"meamh"}, m["meow"])
	require.Equal(t, []string{"húmas"}, m["hummus"])
	require.Equal(t, []string{"húmas"}, m["humus"])
	require.Equal(t, 19, len(m))
}
