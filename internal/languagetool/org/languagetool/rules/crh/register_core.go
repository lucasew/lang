package crh

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreCrimeanTatarRules ports CrimeanTatar.getRelevantRules.
// Java: CommaWhitespace, DoublePunctuation, Uppercase, MultipleWhitespace,
// SentenceWhitespace, WhiteSpaceBeforeParagraphEnd, WhiteSpaceAtBeginOfParagraph,
// Morfologik — no invent SharedLayout extras (no unpaired brackets, empty line,
// long paragraph, word-repeat, whitespace-before-punct, etc.).
func RegisterCoreCrimeanTatarRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "crh")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	// Java WhiteSpaceBeforeParagraphEnd(messages, this) → defaultActive false → setDefaultOff.
	wpe := rules.NewWhiteSpaceBeforeParagraphEnd(map[string]string{
		"whitespace_before_parapgraph_end_msg": "Don't end a paragraph with whitespace",
	})
	lt.AddTextLevelRuleChecker(wpe.GetID(), rules.AsTextLevelChecker(wpe.MatchList))
	if wpe.IsDefaultOff() {
		lt.MarkDefaultOff(wpe.GetID())
	}

	// Java WhiteSpaceAtBeginOfParagraph(messages) → defaultActive false → setDefaultOff.
	wpb := rules.NewWhiteSpaceAtBeginOfParagraph(map[string]string{
		"whitespace_at_begin_parapgraph_msg": "Don't start a paragraph with whitespace",
	})
	lt.AddRuleChecker(wpb.GetID(), rules.AsSentenceCheckerSimple(wpb.Match))
	if wpb.IsDefaultOff() {
		lt.MarkDefaultOff(wpb.GetID())
	}

	// Java CrimeanTatar.getRelevantRules / MorfologikCrimeanTatarSpellerRule.
	sp := NewMorfologikCrimeanTatarSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
