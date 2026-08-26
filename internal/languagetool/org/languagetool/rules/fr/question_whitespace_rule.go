package fr

import (
	"regexp"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	espaceFineInsecable = "\u202F"
	nbsp                = "\u00A0"
)

var urlSchemeRE = regexp.MustCompile(`^(file|s?ftp|finger|git|gopher|hdl|https?|shttp|imap|mailto|mms|nntp|s?news(post|reply)?|prospero|rsync|rtspu|sips?|svn|svn\+ssh|telnet|wais)$`)

var (
	frWhitespaceAntiOnce  sync.Once
	frWhitespaceAntiRules []*disambigrules.DisambiguationPatternRule
)

// QuestionWhitespaceRule ports org.languagetool.rules.fr.QuestionWhitespaceRule.
// Non-strict: any whitespace before ?!; is accepted.
// Strict (Strict=true): only U+202F or U+00A0 count as allowed whitespace; missing
// space is left to FRENCH_WHITESPACE (mutually exclusive).
type QuestionWhitespaceRule struct {
	Messages map[string]string
	Strict   bool
}

func NewQuestionWhitespaceRule(messages map[string]string) *QuestionWhitespaceRule {
	return &QuestionWhitespaceRule{Messages: messages}
}

func NewQuestionWhitespaceStrictRule(messages map[string]string) *QuestionWhitespaceRule {
	return &QuestionWhitespaceRule{Messages: messages, Strict: true}
}

func (r *QuestionWhitespaceRule) GetID() string {
	if r.Strict {
		return "FRENCH_WHITESPACE_STRICT"
	}
	return "FRENCH_WHITESPACE"
}

func (r *QuestionWhitespaceRule) GetDescription() string {
	return "Insertion des espaces fines insécables"
}

// Match ports QuestionWhitespaceRule.match:
// Java: tokens = getSentenceWithImmunization(sentence).getTokens()
func (r *QuestionWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := getSentenceWithFRWhitespaceImmunization(sentence).GetTokens() // include whitespace tokens
	var ruleMatches []*rules.RuleMatch
	prevPrevToken := ""
	prevToken := ""
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if tokens[i].IsImmunized() || prevToken == "(" || prevToken == "[" {
			prevPrevToken = prevToken
			prevToken = token
			continue
		}

		var msg, suggestionText string
		iFrom, iTo := i-1, i
		isPreviousWhitespace := i > 0 && tokens[i-1].IsWhitespace()
		prevTokenToChange := prevToken
		if isPreviousWhitespace {
			prevTokenToChange = ""
		}
		if !r.isAllowedWhitespaceChar(tokens, i-1) {
			if token == "?" && prevToken != "!" {
				msg = "Le point d'interrogation est précédé d'une espace fine insécable."
				suggestionText = prevTokenToChange + espaceFineInsecable + "?"
			} else if token == "!" && prevToken != "?" {
				msg = "Le point d'exclamation est précédé d'une espace fine insécable."
				suggestionText = prevTokenToChange + espaceFineInsecable + "!"
			} else if token == ";" {
				msg = "Le point-virgule est précédé d'une espace fine insécable."
				suggestionText = prevTokenToChange + espaceFineInsecable + ";"
			} else if token == ":" {
				if !urlSchemeRE.MatchString(prevToken) {
					msg = "Les deux-points sont précédés d'une espace insécable."
					suggestionText = prevTokenToChange + nbsp + ":"
				}
			} else if token == "»" {
				if prevPrevToken == "«" {
					msg = "Les guillemets sont toujours accompagnés d'une espace insécable."
					suggestionText = "«" + nbsp + prevTokenToChange + nbsp + "»"
					iFrom = i - 2
				} else {
					msg = "Le guillemet fermant est précédé d'une espace insécable."
					suggestionText = prevTokenToChange + nbsp + "»"
				}
			}
		}

		if prevToken == "«" {
			if tools.IsEmptyStr(token) || token == "" {
				msg = "Le guillemet ouvrant est suivi d'une espace insécable."
				suggestionText = "«" + nbsp
				iTo = i - 1
			} else if !r.isAllowedWhitespaceChar(tokens, i) {
				nextToken := ""
				if i+1 < len(tokens) {
					nextToken = tokens[i+1].GetToken()
				}
				if nextToken != "»" {
					msg = "Le guillemet ouvrant est suivi d'une espace insécable."
					if !tokens[i].IsWhitespace() {
						suggestionText = "«" + nbsp + token
					} else {
						suggestionText = "«" + nbsp
					}
				}
			}
		}

		if msg != "" {
			fromPos := tokens[iFrom].GetStartPos()
			toPos := tokens[iTo].GetEndPos()
			rm := rules.NewRuleMatch(r, sentence, fromPos, toPos, msg)
			rm.ShortMessage = "Insérer une espace insécable"
			rm.SetSuggestedReplacement(suggestionText)
			ruleMatches = append(ruleMatches, rm)
		}
		prevPrevToken = prevToken
		prevToken = token
	}
	return ruleMatches
}

func (r *QuestionWhitespaceRule) isAllowedWhitespaceChar(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 0 || i >= len(tokens) {
		return false
	}
	if r.Strict {
		// Accept fine/nbsp; also treat non-whitespace as "allowed" so missing-space
		// cases are left to FRENCH_WHITESPACE (Java mutual exclusivity).
		t := tokens[i].GetToken()
		return t == " " || t == " " || !tokens[i].IsWhitespace()
	}
	return tokens[i].IsWhitespace()
}

// frWhitespaceAntiPatterns ports QuestionWhitespaceRule.getAntiPatterns (cached IMMUNIZE).
func frWhitespaceAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	frWhitespaceAntiOnce.Do(func() {
		aps := QuestionWhitespaceAntiPatterns
		frWhitespaceAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			// Java makeAntiPatterns / cacheAntiPatterns: INTERNAL_ANTIPATTERN + IMMUNIZE
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "fr",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			frWhitespaceAntiRules = append(frWhitespaceAntiRules, rule)
		}
	})
	return frWhitespaceAntiRules
}

// getSentenceWithFRWhitespaceImmunization ports Rule.getSentenceWithImmunization
// for QuestionWhitespaceRule.ANTI_PATTERNS.
func getSentenceWithFRWhitespaceImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := frWhitespaceAntiPatterns()
	if len(aps) == 0 {
		return sentence
	}
	src := sentence.GetTokens()
	cloned := make([]*languagetool.AnalyzedTokenReadings, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		cloned[i] = languagetool.NewAnalyzedTokenReadingsFromOld(t, t.GetReadings(), "")
	}
	immunized := languagetool.NewAnalyzedSentence(cloned)
	for _, ap := range aps {
		if ap == nil {
			continue
		}
		immunized = ap.Replace(immunized)
	}
	return immunized
}

// QuestionWhitespaceStrictRule is the Java-name twin of the strict whitespace rule.
// Construct with NewQuestionWhitespaceStrictRule.
type QuestionWhitespaceStrictRule = QuestionWhitespaceRule
