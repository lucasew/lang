package en

import (
	"testing"

	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

func TestEnsureDefaultEnglishTagger_POS(t *testing.T) {
	if DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	EnsureDefaultEnglishTagger()
	require.NotEmpty(t, EnglishPOSDictPath())
	tw := EnglishTagWord()
	require.NotNil(t, tw)

	// house / houses
	tags := tw("houses")
	require.NotEmpty(t, tags)
	found := false
	for _, tg := range tags {
		if tg.POS == "NNS" && tg.Lemma == "house" {
			found = true
		}
	}
	require.True(t, found, "%+v", tags)

	// Capitalized How includes WRB from lowercase (Java EnglishTagger)
	tags = tw("How")
	var hasWRB bool
	for _, tg := range tags {
		if tg.POS == "WRB" {
			hasWRB = true
		}
	}
	require.True(t, hasWRB, "%+v", tags)

	// IsTaggedEN wired for tokenizer
	require.NotNil(t, entok.IsTaggedEN)
	require.True(t, entok.IsTaggedEN("doin'"), "doin' in english.dict")
	// keep doin' whole
	toks := entok.NewEnglishWordTokenizer().Tokenize("doin' that")
	require.Equal(t, []string{"doin'", " ", "that"}, toks)
}

func TestAnalyzeEnglishSentence_POSTags(t *testing.T) {
	if DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	sent := AnalyzeEnglishSentence("Houses are big.")
	require.NotNil(t, sent)
	// find Houses token with NNS or similar
	var houses *string
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		if tok.GetToken() == "Houses" || tok.GetToken() == "houses" {
			// should have real POS from tagger
			require.True(t, tok.IsTagged(), "Houses should be tagged")
			houses = new(string)
			*houses = tok.GetToken()
		}
	}
	require.NotNil(t, houses)
}
