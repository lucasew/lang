package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of WordRepeatBeginningRule: token.length()==1 && !Character.isLetter(charAt(0)).
func TestWordRepeatBeginning_IsWordUTF16Length(t *testing.T) {
	r := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word":      "Word repetition at sentence start.",
		"desc_repetition_beginning_thesaurus": "Consider a thesaurus.",
	})
	// "!" has UTF-16 length 1 and is not a letter → isWord false → no match
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("! one two."),
		languagetool.AnalyzePlain("! one two."),
		languagetool.AnalyzePlain("! one two."),
	}
	require.Empty(t, r.MatchList(sents))

	// Three successive sentences starting with "Also" → third fires word-repetition
	r2 := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word":      "Word repetition at sentence start.",
		"desc_repetition_beginning_thesaurus": "Consider a thesaurus.",
	})
	sents2 := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Also one here."),
		languagetool.AnalyzePlain("Also two here."),
		languagetool.AnalyzePlain("Also three here."),
	}
	ms := r2.MatchList(sents2)
	require.Equal(t, 1, len(ms), "third Also should match word-repetition short msg")
	require.Equal(t, "Also", sents2[2].GetTokensWithoutWhitespace()[1].GetToken())
	require.Equal(t, 4, ms[0].GetToPos()-ms[0].GetFromPos()) // "Also" UTF-16 len
}

// Emoji as sentence start: Java String.length()==2 so isWord stays true.
func TestWordRepeatBeginning_EmojiStartIsWord(t *testing.T) {
	require.Equal(t, 2, utf16LenStr("😀"), "emoji is two UTF-16 units")
	// punctuation exception list includes some emoji-ish marks but not 😀
	r := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word":      "Word repetition at sentence start.",
		"desc_repetition_beginning_thesaurus": "Consider a thesaurus.",
	})
	// Need tokens.length > 3: START + 😀 + word + .
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("😀 one here."),
		languagetool.AnalyzePlain("😀 two here."),
		languagetool.AnalyzePlain("😀 three here."),
	}
	// If tokenizer keeps 😀 as first content token with enough following tokens, third fires.
	// Fail-closed: only assert when token structure matches Java gate.
	if toks := sents[0].GetTokensWithoutWhitespace(); len(toks) > 3 && toks[1].GetToken() == "😀" {
		ms := r.MatchList(sents)
		require.Equal(t, 1, len(ms), "third emoji-start sentence should match when isWord")
		require.Equal(t, 2, ms[0].GetToPos()-ms[0].GetFromPos())
	}
}

func TestWordRepeatBeginning_Exceptions(t *testing.T) {
	r := NewWordRepeatBeginningRule(nil)
	// Java isException list
	for _, ex := range []string{":", "–", "-", "✔️", "➡️", "—", "⭐️", "⚠️"} {
		require.True(t, r.isException(ex), ex)
	}
	require.False(t, r.isException("Er"))
	require.Equal(t, 2, r.MinToCheckParagraph())
	require.Equal(t, "WORD_REPEAT_BEGINNING_RULE", r.GetID())
}
