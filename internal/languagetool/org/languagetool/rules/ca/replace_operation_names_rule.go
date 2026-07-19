package ca

import (
	"embed"
	"regexp"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/replace_operationnames.txt
var opNamesFS embed.FS

var (
	opNamesOnce sync.Once
	opNamesMap  map[string][]string

	// Java ReplaceOperationNamesRule POS patterns (full-string match).
	opPrevTokenPOS      = regexp.MustCompile(`^(?:D[^R].*|PX.*|SPS00|SENT_START)$`)
	opPrevTokenPOSExcep = regexp.MustCompile(`^(?:RG_anteposat|N.*|CC|_PUNCT.*|_loc_unavegada|RN)$`)
	opNextTokenPOSExcep = regexp.MustCompile(`^(?:N.*)$`)
	opPuntuacio         = regexp.MustCompile(`^(?:PUNCT.*|SENT_START)$`)
	opDeterminant       = regexp.MustCompile(`^(?:D[^R].M.*)$`)

	opGenderNumberArgs = map[string]string{"lemmaSelect": "[NA].*"}
)

func loadOperationNames() map[string][]string {
	opNamesOnce.Do(func() {
		f, err := opNamesFS.Open("data/replace_operationnames.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		opNamesMap = m
	})
	return opNamesMap
}

// ReplaceOperationNamesRule ports org.languagetool.rules.ca.ReplaceOperationNamesRule.
// Match is POS-gated (prev det/prep/SENT_START, exceptions, next lemmas/_GV_); plural
// forms need Synthesize (NC.P.*). ConvertToGenderAndNumberFilter runs when Filter.Tag
// is set; without POS, mid-sentence matches fail closed (no surface invent).
type ReplaceOperationNamesRule struct {
	*rules.AbstractSimpleReplaceRule
	// Synthesize ports CatalanSynthesizer.synthesize(token, "NC.P.*") for plurals.
	// When nil, tokens ending in "s" produce no match (fail closed).
	Synthesize func(tok *languagetool.AnalyzedToken, postagRE string) []string
	// Filter optional ConvertToGenderAndNumberFilter (Java always applies for single-token).
	// When nil or Tag nil, surface replacements are kept (gender/number det expand incomplete).
	Filter *ConvertToGenderAndNumberFilter
}

func NewReplaceOperationNamesRule(messages map[string]string) *ReplaceOperationNamesRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadOperationNames(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "NOMS_OPERACIONS",
		Description:   "S'ha d'evitar com a nom d'operació tècnica: $match",
		ShortMsg:      "Forma preferible",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Si és el nom d'una operació tècnica, val més usar una altra forma."
		},
		Category: rules.NewCategory(rules.NewCategoryId("FORMES_SECUNDARIES"), "C8) Formes secundàries"),
	}
	return &ReplaceOperationNamesRule{AbstractSimpleReplaceRule: base}
}

// Match ports ReplaceOperationNamesRule.match (not AbstractSimpleReplace surface path).
func (r *ReplaceOperationNamesRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	wrong := loadOperationNames()
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch

loop:
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		token := strings.ToLower(tok.GetToken())
		lookup := token
		if len(lookup) > 3 && strings.HasSuffix(lookup, "s") {
			lookup = lookup[:len(lookup)-1]
		}
		replacementLemmas, ok := wrong[lookup]
		if !ok || len(replacementLemmas) == 0 {
			continue
		}

		// exceptions (Java surface + POS)
		if lookup == "duplicat" && strings.EqualFold(tokens[i-1].GetToken(), "per") {
			continue loop
		}
		if i+1 < len(tokens) && strings.EqualFold(lookup, "polit") &&
			tools.IsCapitalizedWord(tokens[i+1].GetToken()) {
			continue loop
		}
		// Assecat el braç del riu
		if i+1 < len(tokens) &&
			matchPostagRegexp(tokens[i-1], opPuntuacio) &&
			matchPostagRegexp(tokens[i+1], opDeterminant) {
			continue loop
		}

		// relevant token
		if tok.HasPosTag("_GV_") {
			continue loop
		}

		// next token
		if i+1 < len(tokens) {
			next := tokens[i+1]
			if next.HasLemma("per") || next.HasLemma("com") ||
				next.HasLemma("des") || next.HasLemma("amb") ||
				matchPostagRegexp(next, opNextTokenPOSExcep) {
				continue loop
			}
		}

		// prev token
		if !matchPostagRegexp(tokens[i-1], opPrevTokenPOS) ||
			matchPostagRegexp(tokens[i-1], opPrevTokenPOSExcep) {
			continue loop
		}

		// synthesize replacements
		var possibleReplacements []string
		if !strings.HasSuffix(token, "s") {
			possibleReplacements = append(possibleReplacements, replacementLemmas...)
		} else {
			// plural: needs synthesizer (fail closed without Synthesize)
			if r.Synthesize == nil {
				continue loop
			}
			for _, replacementLemma := range replacementLemmas {
				lemma := replacementLemma
				pos := "NCMS000"
				at := languagetool.NewAnalyzedToken(replacementLemma, &pos, &lemma)
				synthesized := r.Synthesize(at, "NC.P.*")
				possibleReplacements = append(possibleReplacements, synthesized...)
			}
		}
		if len(possibleReplacements) == 0 {
			continue loop
		}

		// createRuleMatch (Java AbstractSimpleReplaceRule.createRuleMatch without sub-id path)
		if !r.CaseSensitive && tools.StartsWithUppercase(tok.GetToken()) {
			for j, rep := range possibleReplacements {
				possibleReplacements[j] = tools.UppercaseFirstChar(rep)
			}
		}
		msg := "Si és el nom d'una operació tècnica, val més usar una altra forma."
		if r.MessageFn != nil {
			msg = r.MessageFn(tok.GetToken(), possibleReplacements)
		}
		pos := tok.GetStartPos()
		end := pos + utf16TokenLenOp(tok.GetToken())
		potential := rules.NewRuleMatch(r, sentence, pos, end, msg)
		potential.ShortMessage = r.ShortMsg
		if potential.ShortMessage == "" {
			potential.ShortMessage = "Forma preferible"
		}
		potential.SetSuggestedReplacements(possibleReplacements)
		ruleMatches = append(ruleMatches, potential)
	}

	// ConvertToGenderAndNumberFilter post-pass (Java)
	filter := r.Filter
	if filter == nil {
		return ruleMatches
	}
	// Without Tag the filter returns nil for suggestion-seed path — keep surface match.
	if filter.Tag == nil {
		return ruleMatches
	}
	var filtered []*rules.RuleMatch
	for _, potential := range ruleMatches {
		if potential == nil {
			continue
		}
		if !potential.IsUnderlinedErrorSingleToken() {
			filtered = append(filtered, potential)
			continue
		}
		finalMatch := filter.AcceptRuleMatch(potential, opGenderNumberArgs, 0, nil, nil)
		if finalMatch != nil {
			filtered = append(filtered, finalMatch)
		}
	}
	return filtered
}

// matchPostagRegexp ports ReplaceOperationNamesRule.matchPostagRegexp.
func matchPostagRegexp(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) bool {
	if aToken == nil || pattern == nil {
		return false
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if analyzedToken != nil && analyzedToken.GetPOSTag() != nil {
			posTag = *analyzedToken.GetPOSTag()
		}
		if pattern.MatchString(posTag) {
			return true
		}
	}
	return false
}

func utf16TokenLenOp(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
