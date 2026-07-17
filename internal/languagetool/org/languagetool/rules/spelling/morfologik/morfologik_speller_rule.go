package morfologik

import (
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
)

// MorfologikSpellerRule ports org.languagetool.rules.spelling.morfologik.MorfologikSpellerRule
// (map/dict-backed; binary morfologik deferred).
type MorfologikSpellerRule struct {
	*spelling.SpellingCheckRule
	Speller           *MorfologikSpeller
	IgnoreTaggedWords bool
	// FileName is the dictionary path from getFileName().
	FileName string
}

func NewMorfologikSpellerRule(id, languageCode, fileName string, speller *MorfologikSpeller) *MorfologikSpellerRule {
	if speller == nil {
		speller = NewMorfologikSpeller(fileName, 1)
	}
	r := &MorfologikSpellerRule{
		SpellingCheckRule: spelling.NewSpellingCheckRule(id, "Possible spelling mistake", languageCode),
		Speller:           speller,
		FileName:          fileName,
	}
	r.IsMisspelled = r.Speller.IsMisspelled
	return r
}

func (r *MorfologikSpellerRule) GetFileName() string { return r.FileName }

// Match flags misspelled tokens.
func (r *MorfologikSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	var out []*rules.RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() || tok.IsSentenceEnd() {
			continue
		}
		if tok.IsIgnoredBySpeller() || tok.IsImmunized() {
			continue
		}
		w := tok.GetToken()
		if w == "" || !hasLetter(w) {
			continue
		}
		if r.IgnoreTaggedWords && tok.IsTagged() {
			continue
		}
		if r.AcceptWord(w) {
			continue
		}
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
			"Possible spelling mistake found")
		if sug := r.Speller.FindReplacements(w); len(sug) > 0 {
			m.SetSuggestedReplacements(sug)
		}
		out = append(out, m)
	}
	return out, nil
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
