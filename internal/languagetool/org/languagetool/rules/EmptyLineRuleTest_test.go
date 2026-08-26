package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Behavior matrix for org.languagetool.rules.EmptyLineRule (no Java EmptyLineRuleTest).
// Outcomes twin Java EmptyLineRule.match / helpers + Demo/EN analyze paths.
//
// Java setup pattern (sibling text-level rules):
//
//	JLanguageTool lt = new JLanguageTool(TestTools.getDemoLanguage()); // or en-US
//	EmptyLineRule rule = new EmptyLineRule(messages, lang); // defaultActive false
//	rule.match(lt.analyzeText(input))
//
// Go: demoLT() / NewJLanguageTool("en") + MatchList(lt.Analyze).

func engMessagesEmptyLine() map[string]string {
	// MessagesBundle_en keys used by EmptyLineRule.
	return map[string]string{
		"empty_line_rule_msg":  "Please delete empty line below and use formatting instead",
		"empty_line_rule_desc": "Empty Line",
		"category_style":       "Style",
	}
}

func setUpEmptyLineRule() (*EmptyLineRule, *languagetool.JLanguageTool) {
	// Java: new EmptyLineRule(messages, lang) with Demo → multi-line paragraph marks (false).
	rule := NewEmptyLineRule(engMessagesEmptyLine())
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())
	return rule, lt
}

// matchEmptyLine ports rule.match(lt.analyzeText(text)).
func matchEmptyLine(rule *EmptyLineRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func assertEmptyLineMatches(t *testing.T, rule *EmptyLineRule, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(matchEmptyLine(rule, lt, input))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestEmptyLineRule_Metadata(t *testing.T) {
	// Java EmptyLineRule(messages, lang) → defaultActive false
	rule := NewEmptyLineRule(engMessagesEmptyLine())
	require.Equal(t, "EMPTY_LINE", rule.GetID())
	require.Equal(t, "Empty Line", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.True(t, rule.IsOfficeDefaultOn())
	require.Equal(t, 1, rule.MinToCheckParagraph())
	require.True(t, rule.GetCategory().GetID().Equals(NewCategoryId("STYLE")))
	require.Equal(t, ITSStyle, rule.GetLocQualityIssueType())
	require.False(t, rule.SingleLineBreaksMarksPara) // Demo/SRX default
}

func TestEmptyLineRule_DefaultActive(t *testing.T) {
	// Java 3-arg: defaultActive false → setDefaultOff(); true leaves default on.
	off := NewEmptyLineRuleDefaultActive(engMessagesEmptyLine(), false)
	require.True(t, off.IsDefaultOff())
	on := NewEmptyLineRuleDefaultActive(engMessagesEmptyLine(), true)
	require.False(t, on.IsDefaultOff())
	require.True(t, on.IsOfficeDefaultOn())
}

func TestEmptyLineRule_IsSecondParagraphEndMark(t *testing.T) {
	// Ports EmptyLineRule.isSecondParagraphEndMark private helper (both modes).
	rule := NewEmptyLineRule(engMessagesEmptyLine())

	// Multi-line mode (default): need four line breaks / equivalent CRLF pairs.
	rule.SingleLineBreaksMarksPara = false
	require.True(t, rule.isSecondParagraphEndMark("Hello.\n\n\n\n"))
	require.True(t, rule.isSecondParagraphEndMark("Hello.\n\r\n\r\n\r\n\r"))
	require.True(t, rule.isSecondParagraphEndMark("Hello.\r\n\r\n\r\n\r\n"))
	require.False(t, rule.isSecondParagraphEndMark("Hello.\n\n"))
	require.False(t, rule.isSecondParagraphEndMark("Hello.\n"))
	require.False(t, rule.isSecondParagraphEndMark("Hello."))

	// Single-line-break marks paragraph: empty line is second break → \n\n / \n\r\n\r.
	rule.SingleLineBreaksMarksPara = true
	require.True(t, rule.isSecondParagraphEndMark("Hello.\n\n"))
	require.True(t, rule.isSecondParagraphEndMark("Hello.\n\r\n\r"))
	require.False(t, rule.isSecondParagraphEndMark("Hello.\n"))
	require.False(t, rule.isSecondParagraphEndMark("Hello."))
	// Four newlines still ends with \n\n → true in single-line mode.
	require.True(t, rule.isSecondParagraphEndMark("Hello.\n\n\n\n"))
}

func TestEmptyLineRule_MatchList_DemoMultiLine(t *testing.T) {
	// Demo + SimpleSentenceTokenizer (multi-line default): segment-simple.srx splits so no
	// sentence text ends with \n\n\n\n. Java Demo would likewise yield 0 EmptyLine matches.
	rule, lt := setUpEmptyLineRule()
	require.False(t, rule.SingleLineBreaksMarksPara)

	assertEmptyLineMatches(t, rule, lt, 0, "Hello world.\n\n\n\nNext para.")
	assertEmptyLineMatches(t, rule, lt, 0, "Hello world.\n\nNext para.")
	assertEmptyLineMatches(t, rule, lt, 0, "Hello world.\nNext.")
	assertEmptyLineMatches(t, rule, lt, 0, "A.\n\n\n\nB.\n\n\n\nC.")
	assertEmptyLineMatches(t, rule, lt, 0, "Hi.\n\n\n\nThere.")
	assertEmptyLineMatches(t, rule, lt, 0, "This is a test sentence.")
}

func TestEmptyLineRule_MatchList_ENMultiLine_JavaTwin(t *testing.T) {
	// Java en-US EmptyLineRule.match(lt.analyzeText(...)) outcomes (multi-line paragraph mode).
	// EN SRX keeps trailing newlines on the preceding sentence — required for second end mark.
	rule := NewEmptyLineRule(engMessagesEmptyLine())
	// English default: singleLineBreaksMarksPara=false (matches Java Language.createDefaultSentenceTokenizer).
	require.False(t, rule.SingleLineBreaksMarksPara)
	lt := languagetool.NewJLanguageTool("en")

	// No match: short paragraph breaks / single newline / trailing empty only.
	assertEmptyLineMatches(t, rule, lt, 0, "Hello world.\n\nNext para.")
	assertEmptyLineMatches(t, rule, lt, 0, "Hello world.\nNext.")
	assertEmptyLineMatches(t, rule, lt, 0, "Hello world.\n\n")
	assertEmptyLineMatches(t, rule, lt, 0, "Hello.\n\n\n\n") // last sentence only → loop size-2 empty
	assertEmptyLineMatches(t, rule, lt, 0, "This is a test sentence.")

	// Match: empty-line paragraph end present (four newlines between sentences).
	assertEmptyLineMatches(t, rule, lt, 1, "Hello world.\n\n\n\nNext para.")
	assertEmptyLineMatches(t, rule, lt, 1, "Hello world.\n\n\n\nNext para starts here.")
	assertEmptyLineMatches(t, rule, lt, 1, "Hi.\n\n\n\nThere.")
	assertEmptyLineMatches(t, rule, lt, 2, "A.\n\n\n\nB.\n\n\n\nC.")

	// Positions: last non-ws token of the paragraph-end sentence (the period), absolute.
	m1 := matchEmptyLine(rule, lt, "Hello world.\n\n\n\nNext para.")
	require.Len(t, m1, 1)
	require.Equal(t, 11, m1[0].GetFromPos())
	require.Equal(t, 12, m1[0].GetToPos())
	require.Equal(t, "EMPTY_LINE", rule.GetID())
	require.Equal(t, engMessagesEmptyLine()["empty_line_rule_msg"], m1[0].GetMessage())

	m2 := matchEmptyLine(rule, lt, "A.\n\n\n\nB.\n\n\n\nC.")
	require.Len(t, m2, 2)
	require.Equal(t, 1, m2[0].GetFromPos())
	require.Equal(t, 2, m2[0].GetToPos())
	require.Equal(t, 7, m2[1].GetFromPos())
	require.Equal(t, 8, m2[1].GetToPos())

	m3 := matchEmptyLine(rule, lt, "Hi.\n\n\n\nThere.")
	require.Len(t, m3, 1)
	require.Equal(t, 2, m3[0].GetFromPos())
	require.Equal(t, 3, m3[0].GetToPos())
}

func TestEmptyLineRule_MatchList_SingleLineMode_HelperDriven(t *testing.T) {
	// When SingleLineBreaksMarksPara=true, isSecondParagraphEndMark requires endsWith \n\n.
	// Full EN/Demo tokenize with single-line mode splits each \n, so real Analyze rarely keeps
	// \n\n on a sentence — Java en-US after setSingleLineBreaksMarksParagraph(true) also
	// yields 0 matches. Prove MatchList control flow via Analyze sentences that do end \n\n
	// under multi-line EN tokenize, with rule flag set to single-line (suffix \n\n of \n\n\n\n).
	rule := NewEmptyLineRule(engMessagesEmptyLine())
	lt := languagetool.NewJLanguageTool("en")
	// Multi-line analyze keeps "Hello world.\n\n\n\n" as one sentence.
	sents := lt.Analyze("Hello world.\n\n\n\nNext para.")
	require.GreaterOrEqual(t, len(sents), 2)
	require.True(t, strings.HasSuffix(sents[0].GetText(), "\n\n\n\n"))

	// Single-line rule mode: endsWith \n\n is true → match (same span as multi-line).
	rule.SingleLineBreaksMarksPara = true
	m := rule.MatchList(sents)
	require.Len(t, m, 1)
	require.Equal(t, 11, m[0].GetFromPos())
	require.Equal(t, 12, m[0].GetToPos())

	// Sentence ending with only one extra break is not a second paragraph end mark in SL mode
	// when text is just "\n" — covered by helper tests. Short multi-line \n\n alone:
	sents2 := lt.Analyze("Hello world.\n\nNext para.")
	// first sentence "Hello world.\n\n" ends with \n\n → SL second-mark true + paragraph end
	// → would match under SL flag (Java mixed-mode not typical, but control flow).
	rule.SingleLineBreaksMarksPara = true
	m2 := rule.MatchList(sents2)
	require.Len(t, m2, 1)
	require.Equal(t, 11, m2[0].GetFromPos())
	require.Equal(t, 12, m2[0].GetToPos())

	// Same text multi-line mode: endsWith \n\n is NOT second mark → 0.
	rule.SingleLineBreaksMarksPara = false
	require.Len(t, rule.MatchList(sents2), 0)
}

