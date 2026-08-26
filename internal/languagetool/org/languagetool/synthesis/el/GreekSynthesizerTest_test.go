package el

// Behavior matrix for org.languagetool.synthesis.el.GreekSynthesizer.
// Java has no GreekSynthesizerTest; expected forms are derived from the official
// greek_synth.dict via the same Lookup path BaseSynthesizer uses (no hand-invented maps).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireGreekInstance(t *testing.T) *GreekSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		// Official greek_synth.dict ships in the language-module resource tree;
		// fail closed (not soft-skip) when the inspiration asset is missing.
		t.Fatal("greek_synth.dict / official EL synth resources not available (fail-closed)")
	}
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Same(t, s, INSTANCE)
	return s
}

func dummyToken(tokenStr string) *languagetool.AnalyzedToken {
	// Java: new AnalyzedToken(tokenStr, tokenStr, tokenStr)
	tag := tokenStr
	lem := tokenStr
	return languagetool.NewAnalyzedToken(tokenStr, &tag, &lem)
}

// arraysToString ports Java Arrays.toString(String[]) for twin assertions.
func arraysToString(forms []string) string {
	if forms == nil {
		return "[]"
	}
	return "[" + strings.Join(forms, ", ") + "]"
}

// dictForms is the official-dict ground truth for lemma+exact tag (BaseSynthesizer.Lookup).
func dictForms(t *testing.T, synth *GreekSynthesizer, lemma, posTag string) []string {
	t.Helper()
	require.NotNil(t, synth.Lookup)
	return synth.Lookup(lemma, posTag)
}

// dictFormsRE accumulates Lookup over PossibleTags matching the Java RE path
// (BaseSynthesizer.synthesizeForPosTags order: PossibleTags iteration).
func dictFormsRE(t *testing.T, synth *GreekSynthesizer, lemma, posTagRE string) []string {
	t.Helper()
	forms, err := synth.SynthesizeRE(dummyToken(lemma), posTagRE, true)
	require.NoError(t, err)
	// Also recompute via Lookup + PossibleTags to prove forms come from the dict path.
	var fromDict []string
	for _, tag := range synth.PossibleTags {
		// Java: Pattern.compile(posTag).matcher(tag).matches() with full match via ^(?:...)$
		// SynthesizeRE already does that; here we only collect Lookup forms for tags that
		// contributed — use SynthesizeRE result as outcome and require non-empty for known cases.
		_ = tag
		fromDict = append(fromDict, dictForms(t, synth, lemma, tag)...)
	}
	// For ".*" every PossibleTag contributes; require outcome equals full dict union in tag order.
	if posTagRE == ".*" {
		require.Equal(t, arraysToString(fromDict), arraysToString(forms),
			"SynthesizeRE must equal Lookup over PossibleTags in tag-file order")
	}
	return forms
}

// Behavior matrix: unknown tag → 0; exact multi-form; regex multi-form (Arrays.toString style).
func TestGreekSynthesizer_SynthesizeBehaviorMatrix(t *testing.T) {
	synth := requireGreekInstance(t)

	// Unknown lemma/tag → 0 forms (same as ItalianSynthesizerTest blablabla case).
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// Empty/unknown tag on a known lemma → 0 forms.
	forms, err = synth.Synthesize(dummyToken("έχω"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))
	forms, err = synth.Synthesize(dummyToken("έχω"), "")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// Exact tag, multi-form: lemma έχω + V lives in greek_synth.dict (source greek.txt).
	wantV := dictForms(t, synth, "έχω", "V")
	require.GreaterOrEqual(t, len(wantV), 2, "expected multi-form V paradigm for έχω in official dict")
	forms, err = synth.Synthesize(dummyToken("έχω"), "V")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantV), arraysToString(forms))

	// Exact tag, single-form particle (also in official dict).
	wantParticle := dictForms(t, synth, "θα", "PARTICLE")
	require.Equal(t, []string{"θα"}, wantParticle)
	forms, err = synth.Synthesize(dummyToken("θα"), "PARTICLE")
	require.NoError(t, err)
	require.Equal(t, "[θα]", arraysToString(forms))

	// Exact INF for πάω → single form from dict.
	wantINF := dictForms(t, synth, "πάω", "INF")
	require.NotEmpty(t, wantINF)
	forms, err = synth.Synthesize(dummyToken("πάω"), "INF")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantINF), arraysToString(forms))

	// Regex tag: ".*" over PossibleTags (V, INF, PARTICLE) — multi-form, Arrays.toString style.
	// πάω has both V and INF entries → forms accumulate across matching tags (Java keeps dups).
	formsRE := dictFormsRE(t, synth, "πάω", ".*")
	require.GreaterOrEqual(t, len(formsRE), 2)
	// Also assert RE path equals concatenating exact-tag synth results in PossibleTags order.
	var reWant []string
	for _, tag := range synth.PossibleTags {
		part, err := synth.Synthesize(dummyToken("πάω"), tag)
		require.NoError(t, err)
		reWant = append(reWant, part...)
	}
	require.Equal(t, arraysToString(reWant), arraysToString(formsRE))

	// Regex V|INF for έχω (only V matches) — still multi-form paradigm.
	forms, err = synth.SynthesizeRE(dummyToken("έχω"), "V|INF", true)
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantV), arraysToString(forms))
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestGreekSynthesizer_ResourcePaths(t *testing.T) {
	s := NewGreekSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "el", s.LangShortCode)
	require.Equal(t, "/el/greek_synth.dict", ResourceFilename)
	require.Equal(t, "/el/greek_tags.txt", TagsFileName)
}

// Open path fail-closed + successful load from discovered official dict.
func TestOpenGreekSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenGreekSynthesizerFromDir(""))
	require.Nil(t, OpenGreekSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenGreekSynthesizerFromDictPath(""))

	p := DiscoverGreekSynthDict()
	require.NotEmpty(t, p, "official greek_synth.dict must be discoverable from inspiration tree")
	s := OpenGreekSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, []string{"V", "INF", "PARTICLE"}, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "el", s.LangShortCode)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestGreekSynthesizer_Instance(t *testing.T) {
	s := Instance()
	require.NotNil(t, s, "greek_synth.dict / official EL synth resources must load (fail-closed)")
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
