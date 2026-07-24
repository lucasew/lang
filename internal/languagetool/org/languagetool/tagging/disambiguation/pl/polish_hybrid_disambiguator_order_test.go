package pl

// Outcome twins for PolishHybridDisambiguator full stage order:
// Java PolishHybridDisambiguator.disambiguate:
//   chunker.disambiguate(disambiguator.disambiguate(input))
// i.e. XmlRuleDisambiguator(Polish, useGlobal=false) FIRST, then
// MultiWordChunker.getInstance("/pl/multiwords.txt") defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling).
//
// CRITICAL: inverted vs Romance hybrids (ES/FR/NL/EN: multiword→XML).
// Official polish.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/XML stage leaves).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requirePLHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverPolishMultiwords() == "" {
		t.Skip("official pl/multiwords.txt not discoverable")
	}
	if DiscoverPolishDisambiguationXML() == "" {
		t.Skip("official pl/disambiguation.xml not discoverable")
	}
}

// tokenSentence builds SENT_START + tokens with spaces between word tokens.
func tokenSentence(words ...string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(multiwordTokens(words...))
}

// tokenSentenceNoSpace builds SENT_START + adjacent tokens with spacebefore=no
// (for patterns like 90° single token or URI pieces).
func tokenSentenceNoSpace(words ...string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(words...))
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

// TestNewPolishHybridDisambiguator_WiresBothStages proves Java constructor
// eagerly builds multiwords Chunker and XmlRuleDisambiguator when the same
// official resources Java loads are present — with Polish flags.
func TestNewPolishHybridDisambiguator_WiresBothStages(t *testing.T) {
	requirePLHybridResources(t)

	mw := PolishMultiWordChunker()
	xml := PolishXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewPolishHybridDisambiguator()
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/pl/multiwords.txt) defaults F,F,F")
	require.NotNil(t, d.Rules,
		"disambiguator = new XmlRuleDisambiguator(new Polish()) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Polish multiwords defaults (no invent):
	// NO setRemovePreviousTags, NO setIgnoreSpelling
	// allowFirstCapitalized/allowAllUppercase/allowTitlecase false (outcome-tested elsewhere)
	require.False(t, mw.RemovePreviousTags, "Polish multiwords does NOT setRemovePreviousTags")
	require.False(t, mw.AddIgnoreSpelling, "Polish multiwords does NOT setIgnoreSpelling")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official PL pack is large; smoke that it loaded as a real rule set.
	require.GreaterOrEqual(t, len(xml.Rules), 200,
		"Polish XmlRuleDisambiguator must load official pl/disambiguation.xml rules")
}

// TestPolishHybridDisambiguator_OrderXMLThenMultiword proves stage isolation vs
// full Java order with Java-visible POS / ignore_spelling outcomes.
// Order is XML first → multiword second (NOT multiword→XML).
func TestPolishHybridDisambiguator_OrderXMLThenMultiword(t *testing.T) {
	requirePLHybridResources(t)

	mw := PolishMultiWordChunker()
	xml := PolishXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyMulti := &PolishHybridDisambiguator{Chunker: mw}
	onlyXML := &PolishHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return mw.Disambiguate(xml.Disambiguate(sent))
	}
	// Reverse of Java (Romance-style multiword→XML) — used for leave-one-out / order contrast.
	reverseOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(sent))
	}
	full := NewPolishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Multiword-only phrase: "to znaczy" in pl/multiwords, not an XML ignore rule ---
	// Multiword alone → <TO_ZNACZY> / </TO_ZNACZY> (no removePreviousTags)
	// XML alone → no multiword POS
	// Full hybrid (XML then multiword) → multiword POS (multiword last)
	// Without multiword → no multiword POS
	{
		label := "to znaczy"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("to", "znaczy"))
		}

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "<TO_ZNACZY>"), "%s multiword-only to: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</TO_ZNACZY>"), "%s multiword-only znaczy: %v", label, gotM[1])

		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<TO_ZNACZY>") || hasExactPOS(tags, "</TO_ZNACZY>") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "<TO_ZNACZY>"),
			"%s full hybrid to (XML then multiword): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "</TO_ZNACZY>"),
			"%s full hybrid znaczy: %v", label, gotFull[1])

		gotJO := contentPOSTags(javaOrder(fresh()))
		require.True(t, hasExactPOS(gotJO[0], "<TO_ZNACZY>"), "%s javaOrder to: %v", label, gotJO[0])
		require.True(t, hasExactPOS(gotJO[1], "</TO_ZNACZY>"), "%s javaOrder znaczy: %v", label, gotJO[1])

		// Leave multiword out: XML-only hybrid has no multiword POS.
		noMulti := &PolishHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		// No setIgnoreSpelling on Polish multiwords.
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid token %q must not ignore spelling via multiwords", label, tr.GetToken())
		}

		// Reverse order still keeps multiword POS here (XML does not wipe TO_ZNACZY on this surface).
		// Order proof for independent stages is composition + leave-one-out + stage call order.
		gotRev := contentPOSTags(reverseOrder(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<TO_ZNACZY>") || hasExactPOS(gotRev[0], "TO_ZNACZY"),
			"%s reverse still has multiword tag (stages independent on this surface): %v", label, gotRev[0])
	}

	// Multiword 3-token: "z uwagi na" → PREP:ACC open/close
	{
		label := "z uwagi na"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("z", "uwagi", "na"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 3, label)
		require.True(t, hasExactPOS(gotFull[0], "<PREP:ACC>"), "%s full z: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[2], "</PREP:ACC>"), "%s full na: %v", label, gotFull[2])
		require.False(t, hasAnyAnglePOS(gotFull[1]), "%s interior uwagi: %v", label, gotFull[1])

		// Without multiword → no PREP:ACC angles
		noMulti := &PolishHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "<PREP:ACC>") || hasExactPOS(tags, "</PREP:ACC>"),
				"%s without multi token[%d]: %v", label, i, tags)
		}
	}

	// No-space multiword ELLIPSIS
	{
		label := "..."
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", "."))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 3, label)
		require.True(t, hasExactPOS(gotFull[0], "<ELLIPSIS>"), "%s full open: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[2], "</ELLIPSIS>"), "%s full close: %v", label, gotFull[2])

		// multiword alone same; XML alone no ELLIPSIS
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "<ELLIPSIS>"), "%s multi-only: %v", label, gotM[0])
		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<ELLIPSIS>") || hasExactPOS(tags, "</ELLIPSIS>"),
				"%s xml-only token[%d]: %v", label, i, tags)
		}
	}

	// Casing flags: allowFirstCapitalized=false on multiwords
	{
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("To", "znaczy"))
		}
		got := contentPOSTags(full.Disambiguate(freshCap()))
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"To znaczy first-cap denied full hybrid token[%d]: %v", i, tags)
		}
	}

	// --- (2) XML-only effects: pl ignore_spelling rules (token-built; no polish.dict) ---
	// Chunker does not set ignore_spelling; XML stage does.
	{
		// degree: 90° / 10°C
		sentDeg := tokenSentence("90°")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentDeg), "90°")
		requireIgnored(t, onlyXML.Disambiguate(sentDeg), "90°")
		requireIgnored(t, full.Disambiguate(sentDeg), "90°")
		requireIgnored(t, javaOrder(sentDeg), "90°")

		// leave-one-out: without Rules, multiword alone must not ignore
		noXML := &PolishHybridDisambiguator{Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("90°")), "90°")

		// all inclusive (official pl ignore_spelling multiword expression)
		sentAI := tokenSentence("all", "inclusive")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentAI), "all")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentAI), "inclusive")
		requireIgnored(t, onlyXML.Disambiguate(sentAI), "all", "inclusive")
		requireIgnored(t, full.Disambiguate(sentAI), "all", "inclusive")

		// call center
		sentCC := tokenSentence("call", "center")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentCC), "call")
		requireIgnored(t, full.Disambiguate(sentCC), "call", "center")

		// prez. abbreviation pattern: prez + .
		sentPrez := tokenSentenceNoSpace("prez", ".")
		// pattern: prez|cz|... then spacebefore=no .
		// tokenSentenceNoSpace sets adjacent tokens
		outPrez := full.Disambiguate(sentPrez)
		// At least one of the matched tokens is ignored (whole pattern ignore_spelling)
		// Java ignore_spelling applies to the matched pattern tokens.
		prezTok := tokenBySurface(outPrez, "prez")
		require.NotNil(t, prezTok)
		require.True(t, prezTok.IsIgnoredBySpeller(), "prez. must ignore_spelling via XML")
	}
}

// TestPolishHybridDisambiguator_XMLBeforeMultiword_DoesNotBlockMultiwordPOS proves
// multiword last (after XML) still attaches official multiword open/close POS.
func TestPolishHybridDisambiguator_XMLBeforeMultiword_DoesNotBlockMultiwordPOS(t *testing.T) {
	requirePLHybridResources(t)
	full := NewPolishHybridDisambiguator()

	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("to", "znaczy")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<TO_ZNACZY>"), "to after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</TO_ZNACZY>"), "znaczy after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("co", "do")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<PREP:GEN>"), "co after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</PREP:GEN>"), "do after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", ".")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<ELLIPSIS>"), "... open after full: %v", got[0])
	require.True(t, hasExactPOS(got[2], "</ELLIPSIS>"), "... close after full: %v", got[2])

	// XML ignore still fires (XML first, multiword no-op on this surface).
	requireIgnored(t, full.Disambiguate(tokenSentence("90°")), "90°")
	requireIgnored(t, full.Disambiguate(tokenSentence("all", "inclusive")), "all", "inclusive")
}

// TestPolishHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == multiword(XML(input)) for official isolation surfaces.
func TestPolishHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requirePLHybridResources(t)
	mw := PolishMultiWordChunker()
	xml := PolishXmlRuleDisambiguator()
	full := NewPolishHybridDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantPOS0 string // empty = no multiword angle POS required
		wantPOS1 string
		wantIg   map[string]bool // surface → ignore
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("to", "znaczy")) },
			label:    "to znaczy",
			wantPOS0: "<TO_ZNACZY>",
			wantPOS1: "</TO_ZNACZY>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("to", "jest")) },
			label:    "to jest",
			wantPOS0: "<TO_JEST>",
			wantPOS1: "</TO_JEST>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("z", "uwagi", "na")) },
			label:    "z uwagi na",
			wantPOS0: "<PREP:ACC>",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", "."))
			},
			label:    "...",
			wantPOS0: "<ELLIPSIS>",
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("To", "znaczy")) },
			label:  "To znaczy first-cap denied",
			wantIg: map[string]bool{},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Zxqwv", "Plmnb")) },
			label:  "random non-listed",
			wantIg: map[string]bool{},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("90°") },
			label:  "90°",
			wantIg: map[string]bool{"90°": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("all", "inclusive") },
			label:  "all inclusive",
			wantIg: map[string]bool{"all": true, "inclusive": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("call", "center") },
			label:  "call center",
			wantIg: map[string]bool{"call": true, "center": true},
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
		// Ignore parity.
		ft, mt := outFull.GetTokens(), outManual.GetTokens()
		require.Equal(t, len(ft), len(mt), tc.label+" token count")
		for i := range ft {
			if i == 0 || ft[i].IsWhitespace() {
				continue
			}
			require.Equal(t, ft[i].IsIgnoredBySpeller(), mt[i].IsIgnoredBySpeller(),
				"%s token[%d]=%q ignore parity full vs javaOrder", tc.label, i, ft[i].GetToken())
		}
		for surface, want := range tc.wantIg {
			trF := tokenBySurface(outFull, surface)
			trM := tokenBySurface(outManual, surface)
			require.NotNil(t, trF, "%s full missing %q", tc.label, surface)
			require.NotNil(t, trM, "%s manual missing %q", tc.label, surface)
			require.Equal(t, want, trF.IsIgnoredBySpeller(), "%s full %q", tc.label, surface)
			require.Equal(t, want, trM.IsIgnoredBySpeller(), "%s manual %q", tc.label, surface)
		}
	}
}

// TestPolishHybridDisambiguator_StageOrderIsXMLThenMultiword proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestPolishHybridDisambiguator_StageOrderIsXMLThenMultiword(t *testing.T) {
	requirePLHybridResources(t)
	mw := PolishMultiWordChunker()
	xml := PolishXmlRuleDisambiguator()
	full := NewPolishHybridDisambiguator()

	// Multiword-only surface: only Chunker produces multiword POS.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("to", "znaczy"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "<TO_ZNACZY>"), "full: %v", gotFull[0])

		// Without Chunker → no multiword POS
		noChunk := &PolishHybridDisambiguator{Rules: xml}
		gotNoC := contentPOSTags(noChunk.Disambiguate(fresh()))
		for i, tags := range gotNoC {
			require.False(t, hasAnyAnglePOS(tags), "without chunker token[%d]: %v", i, tags)
		}
		// Only Chunker → multiword POS
		onlyC := &PolishHybridDisambiguator{Chunker: mw}
		gotOnlyC := contentPOSTags(onlyC.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyC[0], "<TO_ZNACZY>"), "only chunker: %v", gotOnlyC[0])
	}

	// XML-only surface: only Rules sets ignore_spelling.
	{
		sent := tokenSentence("90°")
		requireIgnored(t, full.Disambiguate(sent), "90°")
		requireNotIgnored(t, (&PolishHybridDisambiguator{Chunker: mw}).Disambiguate(tokenSentence("90°")), "90°")
		requireIgnored(t, (&PolishHybridDisambiguator{Rules: xml}).Disambiguate(tokenSentence("90°")), "90°")
	}

	// Call-order: Rules then Chunker (Java nested call: outer=chunker, inner=disambiguator).
	{
		var order []string
		rulesStub := &orderStage{name: "rules", order: &order}
		chunkStub := &orderStage{name: "chunker", order: &order}
		d := &PolishHybridDisambiguator{Rules: rulesStub, Chunker: chunkStub}
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

