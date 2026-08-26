package eval

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/dev/errorcorpus"
)

// GoldError is one corpus gold span with optional correction.
type GoldError struct {
	StartPos, EndPos int
	Correction       string
}

// CorpusSentence is one evaluation unit.
type CorpusSentence struct {
	PlainText string
	Errors    []GoldError
}

// RealWordCorpusEvaluator ports scoring from RealWordCorpusEvaluator (injectable Evaluator).
type RealWordCorpusEvaluator struct {
	Evaluator Evaluator

	SentenceCount       int
	ErrorsInCorpusCount int
	PerfectMatches      int // first suggestion perfect
	GoodMatches         int // covers a real error (any suggestion)
	MatchCount          int
}

func NewRealWordCorpusEvaluator(ev Evaluator) *RealWordCorpusEvaluator {
	return &RealWordCorpusEvaluator{Evaluator: ev}
}

func (e *RealWordCorpusEvaluator) Close() error {
	if e.Evaluator != nil {
		return e.Evaluator.Close()
	}
	return nil
}

func (e *RealWordCorpusEvaluator) GetSentencesChecked() int                  { return e.SentenceCount }
func (e *RealWordCorpusEvaluator) GetErrorsChecked() int                     { return e.ErrorsInCorpusCount }
func (e *RealWordCorpusEvaluator) GetRealErrorsFound() int                   { return e.GoodMatches }
func (e *RealWordCorpusEvaluator) GetRealErrorsFoundWithGoodSuggestion() int { return e.PerfectMatches }

// CheckSentence scores one sentence against gold errors.
func (e *RealWordCorpusEvaluator) CheckSentence(s CorpusSentence) error {
	e.SentenceCount++
	e.ErrorsInCorpusCount += len(s.Errors)
	if e.Evaluator == nil {
		return nil
	}
	matches, err := e.Evaluator.Check(s.PlainText)
	if err != nil {
		return err
	}
	for _, m := range matches {
		e.MatchCount++
		ms := Span{StartPos: m.FromPos, EndPos: m.ToPos}
		good := false
		perfect := false
		for _, g := range s.Errors {
			gs := Span{StartPos: g.StartPos, EndPos: g.EndPos}
			if ms.Covers(gs) || gs.Covers(ms) || ms.Overlaps(gs) {
				good = true
				if len(m.SuggestedReplacements) > 0 && g.Correction != "" &&
					m.SuggestedReplacements[0] == g.Correction {
					perfect = true
				}
			}
		}
		if good {
			e.GoodMatches++
		}
		if perfect {
			e.PerfectMatches++
		}
	}
	return nil
}

// FromErrorSentence adapts errorcorpus.ErrorSentence.
func FromErrorSentence(es *errorcorpus.ErrorSentence) CorpusSentence {
	if es == nil {
		return CorpusSentence{}
	}
	out := CorpusSentence{PlainText: es.PlainText}
	for _, er := range es.Errors {
		out.Errors = append(out.Errors, GoldError{
			StartPos:   er.StartPos,
			EndPos:     er.EndPos,
			Correction: er.Correction,
		})
	}
	return out
}

// PrecisionAnySuggestion: goodMatches / matchCount.
func (e *RealWordCorpusEvaluator) PrecisionAnySuggestion() float64 {
	if e.MatchCount == 0 {
		return 0
	}
	return float64(e.GoodMatches) / float64(e.MatchCount)
}

// RecallAnySuggestion: goodMatches / errorsInCorpus (capped style).
func (e *RealWordCorpusEvaluator) RecallAnySuggestion() float64 {
	if e.ErrorsInCorpusCount == 0 {
		return 0
	}
	return float64(e.GoodMatches) / float64(e.ErrorsInCorpusCount)
}

// PrecisionPerfect: perfectMatches / matchCount.
func (e *RealWordCorpusEvaluator) PrecisionPerfect() float64 {
	if e.MatchCount == 0 {
		return 0
	}
	return float64(e.PerfectMatches) / float64(e.MatchCount)
}

// RecallPerfect: perfectMatches / errorsInCorpus.
func (e *RealWordCorpusEvaluator) RecallPerfect() float64 {
	if e.ErrorsInCorpusCount == 0 {
		return 0
	}
	return float64(e.PerfectMatches) / float64(e.ErrorsInCorpusCount)
}

// AnySuggestionPR returns precision/recall for any-suggestion scoring.
func (e *RealWordCorpusEvaluator) AnySuggestionPR() PrecisionRecall {
	return NewPrecisionRecall(e.PrecisionAnySuggestion(), e.RecallAnySuggestion())
}
