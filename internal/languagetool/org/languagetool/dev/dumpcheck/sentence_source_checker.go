package dumpcheck

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceChecker is a pluggable LT stand-in for dump checking.
type SentenceChecker func(text, langCode string) []*rules.RuleMatch

// SentenceSourceChecker ports org.languagetool.dev.dumpcheck.SentenceSourceChecker run loop
// without CLI/database (green vertical slice).
type SentenceSourceChecker struct {
	LangCode string
	Checker  SentenceChecker
	Handler  *ResultHandler
}

func NewSentenceSourceChecker(langCode string, checker SentenceChecker, handler *ResultHandler) *SentenceSourceChecker {
	return &SentenceSourceChecker{LangCode: langCode, Checker: checker, Handler: handler}
}

// Run drains source and handles results until limits or exhaustion.
func (c *SentenceSourceChecker) Run(source SentenceSource) error {
	if c == nil || source == nil {
		return nil
	}
	for source.HasNext() {
		sent, err := source.Next()
		if err != nil {
			return err
		}
		var matches []*rules.RuleMatch
		if c.Checker != nil {
			matches = c.Checker(sent.GetText(), c.LangCode)
		}
		if c.Handler != nil {
			if err := c.Handler.HandleResult(sent, matches, c.LangCode); err != nil {
				return err
			}
		}
	}
	return nil
}
