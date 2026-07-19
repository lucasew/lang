package es

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// PostponedAdjectiveConcordanceFilter ports
// org.languagetool.rules.es.PostponedAdjectiveConcordanceFilter (1:1).
//
// Synthesize is the SpanishSynthesizer.synthesize(token, posTag, true) hook;
// when nil, suggestion synthesis yields empty → Accept returns nil (same as empty synth).
type PostponedAdjectiveConcordanceFilter struct {
	// maxLevels is Java maxLevels (4).
	maxLevels int
	// Synthesize ports SpanishSynthesizer.synthesize(AnalyzedToken, posTagRegex, true).
	Synthesize func(tok *languagetool.AnalyzedToken, posTagRegex string) []string

	// Java instance fields used during acceptRuleMatch.
	adverbAppeared      bool
	conjunctionAppeared bool
	punctuationAppeared bool
}

func NewPostponedAdjectiveConcordanceFilter() *PostponedAdjectiveConcordanceFilter {
	return &PostponedAdjectiveConcordanceFilter{maxLevels: 4}
}

// FreeLing POS patterns — Java Pattern.compile strings, unchanged (including GN_CS || quirk).
var (
	esNOM     = regexp.MustCompile(`N.*`)
	esNOMMS   = regexp.MustCompile(`N.MS.*|PI0MS000`)
	esNOMFS   = regexp.MustCompile(`N.FS.*|PI0FS000`)
	esNOMMP   = regexp.MustCompile(`N.MP.*`)
	esNOMMN   = regexp.MustCompile(`N.MN.*`)
	esNOMFP   = regexp.MustCompile(`N.FP.*`)
	esNOMCS   = regexp.MustCompile(`N.CS.*`)
	esNOMCP   = regexp.MustCompile(`N.CP.*`)
	esNOMDET  = regexp.MustCompile(`N.*|D[NDA0I].*|PI0[MF]S000`)
	esGN      = regexp.MustCompile(`_GN_.*`)
	esGNMS    = regexp.MustCompile(`_GN_MS`)
	esGNFS    = regexp.MustCompile(`_GN_FS`)
	esGNMP    = regexp.MustCompile(`_GN_MP`)
	esGNFP    = regexp.MustCompile(`_GN_FP`)
	esGNCS    = regexp.MustCompile(`_GN_[MF]S`)
	esGNCP    = regexp.MustCompile(`_GN_[MF]P`)
	esDET   = regexp.MustCompile(`D[NDA0IP].*`)
	esDETCS = regexp.MustCompile(`D[NDA0IP]0CS0`)
	esDETMS   = regexp.MustCompile(`D[NDA0IP]0MS0`)
	esDETFS   = regexp.MustCompile(`D[NDA0IP]0FS0`)
	esDETMP   = regexp.MustCompile(`D[NDA0IP]0MP0`)
	esDETFP   = regexp.MustCompile(`D[NDA0IP]0FP0`)
	esGNMSSub = regexp.MustCompile(`N.[MC][SN].*|A..[MC][SN].*|V.P..SM.?|PX.MS.*|D[NDA0I]0MS0|PI0MS000`)
	esGNFSSub = regexp.MustCompile(`N.[FC][SN].*|A..[FC][SN].*|V.P..SF.?|PX.FS.*|D[NDA0I]0FS0|PI0FS000`)
	esGNMPSub = regexp.MustCompile(`N.[MC][PN].*|A..[MC][PN].*|V.P..PM.?|PX.MP.*|D[NDA0I]0MP0`)
	esGNFPSub = regexp.MustCompile(`N.[FC][PN].*|A..[FC][PN].*|V.P..PF.?|PX.FP.*|D[NDA0I]0FP0`)
	esGNCPSub = regexp.MustCompile(`N.[FMC][PN].*|A..[FMC][PN].*|D[NDA0I]0[FM]P0`)
	// Java has double || before PI0 (empty alternative) — keep for bug-for-bug.
	esGNCSSub = regexp.MustCompile(`N.[FMC][SN].*|A..[FMC][SN].*|D[NDA0I]0[FM]S0||PI0[MFC]S000`)
	esADJECTIU   = regexp.MustCompile(`AQ.*|V.P.*|PX.*|.*LOC_ADJ.*`)
	esADJECTIUMS = regexp.MustCompile(`A..[MC][SN].*|V.P..SM.?|PX.MS.*`)
	esADJECTIUFS = regexp.MustCompile(`A..[FC][SN].*|V.P..SF.?|PX.FS.*`)
	esADJECTIUMP = regexp.MustCompile(`A..[MC][PN].*|V.P..PM.?|PX.MP.*`)
	esADJECTIUFP = regexp.MustCompile(`A..[FC][PN].*|V.P..PF.?|PX.FP.*`)
	esADJECTIUCP = regexp.MustCompile(`A..C[PN].*`)
	esADJECTIUCS = regexp.MustCompile(`A..C[SN].*`)
	esADJECTIUS  = regexp.MustCompile(`A...[SN].*|V.P..S..?|PX..S.*`)
	esADJECTIUP  = regexp.MustCompile(`A...[PN].*|V.P..P..?|PX..P.*`)
	esADVERBI    = regexp.MustCompile(`R.|.*LOC_ADV.*`)
	esCONJUNCIO  = regexp.MustCompile(`C.|.*LOC_CONJ.*`)
	esPUNTUACIO  = regexp.MustCompile(`_PUNCT`)
	esLOCADV     = regexp.MustCompile(`.*LOC_ADV.*`)
	esADVERBISOK = regexp.MustCompile(`RG_before`)
	esCOORDIONI  = regexp.MustCompile(`y|e|o|u|ni`)
	esKEEPCOUNT  = regexp.MustCompile(`A.*|N.*|D[NAIDP].*|SPS.*|SP:DA|.*LOC_ADV.*|V.P.*|_PUNCT.*|.*LOC_ADJ.*|PX.*|PI0.S000|UNKNOWN|V.N.{4}`)
	esKEEPCOUNT2 = regexp.MustCompile(`,|y|e|o|ni|u`)
	esSTOPCOUNT  = regexp.MustCompile(`;|lo`)
	esPREPOS     = regexp.MustCompile(`SP.*`)
	esPREPNIVEL  = regexp.MustCompile(`de|del|en|sobre|a|entre|por|con|sin|contra|para`)
	esVERB       = regexp.MustCompile(`V.[^P].*|_GV_`)
	esGV         = regexp.MustCompile(`_GV_`)
)

// AcceptRuleMatch ports PostponedAdjectiveConcordanceFilter.acceptRuleMatch.
func (f *PostponedAdjectiveConcordanceFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	maxLevels := f.maxLevels
	if maxLevels <= 0 {
		maxLevels = 4
	}

	addComma := strings.EqualFold(patterns.GetOptionalDefault("addComma", arguments, "false"), "true")
	tokens := match.Sentence.GetTokensWithoutWhitespace()
	i := patternTokenPos
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return nil
	}

	isPlural := true
	isPrevNoun := false
	var substPattern, gnPattern, adjPattern *regexp.Regexp
	canBeMS, canBeFS, canBeMP, canBeFP, canBeP := false, false, false, false, false

	cNt := make([]int, maxLevels)
	cNMS := make([]int, maxLevels)
	cNFS := make([]int, maxLevels)
	cNMP := make([]int, maxLevels)
	cNMN := make([]int, maxLevels)
	cNFP := make([]int, maxLevels)
	cNCS := make([]int, maxLevels)
	cNCP := make([]int, maxLevels)
	cDMS := make([]int, maxLevels)
	cDFS := make([]int, maxLevels)
	cDMP := make([]int, maxLevels)
	cDFP := make([]int, maxLevels)
	cN := make([]int, maxLevels)
	cD := make([]int, maxLevels)
	level := 0
	j := 1
	f.initializeApparitions()
	for i-j > 0 && f.keepCounting(tokens[i-j]) && level < maxLevels {
		if !isPrevNoun {
			if esMatchPostagRegexp(tokens[i-j], esNOM) ||
				(i-j-1 > 0 && !esMatchPostagRegexp(tokens[i-j], esNOM) && esMatchPostagRegexp(tokens[i-j], esADJECTIU) &&
					esMatchPostagRegexp(tokens[i-j-1], esDET)) {
				if esMatchPostagRegexp(tokens[i-j], esGNMS) {
					cNMS[level]++
					canBeMS = true
				}
				if esMatchPostagRegexp(tokens[i-j], esGNFS) {
					cNFS[level]++
					canBeFS = true
				}
				if esMatchPostagRegexp(tokens[i-j], esGNMP) {
					cNMP[level]++
					canBeMP = true
				}
				if esMatchPostagRegexp(tokens[i-j], esGNFP) {
					cNFP[level]++
					canBeFP = true
				}
			}
			if !esMatchPostagRegexp(tokens[i-j], esGN) {
				if esMatchPostagRegexp(tokens[i-j], esNOMMS) {
					cNMS[level]++
					canBeMS = true
				} else if esMatchPostagRegexp(tokens[i-j], esNOMFS) {
					cNFS[level]++
					canBeFS = true
				} else if esMatchPostagRegexp(tokens[i-j], esNOMMP) {
					cNMP[level]++
					canBeMP = true
				} else if esMatchPostagRegexp(tokens[i-j], esNOMMN) {
					cNMN[level]++
					canBeMS = true
					canBeMP = true
				} else if esMatchPostagRegexp(tokens[i-j], esNOMFP) {
					cNFP[level]++
					canBeFP = true
				} else if esMatchPostagRegexp(tokens[i-j], esNOMCS) {
					cNCS[level]++
					canBeMS = true
					canBeFS = true
				} else if esMatchPostagRegexp(tokens[i-j], esNOMCP) {
					cNCP[level]++
					canBeFP = true
					canBeMP = true
				}
			}
		}
		if esMatchPostagRegexp(tokens[i-j], esNOM) {
			cNt[level]++
			isPrevNoun = true
		} else {
			isPrevNoun = false
		}

		if esMatchPostagRegexp(tokens[i-j], esDETCS) {
			if i-j+1 < len(tokens) && esMatchPostagRegexp(tokens[i-j+1], esNOMMS) {
				cDMS[level]++
				canBeMS = true
			}
			if i-j+1 < len(tokens) && esMatchPostagRegexp(tokens[i-j+1], esNOMFS) {
				cDFS[level]++
				canBeFS = true
			}
		}
		if !esMatchPostagRegexp(tokens[i-j], esADVERBI) {
			if esMatchPostagRegexp(tokens[i-j], esDETMS) {
				cDMS[level]++
				canBeMS = true
			}
			if esMatchPostagRegexp(tokens[i-j], esDETFS) {
				cDFS[level]++
				canBeFS = true
			}
			if esMatchPostagRegexp(tokens[i-j], esDETMP) {
				cDMP[level]++
				canBeMP = true
			}
			if esMatchPostagRegexp(tokens[i-j], esDETFP) {
				cDFP[level]++
				canBeFP = true
			}
		}
		if i-j > 0 {
			if esMatchRegexp(tokens[i-j].GetToken(), esPREPNIVEL) &&
				!esMatchRegexp(tokens[i-j-1].GetToken(), esCOORDIONI) &&
				!esMatchPostagRegexp(tokens[i-j+1], esADVERBI) {
				level++
			}
		}
		if level > 0 && esMatchRegexp(tokens[i-j].GetToken(), esCOORDIONI) {
			k := 1
			for k < 4 && i-j-k > 0 &&
				(esMatchPostagRegexp(tokens[i-j-k], esKEEPCOUNT) ||
					esMatchRegexp(tokens[i-j-k].GetToken(), esKEEPCOUNT2) ||
					esMatchPostagRegexp(tokens[i-j-k], esADVERBISOK)) &&
				!esMatchRegexp(tokens[i-j-k].GetToken(), esSTOPCOUNT) {
				if esMatchPostagRegexp(tokens[i-j-k], esPREPOS) {
					j = j + k
					break
				}
				k++
			}
		}
		f.updateApparitions(tokens[i-j])
		j++
	}
	level++
	if level > maxLevels {
		level = maxLevels
	}
	j = 0
	cNtotal := 0
	cDtotal := 0
	for j < level {
		cN[j] = cNMS[j] + cNFS[j] + cNMP[j] + cNFP[j] + cNCS[j] + cNCP[j] + cNMN[j]
		cD[j] = cDMS[j] + cDFS[j] + cDMP[j] + cDFP[j]
		cNtotal += cN[j]
		cDtotal += cD[j]

		if esMatchPostagRegexp(tokens[i], esADJECTIUMP) && (cN[j] > 1 || cD[j] > 1) &&
			(cNMS[j]+cNMN[j]+cNMP[j]+cNCS[j]+cNCP[j]+cDMS[j]+cDMP[j]) > 0 &&
			(cNFS[j]+cNFP[j] <= cNt[j]) {
			return nil
		}
		if esMatchPostagRegexp(tokens[i], esADJECTIUFP) && (cN[j] > 1 || cD[j] > 1) &&
			((cNMS[j]+cNMP[j]+cNMN[j]+cDMS[j]+cDMP[j]) == 0 || (cNt[j] > 0 && cNFS[j]+cNFP[j] >= cNt[j])) {
			return nil
		}
		if cN[j]+cD[j] > 0 {
			isPlural = isPlural && cD[j] > 1
			canBeP = canBeP || cN[j] > 1
		}
		j++
	}
	// comma + plural noun
	isPlural = isPlural || (i-2 > 0 && cNMP[0]+cNFP[0]+cNCP[0] > 0 && tokens[i-2].GetToken() == ",")

	if cNtotal == 0 && cDtotal == 0 {
		return nil
	}

	if esMatchPostagRegexp(tokens[i], esADJECTIUCS) {
		substPattern = esGNCSSub
		adjPattern = esADJECTIUS
		gnPattern = esGNCS
	} else if esMatchPostagRegexp(tokens[i], esADJECTIUCP) {
		substPattern = esGNCPSub
		adjPattern = esADJECTIUP
		gnPattern = esGNCP
	} else if esMatchPostagRegexp(tokens[i], esADJECTIUMS) {
		substPattern = esGNMSSub
		adjPattern = esADJECTIUMS
		gnPattern = esGNMS
	} else if esMatchPostagRegexp(tokens[i], esADJECTIUFS) {
		substPattern = esGNFSSub
		adjPattern = esADJECTIUFS
		gnPattern = esGNFS
	} else if esMatchPostagRegexp(tokens[i], esADJECTIUMP) {
		substPattern = esGNMPSub
		adjPattern = esADJECTIUMP
		gnPattern = esGNMP
	} else if esMatchPostagRegexp(tokens[i], esADJECTIUFP) {
		substPattern = esGNFPSub
		adjPattern = esADJECTIUFP
		gnPattern = esGNFP
	}

	if substPattern == nil || gnPattern == nil || adjPattern == nil {
		return nil
	}

	j = 1
	keepCount := true
	for i-j > 0 && keepCount {
		if esMatchPostagRegexp(tokens[i-j], esNOMDET) && esMatchPostagRegexp(tokens[i-j], gnPattern) {
			return nil
		} else if !esMatchPostagRegexp(tokens[i-j], esGN) && esMatchPostagRegexp(tokens[i-j], substPattern) {
			return nil
		}
		keepCount = !esMatchPostagRegexp(tokens[i-j], esNOMDET)
		j++
	}

	// Necessary condition: previous token is a non-agreeing noun / adj / accepted adv
	if i-1 < 0 || tokens[i-1] == nil {
		return nil
	}
	cond := (esMatchPostagRegexp(tokens[i-1], esNOM) && !esMatchPostagRegexp(tokens[i-1], substPattern)) ||
		(esMatchPostagRegexp(tokens[i-1], esADJECTIU) && !esMatchPostagRegexp(tokens[i-1], gnPattern)) ||
		(esMatchPostagRegexp(tokens[i-1], esADJECTIU) && !esMatchPostagRegexp(tokens[i-1], adjPattern)) ||
		(i > 2 && esMatchPostagRegexp(tokens[i-1], esADVERBISOK) && !esMatchPostagRegexp(tokens[i-2], esVERB) &&
			!esMatchPostagRegexp(tokens[i-2], esPREPOS)) ||
		(i > 3 && esMatchPostagRegexp(tokens[i-1], esLOCADV) && esMatchPostagRegexp(tokens[i-2], esLOCADV) &&
			!esMatchPostagRegexp(tokens[i-3], esVERB) && !esMatchPostagRegexp(tokens[i-3], esPREPOS))
	if !cond {
		return nil
	}

	if !(isPlural && esMatchPostagRegexp(tokens[i], esADJECTIUS)) {
		j = 1
		f.initializeApparitions()
		for i-j > 0 && f.keepCounting(tokens[i-j]) && (level > 1 || j < 4) {
			if !esMatchPostagRegexp(tokens[i-j], esGN) && esMatchPostagRegexp(tokens[i-j], esNOMDET) &&
				esMatchPostagRegexp(tokens[i-j], substPattern) {
				return nil
			} else if esMatchPostagRegexp(tokens[i-j], gnPattern) {
				return nil
			}
			f.updateApparitions(tokens[i-j])
			j++
		}
	}

	// Synthesize suggestions (SpanishSynthesizer.INSTANCE in Java)
	var suggestions []string
	at := esGetAnalyzedToken(tokens[patternTokenPos], esADJECTIUCS)
	if at != nil {
		suggestions = append(suggestions, f.synth(at, "A..CP.")...)
	}
	if len(suggestions) == 0 {
		at = esGetAnalyzedToken(tokens[patternTokenPos], esADJECTIUCP)
		if at != nil {
			suggestions = append(suggestions, f.synth(at, "A..CS.")...)
		}
	}
	if len(suggestions) == 0 && isPlural {
		at = esGetAnalyzedToken(tokens[patternTokenPos], esADJECTIUP)
		if at != nil {
			suggestions = append(suggestions, f.synth(at, "A...P.|V.P..P.|PX..P.*")...)
		}
	}
	at = esGetAnalyzedToken(tokens[patternTokenPos], esADJECTIU)
	if at != nil && len(suggestions) == 0 {
		if canBeMS && !isPlural {
			suggestions = append(suggestions, f.synth(at, "A..MS.|V.P..SM|PX.MS.*")...)
		}
		if canBeFS && !isPlural {
			suggestions = append(suggestions, f.synth(at, "A..FS.|V.P..SF|PX.FS.*")...)
		}
		if canBeMP {
			suggestions = append(suggestions, f.synth(at, "A..MP.|V.P..PM|PX.MP.*")...)
		}
		if canBeFP {
			suggestions = append(suggestions, f.synth(at, "A..FP.|V.P..PF|PX.FP.*")...)
		}
		if canBeMS && (isPlural || canBeP) {
			suggestions = append(suggestions, f.synth(at, "A..MP.|V.P..PM|PX.MP.*")...)
		}
		if canBeFS && !canBeMS && (isPlural || canBeP) {
			suggestions = append(suggestions, f.synth(at, "A..FP.|V.P..PF|PX.FP.*")...)
		}
	}
	origLower := strings.ToLower(tokens[patternTokenPos].GetToken())
	// Java List.remove(Object) removes first occurrence of equal string
	filtered := suggestions[:0]
	removed := false
	for _, s := range suggestions {
		if !removed && s == origLower {
			removed = true
			continue
		}
		filtered = append(filtered, s)
	}
	suggestions = filtered
	if len(suggestions) == 0 {
		return nil
	}

	var definitive []string
	if addComma {
		definitive = append(definitive, ", "+tokens[patternTokenPos].GetToken())
		for _, s := range suggestions {
			definitive = append(definitive, " "+s)
		}
		match.SetOffsetPosition(match.GetFromPos()-1, match.GetToPos())
		match.SetSentencePosition(match.GetFromPosSentence()-1, match.GetToPosSentence())
	} else {
		definitive = append(definitive, suggestions...)
	}
	match.SetSuggestedReplacements(esDistinct(definitive))
	return match
}

func (f *PostponedAdjectiveConcordanceFilter) synth(at *languagetool.AnalyzedToken, posTagRE string) []string {
	if f == nil || f.Synthesize == nil || at == nil {
		return nil
	}
	return f.Synthesize(at, posTagRE)
}

func (f *PostponedAdjectiveConcordanceFilter) keepCounting(aTr *languagetool.AnalyzedTokenReadings) bool {
	if aTr == nil {
		return false
	}
	if (f.adverbAppeared && f.conjunctionAppeared) || (f.adverbAppeared && f.punctuationAppeared) ||
		(f.conjunctionAppeared && f.punctuationAppeared) || (f.punctuationAppeared && esMatchPostagRegexp(aTr, esPUNTUACIO)) {
		return false
	}
	return (esMatchPostagRegexp(aTr, esKEEPCOUNT) || esMatchRegexp(aTr.GetToken(), esKEEPCOUNT2) ||
		esMatchPostagRegexp(aTr, esADVERBISOK)) && !esMatchRegexp(aTr.GetToken(), esSTOPCOUNT) &&
		(!esMatchPostagRegexp(aTr, esGV) || esMatchPostagRegexp(aTr, esGN))
}

func (f *PostponedAdjectiveConcordanceFilter) initializeApparitions() {
	f.adverbAppeared = false
	f.conjunctionAppeared = false
	f.punctuationAppeared = false
}

func (f *PostponedAdjectiveConcordanceFilter) updateApparitions(aTr *languagetool.AnalyzedTokenReadings) {
	if aTr == nil {
		return
	}
	if esMatchPostagRegexp(aTr, esNOM) || esMatchPostagRegexp(aTr, esADJECTIU) {
		f.initializeApparitions()
		return
	}
	f.adverbAppeared = f.adverbAppeared || esMatchPostagRegexp(aTr, esADVERBI)
	f.conjunctionAppeared = f.conjunctionAppeared || esMatchPostagRegexp(aTr, esCONJUNCIO)
	f.punctuationAppeared = f.punctuationAppeared || (esMatchPostagRegexp(aTr, esPUNTUACIO) || aTr.GetToken() == ",")
}

// esFullMatch is Java Matcher.matches() — entire string, not a substring find.
// Go regexp.MatchString is unanchored (find); Java Pattern used via m.matches() is anchored.
func esFullMatch(pattern *regexp.Regexp, s string) bool {
	if pattern == nil {
		return false
	}
	loc := pattern.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

func esMatchPostagRegexp(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) bool {
	if aToken == nil || pattern == nil {
		return false
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if pt := analyzedToken.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if esFullMatch(pattern, posTag) {
			return true
		}
	}
	return false
}

func esMatchRegexp(s string, pattern *regexp.Regexp) bool {
	return esFullMatch(pattern, s)
}

func esGetAnalyzedToken(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) *languagetool.AnalyzedToken {
	if aToken == nil || pattern == nil {
		return nil
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if pt := analyzedToken.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if esFullMatch(pattern, posTag) {
			return analyzedToken
		}
	}
	return nil
}

func esDistinct(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
