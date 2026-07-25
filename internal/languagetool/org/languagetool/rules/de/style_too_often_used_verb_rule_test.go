package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Behavior matrix twin for org.languagetool.rules.de.StyleTooOftenUsedVerbRule
// (+ AbstractStyleTooOftenUsedWordRule match algorithm). Java has no dedicated *Test class.

func TestStyleTooOftenUsedVerbRule_JavaMeta(t *testing.T) {
	// Java StyleTooOftenUsedVerbRule + AbstractStyleTooOftenUsedWordRule ctor.
	rule := NewStyleTooOftenUsedVerbRule(nil)
	require.Equal(t, "TOO_OFTEN_USED_VERB_DE", rule.GetID())
	require.Equal(t, "Statistische Stilanalyse: Zu häufig genutztes Verb", rule.GetDescription())
	// Java getConfigureText
	require.Equal(t, styleTooOftenVerbConfigureText, rule.GetConfigureText())
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
		"Das Verb wird häufiger verwendet als 5% aller Verben. Möglicherweise ist es besser es durch ein Synonym zu ersetzen.",
		rule.LimitMessage(5))
	require.Equal(t,
		"Das Verb wird häufiger verwendet als 12% aller Verben. Möglicherweise ist es besser es durch ein Synonym zu ersetzen.",
		rule.LimitMessage(12))

	// getRuleOptions: RuleOption(defaultMinPercent, getConfigureText(), 1, 100)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 5, opts[0].DefaultValue)
	require.Equal(t, rule.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 1, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestStyleTooOftenUsedVerbRule_UserConfigMinPercentAndThesaurus(t *testing.T) {
	// Java abstract getMinPercent(userConfig, DEFAULT): configValueByID(getId())[0]
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("TOO_OFTEN_USED_VERB_DE", []any{17})
	ls := languagetool.NewLinguServices()
	uc.LinguServices = ls

	rule := NewStyleTooOftenUsedVerbRuleWithUserConfig(nil, uc)
	require.Equal(t, 17, rule.MinPercent)
	require.Equal(t, 5, rule.DefaultMinPercent) // RuleOption default stays DEFAULT_MIN_PERCENT
	// Java setThesaurusRelevantRule(this)
	require.Equal(t, rule, ls.ThesaurusRelevantRule)

	// nil userConfig: defaults, no thesaurus registration
	ruleNil := NewStyleTooOftenUsedVerbRuleWithUserConfig(nil, nil)
	require.Equal(t, 5, ruleNil.MinPercent)

	// userConfig without LinguServices: no panic, default min percent
	uc2 := languagetool.NewUserConfig()
	rule2 := NewStyleTooOftenUsedVerbRuleWithUserConfig(nil, uc2)
	require.Equal(t, 5, rule2.MinPercent)
}

func TestStyleTooOftenUsedVerbRule_IsToCountedWordAndException(t *testing.T) {
	rule := NewStyleTooOftenUsedVerbRule(nil)

	// isToCountedWord = hasPosTagStartingWith("VER:")
	require.True(t, rule.IsToCountedWord(atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen")))
	require.True(t, rule.IsToCountedWord(atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen")))
	require.True(t, rule.IsToCountedWord(atrWithPOS("kann", "VER:MOD:3:SIN:PRS", "können")))
	require.True(t, rule.IsToCountedWord(atrWithPOS("ist", "VER:AUX:3:SIN:PRS", "sein")))
	require.False(t, rule.IsToCountedWord(atrWithPOS("schnell", "ADV", "schnell")))
	require.False(t, rule.IsToCountedWord(atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto")))
	// VER without colon is not VER: prefix
	require.False(t, rule.IsToCountedWord(atrWithPOS("x", "VER", "x")))

	// isException = VER:MOD || VER:AUX || ART || ADJ (bare ART/ADJ — Java hasPosTagStartingWith).
	// Bare "ART"/"ADJ" match ART:… and ADJ:…; also bare ART/ADJ tags.
	tokens := []*languagetool.AnalyzedTokenReadings{
		sentStartATR(),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"), // 1 pure VER: — not exception
		atrWithPOS("kann", "VER:MOD:3:SIN:PRS", "können"),    // 2 VER:MOD
		atrWithPOS("ist", "VER:AUX:3:SIN:PRS", "sein"),       // 3 VER:AUX
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "der"),      // 4 ART:… matches "ART"
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),          // 5 ADJ:… matches "ADJ"
		atrWithPOS("der", "ART", "der"),                      // 6 bare ART matches "ART"
		atrWithPOS("alt", "ADJ", "alt"),                      // 7 bare ADJ matches "ADJ"
		// VER: that is not MOD/AUX and no ART/ADJ
		atrWithPOS("ging", "VER:1:SIN:PRT:NON", "gehen"), // 8
	}
	require.False(t, rule.IsException(tokens, 1)) // pure VER:
	require.True(t, rule.IsException(tokens, 2))  // VER:MOD
	require.True(t, rule.IsException(tokens, 3))  // VER:AUX
	require.True(t, rule.IsException(tokens, 4))  // ART:DEF… (prefix "ART")
	require.True(t, rule.IsException(tokens, 5))  // ADJ:… (prefix "ADJ")
	require.True(t, rule.IsException(tokens, 6))  // bare ART
	require.True(t, rule.IsException(tokens, 7))  // bare ADJ
	require.False(t, rule.IsException(tokens, 8)) // pure VER: not MOD/AUX

	// Multi-tag: VER: + VER:MOD → counted word AND exception
	verMod := atrWithPOSMulti("kann", "können", "VER:MOD:3:SIN:PRS", "VER:1:SIN:PRS:NON")
	require.True(t, rule.IsToCountedWord(verMod))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{sentStartATR(), verMod}, 1))

	// Multi-tag: VER: + ADJ: → exception via bare "ADJ" prefix (Java hasPosTagStartingWith("ADJ"))
	verAdj := atrWithPOSMulti("x", "x", "VER:INF:NON", "ADJ:PRD:GRU")
	require.True(t, rule.IsToCountedWord(verAdj))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{sentStartATR(), verAdj}, 1))

	// Multi-tag: VER: + ART: → exception via bare "ART"
	verArt := atrWithPOSMulti("x", "x", "VER:INF:NON", "ART:DEF:NOM:SIN:NEU")
	require.True(t, rule.IsToCountedWord(verArt))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{sentStartATR(), verArt}, 1))

	// Note: "VER:MODAL" would also match VER:MOD prefix; "VER:AUXILIARY" would match VER:AUX.
	// Java uses startsWith, not exact tag.
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{
		sentStartATR(),
		atrWithPOS("x", "VER:MODAL:FOO", "x"),
	}, 1))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{
		sentStartATR(),
		atrWithPOS("x", "VER:AUXILIARY", "x"),
	}, 1))
}

func TestStyleTooOftenUsedVerbRule_ToAddedLemma(t *testing.T) {
	rule := NewStyleTooOftenUsedVerbRule(nil)
	// Java: getLemmaForPosTagStartsWith("VER:", token)
	require.Equal(t, "laufen", rule.ToAddedLemma(atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen")))
	require.Equal(t, "laufen", rule.ToAddedLemma(atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen")))
	// No VER: reading → null/empty (no surface invent)
	require.Equal(t, "", rule.ToAddedLemma(atrWithPOS("schön", "ADJ:PRD:GRU", "schön")))
	// VER: reading with nil lemma → empty string path (Java returns null lemma)
	pos := "VER:INF:NON"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", &pos, nil), 0)
	require.Equal(t, "", rule.ToAddedLemma(tok))
	// First VER: reading wins when multi-tag
	adjTag, verTag := "ADJ:PRD:GRU", "VER:INF:NON"
	lemmaAdj, lemmaVer := "alt", "laufen"
	readings := []*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("x", &adjTag, &lemmaAdj),
		languagetool.NewAnalyzedToken("x", &verTag, &lemmaVer),
	}
	multi := languagetool.NewAnalyzedTokenReadingsList(readings, 0)
	require.Equal(t, "laufen", rule.ToAddedLemma(multi))
}

func TestStyleTooOftenUsedVerbRule_Match_MinWordCountGate(t *testing.T) {
	// Java getTooOftenUsedWords: if numWords < MIN_WORD_COUNT (100) → empty.
	rule := NewStyleTooOftenUsedVerbRule(nil)
	require.Equal(t, 100, rule.MinWordCount)

	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"),
		atrWithPOS("schnell", "ADV", "schnell"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("weiter", "ADV", "weiter"),
		atrWithPOS(".", "PKT", "."),
	))
	// Short text: only 2 counted VER lemmas → under 100 → no matches
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2}))
}

func TestStyleTooOftenUsedVerbRule_Match_OverusedLemma(t *testing.T) {
	// Bypass MIN_WORD_COUNT only for percent-threshold unit test (Java production uses 100).
	rule := NewStyleTooOftenUsedVerbRule(nil)
	rule.MinWordCount = 0

	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"),
		atrWithPOS("schnell", "ADV", "schnell"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("weiter", "ADV", "weiter"),
		atrWithPOS(".", "PKT", "."),
	))
	// "laufen" is 100% of verbs ≥ DEFAULT_MIN_PERCENT 5 → both hits
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.GreaterOrEqual(t, len(ms), 2)
	wantMsg := "Das Verb wird häufiger verwendet als 5% aller Verben. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
	for _, m := range ms {
		require.Equal(t, wantMsg, m.GetMessage())
		r, ok := m.GetRule().(*rules.AbstractStyleTooOftenUsedWordRule)
		require.True(t, ok)
		require.Equal(t, "TOO_OFTEN_USED_VERB_DE", r.GetID())
	}

	// Two lemmas each 50% ≥ 5 → both flag once
	s3 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("gingen", "VER:3:PLU:PRT:NON", "gehen"),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("weiter", "ADV", "weiter"),
		atrWithPOS(".", "PKT", "."),
	))
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})
	require.Equal(t, 2, len(ms2))

	// Raise threshold above 50%: neither flags
	rule.MinPercent = 51
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3}))
}

func TestStyleTooOftenUsedVerbRule_Match_ExceptionTagsNotCounted(t *testing.T) {
	// VER:MOD, VER:AUX, and multi-tag with ART/ADJ must not enter the map.
	rule := NewStyleTooOftenUsedVerbRule(nil)
	rule.MinWordCount = 0

	// Only pure VER: "laufen"; modal/aux are exceptions even though isToCountedWord.
	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("können", "VER:MOD:3:PLU:PRS", "können"),
		atrWithPOS("laufen", "VER:INF:NON", "laufen"),
		atrWithPOS("und", "KON", "und"),
		atrWithPOS("sind", "VER:AUX:3:PLU:PRS", "sein"),
		atrWithPOS("gekommen", "VER:PA2:NON", "kommen"),
		atrWithPOS(".", "PKT", "."),
	))
	// counted: laufen + kommen (not können MOD, not sind AUX) → 50% each ≥ 5 → two matches
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 2, len(ms))
	surfaces := []string{extractSurfaceDE(s, ms[0]), extractSurfaceDE(s, ms[1])}
	require.ElementsMatch(t, []string{"laufen", "gekommen"}, surfaces)

	// VER: + ADJ multi-tag exception (bare ADJ prefix) must not be counted
	verAdj := atrWithPOSMulti("x", "x", "VER:INF:NON", "ADJ:PRD:GRU")
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"),
		verAdj,
		atrWithPOS(".", "PKT", "."),
	))
	// only "laufen" counted → 100% ≥ 5 → one match
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{s2})
	require.Equal(t, 1, len(ms2))
	require.Equal(t, "laufen", extractSurfaceDE(s2, ms2[0]))

	// VER: + ART multi-tag exception
	verArt := atrWithPOSMulti("y", "y", "VER:INF:NON", "ART:DEF:NOM:SIN:NEU")
	s3 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("gehen", "VER:3:PLU:PRS:NON", "gehen"),
		verArt,
		atrWithPOS(".", "PKT", "."),
	))
	ms3 := rule.MatchList([]*languagetool.AnalyzedSentence{s3})
	require.Equal(t, 1, len(ms3))
	require.Equal(t, "gehen", extractSurfaceDE(s3, ms3[0]))
}

func TestStyleTooOftenUsedVerbRule_Match_DirectSpeechExclusion(t *testing.T) {
	// Java: withoutDirectSpeech only when setWithoutDirectSpeech(true).
	// Opening quote with next token !isWhitespaceBefore → enter DS; content not counted.
	// Intro uses non-VER so only outside "ging" is counted when DS is excluded
	// (adj/noun twins can keep "sagte" because those rules do not count VER:).
	rule := NewStyleTooOftenUsedVerbRule(nil)
	rule.MinWordCount = 0
	rule.WithoutDirectSpeech = true

	openQ := atrWithPOS("\"", "\"", "\"")
	v1 := atrWithPOS("laufen", "VER:INF:NON", "laufen")
	v1.SetWhitespaceBefore(false)
	v2 := atrWithPOS("laufen", "VER:INF:NON", "laufen")
	v2.SetWhitespaceBefore(true)
	v3 := atrWithPOS("laufen", "VER:INF:NON", "laufen")
	v3.SetWhitespaceBefore(true)
	endQ := atrWithPOS("\"", "\"", "\"")
	endQ.SetWhitespaceBefore(false)
	ging := atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen")
	ging.SetWhitespaceBefore(true)

	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("rief", "INTJ", "rief"), // non-VER intro
		openQ,
		v1,
		v2,
		v3,
		endQ,
		atrWithPOS("und", "KON", "und"),
		ging,
		atrWithPOS("fort", "ADV", "fort"),
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
	require.Equal(t, 1, len(ms), "only outside-quote overused verb should match")
	require.Equal(t, "ging", extractSurfaceDE(s, ms[0]))

	// Default WithoutDirectSpeech=false: quotes do not exclude; "laufen" is majority
	rule.WithoutDirectSpeech = false
	msAll := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.GreaterOrEqual(t, len(msAll), 3) // three "laufen" at least
}

func TestStyleTooOftenUsedVerbRule_Match_UntaggedNoInvent(t *testing.T) {
	// Soft/untagged SplitAndAnalyze must not invent POS → no matches.
	rule := NewStyleTooOftenUsedVerbRule(nil)
	rule.MinWordCount = 0
	plain := languagetool.SplitAndAnalyze("Sie laufen schnell. Dann laufen sie weiter. Sie laufen fort.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
}

func TestStyleTooOftenUsedVerbRule_WithDefaultLimit(t *testing.T) {
	r := NewStyleTooOftenUsedVerbRuleWithDefaultLimit(nil)
	require.Equal(t, 5, r.MinPercent)
	require.Equal(t, 100, r.MinWordCount)
	require.Equal(t, "TOO_OFTEN_USED_VERB_DE", r.GetID())
}
