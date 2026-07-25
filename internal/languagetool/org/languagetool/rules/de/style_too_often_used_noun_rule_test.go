package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Behavior matrix twin for org.languagetool.rules.de.StyleTooOftenUsedNounRule
// (+ AbstractStyleTooOftenUsedWordRule match algorithm). Java has no dedicated *Test class.

func TestStyleTooOftenUsedNounRule_JavaMeta(t *testing.T) {
	// Java StyleTooOftenUsedNounRule + AbstractStyleTooOftenUsedWordRule ctor.
	rule := NewStyleTooOftenUsedNounRule(nil)
	require.Equal(t, "TOO_OFTEN_USED_NOUN_DE", rule.GetID())
	require.Equal(t, "Statistische Stilanalyse: Zu häufig genutztes Substantiv", rule.GetDescription())
	// Java getConfigureText
	require.Equal(t, styleTooOftenNounConfigureText, rule.GetConfigureText())
	require.Equal(t, 5, rule.MinPercent)        // DEFAULT_MIN_PERCENT
	require.Equal(t, 5, rule.DefaultMinPercent) // Java defaultMinPercent
	require.Equal(t, 100, rule.MinWordCount)    // Abstract MIN_WORD_COUNT
	require.True(t, rule.IsDefaultOff())        // DEFAULT_ACTIVATION = false
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	cat := rule.GetCategory()
	require.NotNil(t, cat)
	require.Equal(t, "CREATIVE_WRITING", cat.GetID().String())
	require.True(t, cat.IsDefaultOff())

	// getLimitMessage(limit) — Java has NO space before period (unlike adjective rule).
	require.Equal(t,
		"Das Substantiv wird häufiger verwendet als 5% aller Substantive. Möglicherweise ist es besser es durch ein Synonym zu ersetzen.",
		rule.LimitMessage(5))
	require.Equal(t,
		"Das Substantiv wird häufiger verwendet als 12% aller Substantive. Möglicherweise ist es besser es durch ein Synonym zu ersetzen.",
		rule.LimitMessage(12))

	// getRuleOptions: RuleOption(defaultMinPercent, getConfigureText(), 1, 100)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 5, opts[0].DefaultValue)
	require.Equal(t, rule.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 1, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestStyleTooOftenUsedNounRule_UserConfigMinPercentAndThesaurus(t *testing.T) {
	// Java abstract getMinPercent(userConfig, DEFAULT): configValueByID(getId())[0]
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("TOO_OFTEN_USED_NOUN_DE", []any{17})
	ls := languagetool.NewLinguServices()
	uc.LinguServices = ls

	rule := NewStyleTooOftenUsedNounRuleWithUserConfig(nil, uc)
	require.Equal(t, 17, rule.MinPercent)
	require.Equal(t, 5, rule.DefaultMinPercent) // RuleOption default stays DEFAULT_MIN_PERCENT
	// Java setThesaurusRelevantRule(this)
	require.Equal(t, rule, ls.ThesaurusRelevantRule)

	// nil userConfig: defaults, no thesaurus registration
	ruleNil := NewStyleTooOftenUsedNounRuleWithUserConfig(nil, nil)
	require.Equal(t, 5, ruleNil.MinPercent)

	// userConfig without LinguServices: no panic, default min percent
	uc2 := languagetool.NewUserConfig()
	rule2 := NewStyleTooOftenUsedNounRuleWithUserConfig(nil, uc2)
	require.Equal(t, 5, rule2.MinPercent)
}

func TestStyleTooOftenUsedNounRule_IsToCountedWordAndException(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule(nil)

	// isToCountedWord = hasPosTagStartingWith("SUB:")
	require.True(t, rule.IsToCountedWord(atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus")))
	require.True(t, rule.IsToCountedWord(atrWithPOS("Häuser", "SUB:NOM:PLU:NEU", "Haus")))
	require.False(t, rule.IsToCountedWord(atrWithPOS("schön", "ADJ:PRD:GRU", "schön")))
	require.False(t, rule.IsToCountedWord(atrWithPOS("laufen", "VER:INF:NON", "laufen")))
	// SUB without colon is not SUB: prefix
	require.False(t, rule.IsToCountedWord(atrWithPOS("x", "SUB", "x")))

	// isException: PRO: | Ich | Aber | Ja | Frau/Herr + (EIG: | isPosTagUnknown)
	tokens := []*languagetool.AnalyzedTokenReadings{
		sentStartATR(),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"), // 1 PRO:
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),    // 2 pure SUB:
		atrWithPOS("Ich", "SUB:NOM:SIN:NEU", "Ich"),      // 3 surface "Ich"
		atrWithPOS("Aber", "SUB:NOM:SIN:NEU", "Aber"),    // 4 surface "Aber"
		atrWithPOS("Ja", "SUB:NOM:SIN:NEU", "Ja"),        // 5 surface "Ja"
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),    // 6 Frau + EIG next
		atrWithPOS("Müller", "EIG:NOM:SIN:MAS", "Müller"),
		atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"), // 8 Herr + unknown next
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Xyxyz", nil, nil)),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"), // 10 Frau + tagged non-EIG next → not exception
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"), // 12 last token, no next → not exception via Frau|Herr
	}
	require.True(t, rule.IsException(tokens, 1))   // PRO:
	require.False(t, rule.IsException(tokens, 2))  // pure SUB:
	require.True(t, rule.IsException(tokens, 3))   // "Ich"
	require.True(t, rule.IsException(tokens, 4))   // "Aber"
	require.True(t, rule.IsException(tokens, 5))   // "Ja"
	require.True(t, rule.IsException(tokens, 6))   // Frau + EIG:
	require.False(t, rule.IsException(tokens, 7))  // EIG alone not exception for this rule
	require.True(t, rule.IsException(tokens, 8))   // Herr + isPosTagUnknown
	require.False(t, rule.IsException(tokens, 9))  // unknown next alone
	require.False(t, rule.IsException(tokens, 10)) // Frau + SUB: (not EIG, not unknown)
	require.False(t, rule.IsException(tokens, 11)) // Auto pure
	require.False(t, rule.IsException(tokens, 12)) // Herr at end, no next

	// Multi-tag: SUB: + PRO: → counted word AND exception
	subAndPro := atrWithPOSMulti("mein", "mein", "SUB:NOM:SIN:NEU", "PRO:POS:NOM:SIN:MAS")
	require.True(t, rule.IsToCountedWord(subAndPro))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{sentStartATR(), subAndPro}, 1))

	// Critical: isPosTagUnknown vs !IsTagged divergence
	// Token with SENT_END (or any real POS) is !IsTagged only when hasNoTag paths apply.
	// A reading with POS "UNKNOWN" / SENT_END is tagged in isTagged sense but NOT isPosTagUnknown.
	// Java: isPosTagUnknown only when single reading with null POS at construction.
	sentEnd := "SENT_END"
	withSentEnd := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("X", &sentEnd, nil),
	}, 0)
	// has real POS → not isPosTagUnknown; isTagged depends on HasNoTag for SENT_END
	require.False(t, withSentEnd.IsPosTagUnknown())
	// Frau + next with non-null POS that is not EIG: must NOT exception via isPosTagUnknown
	// even if someone confuses with !IsTagged for special tags.
	frauToks := []*languagetool.AnalyzedTokenReadings{
		sentStartATR(),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		withSentEnd,
	}
	require.False(t, rule.IsException(frauToks, 1), "Frau + non-null POS non-EIG is not exception")

	// Explicit: next isPosTagUnknown → exception for Frau
	unkNext := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Zzz", nil, nil))
	require.True(t, unkNext.IsPosTagUnknown())
	require.False(t, unkNext.IsTagged())
	frauUnk := []*languagetool.AnalyzedTokenReadings{
		sentStartATR(),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		unkNext,
	}
	require.True(t, rule.IsException(frauUnk, 1))
}

func TestStyleTooOftenUsedNounRule_ToAddedLemma(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule(nil)
	// Java: getLemmaForPosTagStartsWith("SUB:", token)
	require.Equal(t, "Haus", rule.ToAddedLemma(atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus")))
	require.Equal(t, "Haus", rule.ToAddedLemma(atrWithPOS("Häuser", "SUB:NOM:PLU:NEU", "Haus")))
	// No SUB: reading → null/empty (no surface invent)
	require.Equal(t, "", rule.ToAddedLemma(atrWithPOS("schön", "ADJ:PRD:GRU", "schön")))
	// SUB: reading with nil lemma → empty string path (Java returns null lemma)
	pos := "SUB:NOM:SIN:NEU"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", &pos, nil), 0)
	require.Equal(t, "", rule.ToAddedLemma(tok))
	// First SUB: reading wins when multi-tag
	adjTag, subTag := "ADJ:PRD:GRU", "SUB:NOM:SIN:NEU"
	lemmaAdj, lemmaSub := "alt", "Substantiv"
	readings := []*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("alt", &adjTag, &lemmaAdj),
		languagetool.NewAnalyzedToken("alt", &subTag, &lemmaSub),
	}
	multi := languagetool.NewAnalyzedTokenReadingsList(readings, 0)
	require.Equal(t, "Substantiv", rule.ToAddedLemma(multi))
}

func TestStyleTooOftenUsedNounRule_Match_MinWordCountGate(t *testing.T) {
	// Java getTooOftenUsedWords: if numWords < MIN_WORD_COUNT (100) → empty.
	rule := NewStyleTooOftenUsedNounRule(nil)
	require.Equal(t, 100, rule.MinWordCount)

	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("sah", "VER:3:SIN:PRT:NON", "sehen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("kaufte", "VER:3:SIN:PRT:SFT", "kaufen"),
		atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	// Short text: only 2 counted SUB lemmas → under 100 → no matches
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2}))
}

func TestStyleTooOftenUsedNounRule_Match_OverusedLemma(t *testing.T) {
	// Bypass MIN_WORD_COUNT only for percent-threshold unit test (Java production uses 100).
	rule := NewStyleTooOftenUsedNounRule(nil)
	rule.MinWordCount = 0

	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("sah", "VER:3:SIN:PRT:NON", "sehen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("kaufte", "VER:3:SIN:PRT:SFT", "kaufen"),
		atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	// "Haus" is 100% of nouns ≥ DEFAULT_MIN_PERCENT 5 → both hits
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.GreaterOrEqual(t, len(ms), 2)
	wantMsg := "Das Substantiv wird häufiger verwendet als 5% aller Substantive. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
	for _, m := range ms {
		require.Equal(t, wantMsg, m.GetMessage())
		r, ok := m.GetRule().(*rules.AbstractStyleTooOftenUsedWordRule)
		require.True(t, ok)
		require.Equal(t, "TOO_OFTEN_USED_NOUN_DE", r.GetID())
	}

	// Two lemmas each 50% ≥ 5 → both flag once
	s3 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("kaufte", "VER:3:SIN:PRT:SFT", "kaufen"),
		atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("die", "ART:DEF:AKK:SIN:FEM", "der"),
		atrWithPOS("Villa", "SUB:AKK:SIN:FEM", "Villa"),
		atrWithPOS(".", "PKT", "."),
	))
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})
	require.Equal(t, 2, len(ms2))

	// Raise threshold above 50%: neither flags
	rule.MinPercent = 51
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3}))
}

func TestStyleTooOftenUsedNounRule_Match_ExceptionTagsNotCounted(t *testing.T) {
	// PRO: and Ich/Aber/Ja and Frau|Herr+EIG/unknown must not enter the map.
	rule := NewStyleTooOftenUsedNounRule(nil)
	rule.MinWordCount = 0

	// Only pure SUB: "Haus"; "sie" has PRO: and isException even if it also had SUB:.
	// "Ich" surface exception; Frau+EIG excluded.
	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("sah", "VER:3:SIN:PRT:NON", "sehen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS("und", "KON", "und"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS("Müller", "EIG:NOM:SIN:MAS", "Müller"),
		atrWithPOS(".", "PKT", "."),
	))
	// only "Haus" counted → 100% ≥ 5 → one match on "Haus" only
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "Haus", extractSurfaceDE(s, ms[0]))

	// Ich/Aber/Ja surfaces with SUB: are exceptions
	sIch := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "SUB:NOM:SIN:NEU", "Ich"),
		atrWithPOS("sah", "VER:3:SIN:PRT:NON", "sehen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	msIch := rule.MatchList([]*languagetool.AnalyzedSentence{sIch})
	require.Equal(t, 1, len(msIch))
	require.Equal(t, "Haus", extractSurfaceDE(sIch, msIch[0]))
}

func TestStyleTooOftenUsedNounRule_Match_FrauHerrUnknownException(t *testing.T) {
	// Frau/Herr + isPosTagUnknown next must not be counted.
	rule := NewStyleTooOftenUsedNounRule(nil)
	rule.MinWordCount = 0

	unkName := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Xyxyz", nil, nil))
	require.True(t, unkName.IsPosTagUnknown())

	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"),
		unkName,
		atrWithPOS("kaufte", "VER:3:SIN:PRT:SFT", "kaufen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "Haus", extractSurfaceDE(s, ms[0]))

	// Frau + ordinary SUB next (not EIG, not unknown) → Frau is counted as noun
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(".", "PKT", "."),
	))
	// Frau + Auto each 50% ≥ 5 → two matches
	require.Equal(t, 2, len(rule.MatchList([]*languagetool.AnalyzedSentence{s2})))
}

func TestStyleTooOftenUsedNounRule_Match_DirectSpeechExclusion(t *testing.T) {
	// Java: withoutDirectSpeech only when setWithoutDirectSpeech(true).
	// Opening quote with next token !isWhitespaceBefore → enter DS; content not counted.
	rule := NewStyleTooOftenUsedNounRule(nil)
	rule.MinWordCount = 0
	rule.WithoutDirectSpeech = true

	// Sentence: Er sagte "Haus Haus Haus" und Villa da.
	// Inside quotes: three "Haus" SUB; outside: "Villa" SUB.
	// With exclusion: only "Villa" counted → 100% of nouns → match on "Villa" only.
	openQ := atrWithPOS("\"", "\"", "\"")
	h1 := atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus")
	h1.SetWhitespaceBefore(false)
	h2 := atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus")
	h2.SetWhitespaceBefore(true)
	h3 := atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus")
	h3.SetWhitespaceBefore(true)
	endQ := atrWithPOS("\"", "\"", "\"")
	endQ.SetWhitespaceBefore(false)
	villa := atrWithPOS("Villa", "SUB:NOM:SIN:FEM", "Villa")
	villa.SetWhitespaceBefore(true)

	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("sagte", "VER:3:SIN:PRT:SFT", "sagen"),
		openQ,
		h1,
		h2,
		h3,
		endQ,
		atrWithPOS("und", "KON", "und"),
		villa,
		atrWithPOS(".", "PKT", "."),
	))
	toks := s.GetTokensWithoutWhitespace()
	var quoteIdxs []int
	for i := 1; i < len(toks); i++ {
		if toks[i].GetToken() == "\"" {
			quoteIdxs = append(quoteIdxs, i)
		}
	}
	require.GreaterOrEqual(t, len(quoteIdxs), 2)
	toks[quoteIdxs[0]+1].SetWhitespaceBefore(false) // enter DS
	toks[quoteIdxs[1]].SetWhitespaceBefore(false)   // exit DS

	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms), "only outside-quote overused noun should match")
	require.Equal(t, "Villa", extractSurfaceDE(s, ms[0]))

	// Default WithoutDirectSpeech=false: quotes do not exclude; "Haus" is majority
	rule.WithoutDirectSpeech = false
	msAll := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.GreaterOrEqual(t, len(msAll), 3) // three "Haus" at least
}

func TestStyleTooOftenUsedNounRule_Match_UntaggedNoInvent(t *testing.T) {
	// Soft/untagged SplitAndAnalyze must not invent POS → no matches.
	rule := NewStyleTooOftenUsedNounRule(nil)
	rule.MinWordCount = 0
	plain := languagetool.SplitAndAnalyze("Er sah das Haus am See. Dann kaufte er das Haus in der Stadt.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
}

func TestStyleTooOftenUsedNounRule_WithDefaultLimit(t *testing.T) {
	r := NewStyleTooOftenUsedNounRuleWithDefaultLimit(nil)
	require.Equal(t, 5, r.MinPercent)
	require.Equal(t, 100, r.MinWordCount)
	require.Equal(t, "TOO_OFTEN_USED_NOUN_DE", r.GetID())
}
