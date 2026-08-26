package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port-style twin of core WordRepeatBeginningRule (no dedicated Java *Test in core).
//
// Java-visible outcomes via demoLT() (xx + SimpleSentenceTokenizer) and/or EN Analyze:
// rule.MatchList(lt.Analyze(text)) — not AnalyzePlain as sole proof.
//
// MessagesBundle_en keys used by the rule.

func wrbMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning":           "Successive sentences beginning with the same word",
		"desc_repetition_beginning_word":      "Three successive sentences begin with the same word.",
		"desc_repetition_beginning_adv":       "Two successive sentences begin with the same adverb.",
		"desc_repetition_beginning_thesaurus": "Consider rewording the sentence or use a thesaurus to find a synonym.",
		"category_repetitions":                "Repetitions",
	}
}

func wrbRule() *WordRepeatBeginningRule {
	return NewWordRepeatBeginningRule(wrbMessages())
}

// matchWRB ports rule.match(lt.analyzeText(text)).
func matchWRB(rule *WordRepeatBeginningRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func TestWordRepeatBeginningRule_Metadata(t *testing.T) {
	r := wrbRule()
	require.Equal(t, "WORD_REPEAT_BEGINNING_RULE", r.GetID())
	require.Equal(t, 2, r.MinToCheckParagraph())
	require.Equal(t, -1, r.EstimateContextForSureMatch())
	require.Equal(t, ITSStyle, r.GetLocQualityIssueType())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, "Successive sentences beginning with the same word", r.GetDescription())
	// Java isException list
	for _, ex := range []string{":", "–", "-", "✔️", "➡️", "—", "⭐️", "⚠️"} {
		require.True(t, r.isException(ex), ex)
	}
	require.False(t, r.isException("Also"))
	require.False(t, r.isAdverb(nil))
	require.False(t, r.isSentenceException(nil))
	require.Empty(t, r.getSuggestions(nil))
}

func TestWordRepeatBeginningRule_ThreeSuccessiveWords(t *testing.T) {
	// Java: three successive sentences begin with the same word → match on third.
	rule := wrbRule()
	lt := demoLT()

	// Two successive: no match (beforeLastToken not yet equal; isAdverb false).
	require.Equal(t, 0, len(matchWRB(rule, lt, "Also one here. Also two here.")))

	text := "Also one here. Also two here. Also three here."
	ms := matchWRB(rule, lt, text)
	require.Equal(t, 1, len(ms), "third Also should fire word-repetition")
	// Positions: sent0 correctedLen=15, sent1=15 → pos=30; "Also" at startPos 0, len 4
	require.Equal(t, 30, ms[0].GetFromPos())
	require.Equal(t, 34, ms[0].GetToPos())
	require.Equal(t, "Three successive sentences begin with the same word.", ms[0].GetShortMessage())
	require.Equal(t,
		"Three successive sentences begin with the same word. Consider rewording the sentence or use a thesaurus to find a synonym.",
		ms[0].GetMessage())
	require.Empty(t, ms[0].GetSuggestedReplacements())

	// EN Analyze path (tokenization only; core rule, not EnglishWordRepeatBeginningRule).
	enLT := languagetool.NewJLanguageTool("en")
	msEN := matchWRB(rule, enLT, text)
	require.Equal(t, 1, len(msEN))
	require.Equal(t, 30, msEN[0].GetFromPos())
	require.Equal(t, 34, msEN[0].GetToPos())
	require.Equal(t, "Three successive sentences begin with the same word.", msEN[0].GetShortMessage())
}

func TestWordRepeatBeginningRule_TwoSuccessiveAdverbs(t *testing.T) {
	// Java: two successive sentences begin with the same adverb → match on second.
	rule := wrbRule()
	rule.IsAdverbFn = func(token *languagetool.AnalyzedTokenReadings) bool {
		return token != nil && token.GetToken() == "Also"
	}
	lt := demoLT()

	text := "Also, I play football. Also, I play basketball."
	ms := matchWRB(rule, lt, text)
	require.Equal(t, 1, len(ms), "second Also adverb should match")
	// sent0 correctedLen=23; "Also" span 0..4 on second sentence → 23..27
	require.Equal(t, 23, ms[0].GetFromPos())
	require.Equal(t, 27, ms[0].GetToPos())
	require.Equal(t, "Two successive sentences begin with the same adverb.", ms[0].GetShortMessage())
	require.Equal(t,
		"Two successive sentences begin with the same adverb. Consider rewording the sentence or use a thesaurus to find a synonym.",
		ms[0].GetMessage())

	// Without IsAdverbFn: only two sentences → no word match (needs three).
	rule2 := wrbRule()
	require.Equal(t, 0, len(matchWRB(rule2, lt, text)))
}

func TestWordRepeatBeginningRule_Exceptions(t *testing.T) {
	// Java: isException on tokens[2] (":") avoids list-style "2007: …" matches.
	rule := wrbRule()
	lt := demoLT()
	require.Equal(t, 0, len(matchWRB(rule, lt,
		"2007: one thing. 2007: two things. 2007: three things.")))
	// Dash exception on second token after start word.
	require.Equal(t, 0, len(matchWRB(rule, lt,
		"A - note here. A - note again. A - note thrice.")))
}

func TestWordRepeatBeginningRule_PrevSentenceMustEndWithPunct(t *testing.T) {
	// Java: prevSentence.getText().trim().matches(".+[.?!]$") — table cells / bare lines skip.
	rule := wrbRule()
	lt := demoLT()
	// No sentence terminators → SimpleSentenceTokenizer yields one segment without .?!
	require.Equal(t, 0, len(matchWRB(rule, lt, "Also one\nAlso two\nAlso three")))

	// Two sentences only: word path needs three.
	require.Equal(t, 0, len(matchWRB(rule, lt, "Also one. Also two")))
}

func TestWordRepeatBeginningRule_SingleLetterNonLetterNotWord(t *testing.T) {
	// Java: token.length()==1 && !Character.isLetter(charAt(0)) → isWord false → no match.
	// Use GetAnalyzedSentence so each unit has tokens[1]=="!" with length>3 (not SRX-split "!" alone).
	rule := wrbRule()
	lt := demoLT()
	sents := []*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("! one two."),
		lt.GetAnalyzedSentence("! one two."),
		lt.GetAnalyzedSentence("! one two."),
	}
	require.Greater(t, len(sents[0].GetTokensWithoutWhitespace()), 3)
	require.Equal(t, "!", sents[0].GetTokensWithoutWhitespace()[1].GetToken())
	require.Equal(t, 1, javaStringLen("!"))
	require.Empty(t, rule.MatchList(sents))

	// Single-letter letter IS a word → third "A" matches.
	sentsA := []*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("A one two."),
		lt.GetAnalyzedSentence("A one two."),
		lt.GetAnalyzedSentence("A one two."),
	}
	ms := rule.MatchList(sentsA)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "Three successive sentences begin with the same word.", ms[0].GetShortMessage())
	// UTF-16 length of "A" is 1
	require.Equal(t, 1, ms[0].GetToPos()-ms[0].GetFromPos())
}

func TestWordRepeatBeginningRule_EmojiStartIsWord(t *testing.T) {
	// Java String.length()==2 for 😀 so isWord stays true (length!=1 gate).
	require.Equal(t, 2, javaStringLen("😀"))
	rule := wrbRule()
	lt := demoLT()
	sents := []*languagetool.AnalyzedSentence{
		lt.GetAnalyzedSentence("😀 one here."),
		lt.GetAnalyzedSentence("😀 two here."),
		lt.GetAnalyzedSentence("😀 three here."),
	}
	toks := sents[0].GetTokensWithoutWhitespace()
	if len(toks) > 3 && toks[1].GetToken() == "😀" {
		ms := rule.MatchList(sents)
		require.Equal(t, 1, len(ms), "third emoji-start sentence should match when isWord")
		require.Equal(t, 2, ms[0].GetToPos()-ms[0].GetFromPos())
	}
}

func TestWordRepeatBeginningRule_SentenceExceptionSkipsUpdates(t *testing.T) {
	// Java: isSentenceException → prevSentence=null; continue
	// continues WITHOUT updating lastToken / beforeLastToken / pos (bug-for-bug).
	// Token history is retained; only prevSentence is cleared.
	rule := wrbRule()
	rule.IsSentenceException = func(s *languagetool.AnalyzedSentence) bool {
		if s == nil {
			return false
		}
		txt := s.GetText()
		return len(txt) > 0 && txt[0] == 'X'
	}
	lt := demoLT()
	// Sequence:
	//   Also one, Also two (beforeLast=Also, last=Also), X exception (prev=null; tokens kept),
	//   Also three (prev==null → no match), Also four (word match), Also five (word match).
	text := "Also one here. Also two here. X exception. Also three here. Also four here. Also five here."
	ms := matchWRB(rule, lt, text)
	require.Equal(t, 2, len(ms), "word matches on 2nd and 3rd Also after exception (history tokens not cleared)")
	for _, m := range ms {
		require.Equal(t, "Three successive sentences begin with the same word.", m.GetShortMessage())
		require.Equal(t, 4, m.GetToPos()-m.GetFromPos()) // "Also"
	}

	// Exception sentence exists in Analyze output but its length is omitted from pos accumulation.
	sents := lt.Analyze(text)
	var foundX bool
	var xLen int
	for _, s := range sents {
		if len(s.GetText()) > 0 && s.GetText()[0] == 'X' {
			foundX = true
			xLen = s.GetCorrectedTextLength()
		}
	}
	require.True(t, foundX)
	require.Greater(t, xLen, 0)

	// Reconstruct expected fromPos for first post-exception word match ("Also four"):
	// pos sums only non-exception sentences before it (Also one + Also two + Also three).
	// sent texts from Analyze include trailing space on non-final segments.
	pos := 0
	nonExc := 0
	var expectedFirstFrom int
	for _, s := range sents {
		if len(s.GetText()) > 0 && s.GetText()[0] == 'X' {
			continue // Java continue: no pos +=
		}
		toks := s.GetTokensWithoutWhitespace()
		if nonExc >= 3 && len(toks) > 1 && toks[1].GetToken() == "Also" {
			// first match is the 4th non-exception Also sentence (index 3 among non-exc)
			// nonExc 0,1 pre-exception; 2 = Also three (no match); 3 = Also four (match)
			if nonExc == 3 {
				expectedFirstFrom = pos + toks[1].GetStartPos()
				break
			}
		}
		pos += s.GetCorrectedTextLength()
		nonExc++
	}
	require.Equal(t, expectedFirstFrom, ms[0].GetFromPos())
}

func TestWordRepeatBeginningRule_SuggestionsHook(t *testing.T) {
	rule := wrbRule()
	rule.GetSuggestionsFn = func(token *languagetool.AnalyzedTokenReadings) []string {
		return []string{"Furthermore"}
	}
	lt := demoLT()
	ms := matchWRB(rule, lt, "Also one here. Also two here. Also three here.")
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Furthermore"}, ms[0].GetSuggestedReplacements())
}
