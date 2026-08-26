package eval

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ConfusionPairEvaluator ports scoring from ConfusionPairEvaluator with injectable check.
// words[0]/words[1] and ruleIDs[0]/ruleIDs[1] form a confusion pair.
type ConfusionPairEvaluator struct {
	Word0, Word1     string
	RuleID0, RuleID1 string
	// Check returns rule IDs that matched the sentence.
	Check func(sentence string) ([]string, error)
	// Tokenize splits a sentence into tokens (defaults to whitespace).
	Tokenize func(sentence string) []string

	// results[wordIdx][0=TP,1=FP,2=TN,3=FN]
	Results [2][4]int
}

const (
	classTP = 0
	classFP = 1
	classTN = 2
	classFN = 3
)

func NewConfusionPairEvaluator(word0, word1, ruleID0, ruleID1 string, check func(string) ([]string, error)) *ConfusionPairEvaluator {
	return &ConfusionPairEvaluator{
		Word0: word0, Word1: word1,
		RuleID0: ruleID0, RuleID1: ruleID1,
		Check: check,
	}
}

func (e *ConfusionPairEvaluator) tokenize(s string) []string {
	if e.Tokenize != nil {
		return e.Tokenize(s)
	}
	return strings.Fields(s)
}

func (e *ConfusionPairEvaluator) containsID(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// AnalyzeSentence ports analyzeSentence(correctSentence, j).
// j is the word index present in the correct sentence (0 or 1).
func (e *ConfusionPairEvaluator) AnalyzeSentence(correctSentence string, j int) error {
	if e.Check == nil {
		return fmt.Errorf("nil Check")
	}
	words := [2]string{e.Word0, e.Word1}
	ruleIDs := [2]string{e.RuleID0, e.RuleID1}

	matchesCorrect, err := e.Check(correctSentence)
	if err != nil {
		return err
	}
	if e.containsID(matchesCorrect, ruleIDs[j]) {
		e.Results[j][classFP]++
	} else {
		e.Results[j][classTN]++
	}

	// swap word j → 1-j
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(words[j]) + `\b`)
	wrongSentence := re.ReplaceAllString(correctSentence, words[1-j])
	matchesWrong, err := e.Check(wrongSentence)
	if err != nil {
		return err
	}
	if e.containsID(matchesWrong, ruleIDs[1-j]) {
		e.Results[1-j][classTP]++
	} else {
		e.Results[1-j][classFN]++
	}
	return nil
}

// ProcessLine sentence-tokenizes a line (one sentence per line for green) and analyzes.
func (e *ConfusionPairEvaluator) ProcessLine(line string) error {
	// green: treat whole line as one sentence
	tokens := e.tokenize(line)
	c0, c1 := 0, 0
	for _, t := range tokens {
		if t == e.Word0 {
			c0++
		}
		if t == e.Word1 {
			c1++
		}
	}
	if c0 > 0 && c1 > 0 {
		return nil // skip mixed
	}
	if c0 == 1 && c1 == 0 {
		return e.AnalyzeSentence(line, 0)
	}
	if c0 == 0 && c1 == 1 {
		return e.AnalyzeSentence(line, 1)
	}
	return nil
}

// ProcessReader processes all lines.
func (e *ConfusionPairEvaluator) ProcessReader(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if err := e.ProcessLine(line); err != nil {
			return err
		}
	}
	return sc.Err()
}

// Precision for rule index i (0 or 1).
func (e *ConfusionPairEvaluator) Precision(i int) float64 {
	tp := float64(e.Results[i][classTP])
	fp := float64(e.Results[i][classFP])
	if tp+fp == 0 {
		return 0
	}
	return tp / (tp + fp)
}

// Recall for rule index i.
func (e *ConfusionPairEvaluator) Recall(i int) float64 {
	tp := float64(e.Results[i][classTP])
	fn := float64(e.Results[i][classFN])
	if tp+fn == 0 {
		return 0
	}
	return tp / (tp + fn)
}
