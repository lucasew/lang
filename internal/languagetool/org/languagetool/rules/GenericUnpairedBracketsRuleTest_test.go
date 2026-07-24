package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.GenericUnpairedBracketsRuleTest (guillemets » «).

func bracketsRule() *GenericUnpairedBracketsRule {
	return NewGenericUnpairedBracketsRule(map[string]string{
		"unpaired_brackets":      "Unpaired bracket, expected %s",
		"desc_unpaired_brackets": "Unpaired brackets",
	}, []string{"\u00bb"}, []string{"\u00ab"})
}

func analyzeForBrackets(input string) []*languagetool.AnalyzedSentence {
	if strings.Contains(input, "\n\n") {
		paras := strings.Split(input, "\n\n")
		var out []*languagetool.AnalyzedSentence
		off := 0
		for pi, para := range paras {
			var sents []*languagetool.AnalyzedSentence
			if strings.Contains(para, ". ") || strings.Contains(para, ".\n") {
				sents = languagetool.SplitAndAnalyze(para)
			} else {
				sents = []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(para)}
			}
			for _, s := range sents {
				if off > 0 {
					for _, tok := range s.GetTokens() {
						tok.SetStartPos(tok.GetStartPos() + off)
					}
				}
				out = append(out, s)
			}
			for _, r := range para {
				if r >= 0x10000 {
					off += 2
				} else {
					off++
				}
			}
			if pi < len(paras)-1 {
				off += 2 // \n\n
			}
		}
		return out
	}
	if strings.Contains(input, ". ") || strings.Contains(input, ".\n") {
		return languagetool.SplitAndAnalyze(input)
	}
	return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)}
}

func assertBracketMatches(t *testing.T, expected int, input string) {
	t.Helper()
	rule := bracketsRule()
	got := len(rule.MatchList(analyzeForBrackets(input)))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestGenericUnpairedBracketsRule_Rule(t *testing.T) {
	assertBracketMatches(t, 0, "This is \u00bbcorrect\u00ab.")
	assertBracketMatches(t, 0, "\u00bbCorrect\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends.\u00ab")
	assertBracketMatches(t, 0, "\u00bbCorrect. This is more than one sentence.\u00ab")
	assertBracketMatches(t, 0, "\u00bbCorrect. This is more than one sentence.\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends.\u00ab")
	assertBracketMatches(t, 0, "\u00bbCorrect\u00ab\n\n\u00bbAnd here it ends.\u00ab\n\nMore text.")
	assertBracketMatches(t, 0, "\u00bbCorrect, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertBracketMatches(t, 0, "\u00bbCorrect, he said.\n\nThis is the next sentence.\u00ab Here's another sentence.")
	assertBracketMatches(t, 0, "\u00bbCorrect, he said.\n\n\n\nThis is the next sentence.\u00ab Here's another sentence.")
	assertBracketMatches(t, 0, "This \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertBracketMatches(t, 0, "Good.\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertBracketMatches(t, 0, "Good.\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertBracketMatches(t, 0, "Good.\n\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertBracketMatches(t, 0, "This is funny :-)")
	assertBracketMatches(t, 0, "This is sad :-( isn't it")
	assertBracketMatches(t, 0, "This is funny :)")
	assertBracketMatches(t, 0, "This is sad :( isn't it")
	assertBracketMatches(t, 0, "a) item one\nb) item two")
	assertBracketMatches(t, 0, "a) item one\nb) item two\nc) item three")
	assertBracketMatches(t, 0, "\na) item one\nb) item two\nc) item three")
	assertBracketMatches(t, 0, "\n\na) item one\nb) item two\nc) item three")
	assertBracketMatches(t, 0, "This is a), not b)")
	assertBracketMatches(t, 0, "This is it (a, not b) some more test")
	assertBracketMatches(t, 0, "This is \u00bbnot an error yet")
	assertBracketMatches(t, 0, "See https://de.wikipedia.org/wiki/Schlammersdorf_(Adelsgeschlecht)")

	assertBracketMatches(t, 1, "This is not correct\u00ab")
	assertBracketMatches(t, 1, "This is \u00bbnot correct.")
	assertBracketMatches(t, 1, "This is \u00bbnot an error yet\n\nBut now it has become one")
	assertBracketMatches(t, 1, "This is correct.\n\n\u00bbBut this is not.")
	assertBracketMatches(t, 1, "This is correct.\n\nBut this is not\u00ab")
	assertBracketMatches(t, 1, "\u00bbThis is correct\u00ab\n\nBut this is not\u00ab")
	assertBracketMatches(t, 1, "\u00bbThis is correct\u00ab\n\nBut this \u00bbis\u00ab not\u00ab")
	assertBracketMatches(t, 1, "This is not correct. No matter if it's more than one sentence\u00ab")
	assertBracketMatches(t, 1, "\u00bbThis is not correct. No matter if it's more than one sentence")
	assertBracketMatches(t, 1, "Correct, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertBracketMatches(t, 1, "\u00bbCorrect, he said. This is the next sentence. Here's another sentence.")
	assertBracketMatches(t, 1, "\u00bbCorrect, he said. This is the next sentence.\n\nHere's another sentence.")
	assertBracketMatches(t, 1, "\u00bbCorrect, he said. This is the next sentence.\n\n\n\nHere's another sentence.")
}

func TestGenericUnpairedBracketsRule_RuleMatchPositions(t *testing.T) {
	rule := bracketsRule()
	match1 := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("This \u00bbis a test.")})
	require.NotEmpty(t, match1)
	require.Equal(t, 5, match1[0].GetFromPos())
	require.Equal(t, 6, match1[0].GetToPos())

	text2 := "This.\nSome stuff.\nIt \u00bbis a test."
	match2 := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(text2)})
	require.NotEmpty(t, match2)
	require.Equal(t, 21, match2[0].GetFromPos())
	require.Equal(t, 22, match2[0].GetToPos())

	// NBSP counts as one UTF-16 unit: Th + nbsp + is + space + »
	match3 := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Th\u00ADis \u00bbis a test.")})
	require.NotEmpty(t, match3)
	require.Equal(t, 6, match3[0].GetFromPos())
	require.Equal(t, 7, match3[0].GetToPos())
}
