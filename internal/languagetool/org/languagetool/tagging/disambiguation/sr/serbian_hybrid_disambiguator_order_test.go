package sr

// Outcome twins for SerbianHybridDisambiguator full stage order:
// Java SerbianHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(input))
// i.e. MultiWordChunker("/sr/multiwords.txt") getInstance defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling) FIRST, then
// XmlRuleDisambiguator(new Serbian(), useGlobal=false).
//
// CRITICAL: multiword→XML (same Romance order as AR/GL/ES/RU;
// opposite of Polish/Swedish XML→multiword).
// Official sr/multiwords.txt is empty → multiword stage is a no-op
// (empty maps; still eagerly wired). Do not invent multiword entries.
// XML stage still applies real official sr/disambiguation.xml outcomes
// (RIMSKI_BROJEVI ignore_spelling on case_sensitive Roman numerals).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tagsr "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/sr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func requireSRHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverSerbianMultiwords() == "" {
		t.Skip("official sr/multiwords.txt not discoverable")
	}
	if DiscoverSerbianDisambiguationXML() == "" {
		t.Skip("official sr/disambiguation.xml not discoverable")
	}
}

func requireSRHybridResourcesWithDict(t *testing.T) {
	t.Helper()
	requireSRHybridResources(t)
	if tagsr.DiscoverEkavianPOSDict() == "" {
		t.Skip("ekavian serbian.dict not in tree")
	}
	tagsr.EnsureDefaultEkavianTagger()
	require.NotNil(t, tagsr.DefaultEkavianTagger)
	require.NotNil(t, tagsr.DefaultEkavianTagger.GetWordTagger())
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

// taggedSRSentence ports the tagging half of Java TestTools.myAssert for SR:
// WordTokenizer + SRXSentenceTokenizer("sr") + EkavianTagger → AnalyzedSentence.
func taggedSRSentence(input string) *languagetool.AnalyzedSentence {
	tagsr.EnsureDefaultEkavianTagger()
	tagger := tagsr.DefaultEkavianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("sr")
	var sentence string
	for _, s := range st.Tokenize(input) {
		sentence = s
		break
	}
	tokens := wt.Tokenize(sentence)
	var noWS []string
	for _, tok := range tokens {
		if srOrderIsWord(tok) {
			noWS = append(noWS, tok)
		}
	}
	aTokens := tagger.Tag(noWS)
	tokenArray := make([]*languagetool.AnalyzedTokenReadings, 0, len(tokens)+1)
	ss := languagetool.SentenceStartTagName
	tokenArray = append(tokenArray, languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", &ss, nil), 0))
	startPos := 0
	noWSCount := 0
	for _, tokenStr := range tokens {
		var posTag *languagetool.AnalyzedTokenReadings
		if srOrderIsWord(tokenStr) {
			posTag = aTokens[noWSCount]
			posTag.SetStartPos(startPos)
			noWSCount++
		} else {
			posTag = languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(tokenStr, nil, nil), startPos)
		}
		tokenArray = append(tokenArray, posTag)
		startPos += tokenizers.UTF16Len(tokenStr)
	}
	return languagetool.NewAnalyzedSentence(tokenArray)
}

func srOrderIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func formatSROrderSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				lemma, pos := "null", "null"
				if r.GetLemma() != nil {
					lemma = *r.GetLemma()
				}
				if r.GetPOSTag() != nil {
					pos = *r.GetPOSTag()
				}
				readings = append(readings, r.GetToken()+"/["+lemma+"]"+pos)
			}
		}
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
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

// TestNewSerbianHybridDisambiguator_WiresBothStages proves Java constructor
// eagerly builds multiwords Chunker and XmlRuleDisambiguator when the same
// official resources Java loads are present — with Serbian flags.
// Official multiwords may be empty; Chunker is still non-nil (Java constructs).
func TestNewSerbianHybridDisambiguator_WiresBothStages(t *testing.T) {
	requireSRHybridResources(t)

	mw := SerbianMultiWordChunker()
	xml := SerbianXmlRuleDisambiguator()
	require.NotNil(t, mw, "Java MultiWordChunker even when multiwords empty")
	require.NotNil(t, xml)

	d := NewSerbianHybridDisambiguator()
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker(/sr/multiwords.txt) defaults F,F,F")
	require.NotNil(t, d.Rules,
		"disambiguator = new XmlRuleDisambiguator(new Serbian()) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Serbian multiwords defaults (no invent):
	// NO setRemovePreviousTags, NO setIgnoreSpelling
	require.False(t, mw.RemovePreviousTags, "Serbian multiwords does NOT setRemovePreviousTags")
	require.False(t, mw.AddIgnoreSpelling, "Serbian multiwords does NOT setIgnoreSpelling")
	// Official file empty → empty phrase list (still loaded).
	require.Empty(t, mw.Lines, "official sr/multiwords.txt empty; no invent phrases")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official SR pack: exactly 1 rule (RIMSKI_BROJEVI).
	require.Equal(t, 1, len(xml.Rules),
		"Serbian XmlRuleDisambiguator must load official sr/disambiguation.xml (1 rule)")
	require.Equal(t, "RIMSKI_BROJEVI", xml.Rules[0].GetID())
}

// TestSerbianHybridDisambiguator_OrderMultiwordThenXML proves stage isolation vs
// full Java order. Multiword stage is no-op (empty official multiwords); XML
// still applies real disambiguation.xml outcomes. Order is multiword first → XML.
func TestSerbianHybridDisambiguator_OrderMultiwordThenXML(t *testing.T) {
	requireSRHybridResourcesWithDict(t)

	mw := SerbianMultiWordChunker()
	xml := SerbianXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyMulti := &SerbianHybridDisambiguator{Chunker: mw}
	onlyXML := &SerbianHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(sent))
	}
	full := NewSerbianHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Multiword no-op: empty official multiwords invent no angle POS ---
	{
		for _, parts := range [][]string{
			{"XX", "vek"},
			{"foo", "bar"},
			{"random", "phrase"},
		} {
			label := strings.Join(parts, " ")
			fresh := func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens(parts...))
			}
			for _, got := range [][][]string{
				contentPOSTags(onlyMulti.Disambiguate(fresh())),
				contentPOSTags(full.Disambiguate(fresh())),
				contentPOSTags(javaOrder(fresh())),
			} {
				for i, tags := range got {
					require.False(t, hasAnyAnglePOS(tags),
						"%s multiword no invent token[%d]: %v", label, i, tags)
				}
			}
			// No setIgnoreSpelling on multiwords (XML may ignore Roman surfaces separately).
			for i, tr := range onlyMulti.Disambiguate(fresh()).GetTokens() {
				if i == 0 || tr.IsWhitespace() {
					continue
				}
				require.False(t, tr.IsIgnoredBySpeller(),
					"%s multiword-only token %q must not ignore spelling", label, tr.GetToken())
			}
		}
	}

	// --- (2) XML RIMSKI_BROJEVI: XII → ignore_spelling ---
	// Multiword alone → no ignore; XML alone / full / javaOrder → same XML outcome.
	{
		label := "XII"
		fresh := func() *languagetool.AnalyzedSentence { return taggedSRSentence("XII") }

		// Multiword-only leaves ignore_spelling false (empty multiwords no-op).
		requireNotIgnored(t, onlyMulti.Disambiguate(fresh()), "XII")

		// XML alone applies RIMSKI_BROJEVI.
		requireIgnored(t, onlyXML.Disambiguate(fresh()), "XII")

		// Full hybrid (multiword then XML) matches XML (multiword empty no-op).
		requireIgnored(t, full.Disambiguate(fresh()), "XII")
		requireIgnored(t, javaOrder(fresh()), "XII")

		// ignore_spelling does not immunize / rewrite readings.
		fullOut := full.Disambiguate(fresh())
		require.False(t, tokenBySurface(fullOut, "XII").IsImmunized())
		// Readings format parity multiword-only vs full (flag only).
		require.Equal(t, formatSROrderSentence(onlyMulti.Disambiguate(fresh())),
			formatSROrderSentence(full.Disambiguate(fresh())),
			"%s ignore_spelling must not alter reading strings", label)

		// Leave multiword out: XML-only hybrid still ignores.
		noMulti := &SerbianHybridDisambiguator{Rules: xml}
		requireIgnored(t, noMulti.Disambiguate(fresh()), "XII")

		// Leave XML out: multiword-only does not ignore.
		noXML := &SerbianHybridDisambiguator{Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(fresh()), "XII")
	}

	// --- (3) XML RIMSKI_BROJEVI in context: "vek XX je" — only XX ignored ---
	{
		label := "vek XX je"
		fresh := func() *languagetool.AnalyzedSentence { return taggedSRSentence("vek XX je") }
		fullOut := full.Disambiguate(fresh())
		javaOut := javaOrder(fresh())
		xmlOut := onlyXML.Disambiguate(fresh())
		mwOut := onlyMulti.Disambiguate(fresh())

		requireIgnored(t, fullOut, "XX")
		requireNotIgnored(t, fullOut, "vek", "je")
		requireIgnored(t, javaOut, "XX")
		requireNotIgnored(t, javaOut, "vek", "je")
		requireIgnored(t, xmlOut, "XX")
		requireNotIgnored(t, mwOut, "XX", "vek", "je")

		// Format parity full vs javaOrder
		require.Equal(t, formatSROrderSentence(fullOut), formatSROrderSentence(javaOut), label+" format")
	}

	// --- (4) case_sensitive: lowercase roman must NOT match ---
	{
		fresh := func() *languagetool.AnalyzedSentence { return taggedSRSentence("xii") }
		requireNotIgnored(t, full.Disambiguate(fresh()), "xii")
		requireNotIgnored(t, onlyXML.Disambiguate(fresh()), "xii")
		requireNotIgnored(t, onlyMulti.Disambiguate(fresh()), "xii")
	}

	// --- (5) Control unmatched stays clean ---
	{
		fresh := func() *languagetool.AnalyzedSentence { return taggedSRSentence("Zdravo svete") }
		for _, got := range []*languagetool.AnalyzedSentence{
			full.Disambiguate(fresh()),
			onlyXML.Disambiguate(fresh()),
			onlyMulti.Disambiguate(fresh()),
		} {
			requireNotIgnored(t, got, "Zdravo", "svete")
		}
	}
}

// TestSerbianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockXML proves
// empty multiword stage first does not block subsequent XML outcomes.
func TestSerbianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockXML(t *testing.T) {
	requireSRHybridResourcesWithDict(t)
	full := NewSerbianHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// Multiword angles never appear (empty official multiwords; no invent).
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("XX", "vek")))
	for i, tags := range contentPOSTags(out) {
		require.False(t, hasAnyAnglePOS(tags), "no invent multiword POS token[%d]: %v", i, tags)
	}

	// XML effects still fire after multiword (multiword no-op on these surfaces).
	for _, roman := range []string{"I", "XII", "XX", "MCMXCIX", "MMXX"} {
		sent := full.Disambiguate(taggedSRSentence(roman))
		requireIgnored(t, sent, roman)
		require.False(t, tokenBySurface(sent, roman).IsImmunized())
	}

	// Phrase: only Roman ignored.
	phrase := full.Disambiguate(taggedSRSentence("vek XX je"))
	requireIgnored(t, phrase, "XX")
	requireNotIgnored(t, phrase, "vek", "je")

	// Lowercase roman still not matched after multiword stage.
	requireNotIgnored(t, full.Disambiguate(taggedSRSentence("xii")), "xii")
}

// TestSerbianHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == XML(multiword(input)) for official isolation surfaces.
func TestSerbianHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireSRHybridResourcesWithDict(t)
	mw := SerbianMultiWordChunker()
	xml := SerbianXmlRuleDisambiguator()
	full := NewSerbianHybridDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		ignored  []string
		clean    []string
		noAngles bool
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("foo", "bar")) },
			label:    "foo bar random",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("random", "phrase")) },
			label:    "random phrase",
			noAngles: true,
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return taggedSRSentence("XII") },
			label:   "XII tagged",
			ignored: []string{"XII"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return taggedSRSentence("MCMXCIX") },
			label:   "MCMXCIX tagged",
			ignored: []string{"MCMXCIX"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return taggedSRSentence("vek XX je") },
			label:   "vek XX je tagged",
			ignored: []string{"XX"},
			clean:   []string{"vek", "je"},
		},
		{
			fresh: func() *languagetool.AnalyzedSentence { return taggedSRSentence("xii") },
			label: "xii lowercase tagged",
			clean: []string{"xii"},
		},
		{
			fresh: func() *languagetool.AnalyzedSentence { return taggedSRSentence("Zdravo svete") },
			label: "Zdravo svete tagged",
			clean: []string{"Zdravo", "svete"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return taggedSRSentence("XXI i MCMXCIX") },
			label:   "XXI i MCMXCIX tagged",
			ignored: []string{"XXI", "MCMXCIX"},
			clean:   []string{"i"},
		},
	}
	for _, tc := range cases {
		outFull := full.Disambiguate(tc.fresh())
		// Java: disambiguator.disambiguate(chunker.disambiguate(input))
		outManual := xml.Disambiguate(mw.Disambiguate(tc.fresh()))

		gotFull := contentPOSTags(outFull)
		gotManual := contentPOSTags(outManual)
		require.Equal(t, len(gotFull), len(gotManual), tc.label+" content POS count")

		if tc.noAngles {
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
		// Full format parity (readings + surfaces).
		require.Equal(t, formatSROrderSentence(outFull), formatSROrderSentence(outManual),
			"%s format parity full vs javaOrder", tc.label)

		if len(tc.ignored) > 0 {
			requireIgnored(t, outFull, tc.ignored...)
			requireIgnored(t, outManual, tc.ignored...)
		}
		if len(tc.clean) > 0 {
			requireNotIgnored(t, outFull, tc.clean...)
			requireNotIgnored(t, outManual, tc.clean...)
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
	}
}

// TestSerbianHybridDisambiguator_StageOrderIsMultiwordThenXML proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestSerbianHybridDisambiguator_StageOrderIsMultiwordThenXML(t *testing.T) {
	requireSRHybridResourcesWithDict(t)
	mw := SerbianMultiWordChunker()
	xml := SerbianXmlRuleDisambiguator()
	full := NewSerbianHybridDisambiguator()

	// Multiword surface: empty official multiwords → no invent POS from any stage combo.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("foo", "bar"))
		}
		for _, got := range [][][]string{
			contentPOSTags(full.Disambiguate(fresh())),
			contentPOSTags((&SerbianHybridDisambiguator{Chunker: mw}).Disambiguate(fresh())),
			contentPOSTags((&SerbianHybridDisambiguator{Rules: xml}).Disambiguate(fresh())),
		} {
			for i, tags := range got {
				require.False(t, hasAnyAnglePOS(tags), "no invent angle POS token[%d]: %v", i, tags)
			}
		}
	}

	// XML-only surface: only Rules sets ignore_spelling (RIMSKI_BROJEVI on XII).
	{
		fresh := func() *languagetool.AnalyzedSentence { return taggedSRSentence("XII") }

		requireIgnored(t, full.Disambiguate(fresh()), "XII")
		requireNotIgnored(t, (&SerbianHybridDisambiguator{Chunker: mw}).Disambiguate(fresh()), "XII")
		requireIgnored(t, (&SerbianHybridDisambiguator{Rules: xml}).Disambiguate(fresh()), "XII")
		// Without Chunker → XML still applies
		requireIgnored(t, (&SerbianHybridDisambiguator{Rules: xml}).Disambiguate(fresh()), "XII")
	}

	// Call-order: Chunker then Rules (Java nested call: outer=disambiguator, inner=chunker).
	{
		var order []string
		rulesStub := &srOrderStage{name: "rules", order: &order}
		chunkStub := &srOrderStage{name: "chunker", order: &order}
		d := &SerbianHybridDisambiguator{Rules: rulesStub, Chunker: chunkStub}
		d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("x")))
		require.Equal(t, []string{"chunker", "rules"}, order,
			"Java: disambiguator.disambiguate(chunker.disambiguate(input)) → multiword then XML")
	}
}

// srOrderStage records Disambiguate call order for stage-order proof.
type srOrderStage struct {
	name  string
	order *[]string
}

func (s *srOrderStage) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s.order != nil {
		*s.order = append(*s.order, s.name)
	}
	return input
}

var _ interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
} = (*srOrderStage)(nil)
