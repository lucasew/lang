package rules

import "fmt"

// DisambiguatedExample ports org.languagetool.tagging.disambiguation.rules.DisambiguatedExample.
type DisambiguatedExample struct {
	Example string
	Input   string // ambiguous forms
	Output  string // disambiguated forms
}

func NewDisambiguatedExample(example string) DisambiguatedExample {
	return DisambiguatedExample{Example: example}
}

func NewDisambiguatedExampleFull(example, input, output string) DisambiguatedExample {
	return DisambiguatedExample{Example: example, Input: input, Output: output}
}

func (e DisambiguatedExample) GetExample() string       { return e.Example }
func (e DisambiguatedExample) GetAmbiguous() string     { return e.Input }
func (e DisambiguatedExample) GetDisambiguated() string { return e.Output }

func (e DisambiguatedExample) String() string {
	return fmt.Sprintf("%s: %s -> %s", e.Example, e.Input, e.Output)
}
