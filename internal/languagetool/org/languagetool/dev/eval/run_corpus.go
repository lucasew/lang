package eval

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/dev/errorcorpus"
)

// RunPedlerCorpus drains a PedlerCorpus into the evaluator.
func (e *RealWordCorpusEvaluator) RunPedlerCorpus(c *errorcorpus.PedlerCorpus) error {
	if e == nil || c == nil {
		return nil
	}
	for c.HasNext() {
		es := c.Next()
		if es == nil {
			break
		}
		cs := CorpusSentence{
			PlainText: es.PlainText,
			Errors:    remapPedlerErrors(es),
		}
		if err := e.CheckSentence(cs); err != nil {
			return err
		}
	}
	return nil
}

// RunSimpleCorpus drains a SimpleCorpus into the evaluator.
func (e *RealWordCorpusEvaluator) RunSimpleCorpus(c *errorcorpus.SimpleCorpus) error {
	if e == nil || c == nil {
		return nil
	}
	for c.HasNext() {
		es, err := c.Next()
		if err != nil {
			return err
		}
		cs := CorpusSentence{
			PlainText: es.PlainText,
			Errors:    remapSimpleErrors(es),
		}
		if err := e.CheckSentence(cs); err != nil {
			return err
		}
	}
	return nil
}

var pedlerERRRE = regexp.MustCompile(`(?i)<ERR\s+targ=([^>]*)>(.*?)</ERR>`)

func remapPedlerErrors(es *errorcorpus.ErrorSentence) []GoldError {
	if es == nil {
		return nil
	}
	var out []GoldError
	for _, m := range pedlerERRRE.FindAllStringSubmatch(es.MarkupText, -1) {
		corr := strings.TrimSpace(m[1])
		surface := m[2]
		idx := strings.Index(es.PlainText, surface)
		if idx < 0 {
			continue
		}
		out = append(out, GoldError{StartPos: idx, EndPos: idx + len(surface), Correction: corr})
	}
	if len(out) == 0 {
		for _, er := range es.Errors {
			out = append(out, GoldError{StartPos: er.StartPos, EndPos: er.EndPos, Correction: er.Correction})
		}
	}
	return out
}

func remapSimpleErrors(es *errorcorpus.ErrorSentence) []GoldError {
	if es == nil {
		return nil
	}
	m := es.MarkupText
	start := strings.Index(m, "_")
	if start < 0 {
		return fromErrors(es.Errors)
	}
	endRel := strings.Index(m[start+1:], "_")
	if endRel < 0 {
		return fromErrors(es.Errors)
	}
	end := start + 1 + endRel
	surface := m[start+1 : end]
	idx := strings.Index(es.PlainText, surface)
	if idx < 0 {
		return fromErrors(es.Errors)
	}
	corr := ""
	if len(es.Errors) > 0 {
		corr = es.Errors[0].Correction
	}
	return []GoldError{{StartPos: idx, EndPos: idx + len(surface), Correction: corr}}
}

func fromErrors(errs []errorcorpus.Error) []GoldError {
	var out []GoldError
	for _, er := range errs {
		out = append(out, GoldError{StartPos: er.StartPos, EndPos: er.EndPos, Correction: er.Correction})
	}
	return out
}
