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
// Java: Objects.requireNonNull(example) — empty string is allowed; only null rejected.
func NewExampleSentence(example string) ExampleSentence {
	// Go has no null string; empty is valid (Java requireNonNull allows "").
	start := strings.Index(example, "<marker>")
	end := strings.Index(example, "</marker>")
	if start != -1 && end == -1 {
		panic(fmt.Sprintf("Example contains <marker> but lacks </marker>:%s", example))
	}
	if start == -1 && end != -1 {
		panic(fmt.Sprintf("Example contains </marker> but lacks <marker>:%s", example))
	}
	// Java: if (markerStart > markerEnd) — after null checks for missing markers.
	if start > end {
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
// Java class has private ctor; Go zero value is fine.
type Example struct{}

// Wrong ports Example.wrong — requires <marker>…</marker>, returns IncorrectExample.
func (Example) Wrong(s string) IncorrectExample { return Wrong(s) }

// Fixed ports Example.fixed — returns CorrectExample.
func (Example) Fixed(s string) CorrectExample { return Fixed(s) }
