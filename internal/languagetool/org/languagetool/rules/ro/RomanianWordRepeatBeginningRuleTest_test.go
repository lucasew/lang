package ro

// Outcome twins for org.languagetool.rules.ro.RomanianWordRepeatBeginningRule.
// Java has no dedicated *Test.java; matrix follows Persian/Greek WRB twin pattern
// + pure-POS isAdverb contract from RomanianWordRepeatBeginningRule.java.
//
// Java setup:
//   JLanguageTool lt = new JLanguageTool(new Romanian());
//   RomanianWordRepeatBeginningRule.match(analyzed sentences)
//
// Go analysis path for this text-level rule:
//   NewJLanguageTool("ro") + SentenceTagger = DefaultRomanianTagger.Tag
//   (official romanian.dict under inspiration/.../resource/ro/)
//   rule.MatchList(sents) isolates ROMANIAN_WORD_REPEAT_BEGINNING_RULE outcomes.
//
// Messages: no MessagesBundle_ro — fallback MessagesBundle.properties keys
// used by WordRepeatBeginningRule (no invent soft strings).

import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	rotag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ro"
	"github.com/stretchr/testify/require"
)

// roWRBMessages ports MessagesBundle.properties keys used by WordRepeatBeginningRule
// (Romanian has no MessagesBundle_ro; Java falls back to default English bundle).
func roWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning":           "Successive sentences beginning with the same word",
		"desc_repetition_beginning_word":      "Three successive sentences begin with the same word.",
		"desc_repetition_beginning_adv":       "Two successive sentences begin with the same adverb.",
		"desc_repetition_beginning_thesaurus": "Consider rewording the sentence or use a thesaurus to find a synonym.",
		// Java Categories.REPETITIONS_STYLE uses typo key cateogry_repetitions_style
		"cateogry_repetitions_style": "Repetitions (Style)",
		"category_repetitions":       "Repetitions",
	}
}

func utf16LenRO(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// atrWithPOSTags builds AnalyzedTokenReadings with the given POS tags (nil entry → null POSTag).
// Order of tags matches Java reading order (control flow is order-sensitive for ambiguous case).
func atrWithPOSTags(surface string, tags []*string) *languagetool.AnalyzedTokenReadings {
	if len(tags) == 0 {
		return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(surface, nil, nil))
	}
	toks := make([]*languagetool.AnalyzedToken, 0, len(tags))
	for _, tag := range tags {
		toks = append(toks, languagetool.NewAnalyzedToken(surface, tag, nil))
	}
	return languagetool.NewAnalyzedTokenReadingsList(toks, 0)
}

func strPtr(s string) *string { return &s }

// analyzeROWRBText ports JLanguageTool(Romanian) sentence tokenize + getAnalyzedSentence
// with DefaultRomanianTagger (official romanian.dict).
func analyzeROWRBText(t *testing.T, text string) []*languagetool.AnalyzedSentence {
	t.Helper()
	rotag.EnsureDefaultRomanianTagger()
	require.NotEmpty(t, rotag.RomanianPOSDictPath(), "official romanian.dict must resolve under inspiration/")
	lt := languagetool.NewJLanguageTool("ro")
	lt.SentenceTagger = rotag.DefaultRomanianTagger.Tag
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents, "Analyze(%q)", text)
	return sents
}

// handBuiltWRBSentence builds tokensWithoutWhitespace layout for WordRepeatBeginningRule:
// SENT_START + first content + rest...; requires len > 3 and text ending in [.?!].
func handBuiltWRBSentence(first *languagetool.AnalyzedTokenReadings, rest ...string) *languagetool.AnalyzedSentence {
	ss := languagetool.SentenceStartTagName
	start := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &ss, nil))
	pos := 0
	firstAt := languagetool.NewAnalyzedTokenReadingsList(first.GetReadings(), pos)
	toks := []*languagetool.AnalyzedTokenReadings{start, firstAt}
	pos = utf16LenRO(firstAt.GetToken())
	for _, s := range rest {
		toks = append(toks, languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken(s, nil, nil), pos))
		pos += utf16LenRO(s)
	}
	return languagetool.NewAnalyzedSentence(toks)
}

// TestRomanianWordRepeatBeginningRule_Metadata covers Java ctor id + category inheritance.
// Java RomanianWordRepeatBeginningRule has no addExamplePair.
func TestRomanianWordRepeatBeginningRule_Metadata(t *testing.T) {
	rule := NewRomanianWordRepeatBeginningRule(roWRBMessages())
	require.Equal(t, "ROMANIAN_WORD_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, 2, rule.MinToCheckParagraph())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), rule.GetCategory().GetID())
	require.Equal(t, roWRBMessages()["desc_repetition_beginning"], rule.GetDescription())
	// no example pairs in Java class
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
	// Java allowAmbiguousAdverbs() default false
	require.False(t, rule.AllowAmbiguousAdverbs())
}

// TestRomanianWordRepeatBeginningRule_IsAdverb ports pure-POS isAdverb matrix
// (Java loop over readings; G prefix; allowAmbiguousAdverbs false early-return).
func TestRomanianWordRepeatBeginningRule_IsAdverb(t *testing.T) {
	r := NewRomanianWordRepeatBeginningRule(roWRBMessages())

	// pure G-tag → true (Java startsWith "G")
	require.True(t, r.isAdverb(atrWithPOSTags("Apoi", []*string{strPtr("G00000p00t")})))
	require.True(t, r.isAdverb(atrWithPOSTags("Astfel", []*string{strPtr("G00000p00m")})))
	require.True(t, r.isAdverb(atrWithPOSTags("x", []*string{strPtr("G")})))
	require.True(t, r.isAdverb(atrWithPOSTags("x", []*string{strPtr("Gxyz")})))

	// multiple pure G readings → true
	require.True(t, r.isAdverb(atrWithPOSTags("x", []*string{strPtr("G0"), strPtr("G1")})))

	// null POS tags skipped; only G remains → true
	require.True(t, r.isAdverb(atrWithPOSTags("x", []*string{nil, strPtr("G00000p00t"), nil})))

	// all null POS → false (isAdverb stays false)
	require.False(t, r.isAdverb(atrWithPOSTags("x", []*string{nil, nil})))
	// no readings / empty surface with nil tag
	require.False(t, r.isAdverb(atrWithPOSTags("x", nil)))

	// non-G only → false
	require.False(t, r.isAdverb(atrWithPOSTags("Eu", []*string{strPtr("N0s1a0p00p")})))
	require.False(t, r.isAdverb(atrWithPOSTags("x", []*string{strPtr("V0s1000iz0")})))
	require.False(t, r.isAdverb(atrWithPOSTags("x", []*string{strPtr("RG")}))) // "RG" starts with R, not G

	// G + other tag, allowAmbiguous false (default) → false
	// Și-style: C, N, G (order as romanian.dict for "Și")
	require.False(t, r.isAdverb(atrWithPOSTags("Și", []*string{
		strPtr("C00000c00c"), strPtr("N0s3gar00n"), strPtr("N0p3gar00n"), strPtr("G00000v000"),
	})))
	// G first then non-G: sets true then returns false on non-G
	require.False(t, r.isAdverb(atrWithPOSTags("Bine", []*string{strPtr("G00000p00m"), strPtr("Sns3anc000")})))
	// non-G first then G: returns false before seeing G
	require.False(t, r.isAdverb(atrWithPOSTags("Rapid", []*string{
		strPtr("Sns3anc000"), strPtr("Ams3an0000"), strPtr("G00000p00m"),
	})))

	// allowAmbiguousAdverbs true: G + other tags → true (Java overridable branch)
	rAmb := NewRomanianWordRepeatBeginningRule(roWRBMessages())
	rAmb.allowAmbiguousAdverbs = true
	require.True(t, rAmb.AllowAmbiguousAdverbs())
	require.True(t, rAmb.isAdverb(atrWithPOSTags("Și", []*string{
		strPtr("C00000c00c"), strPtr("G00000v000"),
	})))
	// still false if no G at all
	require.False(t, rAmb.isAdverb(atrWithPOSTags("Eu", []*string{strPtr("N0s1a0p00p")})))

	require.False(t, r.isAdverb(nil))
}

// TestRomanianWordRepeatBeginningRule_IsAdverb_RealDict exercises isAdverb on
// DefaultRomanianTagger readings (same contract as Match path).
func TestRomanianWordRepeatBeginningRule_IsAdverb_RealDict(t *testing.T) {
	rotag.EnsureDefaultRomanianTagger()
	require.NotEmpty(t, rotag.RomanianPOSDictPath())
	r := NewRomanianWordRepeatBeginningRule(roWRBMessages())

	tagFirst := func(w string) *languagetool.AnalyzedTokenReadings {
		atrs := rotag.DefaultRomanianTagger.Tag([]string{w})
		require.NotEmpty(t, atrs, "tag %q", w)
		return atrs[0]
	}

	// pure G adverb surfaces from dict
	require.True(t, r.isAdverb(tagFirst("Apoi")))
	require.True(t, r.isAdverb(tagFirst("Astfel")))
	require.True(t, r.isAdverb(tagFirst("Totuși")))
	require.True(t, r.isAdverb(tagFirst("Aici")))

	// non-adverb
	require.False(t, r.isAdverb(tagFirst("Eu")))

	// ambiguous (G + other): allowAmbiguous false → not adverb
	require.False(t, r.isAdverb(tagFirst("Și")), "Și is ambiguous G+C+N")
	require.False(t, r.isAdverb(tagFirst("Rapid")), "Rapid is ambiguous G+S+A")
	require.False(t, r.isAdverb(tagFirst("Bine")), "Bine is ambiguous G+S")
}

// TestRomanianWordRepeatBeginningRule_Rule — word/adverb Match outcomes via real RO analysis.
func TestRomanianWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewRomanianWordRepeatBeginningRule(roWRBMessages())

	// ================== correct sentences ===================
	// two successive non-adverb same starts — no error (needs three for word arm)
	require.Equal(t, 0, len(rule.MatchList(analyzeROWRBText(t, "Eu merg. Eu văd."))))
	// two successive ambiguous Și (not isAdverb) — no error
	require.Equal(t, 0, len(rule.MatchList(analyzeROWRBText(t, "Și a venit. Și a plecat."))))
	// two successive Rapid (G + other tags) — not isAdverb under allowAmbiguous=false
	require.Equal(t, 0, len(rule.MatchList(analyzeROWRBText(t, "Rapid am plecat. Rapid am venit."))))
	// two successive Bine ambiguous
	require.Equal(t, 0, len(rule.MatchList(analyzeROWRBText(t, "Bine am făcut. Bine am zis."))))
	// list exception ":" at tokens[2]
	require.Equal(t, 0, len(rule.MatchList(analyzeROWRBText(t, "2011: un. 2011: doi. 2011: trei."))))
	// single sentence
	require.Equal(t, 0, len(rule.MatchList(analyzeROWRBText(t, "Apoi am plecat."))))
	// different starts
	require.Equal(t, 0, len(rule.MatchList(analyzeROWRBText(t, "Apoi am plecat. Astfel am venit."))))

	// =================== errors =============================
	// three successive "Eu" (non-adverb) → word short message
	matchesWord := rule.MatchList(analyzeROWRBText(t, "Eu merg. Eu văd. Eu vorbesc."))
	require.Equal(t, 1, len(matchesWord))
	require.Same(t, rule.WordRepeatBeginningRule, matchesWord[0].GetRule())
	require.Equal(t, "ROMANIAN_WORD_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, roWRBMessages()["desc_repetition_beginning_word"], matchesWord[0].GetShortMessage())
	require.Contains(t, matchesWord[0].GetMessage(), roWRBMessages()["desc_repetition_beginning_thesaurus"])
	require.Empty(t, matchesWord[0].GetSuggestedReplacements())
	require.Equal(t, utf16LenRO("Eu"), matchesWord[0].GetToPos()-matchesWord[0].GetFromPos())

	// three successive ambiguous Și → word path (not adverb exemption)
	matchesSi := rule.MatchList(analyzeROWRBText(t, "Și a venit. Și a plecat. Și a zis."))
	require.Equal(t, 1, len(matchesSi))
	require.Equal(t, roWRBMessages()["desc_repetition_beginning_word"], matchesSi[0].GetShortMessage())

	// two successive pure-G adverbs → adverb short message
	for _, pair := range []struct {
		text, surface string
	}{
		{"Apoi am plecat. Apoi am venit.", "Apoi"},
		{"Astfel am plecat. Astfel am venit.", "Astfel"},
		{"Totuși am plecat. Totuși am venit.", "Totuși"},
		{"Aici stau. Aici lucrez.", "Aici"},
	} {
		ms := rule.MatchList(analyzeROWRBText(t, pair.text))
		require.Equal(t, 1, len(ms), "adverb pair %q", pair.text)
		require.Equal(t, roWRBMessages()["desc_repetition_beginning_adv"], ms[0].GetShortMessage(), pair.surface)
		require.Contains(t, ms[0].GetMessage(), roWRBMessages()["desc_repetition_beginning_thesaurus"])
		require.Empty(t, ms[0].GetSuggestedReplacements())
		require.Equal(t, utf16LenRO(pair.surface), ms[0].GetToPos()-ms[0].GetFromPos(), pair.surface)
	}
}

// TestRomanianWordRepeatBeginningRule_Match_HandBuiltIsAdverbContract exercises
// Match shortMsg arms with hand-built POS-tagged ATR (Java isAdverb contract).
func TestRomanianWordRepeatBeginningRule_Match_HandBuiltIsAdverbContract(t *testing.T) {
	rule := NewRomanianWordRepeatBeginningRule(roWRBMessages())

	// two successive pure-G "Apoi" → adverb match
	gTag := "G00000p00t"
	s1 := handBuiltWRBSentence(atrWithPOSTags("Apoi", []*string{&gTag}), "am", "plecat", ".")
	s2 := handBuiltWRBSentence(atrWithPOSTags("Apoi", []*string{&gTag}), "am", "venit", ".")
	require.Regexp(t, `.+[.?!]$`, s1.GetText())

	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(ms), "hand-built pure G adverb pair; s1.Text=%q s2.Text=%q", s1.GetText(), s2.GetText())
	require.Equal(t, roWRBMessages()["desc_repetition_beginning_adv"], ms[0].GetShortMessage())

	// two successive ambiguous Și → no match (not adverb; only 2 sents)
	siTags := []*string{strPtr("C00000c00c"), strPtr("G00000v000")}
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{
		handBuiltWRBSentence(atrWithPOSTags("Și", siTags), "a", "venit", "."),
		handBuiltWRBSentence(atrWithPOSTags("Și", siTags), "a", "plecat", "."),
	})
	require.Equal(t, 0, len(ms), "ambiguous Și pair must not fire adverb arm")

	// three successive non-G "Eu" → word arm
	nTag := "N0s1a0p00p"
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{
		handBuiltWRBSentence(atrWithPOSTags("Eu", []*string{&nTag}), "merg", "eu", "."),
		handBuiltWRBSentence(atrWithPOSTags("Eu", []*string{&nTag}), "văd", "eu", "."),
		handBuiltWRBSentence(atrWithPOSTags("Eu", []*string{&nTag}), "vorbesc", "eu", "."),
	})
	require.Equal(t, 1, len(ms))
	require.Equal(t, roWRBMessages()["desc_repetition_beginning_word"], ms[0].GetShortMessage())
}
