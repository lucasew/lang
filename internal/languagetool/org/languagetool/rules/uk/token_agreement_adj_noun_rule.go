package uk

import (
	"regexp"
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

const TokenAgreementAdjNounRuleID = "UK_ADJ_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementAdjNounRule ports org.languagetool.rules.uk.TokenAgreementAdjNounRule.
type TokenAgreementAdjNounRule struct {
	*tokenAgreementMatch
	// Synth optional (Java ukrainian.getSynthesizer()); nil → no suggestions.
	Synth synthesis.Synthesizer
}

func NewTokenAgreementAdjNounRule() *TokenAgreementAdjNounRule {
	return NewTokenAgreementAdjNounRuleWithMessages(nil)
}

// NewTokenAgreementAdjNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementAdjNounRuleWithMessages(messages map[string]string) *TokenAgreementAdjNounRule {
	r := &TokenAgreementAdjNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:       TokenAgreementAdjNounRuleID,
		description:  "Узгодження відмінків, роду і числа прикметника та іменника",
		shortMsg:     "Узгодження прикметника та іменника",
		isLeftToken:  HasAdjReading,
		isRightToken: HasNounReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			if IsPredicativeAdjException(left) || IsAdjpException(left) {
				return true
			}
			return AdjNounAgree(CollectPOSTags(left), CollectPOSTags(right))
		},
		exception: IsAdjNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

var (
	adjNounSkipLemmas = []string{"який", "котрий", "сам"}
	adjNounPodibnyi   = []string{"подібний"}
	adjNounDrugyi     = []string{"другий"}
	adjNounAdvSoft    = map[string]struct{}{
		"дуже": {}, "небагато": {}, "багато": {},
	}
	adjNounUYuyuRE   = regexp.MustCompile(`.*[ую]$`)
	adjNounNumDashRE = regexp.MustCompile(`.*([23]-є|[02-9]-а|[0-9]-м[иа])$`)
	adjNounDavNounRE = regexp.MustCompile(`noun.*?:m:v_dav.*`)
)

// Match ports TokenAgreementAdjNounRule.match state machine.
func (r *TokenAgreementAdjNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch

	adjPos := -1
	var adjTok *languagetool.AnalyzedTokenReadings
	var adjTags []string

	start := 1
	if len(tokens) > 0 && tokens[0] != nil && !tokens[0].IsSentenceStart() && firstPOS(tokens[0]) != "SENT_START" {
		start = 0
	}

	for i := start; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			adjPos = -1
			continue
		}
		if firstPOS(tok) == "" {
			adjPos = -1
			continue
		}

		// Java: while state non-empty, soft-skip pure adv / quant before noun under adjp
		if adjPos >= 0 && adjTok != nil {
			if shouldSkipAdvBeforeNoun(tokens, i, adjTags) {
				continue
			}
		}

		// grab adjective
		if HasPosTagStart(tok, "adj") {
			adjPos = -1
			adjTok = nil
			adjTags = nil

			// skip nv / який|котрий|сам / < tags
			if HasPosTagPart(tok, ":nv") ||
				HasLemmaTokenAny(tok, adjNounSkipLemmas) ||
				HasPosTagPart(tok, "<") {
				continue
			}
			// подібний :n: — Java breaks outer loop
			if HasLemmaWithPartPos(tok, adjNounPodibnyi, ":n:") {
				break
			}

			// collect adj readings; mixed POS may clear
			ok := true
			for _, rdg := range tok.GetReadings() {
				if rdg == nil || rdg.GetPOSTag() == nil {
					continue
				}
				pos := *rdg.GetPOSTag()
				if strings.HasPrefix(pos, "adj") {
					adjPos = i
					adjTok = tok
					adjTags = append(adjTags, pos)
					continue
				}
				// Java: !hasLemma(другий, adj:f:) || (next && !FAKE_FEM) && !predict
				// → non-adj reading usually clears unless special "другий" path
				if !HasLemmaWithPartPos(tok, adjNounDrugyi, "adj:f:") {
					ok = false
					break
				}
				nextOK := i+1 < len(tokens) && tokens[i+1] != nil &&
					!HasLemmaWithPartPos(tokens[i+1], FakeFemList, "noun:inanim:m:")
				if nextOK && !isPredictOrInsertPOS(pos) {
					ok = false
					break
				}
			}
			if !ok {
				adjPos = -1
				adjTok = nil
				adjTags = nil
			}
			continue
		}

		if adjPos < 0 || adjTok == nil {
			continue
		}

		// noun-side hard resets: :nv or pron on candidate
		if HasPosTagPart(tok, ":nv") || HasPosTagPart(tok, "pron") {
			adjPos = -1
			continue
		}

		// collect noun readings
		var nounTags []string
		clear := false
		for _, rdg := range tok.GetReadings() {
			if rdg == nil || rdg.GetPOSTag() == nil {
				continue
			}
			pos := *rdg.GetPOSTag()
			if strings.HasPrefix(pos, "noun") {
				nounTags = append(nounTags, pos)
			} else if pos == "SENT_END" || pos == "PARA_END" {
				continue
			} else if !isPredictOrInsertPOS(pos) {
				clear = true
				break
			}
		}
		if clear || len(nounTags) == 0 {
			adjPos = -1
			continue
		}

		master := GetAdjCaseInflections(adjTags)
		slave := GetNounInflectionsFromTags(nounTags, nounVZnaVarIgnore)
		if !InflectionsIntersect(master, slave) {
			if IsAdjNounException(tokens, adjPos, i) {
				adjPos = -1
				continue
			}
			// Java: "… \"%s\": [%s] і \"%s\": [%s]"
			msg := "Потенційна помилка: прикметник не узгоджений з іменником: \"" +
				adjTok.GetToken() + "\": [" + formatAdjNounInflections(master, true) + "] і \"" +
				tok.GetToken() + "\": [" + formatAdjNounInflections(slave, false) + "]"
			// Java message enrichments (else-if chain)
			if HasPosTagPartInTags(adjTags, ":m:v_rod") &&
				adjNounUYuyuRE.MatchString(tok.GetToken()) &&
				HasPosTagRE(tok, adjNounDavNounRE) {
				if UsedUInsteadOfAMsg != "" {
					msg += UsedUInsteadOfAMsg
				}
			} else if strings.Contains(adjTok.GetToken(), "-") &&
				adjNounNumDashRE.MatchString(adjTok.GetToken()) {
				msg += ". Можливо, вжито зайве літерне нарощення після кількісного числівника?"
			} else if strings.HasPrefix(strings.ToLower(adjTok.GetToken()), "не") &&
				hasTagRE(nounTags, regexp.MustCompile(`noun.*?:v_oru.*`)) {
				msg += ". Можливо, тут «не» потрібно писати окремо?"
			} else if !hasTagRE(adjTags, regexp.MustCompile(`adj.*?v_mis.*`)) &&
				hasTagRE(nounTags, regexp.MustCompile(`noun.*?v_mis.*`)) {
				msg += ". Можливо, пропущено прийменник на/в/у...?"
			}
			m := rules.NewRuleMatch(r, sentence, adjTok.GetStartPos(), tok.GetEndPos(), msg)
			m.ShortMessage = r.shortMsg
			if sugs := r.adjNounSuggestions(master, slave, adjTok, tok); len(sugs) > 0 {
				m.SetSuggestedReplacements(sugs)
			}
			// Java: num dash message also adds "N M" surface
			if strings.Contains(msg, "кількісного числівника") {
				suggNum := regexp.MustCompile(`[-–]м[аи]$`).ReplaceAllString(cleanTokenSurface(adjTok), "") +
					" " + tok.GetToken()
				cur := m.GetSuggestedReplacements()
				if !containsStr(cur, suggNum) {
					m.SetSuggestedReplacements(append(cur, suggNum))
				}
			}
			out = append(out, m)
		}
		adjPos = -1
	}
	return out
}

// formatAdjNounInflections ports TokenAgreementAdjNounRule.formatInflections.
func formatAdjNounInflections(infs []Inflection, adj bool) string {
	if len(infs) == 0 {
		return ""
	}
	// sort by gender then case (Java Collections.sort)
	sorted := append([]Inflection(nil), infs...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].CompareTo(sorted[j]) < 0
	})
	// gender → case names (LinkedHashMap order ≈ first-seen gender)
	var order []string
	byGen := map[string][]string{}
	for _, inf := range sorted {
		caseStr := taguk.VidminokName(inf.Case)
		if adj && inf.AnimTag != "" {
			if inf.AnimTag == "anim" {
				caseStr += " (іст.)"
			} else {
				caseStr += " (неіст.)"
			}
		}
		if _, ok := byGen[inf.Gender]; !ok {
			order = append(order, inf.Gender)
		}
		byGen[inf.Gender] = append(byGen[inf.Gender], caseStr)
	}
	var parts []string
	for _, g := range order {
		parts = append(parts, taguk.GenderName(g)+": "+strings.Join(byGen[g], ", "))
	}
	return strings.Join(parts, ", ")
}

// shouldSkipAdvBeforeNoun ports the Java adjp+adv soft skip before the noun check.
func shouldSkipAdvBeforeNoun(tokens []*languagetool.AnalyzedTokenReadings, i int, adjTags []string) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return false
	}
	tok := tokens[i]
	clean := strings.ToLower(cleanTokenSurface(tok))
	_, soft := adjNounAdvSoft[clean]
	if !hasPosTagPartAll(tok, "adv") && !soft {
		return false
	}
	// exclude prep that still has case gov on next token
	if i < len(tokens)-1 && HasPosTagStart(tok, "prep") {
		cases := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(tok, "prep")
		var list []string
		for c := range cases {
			list = append(list, c)
		}
		if HasVidmPosTag(list, tokens[i+1]) {
			return false
		}
	}
	return HasPosTagPartInTags(adjTags, "adjp")
}

// HasPosTagPartInTags reports whether any tag string contains substr.
func HasPosTagPartInTags(tags []string, substr string) bool {
	for _, p := range tags {
		if strings.Contains(p, substr) {
			return true
		}
	}
	return false
}

// adjNounSuggestions ports TokenAgreementAdjNounRule suggestion synthesis loops.
// Requires Synth; returns nil when unset or on empty synthesis.
func (r *TokenAgreementAdjNounRule) adjNounSuggestions(
	master, slave []Inflection,
	adjTok, nounTok *languagetool.AnalyzedTokenReadings,
) []string {
	if r == nil || r.Synth == nil || adjTok == nil || nounTok == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	// noun forms matching adj gender/case
	for _, adjInf := range master {
		if adjInf.Case == "v_kly" {
			continue
		}
		genderTag := ":" + adjInf.Gender + ":"
		vidmTag := adjInf.Case
		if adjInf.Gender != "p" && !HasPosTagPart(nounTok, genderTag) {
			continue
		}
		for _, nr := range nounTok.GetReadings() {
			if nr == nil || nr.GetPOSTag() == nil {
				continue
			}
			old := *nr.GetPOSTag()
			if !strings.HasPrefix(old, "noun") {
				continue
			}
			if adjInf.animMatters() {
				if adjInf.AnimTag != "" && !strings.Contains(old, ":"+adjInf.AnimTag) {
					continue
				}
			}
			newTag := regexp.MustCompile(`:.:v_...`).ReplaceAllString(old, genderTag+vidmTag)
			forms, err := r.Synth.SynthesizeRE(nr, newTag, false)
			if err != nil {
				continue
			}
			for _, s := range forms {
				add(adjTok.GetToken() + " " + s)
			}
		}
	}
	// adj forms matching noun gender/case
	for _, nInf := range slave {
		genderTag := ":" + nInf.Gender + ":"
		vidmTag := nInf.Case
		if nInf.animMatters() && nInf.AnimTag != "" {
			vidmTag += ":r" + nInf.AnimTag
		}
		for _, ar := range adjTok.GetReadings() {
			if ar == nil || ar.GetPOSTag() == nil {
				continue
			}
			old := *ar.GetPOSTag()
			if !strings.HasPrefix(old, "adj") {
				continue
			}
			newTag := regexp.MustCompile(`:.:v_...(?::r(?:in)?anim)?`).ReplaceAllString(old, genderTag+vidmTag)
			forms, err := r.Synth.SynthesizeRE(ar, newTag, false)
			if err != nil {
				continue
			}
			for _, s := range forms {
				add(s + " " + nounTok.GetToken())
			}
		}
	}
	return out
}
