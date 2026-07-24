package synthesis

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// SpellNumber tags used by BaseSynthesizer.
const (
	SpellNumberTag         = "_spell_number_"
	SpellNumberFeminineTag = "_spell_number_:feminine"
	SpellNumberRomanTag    = "_spell_number_:Roman"
)

// BaseSynthesizer ports the non-Morfologik surface of
// org.languagetool.synthesis.BaseSynthesizer — ManualSynthesizer-backed forms
// plus Soros number/Roman spelling.
type BaseSynthesizer struct {
	LangShortCode    string
	ResourceFileName string
	TagFileName      string
	// SorFileName optional SOR resource (Java BaseSynthesizer first ctor arg, e.g. /en/en.sor).
	SorFileName string
	Manual      *ManualSynthesizer
	Removal     *ManualSynthesizer
	// Lookup is optional binary-dict synthesis (lemma+pos → forms).
	Lookup func(lemma, posTag string) []string
	// PossibleTags lists known POS tags when loaded.
	PossibleTags []string
	// NumberSpeller ports Soros from lang.sor (Java createNumberSpeller).
	NumberSpeller *Soros
	// RomanNumberer ports Soros from Roman.sor (Java createRomanNumberer).
	RomanNumberer *Soros
	// IsExceptionFn ports protected BaseSynthesizer.isException (Java virtual dispatch).
	// Nil → always false (Base default). English/French set this for subclass rules.
	IsExceptionFn func(w string) bool
}

func NewBaseSynthesizer(langShortCode string, manual *ManualSynthesizer) *BaseSynthesizer {
	return &BaseSynthesizer{LangShortCode: langShortCode, Manual: manual}
}

// SetNumberSpellerFromSource compiles lang.sor program (Java createNumberSpeller).
func (s *BaseSynthesizer) SetNumberSpellerFromSource(source, lang string) {
	if s == nil {
		return
	}
	if lang == "" {
		lang = s.LangShortCode
	}
	s.NumberSpeller = NewSoros(source, lang)
}

// SetRomanNumbererFromSource compiles Roman.sor program.
func (s *BaseSynthesizer) SetRomanNumbererFromSource(source string) {
	if s == nil {
		return
	}
	s.RomanNumberer = NewSoros(source, "Roman")
}

// LoadNumberSpellersFromDir loads {lang}.sor (or SorFileName basename) and Roman.sor
// from resourceDir / walk-up inspiration paths. Fail-closed (nil) when missing.
func (s *BaseSynthesizer) LoadNumberSpellersFromDir(resourceDir string) {
	if s == nil {
		return
	}
	lang := s.LangShortCode
	if lang == "" {
		lang = "en"
	}
	// Language SOR
	var sorPath string
	if s.SorFileName != "" {
		base := filepath.Base(s.SorFileName)
		if resourceDir != "" {
			cand := filepath.Join(resourceDir, base)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				sorPath = cand
			}
		}
	}
	if sorPath == "" && resourceDir != "" {
		for _, name := range []string{lang + ".sor", filepath.Base(s.SorFileName)} {
			if name == "" || name == "." {
				continue
			}
			cand := filepath.Join(resourceDir, name)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				sorPath = cand
				break
			}
		}
	}
	if sorPath == "" {
		// walk-up inspiration module
		rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", lang,
			"src", "main", "resources", "org", "languagetool", "resource", lang, lang+".sor")
		sorPath = walkUpFile(rel)
		if sorPath == "" && s.SorFileName != "" {
			// e.g. /en/en.sor → en/en.sor under resource
			clean := strings.TrimPrefix(s.SorFileName, "/")
			rel = filepath.Join("inspiration", "languagetool", "languagetool-language-modules", lang,
				"src", "main", "resources", "org", "languagetool", "resource", clean)
			sorPath = walkUpFile(rel)
		}
	}
	if sorPath != "" {
		if b, err := os.ReadFile(sorPath); err == nil {
			s.SetNumberSpellerFromSource(string(b), lang)
		}
	}
	// Roman.sor from core resources
	roman := walkUpFile(filepath.Join("inspiration", "languagetool", "languagetool-core",
		"src", "main", "resources", "org", "languagetool", "resource", "Roman.sor"))
	if roman == "" && resourceDir != "" {
		cand := filepath.Join(resourceDir, "Roman.sor")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			roman = cand
		}
	}
	if roman != "" {
		if b, err := os.ReadFile(roman); err == nil {
			s.SetRomanNumbererFromSource(string(b))
		}
	}
}

func walkUpFile(rel string) string {
	dir, _ := os.Getwd()
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// GetSpelledNumber ports BaseSynthesizer.getSpelledNumber.
func (s *BaseSynthesizer) GetSpelledNumber(arabicNumeral string) string {
	if s != nil && s.NumberSpeller != nil {
		return s.NumberSpeller.Run(arabicNumeral)
	}
	return arabicNumeral
}

// GetRomanNumber ports BaseSynthesizer.getRomanNumber.
func (s *BaseSynthesizer) GetRomanNumber(arabicNumeral string) string {
	if s != nil && s.RomanNumberer != nil {
		return s.RomanNumberer.Run(arabicNumeral)
	}
	return arabicNumeral
}

// IsException ports BaseSynthesizer.isException (default false; subclasses via IsExceptionFn).
func (s *BaseSynthesizer) IsException(w string) bool {
	if s != nil && s.IsExceptionFn != nil {
		return s.IsExceptionFn(w)
	}
	return false
}

// RemoveExceptions ports BaseSynthesizer.removeExceptions.
func (s *BaseSynthesizer) RemoveExceptions(words []string) []string {
	if len(words) == 0 {
		return words
	}
	out := make([]string, 0, len(words))
	for _, w := range words {
		if !s.IsException(w) {
			out = append(out, w)
		}
	}
	return out
}

// Synthesize ports BaseSynthesizer.synthesize for exact POS tags including spell-number tags.
func (s *BaseSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	// Java: spell tags use token.getToken() (surface), not lemma; no removeExceptions.
	switch posTag {
	case SpellNumberTag:
		return []string{s.GetSpelledNumber(token.GetToken())}, nil
	case SpellNumberFeminineTag:
		return []string{s.GetSpelledNumber("feminine " + token.GetToken())}, nil
	case SpellNumberRomanTag:
		return []string{s.GetRomanNumber(token.GetToken())}, nil
	}
	lemma := ""
	if token.GetLemma() != nil {
		lemma = *token.GetLemma()
	}
	if lemma == "" {
		lemma = token.GetToken()
	}
	return s.RemoveExceptions(collectForms(s, lemma, []string{posTag})), nil
}

// SynthesizeRE ports synthesize with optional POS regexp.
// Spell-number tags only apply on the non-regexp path (Java BaseSynthesizer).
// Java non-regexp path: return removeExceptions(synthesize(...)) — spell tags already
// returned without a second filter; regular forms filter once in Synthesize.
func (s *BaseSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	if !posTagRegExp {
		// Java: return removeExceptions(synthesize(token, posTag));
		// Synthesize already removes exceptions for non-spell tags; spell tags skip filter.
		// Applying RemoveExceptions again is a no-op for exceptions already gone and for spell forms.
		forms, err := s.Synthesize(token, posTag)
		if err != nil {
			return nil, err
		}
		return s.RemoveExceptions(forms), nil
	}
	lemma := ""
	if token.GetLemma() != nil {
		lemma = *token.GetLemma()
	}
	if lemma == "" {
		lemma = token.GetToken()
	}
	re, err := regexp.Compile("^(?:" + posTag + ")$")
	if err != nil {
		return nil, err
	}
	return s.SynthesizeForPosTags(lemma, re.MatchString), nil
}

// SynthesizeForPosTags ports BaseSynthesizer.synthesizeForPosTags (Java ≥5.3).
func (s *BaseSynthesizer) SynthesizeForPosTags(lemma string, acceptTag func(string) bool) []string {
	if s == nil || lemma == "" || acceptTag == nil {
		return nil
	}
	var tags []string
	for _, tag := range s.allTags() {
		if acceptTag(tag) {
			tags = append(tags, tag)
		}
	}
	return s.RemoveExceptions(collectForms(s, lemma, tags))
}

func collectForms(s *BaseSynthesizer, lemma string, tags []string) []string {
	if s == nil || lemma == "" {
		return nil
	}
	var forms []string
	seen := map[string]struct{}{}
	for _, tag := range tags {
		for _, f := range s.lookupForms(lemma, tag) {
			if s.isRemoved(lemma, tag, f) {
				continue
			}
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			forms = append(forms, f)
		}
	}
	return forms
}

func (s *BaseSynthesizer) lookupForms(lemma, posTag string) []string {
	var out []string
	if s.Lookup != nil {
		out = append(out, s.Lookup(lemma, posTag)...)
	}
	if s.Manual != nil {
		if v := s.Manual.Lookup(lemma, posTag); len(v) > 0 {
			out = append(out, v...)
		}
	}
	return out
}

func (s *BaseSynthesizer) isRemoved(lemma, posTag, form string) bool {
	if s.Removal == nil {
		return false
	}
	for _, f := range s.Removal.Lookup(lemma, posTag) {
		if f == form {
			return true
		}
	}
	return false
}

func (s *BaseSynthesizer) allTags() []string {
	if len(s.PossibleTags) > 0 {
		return s.PossibleTags
	}
	if s.Manual != nil {
		var tags []string
		for t := range s.Manual.GetPossibleTags() {
			tags = append(tags, t)
		}
		return tags
	}
	return nil
}

// GetTargetPosTag ports BaseSynthesizer.getTargetPosTag (last tag when non-empty list).
func (s *BaseSynthesizer) GetTargetPosTag(posTags []string, posTag string) string {
	if len(posTags) == 0 {
		return posTag
	}
	// Java: return the last one to keep the previous results
	return posTags[len(posTags)-1]
}

// GetPosTagCorrection ports BaseSynthesizer.getPosTagCorrection (identity).
func (s *BaseSynthesizer) GetPosTagCorrection(posTag string) string {
	return posTag
}

var _ Synthesizer = (*BaseSynthesizer)(nil)
