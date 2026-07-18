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
