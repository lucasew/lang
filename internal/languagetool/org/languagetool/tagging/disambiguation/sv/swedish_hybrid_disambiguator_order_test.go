package sv

// Outcome twins for SwedishHybridDisambiguator full stage order:
// Java SwedishHybridDisambiguator.disambiguate:
//   chunker.disambiguate(disambiguator.disambiguate(input))
// i.e. XmlRuleDisambiguator(Swedish, useGlobal=false) FIRST, then
// MultiWordChunker.getInstance("/sv/multiwords.txt") defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling).
//
// CRITICAL: inverted vs Romance hybrids (ES/FR/NL/EN: multiword→XML).
// Same order as Polish. Official swedish.dict is not required: token-built
// AnalyzedSentence patterns (same helpers as ACCEPTed multiword/XML stage leaves).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireSVHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverSwedishMultiwords() == "" {
		t.Skip("official sv/multiwords.txt not discoverable")
	}
	if DiscoverSwedishDisambiguationXML() == "" {
		t.Skip("official sv/disambiguation.xml not discoverable")
	}
}

// tokenSentence builds SENT_START + tokens with spaces between word tokens.
func tokenSentence(words ...string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(multiwordTokens(words...))
}


func tokenBySurface(sent *languagetool.AnalyzedSentence, surface string) *languagetool.AnalyzedTokenReadings {
	if sent == nil {
		return nil
	}
	for _, tr := range sent.GetTokensWithoutWhitespace() {
		if tr != nil && tr.GetToken() == surface {
			return tr
		}
	}
	return nil
}

func requireIgnored(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.True(t, tr.IsIgnoredBySpeller(), "%q must be ignore_spelling", s)
	}
}

func requireNotIgnored(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.False(t, tr.IsIgnoredBySpeller(), "%q must not be ignore_spelling", s)
	}
}

func requireImmunized(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.True(t, tr.IsImmunized(), "%q must be immunized", s)
	}
}

func requireNotImmunized(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.False(t, tr.IsImmunized(), "%q must not be immunized", s)
	}
}

// TestNewSwedishHybridDisambiguator_WiresBothStages proves Java constructor
// eagerly builds multiwords Chunker and XmlRuleDisambiguator when the same
// official resources Java loads are present — with Swedish flags.
func TestNewSwedishHybridDisambiguator_WiresBothStages(t *testing.T) {
	requireSVHybridResources(t)

	mw := SwedishMultiWordChunker()
	xml := SwedishXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewSwedishHybridDisambiguator()
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/sv/multiwords.txt) defaults F,F,F")
	require.NotNil(t, d.Rules,
		"disambiguator = new XmlRuleDisambiguator(new Swedish()) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Swedish multiwords defaults (no invent):
	// NO setRemovePreviousTags, NO setIgnoreSpelling
	// allowFirstCapitalized/allowAllUppercase/allowTitlecase false (outcome-tested elsewhere)
	require.False(t, mw.RemovePreviousTags, "Swedish multiwords does NOT setRemovePreviousTags")
	require.False(t, mw.AddIgnoreSpelling, "Swedish multiwords does NOT setIgnoreSpelling")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official SV pack has ~30-40 ignore_spelling/immunize rules (ACCEPTed 3.A.4).
	require.GreaterOrEqual(t, len(xml.Rules), 30,
		"Swedish XmlRuleDisambiguator must load official sv/disambiguation.xml rules")
	require.LessOrEqual(t, len(xml.Rules), 40,
		"Swedish XmlRuleDisambiguator useGlobal=false should not append global pack")
}

// TestSwedishHybridDisambiguator_OrderXMLThenMultiword proves stage isolation vs
// full Java order with Java-visible POS / ignore_spelling outcomes.
// Order is XML first → multiword second (NOT multiword→XML).
func TestSwedishHybridDisambiguator_OrderXMLThenMultiword(t *testing.T) {
	requireSVHybridResources(t)

	mw := SwedishMultiWordChunker()
	xml := SwedishXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyMulti := &SwedishHybridDisambiguator{Chunker: mw}
	onlyXML := &SwedishHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return mw.Disambiguate(xml.Disambiguate(sent))
	}
	// Reverse of Java (Romance-style multiword→XML) — used for leave-one-out / order contrast.
	reverseOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(sent))
	}
	full := NewSwedishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Multiword-only phrase: "en passant" in sv/multiwords, not an XML invent POS ---
	// Multiword alone → <NN:OF:SIN:NOM:UTR> / </NN:OF:SIN:NOM:UTR> (no removePreviousTags)
	// XML alone → no multiword POS
	// Full hybrid (XML then multiword) → multiword POS (multiword last)
	// Without multiword → no multiword POS
	{
		label := "en passant"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("en", "passant"))
		}

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "<NN:OF:SIN:NOM:UTR>"), "%s multiword-only en: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</NN:OF:SIN:NOM:UTR>"), "%s multiword-only passant: %v", label, gotM[1])

		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<NN:OF:SIN:NOM:UTR>") || hasExactPOS(tags, "</NN:OF:SIN:NOM:UTR>") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "<NN:OF:SIN:NOM:UTR>"),
			"%s full hybrid en (XML then multiword): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "</NN:OF:SIN:NOM:UTR>"),
			"%s full hybrid passant: %v", label, gotFull[1])

		gotJO := contentPOSTags(javaOrder(fresh()))
		require.True(t, hasExactPOS(gotJO[0], "<NN:OF:SIN:NOM:UTR>"), "%s javaOrder en: %v", label, gotJO[0])
		require.True(t, hasExactPOS(gotJO[1], "</NN:OF:SIN:NOM:UTR>"), "%s javaOrder passant: %v", label, gotJO[1])

		// Leave multiword out: XML-only hybrid has no multiword POS.
		noMulti := &SwedishHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		// No setIgnoreSpelling on Swedish multiwords.
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid token %q must not ignore spelling via multiwords", label, tr.GetToken())
		}

		// Reverse order still keeps multiword POS here (XML does not wipe NN on this surface).
		// Order proof for independent stages is composition + leave-one-out + stage call order.
		gotRev := contentPOSTags(reverseOrder(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<NN:OF:SIN:NOM:UTR>") || hasExactPOS(gotRev[0], "NN:OF:SIN:NOM:UTR"),
			"%s reverse still has multiword tag (stages independent on this surface): %v", label, gotRev[0])
	}

	// Multiword 2-token PM:NOM: "Sri Lanka"
	{
		label := "Sri Lanka"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Sri", "Lanka"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "<PM:NOM>"), "%s full Sri: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "</PM:NOM>"), "%s full Lanka: %v", label, gotFull[1])

		// Without multiword → no PM:NOM angles
		noMulti := &SwedishHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "<PM:NOM>") || hasExactPOS(tags, "</PM:NOM>"),
				"%s without multi token[%d]: %v", label, i, tags)
		}
	}

	// No-space multiword ELLIPS
	{
		label := "..."
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", "."))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 3, label)
		require.True(t, hasExactPOS(gotFull[0], "<ELLIPS>"), "%s full open: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[2], "</ELLIPS>"), "%s full close: %v", label, gotFull[2])

		// multiword alone same; XML alone no ELLIPS
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "<ELLIPS>"), "%s multi-only: %v", label, gotM[0])
		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<ELLIPS>") || hasExactPOS(tags, "</ELLIPS>"),
				"%s xml-only token[%d]: %v", label, i, tags)
		}
	}

	// Casing flags: allowFirstCapitalized=false on multiwords
	{
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("En", "passant"))
		}
		got := contentPOSTags(full.Disambiguate(freshCap()))
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"En passant first-cap denied full hybrid token[%d]: %v", i, tags)
		}
	}

	// --- (2) Dual-stage: "ad hoc" is both multiword AB and XML ignore_spelling ---
	// Full hybrid (XML then multiword) must have BOTH ignore_spelling AND multiword POS.
	{
		label := "ad hoc dual-stage"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("ad", "hoc"))
		}

		// Multiword alone → angle POS, no ignore
		outM := onlyMulti.Disambiguate(fresh())
		gotM := contentPOSTags(outM)
		require.True(t, hasExactPOS(gotM[0], "<AB>"), "%s multi AB open: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</AB>"), "%s multi AB close: %v", label, gotM[1])
		requireNotIgnored(t, outM, "ad", "hoc")

		// XML alone → ignore_spelling, no multiword angle POS
		outX := onlyXML.Disambiguate(fresh())
		gotX := contentPOSTags(outX)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<AB>") || hasExactPOS(tags, "</AB>"),
				"%s xml-only token[%d] no multiword POS: %v", label, i, tags)
		}
		requireIgnored(t, outX, "ad", "hoc")

		// Full Java order: XML ignore first, multiword POS second → both present
		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		require.True(t, hasExactPOS(gotFull[0], "<AB>"), "%s full AB open: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "</AB>"), "%s full AB close: %v", label, gotFull[1])
		requireIgnored(t, outFull, "ad", "hoc")

		// javaOrder composition matches full
		outJO := javaOrder(fresh())
		gotJO := contentPOSTags(outJO)
		require.ElementsMatch(t, gotFull[0], gotJO[0], "%s POS0 full vs javaOrder", label)
		require.ElementsMatch(t, gotFull[1], gotJO[1], "%s POS1 full vs javaOrder", label)
		requireIgnored(t, outJO, "ad", "hoc")
	}

	// --- (3) XML-only effects: sv ignore_spelling / immunize rules (token-built; no swedish.dict) ---
	// Chunker does not set ignore_spelling or immunize; XML stage does.
	{
		// pièce de résistance — ignore_spelling only in XML
		sentPR := tokenSentence("pièce", "de", "résistance")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentPR), "pièce", "de", "résistance")
		requireIgnored(t, onlyXML.Disambiguate(sentPR), "pièce", "de", "résistance")
		requireIgnored(t, full.Disambiguate(sentPR), "pièce", "de", "résistance")
		requireIgnored(t, javaOrder(sentPR), "pièce", "de", "résistance")

		// leave-one-out: without Rules, multiword alone must not ignore
		noXML := &SwedishHybridDisambiguator{Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("pièce", "de", "résistance")),
			"pièce", "de", "résistance")

		// World of Warcraft — ignore_spelling
		sentWOW := tokenSentence("World", "of", "Warcraft")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentWOW), "World")
		requireIgnored(t, onlyXML.Disambiguate(sentWOW), "World", "of", "Warcraft")
		requireIgnored(t, full.Disambiguate(sentWOW), "World", "of", "Warcraft")

		// 60-sekunder single-token ignore
		sent60 := tokenSentence("60-sekunder")
		requireNotIgnored(t, onlyMulti.Disambiguate(sent60), "60-sekunder")
		requireIgnored(t, onlyXML.Disambiguate(sent60), "60-sekunder")
		requireIgnored(t, full.Disambiguate(sent60), "60-sekunder")

		// Sambal Oelek — immunize (XML only)
		sentSO := tokenSentence("Sambal", "Oelek")
		requireNotImmunized(t, onlyMulti.Disambiguate(sentSO), "Sambal", "Oelek")
		requireImmunized(t, onlyXML.Disambiguate(sentSO), "Sambal", "Oelek")
		requireImmunized(t, full.Disambiguate(sentSO), "Sambal", "Oelek")
		// leave-one-out without Rules
		requireNotImmunized(t, noXML.Disambiguate(tokenSentence("Sambal", "Oelek")), "Sambal", "Oelek")
	}
}

// TestSwedishHybridDisambiguator_XMLBeforeMultiword_DoesNotBlockMultiwordPOS proves
// multiword last (after XML) still attaches official multiword open/close POS.
func TestSwedishHybridDisambiguator_XMLBeforeMultiword_DoesNotBlockMultiwordPOS(t *testing.T) {
	requireSVHybridResources(t)
	full := NewSwedishHybridDisambiguator()

	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("en", "passant")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<NN:OF:SIN:NOM:UTR>"), "en after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</NN:OF:SIN:NOM:UTR>"), "passant after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("vice", "versa")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<AB>"), "vice after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</AB>"), "versa after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", ".")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<ELLIPS>"), "... open after full: %v", got[0])
	require.True(t, hasExactPOS(got[2], "</ELLIPS>"), "... close after full: %v", got[2])

	// Dual-stage ad hoc: multiword POS survives XML-first ignore_spelling
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("ad", "hoc")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<AB>"), "ad after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</AB>"), "hoc after full order: %v", got[1])
	requireIgnored(t, out, "ad", "hoc")

	// XML ignore still fires (XML first, multiword no-op on this surface).
	requireIgnored(t, full.Disambiguate(tokenSentence("pièce", "de", "résistance")),
		"pièce", "de", "résistance")
	requireIgnored(t, full.Disambiguate(tokenSentence("World", "of", "Warcraft")),
		"World", "of", "Warcraft")
	requireImmunized(t, full.Disambiguate(tokenSentence("Sambal", "Oelek")), "Sambal", "Oelek")
}

// TestSwedishHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == multiword(XML(input)) for official isolation surfaces.
func TestSwedishHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireSVHybridResources(t)
	mw := SwedishMultiWordChunker()
	xml := SwedishXmlRuleDisambiguator()
	full := NewSwedishHybridDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantPOS0 string // empty = no multiword angle POS required
		wantPOS1 string
		wantIg   map[string]bool // surface → ignore
		wantImm  map[string]bool // surface → immunize
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("en", "passant")) },
			label:    "en passant",
			wantPOS0: "<NN:OF:SIN:NOM:UTR>",
			wantPOS1: "</NN:OF:SIN:NOM:UTR>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Sri", "Lanka")) },
			label:    "Sri Lanka",
			wantPOS0: "<PM:NOM>",
			wantPOS1: "</PM:NOM>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("vice", "versa")) },
			label:    "vice versa",
			wantPOS0: "<AB>",
			wantPOS1: "</AB>",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", "."))
			},
			label:    "...",
			wantPOS0: "<ELLIPS>",
		},
		{
			// Dual-stage: multiword AB + XML ignore_spelling
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("ad", "hoc")) },
			label:    "ad hoc",
			wantPOS0: "<AB>",
			wantPOS1: "</AB>",
			wantIg:   map[string]bool{"ad": true, "hoc": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("En", "passant")) },
			label:  "En passant first-cap denied",
			wantIg: map[string]bool{},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Zxqwv", "Plmnb")) },
			label:  "random non-listed",
			wantIg: map[string]bool{},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("pièce", "de", "résistance") },
			label:  "pièce de résistance",
			wantIg: map[string]bool{"pièce": true, "de": true, "résistance": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("World", "of", "Warcraft") },
			label:  "World of Warcraft",
			wantIg: map[string]bool{"World": true, "of": true, "Warcraft": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("60-sekunder") },
			label:  "60-sekunder",
			wantIg: map[string]bool{"60-sekunder": true},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("Sambal", "Oelek") },
			label:   "Sambal Oelek",
			wantImm: map[string]bool{"Sambal": true, "Oelek": true},
		},
	}
	for _, tc := range cases {
		outFull := full.Disambiguate(tc.fresh())
		// Java: chunker.disambiguate(disambiguator.disambiguate(input))
		outManual := mw.Disambiguate(xml.Disambiguate(tc.fresh()))

		gotFull := contentPOSTags(outFull)
		gotManual := contentPOSTags(outManual)
		require.Equal(t, len(gotFull), len(gotManual), tc.label+" content POS count")
		if tc.wantPOS0 != "" {
			require.True(t, hasExactPOS(gotFull[0], tc.wantPOS0), "%s full POS0: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotManual[0], tc.wantPOS0), "%s manual POS0: %v", tc.label, gotManual[0])
			if tc.wantPOS1 != "" {
				require.True(t, hasExactPOS(gotFull[1], tc.wantPOS1), "%s full POS1: %v", tc.label, gotFull[1])
				require.True(t, hasExactPOS(gotManual[1], tc.wantPOS1), "%s manual POS1: %v", tc.label, gotManual[1])
			}
		} else {
			// Non-listed / first-cap / XML-only: no multiword invent angles required
			for i, tags := range gotFull {
				require.False(t, hasAnyAnglePOS(tags),
					"%s full token[%d] must have no multiword angle POS: %v", tc.label, i, tags)
			}
		}
		// Per-token POS membership parity.
		for i := range gotFull {
			require.ElementsMatch(t, gotFull[i], gotManual[i],
				"%s token[%d] POS parity full vs javaOrder", tc.label, i)
		}
		// Ignore / immunize parity.
		ft, mt := outFull.GetTokens(), outManual.GetTokens()
		require.Equal(t, len(ft), len(mt), tc.label+" token count")
		for i := range ft {
			if i == 0 || ft[i].IsWhitespace() {
				continue
			}
			require.Equal(t, ft[i].IsIgnoredBySpeller(), mt[i].IsIgnoredBySpeller(),
				"%s token[%d]=%q ignore parity full vs javaOrder", tc.label, i, ft[i].GetToken())
			require.Equal(t, ft[i].IsImmunized(), mt[i].IsImmunized(),
				"%s token[%d]=%q immunize parity full vs javaOrder", tc.label, i, ft[i].GetToken())
		}
		for surface, want := range tc.wantIg {
			trF := tokenBySurface(outFull, surface)
			trM := tokenBySurface(outManual, surface)
			require.NotNil(t, trF, "%s full missing %q", tc.label, surface)
			require.NotNil(t, trM, "%s manual missing %q", tc.label, surface)
			require.Equal(t, want, trF.IsIgnoredBySpeller(), "%s full %q ignore", tc.label, surface)
			require.Equal(t, want, trM.IsIgnoredBySpeller(), "%s manual %q ignore", tc.label, surface)
		}
		for surface, want := range tc.wantImm {
			trF := tokenBySurface(outFull, surface)
			trM := tokenBySurface(outManual, surface)
			require.NotNil(t, trF, "%s full missing %q", tc.label, surface)
			require.NotNil(t, trM, "%s manual missing %q", tc.label, surface)
			require.Equal(t, want, trF.IsImmunized(), "%s full %q immunize", tc.label, surface)
			require.Equal(t, want, trM.IsImmunized(), "%s manual %q immunize", tc.label, surface)
		}
	}
}

// TestSwedishHybridDisambiguator_StageOrderIsXMLThenMultiword proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestSwedishHybridDisambiguator_StageOrderIsXMLThenMultiword(t *testing.T) {
	requireSVHybridResources(t)
	mw := SwedishMultiWordChunker()
	xml := SwedishXmlRuleDisambiguator()
	full := NewSwedishHybridDisambiguator()

	// Multiword-only surface: only Chunker produces multiword POS.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("en", "passant"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "<NN:OF:SIN:NOM:UTR>"), "full: %v", gotFull[0])

		// Without Chunker → no multiword POS
		noChunk := &SwedishHybridDisambiguator{Rules: xml}
		gotNoC := contentPOSTags(noChunk.Disambiguate(fresh()))
		for i, tags := range gotNoC {
			require.False(t, hasAnyAnglePOS(tags), "without chunker token[%d]: %v", i, tags)
		}
		// Only Chunker → multiword POS
		onlyC := &SwedishHybridDisambiguator{Chunker: mw}
		gotOnlyC := contentPOSTags(onlyC.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyC[0], "<NN:OF:SIN:NOM:UTR>"), "only chunker: %v", gotOnlyC[0])
	}

	// XML-only surface: only Rules sets ignore_spelling.
	{
		sent := tokenSentence("pièce", "de", "résistance")
		requireIgnored(t, full.Disambiguate(sent), "pièce", "de", "résistance")
		requireNotIgnored(t, (&SwedishHybridDisambiguator{Chunker: mw}).Disambiguate(
			tokenSentence("pièce", "de", "résistance")), "pièce", "de", "résistance")
		requireIgnored(t, (&SwedishHybridDisambiguator{Rules: xml}).Disambiguate(
			tokenSentence("pièce", "de", "résistance")), "pièce", "de", "résistance")
	}

	// Call-order: Rules then Chunker (Java nested call: outer=chunker, inner=disambiguator).
	{
		var order []string
		rulesStub := &orderStage{name: "rules", order: &order}
		chunkStub := &orderStage{name: "chunker", order: &order}
		d := &SwedishHybridDisambiguator{Rules: rulesStub, Chunker: chunkStub}
		d.Disambiguate(tokenSentence("x"))
		require.Equal(t, []string{"rules", "chunker"}, order,
			"Java: chunker.disambiguate(disambiguator.disambiguate(input)) → XML then multiword")
	}
}

// orderStage records Disambiguate call order for stage-order proof.
type orderStage struct {
	name  string
	order *[]string
}

func (s *orderStage) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s.order != nil {
		*s.order = append(*s.order, s.name)
	}
	return input
}

var _ interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
} = (*orderStage)(nil)

