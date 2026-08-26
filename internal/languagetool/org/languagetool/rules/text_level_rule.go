package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// TextLevelRule is the Go surface for org.languagetool.rules.TextLevelRule.
// Implementations match across sentences rather than one sentence at a time.
type TextLevelRule interface {
	GetID() string
	GetDescription() string
	// MatchList is the multi-sentence match entry point (Java match(List)).
	MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch
	// MinToCheckParagraph ports minToCheckParagraph (-1 full text, 0 current, >=1 window).
	MinToCheckParagraph() int
}

// TextLevelRuleBase provides defaults shared by text-level rule twins.
type TextLevelRuleBase struct {
	ID          string
	Description string
	// MinParagraphs defaults to -1 (full text) when zero-value used with SetMin.
	MinParagraphs int
	minSet        bool
}

func (b *TextLevelRuleBase) GetID() string          { return b.ID }
func (b *TextLevelRuleBase) GetDescription() string { return b.Description }

func (b *TextLevelRuleBase) SetMinToCheckParagraph(n int) {
	b.MinParagraphs = n
	b.minSet = true
}

func (b *TextLevelRuleBase) MinToCheckParagraph() int {
	if !b.minSet && b.MinParagraphs == 0 {
		// distinguish unset (-1) from explicit 0 via minSet; default -1 like many rules
		return -1
	}
	return b.MinParagraphs
}

// EstimateContextForSureMatch for text-level rules is always -1 in Java.
func (b *TextLevelRuleBase) EstimateContextForSureMatch() int { return -1 }
