package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenPortugueseMultiWordChunker_Settings(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenPortugueseMultiWordChunker applies Java constructor settings).
	// Official pt/multiwords.txt uses tab-separated phrase\ttag lines (default separator).
	r := strings.NewReader("Foo Bar\tNPMS000\n")
	c, err := OpenPortugueseMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.RemovePreviousTags, "Java PortugueseHybridDisambiguator.setRemovePreviousTags(true)")
	require.True(t, c.AddIgnoreSpelling, "Java PortugueseHybridDisambiguator.setIgnoreSpelling(true)")
	require.Contains(t, c.Lines, "Foo Bar\tNPMS000")
}

func TestPortugueseMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverPortugueseMultiwords() == "" {
		t.Skip("official pt/multiwords.txt not discoverable")
	}
	a := PortugueseMultiWordChunker()
	b := PortugueseMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "Adobe Acrobat Reader\tNPMS000")
	require.Contains(t, a.Lines, "Câmara Municipal\tNPFS000")
	require.Contains(t, a.Lines, "Bin Laden\tNPMS000")
	require.Contains(t, a.Lines, "autómato celular\tNCMS000")
	require.Contains(t, a.Lines, "fair play\tNCMS000")
	require.True(t, a.RemovePreviousTags)
	require.True(t, a.AddIgnoreSpelling)

	// Wired on NewPortugueseHybridDisambiguator
	d := NewPortugueseHybridDisambiguator()
	require.NotNil(t, d.Chunker)
	require.Same(t, a, d.Chunker)
	// GlobalChunker wired when spelling_global.txt is discoverable
	if PortugueseGlobalChunker() != nil {
		require.NotNil(t, d.GlobalChunker)
		require.Same(t, PortugueseGlobalChunker(), d.GlobalChunker)
	}
	// Rules wired when official pt + global disambiguation XML load
	if PortugueseXmlRuleDisambiguator() != nil {
		require.NotNil(t, d.Rules)
		require.Same(t, PortugueseXmlRuleDisambiguator(), d.Rules)
	}
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

func TestPortugueseMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverPortugueseMultiwords() == "" {
		t.Skip("official pt/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage (do not re-claim GlobalChunker / Rules).
	c := PortugueseMultiWordChunker()
	require.NotNil(t, c)
	d := &PortugueseHybridDisambiguator{Chunker: c}

	// POS after setRemovePreviousTags(true):
	// multi-token NP tags → plain TAG on every content token of the span.
	// multi-token NC* tags → first token keeps NC…; subsequent use AQ0+gender/number+0
	// (Java MultiWordChunker.getNextPosTag for Romance NC prefixes).
	type phraseCase struct {
		parts    []string
		wantTags []string // one expected tag per content token, in order
		label    string
	}
	positives := []phraseCase{
		{[]string{"Adobe", "Acrobat", "Reader"}, []string{"NPMS000", "NPMS000", "NPMS000"}, "Adobe Acrobat Reader"},
		{[]string{"Câmara", "Municipal"}, []string{"NPFS000", "NPFS000"}, "Câmara Municipal"},
		{[]string{"Bin", "Laden"}, []string{"NPMS000", "NPMS000"}, "Bin Laden"},
		{[]string{"autómato", "celular"}, []string{"NCMS000", "AQ0MS0"}, "autómato celular"},
		// unique NCMS000 line (beta tester is duplicated later as NCCS000_ in the file)
		{[]string{"fair", "play"}, []string{"NCMS000", "AQ0MS0"}, "fair play"},
		// allowFirstCapitalized=true: first-cap of lowercase official entry
		{[]string{"Autómato", "celular"}, []string{"NCMS000", "AQ0MS0"}, "Autómato celular first-cap"},
		{[]string{"Fair", "play"}, []string{"NCMS000", "AQ0MS0"}, "Fair play first-cap"},
		// allowTitlecase=true (unlike ES): full titlecase of all-lower official entry
		{[]string{"Autómato", "Celular"}, []string{"NCMS000", "AQ0MS0"}, "Autómato Celular titlecase"},
		{[]string{"Fair", "Play"}, []string{"NCMS000", "AQ0MS0"}, "Fair Play titlecase"},
		// allowAllUppercase=true
		{[]string{"BIN", "LADEN"}, []string{"NPMS000", "NPMS000"}, "BIN LADEN all-upper"},
		{[]string{"CÂMARA", "MUNICIPAL"}, []string{"NPFS000", "NPFS000"}, "CÂMARA MUNICIPAL all-upper"},
		{[]string{"AUTÓMATO", "CELULAR"}, []string{"NCMS000", "AQ0MS0"}, "AUTÓMATO CELULAR all-upper"},
		{[]string{"ADOBE", "ACROBAT", "READER"}, []string{"NPMS000", "NPMS000", "NPMS000"}, "ADOBE ACROBAT READER all-upper"},
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

	// Negatives: non-listed sequences must not receive multiword POS or ignore-spelling
	negatives := []struct {
		parts []string
		label string
	}{
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed"},
		// all-lower of title-cased official proper name is not generated as a variant
		{[]string{"bin", "laden"}, "bin laden all-lower denied"},
		// wrong middle casing of proper name when not listed
		{[]string{"Bin", "laden"}, "Bin laden mixed denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasExactPOS(tags, "NPMS000") || hasExactPOS(tags, "NPFS000") ||
				hasExactPOS(tags, "NCMS000") || hasExactPOS(tags, "AQ0MS0") || hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
		requireAllContentIgnored(t, out, false, tc.label)
	}

	// NewPortugueseHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewPortugueseHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Bin", "Laden")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPMS000"), "wired hybrid Bin: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMS000"), "wired hybrid Laden: %v", got[1])
	requireAllContentIgnored(t, out, true, "wired hybrid Bin Laden")
}
