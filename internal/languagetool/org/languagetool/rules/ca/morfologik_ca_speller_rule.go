package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	// MorfologikCatalanSpellerRuleID ports MorfologikCatalanSpellerRule.getId().
	// Java returns "MORFOLOGIK_RULE_CA_ES" (not MORFOLOGIK_RULE_CA).
	MorfologikCatalanSpellerRuleID = "MORFOLOGIK_RULE_CA_ES"
	// CatalanSpellerDict ports MorfologikCatalanSpellerRule.getFileName().
	// Java: "/ca/ca-ES_spelling.dict" (not /ca/hunspell/ca.dict).
	CatalanSpellerDict = "/ca/ca-ES_spelling.dict"
)

// MorfologikCatalanSpellerRule ports rules.ca.MorfologikCatalanSpellerRule.
// orderSuggestions: string-level filters + diacritic reorder (lemma/POS arms need TagPOS —
// incomplete without CatalanTagger wiring; no invent).
type MorfologikCatalanSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// TagPOS optional CatalanTagger.tag surface → POS tags (lemma ignore / pronoun reorder).
	// Fail-closed when nil: POS-gated arms of orderSuggestions are skipped.
	TagPOS func(word string) []string
	// HasLemma optional lemma membership for LemmasToIgnore/Allow.
	// Fail-closed when nil: lemma-based drops are skipped.
	HasLemma func(word string, lemmas []string) bool
}

func NewMorfologikCatalanSpellerRule() *MorfologikCatalanSpellerRule {
	r := &MorfologikCatalanSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikCatalanSpellerRuleID, "ca", CatalanSpellerDict, nil),
	}
	// Java MorfologikCatalanSpellerRule ctor: setIgnoreTaggedWords().
	r.IgnoreTaggedWords = true
	// Java tokenizeNewWords() = false; getSpellingFileName → /ca/spelling.txt;
	// getAdditionalSpellingFileNames → /ca/+CUSTOM, GLOBAL, multiwords, spelling-special.
	if r.SpellingCheckRule != nil {
		r.DisableTokenizeNewWords = true
		r.GetSpellingFileNameFn = func() string { return "/ca/spelling.txt" }
		r.GetAdditionalSpellingFileNamesFn = func() []string {
			return []string{
				"/ca/" + spelling.CustomSpellingFile,
				spelling.GlobalSpellingFile,
				"/ca/multiwords.txt",
				"/ca/spelling-special.txt",
			}
		}
		spelling.ReapplyDefaultSpellingWordLists(r.SpellingCheckRule)
	}
	// Java MorfologikSpellerRule.initSpeller + Catalan.prepareLineForSpeller on plain lists.
	r.InitSpellersFromGetters(language.CatalanPrepareLineForSpeller, nil)
	return r
}

// Match ports parent Match + additionalTop + orderSuggestions.
func (r *MorfologikCatalanSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
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
		word := matchSurfaceCA(m, sentence)
		sugs := m.GetSuggestedReplacements()
		if top := r.additionalTopCatalanSuggestions(word); len(top) > 0 {
			sugs = append(top, sugs...)
		}
		if len(sugs) == 0 {
			continue
		}
		m.SetSuggestedReplacements(r.orderCatalanSuggestions(sugs, word))
	}
	return base, nil
}

// Java static lists from MorfologikCatalanSpellerRule.
var (
	caInalambric = map[string]struct{}{
		"inalàmbric": {}, "inalàmbrica": {}, "inalàmbrics": {},
		"inalàmriques": {}, "inalàmbricament": {}, "inalàmbricamente": {},
	}
	// Java: "inal" + "àmbriques" → "inalàmriques"
	caPrefixAmbEspai = map[string]struct{}{
		"pod": {}, "ultra": {}, "eco": {}, "tele": {}, "anti": {}, "re": {}, "des": {},
		"sen": {}, "sem": {}, "s": {}, "avant": {}, "auto": {}, "ex": {}, "extra": {},
		"macro": {}, "mega": {}, "meta": {}, "micro": {}, "multi": {}, "mono": {}, "mini": {},
		"post": {}, "retro": {}, "semi": {}, "super": {}, "trans": {}, "pro": {}, "g": {},
		"l": {}, "m": {}, "e": {}, "pos": {}, "acost": {},
	}
	caEspaiAmbSufixNo  = map[string]struct{}{"mi": {}, "lis": {}}
	caEspaiAmbSufixSi  = map[string]struct{}{"a": {}, "o": {}, "i": {}}
	caParticulaInicial = map[string]struct{}{
		"amb": {}, "sota": {}, "no": {}, "en": {}, "a": {}, "el": {}, "els": {}, "al": {}, "als": {},
		"pel": {}, "pels": {}, "del": {}, "dels": {}, "de": {}, "per": {}, "un": {}, "uns": {},
		"una": {}, "unes": {}, "la": {}, "les": {}, "teu": {}, "meu": {}, "seu": {}, "teus": {},
		"meus": {}, "seus": {},
	}
	caPronomInicial = map[string]struct{}{
		"em": {}, "et": {}, "es": {}, "se": {}, "ens": {}, "us": {}, "vos": {}, "li": {},
		"hi": {}, "ho": {}, "el": {}, "la": {}, "els": {}, "les": {},
	}
	caLemmasToIgnore = []string{"enterar", "sentar", "conseguir", "alcançar", "entimar", "pisar"}
	caLemmasToAllow  = []string{"enter", "sentir"}
)

// orderCatalanSuggestions ports MorfologikCatalanSpellerRule.orderSuggestions
// (string form). Weight-jump filter is N/A without SuggestedReplacement weights.
// Lemma ignore / preposition+verb / balear / pronoun POS reorder need HasLemma/TagPOS.
func (r *MorfologikCatalanSpellerRule) orderCatalanSuggestions(suggestions []string, word string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	wordWithoutDiacritics := tools.RemoveDiacritics(word)
	// membership for lowercase duplicate check (Java replacements list)
	set := make(map[string]struct{}, len(suggestions))
	for _, s := range suggestions {
		set[s] = struct{}{}
	}
	out := make([]string, 0, len(suggestions))
	for i, sug := range suggestions {
		// avoid duplicate capitalized replacement
		if word == strings.ToLower(word) && tools.IsCapitalizedWord(sug) {
			if _, ok := set[strings.ToLower(sug)]; ok {
				continue
			}
		}
		// inalambric → fixed wireless alternatives
		if _, ok := caInalambric[strings.ToLower(sug)]; ok {
			return []string{"sense fils", "sense fil", "sense cables", "autònom"}
		}
		// remove always
		if strings.EqualFold(sug, "como") {
			continue
		}
		// lemma ignore when HasLemma wired (Java tagger.hasAnyLemma)
		if r != nil && r.HasLemma != nil {
			parts := splitOnQuoteHyphenSpace(sug)
			ignore := false
			for _, p := range parts {
				if r.HasLemma(p, caLemmasToIgnore) && !r.HasLemma(p, caLemmasToAllow) {
					ignore = true
					break
				}
			}
			if ignore {
				continue
			}
		}
		// l'_ : remove superfluous space
		if strings.Contains(sug, "' ") {
			sug = strings.ReplaceAll(sug, "' ", "'")
		}
		parts := strings.Split(sug, " ")
		if len(parts) == 2 {
			if strings.EqualFold(parts[1], "s") {
				continue
			}
			if _, bad := caPrefixAmbEspai[strings.ToLower(parts[0])]; bad {
				continue
			}
			if _, bad := caEspaiAmbSufixNo[strings.ToLower(parts[1])]; bad {
				continue
			}
			if tokenizers.UTF16Len(parts[1]) == 1 {
				if _, ok := caEspaiAmbSufixSi[strings.ToLower(parts[1])]; !ok {
					continue
				}
			}
			// preposition + inflected verb drop needs TagPOS (incomplete without it)
			// participle / balear / pronoun reorder needs TagPOS
			if tokenizers.UTF16Len(parts[1]) > 1 {
				if _, ok := caParticulaInicial[strings.ToLower(parts[0])]; ok {
					pos := diacriticFrontPosCA(out, wordWithoutDiacritics)
					out = insertAtCA(out, pos, sug)
					continue
				}
				if _, ok := caPronomInicial[strings.ToLower(parts[0])]; ok {
					// Java only when second is VERB_INDSUBJ; without TagPOS skip reorder
					// (do not invent POS). Fall through to default append unless TagPOS matches.
					if r != nil && r.matchesVerbIndSubjCA(parts[1]) {
						pos := diacriticFrontPosCA(out, wordWithoutDiacritics)
						out = insertAtCA(out, pos, sug)
						continue
					}
				}
			}
		}
		// diacritic-only match → front
		if strings.EqualFold(tools.RemoveDiacritics(sug), wordWithoutDiacritics) {
			pos := diacriticFrontPosCA(out, wordWithoutDiacritics)
			out = insertAtCA(out, pos, sug)
			continue
		}
		// move words with apostrophe or hyphen to second position when clean matches word
		clean := stripQuoteOrHyphen(sug)
		if i > 1 && len(suggestions) > 2 && strings.EqualFold(clean, word) {
			pos := diacriticFrontPosCA(out, wordWithoutDiacritics)
			if pos == 0 {
				pos = 1
			}
			out = insertAtCA(out, pos, sug)
			continue
		}
		// move "queda'n" case: when i==1 and first ends with 'n/'t and no hyphen
		if i == 1 && len(out) > 0 {
			first := out[0]
			if !strings.Contains(first, "-") && (strings.HasSuffix(first, "'n") || strings.HasSuffix(first, "'t")) {
				out = insertAtCA(out, 0, sug)
				continue
			}
		}
		out = append(out, sug)
	}
	return out
}

func (r *MorfologikCatalanSpellerRule) matchesVerbIndSubjCA(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	// Java VERB_INDSUBJ = Pattern.compile("V.[SI].*")
	for _, tag := range r.TagPOS(word) {
		if len(tag) >= 3 && tag[0] == 'V' && (tag[2] == 'S' || tag[2] == 'I') {
			return true
		}
		// case-insensitive fallback
		up := strings.ToUpper(tag)
		if len(up) >= 3 && up[0] == 'V' && (up[2] == 'S' || up[2] == 'I') {
			return true
		}
	}
	return false
}

func splitOnQuoteHyphenSpace(s string) []string {
	// Java: replacement.split("[ '-]")
	var parts []string
	var cur strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\'' || r == '-' {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

func stripQuoteOrHyphen(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\'' || r == '-' || r == '’' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func diacriticFrontPosCA(out []string, wordWithoutDiacritics string) int {
	pos := 0
	for pos < len(out) && strings.EqualFold(tools.RemoveDiacritics(out[pos]), wordWithoutDiacritics) {
		pos++
	}
	return pos
}

func insertAtCA(slice []string, i int, v string) []string {
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

func matchSurfaceCA(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	text := sent.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	if from < 0 || from >= to {
		return ""
	}
	runes := []rune(text)
	if to <= len(runes) {
		return string(runes[from:to])
	}
	if to <= len(text) {
		return text[from:to]
	}
	return ""
}

// UseInOffice ports useInOffice() — force-enable in LO/OO extension.
func (r *MorfologikCatalanSpellerRule) UseInOffice() bool { return true }
