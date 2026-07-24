package disambiguation

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func sentenceStart() *languagetool.AnalyzedTokenReadings {
	// SENT_START pos like Java JLanguageTool.SENTENCE_START_TAGNAME
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil))
}

func TestMultiWordChunkerSpace(t *testing.T) {
	c := NewMultiWordChunker([]string{"New York\tB-NP"}, MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
	})
	// Include sentence-start so whitespace branch (j > 1) matches Java indexing.
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
	found := false
	for _, tok := range out.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() != nil {
				p := *r.GetPOSTag()
				if p == "B-NP" || p == "<B-NP>" || p == "</B-NP>" {
					found = true
				}
			}
		}
	}
	require.True(t, found, "expected multiword POS tag on readings")
}

func TestMultiWordChunkerNoSpace(t *testing.T) {
	c := NewMultiWordChunker([]string{"...\tELLIPSIS"}, MultiWordChunkerSettings{})
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	found := false
	for _, tok := range out.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() != nil {
				p := *r.GetPOSTag()
				if p == "ELLIPSIS" || p == "<ELLIPSIS>" || p == "</ELLIPSIS>" {
					found = true
				}
			}
		}
	}
	require.True(t, found)
}

func TestMultiWordChunker_SeparatorRegExpSemicolon(t *testing.T) {
	// Java #separatorRegExp=[\t;] uses String.split(regex).
	c := NewMultiWordChunker([]string{
		`#separatorRegExp=[\t;]`,
		`home page;N f s`,
		"status quo\tNN",
	}, MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
	})
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("home", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("page", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
	found := false
	for _, tok := range out.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() == nil {
				continue
			}
			p := *r.GetPOSTag()
			if strings.Contains(p, "N f s") {
				found = true
			}
		}
	}
	require.True(t, found, "expected French-style semicolon multiword tag")
}

func TestMultiWordChunker_GermanLineExpander(t *testing.T) {
	// Expand via loadMultiWordLines (reader path).
	r := strings.NewReader("Aston Martin/S\n")
	c, err := NewMultiWordChunkerFromReader(r, MultiWordChunkerSettings{
		DefaultTag:            TagForNotAddingTags,
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
	})
	require.NoError(t, err)
	c.SetIgnoreSpelling(true)
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Aston", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Martins", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
	// tagForNotAddingTags does not add POS; ignore-spelling should fire on the span
	// (requires IsIgnoredBySpeller or similar — check via reading presence of ignore flag).
	// MultiWordChunker with TagForNotAddingTags still sets IgnoreSpelling when configured.
	// Verify chunker loaded expanded forms.
	require.Contains(t, c.Lines, "Aston Martin")
	require.Contains(t, c.Lines, "Aston Martins")
	_ = out
}

func TestRemovePreviousTags_MultiTokenSpan(t *testing.T) {
	// After MultiWordChunker: New has <NNP>, York has </NNP> → removePreviousTags → NNP NNP
	nnpOpen := "<NNP>"
	nnpClose := "</NNP>"
	lemma := "New York"
	newTok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil))
	newTok.AddReading(languagetool.NewAnalyzedToken("New", &nnpOpen, &lemma), "MULTIWORD_CHUNKER")
	yorkTok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil))
	yorkTok.AddReading(languagetool.NewAnalyzedToken("York", &nnpClose, &lemma), "MULTIWORD_CHUNKER")
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		newTok,
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		yorkTok,
	}
	out := removePreviousTags(toks)
	// New → NNP only (original tags removed)
	var newTags []string
	for _, r := range out[1].GetReadings() {
		if r.GetPOSTag() != nil {
			newTags = append(newTags, *r.GetPOSTag())
		}
	}
	require.Equal(t, []string{"NNP"}, newTags)
	// York → NNP (nextPOSTag same as posTag for non-NC tags)
	var yorkTags []string
	for _, r := range out[3].GetReadings() {
		if r.GetPOSTag() != nil {
			yorkTags = append(yorkTags, *r.GetPOSTag())
		}
	}
	require.Equal(t, []string{"NNP"}, yorkTags)
}

func TestGetMultiWordAnalyzedToken_CleanTagJavaSubstring(t *testing.T) {
	// Java cleanTag = substring(1, length-2): "<NPCN000>" → "NPCN00" (not "NPCN000").
	// Low-priority check never matches → equal-distance selection still prefers this tag.
	np := "<NPCN000>"
	nnp := "<NNP>"
	lemma := "x y"
	// Two candidates at same distance: prefer non-low-priority after Java cleanTag mangle.
	// Build tokens: start | "a" with both open tags | " " | "b" with both close tags
	a := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("a", nil, nil))
	a.AddReading(languagetool.NewAnalyzedToken("a", &np, &lemma), "t")
	a.AddReading(languagetool.NewAnalyzedToken("a", &nnp, &lemma), "t")
	b := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("b", nil, nil))
	npC, nnpC := "</NPCN000>", "</NNP>"
	b.AddReading(languagetool.NewAnalyzedToken("b", &npC, &lemma), "t")
	b.AddReading(languagetool.NewAnalyzedToken("b", &nnpC, &lemma), "t")
	toks := []*languagetool.AnalyzedTokenReadings{sentenceStart(), a, b}
	// i=1 is "a"
	sel := getMultiWordAnalyzedToken(toks, 1)
	require.NotNil(t, sel)
	// With equal distance, last non-low-priority (after Java cleanTag) wins: both treated as non-low
	// due to cleanTag mangle for NPCN000, so NNP (second candidate) selected when distance equal.
	require.Equal(t, "<NNP>", *sel.GetPOSTag())
}
