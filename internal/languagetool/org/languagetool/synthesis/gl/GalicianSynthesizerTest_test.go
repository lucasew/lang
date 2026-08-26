package gl

// Behavior matrix for org.languagetool.synthesis.gl.GalicianSynthesizer.
// Java has no GalicianSynthesizerTest; expected forms are derived from the official
// galician_synth.dict via the same Lookup path BaseSynthesizer uses (no hand-invented maps).

import (
	"regexp"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireGalicianInstance(t *testing.T) *GalicianSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		// Official galician_synth.dict ships in the language-module resource tree;
		// fail closed (not soft-skip) when the inspiration asset is missing.
		t.Fatal("galician_synth.dict / official GL synth resources not available (fail-closed)")
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
func dictForms(t *testing.T, synth *GalicianSynthesizer, lemma, posTag string) []string {
	t.Helper()
	require.NotNil(t, synth.Lookup)
	return synth.Lookup(lemma, posTag)
}

// dictFormsRE accumulates Lookup over PossibleTags matching the Java RE path
// (BaseSynthesizer.synthesizeForPosTags order: PossibleTags iteration).
func dictFormsRE(t *testing.T, synth *GalicianSynthesizer, lemma, posTagRE string) []string {
	t.Helper()
	forms, err := synth.SynthesizeRE(dummyToken(lemma), posTagRE, true)
	require.NoError(t, err)
	// Recompute via Lookup + PossibleTags to prove forms come from the dict path.
	// Java: Pattern.compile(posTag).matcher(tag).matches() → full match.
	pat, err := regexp.Compile("^(?:" + posTagRE + ")$")
	require.NoError(t, err)
	var fromDict []string
	for _, tag := range synth.PossibleTags {
		if !pat.MatchString(tag) {
			continue
		}
		fromDict = append(fromDict, dictForms(t, synth, lemma, tag)...)
	}
	require.Equal(t, arraysToString(fromDict), arraysToString(forms),
		"SynthesizeRE must equal Lookup over matching PossibleTags in tag-file order")
	return forms
}

// Behavior matrix: unknown tag → 0; exact single-form; regex multi-form (Arrays.toString style).
func TestGalicianSynthesizer_SynthesizeBehaviorMatrix(t *testing.T) {
	synth := requireGalicianInstance(t)

	// Unknown lemma/tag → 0 forms (same as ItalianSynthesizerTest blablabla case).
	forms, err := synth.Synthesize(dummyToken("blablabla"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// Empty/unknown tag on a known lemma → 0 forms.
	forms, err = synth.Synthesize(dummyToken("casa"), "blablabla")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))
	forms, err = synth.Synthesize(dummyToken("casa"), "")
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))

	// Exact tag, single-form noun (official dict: casa + NCFS000 → casa).
	wantCasa := dictForms(t, synth, "casa", "NCFS000")
	require.NotEmpty(t, wantCasa)
	forms, err = synth.Synthesize(dummyToken("casa"), "NCFS000")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantCasa), arraysToString(forms))

	// Exact tag, plural form for same lemma family (libro + NCMP000).
	wantLibros := dictForms(t, synth, "libro", "NCMP000")
	require.NotEmpty(t, wantLibros)
	forms, err = synth.Synthesize(dummyToken("libro"), "NCMP000")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantLibros), arraysToString(forms))

	// Exact verb infinitive + past (facer paradigm in official dict).
	wantInf := dictForms(t, synth, "facer", "VMN0000")
	require.NotEmpty(t, wantInf)
	forms, err = synth.Synthesize(dummyToken("facer"), "VMN0000")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantInf), arraysToString(forms))

	wantPast := dictForms(t, synth, "facer", "VMIS3S0")
	require.NotEmpty(t, wantPast)
	forms, err = synth.Synthesize(dummyToken("facer"), "VMIS3S0")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantPast), arraysToString(forms))

	// Exact adjective gender/number variants (bo paradigm).
	wantBoMS := dictForms(t, synth, "bo", "AQ0MS0")
	require.NotEmpty(t, wantBoMS)
	forms, err = synth.Synthesize(dummyToken("bo"), "AQ0MS0")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantBoMS), arraysToString(forms))
	wantBoFS := dictForms(t, synth, "bo", "AQ0FS0")
	require.NotEmpty(t, wantBoFS)
	forms, err = synth.Synthesize(dummyToken("bo"), "AQ0FS0")
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantBoFS), arraysToString(forms))

	// Regex multi-form: "AQ0.S0" for lemma "bo" → masculine + feminine singular in tag order.
	formsRE := dictFormsRE(t, synth, "bo", "AQ0.S0")
	require.GreaterOrEqual(t, len(formsRE), 2, "expected multi-form RE paradigm for bo + AQ0.S0")
	// Also assert RE path equals concatenating exact-tag synth results in PossibleTags order.
	pat, err := regexp.Compile("^(?:AQ0.S0)$")
	require.NoError(t, err)
	var reWant []string
	for _, tag := range synth.PossibleTags {
		if !pat.MatchString(tag) {
			continue
		}
		part, err := synth.Synthesize(dummyToken("bo"), tag)
		require.NoError(t, err)
		reWant = append(reWant, part...)
	}
	require.Equal(t, arraysToString(reWant), arraysToString(formsRE))

	// Regex for verb past tags on facer — multi-form when pattern hits several tags.
	forms, err = synth.SynthesizeRE(dummyToken("facer"), "VMIS3S0", true)
	require.NoError(t, err)
	require.Equal(t, arraysToString(wantPast), arraysToString(forms))

	// Broader RE: NCM.* for libro (singular + plural noun masculine common).
	formsRE = dictFormsRE(t, synth, "libro", "NCM.*")
	require.GreaterOrEqual(t, len(formsRE), 2, "expected multi-form RE for libro + NCM.*")
}

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME) without dict.
func TestGalicianSynthesizer_ResourcePaths(t *testing.T) {
	s := NewGalicianSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "gl", s.LangShortCode)
	require.Equal(t, "/gl/galician_synth.dict", ResourceFilename)
	require.Equal(t, "/gl/galician_tags.txt", TagsFileName)
}

// Open path fail-closed + successful load from discovered official dict.
func TestOpenGalicianSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenGalicianSynthesizerFromDir(""))
	require.Nil(t, OpenGalicianSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenGalicianSynthesizerFromDictPath(""))

	p := DiscoverGalicianSynthDict()
	require.NotEmpty(t, p, "official galician_synth.dict must be discoverable from inspiration tree")
	s := OpenGalicianSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "gl", s.LangShortCode)
	// First tags from official galician_tags.txt (Freeling-style).
	require.Equal(t, "AO0FP0", s.PossibleTags[0])
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestGalicianSynthesizer_Instance(t *testing.T) {
	s := Instance()
	require.NotNil(t, s, "galician_synth.dict / official GL synth resources must load (fail-closed)")
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
