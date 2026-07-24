package dumpcheck

import (
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const (
	DefaultContextSize = 50
	MarkerStart        = "<err>"
	MarkerEnd          = "</err>"
)

// ResultHandler ports org.languagetool.dev.dumpcheck.ResultHandler.
type ResultHandler struct {
	MaxSentences  int
	MaxErrors     int
	SentenceCount int
	ErrorCount    int
	// Handle is called for each sentence (optional inject for tests).
	Handle func(sentence Sentence, matches []*rules.RuleMatch, langCode string) error
}

func NewResultHandler(maxSentences, maxErrors int) *ResultHandler {
	return &ResultHandler{MaxSentences: maxSentences, MaxErrors: maxErrors}
}

func (h *ResultHandler) CheckMaxSentences() error {
	if h.MaxSentences > 0 && h.SentenceCount >= h.MaxSentences {
		return DocumentLimitReachedError{Limit: h.MaxSentences}
	}
	return nil
}

func (h *ResultHandler) CheckMaxErrors() error {
	if h.MaxErrors > 0 && h.ErrorCount >= h.MaxErrors {
		return ErrorLimitReachedError{Limit: h.MaxErrors}
	}
	return nil
}

// HandleResult increments counters and invokes Handle; enforces limits after.
func (h *ResultHandler) HandleResult(sentence Sentence, matches []*rules.RuleMatch, langCode string) error {
	if h.Handle != nil {
		if err := h.Handle(sentence, matches, langCode); err != nil {
			return err
		}
	}
	h.ErrorCount += len(matches)
	if err := h.CheckMaxErrors(); err != nil {
		return err
	}
	h.SentenceCount++
	return h.CheckMaxSentences()
}

// StdoutHandler ports org.languagetool.dev.dumpcheck.StdoutHandler (writes to W).
type StdoutHandler struct {
	*ResultHandler
	W           io.Writer
	ContextSize int
	Verbose     bool
}

func NewStdoutHandler(w io.Writer, maxSentences, maxErrors, contextSize int) *StdoutHandler {
	if contextSize <= 0 {
		contextSize = DefaultContextSize
	}
	h := &StdoutHandler{
		ResultHandler: NewResultHandler(maxSentences, maxErrors),
		W:             w,
		ContextSize:   contextSize,
	}
	h.Handle = h.printResult
	return h
}

func (h *StdoutHandler) printResult(sentence Sentence, matches []*rules.RuleMatch, langCode string) error {
	if len(matches) == 0 {
		return nil
	}
	fmt.Fprintf(h.W, "\nTitle: %s\n", sentence.GetTitle())
	for i, match := range matches {
		id := ruleIDOf(match)
		fmt.Fprintf(h.W, "%d.) Rule ID: %s\n", i+1, id)
		if match != nil && match.GetMessage() != "" {
			msg := match.GetMessage()
			msg = strings.ReplaceAll(msg, "<suggestion>", "'")
			msg = strings.ReplaceAll(msg, "</suggestion>", "'")
			fmt.Fprintf(h.W, "Message: %s\n", msg)
		}
		if match != nil {
			repls := match.GetSuggestedReplacements()
			if len(repls) > 0 {
				n := len(repls)
				if n > 5 {
					n = 5
				}
				fmt.Fprintf(h.W, "Suggestion: %s\n", strings.Join(repls[:n], "; "))
			}
			fmt.Fprintln(h.W, plainTextContext(match.FromPos, match.ToPos, sentence.GetText(), h.ContextSize))
		}
	}
	return nil
}

func plainTextContext(from, to int, text string, ctx int) string {
	runes := []rune(text)
	if from < 0 {
		from = 0
	}
	if to > len(runes) {
		to = len(runes)
	}
	if from > to {
		from = to
	}
	start := from - ctx
	if start < 0 {
		start = 0
	}
	end := to + ctx
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:from]) + MarkerStart + string(runes[from:to]) + MarkerEnd + string(runes[to:end])
}

type hasID interface{ GetID() string }

func ruleIDOf(match *rules.RuleMatch) string {
	if match == nil {
		return ""
	}
	if h, ok := match.GetRule().(hasID); ok {
		return h.GetID()
	}
	return ""
}
