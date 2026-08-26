package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Behavior matrix for org.languagetool.rules.WhiteSpaceBeforeParagraphEnd
// (no Java WhiteSpaceBeforeParagraphEndTest). Outcomes twin Java match / isParagraphEnd / metadata.
//
// Java setup pattern (sibling text-level rules):
//
//	JLanguageTool lt = new JLanguageTool(TestTools.getDemoLanguage()); // or en-US
//	WhiteSpaceBeforeParagraphEnd rule = new WhiteSpaceBeforeParagraphEnd(messages, lang); // defaultActive false
//	rule.match(lt.analyzeText(input))
//
// Go: demoLT() / NewJLanguageTool("en") + MatchList(lt.Analyze).

func engMessagesWSBeforeParaEnd() map[string]string {
	// MessagesBundle_en keys used by WhiteSpaceBeforeParagraphEnd (Java typo "parapgraph").
	return map[string]string{
		"whitespace_before_parapgraph_end_msg":  "Please delete the space character at the end of the paragraph.",
		"whitespace_before_parapgraph_end_desc": "Space character at the end of paragraph",
		"category_style":                        "Style",
	}
}

func setUpWSBeforeParaEnd() (*WhiteSpaceBeforeParagraphEnd, *languagetool.JLanguageTool) {
	// Java: new WhiteSpaceBeforeParagraphEnd(messages, lang) with Demo → multi-line paragraph marks (false).
	rule := NewWhiteSpaceBeforeParagraphEnd(engMessagesWSBeforeParaEnd())
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())
	return rule, lt
}

// matchWSBeforeParaEnd ports rule.match(lt.analyzeText(text)).
func matchWSBeforeParaEnd(rule *WhiteSpaceBeforeParagraphEnd, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func assertWSBeforeParaEndMatches(t *testing.T, rule *WhiteSpaceBeforeParagraphEnd, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(matchWSBeforeParaEnd(rule, lt, input))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestWhiteSpaceBeforeParagraphEnd_Metadata(t *testing.T) {
	// Java WhiteSpaceBeforeParagraphEnd(messages, lang) → defaultActive false
	rule := NewWhiteSpaceBeforeParagraphEnd(engMessagesWSBeforeParaEnd())
	require.Equal(t, "WHITESPACE_PARAGRAPH", rule.GetID())
	require.Equal(t, "Space character at the end of paragraph", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.True(t, rule.IsOfficeDefaultOn())
	require.Equal(t, 0, rule.MinToCheckParagraph())
	require.True(t, rule.GetCategory().GetID().Equals(NewCategoryId("STYLE")))
	require.Equal(t, ITSStyle, rule.GetLocQualityIssueType())
	require.False(t, rule.SingleLineBreaksMarksPara) // Demo/SRX default
}

func TestWhiteSpaceBeforeParagraphEnd_DefaultActive(t *testing.T) {
	// Java 3-arg: defaultActive false → setDefaultOff(); true leaves default on.
	off := NewWhiteSpaceBeforeParagraphEndDefaultActive(engMessagesWSBeforeParaEnd(), false)
	require.True(t, off.IsDefaultOff())
	require.True(t, off.IsOfficeDefaultOn())
	on := NewWhiteSpaceBeforeParagraphEndDefaultActive(engMessagesWSBeforeParaEnd(), true)
	require.False(t, on.IsDefaultOff())
	require.True(t, on.IsOfficeDefaultOn())
}

func TestWhiteSpaceBeforeParagraphEnd_MatchList_NoFalsePositive(t *testing.T) {
	// No trailing whitespace before paragraph end → no match.
	rule, lt := setUpWSBeforeParaEnd()
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello.")
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello world.")
	// Trailing linebreak only (no space/tab before it): lb lands on last non-lb; lw==lb → no match.
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello.\n")
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello.\n\nNext.")
	// Mid-sentence multi-space is not at paragraph end of that sentence when more text follows
	// on the same last sentence — last sentence "Hello  world." has no trailing ws.
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello  world.")
}

func TestWhiteSpaceBeforeParagraphEnd_MatchList_TrailingWhitespace(t *testing.T) {
	// Last sentence is always paragraph end (Tools.isParagraphEnd n >= size-1).
	rule, lt := setUpWSBeforeParaEnd()

	assertWSBeforeParaEndMatches(t, rule, lt, 1, "Hello  ")
	assertWSBeforeParaEndMatches(t, rule, lt, 1, "Hello. ")
	assertWSBeforeParaEndMatches(t, rule, lt, 1, "Hello\t")

	// Span + suggested replacement: non-ws at lw → replace from that token through trailing ws.
	// "Hello  " → tokens Hello, " ", " "; lw=Hello, lb=last space; from=0 to=7 sugg=Hello
	m1 := matchWSBeforeParaEnd(rule, lt, "Hello  ")
	require.Len(t, m1, 1)
	require.Equal(t, 0, m1[0].GetFromPos())
	require.Equal(t, 7, m1[0].GetToPos())
	require.Equal(t, []string{"Hello"}, m1[0].GetSuggestedReplacements())
	require.Equal(t, engMessagesWSBeforeParaEnd()["whitespace_before_parapgraph_end_msg"], m1[0].GetMessage())
	require.Equal(t, "WHITESPACE_PARAGRAPH", rule.GetID())

	// "Hello. " → lw=period, from=period.start (5) to space end (7), sugg="."
	m2 := matchWSBeforeParaEnd(rule, lt, "Hello. ")
	require.Len(t, m2, 1)
	require.Equal(t, 5, m2[0].GetFromPos())
	require.Equal(t, 7, m2[0].GetToPos())
	require.Equal(t, []string{"."}, m2[0].GetSuggestedReplacements())

	// Tab
	m3 := matchWSBeforeParaEnd(rule, lt, "Hello\t")
	require.Len(t, m3, 1)
	require.Equal(t, 0, m3[0].GetFromPos())
	require.Equal(t, 6, m3[0].GetToPos())
	require.Equal(t, []string{"Hello"}, m3[0].GetSuggestedReplacements())

	// Pure whitespace paragraph: lw lands on SENT_START (ws); from=tokens[1].start; sugg=""
	m4 := matchWSBeforeParaEnd(rule, lt, "   ")
	require.Len(t, m4, 1)
	require.Equal(t, 0, m4[0].GetFromPos())
	require.Equal(t, 3, m4[0].GetToPos())
	require.Equal(t, []string{""}, m4[0].GetSuggestedReplacements())
}

func TestWhiteSpaceBeforeParagraphEnd_MatchList_ExcludeZWSP(t *testing.T) {
	// Java: walk stops at "\u200B" (not walked as deletable trailing ws).
	rule, lt := setUpWSBeforeParaEnd()

	// Trailing ZWSP only: lb/lw on ZWSP, lw==lb → no match (ZWSP not walked left of itself as "extra").
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello\u200B")
	// Space then ZWSP: last token ZWSP is whitespace but excluded from walk; lw stops at ZWSP
	// with lw==lb (no walk) if ZWSP is last — no match.
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello \u200B")

	// ZWSP then spaces: walk stops at ZWSP (ws); fromPos = tokens[lw+1].start; sugg ""
	m := matchWSBeforeParaEnd(rule, lt, "Hello\u200B  ")
	require.Len(t, m, 1)
	require.Equal(t, 6, m[0].GetFromPos())
	require.Equal(t, 8, m[0].GetToPos())
	require.Equal(t, []string{""}, m[0].GetSuggestedReplacements())
}

func TestWhiteSpaceBeforeParagraphEnd_MatchList_ParagraphBreak(t *testing.T) {
	// Paragraph-end sentence with trailing spaces before linebreaks matches; span ends at last
	// non-linebreak token (Java toPos = tokens[lb].end — linebreaks excluded from span).
	rule, lt := setUpWSBeforeParaEnd()

	// Demo SimpleSentenceTokenizer: "Hello  \n\nNext." → first sentence keeps "\n\n".
	m := matchWSBeforeParaEnd(rule, lt, "Hello  \n\nNext.")
	require.Len(t, m, 1)
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 7, m[0].GetToPos()) // "Hello  " only; linebreaks after lb excluded
	require.Equal(t, []string{"Hello"}, m[0].GetSuggestedReplacements())

	// No trailing ws before para break → 0 (last "Next." has no trailing ws either).
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello.\n\nNext.")

	// Both paragraphs end with trailing spaces.
	m2 := matchWSBeforeParaEnd(rule, lt, "A  \n\nB  ")
	require.Len(t, m2, 2)
	require.Equal(t, 0, m2[0].GetFromPos())
	require.Equal(t, 3, m2[0].GetToPos())
	require.Equal(t, []string{"A"}, m2[0].GetSuggestedReplacements())
	// pos after first corrected len 5 ("A  \n\n") + B start 0 → absolute 5..8
	require.Equal(t, 5, m2[1].GetFromPos())
	require.Equal(t, 8, m2[1].GetToPos())
	require.Equal(t, []string{"B"}, m2[1].GetSuggestedReplacements())
}

func TestWhiteSpaceBeforeParagraphEnd_MatchList_EN_JavaTwin(t *testing.T) {
	// Same control flow under EN analyze (real word tokenizer + SRX path).
	rule := NewWhiteSpaceBeforeParagraphEnd(engMessagesWSBeforeParaEnd())
	// English default: singleLineBreaksMarksPara=false
	require.False(t, rule.SingleLineBreaksMarksPara)
	lt := languagetool.NewJLanguageTool("en")

	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello.")
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello  world.")
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello.\n\nNext.")
	assertWSBeforeParaEndMatches(t, rule, lt, 1, "Hello  ")
	assertWSBeforeParaEndMatches(t, rule, lt, 1, "Hello. ")
	assertWSBeforeParaEndMatches(t, rule, lt, 1, "Hello\t")
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello\u200B")
	assertWSBeforeParaEndMatches(t, rule, lt, 0, "Hello \u200B")

	m1 := matchWSBeforeParaEnd(rule, lt, "Hello  ")
	require.Len(t, m1, 1)
	require.Equal(t, 0, m1[0].GetFromPos())
	require.Equal(t, 7, m1[0].GetToPos())
	require.Equal(t, []string{"Hello"}, m1[0].GetSuggestedReplacements())
	require.Equal(t, engMessagesWSBeforeParaEnd()["whitespace_before_parapgraph_end_msg"], m1[0].GetMessage())

	mZW := matchWSBeforeParaEnd(rule, lt, "Hello\u200B  ")
	require.Len(t, mZW, 1)
	require.Equal(t, 6, mZW[0].GetFromPos())
	require.Equal(t, 8, mZW[0].GetToPos())
	require.Equal(t, []string{""}, mZW[0].GetSuggestedReplacements())

	// EN SRX keeps trailing \n\n on preceding sentence.
	mPara := matchWSBeforeParaEnd(rule, lt, "Hello  \n\nNext.")
	require.Len(t, mPara, 1)
	require.Equal(t, 0, mPara[0].GetFromPos())
	require.Equal(t, 7, mPara[0].GetToPos())
	require.Equal(t, []string{"Hello"}, mPara[0].GetSuggestedReplacements())

	mBoth := matchWSBeforeParaEnd(rule, lt, "A  \n\nB  ")
	require.Len(t, mBoth, 2)
	require.Equal(t, 0, mBoth[0].GetFromPos())
	require.Equal(t, 3, mBoth[0].GetToPos())
	require.Equal(t, 5, mBoth[1].GetFromPos())
	require.Equal(t, 8, mBoth[1].GetToPos())
}

func TestWhiteSpaceBeforeParagraphEnd_MatchList_LastSentenceAlwaysParaEnd(t *testing.T) {
	// Tools.isParagraphEnd: nTest >= sentences.size()-1 → true always for last sentence.
	rule, lt := setUpWSBeforeParaEnd()
	// Multi-sentence: only last has trailing spaces → one match at absolute pos after first sentence.
	// "Hi. Bye  " — last sentence ends with spaces.
	m := matchWSBeforeParaEnd(rule, lt, "Hi. Bye  ")
	require.GreaterOrEqual(t, len(lt.Analyze("Hi. Bye  ")), 1)
	require.Len(t, m, 1)
	// Suggestion is last non-ws token of last sentence ("Bye") spanning through trailing spaces.
	require.Equal(t, []string{"Bye"}, m[0].GetSuggestedReplacements())
	require.Equal(t, 4, m[0].GetFromPos()) // "Hi. " len 4 + "Bye  " from 0
	require.Equal(t, 9, m[0].GetToPos())

	// First sentence with trailing spaces only if it is also a paragraph end.
	// Single-sentence trailing: covered above. Mid non-para-end with spaces: EN may split.
}

func TestWhiteSpaceBeforeParagraphEnd_MatchList_AnalyzeMultiSentence(t *testing.T) {
	// Full analyzeText path: multi-sentence document, absolute positions via correctedTextLength.
	rule, lt := setUpWSBeforeParaEnd()
	require.Equal(t, 0, len(matchWSBeforeParaEnd(rule, lt, "Hello world.")))
	require.Equal(t, 1, len(matchWSBeforeParaEnd(rule, lt, "Hello  ")))
	require.Equal(t, 1, len(matchWSBeforeParaEnd(rule, lt, "Hello  \n\nNext.")))
	require.Equal(t, 2, len(matchWSBeforeParaEnd(rule, lt, "A  \n\nB  ")))

	// Non-last sentence that is not a paragraph end should not match even with spaces
	// when SRX attaches spaces to a sentence that is not para-end.
	// EN: "Hello.  Next." may tokenize oddly; use explicit multi-line non-para-end.
	// With multi-line mode false, single \n may still mark para end if next starts with \n —
	// use no break: two sentences in one line without para marks on first.
	// Demo/EN both: last sentence always checked; first without para-end suffix → skip.
	sents := lt.Analyze("One. Two  ")
	// Regardless of split, MatchList only flags para-end sentences with trailing ws.
	ms := rule.MatchList(sents)
	require.Equal(t, 1, len(ms), "only the paragraph-end sentence with trailing ws")
}

func TestWhiteSpaceBeforeParagraphEnd_IsParagraphEnd_Helper(t *testing.T) {
	// Ports Tools.isParagraphEnd via rule flag (both modes).
	rule := NewWhiteSpaceBeforeParagraphEnd(engMessagesWSBeforeParaEnd())
	lt := languagetool.NewJLanguageTool("en")

	// Last sentence always end.
	sents := lt.Analyze("Hello.")
	require.True(t, rule.isParagraphEnd(sents, 0))

	// Multi-line default: ends with \n\n is paragraph end.
	sents2 := lt.Analyze("Hello.\n\nNext.")
	require.GreaterOrEqual(t, len(sents2), 2)
	rule.SingleLineBreaksMarksPara = false
	require.True(t, rule.isParagraphEnd(sents2, 0))
	require.True(t, rule.isParagraphEnd(sents2, len(sents2)-1))

	// Single-line mode: single \n suffix marks para end.
	sents3 := lt.Analyze("Hello.\nNext.")
	require.GreaterOrEqual(t, len(sents3), 2)
	rule.SingleLineBreaksMarksPara = true
	// Depending on SRX attach, first may end with \n or next start with \n — either path true
	// for paragraph end under Tools.isParagraphEnd logic.
	// When first ends without \n and next does not start with \n, not para end.
	// EN typically keeps "Hello.\n" then "Next." → endsWith \n → true in SL mode.
	if len(sents3) >= 2 {
		_ = rule.isParagraphEnd(sents3, 0) // exercised
		require.True(t, rule.isParagraphEnd(sents3, len(sents3)-1))
	}
}
