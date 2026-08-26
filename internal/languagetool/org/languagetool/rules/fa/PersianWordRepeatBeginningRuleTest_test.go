package fa

// Outcome twins for org.languagetool.rules.fa.PersianWordRepeatBeginningRule.
// Java has no dedicated *Test.java; matrix follows GreekWordRepeatBeginningRuleTest pattern
// + Java example pair surfaces and ADVERBS set.
//
// Java setup:
//   JLanguageTool lt = new JLanguageTool(new Persian());
//   rule.match(lt.check text-level path via analyzed sentences)
//
// Go analysis path for this text-level rule:
//   NewJLanguageTool("fa").Analyze — Persian SRX sentence tokenize + PersianWordTokenizer
//   (splits Arabic comma U+060C; isAdverb sees bare "همچنین", not "همچنین،").
//   rule.MatchList(sents) isolates PERSIAN_WORD_REPEAT_BEGINNING_RULE outcomes.
//
// Messages: MessagesBundle_fa.properties keys used by WordRepeatBeginningRule
// (no invent soft strings).

import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// faWRBMessages ports MessagesBundle_fa.properties keys used by WordRepeatBeginningRule.
func faWRBMessages() map[string]string {
	return map[string]string{
		// MessagesBundle_fa desc_repetition_beginning* (exact property values)
		"desc_repetition_beginning":           "جملهٔ  بعدی هم  با کلمه‌ای مشابه شروع شده‌است",
		"desc_repetition_beginning_word":      "سه جملهٔ پشت سر هم با کلمه‌ای مشابه شروع شده‌اند",
		"desc_repetition_beginning_adv":       "دو جملهٔ پست سر هم با حرف ربط مشابه شروع شده‌اند",
		"desc_repetition_beginning_thesaurus": "Consider rewording the sentence or use a thesaurus to find a synonym.",
		// Java Categories.REPETITIONS_STYLE uses typo key cateogry_repetitions_style
		"cateogry_repetitions_style": "Repetitions (Style)",
		"category_repetitions":       "Repetitions",
	}
}

func faWRBAtrSurface(tok string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(tok, nil, nil))
}

// analyzeFAWRBText ports JLanguageTool(Persian) sentence tokenize + getAnalyzedSentence per part.
func analyzeFAWRBText(t *testing.T, text string) []*languagetool.AnalyzedSentence {
	t.Helper()
	lt := languagetool.NewJLanguageTool("fa")
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents, "Analyze(%q)", text)
	return sents
}

func utf16LenFA(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// TestPersianWordRepeatBeginningRule_Rule — adverb 2-sentence match, non-adverb 0, word 3-sentence.
func TestPersianWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewPersianWordRepeatBeginningRule(faWRBMessages())

	// ================== correct sentences ===================
	// two successive sentences that start with the same non-adverb word.
	require.Equal(t, 0, len(rule.MatchList(analyzeFAWRBText(t, "این خوب است. این بهتر است."))))
	// two successive different adverbs — no repeat.
	require.Equal(t, 0, len(rule.MatchList(analyzeFAWRBText(t, "همچنین، خیابان مسکونی است. بنابراین، بهتر است."))))
	// single adverb sentence only — no previous to compare.
	require.Equal(t, 0, len(rule.MatchList(analyzeFAWRBText(t, "همچنین، خیابان مسکونی است."))))

	// =================== errors =============================
	// Java example shape: two successive sentence-initial همچنین (with Arabic comma after).
	matchesAdv := rule.MatchList(analyzeFAWRBText(t,
		"همچنین، خیابان مسکونی است. همچنین، به افتخار یک شاعر نامگذاری شده‌است."))
	require.Equal(t, 1, len(matchesAdv))
	// match.Rule is the rule instance passed to NewRuleMatch (same as Java getRule())
	require.Same(t, rule.WordRepeatBeginningRule, matchesAdv[0].GetRule())
	require.Equal(t, "PERSIAN_WORD_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, faWRBMessages()["desc_repetition_beginning_adv"], matchesAdv[0].GetShortMessage())
	require.Contains(t, matchesAdv[0].GetMessage(), faWRBMessages()["desc_repetition_beginning_thesaurus"])
	// Persian has no getSuggestions override → empty
	require.Empty(t, matchesAdv[0].GetSuggestedReplacements())
	// span is UTF-16 length of "همچنین" (all BMP)
	require.Equal(t, utf16LenFA("همچنین"), matchesAdv[0].GetToPos()-matchesAdv[0].GetFromPos())

	// each single-token ADVERBS member (multi-word set members never equal one token surface)
	for _, adv := range []string{"هم", "همچنین", "نیز", "بنابراین", "حتی", "چنانچه"} {
		text := adv + " یک است. " + adv + " دو است."
		ms := rule.MatchList(analyzeFAWRBText(t, text))
		require.Equal(t, 1, len(ms), "adverb %q should match on second successive start", adv)
		require.Equal(t, faWRBMessages()["desc_repetition_beginning_adv"], ms[0].GetShortMessage(), adv)
		require.Equal(t, utf16LenFA(adv), ms[0].GetToPos()-ms[0].GetFromPos(), adv)
	}

	// three successive sentences that start with the same non-exception non-adverb word.
	matchesWord := rule.MatchList(analyzeFAWRBText(t, "من اینجا هستم. من آنجا هستم. من خوب هستم."))
	require.Equal(t, 1, len(matchesWord))
	require.Equal(t, faWRBMessages()["desc_repetition_beginning_word"], matchesWord[0].GetShortMessage())
	require.Contains(t, matchesWord[0].GetMessage(), faWRBMessages()["desc_repetition_beginning_thesaurus"])
	require.Empty(t, matchesWord[0].GetSuggestedReplacements())
	require.Equal(t, utf16LenFA("من"), matchesWord[0].GetToPos()-matchesWord[0].GetFromPos())
}

// TestPersianWordRepeatBeginningRule_Metadata covers Java ctor id + example pair + category.
func TestPersianWordRepeatBeginningRule_Metadata(t *testing.T) {
	rule := NewPersianWordRepeatBeginningRule(faWRBMessages())
	require.Equal(t, "PERSIAN_WORD_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, 2, rule.MinToCheckParagraph())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), rule.GetCategory().GetID())
	require.Equal(t, faWRBMessages()["desc_repetition_beginning"], rule.GetDescription())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t,
		"همچنین، خیابان تقریباً کاملاً مسکونی است. <marker>همچنین</marker>، به افتخار یک شاعر نامگذاری شده‌است.",
		inc[0].GetExample())
	require.Equal(t, []string{"این خیابان"}, inc[0].GetCorrections())
	require.Equal(t,
		"همچنین، خیابان تقریباً مسکونی است. <marker>این خیابان</marker> به افتخار یک شاعر نامگذاری شده‌است.",
		rule.GetCorrectExamples()[0].GetExample())
}

// TestPersianWordRepeatBeginningRule_IsAdverb ports ADVERBS.contains(token.getToken()) exactly.
func TestPersianWordRepeatBeginningRule_IsAdverb(t *testing.T) {
	r := NewPersianWordRepeatBeginningRule(faWRBMessages())

	// single-token members (Java static block order)
	for _, adv := range []string{"هم", "همچنین", "نیز", "بنابراین", "حتی", "چنانچه"} {
		require.True(t, r.isAdverb(faWRBAtrSurface(adv)), "isAdverb(%q)", adv)
	}
	// multi-word members: contain only if token surface is the whole phrase (Java HashSet)
	for _, adv := range []string{"از یک سو", "از یک طرف", "از طرف ديگر"} {
		require.True(t, r.isAdverb(faWRBAtrSurface(adv)), "isAdverb whole phrase %q", adv)
	}
	// tokenizer would yield first word only — not in set as adverb start for multi-word phrases
	require.False(t, r.isAdverb(faWRBAtrSurface("از")))
	// non-adverbs / wrong surfaces
	require.False(t, r.isAdverb(faWRBAtrSurface("این")))
	require.False(t, r.isAdverb(faWRBAtrSurface("همچنین،"))) // no invent strip of Arabic comma
	// "دیگر" with Persian yeh U+06CC must not match Java "ديگر" (Arabic yeh U+064A)
	require.False(t, r.isAdverb(faWRBAtrSurface("از طرف دیگر")))
	require.False(t, r.isAdverb(nil))

	// set size matches Java static block (9 adds)
	require.Equal(t, 9, len(persianAdverbs))

	// base exception list still applies via WordRepeatBeginningRule (no FA isException override)
	require.Equal(t, 0, len(r.MatchList(analyzeFAWRBText(t,
		"2011: یک. 2011: دو. 2011: سه."))))
}
