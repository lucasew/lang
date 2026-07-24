package km

// Outcome twins for Khmer XmlRuleDisambiguator as used by Khmer.createDefaultDisambiguator:
// Java new XmlRuleDisambiguator(this) with useGlobalDisambiguation=false.
// Cases derived from official resource/km/disambiguation.xml rule patterns
// (TIME_ADJ, TAG_AS_NOUN, TAG_AS_NOUN_END, USAGE_ALL, DOWY_SA, TIME_AS_ADJ,
// TOBE_YOU*, CLS_YOU*, *PTOAL*, self_*, timoy, TOBE_LONG, ADJ_HEART,
// WITH_PREP*, PAST_TENSE_VERB, ADVERB_MARKER, MON_PREP)
// + real KhmerTagger / khmer.dict readings — same bar as EO/BR/DA RuleDisambiguator tests.
// Upstream has no KhmerRuleDisambiguatorTest.java.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	tagkm "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/km"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	kmtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/km"
	"github.com/stretchr/testify/require"
)

// loadKMXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Khmer())
// (useGlobalDisambiguation default false) over official resource/km/disambiguation.xml.
func loadKMXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	// Prefer process cache (tagging/km loader); fall back to discover for isolation.
	if x := tagkm.KhmerXmlRuleDisambiguator(); x != nil {
		return x
	}
	p := discoverKMDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "km", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverKMDisambiguationXML() string {
	if p := os.Getenv("LANG_KM_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "km",
		"src", "main", "resources", "org", "languagetool", "resource", "km", "disambiguation.xml")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func setupKMDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagkm.DiscoverKhmerPOSDict() == "" {
		t.Skip("khmer.dict not in tree")
	}
	tagkm.EnsureDefaultKhmerTagger()
	require.NotNil(t, tagkm.DefaultKhmerTagger)
	require.NotNil(t, tagkm.DefaultKhmerTagger.GetWordTagger())
	require.NotEmpty(t, tagkm.KhmerPOSDictPath(), "real khmer.dict must load")

	xml = loadKMXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("km/disambiguation.xml not in tree or failed to load")
	}
	// Official KM pack: 28 rules (TIME_ADJ … MON_PREP).
	require.GreaterOrEqual(t, len(xml.Rules), 28)
	return disambigxx.NewDemoDisambiguator(), xml
}

// TIME_ADJ: ពេល + ដែល → REPLACE first token to JJ.
func TestKhmerRuleDisambiguator_TimeAdj(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ពេល ដែល"
	require.Equal(t,
		"/[null]SENT_START ពេល/[ពេល]JJ|ពេល/[ពេល]NN  /[null]null ដែល/[ដែល]IN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ពេល/[ពេល]JJ  /[null]null ដែល/[ដែល]IN",
		myAssertDisambiguate(input, xmlDisam),
		"xml TIME_ADJ ពេល ដែល")
}

// TAG_AS_NOUN: ការ + VB|JJ|RB → marker becomes NN (lemma kept from NN reading when present).
func TestKhmerRuleDisambiguator_TagAsNoun(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START ការ/[ការ]NN|ការ/[ការ]VB  /[null]null សង្គ្រោះ/[សង្គ្រោះ]JJ|សង្គ្រោះ/[សង្គ្រោះ]NN|សង្គ្រោះ/[សង្គ្រោះ]VB",
		myAssertDisambiguate("ការ សង្គ្រោះ", demo),
		"demo ការ សង្គ្រោះ")
	require.Equal(t,
		"/[null]SENT_START ការ/[ការ]NN|ការ/[ការ]VB  /[null]null សង្គ្រោះ/[សង្គ្រោះ]NN",
		myAssertDisambiguate("ការ សង្គ្រោះ", xmlDisam),
		"xml TAG_AS_NOUN ការ សង្គ្រោះ")

	// ភាព + ល្អ (JJ|RB only) → REPLACE to NN (no NN reading: lemma falls back to surface lemma).
	require.Equal(t,
		"/[null]SENT_START ភាព/[ភាព]NN  /[null]null ល្អ/[ល្អ]JJ|ល្អ/[ល្អ]RB",
		myAssertDisambiguate("ភាព ល្អ", demo),
		"demo ភាព ល្អ")
	require.Equal(t,
		"/[null]SENT_START ភាព/[ភាព]NN  /[null]null ល្អ/[ល្អ]NN",
		myAssertDisambiguate("ភាព ល្អ", xmlDisam),
		"xml TAG_AS_NOUN ភាព ល្អ")
}

// TAG_AS_NOUN_END: VB|JJ|RB + ធម៌|និយម|ករ → marker becomes NN.
func TestKhmerRuleDisambiguator_TagAsNounEnd(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "សង្គ្រោះ ធម៌"
	require.Equal(t,
		"/[null]SENT_START សង្គ្រោះ/[សង្គ្រោះ]JJ|សង្គ្រោះ/[សង្គ្រោះ]NN|សង្គ្រោះ/[សង្គ្រោះ]VB  /[null]null ធម៌/[ធម៌]JJ|ធម៌/[ធម៌]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START សង្គ្រោះ/[សង្គ្រោះ]NN  /[null]null ធម៌/[ធម៌]JJ|ធម៌/[ធម៌]NN",
		myAssertDisambiguate(input, xmlDisam),
		"xml TAG_AS_NOUN_END សង្គ្រោះ ធម៌")
}

// USAGE_ALL: PRO + អស់ + NN → អស់ becomes RB.
func TestKhmerRuleDisambiguator_UsageAll(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ខ្ញុំ អស់ មនុស្ស"
	require.Equal(t,
		"/[null]SENT_START ខ្ញុំ/[ខ្ញុំ]PRO  /[null]null អស់/[អស់]IN|អស់/[អស់]JJ|អស់/[អស់]NN|អស់/[អស់]RB|អស់/[អស់]VB  /[null]null មនុស្ស/[មនុស្ស]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ខ្ញុំ/[ខ្ញុំ]PRO  /[null]null អស់/[អស់]RB  /[null]null មនុស្ស/[មនុស្ស]NN",
		myAssertDisambiguate(input, xmlDisam),
		"xml USAGE_ALL ខ្ញុំ អស់ មនុស្ស")
}

// DOWY_SA: ដោយ + សារ → marker សារ becomes CC.
func TestKhmerRuleDisambiguator_DowySa(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ដោយ សារ"
	require.Equal(t,
		"/[null]SENT_START ដោយ/[ដោយ]CC|ដោយ/[ដោយ]IN|ដោយ/[ដោយ]JJ|ដោយ/[ដោយ]RB|ដោយ/[ដោយ]VB  /[null]null សារ/[សារ]JJ|សារ/[សារ]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ដោយ/[ដោយ]CC|ដោយ/[ដោយ]IN|ដោយ/[ដោយ]JJ|ដោយ/[ដោយ]RB|ដោយ/[ដោយ]VB  /[null]null សារ/[សារ]CC",
		myAssertDisambiguate(input, xmlDisam),
		"xml DOWY_SA ដោយ សារ")
}

// TIME_AS_ADJ: JJ + ពេល + NN|VB|PRO → ពេល becomes JJ.
func TestKhmerRuleDisambiguator_TimeAsAdj(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ល្អ ពេល មនុស្ស"
	require.Equal(t,
		"/[null]SENT_START ល្អ/[ល្អ]JJ|ល្អ/[ល្អ]RB  /[null]null ពេល/[ពេល]JJ|ពេល/[ពេល]NN  /[null]null មនុស្ស/[មនុស្ស]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ល្អ/[ល្អ]JJ|ល្អ/[ល្អ]RB  /[null]null ពេល/[ពេល]JJ  /[null]null មនុស្ស/[មនុស្ស]NN",
		myAssertDisambiguate(input, xmlDisam),
		"xml TIME_AS_ADJ ល្អ ពេល មនុស្ស")
}

// TOBE_YOU_as_Noun + TOBE_YOU_as_Noun2: ជា + អ្នក + VB|NN|JJ → both markers → NN.
func TestKhmerRuleDisambiguator_TobeYou(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ជា អ្នក សង្គ្រោះ"
	require.Equal(t,
		"/[null]SENT_START ជា/[ជា]VB  /[null]null អ្នក/[អ្នក]NN|អ្នក/[អ្នក]PRO|អ្នក/[អ្នក]PRP  /[null]null សង្គ្រោះ/[សង្គ្រោះ]JJ|សង្គ្រោះ/[សង្គ្រោះ]NN|សង្គ្រោះ/[សង្គ្រោះ]VB",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ជា/[ជា]VB  /[null]null អ្នក/[អ្នក]NN  /[null]null សង្គ្រោះ/[សង្គ្រោះ]NN",
		myAssertDisambiguate(input, xmlDisam),
		"xml TOBE_YOU cascade ជា អ្នក សង្គ្រោះ")
}

// CLS_YOU_as_Noun + CLS_YOU_as_Noun2: CLS + អ្នក + VB|NN|JJ → both → NN.
// ក្រុម has CLS|NN readings so the CLS pattern matches.
func TestKhmerRuleDisambiguator_ClsYou(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ក្រុម អ្នក សិល្បៈ"
	require.Equal(t,
		"/[null]SENT_START ក្រុម/[ក្រុម]CLS|ក្រុម/[ក្រុម]NN  /[null]null អ្នក/[អ្នក]NN|អ្នក/[អ្នក]PRO|អ្នក/[អ្នក]PRP  /[null]null សិល្បៈ/[សិល្បៈ]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ក្រុម/[ក្រុម]CLS|ក្រុម/[ក្រុម]NN  /[null]null អ្នក/[អ្នក]NN  /[null]null សិល្បៈ/[សិល្បៈ]NN",
		myAssertDisambiguate(input, xmlDisam),
		"xml CLS_YOU ក្រុម អ្នក សិល្បៈ")
}

// NOUN_PTOAL: NN (not PRO|PRP) + ផ្ទាល់ + PRO → ផ្ទាល់ becomes PRO.
func TestKhmerRuleDisambiguator_NounPtoal(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "រឿង ផ្ទាល់ ខ្លួន"
	require.Equal(t,
		"/[null]SENT_START រឿង/[រឿង]JJ|រឿង/[រឿង]NN  /[null]null ផ្ទាល់/[ផ្ទាល់]AW|ផ្ទាល់/[ផ្ទាល់]IN|ផ្ទាល់/[ផ្ទាល់]JJ|ផ្ទាល់/[ផ្ទាល់]NN|ផ្ទាល់/[ផ្ទាល់]PRO|ផ្ទាល់/[ផ្ទាល់]RB|ផ្ទាល់/[ផ្ទាល់]VB  /[null]null ខ្លួន/[ខ្លួន]NN|ខ្លួន/[ខ្លួន]PRO",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START រឿង/[រឿង]JJ|រឿង/[រឿង]NN  /[null]null ផ្ទាល់/[ផ្ទាល់]PRO  /[null]null ខ្លួន/[ខ្លួន]NN|ខ្លួន/[ខ្លួន]PRO",
		myAssertDisambiguate(input, xmlDisam),
		"xml NOUN_PTOAL រឿង ផ្ទាល់ ខ្លួន")
}

// self_PTOAL{,2,3}: ខ្លួន + ឯង + ផ្ទាល់ → all three markers become PRO.
func TestKhmerRuleDisambiguator_SelfPtoal(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ខ្លួន ឯង ផ្ទាល់"
	require.Equal(t,
		"/[null]SENT_START ខ្លួន/[ខ្លួន]NN|ខ្លួន/[ខ្លួន]PRO  /[null]null ឯង/[ឯង]IN|ឯង/[ឯង]JJ|ឯង/[ឯង]NN|ឯង/[ឯង]PRP|ឯង/[ឯង]RB  /[null]null ផ្ទាល់/[ផ្ទាល់]AW|ផ្ទាល់/[ផ្ទាល់]IN|ផ្ទាល់/[ផ្ទាល់]JJ|ផ្ទាល់/[ផ្ទាល់]NN|ផ្ទាល់/[ផ្ទាល់]PRO|ផ្ទាល់/[ផ្ទាល់]RB|ផ្ទាល់/[ផ្ទាល់]VB",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ខ្លួន/[ខ្លួន]PRO  /[null]null ឯង/[ឯង]PRO  /[null]null ផ្ទាល់/[ផ្ទាល់]PRO",
		myAssertDisambiguate(input, xmlDisam),
		"xml self_PTOAL cascade ខ្លួន ឯង ផ្ទាល់")
}

// Ing-pronoun: PRO|NNP + ឯង → ឯង becomes PRO.
func TestKhmerRuleDisambiguator_IngPronoun(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ខ្លួន ឯង"
	require.Equal(t,
		"/[null]SENT_START ខ្លួន/[ខ្លួន]NN|ខ្លួន/[ខ្លួន]PRO  /[null]null ឯង/[ឯង]IN|ឯង/[ឯង]JJ|ឯង/[ឯង]NN|ឯង/[ឯង]PRP|ឯង/[ឯង]RB",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ខ្លួន/[ខ្លួន]NN|ខ្លួន/[ខ្លួន]PRO  /[null]null ឯង/[ឯង]PRO",
		myAssertDisambiguate(input, xmlDisam),
		"xml Ing-pronoun ខ្លួន ឯង")
}

// ptoal-pronoun: PRO|NNP + ផ្ទាល់ → ផ្ទាល់ becomes PRO.
func TestKhmerRuleDisambiguator_PtoalPronoun(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ខ្លួន ផ្ទាល់"
	require.Equal(t,
		"/[null]SENT_START ខ្លួន/[ខ្លួន]NN|ខ្លួន/[ខ្លួន]PRO  /[null]null ផ្ទាល់/[ផ្ទាល់]AW|ផ្ទាល់/[ផ្ទាល់]IN|ផ្ទាល់/[ផ្ទាល់]JJ|ផ្ទាល់/[ផ្ទាល់]NN|ផ្ទាល់/[ផ្ទាល់]PRO|ផ្ទាល់/[ផ្ទាល់]RB|ផ្ទាល់/[ផ្ទាល់]VB",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ខ្លួន/[ខ្លួន]NN|ខ្លួន/[ខ្លួន]PRO  /[null]null ផ្ទាល់/[ផ្ទាល់]PRO",
		myAssertDisambiguate(input, xmlDisam),
		"xml ptoal-pronoun ខ្លួន ផ្ទាល់")
}

// timoy: NN + តែ + មួយ → REPLACE first (no marker) to JJ.
func TestKhmerRuleDisambiguator_Timoy(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ព្រះ តែ មួយ"
	require.Equal(t,
		"/[null]SENT_START ព្រះ/[ព្រះ]NN|ព្រះ/[ព្រះ]ROY  /[null]null តែ/[តែ]AW|តែ/[តែ]CC|តែ/[តែ]IN|តែ/[តែ]NN|តែ/[តែ]RB|តែ/[តែ]VB  /[null]null មួយ/[មួយ]NUM",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ព្រះ/[ព្រះ]JJ  /[null]null តែ/[តែ]AW|តែ/[តែ]CC|តែ/[តែ]IN|តែ/[តែ]NN|តែ/[តែ]RB|តែ/[តែ]VB  /[null]null មួយ/[មួយ]NUM",
		myAssertDisambiguate(input, xmlDisam),
		"xml timoy ព្រះ តែ មួយ")
}

// TOBE_LONG: ជា + យូរ + NN → យូរ becomes RB.
func TestKhmerRuleDisambiguator_TobeLong(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ជា យូរ ឆ្នាំ"
	require.Equal(t,
		"/[null]SENT_START ជា/[ជា]VB  /[null]null យូរ/[យូរ]JJ|យូរ/[យូរ]RB  /[null]null ឆ្នាំ/[ឆ្នាំ]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ជា/[ជា]VB  /[null]null យូរ/[យូរ]RB  /[null]null ឆ្នាំ/[ឆ្នាំ]NN",
		myAssertDisambiguate(input, xmlDisam),
		"xml TOBE_LONG ជា យូរ ឆ្នាំ")
}

// ADJ_HEART: PRO + JJ + ចិត្ត → ចិត្ត becomes JJ.
func TestKhmerRuleDisambiguator_AdjHeart(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "អ្នក សប្បាយ ចិត្ត"
	require.Equal(t,
		"/[null]SENT_START អ្នក/[អ្នក]NN|អ្នក/[អ្នក]PRO|អ្នក/[អ្នក]PRP  /[null]null សប្បាយ/[សប្បាយ]JJ  /[null]null ចិត្ត/[ចិត្ត]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START អ្នក/[អ្នក]NN|អ្នក/[អ្នក]PRO|អ្នក/[អ្នក]PRP  /[null]null សប្បាយ/[សប្បាយ]JJ  /[null]null ចិត្ត/[ចិត្ត]JJ",
		myAssertDisambiguate(input, xmlDisam),
		"xml ADJ_HEART អ្នក សប្បាយ ចិត្ត")
}

// WITH_PREP: ជា + មួយ → REPLACE first to IN.
func TestKhmerRuleDisambiguator_WithPrep(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "ជា មួយ"
	require.Equal(t,
		"/[null]SENT_START ជា/[ជា]VB  /[null]null មួយ/[មួយ]NUM",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ជា/[ជា]IN  /[null]null មួយ/[មួយ]NUM",
		myAssertDisambiguate(input, xmlDisam),
		"xml WITH_PREP ជា មួយ")
}

// WITH_PREP2 + WITH_PREP cascade: នៅ ជា មួយ → នៅ becomes IN (PREP2), then ជា becomes IN (PREP).
func TestKhmerRuleDisambiguator_WithPrep2(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "នៅ ជា មួយ"
	require.Equal(t,
		"/[null]SENT_START នៅ/[នៅ]IN|នៅ/[នៅ]RB|នៅ/[នៅ]VB  /[null]null ជា/[ជា]VB  /[null]null មួយ/[មួយ]NUM",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START នៅ/[នៅ]IN  /[null]null ជា/[ជា]IN  /[null]null មួយ/[មួយ]NUM",
		myAssertDisambiguate(input, xmlDisam),
		"xml WITH_PREP2 cascade នៅ ជា មួយ")
}

// PAST_TENSE_VERB: បាន + VB → បាន becomes PAS.
func TestKhmerRuleDisambiguator_PastTenseVerb(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "បាន សង្គ្រោះ"
	require.Equal(t,
		"/[null]SENT_START បាន/[បាន]AUX  /[null]null សង្គ្រោះ/[សង្គ្រោះ]JJ|សង្គ្រោះ/[សង្គ្រោះ]NN|សង្គ្រោះ/[សង្គ្រោះ]VB",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START បាន/[បាន]PAS  /[null]null សង្គ្រោះ/[សង្គ្រោះ]JJ|សង្គ្រោះ/[សង្គ្រោះ]NN|សង្គ្រោះ/[សង្គ្រោះ]VB",
		myAssertDisambiguate(input, xmlDisam),
		"xml PAST_TENSE_VERB បាន សង្គ្រោះ")
}

// ADVERB_MARKER: យ៉ាង + JJ|RB → យ៉ាង becomes RB.
func TestKhmerRuleDisambiguator_AdverbMarker(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "យ៉ាង ទន់ភ្លន់"
	require.Equal(t,
		"/[null]SENT_START យ៉ាង/[យ៉ាង]NN  /[null]null ទន់ភ្លន់/[ទន់ភ្លន់]JJ",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START យ៉ាង/[យ៉ាង]RB  /[null]null ទន់ភ្លន់/[ទន់ភ្លន់]JJ",
		myAssertDisambiguate(input, xmlDisam),
		"xml ADVERB_MARKER យ៉ាង ទន់ភ្លន់")
}

// MON_PREP: មុន + ពេល → មុន becomes IN.
func TestKhmerRuleDisambiguator_MonPrep(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "មុន ពេល"
	require.Equal(t,
		"/[null]SENT_START មុន/[មុន]NN|មុន/[មុន]RB  /[null]null ពេល/[ពេល]JJ|ពេល/[ពេល]NN",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START មុន/[មុន]IN  /[null]null ពេល/[ពេល]JJ|ពេល/[ពេល]NN",
		myAssertDisambiguate(input, xmlDisam),
		"xml MON_PREP មុន ពេល")
}

// PAST_TENSE_VERB + VERB_PTOAL cascade: បាន + VB + PRO + ផ្ទាល់.
func TestKhmerRuleDisambiguator_PastAndVerbPtoal(t *testing.T) {
	demo, xmlDisam := setupKMDisambiguation(t)
	const input = "បាន ជា គេ ផ្ទាល់"
	require.Equal(t,
		"/[null]SENT_START បាន/[បាន]AUX  /[null]null ជា/[ជា]VB  /[null]null គេ/[គេ]NN|គេ/[គេ]PRO|គេ/[គេ]PRP  /[null]null ផ្ទាល់/[ផ្ទាល់]AW|ផ្ទាល់/[ផ្ទាល់]IN|ផ្ទាល់/[ផ្ទាល់]JJ|ផ្ទាល់/[ផ្ទាល់]NN|ផ្ទាល់/[ផ្ទាល់]PRO|ផ្ទាល់/[ផ្ទាល់]RB|ផ្ទាល់/[ផ្ទាល់]VB",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START បាន/[បាន]PAS  /[null]null ជា/[ជា]VB  /[null]null គេ/[គេ]NN|គេ/[គេ]PRO|គេ/[គេ]PRP  /[null]null ផ្ទាល់/[ផ្ទាល់]PRO",
		myAssertDisambiguate(input, xmlDisam),
		"xml PAST_TENSE_VERB+VERB_PTOAL បាន ជា គេ ផ្ទាល់")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// KhmerWordTokenizer, SRXSentenceTokenizer(Khmer), KhmerTagger, disambiguator).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagkm.EnsureDefaultKhmerTagger()
	tagger := tagkm.DefaultKhmerTagger
	wt := kmtok.NewKhmerWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("km")
	var out strings.Builder
	for _, sentence := range st.Tokenize(input) {
		tokens := wt.Tokenize(sentence)
		var noWS []string
		for _, tok := range tokens {
			if testToolsIsWord(tok) {
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
			if testToolsIsWord(tokenStr) {
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
		finalSentence := languagetool.NewAnalyzedSentence(tokenArray)
		if dis != nil {
			finalSentence = dis.Disambiguate(finalSentence)
		}
		out.WriteString(formatMyAssertSentence(finalSentence))
	}
	return out.String()
}

// testToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func testToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// formatMyAssertSentence ports TestTools.getAsStrings + join for one sentence.
func formatMyAssertSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				readings = append(readings, testToolsGetAsString(r))
			}
		}
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}

// testToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
func testToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
