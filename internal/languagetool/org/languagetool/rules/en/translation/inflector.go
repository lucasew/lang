package translation

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	ensynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Inflector ports org.languagetool.rules.en.translation.Inflector.
// Inflects English words according to a German POS tag.
//
// Java:
//
//	private final Synthesizer enSynth = Languages.getLanguageForShortCode("en").getSynthesizer();
type Inflector struct {
	enSynth synthesis.Synthesizer
}

var (
	enSynthOnce   sync.Once
	enSynthCached synthesis.Synthesizer
)

// NewInflector ports Inflector() — English synthesizer from Languages.getLanguageForShortCode("en").getSynthesizer()
// (English.createDefaultSynthesizer → EnglishSynthesizer.INSTANCE with official /en/english_synth.dict).
func NewInflector() *Inflector {
	return &Inflector{enSynth: defaultEnglishSynthesizer()}
}

// defaultEnglishSynthesizer ports Languages.getLanguageForShortCode("en").getSynthesizer().
// Loads the same official english_synth.dict resources Java INSTANCE uses. Nil when dict missing (fail-closed).
func defaultEnglishSynthesizer() synthesis.Synthesizer {
	enSynthOnce.Do(func() {
		if p := discoverEnglishSynthDict(); p != "" {
			enSynthCached = ensynth.OpenEnglishSynthesizerFromDictPath(p)
		}
	})
	return enSynthCached
}

// discoverEnglishSynthDict finds english_synth.dict (Java /en/english_synth.dict).
func discoverEnglishSynthDict() string {
	if p := os.Getenv("LANG_ENGLISH_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english_synth.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "english_synth.dict"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		for _, rel := range relPaths {
			cand := filepath.Join(dir, rel)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				return cand
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// Inflect ports Inflector.inflect(enToken, dePosTag).
// enToken is the base form of the English token (from BeoLingus).
func (inf *Inflector) Inflect(enToken, dePosTag string) []string {
	// Java: StringUtils.replaceOnce(enToken, "to ", "").split(" ")
	parts := strings.Split(strings.Replace(enToken, "to ", "", 1), " ")
	lastPartForms := inf.inflectSingleWord(parts[len(parts)-1], dePosTag)
	startParts := ""
	if len(parts) > 1 {
		// Java: String.join(" ", parts.subList(0, parts.size()-1))
		startParts = strings.Join(parts[:len(parts)-1], " ")
	}
	result := make([]string, 0, len(lastPartForms))
	for _, lastPartForm := range lastPartForms {
		// only the last part is inflected, e.g. "tire pump" -> "tire pumps"
		// Java: (startParts + " " + lastPartForm).trim()
		result = append(result, tools.JavaStringTrim(startParts+" "+lastPartForm))
	}
	return result
}

// inflectSingleWord ports Inflector.inflectSingleWord.
// dePosTag empty string maps Java null (Go has no null string); "" also falls through
// the same default as Java empty-string (no regex match → base form).
func (inf *Inflector) inflectSingleWord(enToken, dePosTag string) []string {
	forms := make([]string, 0)
	// Java: if (dePosTag == null) { forms.add(enToken); return forms; }
	// Go: empty string is the null stand-in for optional tags from callers;
	// Java "" is non-null and hits the default branch — same result (base form).
	if dePosTag == "" {
		forms = append(forms, enToken)
		return forms
	}
	// Java String.matches → full-string match
	switch {
	case matchRE(dePosTag, `SUB.*PLU.*`):
		forms = append(forms, inf.getForms(enToken, "NNP?S")...)
	case matchRE(dePosTag, `VER:3:SIN:PRÄ.*`):
		forms = append(forms, inf.getForms(enToken, "VBZ")...)
	case matchRE(dePosTag, `VER:3:SIN:PRT:.*`):
		forms = append(forms, inf.getForms(enToken, "VBD")...)
	case matchRE(dePosTag, `PA1:PRD:GRU:VER`):
		forms = append(forms, inf.getForms(enToken, "VBG")...)
	case matchRE(dePosTag, `PA2:PRD:GRU:VER`):
		forms = append(forms, inf.getForms(enToken, "VBN")...)
	case matchRE(dePosTag, `ADJ:PRD:KOM|ADJ:.*:KOM.*`):
		forms = append(forms, inf.getForms(enToken, "JJR")...)
	case matchRE(dePosTag, `ADJ:.*:SUP.*`):
		forms = append(forms, inf.getForms(enToken, "JJS")...)
	default:
		forms = append(forms, enToken)
	}
	return forms
}

// getForms ports Inflector.getForms — synthesize(AnalyzedToken(enToken,"fake-value",enToken), posTagRegex, true).
func (inf *Inflector) getForms(enToken, posTagRegex string) []string {
	if inf == nil || inf.enSynth == nil {
		// Java always has English synthesizer; fail-closed empty if dict not discoverable.
		return nil
	}
	// Java: new AnalyzedToken(enToken, "fake-value", enToken)
	pos := "fake-value"
	lemma := enToken
	tok := languagetool.NewAnalyzedToken(enToken, &pos, &lemma)
	// Java: enSynth.synthesize(..., posTagRegex, true)
	forms, err := inf.enSynth.SynthesizeRE(tok, posTagRegex, true)
	if err != nil {
		// Java: catch (IOException e) { throw new RuntimeException(e); }
		panic(err)
	}
	return forms
}

// matchRE ports Java String.matches (full-string match of the pattern).
func matchRE(s, re string) bool {
	ok, err := regexp.MatchString("^(?:"+re+")$", s)
	return err == nil && ok
}
