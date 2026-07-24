package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.GenericUnpairedBracketsRuleTest — Java-visible outcomes.
//
// Java setup (setUpRule):
//
//	lt = new JLanguageTool(new FakeLanguage());  // short code "yy"
//	// disable all rules, then:
//	rule = new GenericUnpairedBracketsRule(TestTools.getEnglishMessages(),
//	        Arrays.asList("»"), Arrays.asList("«"));
//	lt.addRule(rule);
//
// FakeLanguage → Language.createDefaultSentenceTokenizer → SimpleSentenceTokenizer.
// assertMatches → lt.check(input).size()
// testRuleMatchPositions → lt.check(...) positions + line/column.

func engMessagesGUBR() map[string]string {
	// MessagesBundle_en keys used by GenericUnpairedBracketsRule.
	return map[string]string{
		"unpaired_brackets":      "Unpaired symbol: ''{0}'' seems to be missing",
		"desc_unpaired_brackets": "Unpaired braces, brackets, quotation marks and similar symbols",
		"category_punctuation":   "Punctuation",
	}
}

// fakeLanguageLTGUBR ports new JLanguageTool(new FakeLanguage()).
func fakeLanguageLTGUBR() *languagetool.JLanguageTool {
	lt := languagetool.NewJLanguageTool("yy")
	lt.SentenceTokenizer = tokenizers.NewSimpleSentenceTokenizer()
	return lt
}

// setUpRuleGUBR ports GenericUnpairedBracketsRuleTest.setUpRule(new FakeLanguage()).
func setUpRuleGUBR() (*GenericUnpairedBracketsRule, *languagetool.JLanguageTool) {
	lt := fakeLanguageLTGUBR()
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	rule := NewGenericUnpairedBracketsRule(engMessagesGUBR(),
		[]string{"\u00bb"}, []string{"\u00ab"})
	lt.AddTextLevelRuleChecker(rule.GetID(), AsTextLevelChecker(rule.MatchList))
	return rule, lt
}

// assertMatchesGUBR ports assertMatches(expected, input) via lt.check(input).
func assertMatchesGUBR(t *testing.T, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(lt.Check(input))
	require.Equal(t, expected, got, "input=%q", input)
}

// matchGUBR ports rule.match(lt.analyzeText(input)) for direct TextLevelRule path.
func matchGUBR(rule *GenericUnpairedBracketsRule, lt *languagetool.JLanguageTool, input string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(input))
}

func TestGenericUnpairedBracketsRule_Rule(t *testing.T) {
	_, lt := setUpRuleGUBR()

	// Java assertMatches(0, ...):
	assertMatchesGUBR(t, lt, 0, "This is \u00bbcorrect\u00ab.")
	assertMatchesGUBR(t, lt, 0, "\u00bbCorrect\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends.\u00ab")
	assertMatchesGUBR(t, lt, 0, "\u00bbCorrect. This is more than one sentence.\u00ab")
	assertMatchesGUBR(t, lt, 0, "\u00bbCorrect. This is more than one sentence.\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends.\u00ab")
	assertMatchesGUBR(t, lt, 0, "\u00bbCorrect\u00ab\n\n\u00bbAnd here it ends.\u00ab\n\nMore text.")
	assertMatchesGUBR(t, lt, 0, "\u00bbCorrect, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertMatchesGUBR(t, lt, 0, "\u00bbCorrect, he said.\n\nThis is the next sentence.\u00ab Here's another sentence.")
	assertMatchesGUBR(t, lt, 0, "\u00bbCorrect, he said.\n\n\n\nThis is the next sentence.\u00ab Here's another sentence.")
	assertMatchesGUBR(t, lt, 0, "This \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertMatchesGUBR(t, lt, 0, "Good.\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertMatchesGUBR(t, lt, 0, "Good.\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertMatchesGUBR(t, lt, 0, "Good.\n\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertMatchesGUBR(t, lt, 0, "This is funny :-)")
	assertMatchesGUBR(t, lt, 0, "This is sad :-( isn't it")
	assertMatchesGUBR(t, lt, 0, "This is funny :)")
	assertMatchesGUBR(t, lt, 0, "This is sad :( isn't it")
	assertMatchesGUBR(t, lt, 0, "a) item one\nb) item two")
	assertMatchesGUBR(t, lt, 0, "a) item one\nb) item two\nc) item three")
	assertMatchesGUBR(t, lt, 0, "\na) item one\nb) item two\nc) item three")
	assertMatchesGUBR(t, lt, 0, "\n\na) item one\nb) item two\nc) item three")
	assertMatchesGUBR(t, lt, 0, "This is a), not b)")
	assertMatchesGUBR(t, lt, 0, "This is it (a, not b) some more test")
	assertMatchesGUBR(t, lt, 0, "This is \u00bbnot an error yet")
	assertMatchesGUBR(t, lt, 0, "See https://de.wikipedia.org/wiki/Schlammersdorf_(Adelsgeschlecht)")

	// Java assertMatches(1, ...):
	assertMatchesGUBR(t, lt, 1, "This is not correct\u00ab")
	assertMatchesGUBR(t, lt, 1, "This is \u00bbnot correct.")
	assertMatchesGUBR(t, lt, 1, "This is \u00bbnot an error yet\n\nBut now it has become one")
	assertMatchesGUBR(t, lt, 1, "This is correct.\n\n\u00bbBut this is not.")
	assertMatchesGUBR(t, lt, 1, "This is correct.\n\nBut this is not\u00ab")
	assertMatchesGUBR(t, lt, 1, "\u00bbThis is correct\u00ab\n\nBut this is not\u00ab")
	assertMatchesGUBR(t, lt, 1, "\u00bbThis is correct\u00ab\n\nBut this \u00bbis\u00ab not\u00ab")
	assertMatchesGUBR(t, lt, 1, "This is not correct. No matter if it's more than one sentence\u00ab")
	assertMatchesGUBR(t, lt, 1, "\u00bbThis is not correct. No matter if it's more than one sentence")
	assertMatchesGUBR(t, lt, 1, "Correct, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertMatchesGUBR(t, lt, 1, "\u00bbCorrect, he said. This is the next sentence. Here's another sentence.")
	assertMatchesGUBR(t, lt, 1, "\u00bbCorrect, he said. This is the next sentence.\n\nHere's another sentence.")
	assertMatchesGUBR(t, lt, 1, "\u00bbCorrect, he said. This is the next sentence.\n\n\n\nHere's another sentence.")
}

func TestGenericUnpairedBracketsRule_RuleMatchPositions(t *testing.T) {
	// Java: setUpRule + lt.check(...).get(0) line/column/from/to
	_, lt := setUpRuleGUBR()

	match1 := lt.Check("This \u00bbis a test.")
	require.NotEmpty(t, match1)
	require.Equal(t, 5, match1[0].FromPos)
	require.Equal(t, 6, match1[0].ToPos)
	require.Equal(t, 0, match1[0].Line)
	require.Equal(t, 0, match1[0].EndLine)
	require.Equal(t, 5, match1[0].Column)
	require.Equal(t, 6, match1[0].EndColumn)

	match2 := lt.Check("This.\nSome stuff.\nIt \u00bbis a test.")
	require.NotEmpty(t, match2)
	require.Equal(t, 21, match2[0].FromPos)
	require.Equal(t, 22, match2[0].ToPos)
	require.Equal(t, 2, match2[0].Line) // first line is 0
	require.Equal(t, 2, match2[0].EndLine)
	require.Equal(t, 4, match2[0].Column)
	require.Equal(t, 5, match2[0].EndColumn)

	// NBSP: Th + nbsp + is + space + »
	match3 := lt.Check("Th\u00ADis \u00bbis a test.")
	require.NotEmpty(t, match3)
	require.Equal(t, 6, match3[0].FromPos)
	require.Equal(t, 7, match3[0].ToPos)
}

func TestGenericUnpairedBracketsRule_AnalyzeTextPath(t *testing.T) {
	// Direct Java rule.match(lt.analyzeText(...)) path (same counts as check).
	rule, lt := setUpRuleGUBR()
	require.Equal(t, "UNPAIRED_BRACKETS", rule.GetID())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.Equal(t, engMessagesGUBR()["desc_unpaired_brackets"], rule.GetDescription())

	require.Equal(t, 0, len(matchGUBR(rule, lt, "This is \u00bbcorrect\u00ab.")))
	require.Equal(t, 1, len(matchGUBR(rule, lt, "This is \u00bbnot correct.")))
	require.Equal(t, 0, len(matchGUBR(rule, lt, "This is \u00bbnot an error yet")))
	require.Equal(t, 1, len(matchGUBR(rule, lt, "This is \u00bbnot an error yet\n\nBut now it has become one")))
}

func TestGenericUnpairedBracketsRule_MessageFormat(t *testing.T) {
	rule, lt := setUpRuleGUBR()
	ms := matchGUBR(rule, lt, "This is \u00bbnot correct.")
	require.NotEmpty(t, ms)
	// MessageFormat: Unpaired symbol: ''{0}'' seems to be missing → quotes around missing symbol «
	require.Equal(t, "Unpaired symbol: '\u00ab' seems to be missing", ms[0].GetMessage())
}
