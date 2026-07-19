package de

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RedundantModalOrAuxiliaryVerb ports org.languagetool.rules.de.RedundantModalOrAuxiliaryVerb.
// Java: VER:MOD / VER:AUX only (no surface form invent). Default off.
// Java: STYLE, Style, setDefaultOff().
type RedundantModalOrAuxiliaryVerb struct {
	Messages map[string]string
	// Category ports setCategory(STYLE).
	Category *rules.Category
	// IssueType ports setLocQualityIssueType(Style).
	IssueType rules.ITSIssueType
	// DefaultOff mirrors Java setDefaultOff().
	DefaultOff bool
}

const (
	redundantVerbText = " scheint redundant zu sein. Prüfen Sie, ob es gelöscht oder der Satz umformuliert werden kann."
	redundantSubText  = "Der Satzteil scheint redundant zu sein. Prüfen Sie, ob es gelöscht oder der Satz umformuliert werden kann."
)

var redundantMarksRE = regexp.MustCompile(`^[,;.:?!\-–—’'"„“”»«‚‘›‹()\[\]]$`)

func NewRedundantModalOrAuxiliaryVerb(messages map[string]string) *RedundantModalOrAuxiliaryVerb {
	return &RedundantModalOrAuxiliaryVerb{
		Messages:   messages,
		Category:   rules.CatStyle.GetCategory(messages),
		IssueType:  rules.ITSStyle,
		DefaultOff: true,
	}
}

func (r *RedundantModalOrAuxiliaryVerb) GetID() string { return "REDUNDANT_MODAL_VERB" }

func (r *RedundantModalOrAuxiliaryVerb) GetDescription() string {
	return "Redundantes Modal- oder Hilfsverb"
}

func (r *RedundantModalOrAuxiliaryVerb) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *RedundantModalOrAuxiliaryVerb) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSStyle
	}
	return r.IssueType
}

func (r *RedundantModalOrAuxiliaryVerb) IsDefaultOff() bool { return r != nil && r.DefaultOff }

func isBreakTokenRedundant(sToken string) bool {
	return redundantMarksRE.MatchString(sToken) ||
		sToken == "und" || sToken == "oder" || sToken == "sowie"
}

func isMarkTokenRedundant(sToken string) bool {
	return redundantMarksRE.MatchString(sToken)
}

func hasParticipleAt(nConjunction, nStart int, tokens []*languagetool.AnalyzedTokenReadings) int {
	// Java: only hasPosTagStartingWith("PA2") — no surface invent
	if nConjunction < 1 || tokens[nConjunction-1] == nil {
		return -1
	}
	if !tokens[nConjunction-1].HasPosTagStartingWith("PA2") {
		return -1
	}
	sParticiple := tokens[nConjunction-1].GetToken()
	for i := nStart; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		sToken := tokens[i].GetToken()
		if isBreakTokenRedundant(sToken) {
			return -1
		}
		if sToken == sParticiple {
			if i == len(tokens)-1 || (i+1 < len(tokens) && tokens[i+1] != nil && isBreakTokenRedundant(tokens[i+1].GetToken())) {
				return i
			}
			return -1
		}
	}
	return -1
}

func isModalOrAux(tokens []*languagetool.AnalyzedTokenReadings, nt int) (isMod bool, ok bool) {
	// Java: hasPosTagStartingWith("VER:MOD") / "VER:AUX" only
	if nt < 0 || nt >= len(tokens) || tokens[nt] == nil {
		return false, false
	}
	t := tokens[nt]
	if t.HasPosTagStartingWith("VER:MOD") {
		return true, true
	}
	if t.HasPosTagStartingWith("VER:AUX") {
		return false, true
	}
	return false, false
}

func (r *RedundantModalOrAuxiliaryVerb) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	for nt := 2; nt < len(tokens); nt++ {
		isModVerb, ok := isModalOrAux(tokens, nt)
		if !ok {
			continue
		}
		if nt+1 >= len(tokens) || tokens[nt-1] == nil || tokens[nt+1] == nil {
			continue
		}
		if tokens[nt-1].GetToken() == tokens[nt+1].GetToken() {
			continue
		}
		sVerb := tokens[nt].GetToken()
		nVerb := nt
		doBreak := false
		var suggestion *string
		for nt++; nt < len(tokens); nt++ {
			if tokens[nt] == nil {
				continue
			}
			sToken := tokens[nt].GetToken()
			if isMarkTokenRedundant(sToken) {
				break
			}
			if sToken == "und" || sToken == "oder" || sToken == "sowie" {
				nConjunction := nt
				doBreak = true
				for nt++; nt < len(tokens); nt++ {
					if tokens[nt] == nil {
						continue
					}
					sToken = tokens[nt].GetToken()
					if isBreakTokenRedundant(sToken) {
						break
					}
					if sToken != sVerb {
						continue
					}
					var ruleMatch *rules.RuleMatch
					if nt-1 == nConjunction {
						if nVerb == nConjunction-1 {
							break
						}
						n := 1
						for n+nt < len(tokens) && nVerb+n < len(tokens) &&
							tokens[nt+n] != nil && tokens[nVerb+n] != nil &&
							strings.EqualFold(tokens[nt+n].GetToken(), tokens[nVerb+n].GetToken()) {
							n++
						}
						if n > 1 {
							if nVerb+n == nConjunction {
								break
							}
							ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-1].GetEndPos(), tokens[nt+n-1].GetEndPos(), redundantSubText)
						} else {
							msg := "Das " + verbKind(isModVerb) + redundantVerbText
							ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-1].GetEndPos(), tokens[nt].GetEndPos(), msg)
						}
					} else if strings.EqualFold(tokens[nt-1].GetToken(), tokens[nVerb-1].GetToken()) &&
						tokens[nt-1].HasPosTagStartingWith("PRO:PER") && !tokens[nt-1].HasPosTagStartingWith("ART") {
						// morph-only personal pronoun branch
						if nVerb == nConjunction-1 {
							if nVerb >= 2 {
								ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nVerb-2].GetEndPos(), tokens[nVerb].GetEndPos(), redundantSubText)
							}
						} else if nt >= 2 {
							ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-2].GetEndPos(), tokens[nt].GetEndPos(), redundantSubText)
						}
					} else if nt+1 < len(tokens) && tokens[nt+1] != nil && tokens[nVerb+1] != nil &&
						strings.EqualFold(tokens[nt+1].GetToken(), tokens[nVerb+1].GetToken()) &&
						(tokens[nt+1].HasPosTagStartingWith("PRO:IND") ||
							(tokens[nt+1].HasPosTagStartingWith("PRO:PER") && tokens[nt+1].GetToken() != "Sie" &&
								!tokens[nt+1].HasPosTagStartingWith("ART"))) {
						if tokens[nt+1].HasPosTagStartingWith("PRO:PER:AKK") &&
							tokens[nt].MatchesPosTagRegex(`VER:(AUX|MOD):.*KJ1`) {
							msg := "Das " + verbKind(isModVerb) + redundantVerbText
							ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-1].GetEndPos(), tokens[nt].GetEndPos(), msg)
						} else {
							ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-1].GetEndPos(), tokens[nt+1].GetEndPos(), redundantSubText)
						}
					} else {
						// surface-friendly branch + morph exceptions
						if shouldSkipRedundantBranch(tokens, nt, nVerb, nConjunction) {
							break
						}
						if nVerb == nConjunction-1 {
							n := 1
							for nVerb-n > 0 && nt-n > nConjunction &&
								tokens[nVerb-n] != nil && tokens[nt-n] != nil &&
								tokens[nVerb-n].GetToken() == tokens[nt-n].GetToken() {
								n++
							}
							if n > 1 {
								ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nVerb-n].GetEndPos(), tokens[nVerb].GetEndPos(), redundantSubText)
							} else {
								msg := "Das " + verbKind(isModVerb) + redundantVerbText
								ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nVerb-1].GetEndPos(), tokens[nVerb].GetEndPos(), msg)
							}
						} else {
							paAt := hasParticipleAt(nConjunction, nt+1, tokens)
							if paAt > 0 {
								ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-1].GetEndPos(), tokens[paAt].GetEndPos(), redundantSubText)
								sug := ""
								for i := nt + 1; i < paAt; i++ {
									if tokens[i] != nil {
										sug += " " + tokens[i].GetToken()
									}
								}
								suggestion = &sug
							} else {
								n := 1
								for n+nVerb < nConjunction && n+nt < len(tokens) &&
									tokens[nVerb+n] != nil && tokens[nt+n] != nil &&
									tokens[nVerb+n].GetToken() == tokens[nt+n].GetToken() {
									n++
								}
								if n+nVerb == nConjunction {
									ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-1].GetEndPos(), tokens[nt+n-1].GetEndPos(), redundantSubText)
								} else {
									msg := "Das " + verbKind(isModVerb) + redundantVerbText
									ruleMatch = rules.NewRuleMatch(r, sentence, tokens[nt-1].GetEndPos(), tokens[nt].GetEndPos(), msg)
								}
							}
						}
					}
					if ruleMatch != nil {
						// Java: RuleMatch without shortMessage.
						if suggestion == nil {
							ruleMatch.SetSuggestedReplacement("")
						} else {
							ruleMatch.SetSuggestedReplacement(*suggestion)
						}
						ruleMatches = append(ruleMatches, ruleMatch)
					}
					break
				}
			}
			if doBreak {
				break
			}
		}
	}
	return ruleMatches
}

func verbKind(isMod bool) string {
	if isMod {
		return "Modalverb"
	}
	return "Hilfsverb"
}

// shouldSkipRedundantBranch ports the large Java else-if guard that breaks without match.
func shouldSkipRedundantBranch(tokens []*languagetool.AnalyzedTokenReadings, nt, nVerb, nConjunction int) bool {
	if nt < 1 || nVerb < 1 || nVerb+1 >= len(tokens) || tokens[nt-1] == nil {
		return true
	}
	// Java: if (tokens[nt-1].hasPosTagStartingWith("PRO:PER") || da/zu || ... || break conditions)
	// Without PRO tags, still honor da/zu and token equality checks.
	if tokens[nt-1].HasPosTagStartingWith("PRO:PER") ||
		tokens[nt-1].GetToken() == "da" || tokens[nt-1].GetToken() == "zu" ||
		(tokens[nVerb+1] != nil && tokens[nVerb+1].GetToken() == tokens[nt-1].GetToken()) ||
		(nt+1 < len(tokens) && tokens[nt+1] != nil &&
			(tokens[nt+1].HasPosTagStartingWith("PRO:PER") ||
				tokens[nt-1].GetToken() == tokens[nt+1].GetToken() ||
				tokens[nVerb-1].GetToken() == tokens[nt+1].GetToken() ||
				(tokens[nVerb+1].HasPosTagStartingWith("VER:MOD") && tokens[nt+1].HasPosTagStartingWith("VER:MOD")) ||
				(nVerb == nConjunction-1 && !isBreakTokenRedundant(tokens[nt+1].GetToken())))) ||
		(nVerb < nConjunction-1 && (nt+1 == len(tokens) || (nt+1 < len(tokens) && tokens[nt+1] != nil && isBreakTokenRedundant(tokens[nt+1].GetToken())))) {
		return true
	}
	return false
}
