package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const TokenAgreementAdjNounRuleID = "UK_ADJ_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementAdjNounRule ports org.languagetool.rules.uk.TokenAgreementAdjNounRule.
type TokenAgreementAdjNounRule struct {
	*tokenAgreementMatch
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
			msg := "Потенційна помилка: прикметник не узгоджений з іменником: \"" +
				adjTok.GetToken() + "\" і \"" + tok.GetToken() + "\""
			// Java message enrichments (no synthesizer suggestions yet)
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
				HasPosTagPartInTags(nounTags, "v_oru") {
				msg += ". Можливо, тут «не» потрібно писати окремо?"
			} else if !HasPosTagPartInTags(adjTags, "v_mis") &&
				HasPosTagPartInTags(nounTags, "v_mis") {
				msg += ". Можливо, пропущено прийменник на/в/у...?"
			}
			m := rules.NewRuleMatch(r, sentence, adjTok.GetStartPos(), tok.GetEndPos(), msg)
			m.ShortMessage = r.shortMsg
			out = append(out, m)
		}
		adjPos = -1
	}
	return out
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
