package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Behavior matrix twin for org.languagetool.rules.de.StyleTooOftenUsedAdjectiveRule
// (+ AbstractStyleTooOftenUsedWordRule match algorithm). Java has no dedicated *Test class.

func TestStyleTooOftenUsedAdjectiveRule_JavaMeta(t *testing.T) {
	// Java StyleTooOftenUsedAdjectiveRule + AbstractStyleTooOftenUsedWordRule ctor.
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_DE", rule.GetID())
	require.Equal(t, "Statistische Stilanalyse: Zu häufig genutztes Adjektiv", rule.GetDescription())
	// Java getConfigureText
	require.Equal(t, styleTooOftenAdjConfigureText, rule.GetConfigureText())
	require.Equal(t, 5, rule.MinPercent)        // DEFAULT_MIN_PERCENT
	require.Equal(t, 5, rule.DefaultMinPercent) // Java defaultMinPercent
	require.Equal(t, 100, rule.MinWordCount)    // Abstract MIN_WORD_COUNT
	require.True(t, rule.IsDefaultOff())        // DEFAULT_ACTIVATION = false
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	cat := rule.GetCategory()
	require.NotNil(t, cat)
	require.Equal(t, "CREATIVE_WRITING", cat.GetID().String())
	require.True(t, cat.IsDefaultOff())

	// getLimitMessage(limit) — Java has space before period: "Adjektive . "
	require.Equal(t,
		"Das Adjektiv wird häufiger verwendet als 5% aller Adjektive . Möglicherweise ist es besser es durch ein Synonym zu ersetzen.",
		rule.LimitMessage(5))
	require.Equal(t,
		"Das Adjektiv wird häufiger verwendet als 12% aller Adjektive . Möglicherweise ist es besser es durch ein Synonym zu ersetzen.",
		rule.LimitMessage(12))

	// getRuleOptions: RuleOption(defaultMinPercent, getConfigureText(), 1, 100)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 5, opts[0].DefaultValue)
	require.Equal(t, rule.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 1, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestStyleTooOftenUsedAdjectiveRule_UserConfigMinPercentAndThesaurus(t *testing.T) {
	// Java abstract getMinPercent(userConfig, DEFAULT): configValueByID(getId())[0]
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("TOO_OFTEN_USED_ADJECTIVE_DE", []any{17})
	ls := languagetool.NewLinguServices()
	uc.LinguServices = ls

	rule := NewStyleTooOftenUsedAdjectiveRuleWithUserConfig(nil, uc)
	require.Equal(t, 17, rule.MinPercent)
	require.Equal(t, 5, rule.DefaultMinPercent) // RuleOption default stays DEFAULT_MIN_PERCENT
	// Java setThesaurusRelevantRule(this)
	require.Equal(t, rule, ls.ThesaurusRelevantRule)

	// nil userConfig: defaults, no thesaurus registration
	ruleNil := NewStyleTooOftenUsedAdjectiveRuleWithUserConfig(nil, nil)
	require.Equal(t, 5, ruleNil.MinPercent)

	// userConfig without LinguServices: no panic, default min percent
	uc2 := languagetool.NewUserConfig()
	rule2 := NewStyleTooOftenUsedAdjectiveRuleWithUserConfig(nil, uc2)
	require.Equal(t, 5, rule2.MinPercent)
}

func TestStyleTooOftenUsedAdjectiveRule_IsToCountedWordAndException(t *testing.T) {
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)

	// isToCountedWord = hasPosTagStartingWith("ADJ:")
	require.True(t, rule.IsToCountedWord(atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön")))
	require.True(t, rule.IsToCountedWord(atrWithPOS("schöner", "ADJ:NOM:SIN:MAS:KOM:IND", "schön")))
	require.True(t, rule.IsToCountedWord(atrWithPOS("schön", "ADJ:PRD:GRU", "schön")))
	require.False(t, rule.IsToCountedWord(atrWithPOS("schnell", "ADV", "schnell")))
	require.False(t, rule.IsToCountedWord(atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto")))
	// ADJ without colon is not ADJ: prefix
	require.False(t, rule.IsToCountedWord(atrWithPOS("x", "ADJ", "x")))

	// isException = PRO: || ADV: || ZUS (starts with). Plain "ADV" is NOT "ADV:".
	tokens := []*languagetool.AnalyzedTokenReadings{
		sentStartATR(),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("dort", "ADV:LOK", "dort"),
		atrWithPOS("mit", "ZUS", "mit"), // particle / ZUS prefix
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		atrWithPOS("herauf", "ZUS:VPART", "herauf"),
		atrWithPOS("schnell", "ADV", "schnell"), // bare ADV does not start with "ADV:"
	}
	require.True(t, rule.IsException(tokens, 1))  // PRO:
	require.True(t, rule.IsException(tokens, 2))  // ADV:LOK
	require.True(t, rule.IsException(tokens, 3))  // ZUS
	require.False(t, rule.IsException(tokens, 4)) // pure ADJ:
	require.True(t, rule.IsException(tokens, 5))  // ZUS:VPART
	require.False(t, rule.IsException(tokens, 6)) // bare ADV ≠ ADV:

	// Multi-tag: ADJ: + PRO: → counted word AND exception (Java ORs all tags)
	adjAndPro := atrWithPOSMulti("mein", "mein", "ADJ:NOM:SIN:MAS:GRU:IND", "PRO:POS:NOM:SIN:MAS")
	require.True(t, rule.IsToCountedWord(adjAndPro))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{sentStartATR(), adjAndPro}, 1))

	// Multi-tag: ADJ: + ADV:MOD → exception
	adjAndAdv := atrWithPOSMulti("gut", "gut", "ADJ:PRD:GRU", "ADV:MOD")
	require.True(t, rule.IsToCountedWord(adjAndAdv))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{sentStartATR(), adjAndAdv}, 1))
}

func TestStyleTooOftenUsedAdjectiveRule_ToAddedLemma(t *testing.T) {
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	// Java: getLemmaForPosTagStartsWith("ADJ:", token)
	require.Equal(t, "schön", rule.ToAddedLemma(atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön")))
	require.Equal(t, "schön", rule.ToAddedLemma(atrWithPOS("schöner", "ADJ:NOM:SIN:MAS:KOM:IND", "schön")))
	// No ADJ: reading → null/empty (no surface invent)
	require.Equal(t, "", rule.ToAddedLemma(atrWithPOS("schnell", "ADV", "schnell")))
	// ADJ: reading with nil lemma → empty string path (Java returns null lemma)
	pos := "ADJ:PRD:GRU"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", &pos, nil), 0)
	require.Equal(t, "", rule.ToAddedLemma(tok))
	// First ADJ: reading wins when multi-tag
	multi := atrWithPOSMulti("alt", "alt", "SUB:NOM:SIN:NEU", "ADJ:PRD:GRU")
	// readings order: SUB first, ADJ second — lemma for ADJ: is "alt" (lemma shared in helper)
	// Rebuild with distinct lemmas to assert first matching ADJ: reading
	subTag, adjTag := "SUB:NOM:SIN:NEU", "ADJ:PRD:GRU"
	lemmaSub, lemmaAdj := "Substantiv", "alt"
	readings := []*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("alt", &subTag, &lemmaSub),
		languagetool.NewAnalyzedToken("alt", &adjTag, &lemmaAdj),
	}
	multi = languagetool.NewAnalyzedTokenReadingsList(readings, 0)
	require.Equal(t, "alt", rule.ToAddedLemma(multi))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_MinWordCountGate(t *testing.T) {
	// Java getTooOftenUsedWords: if numWords < MIN_WORD_COUNT (100) → empty.
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	require.Equal(t, 100, rule.MinWordCount)

	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Noch", "ADV", "noch"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	// Short text: only 2 counted ADJ lemmas → under 100 → no matches
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2}))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_OverusedLemma(t *testing.T) {
	// Bypass MIN_WORD_COUNT only for percent-threshold unit test (Java production uses 100).
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	rule.MinWordCount = 0

	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Noch", "ADV", "noch"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	// "schön" is 100% of adjectives ≥ DEFAULT_MIN_PERCENT 5 → both hits
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.GreaterOrEqual(t, len(ms), 2)
	wantMsg := "Das Adjektiv wird häufiger verwendet als 5% aller Adjektive . Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
	for _, m := range ms {
		require.Equal(t, wantMsg, m.GetMessage())
		r, ok := m.GetRule().(*rules.AbstractStyleTooOftenUsedWordRule)
		require.True(t, ok)
		require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_DE", r.GetID())
	}

	// Two lemmas each 50% ≥ 5 → both flag once
	s3 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("hässliches", "ADJ:NOM:SIN:NEU:GRU:IND", "hässlich"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})
	require.Equal(t, 2, len(ms2))

	// Raise threshold above 50%: neither flags
	rule.MinPercent = 51
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3}))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_ExceptionTagsNotCounted(t *testing.T) {
	// Tokens that have ADJ: but also PRO:/ADV:/ZUS are exceptions and must not enter the map.
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	rule.MinWordCount = 0

	// Only pure ADJ: "schön"; "gut" has ADJ:+ADV:MOD → isException (Java ADV: prefix)
	adjAdv := atrWithPOSMulti("gut", "gut", "ADJ:PRD:GRU", "ADV:MOD")
	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		adjAdv,
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(".", "PKT", "."),
	))
	// only "schön" counted → 100% ≥ 5 → one match on "schönes" only
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "schönes", extractSurfaceDE(s, ms[0]))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_DirectSpeechExclusion(t *testing.T) {
	// Java: withoutDirectSpeech only when setWithoutDirectSpeech(true).
	// Opening quote with next token !isWhitespaceBefore → enter DS; content not counted.
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	rule.MinWordCount = 0
	rule.WithoutDirectSpeech = true

	// Sentence: Er sagte "schön schön schön" und nettes Auto.
	// Inside quotes: three "schön" ADJ; outside: "nettes" ADJ.
	// With exclusion: only "nett" counted → 100% of adjectives → match on "nettes" only.
	openQ := atrWithPOS("\"", "\"", "\"")
	good1 := atrWithPOS("schön", "ADJ:PRD:GRU", "schön")
	good1.SetWhitespaceBefore(false)
	good2 := atrWithPOS("schön", "ADJ:PRD:GRU", "schön")
	good2.SetWhitespaceBefore(true)
	good3 := atrWithPOS("schön", "ADJ:PRD:GRU", "schön")
	good3.SetWhitespaceBefore(true)
	endQ := atrWithPOS("\"", "\"", "\"")
	endQ.SetWhitespaceBefore(false)
	nice := atrWithPOS("nettes", "ADJ:NOM:SIN:NEU:GRU:IND", "nett")
	nice.SetWhitespaceBefore(true)

	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("sagte", "VER:3:SIN:PRT:SFT", "sagen"),
		openQ,
		good1,
		good2,
		good3,
		endQ,
		atrWithPOS("und", "KON", "und"),
		nice,
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
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
	require.Equal(t, 1, len(ms), "only outside-quote overused adjective should match")
	require.Equal(t, "nettes", extractSurfaceDE(s, ms[0]))

	// Default WithoutDirectSpeech=false: quotes do not exclude; "schön" is majority
	rule.WithoutDirectSpeech = false
	msAll := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.GreaterOrEqual(t, len(msAll), 3) // three "schön" at least
}

func TestStyleTooOftenUsedAdjectiveRule_Match_UntaggedNoInvent(t *testing.T) {
	// Soft/untagged SplitAndAnalyze must not invent POS → no matches.
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	rule.MinWordCount = 0
	plain := languagetool.SplitAndAnalyze("Ein schönes Auto. Noch ein schönes Haus. Ein schönes Kind.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
}

func TestStyleTooOftenUsedAdjectiveRule_WithDefaultLimit(t *testing.T) {
	r := NewStyleTooOftenUsedAdjectiveRuleWithDefaultLimit(nil)
	require.Equal(t, 5, r.MinPercent)
	require.Equal(t, 100, r.MinWordCount)
	require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_DE", r.GetID())
}

// extractSurfaceDE returns the token surface covered by match positions (document offset 0).
func extractSurfaceDE(sent *languagetool.AnalyzedSentence, m *rules.RuleMatch) string {
	from, to := m.GetFromPos(), m.GetToPos()
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok.GetStartPos() == from && tok.GetEndPos() == to {
			return tok.GetToken()
		}
	}
	return ""
}
