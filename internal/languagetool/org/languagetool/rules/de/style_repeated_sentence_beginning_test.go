package de

// Twin of org.languagetool.rules.de.StyleRepeatedSentenceBeginning
// (no Java *Test class — behavior matrix + real de-DE analysis when GermanTagger supplies POS).
//
// Java-critical:
//   MIN_REPEATED=3; tokens[1] ART:DEF:NOM|ART:IND:NOM (SUB span) or PRO:PER:NOM (pronoun-only);
//   cumulative pos via getCorrectedTextLength; AnalyzePlain fail-closed (POS only).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

func artNomSent(art, noun, restVerb string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS(art, "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS(noun, "SUB:NOM:SIN:NEU", noun),
		atrWithPOS(restVerb, "VER:3:SIN:PRT:NON", restVerb),
		atrWithPOS(".", "PKT", "."),
	))
}

func artIndNomSent(art, noun, restVerb string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS(art, "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS(noun, "SUB:NOM:SIN:MAS", noun),
		atrWithPOS(restVerb, "VER:3:SIN:PRT:NON", restVerb),
		atrWithPOS(".", "PKT", "."),
	))
}

// artNomOnlySent: ART:DEF:NOM then VER with no SUB in between → Java noSub end at article.
func artNomOnlySent(art, restVerb string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS(art, "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS(restVerb, "VER:3:SIN:PRT:NON", restVerb),
		atrWithPOS(".", "PKT", "."),
	))
}

func proNomSent(pro, restVerb string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS(pro, "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS(restVerb, "VER:3:SIN:PRT:NON", restVerb),
		atrWithPOS(".", "PKT", "."),
	))
}

func TestStyleRepeatedSentenceBeginning_Meta(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	require.Equal(t, "STYLE_REPEATED_SENTENCE_BEGINNING", rule.GetID())
	require.Equal(t, "Subjekt als wiederholter Satzanfang", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 3, rule.MinRepeated)
	require.Equal(t, 3, rule.MinToCheckParagraph())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.CategoryInternal, rule.GetCategory().GetLocation())
	require.True(t, rule.GetCategory().IsDefaultOff())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	// Java addExamplePair multi-marker wrong / fixed without markers
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Contains(t, inc[0].GetExample(), "<marker>Das Auto</marker>")
	require.Contains(t, inc[0].GetExample(), "<marker>Der Hund</marker>")
	require.Contains(t, inc[0].GetExample(), "<marker>Die Reifen</marker>")
	cor := rule.GetCorrectExamples()
	require.Len(t, cor, 1)
	require.Equal(t,
		"Das Auto kam näher. Langsam lief der Hund über die Straße. Die Reifen quietschten.",
		cor[0].GetExample())
	require.NotContains(t, cor[0].GetExample(), "<marker>")

	// messages map → category name from category_creative_writing
	msgs := map[string]string{"category_creative_writing": "Kreatives Schreiben"}
	r2 := NewStyleRepeatedSentenceBeginning(msgs)
	require.Equal(t, "Kreatives Schreiben", r2.GetCategory().GetName())
}

func TestStyleRepeatedSentenceBeginning_ArtDefNomStreak(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// Java example shape: three sentences starting with ART:DEF:NOM → marker ART…SUB
	sents := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
		artNomSent("Die", "Reifen", "quietschten"),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 3, len(matches))
	for _, m := range matches {
		require.Equal(t, "Subjekt als wiederholter Satzanfang", m.GetMessage())
	}
	// first match local (pos=0): ART tokens[1] … SUB tokens[2]
	toks0 := sents[0].GetTokensWithoutWhitespace()
	require.Equal(t, toks0[1].GetStartPos(), matches[0].GetFromPos())
	require.Equal(t, toks0[2].GetEndPos(), matches[0].GetToPos())
}

func TestStyleRepeatedSentenceBeginning_ArtIndNomStreak(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	sents := []*languagetool.AnalyzedSentence{
		artIndNomSent("Ein", "Hund", "bellte"),
		artIndNomSent("Eine", "Katze", "schlief"),
		artIndNomSent("Ein", "Auto", "fuhr"),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 3, len(matches))
	toks0 := sents[0].GetTokensWithoutWhitespace()
	require.Equal(t, toks0[1].GetStartPos(), matches[0].GetFromPos())
	require.Equal(t, toks0[2].GetEndPos(), matches[0].GetToPos())
}

func TestStyleRepeatedSentenceBeginning_ArtNoSubEndsAtArticle(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// ART then VER (no SUB before VER) → Java noSub: endPos = tokens[1].endPos + pos
	sents := []*languagetool.AnalyzedSentence{
		artNomOnlySent("Das", "geschieht"),
		artNomOnlySent("Der", "folgt"),
		artNomOnlySent("Die", "beginnt"),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 3, len(matches))
	toks0 := sents[0].GetTokensWithoutWhitespace()
	require.Equal(t, toks0[1].GetStartPos(), matches[0].GetFromPos())
	require.Equal(t, toks0[1].GetEndPos(), matches[0].GetToPos(), "noSub: span article only")
}

func TestStyleRepeatedSentenceBeginning_ProPerNomStreak(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	proSents := []*languagetool.AnalyzedSentence{
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
			atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen"),
			atrWithPOS(".", "PKT", "."),
		)),
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Sie", "PRO:PER:NOM:SIN:FEM", "sie"),
			atrWithPOS("rief", "VER:3:SIN:PRT:SFT", "rufen"),
			atrWithPOS(".", "PKT", "."),
		)),
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
			atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen"),
			atrWithPOS(".", "PKT", "."),
		)),
	}
	msPro := rule.MatchList(proSents)
	require.Equal(t, 3, len(msPro))
	// first span = pronoun only
	pt := proSents[0].GetTokensWithoutWhitespace()
	require.Equal(t, pt[1].GetStartPos(), msPro[0].GetFromPos())
	require.Equal(t, pt[1].GetEndPos(), msPro[0].GetToPos())
}

func TestStyleRepeatedSentenceBeginning_StreakBroken(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// mixed starts — no streak of 3 (middle not ART/PRO:PER:NOM)
	sents2 := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Langsam", "ADV", "langsam"),
			atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen"),
			atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
			atrWithPOS("Hund", "SUB:NOM:SIN:MAS", "Hund"),
			atrWithPOS(".", "PKT", "."),
		)),
		artNomSent("Die", "Reifen", "quietschten"),
	}
	require.Equal(t, 0, len(rule.MatchList(sents2)))
}

func TestStyleRepeatedSentenceBeginning_BelowMinRepeated(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// fewer than MIN_REPEATED → 0 (early return on len < 3)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
	})))
	// empty / single
	require.Equal(t, 0, len(rule.MatchList(nil)))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
	})))
}

func TestStyleRepeatedSentenceBeginning_CumulativePos(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	sents := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
		artNomSent("Die", "Reifen", "quietschten"),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 3, len(matches))

	// Java: pos accumulates getCorrectedTextLength(); start/end = token local + pos
	pos := 0
	for i, s := range sents {
		toks := s.GetTokensWithoutWhitespace()
		require.Equal(t, toks[1].GetStartPos()+pos, matches[i].GetFromPos(), "match %d FromPos", i)
		require.Equal(t, toks[2].GetEndPos()+pos, matches[i].GetToPos(), "match %d ToPos (ART…SUB)", i)
		pos += s.GetCorrectedTextLength()
	}
	// absolute offsets strictly increase across sentences
	require.Greater(t, matches[1].GetFromPos(), matches[0].GetToPos())
	require.Greater(t, matches[2].GetFromPos(), matches[1].GetToPos())
}

func TestStyleRepeatedSentenceBeginning_FlushOnBreakThenNewStreak(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// 3 ART, break (ADV), 3 PRO → 6 matches (two streaks flushed)
	adv := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Langsam", "ADV", "langsam"),
		atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen"),
		atrWithPOS(".", "PKT", "."),
	))
	sents := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
		artNomSent("Die", "Reifen", "quietschten"),
		adv,
		proNomSent("Er", "lief"),
		proNomSent("Sie", "rief"),
		proNomSent("Es", "ging"),
	}
	ms := rule.MatchList(sents)
	require.Equal(t, 6, len(ms))
}

func TestStyleRepeatedSentenceBeginning_AnalyzePlainFailClosed(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	// untagged AnalyzePlain must not invent article/pronoun matches (POS only)
	plain := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund lief langsam über die Straße."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	require.Equal(t, 0, len(rule.MatchList(plain)))
}

func setUpStyleRepeatedSentenceBeginning(t *testing.T) (*StyleRepeatedSentenceBeginning, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for StyleRepeatedSentenceBeginning twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")
	return NewStyleRepeatedSentenceBeginning(nil), lt, tagger
}

func TestStyleRepeatedSentenceBeginning_RealGermanTagger(t *testing.T) {
	rule, lt, tagger := setUpStyleRepeatedSentenceBeginning(t)

	// ART:DEF:NOM ×3 with SUB spans (Java example family)
	defSents := analyzeTextWithGermanTagger(lt, tagger,
		"Das Auto kam näher. Der Hund lief langsam. Die Reifen quietschten.")
	msDef := rule.MatchList(defSents)
	require.Equal(t, 3, len(msDef))
	// cumulative absolute positions
	pos := 0
	for i, s := range defSents {
		toks := s.GetTokensWithoutWhitespace()
		require.True(t, toks[1].HasPosTagStartingWith("ART:DEF:NOM"), "tok1 DEF:NOM in sent %d", i)
		require.Equal(t, toks[1].GetStartPos()+pos, msDef[i].GetFromPos())
		// end at first SUB before VER
		wantTo := toks[1].GetEndPos() + pos
		for j := 2; j < len(toks) && toks[j] != nil && !toks[j].HasPosTagStartingWith("VER"); j++ {
			if toks[j].HasPosTagStartingWith("SUB") {
				wantTo = toks[j].GetEndPos() + pos
				break
			}
		}
		require.Equal(t, wantTo, msDef[i].GetToPos())
		pos += s.GetCorrectedTextLength()
	}

	// ART:IND:NOM streak
	indSents := analyzeTextWithGermanTagger(lt, tagger,
		"Ein Hund bellte. Eine Katze schlief. Ein Auto fuhr.")
	require.Equal(t, 3, len(rule.MatchList(indSents)))

	// PRO:PER:NOM ×3 — pronoun-only spans
	proSents := analyzeTextWithGermanTagger(lt, tagger, "Er lief. Sie rief. Es ging.")
	msPro := rule.MatchList(proSents)
	require.Equal(t, 3, len(msPro))
	pt := proSents[0].GetTokensWithoutWhitespace()
	require.Equal(t, pt[1].GetStartPos(), msPro[0].GetFromPos())
	require.Equal(t, pt[1].GetEndPos(), msPro[0].GetToPos())

	// streak broken by ADV start
	broken := analyzeTextWithGermanTagger(lt, tagger,
		"Das Auto kam näher. Langsam lief der Hund. Die Reifen quietschten.")
	require.Equal(t, 0, len(rule.MatchList(broken)))
}

// TestStyleRepeatedSentenceBeginning keeps the historical smoke entrypoint name used by package scans.
func TestStyleRepeatedSentenceBeginning(t *testing.T) {
	rule := NewStyleRepeatedSentenceBeginning(nil)
	require.Equal(t, "STYLE_REPEATED_SENTENCE_BEGINNING", rule.GetID())
	sents := []*languagetool.AnalyzedSentence{
		artNomSent("Das", "Auto", "kam"),
		artNomSent("Der", "Hund", "lief"),
		artNomSent("Die", "Reifen", "quietschten"),
	}
	require.Equal(t, 3, len(rule.MatchList(sents)))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam."),
		languagetool.AnalyzePlain("Der Hund lief."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	})))
}
