package pl

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java PolishSynthesizer.RESOURCE_FILENAME / TAGS_FILE_NAME and internal tags.
const (
	ResourceFilename     = "/pl/polish_synth.dict"
	TagsFileName         = "/pl/polish_tags.txt"
	potentialNegationTag = ":aff"
	negationTag          = ":neg"
	compTag              = "com"
	supTag               = "sup"
)

var (
	// Java PATTERN = Pattern.compile(".*[a-z]\\.[a-z].*")
	plPosDotSegment = regexp.MustCompile(`.*[a-z]\.[a-z].*`)
	// Java NEGATION_TAG_PATTERN = Pattern.compile(NEGATION_TAG)
	negationTagRE = regexp.MustCompile(negationTag)
)

// PolishSynthesizer ports org.languagetool.synthesis.pl.PolishSynthesizer.
type PolishSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewPolishSynthesizer ports private PolishSynthesizer() / deprecated Language ctor.
// Java: super(RESOURCE_FILENAME, TAGS_FILE_NAME, "pl")
func NewPolishSynthesizer(manual *synthesis.ManualSynthesizer) *PolishSynthesizer {
	base := synthesis.NewBaseSynthesizer("pl", manual)
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &PolishSynthesizer{BaseSynthesizer: base}
}

// Synthesize ports PolishSynthesizer.synthesize(token, posTag).
// Java null posTag → nil; empty string is not null in Java (Go uses "" for null).
func (s *PolishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if posTag == "" {
		// Java: if (posTag == null) return null;
		return nil, nil
	}
	if s == nil || s.BaseSynthesizer == nil || token == nil {
		return nil, nil
	}
	isNegated := isNegatedPL(token, posTag)
	// Java: if (posTag.indexOf('+') > 0) return synthesize(token, posTag, true);
	// indexOf > 0: '+' at position 0 does not force the regexp path.
	if strings.Index(posTag, "+") > 0 {
		return s.SynthesizeRE(token, posTag, true)
	}
	return s.getWordForms(token, posTag, isNegated), nil
}

// SynthesizeRE ports PolishSynthesizer.synthesize(token, pos, posTagRegExp).
func (s *PolishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, pos string, posTagRegExp bool) ([]string, error) {
	if pos == "" {
		// Java: if (pos == null) return null;
		return nil, nil
	}
	if s == nil || s.BaseSynthesizer == nil || token == nil {
		return nil, nil
	}
	if !posTagRegExp {
		return s.Synthesize(token, pos)
	}
	posTag := pos
	isNegated := isNegatedPL(token, posTag)
	if isNegated {
		// Java: posTag = posTag.replace(NEGATION_TAG, POTENTIAL_NEGATION_TAG + "?");
		posTag = strings.ReplaceAll(posTag, negationTag, potentialNegationTag+"?")
	}
	// Java: Pattern.compile(posTag.replace('+', '|')); Matcher.matches() on full tag.
	pattern := strings.ReplaceAll(posTag, "+", "|")
	re, err := regexp.Compile("^(?:" + pattern + ")$")
	if err != nil {
		// Java PatternSyntaxException: printStackTrace, return collected (empty) after HashSet.
		return []string{}, nil
	}
	// Java: Set<String> hs = new HashSet<>(results); results.clear(); results.addAll(hs);
	seen := map[string]struct{}{}
	var results []string
	for _, tag := range s.possibleTags() {
		if !re.MatchString(tag) {
			continue
		}
		for _, f := range s.getWordForms(token, tag, isNegated) {
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			results = append(results, f)
		}
	}
	return results, nil
}

// isNegatedPL ports Java isNegated calculation (operator precedence && before ||).
//
//	Only when token.getPOSTag() != null:
//	posTag.indexOf(":neg") > 0
//	  || token.getPOSTag().indexOf(":neg") > 0
//	     && !(posTag.indexOf("com") > 0)
//	     && !(posTag.indexOf("sup") > 0)
func isNegatedPL(token *languagetool.AnalyzedToken, posTag string) bool {
	if token == nil || token.GetPOSTag() == nil {
		// Java: isNegated stays false when token POS is null.
		return false
	}
	tokenPOS := *token.GetPOSTag()
	// Java indexOf > 0 (not found is -1; match at index 0 is false)
	a := strings.Index(posTag, negationTag) > 0
	b := strings.Index(tokenPOS, negationTag) > 0
	c := !(strings.Index(posTag, compTag) > 0)
	d := !(strings.Index(posTag, supTag) > 0)
	return a || (b && c && d)
}

// getWordForms ports PolishSynthesizer.getWordForms.
// Java uses DictionaryLookup(getDictionary()) only — not Base.lookup (no manual/removal).
// Negated: lookup lemma|:aff (neg→aff replaceFirst) and prefix "nie".
func (s *PolishSynthesizer) getWordForms(token *languagetool.AnalyzedToken, posTag string, isNegated bool) []string {
	lemma := ""
	if token != nil && token.GetLemma() != nil {
		lemma = *token.GetLemma()
	}
	if isNegated {
		// Java: NEGATION_TAG_PATTERN.matcher(posTag).replaceFirst(POTENTIAL_NEGATION_TAG)
		affTag := posTag
		if idx := negationTagRE.FindStringIndex(posTag); idx != nil {
			affTag = posTag[:idx[0]] + potentialNegationTag + posTag[idx[1]:]
		}
		var forms []string
		for _, stem := range s.lookupDictOnly(lemma, affTag) {
			forms = append(forms, "nie"+stem)
		}
		return forms
	}
	return s.lookupDictOnly(lemma, posTag)
}

// lookupDictOnly is Java DictionaryLookup(getDictionary()).lookup(lemma+"|"+posTag) stems.
func (s *PolishSynthesizer) lookupDictOnly(lemma, posTag string) []string {
	if s == nil || s.BaseSynthesizer == nil || s.Lookup == nil {
		return nil
	}
	return s.Lookup(lemma, posTag)
}

func (s *PolishSynthesizer) possibleTags() []string {
	if s == nil || s.BaseSynthesizer == nil {
		return nil
	}
	// Java: load possibleTags from TAGS_FILE_NAME once; Open path sets PossibleTags.
	if len(s.PossibleTags) > 0 {
		return s.PossibleTags
	}
	return nil
}

// GetPosTagCorrection ports PolishSynthesizer.getPosTagCorrection.
func (s *PolishSynthesizer) GetPosTagCorrection(posTag string) string {
	if !strings.Contains(posTag, ".") {
		return posTag
	}
	tags := strings.Split(posTag, ":")
	pos := -1
	for i, t := range tags {
		if plPosDotSegment.MatchString(t) {
			// Java: "(.*" + DOT.matcher(tags[i]).replaceAll(".*|.*") + ".*)"
			// DOT is Pattern.LITERAL "."
			tags[i] = "(.*" + strings.ReplaceAll(t, ".", ".*|.*") + ".*)"
			pos = i
		}
	}
	if pos == -1 {
		return posTag
	}
	return strings.Join(tags, ":")
}

var _ synthesis.Synthesizer = (*PolishSynthesizer)(nil)
