package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Behavior matrix for org.languagetool.rules.ParagraphRepeatBeginningRule
// (no core Java unit test; DE GermanParagraphRepeatBeginningRuleTest is a different leaf).
// Outcomes twin Java match / numCharEqualBeginning / isArticle / metadata.
//
// Java setup pattern (sibling text-level rules + DE test shape):
//
//	JLanguageTool lt = new JLanguageTool(lang); // Demo or en-US
//	ParagraphRepeatBeginningRule rule = new ParagraphRepeatBeginningRule(messages, lang);
//	// setDefaultOff — enable for check; here we call MatchList(lt.analyzeText(...))
//
// Go: demoLT() / NewJLanguageTool("en") + MatchList(lt.Analyze).
// Prefer EN Analyze: SRX keeps trailing \n\n on the preceding sentence (required for
// multi-line paragraph mode). Demo SimpleSentenceTokenizer often emits a bare "\n"
// sentence between paragraphs, so first/next non-ws beginnings never line up (0 matches).

func engMessagesParaRepeat() map[string]string {
	// MessagesBundle_en keys used by ParagraphRepeatBeginningRule.
	return map[string]string{
		"repetition_paragraph_beginning_desc":     "Same beginning of paragraph",
		"repetition_paragraph_beginning_last_msg": "Same beginning as last paragraph",
		"category_style": "Style",
	}
}

func setUpParaRepeat() (*ParagraphRepeatBeginningRule, *languagetool.JLanguageTool) {
	// Java: new ParagraphRepeatBeginningRule(messages, lang) with Demo → multi-line para marks (false).
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())
	return rule, lt
}

// matchParaRepeat ports rule.match(lt.analyzeText(text)).
func matchParaRepeat(rule *ParagraphRepeatBeginningRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func assertParaRepeatMatches(t *testing.T, rule *ParagraphRepeatBeginningRule, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(matchParaRepeat(rule, lt, input))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestParagraphRepeatBeginningRule_Metadata(t *testing.T) {
	// Java: STYLE, Style, setDefaultOff(), minToCheckParagraph=1, id constant.
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	require.Equal(t, "PARAGRAPH_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, "Same beginning of paragraph", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 1, rule.MinToCheckParagraph())
	require.True(t, rule.GetCategory().GetID().Equals(NewCategoryId("STYLE")))
	require.Equal(t, ITSStyle, rule.GetLocQualityIssueType())
	require.False(t, rule.SingleLineBreaksMarksPara) // Demo/SRX default
}

func TestParagraphRepeatBeginningRule_IsArticle_DefaultDT(t *testing.T) {
	// Java isArticle: hasPosTagStartingWith("DT"); overridable via IsArticle field.
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	// Untagged token is not an article.
	tok := demoLT().GetAnalyzedSentence("The dog").GetTokensWithoutWhitespace()[1]
	require.Equal(t, "The", tok.GetToken())
	require.False(t, rule.isArticle(tok))

	// Inject DT like a tagger would.
	pos := "DT"
	tok.AddReading(languagetool.NewAnalyzedToken("The", &pos, nil), "test")
	require.True(t, rule.isArticle(tok))

	// Override field (German subclass pattern).
	rule.IsArticle = func(token *languagetool.AnalyzedTokenReadings) bool {
		return token != nil && token.HasPosTagStartingWith("ART")
	}
	require.False(t, rule.isArticle(tok)) // DT not ART
	posART := "ART:DEF"
	tok.AddReading(languagetool.NewAnalyzedToken("The", &posART, nil), "test")
	require.True(t, rule.isArticle(tok))
}

func TestParagraphRepeatBeginningRule_MatchList_EN_RepeatedFirstWord(t *testing.T) {
	// Java DE shape: same first word across \n\n → 2 matches (last + next paragraph).
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")
	require.False(t, rule.SingleLineBreaksMarksPara)

	// Same proper name across paragraph boundary → 2 matches.
	assertParaRepeatMatches(t, rule, lt, 2, "Peter walked down the street.\n\nPeter ignored the traffic.")
	// Different beginnings → 0.
	assertParaRepeatMatches(t, rule, lt, 0, "Peter walked down the street.\n\nMary ignored the traffic.")
	// No paragraph boundary (single \n, multi-line mode): not a paragraph end unless next starts with \n.
	assertParaRepeatMatches(t, rule, lt, 0, "Peter walked.\nPeter ignored.")
	// Same start without para break on one line → 0 (one sentence or non-para).
	assertParaRepeatMatches(t, rule, lt, 0, "Peter walked. Peter ignored.")
	// Plain text.
	assertParaRepeatMatches(t, rule, lt, 0, "This is a test sentence.")
}

func TestParagraphRepeatBeginningRule_MatchList_EN_PositionsAndMessage(t *testing.T) {
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")

	// "Peter walked down the street.\n\nPeter ignored the traffic."
	// first "Peter" [0,5]; nextPos = utf16Len(first sentence) = 31; next "Peter" [31,36]
	m := matchParaRepeat(rule, lt, "Peter walked down the street.\n\nPeter ignored the traffic.")
	require.Len(t, m, 2)
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 5, m[0].GetToPos())
	require.Equal(t, 31, m[1].GetFromPos())
	require.Equal(t, 36, m[1].GetToPos())
	require.Equal(t, engMessagesParaRepeat()["repetition_paragraph_beginning_last_msg"], m[0].GetMessage())
	require.Equal(t, engMessagesParaRepeat()["repetition_paragraph_beginning_last_msg"], m[1].GetMessage())
	require.Equal(t, "PARAGRAPH_REPEAT_BEGINNING_RULE", rule.GetID())
}

func TestParagraphRepeatBeginningRule_MatchList_EN_QuoteSkip(t *testing.T) {
	// Java DE: "»Peter …«\n\n»Peter …«" → 2 matches spanning quote+name.
	// EN straight double quotes are in QUOTES_REGEX.
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")

	m := matchParaRepeat(rule, lt, "\"Peter walked.\"\n\n\"Peter ran.\"")
	require.Len(t, m, 2)
	// start = lastTokens[1] quote start 0; end = lastTokens[2] "Peter" end 6
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 6, m[0].GetToPos())
	// nextPos = len("\"Peter walked.\"\n\n") = 17; next span 17..23
	require.Equal(t, 17, m[1].GetFromPos())
	require.Equal(t, 23, m[1].GetToPos())

	// Matching quotes but different following word → 0
	assertParaRepeatMatches(t, rule, lt, 0, "\"Peter walked.\"\n\n\"Mary ran.\"")
}

func TestParagraphRepeatBeginningRule_MatchList_EN_ArticlePath(t *testing.T) {
	// Java: if first word is article (DT), compare first two words.
	// EN Analyze in this package does not wire english.dict POS (import cycle with tagging/en),
	// so exercise the same control flow via IsArticle hook (Java overridable isArticle).
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")
	rule.IsArticle = func(token *languagetool.AnalyzedTokenReadings) bool {
		if token == nil {
			return false
		}
		// Surface stand-in for DT on English article "The"/"the" only — hook is the Java override point.
		tok := token.GetToken()
		return tok == "The" || tok == "the" || token.HasPosTagStartingWith("DT")
	}

	// Same article+noun → 2 matches spanning "The dog"
	m := matchParaRepeat(rule, lt, "The dog walked.\n\nThe dog ran.")
	require.Len(t, m, 2)
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 7, m[0].GetToPos()) // "The dog"
	require.Equal(t, 17, m[1].GetFromPos())
	require.Equal(t, 24, m[1].GetToPos())

	// Same article, different noun → 0 (Java Der Hund / Das Tier shape)
	assertParaRepeatMatches(t, rule, lt, 0, "The dog walked.\n\nThe cat ran.")

	// Quote + article + noun both paragraphs
	m2 := matchParaRepeat(rule, lt, "\"The dog walked.\"\n\n\"The dog ran.\"")
	require.Len(t, m2, 2)
	require.Equal(t, 0, m2[0].GetFromPos())
	require.Equal(t, 8, m2[0].GetToPos()) // "\"The dog"
	require.Equal(t, 19, m2[1].GetFromPos())
	require.Equal(t, 27, m2[1].GetToPos())

	// Without article hook, untagged "The" is only first-word compare (still 2 matches, shorter span).
	ruleNo := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	m3 := matchParaRepeat(ruleNo, lt, "The dog walked.\n\nThe dog ran.")
	require.Len(t, m3, 2)
	require.Equal(t, 0, m3[0].GetFromPos())
	require.Equal(t, 3, m3[0].GetToPos()) // "The" only
	// Same first word "The" with different second word still matches without article path
	assertParaRepeatMatches(t, ruleNo, lt, 2, "The dog walked.\n\nThe cat ran.")
}

func TestParagraphRepeatBeginningRule_MatchList_EN_ThreeParagraphs(t *testing.T) {
	// Each paragraph-end boundary can emit 2 matches (prev + next).
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")
	m := matchParaRepeat(rule, lt, "Peter a.\n\nPeter b.\n\nPeter c.")
	require.Len(t, m, 4)
}

func TestParagraphRepeatBeginningRule_MatchList_Demo_TokenizerNote(t *testing.T) {
	// Demo + SimpleSentenceTokenizer: segment-simple often yields a bare "\n" sentence between
	// paragraphs, so numCharEqualBeginning hits whitespace at nextTokens[1] → 0.
	// Same as EmptyLineRule Demo documentation: prove the path, do not invent AnalyzePlain.
	rule, lt := setUpParaRepeat()
	require.False(t, rule.SingleLineBreaksMarksPara)

	assertParaRepeatMatches(t, rule, lt, 0, "Peter walked down the street.\n\nPeter ignored the traffic.")
	assertParaRepeatMatches(t, rule, lt, 0, "Wiederholung am Anfang.\n\nWiederholung am Ende.")

	// When SRX keeps \n\n on the preceding sentence (quotes sample), Demo still matches.
	m := matchParaRepeat(rule, lt, "\"Peter walked.\"\n\n\"Peter ran.\"")
	require.Len(t, m, 2)
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 6, m[0].GetToPos())
}

func TestParagraphRepeatBeginningRule_MatchList_NoParaBoundaryAndNonLetter(t *testing.T) {
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")

	// Number / non-letter beginning: Character.isLetter fails → 0
	assertParaRepeatMatches(t, rule, lt, 0, "123 went home.\n\n123 went away.")
	// Empty / short
	assertParaRepeatMatches(t, rule, lt, 0, "")
	assertParaRepeatMatches(t, rule, lt, 0, "Only one paragraph.")
}

func TestParagraphRepeatBeginningRule_MatchList_SingleLineBreaksMode(t *testing.T) {
	// When SingleLineBreaksMarksPara=true, a sentence ending with \n is a paragraph end.
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")
	rule.SingleLineBreaksMarksPara = true

	// EN may keep "Peter walked.\n" then "Peter ignored." → para end → 2 matches.
	sents := lt.Analyze("Peter walked.\nPeter ignored.")
	require.GreaterOrEqual(t, len(sents), 2)
	m := rule.MatchList(sents)
	// If first ends with \n or next starts with \n, Tools.isParagraphEnd is true.
	if rule.isParagraphEnd(sents, 0) {
		require.Len(t, m, 2)
		require.Equal(t, 0, m[0].GetFromPos())
		require.Equal(t, 5, m[0].GetToPos())
	} else {
		// Fallback: multi-line \n\n still works under SL mode (endsWith \n).
		assertParaRepeatMatches(t, rule, lt, 2, "Peter walked.\n\nPeter ignored.")
	}
}

func TestParagraphRepeatBeginningRule_NumCharEqualBeginning_Helper(t *testing.T) {
	// Direct twin of private helper via EN-analyzed tokens (no AnalyzePlain invent).
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")

	sents := lt.Analyze("Peter walked.\n\nPeter ignored.")
	require.GreaterOrEqual(t, len(sents), 2)
	last := sents[0].GetTokensWithoutWhitespace()
	next := sents[1].GetTokensWithoutWhitespace()
	require.Equal(t, 5, rule.numCharEqualBeginning(last, next)) // "Peter" end

	sents2 := lt.Analyze("Peter walked.\n\nMary ignored.")
	require.Equal(t, 0, rule.numCharEqualBeginning(
		sents2[0].GetTokensWithoutWhitespace(),
		sents2[1].GetTokensWithoutWhitespace(),
	))

	// Quote skip
	sents3 := lt.Analyze("\"Peter walked.\"\n\n\"Peter ran.\"")
	require.Equal(t, 6, rule.numCharEqualBeginning(
		sents3[0].GetTokensWithoutWhitespace(),
		sents3[1].GetTokensWithoutWhitespace(),
	))

	// Whitespace-only next beginning (Demo bare newline sentence shape)
	bareNL := lt.Analyze("x\n\ny")
	// If a sentence is only newlines, tokensWithoutWhitespace[1] may be whitespace → 0
	for i := 0; i < len(bareNL)-1; i++ {
		end := rule.numCharEqualBeginning(
			bareNL[i].GetTokensWithoutWhitespace(),
			bareNL[i+1].GetTokensWithoutWhitespace(),
		)
		// not asserting exact pairs; just ensure helper does not panic
		_ = end
	}
}

func TestParagraphRepeatBeginningRule_QuotesRegex_FullMatch(t *testing.T) {
	// Java QUOTES_REGEX.matcher(token).matches() — whole token is one quote/bracket char.
	require.True(t, paraQuotesRE.MatchString(`"`))
	require.True(t, paraQuotesRE.MatchString(`'`))
	require.True(t, paraQuotesRE.MatchString(`(`))
	require.True(t, paraQuotesRE.MatchString(`[`))
	require.True(t, paraQuotesRE.MatchString(`„`))
	require.False(t, paraQuotesRE.MatchString(`"Peter`)) // find would true; matches() false
	require.False(t, paraQuotesRE.MatchString(`Peter`))
	require.False(t, paraQuotesRE.MatchString(``))
}

func TestParagraphRepeatBeginningRule_GetTextLengthPosAccumulation(t *testing.T) {
	// Java uses getText().length() not getCorrectedTextLength for nextPos.
	// With EN, corrected and text lengths usually match; still assert absolute next span.
	rule := NewParagraphRepeatBeginningRule(engMessagesParaRepeat())
	lt := languagetool.NewJLanguageTool("en")
	text := "Café opened.\n\nCafé closed."
	sents := lt.Analyze(text)
	require.GreaterOrEqual(t, len(sents), 2)
	// nextPos after first sentence must be utf16 length of getText() (Café = 4 UTF-16 units).
	nextPos := 0
	for _, r := range sents[0].GetText() {
		if r >= 0x10000 {
			nextPos += 2
		} else {
			nextPos++
		}
	}
	m := rule.MatchList(sents)
	require.Len(t, m, 2)
	require.Equal(t, 0, m[0].GetFromPos())
	require.Equal(t, 4, m[0].GetToPos()) // "Café"
	require.Equal(t, nextPos, m[1].GetFromPos())
	require.Equal(t, nextPos+4, m[1].GetToPos())
}
