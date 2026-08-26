package pt

// Outcome twins for PortugueseHybridDisambiguator full stage order:
// Java PortugueseHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
// i.e. spelling_global → /pt/multiwords.txt (setRemovePreviousTags + setIgnoreSpelling)
//     → XmlRuleDisambiguator(lang, true).
// Official portuguese.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/global/XML stage leaves).
//
// Differs from Spanish hybrid:
//   - allowAllUppercase=true on both chunkers; allowTitlecase=true on both
//   - setIgnoreSpelling(true) on BOTH chunkerGlobal and chunker (ES: neither)

import (
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requirePTHybridResources(t *testing.T) {
	t.Helper()
	// Official resources exist under inspiration pt module + core (same as Java).
	// Fail hard — no invent Skip when assets are in-tree (PolishWordTokenizer / PL Hybrid pattern).
	require.NotEmpty(t, DiscoverPortugueseGlobalChunker(),
		"official spelling_global.txt must be discoverable (inspiration core)")
	require.NotEmpty(t, DiscoverPortugueseMultiwords(),
		"official pt/multiwords.txt must be discoverable (inspiration pt module)")
	require.NotEmpty(t, DiscoverPortugueseDisambiguationXML(),
		"official pt/disambiguation.xml must be discoverable (inspiration pt module)")
	require.NotEmpty(t, DiscoverGlobalDisambiguationXML(),
		"official disambiguation-global.xml must be discoverable (inspiration core)")
}

// TestNewPortugueseHybridDisambiguator_WiresAllThreeStages proves Java constructor
// eagerly builds GlobalChunker, multiwords Chunker, and XmlRuleDisambiguator when
// the same official resources Java loads are present — with Portuguese flags.
func TestNewPortugueseHybridDisambiguator_WiresAllThreeStages(t *testing.T) {
	requirePTHybridResources(t)

	gPath := DiscoverPortugueseGlobalChunker()
	mwPath := DiscoverPortugueseMultiwords()
	ptXML := DiscoverPortugueseDisambiguationXML()
	gXML := DiscoverGlobalDisambiguationXML()
	require.Contains(t, gPath, "spelling_global.txt", "Java resource /spelling_global.txt")
	require.Contains(t, mwPath, "multiwords.txt", "Java resource /pt/multiwords.txt")
	require.Contains(t, ptXML, "disambiguation.xml", "Java resource /pt/disambiguation.xml")
	require.Contains(t, gXML, "disambiguation-global.xml", "Java resource disambiguation-global.xml")
	// Path under inspiration (same layout Java ships).
	require.Contains(t, gPath, "languagetool-core")
	require.Contains(t, mwPath, "languagetool-language-modules")
	require.Contains(t, mwPath, filepath.Join("resource", "pt"))
	require.Contains(t, ptXML, filepath.Join("resource", "pt"))

	g := PortugueseGlobalChunker()
	mw := PortugueseMultiWordChunker()
	xml := PortugueseXmlRuleDisambiguator()
	require.NotNil(t, g, "MultiWordChunker.getInstance(/spelling_global.txt, false, true, true, NPCN000) must load")
	require.NotNil(t, mw, "MultiWordChunker.getInstance(/pt/multiwords.txt, true, true, true) must load")
	require.NotNil(t, xml, "XmlRuleDisambiguator(lang, true) must load pt + global disambiguation XML")

	d := NewPortugueseHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker,
		"chunkerGlobal = MultiWordChunker.getInstance(/spelling_global.txt, false, true, true, NPCN000)")
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/pt/multiwords.txt, true, true, true) + setRemovePreviousTags + setIgnoreSpelling")
	require.NotNil(t, d.Rules, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.Same(t, g, d.GlobalChunker)
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Portuguese settings (no invent): ignoreSpelling on BOTH chunkers; multiwords removePreviousTags.
	require.True(t, g.AddIgnoreSpelling, "Portuguese chunkerGlobal.setIgnoreSpelling(true)")
	require.False(t, g.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	require.True(t, mw.RemovePreviousTags, "Portuguese chunker.setRemovePreviousTags(true)")
	require.True(t, mw.AddIgnoreSpelling, "Portuguese chunker.setIgnoreSpelling(true)")

	// useGlobal=true: language XML + disambiguation-global pack.
	require.GreaterOrEqual(t, len(xml.Rules), 200,
		"Portuguese XmlRuleDisambiguator must load official pt + global disambiguation rules")
	// Process-cached like Java static MultiWordChunker.getInstance + field finals.
	require.Same(t, PortugueseGlobalChunker(), g)
	require.Same(t, PortugueseMultiWordChunker(), mw)
	require.Same(t, PortugueseXmlRuleDisambiguator(), xml)
}

// TestPortugueseHybridDisambiguator_OrderGlobalThenMultiwordThenXML proves stage
// isolation vs full Java order with Java-visible POS / ignore_spelling outcomes.
func TestPortugueseHybridDisambiguator_OrderGlobalThenMultiwordThenXML(t *testing.T) {
	requirePTHybridResources(t)

	g := PortugueseGlobalChunker()
	mw := PortugueseMultiWordChunker()
	xml := PortugueseXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	// Isolation hybrids (only one stage) vs full order.
	onlyGlobal := &PortugueseHybridDisambiguator{GlobalChunker: g}
	onlyMulti := &PortugueseHybridDisambiguator{Chunker: mw}
	onlyXML := &PortugueseHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(g.Disambiguate(sent)))
	}
	// Reverse chunk stages for order proof (multiword then global — opposite of Java).
	reverseChunks := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return g.Disambiguate(mw.Disambiguate(sent))
	}
	full := NewPortugueseHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Global-only phrase: "Microsoft Entra" in spelling_global, not pt/multiwords ---
	// Global alone → open/close <NPCN000></NPCN000> + ignore_spelling
	// Multiword alone → no multiword POS, no ignore
	// Full hybrid (global→multiword removePreviousTags) → plain NPCN000 NPCN000 + ignore
	// Wrong chunk order (multiword→global) → angle tags remain (global last, no flatten after)
	{
		label := "Microsoft Entra"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		require.Len(t, gotG, 2, label)
		require.True(t, hasExactPOS(gotG[0], "<NPCN000>"), "%s global-only open: %v", label, gotG[0])
		require.True(t, hasExactPOS(gotG[1], "</NPCN000>"), "%s global-only close: %v", label, gotG[1])
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global-only ignore")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		for i, tags := range gotM {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasExactPOS(tags, "<NPCN000>") ||
				hasExactPOS(tags, "</NPCN000>") || hasAnyAnglePOS(tags),
				"%s multiword-only token[%d] must have no global/multiword POS, got %v", label, i, tags)
		}
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), false, label+" multiword-only no ignore")

		// XML alone: no POS invent for this surface; no angle tags
		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no chunk POS, got %v", label, i, tags)
		}

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "NPCN000"),
			"%s full hybrid Microsoft flattened (global then removePreviousTags): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPCN000"),
			"%s full hybrid Entra flattened: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must not keep angle tags: %v %v", label, gotFull[0], gotFull[1])
		// PT: both chunkers setIgnoreSpelling — full hybrid ignores matched global phrases.
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")

		gotJO := contentPOSTags(javaOrder(fresh()))
		require.True(t, hasExactPOS(gotJO[0], "NPCN000"), "%s javaOrder Microsoft: %v", label, gotJO[0])
		require.True(t, hasExactPOS(gotJO[1], "NPCN000"), "%s javaOrder Entra: %v", label, gotJO[1])
		requireAllContentIgnored(t, javaOrder(fresh()), true, label+" javaOrder ignore")

		// Reverse chunk order differs: multiword no-op then global leaves angles
		gotRev := contentPOSTags(reverseChunks(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<NPCN000>"),
			"%s reverse order leaves open angle (proves multiword-after-global needed): %v", label, gotRev[0])
		require.True(t, hasExactPOS(gotRev[1], "</NPCN000>"),
			"%s reverse order leaves close angle: %v", label, gotRev[1])
	}

	// --- (2) Multiword-only phrase: "Bin Laden" in pt/multiwords.txt, not spelling_global ---
	// Official line: Bin Laden\tNPMS000 → after removePreviousTags: NPMS000 NPMS000 + ignore
	{
		label := "Bin Laden"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Bin", "Laden"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "NPMS000") || hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no ignore")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "NPMS000"), "%s multiword-only Bin: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NPMS000"), "%s multiword-only Laden: %v", label, gotM[1])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]))
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), true, label+" multiword-only ignore")

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "NPMS000"), "%s full hybrid Bin: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPMS000"), "%s full hybrid Laden: %v", label, gotFull[1])
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")

		// Leave multiword out: global+XML hybrid has no multiword POS.
		noMulti := &PortugueseHybridDisambiguator{GlobalChunker: g, Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "NPMS000") || hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}
	}

	// Multiword NC phrase (Romance getNextPosTag): fair play;NCMS000 → NCMS000 AQ0MS0
	{
		label := "fair play"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("fair", "play"))
		}
		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "NCMS000") || hasExactPOS(tags, "AQ0MS0") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NCMS000"), "%s multiword-only fair: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "AQ0MS0"), "%s multiword-only play: %v", label, gotM[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NCMS000"), "%s full hybrid fair: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "AQ0MS0"), "%s full hybrid play: %v", label, gotFull[1])
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")
	}

	// --- (3) Shared phrase: "Peter Pan" listed in BOTH global and multiwords ---
	// Official multiwords: Peter Pan;NPMS000_
	// Official global: Peter Pan → NPCN000 open/close
	// Java order: global tags first, multiword adds NPMS000_, removePreviousTags → NPMS000_
	// (multiword tag preferred over low-priority NPCN000 when both present).
	{
		label := "Peter Pan"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotG[0], "<NPCN000>"), "%s global-only: %v", label, gotG[0])
		require.True(t, hasExactPOS(gotG[1], "</NPCN000>"), "%s global-only: %v", label, gotG[1])

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NPMS000_"), "%s multiword-only: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NPMS000_"), "%s multiword-only: %v", label, gotM[1])

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPMS000_"),
			"%s full hybrid prefers multiword NPMS000_ after global+removePreviousTags: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPMS000_"),
			"%s full hybrid Pan: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must flatten angles: %v %v", label, gotFull[0], gotFull[1])
		// Must not leave only NPCN000 from global without multiword tag.
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")

		// Without multiword, global leaves angles (or flattened only if multiword ran).
		noMulti := &PortugueseHybridDisambiguator{GlobalChunker: g, Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotNoM[0], "<NPCN000>"),
			"%s without multiword keeps global open angle: %v", label, gotNoM[0])
		require.True(t, hasExactPOS(gotNoM[1], "</NPCN000>"),
			"%s without multiword keeps global close angle: %v", label, gotNoM[1])
	}

	// --- (4) XML-only effects: pt UNIVERSITY_OF + global proper noun ---
	// Chunkers do not match these surfaces; XML stage sets ignore_spelling.
	{
		// pt rule: case_sensitive University of
		sentUni := tokenSentence("University", "of")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentUni), "University", "of")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentUni), "University", "of")
		requireIgnored(t, onlyXML.Disambiguate(sentUni), "University", "of")
		requireIgnored(t, full.Disambiguate(sentUni), "University", "of")
		requireIgnored(t, javaOrder(sentUni), "University", "of")

		// leave-one-out: without Rules, chunkers alone must not ignore
		noXML := &PortugueseHybridDisambiguator{GlobalChunker: g, Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("University", "of")), "University", "of")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// pt ROMAN_NUMBER_IGNORE_SPELLING: XIV (XML after chunkers still fires)
		sentXIV := tokenSentence("XIV")
		requireIgnored(t, full.Disambiguate(sentXIV), "XIV")
	}
}

// TestPortugueseHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkPOS proves
// XML last does not wipe multiword/global POS on official surfaces.
func TestPortugueseHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkPOS(t *testing.T) {
	requirePTHybridResources(t)
	full := NewPortugueseHybridDisambiguator()
	// Peter Pan multiword POS survives XML stage.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPMS000_"), "Peter after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMS000_"), "Pan after full order: %v", got[1])
	// Microsoft Entra flattened global POS survives XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")))
	got = contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPCN000"), "Microsoft after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPCN000"), "Entra after full order: %v", got[1])
	// Bin Laden multiword POS survives XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Bin", "Laden")))
	got = contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPMS000"), "Bin after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMS000"), "Laden after full order: %v", got[1])

	// XML ignore still fires (XML last, chunkers no-op on this surface).
	requireIgnored(t, full.Disambiguate(tokenSentence("University", "of")), "University", "of")
	requireIgnored(t, full.Disambiguate(tokenSentence("QB|LT")), "QB|LT")
}

// TestPortugueseHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == xml(mw(g(input))) for official isolation surfaces.
func TestPortugueseHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requirePTHybridResources(t)
	g := PortugueseGlobalChunker()
	mw := PortugueseMultiWordChunker()
	xml := PortugueseXmlRuleDisambiguator()
	full := NewPortugueseHybridDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantPOS0 string // empty = no multiword/global invent POS required
		wantPOS1 string
		wantIg   map[string]bool // surface → ignore
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")) },
			label:    "Microsoft Entra",
			wantPOS0: "NPCN000",
			wantPOS1: "NPCN000",
			wantIg:   map[string]bool{"Microsoft": true, "Entra": true},
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Bin", "Laden")) },
			label:    "Bin Laden",
			wantPOS0: "NPMS000",
			wantPOS1: "NPMS000",
			wantIg:   map[string]bool{"Bin": true, "Laden": true},
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("fair", "play")) },
			label:    "fair play",
			wantPOS0: "NCMS000",
			wantPOS1: "AQ0MS0",
			wantIg:   map[string]bool{"fair": true, "play": true},
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan")) },
			label:    "Peter Pan",
			wantPOS0: "NPMS000_",
			wantPOS1: "NPMS000_",
			wantIg:   map[string]bool{"Peter": true, "Pan": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Zxqwv", "Plmnb")) },
			label:  "random non-listed",
			wantIg: map[string]bool{},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("University", "of") },
			label:  "University of",
			wantIg: map[string]bool{"University": true, "of": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("QB|LT") },
			label:  "QB|LT",
			wantIg: map[string]bool{"QB|LT": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("XIV") },
			label:  "XIV",
			wantIg: map[string]bool{"XIV": true},
		},
	}
	for _, tc := range cases {
		outFull := full.Disambiguate(tc.fresh())
		// Java: disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
		outManual := xml.Disambiguate(mw.Disambiguate(g.Disambiguate(tc.fresh())))

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
		} else if len(tc.wantIg) == 0 {
			// Non-listed: no multiword/global invent angles required
			for i, tags := range gotFull {
				require.False(t, hasAnyAnglePOS(tags),
					"%s full token[%d] must have no chunk angle POS: %v", tc.label, i, tags)
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

// TestPortugueseHybridDisambiguator_StageOrderIsGlobalThenMultiwordThenXML proves each
// stage occupies its Java slot via leave-one-out isolation and call-order recording.
func TestPortugueseHybridDisambiguator_StageOrderIsGlobalThenMultiwordThenXML(t *testing.T) {
	requirePTHybridResources(t)
	g := PortugueseGlobalChunker()
	mw := PortugueseMultiWordChunker()
	xml := PortugueseXmlRuleDisambiguator()
	full := NewPortugueseHybridDisambiguator()

	// Global-only surface (then multiword flatten): only GlobalChunker+Chunker produce NPCN000.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPCN000"), "full: %v", gotFull[0])

		// Without GlobalChunker → no NPCN000
		noGlobal := &PortugueseHybridDisambiguator{Chunker: mw, Rules: xml}
		gotNoG := contentPOSTags(noGlobal.Disambiguate(fresh()))
		for i, tags := range gotNoG {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasExactPOS(tags, "<NPCN000>"),
				"without global token[%d]: %v", i, tags)
		}
		// Only GlobalChunker → open/close angles
		onlyG := &PortugueseHybridDisambiguator{GlobalChunker: g}
		gotOnlyG := contentPOSTags(onlyG.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyG[0], "<NPCN000>"), "only global: %v", gotOnlyG[0])
	}

	// Multiword-only surface: only Chunker produces multiword POS.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Bin", "Laden"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPMS000"), "full: %v", gotFull[0])

		noChunk := &PortugueseHybridDisambiguator{GlobalChunker: g, Rules: xml}
		gotNoC := contentPOSTags(noChunk.Disambiguate(fresh()))
		for i, tags := range gotNoC {
			require.False(t, hasExactPOS(tags, "NPMS000") || hasAnyAnglePOS(tags),
				"without chunker token[%d]: %v", i, tags)
		}
		onlyC := &PortugueseHybridDisambiguator{Chunker: mw}
		gotOnlyC := contentPOSTags(onlyC.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyC[0], "NPMS000"), "only chunker: %v", gotOnlyC[0])
	}

	// XML-only surface: only Rules sets ignore_spelling.
	{
		sent := tokenSentence("University", "of")
		requireIgnored(t, full.Disambiguate(sent), "University", "of")
		requireNotIgnored(t, (&PortugueseHybridDisambiguator{GlobalChunker: g, Chunker: mw}).Disambiguate(tokenSentence("University", "of")), "University", "of")
		requireIgnored(t, (&PortugueseHybridDisambiguator{Rules: xml}).Disambiguate(tokenSentence("University", "of")), "University", "of")
	}

	// Call-order: GlobalChunker then Chunker then Rules
	// (Java nested: outer=disambiguator, mid=chunker, inner=chunkerGlobal).
	{
		var order []string
		globalStub := &orderStage{name: "global", order: &order}
		chunkStub := &orderStage{name: "chunker", order: &order}
		rulesStub := &orderStage{name: "rules", order: &order}
		d := &PortugueseHybridDisambiguator{GlobalChunker: globalStub, Chunker: chunkStub, Rules: rulesStub}
		d.Disambiguate(tokenSentence("x"))
		require.Equal(t, []string{"global", "chunker", "rules"}, order,
			"Java: disambiguator(chunker(chunkerGlobal(input))) → global then multiword then XML")
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
