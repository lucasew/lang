package rules

import (
	"fmt"
	"strings"
)

// ExampleSentence ports org.languagetool.rules.ExampleSentence.
type ExampleSentence struct {
	Example string
}

// NewExampleSentence validates optional <marker>…</marker> pairing.
func NewExampleSentence(example string) ExampleSentence {
	if example == "" {
		panic("example must not be empty")
	}
	start := strings.Index(example, "<marker>")
	end := strings.Index(example, "</marker>")
	if start != -1 && end == -1 {
		panic(fmt.Sprintf("Example contains <marker> but lacks </marker>:%s", example))
	}
	if start == -1 && end != -1 {
		panic(fmt.Sprintf("Example contains </marker> but lacks <marker>:%s", example))
	}
	if start > end && end != -1 {
		panic(fmt.Sprintf("Example <marker> comes before </marker>:%s", example))
	}
	return ExampleSentence{Example: example}
}

// CleanMarkersInExample strips marker tags.
func CleanMarkersInExample(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, "<marker>", ""), "</marker>", "")
}

func (e ExampleSentence) GetExample() string { return e.Example }
func (e ExampleSentence) String() string     { return e.Example }

// IncorrectExample ports org.languagetool.rules.IncorrectExample.
type IncorrectExample struct {
	ExampleSentence
	Corrections []string
}

// NewIncorrectExample builds an incorrect example; corrections may be empty.
func NewIncorrectExample(example string, corrections ...string) IncorrectExample {
	return IncorrectExample{
		ExampleSentence: NewExampleSentence(example),
		Corrections:     append([]string(nil), corrections...),
	}
}

func (e IncorrectExample) GetCorrections() []string {
	if e.Corrections == nil {
		return nil
	}
	return append([]string(nil), e.Corrections...)
}

func (e IncorrectExample) String() string {
	return e.Example + " " + fmt.Sprint(e.GetCorrections())
}

// CorrectExample ports org.languagetool.rules.CorrectExample.
type CorrectExample struct {
	ExampleSentence
}

func NewCorrectExample(example string) CorrectExample {
	return CorrectExample{ExampleSentence: NewExampleSentence(example)}
}

// ErrorTriggeringExample ports org.languagetool.rules.ErrorTriggeringExample.
type ErrorTriggeringExample struct {
	ExampleSentence
}

func NewErrorTriggeringExample(example string) ErrorTriggeringExample {
	return ErrorTriggeringExample{ExampleSentence: NewExampleSentence(example)}
}

// Wrong creates an IncorrectExample that must contain <marker>…</marker>.
func Wrong(example string) IncorrectExample {
	if !strings.Contains(example, "<marker>") || !strings.Contains(example, "</marker>") {
		panic("Example text must contain '<marker>...</marker>': " + example)
	}
	return NewIncorrectExample(example)
}

// Fixed creates a CorrectExample (markers optional).
func Fixed(example string) CorrectExample {
	return NewCorrectExample(example)
}

// Example is the Java-name twin for building example sentences (Example.wrong / Example.fixed).
type Example struct{}

// Wrong wraps a wrong-example string.
func (Example) Wrong(s string) ExampleSentence { return NewExampleSentence(s) }

// Fixed wraps a fixed-example string.
func (Example) Fixed(s string) ExampleSentence { return NewExampleSentence(s) }
