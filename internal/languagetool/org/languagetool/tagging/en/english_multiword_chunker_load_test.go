package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenEnglishMultiWordChunker_Settings(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenEnglishMultiWordChunker applies Java constructor settings).
	// Official en/multiwords.txt uses tab-separated phrase\ttag lines (default separator).
	r := strings.NewReader("Foo Bar\tNNP\n")
	c, err := OpenEnglishMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.RemovePreviousTags, "Java EnglishHybridDisambiguator.setRemovePreviousTags(true)")
	require.True(t, c.AddIgnoreSpelling, "Java EnglishHybridDisambiguator.setIgnoreSpelling(true)")
	require.Contains(t, c.Lines, "Foo Bar\tNNP")
}

func TestEnglishMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverEnglishMultiwords() == "" {
		t.Skip("official en/multiwords.txt not discoverable")
	}
	a := EnglishMultiWordChunker()
	b := EnglishMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "New York Post\tNNP")
	require.Contains(t, a.Lines, "Taj Mahal\tNNP")
	require.Contains(t, a.Lines, "status quo\tNN")
	require.Contains(t, a.Lines, "quid pro quo\tNN")
	require.Contains(t, a.Lines, "Qur'an\tNNP")
	require.True(t, a.RemovePreviousTags)
	require.True(t, a.AddIgnoreSpelling)

	// Wired on DefaultEnglishHybridDisambiguator (process-cached hybrid)
	d := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, d)
	require.NotNil(t, d.Chunker)
	require.Same(t, a, d.Chunker)
}

// multiwordTokens builds SENT_START + alternating content/space tokens for MultiWordChunker.
func multiwordTokens(parts ...string) []*languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for i, p := range parts {
		if i > 0 {
			toks = append(toks, languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)))
		}
		toks = append(toks, languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(p, nil, nil)))
	}
	return toks
}

// multiwordNoSpaceTokens builds SENT_START + consecutive content tokens (no whitespace)
// for no-space multiwords like Qur'an (Java mFullNoSpace path).
func multiwordNoSpaceTokens(parts ...string) []*languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for _, p := range parts {
		toks = append(toks, languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(p, nil, nil)))
	}
	return toks
}

func contentPOSTags(out *languagetool.AnalyzedSentence) [][]string {
	var all [][]string
	for i, tr := range out.GetTokens() {
		if i == 0 || tr.IsWhitespace() {
			continue
		}
		var tags []string
		for _, r := range tr.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
		all = append(all, tags)
	}
	return all
}

func hasExactPOS(tags []string, want string) bool {
	for _, p := range tags {
		if p == want {
			return true
		}
	}
	return false
}

func hasAnyAnglePOS(tags []string) bool {
	for _, p := range tags {
		if strings.Contains(p, "<") {
			return true
		}
	}
	return false
}

func requireAllContentIgnored(t *testing.T, out *languagetool.AnalyzedSentence, want bool, label string) {
	t.Helper()
	toks := out.GetTokens()
	for i, tr := range toks {
		if i == 0 || tr.IsWhitespace() {
			continue
		}
		if want {
			require.True(t, tr.IsIgnoredBySpeller(), "%s token[%d]=%q", label, i, tr.GetToken())
		} else {
			require.False(t, tr.IsIgnoredBySpeller(), "%s token[%d]=%q should NOT ignore", label, i, tr.GetToken())
		}
	}
}

func TestEnglishMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverEnglishMultiwords() == "" {
		t.Skip("official en/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage (do not re-claim GlobalChunker / Rules).
	c := EnglishMultiWordChunker()
	require.NotNil(t, c)
	d := &EnglishHybridDisambiguator{Chunker: c}

	// POS after setRemovePreviousTags(true): English multiwords use NNP/NN/NNS (not NC*/N ),
	// so getNextPosTag returns the same tag on every content token of the span.
	// Java fillMaps last-write-wins only for keys re-inserted: later "status quo\tNN"
	// overwrites the exact key "status quo", but first-cap/all-upper variants already in
	// the map from earlier "status quo\tNN:UN" are not re-emitted (tokenLettercaseVariants
	// skips keys already present) — so those stay NN:UN (bug-for-bug with Java).
	type phraseCase struct {
		parts    []string
		wantTags []string // one expected tag per content token, in order
		label    string
	}
	positives := []phraseCase{
		{[]string{"New", "York", "Post"}, []string{"NNP", "NNP", "NNP"}, "New York Post"},
		{[]string{"Taj", "Mahal"}, []string{"NNP", "NNP"}, "Taj Mahal"},
		{[]string{"quid", "pro", "quo"}, []string{"NN", "NN", "NN"}, "quid pro quo"},
		{[]string{"status", "quo"}, []string{"NN", "NN"}, "status quo"},
		// allowFirstCapitalized=true: first-cap of lowercase official entry
		{[]string{"Quid", "pro", "quo"}, []string{"NN", "NN", "NN"}, "Quid pro quo first-cap"},
		// Status quo first-cap was inserted with NN:UN and not overwritten (see above)
		{[]string{"Status", "quo"}, []string{"NN:UN", "NN:UN"}, "Status quo first-cap NN:UN"},
		// Official listed "Status Quo" later line overwrites to NN
		{[]string{"Status", "Quo"}, []string{"NN", "NN"}, "Status Quo listed"},
		// allowAllUppercase=true
		{[]string{"NEW", "YORK", "POST"}, []string{"NNP", "NNP", "NNP"}, "NEW YORK POST all-upper"},
		{[]string{"TAJ", "MAHAL"}, []string{"NNP", "NNP"}, "TAJ MAHAL all-upper"},
		{[]string{"QUID", "PRO", "QUO"}, []string{"NN", "NN", "NN"}, "QUID PRO QUO all-upper"},
		// STATUS QUO all-upper stays NN:UN (variant from early line, not overwritten)
		{[]string{"STATUS", "QUO"}, []string{"NN:UN", "NN:UN"}, "STATUS QUO all-upper NN:UN"},
		// ad hoc (official; last exact-key write is RB) + first-cap variant keeps earlier JJ
		{[]string{"ad", "hoc"}, []string{"RB", "RB"}, "ad hoc"},
		{[]string{"Ad", "hoc"}, []string{"JJ", "JJ"}, "Ad hoc first-cap"},
	}
	for _, tc := range positives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		require.Len(t, got, len(tc.wantTags), "%s content token count", tc.label)
		for i, want := range tc.wantTags {
			require.True(t, hasExactPOS(got[i], want),
				"%s token[%d] want %q in %v", tc.label, i, want, got[i])
			require.False(t, hasAnyAnglePOS(got[i]),
				"%s token[%d] angle-bracket chunk tags should be flattened by removePreviousTags: %v",
				tc.label, i, got[i])
		}
		requireAllContentIgnored(t, out, true, tc.label+" ignore spelling")
	}

	// No-space multiword: Qur'an (official line Qur'an\tNNP) — tokens without intervening space.
	outQ := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordNoSpaceTokens("Qur'", "an")))
	gotQ := contentPOSTags(outQ)
	require.Len(t, gotQ, 2, "Qur'an content tokens")
	require.True(t, hasExactPOS(gotQ[0], "NNP"), "Qur' want NNP in %v", gotQ[0])
	require.True(t, hasExactPOS(gotQ[1], "NNP"), "an want NNP in %v", gotQ[1])
	require.False(t, hasAnyAnglePOS(gotQ[0]) || hasAnyAnglePOS(gotQ[1]), "Qur'an flattened")
	requireAllContentIgnored(t, outQ, true, "Qur'an ignore spelling")

	// Negatives: non-listed sequences must not receive multiword POS or ignore-spelling.
	// Note: "Quid Pro Quo" / "Status Quo" / "Status Quo Ante" are official listed lines.
	// allowTitlecase=false: "Ad Hoc" is not generated from "ad hoc" (only first-cap "Ad hoc").
	negatives := []struct {
		parts []string
		label string
	}{
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed"},
		// all-lower of title-cased official proper name is not generated as a variant
		{[]string{"new", "york", "post"}, "new york post all-lower denied"},
		{[]string{"taj", "mahal"}, "taj mahal all-lower denied"},
		// wrong middle casing of proper name when not listed
		{[]string{"New", "york", "Post"}, "New york Post mixed denied"},
		// allowTitlecase=false: full titlecase of lower official entry denied
		{[]string{"Ad", "Hoc"}, "Ad Hoc titlecase denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasExactPOS(tags, "NNP") || hasExactPOS(tags, "NN") ||
				hasExactPOS(tags, "NNS") || hasExactPOS(tags, "NN:UN") ||
				hasExactPOS(tags, "JJ") || hasExactPOS(tags, "RB") || hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
		requireAllContentIgnored(t, out, false, tc.label)
	}

	// DefaultEnglishHybridDisambiguator wires Chunker and still tags official phrases.
	full := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("New", "York", "Post")))
	got := contentPOSTags(out)
	require.Len(t, got, 3)
	require.True(t, hasExactPOS(got[0], "NNP"), "wired hybrid New: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NNP"), "wired hybrid York: %v", got[1])
	require.True(t, hasExactPOS(got[2], "NNP"), "wired hybrid Post: %v", got[2])
	requireAllContentIgnored(t, out, true, "wired hybrid New York Post ignore")
}
