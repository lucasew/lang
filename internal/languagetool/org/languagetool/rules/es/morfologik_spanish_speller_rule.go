package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	// MorfologikSpanishSpellerRuleID ports MorfologikSpanishSpellerRule.getId().
	MorfologikSpanishSpellerRuleID = "MORFOLOGIK_RULE_ES"
	// SpanishSpellerDict ports MorfologikSpanishSpellerRule.getFileName().
	// Java: "/es/es-ES.dict" (not /es/hunspell/es.dict).
	SpanishSpellerDict = "/es/es-ES.dict"
)

// Java VERB_INDSUBJ = Pattern.compile("V.[SI].*") — '.' is any char.
var spanishVerbIndSubj = regexp.MustCompile(`(?i)^V.[SI]`)

// MorfologikSpanishSpellerRule ports rules.es.MorfologikSpanishSpellerRule.
// orderSuggestions: REMOVE_FROM_SUGGESTIONS + prefix drops + pronoun reorder when TagPOS set.
type MorfologikSpanishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// TagPOS optional SpanishTagger.tag surface → POS tags (pronoun-split reorder).
	TagPOS func(word string) []string
}

func NewMorfologikSpanishSpellerRule() *MorfologikSpanishSpellerRule {
	r := &MorfologikSpanishSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikSpanishSpellerRuleID, "es", SpanishSpellerDict, nil),
	}
	// Java MorfologikSpanishSpellerRule ctor: setIgnoreTaggedWords().
	r.IgnoreTaggedWords = true
	// Java tokenizeNewWords() = false; getAdditionalSpellingFileNames:
	// "/es/"+CUSTOM, GLOBAL, "/es/multiwords.txt"
	if r.SpellingCheckRule != nil {
		r.DisableTokenizeNewWords = true
		r.GetAdditionalSpellingFileNamesFn = func() []string {
			return []string{
				"/es/" + spelling.CustomSpellingFile,
				spelling.GlobalSpellingFile,
				"/es/multiwords.txt",
			}
		}
		spelling.ReapplyDefaultSpellingWordLists(r.SpellingCheckRule)
	}
	// Java MorfologikSpellerRule.initSpeller + Spanish.prepareLineForSpeller on plain lists.
	r.InitSpellersFromGetters(language.SpanishPrepareLineForSpeller, nil)
	return r
}

// Match ports parent Match + Spanish orderSuggestions + additional top suggestions.
func (r *MorfologikSpanishSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurfaceES(m, sentence)
		sugs := m.GetSuggestedReplacements()
		if top := r.additionalTopSpanishSuggestions(word); len(top) > 0 {
			sugs = append(top, sugs...)
		}
		if len(sugs) == 0 {
			continue
		}
		m.SetSuggestedReplacements(r.orderSpanishSuggestions(sugs, word))
	}
	return base, nil
}

// additionalTopSpanishSuggestions ports getAdditionalTopSuggestionsString
// (camelCase join + digit split; digit arm needs TagPOS for isTagged).
func (r *MorfologikSpanishSpellerRule) additionalTopSpanishSuggestions(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	if parts := splitCamelCaseES(word); len(parts) > 1 {
		ok := true
		for _, p := range parts {
			if r.wordIsMisspelled(p) {
				ok = false
				break
			}
		}
		if ok {
			return []string{strings.Join(parts, " ")}
		}
	}
	if parts := splitDigitsAtEndES(word); len(parts) == 2 {
		p0 := parts[0]
		_, shortOK := spanishSplitDigitsAtEnd[strings.ToLower(p0)]
		if r.isTagged(p0) && (tokenizers.UTF16Len(p0) > 2 || shortOK) {
			return []string{strings.Join(parts, " ")}
		}
	}
	return nil
}

func (r *MorfologikSpanishSpellerRule) wordIsMisspelled(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if r.IsMisspelled != nil {
		return r.IsMisspelled(word)
	}
	if r.Speller != nil {
		return r.Speller.IsMisspelled(word)
	}
	return false
}

func (r *MorfologikSpanishSpellerRule) isTagged(word string) bool {
	if r == nil || r.TagPOS == nil {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if t != "" {
			return true
		}
	}
	return false
}

// splitCamelCaseES delegates to tools.SplitCamelCase (StringTools).
func splitCamelCaseES(word string) []string {
	if word == "" {
		return nil
	}
	return tools.SplitCamelCase(word)
}

// splitDigitsAtEndES delegates to tools.SplitDigitsAtEnd (StringTools).
func splitDigitsAtEndES(input string) []string {
	if input == "" {
		return nil
	}
	return tools.SplitDigitsAtEnd(input)
}

var spanishSplitDigitsAtEnd = map[string]struct{}{
	"en": {}, "de": {}, "del": {}, "al": {}, "a": {}, "y": {}, "o": {}, "con": {},
}

func matchSurfaceES(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	// Java RuleMatch FromPos/ToPos are UTF-16 code units (String.substring).
	return rules.UTF16Substring(sent.GetText(), m.GetFromPos(), m.GetToPos())
}


// orderSpanishSuggestions ports MorfologikSpanishSpellerRule.orderSuggestions.
func (r *MorfologikSpanishSpellerRule) orderSpanishSuggestions(suggestions []string, word string) []string {
	wordWithoutDiacritics := tools.RemoveDiacritics(strings.ToLower(word))
	var out []string
	for _, sug := range suggestions {
		low := strings.ToLower(sug)
		if _, bad := spanishRemoveFromSuggestions[low]; bad {
			continue
		}
		parts := strings.Split(low, " ")
		if len(parts) == 2 {
			if parts[1] == "s" {
				continue
			}
			if _, bad := spanishPrefixWithWhitespace[parts[0]]; bad {
				continue
			}
			// pronoun + verb: move near front when TagPOS says second is V.[SI].*
			if _, isPron := spanishPronombreInicial[parts[0]]; isPron && len(parts[1]) > 1 {
				if r != nil && r.matchesVerbIndSubj(parts[1]) {
					pos := diacriticFrontPos(out, wordWithoutDiacritics)
					out = insertAtES(out, pos, sug)
					continue
				}
			}
			if _, ok := spanishParticulaFinal[parts[1]]; ok {
				out = append([]string{sug}, out...)
				continue
			}
		}
		if tools.RemoveDiacritics(low) == wordWithoutDiacritics {
			pos := diacriticFrontPos(out, wordWithoutDiacritics)
			out = insertAtES(out, pos, sug)
			continue
		}
		out = append(out, sug)
	}
	return out
}

func (r *MorfologikSpanishSpellerRule) matchesVerbIndSubj(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	for _, tag := range r.TagPOS(word) {
		if spanishVerbIndSubj.MatchString(tag) {
			return true
		}
	}
	return false
}

func diacriticFrontPos(out []string, wordWithoutDiacritics string) int {
	pos := 0
	for pos < len(out) && tools.RemoveDiacritics(strings.ToLower(out[pos])) == wordWithoutDiacritics {
		pos++
	}
	return pos
}

func insertAtES(slice []string, i int, v string) []string {
	if i < 0 {
		i = 0
	}
	if i >= len(slice) {
		return append(slice, v)
	}
	out := make([]string, 0, len(slice)+1)
	out = append(out, slice[:i]...)
	out = append(out, v)
	out = append(out, slice[i:]...)
	return out
}

// orderSpanishSuggestions package-level for tests without TagPOS.
func orderSpanishSuggestions(suggestions []string, word string) []string {
	return (&MorfologikSpanishSpellerRule{}).orderSpanishSuggestions(suggestions, word)
}

// UseInOffice ports useInOffice() — force-enable in LO/OO extension.
func (r *MorfologikSpanishSpellerRule) UseInOffice() bool { return true }
