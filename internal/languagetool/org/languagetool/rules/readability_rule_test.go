package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// demoLTReadability ports new JLanguageTool(new Demo()) for multi-sentence Analyze.
func demoLTReadability() *languagetool.JLanguageTool {
	lt := languagetool.NewJLanguageTool("xx")
	lt.IgnoredCharacters = languagetool.GermanIgnoredCharactersRegex
	lt.SentenceTokenizer = tokenizers.NewSimpleSentenceTokenizer()
	return lt
}

func TestReadabilityRule_IDsAndMetadata(t *testing.T) {
	simple := NewReadabilityRule(true, -1)
	diff := NewReadabilityRule(false, -1)

	require.Equal(t, "READABILITY_RULE_SIMPLE", simple.GetID())
	require.Equal(t, "READABILITY_RULE_DIFFICULT", diff.GetID())
	require.Equal(t, "READABILITY_RULE_SIMPLE", simple.GetIDFor(true))
	require.Equal(t, "READABILITY_RULE_DIFFICULT", simple.GetIDFor(false))

	require.Equal(t, "Readability: Too easy text", simple.GetDescription())
	require.Equal(t, "Readability: Too difficult text", diff.GetDescription())

	require.Equal(t, 3, simple.Level)
	require.Equal(t, 3, diff.Level)
	require.True(t, simple.IsDefaultOff())
	require.True(t, diff.IsDefaultOff())
	require.Equal(t, ITSStyle, simple.GetLocQualityIssueType())
	require.NotNil(t, simple.GetCategory())
	require.Equal(t, "TEXT_ANALYSIS", simple.GetCategory().GetID().String())
	require.Equal(t, -1, simple.MinToCheckParagraph())

	require.Equal(t, "Level of readability 0 (very difficult) to 6 (very easy):", simple.GetConfigureText())
	opts := simple.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 3, opts[0].DefaultValue)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 6, opts[0].MaxConfigurableValue)
}

func TestReadabilityRule_CtorLevelAndUserConfig(t *testing.T) {
	r := NewReadabilityRule(false, 5)
	require.Equal(t, 5, r.Level)

	r2 := NewReadabilityRuleFull(nil, nil, false, -1, true)
	require.Equal(t, 3, r2.Level)
	require.False(t, r2.IsDefaultOff())

	uc := languagetool.NewUserConfig()
	// Config key is getId() at construction time → DIFFICULT.
	uc.SetConfigValueByID("READABILITY_RULE_DIFFICULT", []any{1})
	r3 := NewReadabilityRuleFull(nil, uc, false, 5, false)
	require.Equal(t, 1, r3.Level)
}

func TestReadabilityRule_PrintMessageLevel(t *testing.T) {
	r := NewReadabilityRule(false, 3)
	require.Equal(t, " {Level 0: Very difficult}", r.PrintMessageLevel(0))
	require.Equal(t, " {Level 1: Difficult}", r.PrintMessageLevel(1))
	require.Equal(t, " {Level 2: Fairly difficult}", r.PrintMessageLevel(2))
	require.Equal(t, " {Level 3: Medium}", r.PrintMessageLevel(3))
	require.Equal(t, " {Level 4: Fairly easy}", r.PrintMessageLevel(4))
	require.Equal(t, " {Level 5: Easy}", r.PrintMessageLevel(5))
	require.Equal(t, " {Level 6: Very easy}", r.PrintMessageLevel(6))
	require.Equal(t, "", r.PrintMessageLevel(7))
	require.Equal(t, "", r.PrintMessageLevel(-1))
}

func TestReadabilityRule_GetMessage(t *testing.T) {
	diff := NewReadabilityRule(false, 3)
	msg := diff.GetMessage(1, 40, 20, 2)
	require.Equal(t,
		"Readability: The text of this paragraph is too difficult {Level 1: Difficult}. Too many words per sentence and too many syllables per word.",
		msg)

	easy := NewReadabilityRule(true, 3)
	msgE := easy.GetMessage(5, 85, 8, 1)
	require.Equal(t,
		"Readability: The text of this paragraph is too simple {Level 5: Easy}. Too few words per sentence and too few syllables per word.",
		msgE)
}

func TestReadabilityRule_GetReadabilityLevel_Boundaries(t *testing.T) {
	r := NewReadabilityRule(false, 3)
	// fre < 30 → 0
	require.Equal(t, 0, r.GetReadabilityLevel(29.999))
	require.Equal(t, 0, r.GetReadabilityLevel(-10))
	// fre < 50 → 1
	require.Equal(t, 1, r.GetReadabilityLevel(30))
	require.Equal(t, 1, r.GetReadabilityLevel(49.999))
	// fre < 60 → 2
	require.Equal(t, 2, r.GetReadabilityLevel(50))
	require.Equal(t, 2, r.GetReadabilityLevel(59.999))
	// fre < 70 → 3
	require.Equal(t, 3, r.GetReadabilityLevel(60))
	require.Equal(t, 3, r.GetReadabilityLevel(69.999))
	// fre < 80 → 4
	require.Equal(t, 4, r.GetReadabilityLevel(70))
	require.Equal(t, 4, r.GetReadabilityLevel(79.999))
	// fre < 90 → 5
	require.Equal(t, 5, r.GetReadabilityLevel(80))
	require.Equal(t, 5, r.GetReadabilityLevel(89.999))
	// else → 6
	require.Equal(t, 6, r.GetReadabilityLevel(90))
	require.Equal(t, 6, r.GetReadabilityLevel(100))
}

func TestReadabilityRule_GetFleschReadingEase(t *testing.T) {
	r := NewReadabilityRule(false, 3)
	// FRE = 206.835 - 1.015*asl - 84.6*asw
	require.InDelta(t, 206.835-1.015*10-84.6*1.5, r.GetFleschReadingEase(10, 1.5), 1e-9)
	require.InDelta(t, 206.835, r.GetFleschReadingEase(0, 0), 1e-9)
	require.InDelta(t, 206.835-1.015*20-84.6*2, r.GetFleschReadingEase(20, 2), 1e-9)
}

func TestReadabilityRule_SimpleSyllablesCount(t *testing.T) {
	r := NewReadabilityRule(false, 3)
	require.Equal(t, 0, r.SimpleSyllablesCount(""))
	require.Equal(t, 1, r.SimpleSyllablesCount("a"))
	require.Equal(t, 1, r.SimpleSyllablesCount("I"))
	require.Equal(t, 1, r.SimpleSyllablesCount("cat"))
	require.Equal(t, 1, r.SimpleSyllablesCount("dog"))
	// "the": t,h,e — loop counts no vowel with next; last e silent path → nSyllables 0 → 1
	require.Equal(t, 1, r.SimpleSyllablesCount("the"))
	// "to": t,o — o is last vowel != e → +1 after loop counted? loop i=0 't'; i ends; last o vowel → +1 → 1
	require.Equal(t, 1, r.SimpleSyllablesCount("to"))
	// "idea": i,d,e,a
	// i with d: vowel single → 1; e with a: e+a double lastDouble; last a vowel → after lastDouble cleared in loop?
	// Trace:
	// i=0 'i' vowel cn='d' not double → n=1
	// i=1 'd' not
	// i=2 'e' vowel cn='a' → e+a → lastDouble=true
	// last c='a' cl='e': not (e+s/d or u+e); isVowel(a)&&a!='e' → n++ → 2
	require.Equal(t, 2, r.SimpleSyllablesCount("idea"))
	// "people": p,e,o,p,l,e
	// e+o double; last e with cl=l → not special; isVowel(e)&&e!='e' false → n from loop
	// i=0 p; i=1 e cn=o → lastDouble; i=2 o lastDouble → n++ lastDouble=false; i=3 p; i=4 l
	// last e cl=l → no change; n=1 → 1
	require.Equal(t, 1, r.SimpleSyllablesCount("people"))
	// "needed": n,e,e,d,e,d — e+e double; last d cl=e → e+d → n--
	// i=0 n; i=1 e cn=e → lastDouble; i=2 e lastDouble → n=1 lastDouble=false; i=3 d; i=4 e cn=d → n=2
	// last d cl=e → e+d → n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("needed"))
	// "true": t,r,u,e — last e cl=u → u+e → n--
	// i=0 t; i=1 r; i=2 u cn=e → u+e not in double list (u only with i/a); n=1
	// last e cl=u → n=0 → 1
	require.Equal(t, 1, r.SimpleSyllablesCount("true"))
	// "beautiful" — multi-syllable word (Java algorithm)
	require.GreaterOrEqual(t, r.SimpleSyllablesCount("beautiful"), 2)
	// "responsibility" — many vowels
	require.GreaterOrEqual(t, r.SimpleSyllablesCount("responsibility"), 4)
	// trailing "es" / "ed"
	require.Equal(t, r.SimpleSyllablesCount("likes"), 1) // li+ke+s: i; e; last s cl=e → --
	// "xx" no vowels → 1
	require.Equal(t, 1, r.SimpleSyllablesCount("xx"))
}

func TestReadabilityRule_Match_BelowMinWords(t *testing.T) {
	lt := demoLTReadability()
	// Difficult: few words → no per-para match; final rLevel from few short words may or may not
	// be < level — but no ruleMatches collected if nWords < MIN_WORDS.
	r := NewReadabilityRule(false, 3)
	matches := r.MatchList(lt.Analyze("Hi there."))
	require.Empty(t, matches)
	// too easy single short sentence
	rEasy := NewReadabilityRule(true, 3)
	require.Empty(t, rEasy.MatchList(lt.Analyze("Hi there friend.")))
}

func TestReadabilityRule_Match_Difficult(t *testing.T) {
	lt := demoLTReadability()
	// Many long multisyllabic words, one sentence → low FRE → rLevel low.
	// difficult test (tooEasy=false) fires when rLevel < level (default 3).
	// Final quirk: (!tooEasy && rLevel < level) returns collected matches without needing nParagraph>1.
	var b strings.Builder
	words := []string{
		"incomprehensibilities", "conceptualizations", "internationalization",
		"responsibilities", "characterization", "individualization",
		" entation", "constitutionalism", "overintellectualization",
		"counterrevolutionary", "electroencephalography",
	}
	// ensure >= MIN_WORDS (10)
	for i, w := range words {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strings.TrimSpace(w))
	}
	b.WriteByte('.')
	text := b.String()
	r := NewReadabilityRule(false, 3)
	matches := r.MatchList(lt.Analyze(text))
	require.NotEmpty(t, matches, "expected difficult match for long multisyllabic words")
	require.Contains(t, matches[0].Message, "too difficult")
	require.Contains(t, matches[0].Message, "Too many words per sentence")
	// Span: first sentence tokens[1].start .. tokens[MARK_WORDS].end (MARK_WORDS=3)
	require.GreaterOrEqual(t, matches[0].GetFromPos(), 0)
	require.Greater(t, matches[0].GetToPos(), matches[0].GetFromPos())
	// tokens without whitespace: SENT_START + words + '.' → MARK_WORDS index is 3rd word end
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents)
	toks := sents[0].GetTokensWithoutWhitespace()
	require.Greater(t, len(toks), readabilityMarkWords)
	require.Equal(t, toks[1].GetStartPos(), matches[0].GetFromPos())
	require.Equal(t, toks[readabilityMarkWords].GetEndPos(), matches[0].GetToPos())
}

func TestReadabilityRule_Match_TooEasy_NeedsMultipleParagraphs(t *testing.T) {
	lt := demoLTReadability()
	// Easy text: short words, short sentences → high FRE → high rLevel.
	// Single paragraph: Java final condition requires nParagraph > 1 for tooEasyTest → empty.
	easyPara := "The cat sat on the mat. The dog ran to the park. A man saw a big red car near the house."
	r := NewReadabilityRule(true, 3)
	require.Empty(t, r.MatchList(lt.Analyze(easyPara)), "single paragraph tooEasy must not return matches (Java quirk)")

	// Two paragraphs (≥ MIN_WORDS each preferred): nParagraph becomes 2; final may return matches.
	twoPara := easyPara + "\n\n" + "The sun is hot. The sky is blue. The bird can fly high over the tall tree near us now."
	matches := r.MatchList(lt.Analyze(twoPara))
	// If FRE level > 3, expect matches; otherwise still a valid empty (level not exceeded).
	if len(matches) > 0 {
		require.Contains(t, matches[0].Message, "too simple")
		require.Contains(t, matches[0].Message, "Too few words per sentence")
	} else {
		// Verify overall level is not above threshold (or MIN_WORDS not met in a way that collected nothing)
		// Re-check with very easy forced text repeated.
		easyLots := strings.Repeat("The cat sat on the mat. ", 8) + "\n\n" + strings.Repeat("The dog ran by me. ", 8)
		matches2 := r.MatchList(lt.Analyze(easyLots))
		require.NotEmpty(t, matches2, "very easy multi-paragraph text should match tooEasy rule")
		require.Contains(t, matches2[0].Message, "too simple")
	}
}

func TestReadabilityRule_Match_Difficult_DoesNotNeedTwoParagraphs(t *testing.T) {
	lt := demoLTReadability()
	// Confirm single-paragraph difficult path works (final || branch).
	text := "incomprehensibilities conceptualizations internationalization responsibilities characterization individualization differentiation constitutionalism overintellectualization counterrevolutionary."
	r := NewReadabilityRule(false, 6) // level 6: any rLevel < 6 matches (almost always for real text with asw>0)
	// Even medium text has fre < something; with level 6, rLevel is at most 6, < 6 needed.
	// Pure hard text should be rLevel 0..2.
	r.Level = 3
	matches := r.MatchList(lt.Analyze(text))
	require.NotEmpty(t, matches)
}

func TestReadabilityRule_Match_MessageUsesParagraphLevel(t *testing.T) {
	lt := demoLTReadability()
	r := NewReadabilityRule(false, 3)
	text := "incomprehensibilities conceptualizations internationalization responsibilities characterization individualization differentiation constitutionalism overintellectualization counterrevolutionary electroencephalography."
	matches := r.MatchList(lt.Analyze(text))
	require.NotEmpty(t, matches)
	require.Regexp(t, `\{Level [0-6]: `, matches[0].Message)
}

func TestReadabilityRule_GetAllStatsAfterMatch(t *testing.T) {
	lt := demoLTReadability()
	r := NewReadabilityRule(false, 3)
	text := "The cat sat on the mat and looked at the dog carefully nearby more words today."
	_ = r.MatchList(lt.Analyze(text))
	require.Greater(t, r.GetAllWords(), 0)
	require.Greater(t, r.GetAllSentences(), 0)
	require.Greater(t, r.GetAllSyllables(), 0)
}
