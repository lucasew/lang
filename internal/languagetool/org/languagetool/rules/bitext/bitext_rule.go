package bitext

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// BitextRule ports org.languagetool.rules.bitext.BitextRule.
type BitextRule interface {
	GetID() string
	GetDescription() string
	GetMessage() string
	// MatchBitext checks source/target pair (Java match(src, trg)).
	MatchBitext(source, target *languagetool.AnalyzedSentence) []*rules.RuleMatch
}

// BitextRuleBase holds shared fields.
type BitextRuleBase struct {
	ID          string
	Description string
	Message     string
	IssueType   string
	SourceLang  string
}

func (b *BitextRuleBase) GetID() string                 { return b.ID }
func (b *BitextRuleBase) GetDescription() string        { return b.Description }
func (b *BitextRuleBase) GetMessage() string            { return b.Message }
func (b *BitextRuleBase) SetSourceLanguage(code string) { b.SourceLang = code }
func (b *BitextRuleBase) GetSourceLanguage() string     { return b.SourceLang }

// RelevantBitextRules ports BitextRule.getRelevantRules factory list.
func RelevantBitextRules() []BitextRule {
	return []BitextRule{
		NewDifferentLengthRule(),
		NewSameTranslationRule(),
		NewDifferentPunctuationRule(),
	}
}

func targetSpan(target *languagetool.AnalyzedSentence) (from, to int) {
	if target == nil {
		return 0, 0
	}
	toks := target.GetTokens()
	if len(toks) == 0 {
		return 0, 0
	}
	last := toks[len(toks)-1]
	return 0, last.GetStartPos() + len([]rune(last.GetToken())) // approximate; use GetEndPos
}

func targetEndPos(target *languagetool.AnalyzedSentence) int {
	toks := target.GetTokens()
	if len(toks) == 0 {
		return 0
	}
	return toks[len(toks)-1].GetEndPos()
}
