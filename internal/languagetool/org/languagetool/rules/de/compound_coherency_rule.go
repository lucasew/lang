package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CompoundCoherencyRule ports org.languagetool.rules.de.CompoundCoherencyRule.
// Uses reading lemmas when hasSameLemmas + readings exist; otherwise surface token
// (Java: lemmaOrNull != null ? lemmaOrNull : token).
// Java: STYLE category.
type CompoundCoherencyRule struct {
	Messages map[string]string
	Category *rules.Category
}

func NewCompoundCoherencyRule(messages map[string]string) *CompoundCoherencyRule {
	return &CompoundCoherencyRule{
		Messages: messages,
		Category: rules.CatStyle.GetCategory(messages),
	}
}

func (r *CompoundCoherencyRule) GetID() string { return "DE_COMPOUND_COHERENCY" }

// GetDescription ports CompoundCoherencyRule.getDescription.
func (r *CompoundCoherencyRule) GetDescription() string {
	return "Einheitliche Schreibweise bei Komposita (mit oder ohne Bindestrich)"
}

func (r *CompoundCoherencyRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns -1).
func (r *CompoundCoherencyRule) MinToCheckParagraph() int { return -1 }

func containsHyphenInside(token string) bool {
	return strings.Contains(token, "-") && !strings.HasPrefix(token, "-") && !strings.HasSuffix(token, "-")
}

func isNumericToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// compoundCoherencyLemma ports CompoundCoherencyRule.getLemma.
// Java: lemmaOrNull = atr.hasSameLemmas() && readingsLength > 0 ? readings.get(0).getLemma() : null
func compoundCoherencyLemma(atr *languagetool.AnalyzedTokenReadings) string {
	if atr == nil || !atr.HasSameLemmas() || atr.GetReadingsLength() == 0 {
		return ""
	}
	first := atr.GetReadings()[0].GetLemma()
	if first == nil {
		return ""
	}
	lemmaOrNull := *first
	token := atr.GetToken()
	// Java charAt indices; use runes for DE letters (äöüß) so indices stay aligned with graphemes.
	if !strings.Contains(lemmaOrNull, "-") && strings.Contains(token, "-") {
		var b strings.Builder
		lemmaRunes := []rune(lemmaOrNull)
		tokenRunes := []rune(token)
		for lemmaPos, tokenPos := 0, 0; lemmaPos < len(lemmaRunes); lemmaPos, tokenPos = lemmaPos+1, tokenPos+1 {
			if tokenPos >= len(tokenRunes) {
				break
			}
			lemmaChar := lemmaRunes[lemmaPos]
			tokenChar := tokenRunes[tokenPos]
			if lemmaChar == tokenChar {
				b.WriteRune(lemmaChar)
			} else if tokenChar == '-' {
				tokenPos++ // skip hyphen; for-loop still increments tokenPos (Java twin)
				b.WriteByte('-')
				// Java: lemmaPos + 1 < token.length() && isUpperCase(token.charAt(tokenPos))
				if lemmaPos+1 < len(tokenRunes) && tokenPos < len(tokenRunes) && unicode.IsUpper(tokenRunes[tokenPos]) {
					b.WriteRune(unicode.ToUpper(lemmaChar))
				} else {
					b.WriteRune(lemmaChar)
				}
			}
			// else: Java appends nothing on other mismatches
		}
		return b.String()
	}
	return lemmaOrNull
}

func (r *CompoundCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	normToText := map[string][]string{}
	pos := 0
	for _, sentence := range sentences {
		for _, atr := range sentence.GetTokensWithoutWhitespace() {
			token := atr.GetToken()
			if token == "" {
				continue
			}
			lemmaOrNull := compoundCoherencyLemma(atr)
			lemma := token
			if lemmaOrNull != "" {
				lemma = lemmaOrNull
			}
			normToken := strings.ToLower(strings.ReplaceAll(lemma, "-", ""))
			if isNumericToken(normToken) {
				break
			}
			if occ, ok := normToText[normToken]; ok {
				foundSame := false
				for _, f := range occ {
					if strings.EqualFold(f, lemma) {
						foundSame = true
						break
					}
				}
				if !foundSame {
					other := occ[0]
					if containsHyphenInside(other) || containsHyphenInside(token) {
						msg := "Uneinheitliche Verwendung von Bindestrichen. Der Text enthält sowohl '" +
							token + "' als auch '" + other + "'."
						rm := rules.NewRuleMatch(r, sentence, pos+atr.GetStartPos(), pos+atr.GetEndPos(), msg)
						if strings.EqualFold(strings.ReplaceAll(token, "-", ""), strings.ReplaceAll(other, "-", "")) {
							rm.SetSuggestedReplacement(other)
						}
						ruleMatches = append(ruleMatches, rm)
					}
				}
			} else {
				normToText[normToken] = []string{lemma}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
