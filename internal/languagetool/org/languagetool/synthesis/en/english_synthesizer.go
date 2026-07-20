package en

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	EnglishSynthResource = "/en/english_synth.dict"
	EnglishTagsFile      = "/en/english_tags.txt"
	EnglishSorFile       = "/en/en.sor"

	// Special synthesizer tags (Java EnglishSynthesizer.ADD_*).
	AddDeterminer    = "+DT"
	AddIndDeterminer = "+INDT"
)

// Java EnglishSynthesizer.exceptions (ne'er, e'er, …).
var englishSynthExceptions = map[string]struct{}{
	"ne'er": {}, "e'er": {}, "o'er": {}, "ol'": {}, "ma'am": {}, "n't": {}, "informations": {},
}

// DefaultSuggestAorAn is set by package rules/en to AvsAnRule.SuggestAorAn
// (avoids synthesis/en → rules/en import cycle). When nil, SuggestAorAn fails
// closed to the bare word (no soft phonetic invent).
var DefaultSuggestAorAn func(word string) string

// EnglishSynthesizer ports org.languagetool.synthesis.en.EnglishSynthesizer.
type EnglishSynthesizer struct {
	*synthesis.BaseSynthesizer
	// SuggestAorAn ports AvsAnRule.suggestAorAn ("a word" / "an word" / word).
	// Nil → DefaultSuggestAorAn; both nil → bare word (fail-closed).
	SuggestAorAn func(word string) string
}

func NewEnglishSynthesizer(manual *synthesis.ManualSynthesizer) *EnglishSynthesizer {
	base := synthesis.NewBaseSynthesizer("en", manual)
	// Java: super(SOR_FILE_NAME, RESOURCE_FILENAME, TAGS_FILE_NAME, "en")
	base.SorFileName = EnglishSorFile
	base.ResourceFileName = EnglishSynthResource
	base.TagFileName = EnglishTagsFile
	return &EnglishSynthesizer{BaseSynthesizer: base}
}

// INSTANCE ports EnglishSynthesizer.INSTANCE.
var INSTANCE = NewEnglishSynthesizer(nil)

func (s *EnglishSynthesizer) suggestAorAn(word string) string {
	if s != nil && s.SuggestAorAn != nil {
		return s.SuggestAorAn(word)
	}
	if DefaultSuggestAorAn != nil {
		return DefaultSuggestAorAn(word)
	}
	// Fail closed without AvsAnRule wiring (no soft invent lexicon).
	return word
}

// IsException ports EnglishSynthesizer.isException: leading apostrophe or exceptions list.
func (s *EnglishSynthesizer) IsException(w string) bool {
	if strings.HasPrefix(w, "'") {
		return true
	}
	_, ok := englishSynthExceptions[w]
	return ok
}

func (s *EnglishSynthesizer) removeExceptions(words []string) []string {
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

// Synthesize ports EnglishSynthesizer.synthesize(token, posTag).
func (s *EnglishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	if strings.HasPrefix(posTag, synthesis.SpellNumberTag) {
		return s.BaseSynthesizer.Synthesize(token, posTag)
	}
	// Java uses token.getToken() (surface), not lemma, for suggestAorAn / "the …".
	surface := token.GetToken()
	aOrAn := s.suggestAorAn(surface)
	switch posTag {
	case AddDeterminer:
		// { aOrAn, "the " + lowercaseFirstCharIfCapitalized(token) }
		return []string{aOrAn, "the " + tools.LowercaseFirstCharIfCapitalized(surface)}, nil
	case AddIndDeterminer:
		return []string{aOrAn}, nil
	default:
		forms, err := s.BaseSynthesizer.Synthesize(token, posTag)
		if err != nil {
			return nil, err
		}
		return s.removeExceptions(forms), nil
	}
}

// SynthesizeRE ports EnglishSynthesizer.synthesize(token, posTag, posTagRegExp)
// including regexp tags ending with \\+INDT or \\+DT.
func (s *EnglishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	if strings.HasPrefix(posTag, synthesis.SpellNumberTag) {
		return s.Synthesize(token, posTag)
	}
	if !posTagRegExp {
		return s.removeExceptionsFromSynth(token, posTag)
	}
	myPosTag := posTag
	det := ""
	if strings.HasSuffix(posTag, AddIndDeterminer) {
		// Java: indexOf(+INDT) - "\\".length()  (pattern is …\\+INDT)
		idx := strings.Index(myPosTag, AddIndDeterminer)
		if idx >= 1 && myPosTag[idx-1] == '\\' {
			myPosTag = myPosTag[:idx-1]
		} else if idx >= 0 {
			myPosTag = myPosTag[:idx]
		}
		lemma := ""
		if token.GetLemma() != nil {
			lemma = *token.GetLemma()
		}
		full := s.suggestAorAn(lemma)
		// det = article + space only (substring to first space inclusive)
		if sp := strings.IndexByte(full, ' '); sp >= 0 {
			det = full[:sp+1]
		}
	} else if strings.HasSuffix(posTag, AddDeterminer) {
		idx := strings.Index(myPosTag, AddDeterminer)
		if idx >= 1 && myPosTag[idx-1] == '\\' {
			myPosTag = myPosTag[:idx-1]
		} else if idx >= 0 {
			myPosTag = myPosTag[:idx]
		}
		det = "the "
	}

	re, err := regexp.Compile("^(?:" + myPosTag + ")$")
	if err != nil {
		return nil, err
	}
	lemma := ""
	if token.GetLemma() != nil {
		lemma = *token.GetLemma()
	}
	if lemma == "" {
		return s.removeExceptions(nil), nil
	}
	var results []string
	for _, tag := range s.possibleTags() {
		if re.MatchString(tag) {
			for _, form := range s.lookupLemmaTag(lemma, tag) {
				results = append(results, det+form)
			}
		}
	}
	return s.removeExceptions(results), nil
}

func (s *EnglishSynthesizer) removeExceptionsFromSynth(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	forms, err := s.Synthesize(token, posTag)
	if err != nil {
		return nil, err
	}
	// Synthesize already removeExceptions for non-DT paths; DT paths have no exceptions.
	return forms, nil
}

func (s *EnglishSynthesizer) possibleTags() []string {
	if s == nil || s.BaseSynthesizer == nil {
		return nil
	}
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

func (s *EnglishSynthesizer) lookupLemmaTag(lemma, posTag string) []string {
	if s == nil || s.BaseSynthesizer == nil || lemma == "" {
		return nil
	}
	var out []string
	if s.Lookup != nil {
		out = append(out, s.Lookup(lemma, posTag)...)
	}
	if s.Manual != nil {
		out = append(out, s.Manual.Lookup(lemma, posTag)...)
	}
	if s.Removal != nil {
		filtered := out[:0]
		for _, f := range out {
			removed := false
			for _, r := range s.Removal.Lookup(lemma, posTag) {
				if r == f {
					removed = true
					break
				}
			}
			if !removed {
				filtered = append(filtered, f)
			}
		}
		out = filtered
	}
	return out
}

var _ synthesis.Synthesizer = (*EnglishSynthesizer)(nil)
