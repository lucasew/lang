package de

// Twin of org.languagetool.rules.de.StyleRepeatedVeryShortSentences
// (no Java *Test class — behavior matrix via AnalyzePlain length + paragraph/quote rules).
//
// Java-critical:
//   MIN_REPEATED=3, MIN_WORDS=4, EXCLUDE_DIRECT_SPEECH=true; UserConfig triple;
//   tokens.length > 3 && <= minWords+2; span tokens[len-2]..tokens[len-1];
//   Tools.isParagraphEnd + nPara; OPENING/ENDING quotes; cumulative pos.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestStyleRepeatedVeryShortSentences_Meta(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	require.Equal(t, "STYLE_REPEATED_SHORT_SENTENCES", rule.GetID())
	require.Equal(t, "Stakkato-Sätze", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 4, rule.MinWords)
	require.Equal(t, 3, rule.MinRepeated)
	require.True(t, rule.ExcludeDirectSpeech)
	require.False(t, rule.SingleLineBreaksMarksPara)
	require.Equal(t, 3, rule.MinToCheckParagraph())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.CategoryInternal, rule.GetCategory().GetLocation())
	require.True(t, rule.GetCategory().IsDefaultOff())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	// Java addExamplePair multi-marker wrong / fixed without markers
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Contains(t, inc[0].GetExample(), "<marker>näher.</marker>")
	require.Contains(t, inc[0].GetExample(), "<marker>schlief.</marker>")
	require.Contains(t, inc[0].GetExample(), "<marker>quietschten.</marker>")
	cor := rule.GetCorrectExamples()
	require.Len(t, cor, 1)
	require.Equal(t,
		"Das Auto kam näher. Tief und fest schlief der Hund. Die Reifen quietschten.",
		cor[0].GetExample())
	require.NotContains(t, cor[0].GetExample(), "<marker>")

	// messages map → category name from category_creative_writing
	msgs := map[string]string{"category_creative_writing": "Kreatives Schreiben"}
	r2 := NewStyleRepeatedVeryShortSentences(msgs)
	require.Equal(t, "Kreatives Schreiben", r2.GetCategory().GetName())
}

func TestStyleRepeatedVeryShortSentences_GetRuleOptions(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 3)

	require.Equal(t, 3, opts[0].DefaultValue)
	require.Equal(t, "Anzahl von kurzen Sätzen hintereinander", opts[0].ConfigureText)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 10, opts[0].MaxConfigurableValue)

	require.Equal(t, 4, opts[1].DefaultValue)
	require.Equal(t, "Maximale Länge eines kurzen Satzes", opts[1].ConfigureText)
	require.Equal(t, 0, opts[1].MinConfigurableValue)
	require.Equal(t, 10, opts[1].MaxConfigurableValue)

	// Java 2-arg RuleOption(Boolean, text) → min/max default 0/100 (non-numeric defaultValue)
	require.Equal(t, true, opts[2].DefaultValue)
	require.Equal(t, "Direkte Rede und Zitate ausschließen", opts[2].ConfigureText)
	require.Equal(t, 0, opts[2].MinConfigurableValue)
	require.Equal(t, 100, opts[2].MaxConfigurableValue)
}

func TestStyleRepeatedVeryShortSentences_DefaultShortx3(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// Java example: three short sentences
	// tokens.length > 3 && <= minWords+2 (SENT_START + words + punct)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund schlief."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	for _, s := range sents {
		n := len(s.GetTokensWithoutWhitespace())
		require.Greater(t, n, 3, s.GetText())
		require.LessOrEqual(t, n, rule.MinWords+2, s.GetText())
	}
	ms := rule.MatchList(sents)
	require.Equal(t, 3, len(ms))
	for _, m := range ms {
		require.Equal(t, "Stakkato-Sätze", m.GetMessage())
		require.Greater(t, m.GetToPos(), m.GetFromPos())
	}
	// Java: start = tokens[len-2].startPos (+pos), end = tokens[len-1].endPos (+pos)
	// first sentence has pos=0
	toks0 := sents[0].GetTokensWithoutWhitespace()
	require.Equal(t, toks0[len(toks0)-2].GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, toks0[len(toks0)-1].GetEndPos(), ms[0].GetToPos())
}

func TestStyleRepeatedVeryShortSentences_LongBreaksStreak(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// long sentence breaks the streak; only 2 shorts before + 1 after → no emit
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
		languagetool.AnalyzePlain("Dieser deutlich längere Satz unterbricht die Serie von kurzen Sätzen hier."),
		languagetool.AnalyzePlain("Er ging."),
	}
	require.Equal(t, 0, len(rule.MatchList(sents)))
}

func TestStyleRepeatedVeryShortSentences_BelowMinRepeated(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// fewer than minRepeated → early empty (and also streak too short)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
	})))
}

func TestStyleRepeatedVeryShortSentences_UserConfigOverrides(t *testing.T) {
	// cf[0]=minRepeated, cf[1]=minWords, cf[2]=excludeDirectSpeech
	uc := languagetool.NewUserConfigWithWords(nil, map[string][]any{
		"STYLE_REPEATED_SHORT_SENTENCES": {2, 10, false},
	})
	rule := NewStyleRepeatedVeryShortSentencesWithUserConfig(nil, uc)
	require.Equal(t, 2, rule.MinRepeated)
	require.Equal(t, 10, rule.MinWords)
	require.False(t, rule.ExcludeDirectSpeech)
	require.Equal(t, 2, rule.MinToCheckParagraph())

	// minRepeated=2 → two shorts match
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
	}
	require.Equal(t, 2, len(rule.MatchList(sents)))

	// wrong types ignored (Java instanceof Integer/Boolean)
	ucBad := languagetool.NewUserConfigWithWords(nil, map[string][]any{
		"STYLE_REPEATED_SHORT_SENTENCES": {"3", 4.0, "true"},
	})
	ruleBad := NewStyleRepeatedVeryShortSentencesWithUserConfig(nil, ucBad)
	require.Equal(t, 3, ruleBad.MinRepeated) // defaults preserved
	require.Equal(t, 4, ruleBad.MinWords)
	require.True(t, ruleBad.ExcludeDirectSpeech)

	// partial cf: only minRepeated
	ucPart := languagetool.NewUserConfigWithWords(nil, map[string][]any{
		"STYLE_REPEATED_SHORT_SENTENCES": {5},
	})
	rulePart := NewStyleRepeatedVeryShortSentencesWithUserConfig(nil, ucPart)
	require.Equal(t, 5, rulePart.MinRepeated)
	require.Equal(t, 4, rulePart.MinWords)
	require.True(t, rulePart.ExcludeDirectSpeech)

	// nil userConfig keeps defaults
	ruleNil := NewStyleRepeatedVeryShortSentencesWithUserConfig(nil, nil)
	require.Equal(t, 3, ruleNil.MinRepeated)
	require.Equal(t, 4, ruleNil.MinWords)
	require.True(t, ruleNil.ExcludeDirectSpeech)
}

func TestStyleRepeatedVeryShortSentences_DirectSpeechExclude(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// long quoted (tokens > minWords+2) also excluded by length, but exclude path fires first
	quotedLong := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("„Das Auto kam näher.“"),
		languagetool.AnalyzePlain("„Der Hund schlief.“"),
		languagetool.AnalyzePlain("„Die Reifen quietschten.“"),
	}
	require.Equal(t, 0, len(rule.MatchList(quotedLong)), "quoted short sentences excluded")

	// short quoted: nTok=6 <= minWords+2; OPENING quote + !wsBefore next → beginsWithDirectSpeech
	quotedShort := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("„Er lief.“"),
		languagetool.AnalyzePlain("„Sie rief.“"),
		languagetool.AnalyzePlain("„Er ging.“"),
	}
	for _, s := range quotedShort {
		n := len(s.GetTokensWithoutWhitespace())
		require.Greater(t, n, 3)
		require.LessOrEqual(t, n, rule.MinWords+2)
	}
	require.Equal(t, 0, len(rule.MatchList(quotedShort)), "short quoted excluded by direct speech")

	// excludeDirectSpeech=false → quoted shorts match
	uc := languagetool.NewUserConfigWithWords(nil, map[string][]any{
		"STYLE_REPEATED_SHORT_SENTENCES": {3, 4, false},
	})
	ruleIncl := NewStyleRepeatedVeryShortSentencesWithUserConfig(nil, uc)
	require.False(t, ruleIncl.ExcludeDirectSpeech)
	ms := ruleIncl.MatchList(quotedShort)
	require.Equal(t, 3, len(ms), "quoted shorts match when excludeDirectSpeech=false")
}

// TestStyleRepeatedVeryShortSentences_QuoteFullTokenMatch asserts Java
// OPENING_QUOTES/ENDING_QUOTES.matcher(token).matches() — full token only.
// Multi-char surfaces that merely contain a quote (e.g. „foo, "bar) must not
// open/close direct speech the way unanchored MatchString would.
func TestStyleRepeatedVeryShortSentences_QuoteFullTokenMatch(t *testing.T) {
	// Pattern-level: single quote char matches; multi-char containing quote does not.
	require.True(t, styleShortOpenQuotes.MatchString(`„`))
	require.True(t, styleShortOpenQuotes.MatchString(`"`))
	require.False(t, styleShortOpenQuotes.MatchString(`„foo`), "substring open must not match")
	require.False(t, styleShortOpenQuotes.MatchString(`"bar`), "substring open must not match")
	require.True(t, styleShortEndingQuotes.MatchString(`“`))
	require.True(t, styleShortEndingQuotes.MatchString(`"`))
	require.False(t, styleShortEndingQuotes.MatchString(`"bar`), "substring end must not match")
	require.False(t, styleShortEndingQuotes.MatchString(`foo“`), "substring end must not match")

	rule := NewStyleRepeatedVeryShortSentences(nil)
	// Hand-built: token surface contains open quote but is not a single-char quote token.
	// Next token has !WhitespaceBefore (zero-value), so unanchored open would set
	// beginsWithDirectSpeech and exclude; Java full-token does not.
	// tokens.length=4 → >3 && <= minWords+2 → eligible short.
	multiOpenShort := func(openContaining, word string) *languagetool.AnalyzedSentence {
		next := atrWithPOS(word, "SUB:NOM:SIN:NEU", word)
		// explicit: no whitespace before next (would satisfy Java open-quote side condition)
		next.SetWhitespaceBefore(false)
		return languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS(openContaining, "UNKNOWN", openContaining),
			next,
			atrWithPOS(".", "PKT", "."),
		))
	}
	sents := []*languagetool.AnalyzedSentence{
		multiOpenShort(`„foo`, "lief"),
		multiOpenShort(`"bar`, "rief"),
		multiOpenShort(`„baz`, "ging"),
	}
	for _, s := range sents {
		n := len(s.GetTokensWithoutWhitespace())
		require.Greater(t, n, 3)
		require.LessOrEqual(t, n, rule.MinWords+2)
		// sanity: multi-char first content token contains a quote char
		require.Greater(t, len([]rune(s.GetTokensWithoutWhitespace()[1].GetToken())), 1)
	}
	ms := rule.MatchList(sents)
	require.Equal(t, 3, len(ms),
		"multi-char tokens containing quotes must NOT start direct speech (Java Matcher.matches)")
}

func TestStyleRepeatedVeryShortSentences_ParagraphEndInteraction(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// Java: include when (!paragraphEnd || nPara > 0)
	// Solo short per paragraph (each ends \n\n; last is always para end):
	// every sentence has nPara==0 && paragraphEnd → never included.
	soloParas := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief.\n\n"),
		languagetool.AnalyzePlain("Sie rief.\n\n"),
		languagetool.AnalyzePlain("Er ging."),
	}
	require.Equal(t, 0, len(rule.MatchList(soloParas)), "solo short per paragraph not staccato")

	// Three shorts in one paragraph (no double newline) → match
	samePara := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
		languagetool.AnalyzePlain("Er ging."),
	}
	require.Equal(t, 3, len(rule.MatchList(samePara)))

	// First alone as paragraph end (nPara=0) skipped; then three consecutive shorts → match
	// s0 ends \n\n → para end, nPara=0 → skip/flush; nPara=-1
	// s1,s2,s3 consecutive → 3 matches
	mixed := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er kam.\n\n"),
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
		languagetool.AnalyzePlain("Er ging."),
	}
	require.Equal(t, 3, len(rule.MatchList(mixed)))
}

func TestStyleRepeatedVeryShortSentences_CumulativePos(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund schlief."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	ms := rule.MatchList(sents)
	require.Equal(t, 3, len(ms))

	// Java: pos accumulates getCorrectedTextLength(); start/end = token local + pos
	pos := 0
	for i, s := range sents {
		toks := s.GetTokensWithoutWhitespace()
		require.Equal(t, toks[len(toks)-2].GetStartPos()+pos, ms[i].GetFromPos(), "match %d FromPos", i)
		require.Equal(t, toks[len(toks)-1].GetEndPos()+pos, ms[i].GetToPos(), "match %d ToPos", i)
		pos += s.GetCorrectedTextLength()
	}
	// absolute offsets strictly increase across sentences
	require.Greater(t, ms[1].GetFromPos(), ms[0].GetToPos())
	require.Greater(t, ms[2].GetFromPos(), ms[1].GetToPos())
}

func TestStyleRepeatedVeryShortSentences_FlushOnBreakThenNewStreak(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// 3 short, long break, 3 short → 6 matches
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
		languagetool.AnalyzePlain("Er ging."),
		languagetool.AnalyzePlain("Dieser deutlich längere Satz unterbricht die Serie von kurzen Sätzen hier."),
		languagetool.AnalyzePlain("Er kam."),
		languagetool.AnalyzePlain("Sie sah."),
		languagetool.AnalyzePlain("Es war."),
	}
	ms := rule.MatchList(sents)
	require.Equal(t, 6, len(ms))
}

func TestStyleRepeatedVeryShortSentences_OneWordSkipped(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// Java: tokens.length > 3 skips one-word sentences (SENT_START + word + punct = 3)
	// "Ja." → SENT_START, Ja, . → length 3 → not short-streak candidate
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Ja."),
		languagetool.AnalyzePlain("Nein."),
		languagetool.AnalyzePlain("Doch."),
		languagetool.AnalyzePlain("Er lief."),
	}
	// only one real short at end → no match
	require.Equal(t, 0, len(rule.MatchList(sents)))
	for _, s := range sents[:3] {
		require.LessOrEqual(t, len(s.GetTokensWithoutWhitespace()), 3, s.GetText())
	}
}

// TestStyleRepeatedVeryShortSentences keeps the historical smoke entrypoint name used by package scans.
func TestStyleRepeatedVeryShortSentences(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	require.Equal(t, "STYLE_REPEATED_SHORT_SENTENCES", rule.GetID())
	require.Equal(t, "Stakkato-Sätze", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 4, rule.MinWords)
	require.Equal(t, 3, rule.MinRepeated)
	require.True(t, rule.ExcludeDirectSpeech)
	require.Equal(t, 3, rule.MinToCheckParagraph())
	require.NotEmpty(t, rule.GetIncorrectExamples())

	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund schlief."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	ms := rule.MatchList(sents)
	require.Equal(t, 3, len(ms))

	// long break
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
		languagetool.AnalyzePlain("Dieser deutlich längere Satz unterbricht die Serie von kurzen Sätzen hier."),
		languagetool.AnalyzePlain("Er ging."),
	})))

	// below min
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
	})))

	// direct speech exclusion
	quoted := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("„Das Auto kam näher.“"),
		languagetool.AnalyzePlain("„Der Hund schlief.“"),
		languagetool.AnalyzePlain("„Die Reifen quietschten.“"),
	}
	require.Equal(t, 0, len(rule.MatchList(quoted)), "quoted short sentences excluded")
}
