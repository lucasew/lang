package pt

import (
	"embed"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Official resources Java loads via
// JLanguageTool.getDataBroker().getFromRulesDirAsStream:
//
//	/pt/verbos_sem_acento_nomes_com_acento.txt
//	/pt/verbos_sem_acento_adj_com_acento.txt
//
//go:embed data/verbos_sem_acento_nomes_com_acento.txt data/verbos_sem_acento_adj_com_acento.txt
var accentuationCheckFS embed.FS

// Classpath paths (PortugueseAccentuationCheckRule static loaders).
const (
	PortugueseAccentuationNounsPath = "/pt/verbos_sem_acento_nomes_com_acento.txt"
	PortugueseAccentuationAdjsPath  = "/pt/verbos_sem_acento_adj_com_acento.txt"
)

// Message strings from PortugueseAccentuationCheckRule.match (Java literals).
const (
	portugueseAccentuationMessage = "Se é um nome ou um adjectivo, tem acento."
	portugueseAccentuationShort   = "Falta um acento"
)

// Patterns from PortugueseAccentuationCheckRule (Java Pattern.compile + Matcher.matches
// ⇒ full-string match). Anchors added once at compile time.
//
// Note: ARTIGO_O_MP / ARTIGO_O_FP keep Java's swapped forms (as|As for MP, os|Os for FP).
// PREPOSICAO_DE keeps the Java character class [a|o] (literal '|' included).
var (
	ptPreposicaoDE      = regexp.MustCompile(`^(?:de|d[a|o]s?)$`)
	ptArtigoOMS         = regexp.MustCompile(`^(?:o|O)$`)
	ptArtigoOFS         = regexp.MustCompile(`^(?:a|A)$`)
	ptArtigoOMP         = regexp.MustCompile(`^(?:as|As)$`)
	ptArtigoOFP         = regexp.MustCompile(`^(?:os|Os)$`)
	ptDeterminante      = regexp.MustCompile(`^D[^R].*$`)
	ptDeterminanteMS    = regexp.MustCompile(`^D[^R].[MC][SN].*$`)
	ptDeterminanteFS    = regexp.MustCompile(`^D[^R].[FC][SN].*$`)
	ptDeterminanteMP    = regexp.MustCompile(`^D[^R].[MC][PN].*$`)
	ptDeterminanteFP    = regexp.MustCompile(`^D[^R].[FC][PN].*$`)
	ptNomeMS            = regexp.MustCompile(`^NC[MC][SN].*$`)
	ptNomeFS            = regexp.MustCompile(`^NC[FC][SN].*$`)
	ptNomeMP            = regexp.MustCompile(`^NC[MC][PN].*$`)
	ptNomeFP            = regexp.MustCompile(`^NC[FC][PN].*$`)
	ptAdjetivoMS        = regexp.MustCompile(`^(?:A..[MC][SN].*|V.P..SM.?|PX.MS.*)$`)
	ptAdjetivoFS        = regexp.MustCompile(`^(?:A..[FC][SN].*|V.P..SF.?|PX.FS.*)$`)
	ptAdjetivoMP        = regexp.MustCompile(`^(?:A..[MC][PN].*|V.P..PM.?|PX.MP.*)$`)
	ptAdjetivoFP        = regexp.MustCompile(`^(?:A..[FC][PN].*|V.P..PF.?|PX.FP.*)$`)
	ptInfinitivo        = regexp.MustCompile(`^V.N.*$`)
	ptVerboConjugado    = regexp.MustCompile(`^(?:V.[^NGP].*|_GV_)$`)
	ptParticipioMS      = regexp.MustCompile(`^V.P.*SM.?$`)
	ptGrupoVerbal       = regexp.MustCompile(`^_GV_$`)
	ptVerbo3S           = regexp.MustCompile(`^V...3S..?$`)
	ptNotInPrevToken    = regexp.MustCompile(`^(?:V..*|PP.*|P0.*|V.P.*)$`)
	ptBeforeAdjectiveMS = regexp.MustCompile(`^(?:SPS00|D[^R].[MC][SN].*|V.[^NGP].*|PX.*)$`)
	ptBeforeAdjectiveFS = regexp.MustCompile(`^(?:SPS00|D[^R].[FC][SN].*|V.[^NGP].*|PX.*)$`)
	ptBeforeAdjectiveMP = regexp.MustCompile(`^(?:SPS00|D[^R].[MC][PN].*|V.[^NGP].*|PX.*)$`)
	ptBeforeAdjectiveFP = regexp.MustCompile(`^(?:SPS00|D[^R].[FC][PN].*|V.[^NGP].*|PX.*)$`)
	ptGN                = regexp.MustCompile(`^(?:.*_GN_.*|<?/?N[CP].*)$`)
	// Java: Pattern.CASE_INSENSITIVE | Pattern.UNICODE_CASE
	ptExcepcoesAntesDE = regexp.MustCompile(`(?i)^(?:forma|manera|por|costat)$`)
	ptPronomePessoal   = regexp.MustCompile(`^(?:P0.{6}|PP3CN000|PP3NN000|PP3CP000|PP3CSD00)$`)
)

var (
	relevantWordsOnce    sync.Once
	relevantWordsNoun    map[string]*languagetool.AnalyzedTokenReadings
	relevantWordsAdjOnce sync.Once
	relevantWordsAdj     map[string]*languagetool.AnalyzedTokenReadings
)

// PortugueseAccentuationCheckRule ports
// org.languagetool.rules.pt.PortugueseAccentuationCheckRule.
type PortugueseAccentuationCheckRule struct {
	// relevantWords / relevantWords2 twin Java static maps (verb→noun / verb→adj).
	relevantWords  map[string]*languagetool.AnalyzedTokenReadings
	relevantWords2 map[string]*languagetool.AnalyzedTokenReadings
	// DefaultOff ports setDefaultOff() in the Java constructor.
	DefaultOff bool
	// Category / IssueType twin constructor setCategory / setLocQualityIssueType.
	Category  *rules.Category
	IssueType rules.ITSIssueType
}

// NewPortugueseAccentuationCheckRule ports the Java constructor
// PortugueseAccentuationCheckRule(ResourceBundle messages).
// messages may be nil (category display name falls back to id).
func NewPortugueseAccentuationCheckRule(messages map[string]string) *PortugueseAccentuationCheckRule {
	return &PortugueseAccentuationCheckRule{
		relevantWords:  loadRelevantWordsNoun(),
		relevantWords2: loadRelevantWordsAdj(),
		DefaultOff:     true, // Java setDefaultOff()
		Category:       rules.CatConfusedWords.GetCategory(messages),
		IssueType:      rules.ITSMisspelling,
	}
}

func loadRelevantWordsNoun() map[string]*languagetool.AnalyzedTokenReadings {
	relevantWordsOnce.Do(func() {
		f, err := accentuationCheckFS.Open("data/verbos_sem_acento_nomes_com_acento.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := PortugueseAccentuationDataLoader{}.LoadWords(f, PortugueseAccentuationNounsPath)
		if err != nil {
			panic(err)
		}
		relevantWordsNoun = m
	})
	return relevantWordsNoun
}

func loadRelevantWordsAdj() map[string]*languagetool.AnalyzedTokenReadings {
	relevantWordsAdjOnce.Do(func() {
		f, err := accentuationCheckFS.Open("data/verbos_sem_acento_adj_com_acento.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := PortugueseAccentuationDataLoader{}.LoadWords(f, PortugueseAccentuationAdjsPath)
		if err != nil {
			panic(err)
		}
		relevantWordsAdj = m
	})
	return relevantWordsAdj
}

func (r *PortugueseAccentuationCheckRule) GetID() string { return "ACCENTUATION_CHECK_PT" }

func (r *PortugueseAccentuationCheckRule) GetDescription() string {
	// Java: "Confusão com acentos gráficos" (unicode escapes in source)
	return "Confusão com acentos gráficos"
}

func (r *PortugueseAccentuationCheckRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

func (r *PortugueseAccentuationCheckRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *PortugueseAccentuationCheckRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil {
		return ""
	}
	return r.IssueType
}

// Match ports PortugueseAccentuationCheckRule.match.
func (r *PortugueseAccentuationCheckRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		// ignoring token 0, i.e. SENT_START
		var token string
		if i == 1 {
			token = strings.ToLower(tokens[i].GetToken())
		} else {
			token = tokens[i].GetToken()
		}
		prevToken := tokens[i-1].GetToken()
		prevPrevToken := ""
		if i > 2 {
			prevPrevToken = tokens[i-2].GetToken()
		}
		nextToken := ""
		if i < len(tokens)-1 {
			nextToken = tokens[i+1].GetToken()
		}
		nextNextToken := ""
		if i < len(tokens)-2 {
			nextNextToken = tokens[i+2].GetToken()
		}
		if tools.IsEmptyStr(token) {
			continue
		}
		isRelevantWord := false
		isRelevantWord2 := false
		if _, ok := r.relevantWords[token]; ok {
			isRelevantWord = true
		}
		if _, ok := r.relevantWords2[token]; ok {
			isRelevantWord2 = true
		}
		if !isRelevantWord && !isRelevantWord2 {
			continue
		}

		// verbo precedido de pronome reflexo
		if matchPostagRegexpPT(tokens[i-1], ptPronomePessoal) && !startsWithHyphen(prevToken) {
			continue
		}

		var replacement string
		// Java creates matchers once per token; re-match on demand (stateless).
		mPreposicaoDE := ptPreposicaoDE.MatchString(nextToken)
		mExcepcoesDE := ptExcepcoesAntesDE.MatchString(nextNextToken)
		mArtigoOMS := ptArtigoOMS.MatchString(prevToken)
		mArtigoOFS := ptArtigoOFS.MatchString(prevToken)
		mArtigoOMP := ptArtigoOMP.MatchString(prevToken)
		mArtigoOFP := ptArtigoOFP.MatchString(prevToken)

		// VERB WITHOUT ACCENT → NOUN WITH ACCENT
		if isRelevantWord && !matchPostagRegexpPT(tokens[i], ptGN) {
			nounATR := r.relevantWords[token]
			// amb renuncies
			if tokens[i-1].HasPosTag("SPS00") && !tokens[i-1].HasPosTag("RG") &&
				!matchPostagRegexpPT(tokens[i-1], ptDeterminante) &&
				!matchPostagRegexpPT(tokens[i], ptInfinitivo) {
				replacement = nounATR.GetToken()
			} else if (matchPostagRegexpPT(tokens[i-1], ptDeterminanteMS) && matchPostagRegexpPT(nounATR, ptNomeMS)) ||
				(matchPostagRegexpPT(tokens[i-1], ptDeterminanteMP) && matchPostagRegexpPT(nounATR, ptNomeMP)) ||
				(matchPostagRegexpPT(tokens[i-1], ptDeterminanteFS) && matchPostagRegexpPT(nounATR, ptNomeFS)) ||
				(matchPostagRegexpPT(tokens[i-1], ptDeterminanteFP) && matchPostagRegexpPT(nounATR, ptNomeFP)) {
				// aquestes renuncies
				replacement = nounATR.GetToken()
			} else if i > 2 &&
				matchPostagRegexpPT(tokens[i-2], ptVerboConjugado) &&
				((matchPostagRegexpPT(tokens[i-1], ptDeterminanteMS) && matchPostagRegexpPT(nounATR, ptNomeMS)) ||
					(matchPostagRegexpPT(tokens[i-1], ptDeterminanteMP) && matchPostagRegexpPT(nounATR, ptNomeMP)) ||
					(matchPostagRegexpPT(tokens[i-1], ptDeterminanteFS) && matchPostagRegexpPT(nounATR, ptNomeFS)) ||
					(matchPostagRegexpPT(tokens[i-1], ptDeterminanteFP) && matchPostagRegexpPT(nounATR, ptNomeFP))) {
				// fumaré una faria (correct: fària)
				replacement = nounATR.GetToken()
			} else if i > 2 &&
				matchPostagRegexpPT(tokens[i-2], ptVerboConjugado) &&
				((mArtigoOMS && matchPostagRegexpPT(nounATR, ptNomeMS)) ||
					(mArtigoOMP && matchPostagRegexpPT(nounATR, ptNomeMP)) ||
					(mArtigoOFS && matchPostagRegexpPT(nounATR, ptNomeFS)) ||
					(mArtigoOFP && matchPostagRegexpPT(nounATR, ptNomeFP))) {
				// fem la copia (correct: còpia)
				replacement = nounATR.GetToken()
			} else if !matchPostagRegexpPT(tokens[i], ptParticipioMS) &&
				!matchPostagRegexpPT(tokens[i-1], ptNotInPrevToken) &&
				(i < len(tokens)-2) &&
				!matchPostagRegexpPT(tokens[i+2], ptInfinitivo) &&
				!mExcepcoesDE &&
				!tokens[i-1].HasPosTag("RG") {
				// circumstancies d'una altra classe
				// (mPreposicaoDE.matches() is commented out in Java)
				replacement = nounATR.GetToken()
			} else if ((mArtigoOMS && matchPostagRegexpPT(nounATR, ptNomeMS)) ||
				(mArtigoOFS && matchPostagRegexpPT(nounATR, ptNomeFS)) ||
				(mArtigoOMP && matchPostagRegexpPT(nounATR, ptNomeMP)) ||
				(mArtigoOFP && matchPostagRegexpPT(nounATR, ptNomeFP))) &&
				mPreposicaoDE {
				// la renuncia del president.
				replacement = nounATR.GetToken()
			} else if (i < len(tokens)-1) &&
				((matchPostagRegexpPT(nounATR, ptNomeMS) && matchPostagRegexpPT(tokens[i+1], ptAdjetivoMS)) ||
					(matchPostagRegexpPT(nounATR, ptNomeFS) && matchPostagRegexpPT(tokens[i+1], ptAdjetivoFS)) ||
					(matchPostagRegexpPT(nounATR, ptNomeMP) && matchPostagRegexpPT(tokens[i+1], ptAdjetivoMP)) ||
					(matchPostagRegexpPT(nounATR, ptNomeFP) && matchPostagRegexpPT(tokens[i+1], ptAdjetivoFP))) {
				// circunstancias extraordináries
				replacement = nounATR.GetToken()
			} else if (matchPostagRegexpPT(nounATR, ptNomeMS) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoMS) &&
				!matchPostagRegexpPT(tokens[i], ptVerbo3S) && !matchPostagRegexpPT(tokens[i], ptGrupoVerbal)) ||
				(matchPostagRegexpPT(nounATR, ptNomeFS) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoFS) &&
					!matchPostagRegexpPT(tokens[i], ptVerbo3S)) ||
				(matchPostagRegexpPT(nounATR, ptNomeMP) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoMP)) ||
				(matchPostagRegexpPT(nounATR, ptNomeFP) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoFP)) {
				// les seves contraries
				replacement = nounATR.GetToken()
			} else if nextToken == "que" && i > 2 &&
				((matchPostagRegexpPT(nounATR, ptNomeMS) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoMS) &&
					matchPostagRegexpPT(tokens[i-2], ptDeterminanteMS)) ||
					(matchPostagRegexpPT(nounATR, ptNomeFS) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoFS) &&
						matchPostagRegexpPT(tokens[i-2], ptDeterminanteFS)) ||
					(matchPostagRegexpPT(nounATR, ptNomeMP) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoMP) &&
						matchPostagRegexpPT(tokens[i-2], ptDeterminanteMP)) ||
					(matchPostagRegexpPT(nounATR, ptNomeFP) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoFP) &&
						matchPostagRegexpPT(tokens[i-2], ptDeterminanteFP))) {
				// uma nova formula que (fórmula)
				replacement = nounATR.GetToken()
			} else if nextToken == "que" &&
				((mArtigoOMS && matchPostagRegexpPT(nounATR, ptNomeMS)) ||
					(mArtigoOFS && matchPostagRegexpPT(nounATR, ptNomeFS)) ||
					(mArtigoOMP && matchPostagRegexpPT(nounATR, ptNomeMP)) ||
					(mArtigoOFP && matchPostagRegexpPT(nounATR, ptNomeFP))) {
				// les circumstancies que ens envolten
				replacement = nounATR.GetToken()
			}
			// de positiva influencia — Java uses plain if (not else if): may overwrite.
			if i > 2 &&
				tokens[i-2].HasPosTag("SPS00") && !tokens[i-2].HasPosTag("RG") &&
				((matchPostagRegexpPT(nounATR, ptNomeMS) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoMS)) ||
					(matchPostagRegexpPT(nounATR, ptNomeFS) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoFS)) ||
					(matchPostagRegexpPT(nounATR, ptNomeMP) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoMP)) ||
					(matchPostagRegexpPT(nounATR, ptNomeFP) && matchPostagRegexpPT(tokens[i-1], ptAdjetivoFP))) {
				replacement = nounATR.GetToken()
			}
		}

		// VERB WITHOUT ACCENT → ADJECTIVE WITH ACCENT
		if isRelevantWord2 && !matchPostagRegexpPT(tokens[i], ptGN) {
			adjATR := r.relevantWords2[token]
			// de maneira obvia, circumstancias extraordinarias.
			if (matchPostagRegexpPT(adjATR, ptAdjetivoMS) && matchPostagRegexpPT(tokens[i-1], ptNomeMS) &&
				!tokens[i-1].HasPosTag("_GN_FS") && matchPostagRegexpPT(tokens[i], ptVerboConjugado) &&
				!matchPostagRegexpPT(tokens[i], ptVerbo3S)) ||
				(matchPostagRegexpPT(adjATR, ptAdjetivoFS) && equalsIgnoreCase(prevPrevToken, "de") &&
					(prevToken == "maneira" || prevToken == "forma")) ||
				(matchPostagRegexpPT(adjATR, ptAdjetivoMP) && matchPostagRegexpPT(tokens[i-1], ptNomeMP)) ||
				(matchPostagRegexpPT(adjATR, ptAdjetivoFP) && matchPostagRegexpPT(tokens[i-1], ptNomeFP)) {
				replacement = adjATR.GetToken()
			} else if (i < len(tokens)-1) &&
				prevToken != "que" &&
				!matchPostagRegexpPT(tokens[i-1], ptNotInPrevToken) &&
				((matchPostagRegexpPT(adjATR, ptAdjetivoMS) && matchPostagRegexpPT(tokens[i+1], ptNomeMS) &&
					matchPostagRegexpPT(tokens[i-1], ptBeforeAdjectiveMS)) ||
					(matchPostagRegexpPT(adjATR, ptAdjetivoFS) && matchPostagRegexpPT(tokens[i+1], ptNomeFS) &&
						matchPostagRegexpPT(tokens[i-1], ptBeforeAdjectiveFS)) ||
					(matchPostagRegexpPT(adjATR, ptAdjetivoMP) && matchPostagRegexpPT(tokens[i+1], ptNomeMP) &&
						matchPostagRegexpPT(tokens[i-1], ptBeforeAdjectiveMP)) ||
					(matchPostagRegexpPT(adjATR, ptAdjetivoFP) && matchPostagRegexpPT(tokens[i+1], ptNomeFP) &&
						matchPostagRegexpPT(tokens[i-1], ptBeforeAdjectiveFP))) {
				// de continua disputa
				replacement = adjATR.GetToken()
			} else if (i < len(tokens)-1) &&
				((matchPostagRegexpPT(adjATR, ptAdjetivoMS) &&
					matchPostagRegexpPT(tokens[i+1], ptNomeMS) && mArtigoOMS) ||
					(matchPostagRegexpPT(adjATR, ptAdjetivoFS) &&
						matchPostagRegexpPT(tokens[i+1], ptNomeFS) && mArtigoOFS) ||
					(matchPostagRegexpPT(adjATR, ptAdjetivoMP) &&
						matchPostagRegexpPT(tokens[i+1], ptNomeMP) && mArtigoOMP) ||
					(matchPostagRegexpPT(adjATR, ptAdjetivoFP) &&
						matchPostagRegexpPT(tokens[i+1], ptNomeFP) && mArtigoOFP)) {
				// a magnifica conservação
				replacement = adjATR.GetToken()
			}
		}

		if replacement != "" {
			msg := portugueseAccentuationMessage
			ruleMatch := rules.NewRuleMatch(r, sentence,
				tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
			ruleMatch.SetShortMessage(portugueseAccentuationShort)
			// Java: setSuggestedReplacement(replacement) — map form, no case preserve.
			ruleMatch.SetSuggestedReplacement(replacement)
			ruleMatches = append(ruleMatches, ruleMatch)
		}
	}
	return ruleMatches
}

// matchPostagRegexpPT ports PortugueseAccentuationCheckRule.matchPostagRegexp:
// any non-null POS tag full-matches the pattern.
func matchPostagRegexpPT(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) bool {
	if aToken == nil || pattern == nil {
		return false
	}
	for _, analyzedToken := range aToken.GetReadings() {
		if analyzedToken == nil {
			continue
		}
		posTag := analyzedToken.GetPOSTag()
		if posTag != nil && pattern.MatchString(*posTag) {
			return true
		}
	}
	return false
}

func startsWithHyphen(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

func equalsIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
}
