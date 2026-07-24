package pt

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortugueseAccentuationCheckRule ports
// org.languagetool.rules.pt.PortugueseAccentuationCheckRule (subset of heuristics).
// Maps are injectable; production can load via PortugueseAccentuationDataLoader.
type PortugueseAccentuationCheckRule struct {
	// VerbToNoun: unaccented verb form → accented noun readings
	VerbToNoun map[string]*languagetool.AnalyzedTokenReadings
	// VerbToAdj: unaccented verb form → accented adjective readings
	VerbToAdj map[string]*languagetool.AnalyzedTokenReadings
	DefaultOff bool
}

func NewPortugueseAccentuationCheckRule() *PortugueseAccentuationCheckRule {
	return &PortugueseAccentuationCheckRule{
		VerbToNoun: map[string]*languagetool.AnalyzedTokenReadings{},
		VerbToAdj:  map[string]*languagetool.AnalyzedTokenReadings{},
		DefaultOff: true, // matches Java setDefaultOff
	}
}

func (r *PortugueseAccentuationCheckRule) GetID() string { return "ACCENTUATION_CHECK_PT" }

func (r *PortugueseAccentuationCheckRule) GetDescription() string {
	return "Confusão com acentos gráficos"
}

// Patterns for MatchesPosTagRegex (which wraps ^(?:...)$).
const (
	ptDetMS          = `D[^R].[MC][SN].*`
	ptDetFS          = `D[^R].[FC][SN].*`
	ptDetMP          = `D[^R].[MC][PN].*`
	ptDetFP          = `D[^R].[FC][PN].*`
	ptNomeMS         = `NC[MC][SN].*`
	ptNomeFS         = `NC[FC][SN].*`
	ptNomeMP         = `NC[MC][PN].*`
	ptNomeFP         = `NC[FC][PN].*`
	ptAdjMS          = `A..[MC][SN].*|V.P..SM.?|PX.MS.*`
	ptAdjFS          = `A..[FC][SN].*|V.P..SF.?|PX.FS.*`
	ptAdjMP          = `A..[MC][PN].*|V.P..PM.?|PX.MP.*`
	ptAdjFP          = `A..[FC][PN].*|V.P..PF.?|PX.FP.*`
	ptPronomePessoal = `P0.{6}|PP3CN000|PP3NN000|PP3CP000|PP3CSD00`
	ptSPS00          = `SPS00`
	ptInfinitivo     = `V.N.*`
)

var (
	ptPreposicaoDE = regexp.MustCompile(`(?i)^(?:de|d[ao]s?)$`)
	ptArtigoMS     = regexp.MustCompile(`(?i)^o$`)
	ptArtigoFS     = regexp.MustCompile(`(?i)^a$`)
	ptArtigoMP     = regexp.MustCompile(`(?i)^os$`)
	ptArtigoFP     = regexp.MustCompile(`(?i)^as$`)
)

func (r *PortugueseAccentuationCheckRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i].GetToken()
		if i == 1 {
			tok = strings.ToLower(tok)
		}
		if tools.IsEmptyStr(tok) {
			continue
		}
		nounATR, isNoun := r.VerbToNoun[tok]
		adjATR, isAdj := r.VerbToAdj[tok]
		if !isNoun && !isAdj {
			continue
		}
		prev := tokens[i-1]
		// skip reflexive pronoun before verb
		if prev.MatchesPosTagRegex(ptPronomePessoal) && !strings.HasPrefix(prev.GetToken(), "-") {
			continue
		}
		var replacement string
		// VERB → NOUN: det + matching gender/number noun reading
		if isNoun && nounATR != nil {
			if matchesDetNoun(prev, nounATR) {
				replacement = nounATR.GetToken()
			} else if prev.MatchesPosTagRegex(ptSPS00) && !prev.HasPosTag("RG") {
				// preposition before: amb renuncies style
				if !tokens[i].MatchesPosTagRegex(ptInfinitivo) {
					replacement = nounATR.GetToken()
				}
			} else if i+1 < len(tokens) && ptPreposicaoDE.MatchString(tokens[i+1].GetToken()) {
				// artigo + de …
				if ptArtigoMS.MatchString(prev.GetToken()) && nounATR.MatchesPosTagRegex(ptNomeMS) ||
					ptArtigoFS.MatchString(prev.GetToken()) && nounATR.MatchesPosTagRegex(ptNomeFS) ||
					ptArtigoMP.MatchString(prev.GetToken()) && nounATR.MatchesPosTagRegex(ptNomeMP) ||
					ptArtigoFP.MatchString(prev.GetToken()) && nounATR.MatchesPosTagRegex(ptNomeFP) {
					replacement = nounATR.GetToken()
				}
			}
		}
		// VERB → ADJ: det/article before adjective reading
		if replacement == "" && isAdj && adjATR != nil {
			if matchesDetAdj(prev, adjATR) {
				replacement = adjATR.GetToken()
			}
		}
		if replacement == "" || replacement == tokens[i].GetToken() {
			continue
		}
		// preserve case of surface token
		rep := tools.PreserveCase(replacement, tokens[i].GetToken())
		msg := "Possível confusão com acento gráfico: «" + tokens[i].GetToken() + "» → «" + rep + "»."
		rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
		rm.SetSuggestedReplacements([]string{rep})
		out = append(out, rm)
	}
	return out
}

func matchesDetNoun(prev *languagetool.AnalyzedTokenReadings, noun *languagetool.AnalyzedTokenReadings) bool {
	if prev.MatchesPosTagRegex(ptDetMS) && noun.MatchesPosTagRegex(ptNomeMS) {
		return true
	}
	if prev.MatchesPosTagRegex(ptDetMP) && noun.MatchesPosTagRegex(ptNomeMP) {
		return true
	}
	if prev.MatchesPosTagRegex(ptDetFS) && noun.MatchesPosTagRegex(ptNomeFS) {
		return true
	}
	if prev.MatchesPosTagRegex(ptDetFP) && noun.MatchesPosTagRegex(ptNomeFP) {
		return true
	}
	// surface articles
	t := prev.GetToken()
	if ptArtigoMS.MatchString(t) && noun.MatchesPosTagRegex(ptNomeMS) {
		return true
	}
	if ptArtigoFS.MatchString(t) && noun.MatchesPosTagRegex(ptNomeFS) {
		return true
	}
	if ptArtigoMP.MatchString(t) && noun.MatchesPosTagRegex(ptNomeMP) {
		return true
	}
	if ptArtigoFP.MatchString(t) && noun.MatchesPosTagRegex(ptNomeFP) {
		return true
	}
	return false
}

func matchesDetAdj(prev *languagetool.AnalyzedTokenReadings, adj *languagetool.AnalyzedTokenReadings) bool {
	if prev.MatchesPosTagRegex(ptDetMS) && adj.MatchesPosTagRegex(ptAdjMS) {
		return true
	}
	if prev.MatchesPosTagRegex(ptDetFS) && adj.MatchesPosTagRegex(ptAdjFS) {
		return true
	}
	if prev.MatchesPosTagRegex(ptDetMP) && adj.MatchesPosTagRegex(ptAdjMP) {
		return true
	}
	if prev.MatchesPosTagRegex(ptDetFP) && adj.MatchesPosTagRegex(ptAdjFP) {
		return true
	}
	return false
}
