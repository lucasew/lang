package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.GenericUnpairedQuotesRuleTest — Java-visible outcomes.
//
// Java setup (setUpRule):
//
//	lt = new JLanguageTool(new FakeLanguage());  // short code "yy"
//	// disable all rules, then:
//	rule = new GenericUnpairedQuotesRule(TestTools.getEnglishMessages(),
//	        Arrays.asList("»", "«", "\"", "'", "›", "‹"),
//	        Arrays.asList("«", "»", "\"", "'", "‹", "›"));
//	lt.addRule(rule);
//
// FakeLanguage → Language.createDefaultSentenceTokenizer → SimpleSentenceTokenizer.
// assertMatches → lt.check(input).size()
// testRuleMatchPositions → lt.check(...) positions + line/column.

func engMessagesGUQR() map[string]string {
	// MessagesBundle_en keys used by GenericUnpairedQuotesRule.
	return map[string]string{
		"unpaired_brackets":      "Unpaired symbol: ''{0}'' seems to be missing",
		"desc_unpaired_brackets": "Unpaired braces, brackets, quotation marks and similar symbols",
		"desc_unpaired_quotes":   "Unpaired quotation marks",
		"category_punctuation":   "Punctuation",
	}
}

// fakeLanguageLTGUQR ports new JLanguageTool(new FakeLanguage()).
func fakeLanguageLTGUQR() *languagetool.JLanguageTool {
	lt := languagetool.NewJLanguageTool("yy")
	lt.SentenceTokenizer = tokenizers.NewSimpleSentenceTokenizer()
	return lt
}

// setUpRuleGUQR ports GenericUnpairedQuotesRuleTest.setUpRule(new FakeLanguage()).
func setUpRuleGUQR() (*GenericUnpairedQuotesRule, *languagetool.JLanguageTool) {
	lt := fakeLanguageLTGUQR()
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	rule := NewGenericUnpairedQuotesRule(engMessagesGUQR(),
		[]string{"\u00bb", "\u00ab", "\"", "'", "\u203a", "\u2039"},
		[]string{"\u00ab", "\u00bb", "\"", "'", "\u2039", "\u203a"})
	lt.AddTextLevelRuleChecker(rule.GetID(), AsTextLevelChecker(rule.MatchList))
	return rule, lt
}

// assertMatchesGUQR ports assertMatches(expected, input) via lt.check(input).
func assertMatchesGUQR(t *testing.T, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(lt.Check(input))
	require.Equal(t, expected, got, "input=%q", input)
}

// matchGUQR ports rule.match(lt.analyzeText(input)) for direct TextLevelRule path.
func matchGUQR(rule *GenericUnpairedQuotesRule, lt *languagetool.JLanguageTool, input string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(input))
}

func TestGenericUnpairedQuotesRule_Rule(t *testing.T) {
	_, lt := setUpRuleGUQR()

	// Java assertMatches(0, ...):
	assertMatchesGUQR(t, lt, 0, "This is \u00bbcorrect\u00ab.")
	assertMatchesGUQR(t, lt, 0, "This is \u00abcorrect\u00bb.")
	assertMatchesGUQR(t, lt, 0, "\u00bbCorrect\u00ab\n\u00bbAnd \u203ahere\u2039 it ends.\u00ab")
	assertMatchesGUQR(t, lt, 0, "\u00abCorrect\u00bb\n\u00abAnd \u2039here\u203a it ends.\u00bb")
	assertMatchesGUQR(t, lt, 0, "\u00bbCorrect. This is more than one sentence.\u00ab")
	assertMatchesGUQR(t, lt, 0, "\u00bbCorrect. This is more than one sentence.\u00ab\n\u00bbAnd \u203ahere\u2039 it ends.\u00ab")
	assertMatchesGUQR(t, lt, 0, "\u00bbCorrect\u00ab\n\n\u00bbAnd here it ends.\u00ab\n\nMore text.")
	assertMatchesGUQR(t, lt, 0, "\u00bbCorrect, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertMatchesGUQR(t, lt, 0, "\u00bbCorrect\u00ab, he said.\n\n\u00bbThis is the next sentence.\u00ab \u00bbHere's another sentence.\u00ab")
	assertMatchesGUQR(t, lt, 0, "\u00bbCorrect\u00ab, he said. \u00bbThis is the next sentence.\u00ab\n\u00bbHere's another sentence.\u00ab")
	assertMatchesGUQR(t, lt, 0, "This \u00bbis also \u203acorrect\u2039\u00ab.")
	assertMatchesGUQR(t, lt, 0, "Good.\n\nThis \u00bbis also \u203acorrect\u2039\u00ab.")
	assertMatchesGUQR(t, lt, 0, "Good.\n\n\nThis \u00bbis also \u203acorrect\u2039\u00ab.")
	assertMatchesGUQR(t, lt, 0, "Good.\n\n\n\nThis \u00bbis also \u203acorrect\u2039\u00ab.")

	assertMatchesGUQR(t, lt, 0, "This is \"correct\".")
	assertMatchesGUQR(t, lt, 0, "This is 'correct'.")
	assertMatchesGUQR(t, lt, 0, "\"Correct\"\n\"And 'here' it ends.\"")
	assertMatchesGUQR(t, lt, 0, "'Correct'\n'And \"here\" it ends.'")
	assertMatchesGUQR(t, lt, 0, "This \"is also 'correct'\".")
	assertMatchesGUQR(t, lt, 0, "This isn't \"incorrect\".")
	assertMatchesGUQR(t, lt, 0, "\"This isn't incorrect.\"")
	assertMatchesGUQR(t, lt, 0, "The screen is 20\" wide.")
	assertMatchesGUQR(t, lt, 0, "\"The screen is 20\" wide.\"")

	// Java assertMatches(1, ...):
	assertMatchesGUQR(t, lt, 1, "This is not correct\u00ab")
	assertMatchesGUQR(t, lt, 1, "This is \u00bbnot correct.")
	assertMatchesGUQR(t, lt, 1, "This is \u00bbnot correct")
	assertMatchesGUQR(t, lt, 1, "This is \u00bbnot an error yet\n\nBut now it has become one")
	assertMatchesGUQR(t, lt, 1, "This is correct.\n\n\u00bbBut this is not.")
	assertMatchesGUQR(t, lt, 1, "This is correct.\n\nBut this is not\u00ab")
	assertMatchesGUQR(t, lt, 1, "\u00bbThis is correct\u00ab\n\nBut this is not\u00ab")
	assertMatchesGUQR(t, lt, 1, "\u00bbThis is correct\u00ab\n\nBut this \u00bbis\u00ab not\u00ab")
	assertMatchesGUQR(t, lt, 1, "This is not correct. No matter if it's more than one sentence\u00ab")
	assertMatchesGUQR(t, lt, 1, "\u00bbThis is not correct. No matter if it's more than one sentence")
	assertMatchesGUQR(t, lt, 1, "Correct, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertMatchesGUQR(t, lt, 1, "\u00bbCorrect, he said. This is the next sentence. Here's another sentence.")
	assertMatchesGUQR(t, lt, 1, "\u00bbCorrect, he said. This is the next sentence.\n\nHere's another sentence.")
	assertMatchesGUQR(t, lt, 1, "\u00bbCorrect, he said. This is the next sentence.\n\n\n\nHere's another sentence.")

	// Java assertMatches(2, ...):
	assertMatchesGUQR(t, lt, 2, "\u00bbCorrect\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends with two matches.\u00ab")
	assertMatchesGUQR(t, lt, 2, "\u00bbCorrect. This is more than one sentence.\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends.\u00ab")
	assertMatchesGUQR(t, lt, 2, "Good.\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertMatchesGUQR(t, lt, 2, "Good.\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertMatchesGUQR(t, lt, 2, "Good.\n\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
}

func TestGenericUnpairedQuotesRule_RuleMatchPositions(t *testing.T) {
	// Java: setUpRule + lt.check(...).get(0) line/column/from/to
	_, lt := setUpRuleGUQR()

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

func TestGenericUnpairedQuotesRule_AnalyzeTextPath(t *testing.T) {
	// Direct Java rule.match(lt.analyzeText(...)) path (same counts as check).
	rule, lt := setUpRuleGUQR()
	require.Equal(t, "UNPAIRED_QUOTES", rule.GetID())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.Equal(t, engMessagesGUQR()["desc_unpaired_quotes"], rule.GetDescription())

	require.Equal(t, 0, len(matchGUQR(rule, lt, "This is \u00bbcorrect\u00ab.")))
	require.Equal(t, 1, len(matchGUQR(rule, lt, "This is \u00bbnot correct.")))
	require.Equal(t, 1, len(matchGUQR(rule, lt, "This is \u00bbnot an error yet\n\nBut now it has become one")))
	require.Equal(t, 0, len(matchGUQR(rule, lt, "The screen is 20\" wide.")))
	require.Equal(t, 0, len(matchGUQR(rule, lt, "This isn't \"incorrect\".")))
}

func TestGenericUnpairedQuotesRule_MessageFormat(t *testing.T) {
	rule, lt := setUpRuleGUQR()
	ms := matchGUQR(rule, lt, "This is \u00bbnot correct.")
	require.NotEmpty(t, ms)
	// MessageFormat: Unpaired symbol: ''{0}'' seems to be missing → quotes around missing symbol «
	require.Equal(t, "Unpaired symbol: '\u00ab' seems to be missing", ms[0].GetMessage())
}

func TestGenericUnpairedQuotesRule_DefaultCtor(t *testing.T) {
	// Java: GenericUnpairedQuotesRule(messages) → “” "" ‘’ ''
	rule := NewGenericUnpairedQuotesRuleDefault(engMessagesGUQR())
	require.Equal(t, "UNPAIRED_QUOTES", rule.GetID())
	require.Equal(t, []string{"“", "\"", "‘", "'"}, rule.StartSymbols)
	require.Equal(t, []string{"”", "\"", "’", "'"}, rule.EndSymbols)
	require.Equal(t, ITSTypographical, rule.GetLocQualityIssueType())
}

func TestGenericUnpairedQuotesRule_WithID(t *testing.T) {
	id := "CUSTOM_QUOTES"
	rule := NewGenericUnpairedQuotesRuleWithID(&id, engMessagesGUQR(),
		[]string{"»"}, []string{"«"})
	require.Equal(t, "CUSTOM_QUOTES", rule.GetID())
	require.Equal(t, "«", rule.FindCorrespondingSymbol("»"))
	require.Equal(t, "»", rule.FindCorrespondingSymbol("«"))
}
