package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// PostponedAdjectiveConcordanceFilter ports
// org.languagetool.rules.ca.PostponedAdjectiveConcordanceFilter (1:1).
//
// Synthesize ports Synthesizer.synthesize(token, posTag, true) from
// getSynthesizerFromRuleMatch; when nil, empty suggestions → return nil.
type PostponedAdjectiveConcordanceFilter struct {
	maxLevels int
	// Synthesize ports synth.synthesize(AnalyzedToken, posTagRegex, true).
	Synthesize func(tok *languagetool.AnalyzedToken, posTagRegex string) []string

	adverbAppeared      bool
	conjunctionAppeared bool
	punctuationAppeared bool
}

func NewPostponedAdjectiveConcordanceFilter() *PostponedAdjectiveConcordanceFilter {
	return &PostponedAdjectiveConcordanceFilter{maxLevels: 4}
}

// Java Pattern.compile strings (CA module) — full-string match via caFullMatch.
var (
	caNOM     = regexp.MustCompile(`N.*`)
	caNOMMS   = regexp.MustCompile(`N.MS.*`)
	caNOMFS   = regexp.MustCompile(`N.FS.*`)
	caNOMMP   = regexp.MustCompile(`N.MP.*`)
	caNOMMN   = regexp.MustCompile(`N.MN.*`)
	caNOMFP   = regexp.MustCompile(`N.FP.*`)
	caNOMCS   = regexp.MustCompile(`N.CS.*`)
	caNOMCP   = regexp.MustCompile(`N.CP.*`)
	caNOMDET  = regexp.MustCompile(`N.*|D[NDA0I].*`)
	caGN      = regexp.MustCompile(`_GN_.*`)
	caGNMS    = regexp.MustCompile(`_GN_MS`)
	caGNFS    = regexp.MustCompile(`_GN_FS`)
	caGNMP    = regexp.MustCompile(`_GN_MP`)
	caGNFP    = regexp.MustCompile(`_GN_FP`)
	caGNCS    = regexp.MustCompile(`_GN_[MF]S`)
	caGNCP    = regexp.MustCompile(`_GN_[MF]P`)
	caDET     = regexp.MustCompile(`D[NDA0IP].*`)
	caDETCS   = regexp.MustCompile(`D[NDA0IP]0CS0`)
	caDETMS   = regexp.MustCompile(`D[NDA0IP]0MS0`)
	caDETFS   = regexp.MustCompile(`D[NDA0IP]0FS0`)
	caDETMP   = regexp.MustCompile(`D[NDA0IP]0MP0`)
	caDETFP   = regexp.MustCompile(`D[NDA0IP]0FP0`)
	caGNMSSub = regexp.MustCompile(`N.[MC][SN].*|A..[MC][SN].*|V.P..SM.?|PX.MS.*|D[NDA0I]0MS0|PI0MS000`)
	caGNFSSub = regexp.MustCompile(`N.[FC][SN].*|A..[FC][SN].*|V.P..SF.?|PX.FS.*|D[NDA0I]0FS0|PI0FS000`)
	caGNMPSub = regexp.MustCompile(`N.[MC][PN].*|A..[MC][PN].*|V.P..PM.?|PX.MP.*|D[NDA0I]0MP0`)
	caGNFPSub = regexp.MustCompile(`N.[FC][PN].*|A..[FC][PN].*|V.P..PF.?|PX.FP.*|D[NDA0I]0FP0`)
	caGNCPSub = regexp.MustCompile(`N.[FMC][PN].*|A..[FMC][PN].*|D[NDA0I]0[FM]P0`)
	caGNCSSub = regexp.MustCompile(`N.[FMC][SN].*|A..[FMC][SN].*|D[NDA0I]0[FM]S0`)
	caADJECTIU   = regexp.MustCompile(`AQ.*|V.P.*|PX.*|.*LOC_ADJ.*`)
	caADJECTIUMS = regexp.MustCompile(`A..[MC][SN].*|V.P..SM.?|PX.MS.*`)
	caADJECTIUFS = regexp.MustCompile(`A..[FC][SN].*|V.P..SF.?|PX.FS.*`)
	caADJECTIUMP = regexp.MustCompile(`A..[MC][PN].*|V.P..PM.?|PX.MP.*`)
	caADJECTIUFP = regexp.MustCompile(`A..[FC][PN].*|V.P..PF.?|PX.FP.*`)
	caADJECTIUCP = regexp.MustCompile(`A..C[PN].*`)
	caADJECTIUCS = regexp.MustCompile(`A..C[SN].*`)
	caADJECTIUS  = regexp.MustCompile(`A...[SN].*|V.P..S..?|PX..S.*`)
	caADJECTIUP  = regexp.MustCompile(`A...[PN].*|V.P..P..?|PX..P.*`)
	caADVERBI    = regexp.MustCompile(`R.|.*LOC_ADV.*`)
	caCONJUNCIO  = regexp.MustCompile(`C.|.*LOC_CONJ.*`)
	caPUNTUACIO  = regexp.MustCompile(`_PUNCT`)
	caLOCADV     = regexp.MustCompile(`.*LOC_ADV.*`)
	caADVERBISOK = regexp.MustCompile(`RG_anteposat`)
	caCOORDIONI  = regexp.MustCompile(`i|o|ni`)
	caKEEPCOUNT  = regexp.MustCompile(`A.*|N.*|D[NAIDP].*|SPS.*|.*LOC_ADV.*|V.P.*|_PUNCT.*|.*LOC_ADJ.*|PX.*|PI0.S000|UNKNOWN`)
	caKEEPCOUNT2 = regexp.MustCompile(`,|i|o|ni`)
	caSTOPCOUNT  = regexp.MustCompile(`[;:]`)
	caPREPOS     = regexp.MustCompile(`SPS.*`)
	caPREPNIVEL  = regexp.MustCompile(`de|d'|en|sobre|a|entre|per|pe|amb|sense|contra|com|envers`)
	caVERB       = regexp.MustCompile(`V.[^P].*|_GV_`)
	caGV         = regexp.MustCompile(`_GV_`)
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
			if caMatchPostagRegexp(tokens[i-j], caNOM) ||
				(i-j-1 > 0 && !caMatchPostagRegexp(tokens[i-j], caNOM) && caMatchPostagRegexp(tokens[i-j], caADJECTIU) &&
					caMatchPostagRegexp(tokens[i-j-1], caDET)) {
				if caMatchPostagRegexp(tokens[i-j], caGNMS) {
					cNMS[level]++
					canBeMS = true
				}
				if caMatchPostagRegexp(tokens[i-j], caGNFS) {
					cNFS[level]++
					canBeFS = true
				}
				if caMatchPostagRegexp(tokens[i-j], caGNMP) {
					cNMP[level]++
					canBeMP = true
				}
				if caMatchPostagRegexp(tokens[i-j], caGNFP) {
					cNFP[level]++
					canBeFP = true
				}
			}
			if !caMatchPostagRegexp(tokens[i-j], caGN) {
				if caMatchPostagRegexp(tokens[i-j], caNOMMS) {
					cNMS[level]++
					canBeMS = true
				} else if caMatchPostagRegexp(tokens[i-j], caNOMFS) {
					cNFS[level]++
					canBeFS = true
				} else if caMatchPostagRegexp(tokens[i-j], caNOMMP) {
					cNMP[level]++
					canBeMP = true
				} else if caMatchPostagRegexp(tokens[i-j], caNOMMN) {
					cNMN[level]++
					canBeMS = true
					canBeMP = true
				} else if caMatchPostagRegexp(tokens[i-j], caNOMFP) {
					cNFP[level]++
					canBeFP = true
				} else if caMatchPostagRegexp(tokens[i-j], caNOMCS) {
					cNCS[level]++
					canBeMS = true
					canBeFS = true
				} else if caMatchPostagRegexp(tokens[i-j], caNOMCP) {
					cNCP[level]++
					canBeFP = true
					canBeMP = true
				}
			}
		}
		if caMatchPostagRegexp(tokens[i-j], caNOM) {
			cNt[level]++
			isPrevNoun = true
		} else {
			isPrevNoun = false
		}

		if caMatchPostagRegexp(tokens[i-j], caDETCS) {
			if i-j+1 < len(tokens) && caMatchPostagRegexp(tokens[i-j+1], caNOMMS) {
				cDMS[level]++
				canBeMS = true
			}
			if i-j+1 < len(tokens) && caMatchPostagRegexp(tokens[i-j+1], caNOMFS) {
				cDFS[level]++
				canBeFS = true
			}
		}
		if !caMatchPostagRegexp(tokens[i-j], caADVERBI) {
			// exception: tot el
			skipDet := false
			if i-j+1 < len(tokens) && tokens[i-j].HasAnyLemma("tot") && tokens[i-j+1].HasAnyLemma("el") {
				skipDet = true
			}
			if !skipDet {
				if caMatchPostagRegexp(tokens[i-j], caDETMS) {
					cDMS[level]++
					canBeMS = true
				}
				if caMatchPostagRegexp(tokens[i-j], caDETFS) {
					cDFS[level]++
					canBeFS = true
				}
				if caMatchPostagRegexp(tokens[i-j], caDETMP) {
					cDMP[level]++
					canBeMP = true
				}
				if caMatchPostagRegexp(tokens[i-j], caDETFP) {
					cDFP[level]++
					canBeFP = true
				}
			}
		}
		if i-j-1 > 0 {
			if caMatchRegexp(tokens[i-j].GetToken(), caPREPNIVEL) &&
				!caMatchPostagRegexp(tokens[i-j], caCONJUNCIO) &&
				!caMatchRegexp(tokens[i-j-1].GetToken(), caCOORDIONI) &&
				i-j+1 < len(tokens) && !caMatchPostagRegexp(tokens[i-j+1], caADVERBI) {
				level++
			}
		}
		j = f.updateJValue(tokens, i, j, level)
		// After updateJValue, Java still uses tokens[i-j] for updateApparitions
		// (j may have been increased by 1 for "dos o més").
		if i-j >= 0 && i-j < len(tokens) {
			f.updateApparitions(tokens[i-j])
		}
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

		if caMatchPostagRegexp(tokens[i], caADJECTIUMP) && (cN[j] > 1 || cD[j] > 1) &&
			(cNMS[j]+cNMN[j]+cNMP[j]+cNCS[j]+cNCP[j]+cDMS[j]+cDMP[j]) > 0 &&
			(cNFS[j]+cNFP[j] <= cNt[j]) {
			return nil
		}
		if caMatchPostagRegexp(tokens[i], caADJECTIUFP) && (cN[j] > 1 || cD[j] > 1) &&
			((cNMS[j]+cNMP[j]+cNMN[j]+cDMS[j]+cDMP[j]) == 0 || (cNt[j] > 0 && cNFS[j]+cNFP[j] >= cNt[j])) {
			return nil
		}
		if cN[j]+cD[j] > 0 {
			isPlural = isPlural && cD[j] > 1
			canBeP = canBeP || cN[j] > 1
		}
		j++
	}
	isPlural = isPlural || (i-2 > 0 && cNMP[0]+cNFP[0]+cNCP[0] > 0 && tokens[i-2].GetToken() == ",")

	if cNtotal == 0 && cDtotal == 0 {
		return nil
	}

	if caMatchPostagRegexp(tokens[i], caADJECTIUCS) {
		substPattern = caGNCSSub
		adjPattern = caADJECTIUS
		gnPattern = caGNCS
	} else if caMatchPostagRegexp(tokens[i], caADJECTIUCP) {
		substPattern = caGNCPSub
		adjPattern = caADJECTIUP
		gnPattern = caGNCP
	} else if caMatchPostagRegexp(tokens[i], caADJECTIUMS) {
		substPattern = caGNMSSub
		adjPattern = caADJECTIUMS
		gnPattern = caGNMS
	} else if caMatchPostagRegexp(tokens[i], caADJECTIUFS) {
		substPattern = caGNFSSub
		adjPattern = caADJECTIUFS
		gnPattern = caGNFS
	} else if caMatchPostagRegexp(tokens[i], caADJECTIUMP) {
		substPattern = caGNMPSub
		adjPattern = caADJECTIUMP
		gnPattern = caGNMP
	} else if caMatchPostagRegexp(tokens[i], caADJECTIUFP) {
		substPattern = caGNFPSub
		adjPattern = caADJECTIUFP
		gnPattern = caGNFP
	}

	if substPattern == nil || gnPattern == nil || adjPattern == nil {
		return nil
	}

	j = 1
	keepCount := true
	for i-j > 0 && keepCount {
		if caMatchPostagRegexp(tokens[i-j], caNOMDET) && caMatchPostagRegexp(tokens[i-j], gnPattern) {
			return nil
		} else if !caMatchPostagRegexp(tokens[i-j], caGN) && caMatchPostagRegexp(tokens[i-j], substPattern) {
			return nil
		}
		keepCount = !caMatchPostagRegexp(tokens[i-j], caNOMDET)
		j++
	}

	if i-1 < 0 || tokens[i-1] == nil {
		return nil
	}
	cond := (caMatchPostagRegexp(tokens[i-1], caNOM) && !caMatchPostagRegexp(tokens[i-1], substPattern)) ||
		(caMatchPostagRegexp(tokens[i-1], caADJECTIU) && !caMatchPostagRegexp(tokens[i-1], gnPattern)) ||
		(caMatchPostagRegexp(tokens[i-1], caADJECTIU) && !caMatchPostagRegexp(tokens[i-1], adjPattern)) ||
		(i > 2 && caMatchPostagRegexp(tokens[i-1], caADVERBISOK) && !caMatchPostagRegexp(tokens[i-2], caVERB) &&
			!caMatchPostagRegexp(tokens[i-2], caPREPOS)) ||
		(i > 3 && caMatchPostagRegexp(tokens[i-1], caLOCADV) && caMatchPostagRegexp(tokens[i-2], caLOCADV) &&
			!caMatchPostagRegexp(tokens[i-3], caVERB) && !caMatchPostagRegexp(tokens[i-3], caPREPOS))
	if !cond {
		return nil
	}

	// Adjective can't be singular. The rule matches
	// Java: while (i - j > 0 && keepCounting(...)) { //&& (level > 1 || j < 4)  — constraint commented out
	if !(isPlural && caMatchPostagRegexp(tokens[i], caADJECTIUS)) {
		j = 1
		f.initializeApparitions()
		for i-j > 0 && f.keepCounting(tokens[i-j]) {
			if !caMatchPostagRegexp(tokens[i-j], caGN) && caMatchPostagRegexp(tokens[i-j], caNOMDET) &&
				caMatchPostagRegexp(tokens[i-j], substPattern) {
				return nil
			} else if caMatchPostagRegexp(tokens[i-j], gnPattern) {
				return nil
			}
			j = f.updateJValue(tokens, i, j, 0)
			if i-j >= 0 && i-j < len(tokens) {
				f.updateApparitions(tokens[i-j])
			}
			j++
		}
	}

	var suggestions []string
	at := caGetAnalyzedToken(tokens[patternTokenPos], caADJECTIUCS)
	if at != nil {
		suggestions = append(suggestions, f.synth(at, "A..CP.")...)
	}
	if len(suggestions) == 0 {
		at = caGetAnalyzedToken(tokens[patternTokenPos], caADJECTIUCP)
		if at != nil {
			suggestions = append(suggestions, f.synth(at, "A..CS.")...)
		}
	}
	if len(suggestions) == 0 && isPlural {
		at = caGetAnalyzedToken(tokens[patternTokenPos], caADJECTIUP)
		if at != nil {
			// Java: "A...P.|V.P..P..|PX..P.*" (note double dots after P)
			suggestions = append(suggestions, f.synth(at, "A...P.|V.P..P..|PX..P.*")...)
		}
	}
	at = caGetAnalyzedToken(tokens[patternTokenPos], caADJECTIU)
	if at != nil && len(suggestions) == 0 {
		if canBeMS && !isPlural {
			suggestions = append(suggestions, f.synth(at, "A..MS.|V.P..SM.|PX.MS.*")...)
		}
		if canBeFS && !isPlural {
			suggestions = append(suggestions, f.synth(at, "A..FS.|V.P..SF.|PX.FS.*")...)
		}
		if canBeMP {
			suggestions = append(suggestions, f.synth(at, "A..MP.|V.P..PM.|PX.MP.*")...)
		}
		if canBeFP {
			suggestions = append(suggestions, f.synth(at, "A..FP.|V.P..PF.|PX.FP.*")...)
		}
		if canBeMS && (isPlural || canBeP) {
			suggestions = append(suggestions, f.synth(at, "A..MP.|V.P..PM.|PX.MP.*")...)
		}
		if canBeFS && !canBeMS && (isPlural || canBeP) {
			suggestions = append(suggestions, f.synth(at, "A..FP.|V.P..PF.|PX.FP.*")...)
		}
	}
	origLower := strings.ToLower(tokens[patternTokenPos].GetToken())
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
	match.SetSuggestedReplacements(caDistinct(definitive))
	return match
}

func (f *PostponedAdjectiveConcordanceFilter) synth(at *languagetool.AnalyzedToken, posTagRE string) []string {
	if f == nil || f.Synthesize == nil || at == nil {
		return nil
	}
	return f.Synthesize(at, posTagRE)
}

// updateJValue ports PostponedAdjectiveConcordanceFilter.updateJValue ("dos o més").
func (f *PostponedAdjectiveConcordanceFilter) updateJValue(tokens []*languagetool.AnalyzedTokenReadings, i, j, level int) int {
	_ = f
	_ = level
	// commented-out level/coordinacio block in Java omitted (still commented).
	if caMatchRegexp(tokens[i-j].GetToken(), caCOORDIONI) {
		if i-j-1 > 0 && i-j+1 < len(tokens) {
			if caMatchPostagRegexp(tokens[i-j-1], caDET) && tokens[i-j+1].GetToken() == "més" {
				j = j + 1
			}
		}
	}
	return j
}

func (f *PostponedAdjectiveConcordanceFilter) keepCounting(aTr *languagetool.AnalyzedTokenReadings) bool {
	if aTr == nil {
		return false
	}
	if caMatchRegexp(aTr.GetToken(), caPREPNIVEL) {
		return true
	}
	if aTr.GetToken() == "." { // abbreviation, not sentence end
		return true
	}
	if (f.adverbAppeared && f.conjunctionAppeared) || (f.adverbAppeared && f.punctuationAppeared) ||
		(f.conjunctionAppeared && f.punctuationAppeared) || (f.punctuationAppeared && caMatchPostagRegexp(aTr, caPUNTUACIO)) {
		return false
	}
	return (caMatchPostagRegexp(aTr, caKEEPCOUNT) || caMatchRegexp(aTr.GetToken(), caKEEPCOUNT2) ||
		caMatchPostagRegexp(aTr, caADVERBISOK)) && !caMatchRegexp(aTr.GetToken(), caSTOPCOUNT) &&
		(!caMatchPostagRegexp(aTr, caGV) || caMatchPostagRegexp(aTr, caGN))
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
	f.conjunctionAppeared = f.conjunctionAppeared || caMatchPostagRegexp(aTr, caCONJUNCIO)
	if aTr.GetToken() == "com" {
		return
	}
	if caMatchPostagRegexp(aTr, caNOM) || caMatchPostagRegexp(aTr, caADJECTIU) {
		f.initializeApparitions()
		return
	}
	f.adverbAppeared = f.adverbAppeared || caMatchPostagRegexp(aTr, caADVERBI)
	f.punctuationAppeared = f.punctuationAppeared || (caMatchPostagRegexp(aTr, caPUNTUACIO) || aTr.GetToken() == ",")
}

// caFullMatch is Java Matcher.matches() (entire string).
func caFullMatch(pattern *regexp.Regexp, s string) bool {
	if pattern == nil {
		return false
	}
	loc := pattern.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

func caMatchPostagRegexp(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) bool {
	if aToken == nil || pattern == nil {
		return false
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if pt := analyzedToken.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if caFullMatch(pattern, posTag) {
			return true
		}
	}
	return false
}

func caMatchRegexp(s string, pattern *regexp.Regexp) bool {
	return caFullMatch(pattern, s)
}

func caGetAnalyzedToken(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) *languagetool.AnalyzedToken {
	if aToken == nil || pattern == nil {
		return nil
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if pt := analyzedToken.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if caFullMatch(pattern, posTag) {
			return analyzedToken
		}
	}
	return nil
}

func caDistinct(in []string) []string {
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
