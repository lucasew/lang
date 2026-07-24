package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Behavior matrix for org.languagetool.rules.WhiteSpaceAtBeginOfParagraph
// (no Java WhiteSpaceAtBeginOfParagraphTest). Outcomes twin Java match / isWhitespaceDel / metadata.
//
// Java setup pattern (sentence-level Rule siblings):
//
//	JLanguageTool lt = new JLanguageTool(TestTools.getDemoLanguage()); // or en-US
//	WhiteSpaceAtBeginOfParagraph rule = new WhiteSpaceAtBeginOfParagraph(messages); // defaultActive false
//	rule.match(lt.getAnalyzedSentence(input)) / sum over lt.analyzeText(input)
//
// Go: demoLT() / NewJLanguageTool("en") + Match(GetAnalyzedSentence) / Match over Analyze.

func engMessagesWSBegin() map[string]string {
	// MessagesBundle_en keys used by WhiteSpaceAtBeginOfParagraph (Java typo "parapgraph").
	return map[string]string{
		"whitespace_at_begin_parapgraph_msg":  "Please delete the space character at the beginning of the paragraph.",
		"whitespace_at_begin_parapgraph_desc": "Space character at the beginning of paragraph",
		"category_style":                      "Style",
	}
}

func setUpWSBeginRule() (*WhiteSpaceAtBeginOfParagraph, *languagetool.JLanguageTool) {
	// Java: new WhiteSpaceAtBeginOfParagraph(messages) with Demo.
	rule := NewWhiteSpaceAtBeginOfParagraph(engMessagesWSBegin())
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())
	return rule, lt
}

// matchWSBegin ports rule.match(lt.getAnalyzedSentence(text)).
func matchWSBegin(rule *WhiteSpaceAtBeginOfParagraph, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.Match(lt.GetAnalyzedSentence(text))
}

// matchWSBeginAnalyze ports sum of rule.match over lt.analyzeText(text).
func matchWSBeginAnalyze(rule *WhiteSpaceAtBeginOfParagraph, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	var out []*RuleMatch
	for _, s := range lt.Analyze(text) {
		out = append(out, rule.Match(s)...)
	}
	return out
}

func assertWSBeginMatches(t *testing.T, rule *WhiteSpaceAtBeginOfParagraph, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(matchWSBegin(rule, lt, input))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestWhiteSpaceAtBeginOfParagraph_Metadata(t *testing.T) {
	// Java WhiteSpaceAtBeginOfParagraph(messages) → defaultActive false
	rule := NewWhiteSpaceAtBeginOfParagraph(engMessagesWSBegin())
	require.Equal(t, "WHITESPACE_PARAGRAPH_BEGIN", rule.GetID())
	require.Equal(t, "Space character at the beginning of paragraph", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.True(t, rule.IsOfficeDefaultOn())
	require.True(t, rule.GetCategory().GetID().Equals(NewCategoryId("STYLE")))
	require.Equal(t, ITSStyle, rule.GetLocQualityIssueType())
}

func TestWhiteSpaceAtBeginOfParagraph_DefaultActive(t *testing.T) {
	// Java 2-arg: defaultActive false → setDefaultOff(); true leaves default on.
	off := NewWhiteSpaceAtBeginOfParagraphDefaultActive(engMessagesWSBegin(), false)
	require.True(t, off.IsDefaultOff())
	require.True(t, off.IsOfficeDefaultOn())
	on := NewWhiteSpaceAtBeginOfParagraphDefaultActive(engMessagesWSBegin(), true)
	require.False(t, on.IsDefaultOff())
	require.True(t, on.IsOfficeDefaultOn())
}

func TestWhiteSpaceAtBeginOfParagraph_IsWhitespaceDel(t *testing.T) {
	// Ports WhiteSpaceAtBeginOfParagraph.isWhitespaceDel private helper via Analyze tokens.
	rule, lt := setUpWSBeginRule()
	// Leading space: deletable
	s := lt.GetAnalyzedSentence(" Hello.")
	toks := s.GetTokens()
	require.GreaterOrEqual(t, len(toks), 3)
	require.True(t, rule.isWhitespaceDel(toks[1]), "space should be deletable")
	require.False(t, rule.isWhitespaceDel(toks[2]), "Hello not whitespace")

	// ZWSP: whitespace but excluded (LO/OO function markers)
	sZW := lt.GetAnalyzedSentence("\u200BHello.")
	toksZW := sZW.GetTokens()
	require.True(t, toksZW[1].IsWhitespace())
	require.Equal(t, "\u200B", toksZW[1].GetToken())
	require.False(t, rule.isWhitespaceDel(toksZW[1]), "ZWSP must not be deletable")

	// Linebreak: whitespace but excluded
	sLB := lt.GetAnalyzedSentence("\nHello.")
	toksLB := sLB.GetTokens()
	require.True(t, toksLB[1].IsLinebreak())
	require.False(t, rule.isWhitespaceDel(toksLB[1]), "linebreak must not be deletable")
}

func TestWhiteSpaceAtBeginOfParagraph_Match_Demo(t *testing.T) {
	// Java-visible match counts / spans / suggestions via Demo analyze.
	rule, lt := setUpWSBeginRule()

	// No match without leading whitespace after SENT_START.
	assertWSBeginMatches(t, rule, lt, 0, "Hello world.")
	assertWSBeginMatches(t, rule, lt, 0, "Hello  world.") // mid-sentence multi-space only

	// Match: leading space(s) / tab after SENT_START.
	assertWSBeginMatches(t, rule, lt, 1, " Hello.")
	assertWSBeginMatches(t, rule, lt, 1, "  Hello world.")
	assertWSBeginMatches(t, rule, lt, 1, "\tHello.")

	// Span + suggested replacement (Java: from tokens[1].startPos to tokens[i].endPos,
	// suggestedReplacement = tokens[i].getToken() — the first non-deletable token).
	m1 := matchWSBegin(rule, lt, " Hello.")
	require.Len(t, m1, 1)
	require.Equal(t, 0, m1[0].GetFromPos())
	require.Equal(t, 6, m1[0].GetToPos()) // " Hello" → start of space through end of Hello
	require.Equal(t, []string{"Hello"}, m1[0].GetSuggestedReplacements())
	require.Equal(t, engMessagesWSBegin()["whitespace_at_begin_parapgraph_msg"], m1[0].GetMessage())
	require.Equal(t, "WHITESPACE_PARAGRAPH_BEGIN", rule.GetID())

	m2 := matchWSBegin(rule, lt, "  Hello world.")
	require.Len(t, m2, 1)
	require.Equal(t, 0, m2[0].GetFromPos())
	require.Equal(t, 7, m2[0].GetToPos()) // two spaces + "Hello"
	require.Equal(t, []string{"Hello"}, m2[0].GetSuggestedReplacements())

	m3 := matchWSBegin(rule, lt, "\tHello.")
	require.Len(t, m3, 1)
	require.Equal(t, 0, m3[0].GetFromPos())
	require.Equal(t, 6, m3[0].GetToPos())
	require.Equal(t, []string{"Hello"}, m3[0].GetSuggestedReplacements())
}

func TestWhiteSpaceAtBeginOfParagraph_Match_ExcludeZWSP(t *testing.T) {
	// Java isWhitespaceDel: !token.equals("\u200B") — pure ZWSP begin does not match.
	rule, lt := setUpWSBeginRule()
	assertWSBeginMatches(t, rule, lt, 0, "\u200BHello.")
	assertWSBeginMatches(t, rule, lt, 0, "\u200B")

	// Spaces before ZWSP: loop stops at ZWSP (not deletable) with i>1 → match;
	// suggestedReplacement is the ZWSP token (Java bug-for-bug).
	m := matchWSBegin(rule, lt, "  \u200BHello.")
	require.Len(t, m, 1)
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 3, m[0].GetToPos()) // two spaces + ZWSP
	require.Equal(t, []string{"\u200B"}, m[0].GetSuggestedReplacements())
}

func TestWhiteSpaceAtBeginOfParagraph_Match_ExcludeLinebreak(t *testing.T) {
	// Pure linebreak / leading linebreak cases: no match (isWhitespaceDel false for lb,
	// or tokens[i].isLinebreak() after skipping spaces).
	rule, lt := setUpWSBeginRule()
	assertWSBeginMatches(t, rule, lt, 0, "\n")
	assertWSBeginMatches(t, rule, lt, 0, "\nHello.")
	assertWSBeginMatches(t, rule, lt, 0, "\n Hello.") // first token after SENT_START is linebreak
	assertWSBeginMatches(t, rule, lt, 0, " \nHello.") // i lands on linebreak → !isLinebreak fails
	assertWSBeginMatches(t, rule, lt, 0, "  \n")      // only spaces then linebreak

	// Only whitespace (no non-ws token after): i == tokens.length → no match.
	assertWSBeginMatches(t, rule, lt, 0, "   ")
	assertWSBeginMatches(t, rule, lt, 0, " ")
	assertWSBeginMatches(t, rule, lt, 0, "\t")
}

func TestWhiteSpaceAtBeginOfParagraph_Match_EN_JavaTwin(t *testing.T) {
	// Same control flow under EN analyze (real word tokenizer path).
	rule := NewWhiteSpaceAtBeginOfParagraph(engMessagesWSBegin())
	lt := languagetool.NewJLanguageTool("en")

	assertWSBeginMatches(t, rule, lt, 0, "Hello world.")
	assertWSBeginMatches(t, rule, lt, 1, " Hello.")
	assertWSBeginMatches(t, rule, lt, 1, "  Hello world.")
	assertWSBeginMatches(t, rule, lt, 0, "\u200BHello.")
	assertWSBeginMatches(t, rule, lt, 0, "\nHello.")
	assertWSBeginMatches(t, rule, lt, 0, " \nHello.")
	assertWSBeginMatches(t, rule, lt, 0, "   ")

	m := matchWSBegin(rule, lt, "  Hello world.")
	require.Len(t, m, 1)
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 7, m[0].GetToPos())
	require.Equal(t, []string{"Hello"}, m[0].GetSuggestedReplacements())
	require.Equal(t, engMessagesWSBegin()["whitespace_at_begin_parapgraph_msg"], m[0].GetMessage())
}

func TestWhiteSpaceAtBeginOfParagraph_Match_AnalyzeMultiSentence(t *testing.T) {
	// Sum of sentence-level matches over analyzeText (lt.check-style with only this rule).
	rule, lt := setUpWSBeginRule()
	// First sentence has no leading ws; second sentence after \n\n may retain leading spaces
	// depending on SRX — assert total via Analyze path.
	require.Equal(t, 1, len(matchWSBeginAnalyze(rule, lt, "  Hello world.")))
	require.Equal(t, 0, len(matchWSBeginAnalyze(rule, lt, "Hello world.")))
	// Mid-text double space is not paragraph-begin after SENT_START of that sentence.
	require.Equal(t, 0, len(matchWSBeginAnalyze(rule, lt, "Hello  world.")))
}
