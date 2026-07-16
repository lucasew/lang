package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RedundantModalOrAuxiliaryVerb is a surface stand-in for
// org.languagetool.rules.de.RedundantModalOrAuxiliaryVerb.
// Without VER:MOD/VER:AUX tags, common modal/auxiliary surface forms are tracked;
// a second identical form after und/oder/sowie is flagged (simplified).
type RedundantModalOrAuxiliaryVerb struct {
	Messages map[string]string
	forms    map[string]bool // true = modal, false = aux
}

func NewRedundantModalOrAuxiliaryVerb(messages map[string]string) *RedundantModalOrAuxiliaryVerb {
	forms := map[string]bool{}
	// modals
	for _, w := range []string{
		"kann", "kannst", "können", "könnt", "konnte", "konntest", "konnten", "könnte", "könntest", "könnten",
		"muss", "muß", "musst", "mußt", "müssen", "müsst", "musste", "mußte", "müsste", "müßten",
		"soll", "sollst", "sollen", "sollt", "sollte", "solltest", "sollten",
		"will", "willst", "wollen", "wollt", "wollte", "wolltest", "wollten",
		"darf", "darfst", "dürfen", "dürft", "durfte", "dürfte",
		"mag", "magst", "mögen", "mögt", "mochte", "möchte", "möchtest", "möchten",
		"werde", "wirst", "wird", "werden", "werdet", "wurde", "wurden", "würde", "würden",
	} {
		forms[w] = true
	}
	// auxiliaries (haben/sein/werden already partly modal list)
	for _, w := range []string{
		"habe", "hast", "hat", "haben", "habt", "hatte", "hatten", "hätte", "hätten",
		"bin", "bist", "ist", "sind", "seid", "war", "warst", "waren", "wär", "wäre", "wären",
		"sei", "seist", "seien",
	} {
		forms[w] = false
	}
	return &RedundantModalOrAuxiliaryVerb{Messages: messages, forms: forms}
}

func (r *RedundantModalOrAuxiliaryVerb) GetID() string { return "REDUNDANT_MODAL_VERB" }

func isRedundantBreak(s string) bool {
	switch s {
	case ",", ";", ".", ":", "?", "!", "-", "–", "—",
		"'", "\"", "„", "“", "”", "»", "«", "(", ")", "[", "]":
		return true
	}
	return false
}

func (r *RedundantModalOrAuxiliaryVerb) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for nt := 2; nt < len(tokens); nt++ {
		sVerb := tokens[nt].GetToken()
		lc := strings.ToLower(sVerb)
		isMod, ok := r.forms[lc]
		if !ok {
			continue
		}
		if nt+1 < len(tokens) && tokens[nt-1].GetToken() == tokens[nt+1].GetToken() {
			continue
		}
		nVerb := nt
		// scan forward for und/oder/sowie then same verb form
		for j := nt + 1; j < len(tokens); j++ {
			s := tokens[j].GetToken()
			if isRedundantBreak(s) {
				break
			}
			if s != "und" && s != "oder" && s != "sowie" {
				continue
			}
			// after conjunction, find same surface form
			for k := j + 1; k < len(tokens); k++ {
				sk := tokens[k].GetToken()
				if isRedundantBreak(sk) {
					break
				}
				if sk != sVerb {
					continue
				}
				// skip if next to conjunction only and previous was also just verb? keep simple
				// Exception soft: quoted "nicht dürfen" und "nicht müssen" — if token is after quote
				if k > 1 && (tokens[k-1].GetToken() == "\"" || tokens[k-1].GetToken() == "„") {
					break
				}
				kind := "Hilfsverb"
				if isMod {
					kind = "Modalverb"
				}
				msg := "Das " + kind + " scheint redundant zu sein. Prüfen Sie, ob es gelöscht oder der Satz umformuliert werden kann."
				rm := rules.NewRuleMatch(r, sentence, tokens[k].GetStartPos(), tokens[k].GetEndPos(), msg)
				rm.ShortMessage = "redundantes " + kind
				rm.SetSuggestedReplacement("")
				matches = append(matches, rm)
				_ = nVerb
				return matches // one match like many Java cases
			}
			break
		}
	}
	return matches
}
