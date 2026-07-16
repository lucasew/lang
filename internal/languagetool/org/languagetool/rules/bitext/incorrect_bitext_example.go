package bitext

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/bitext"
)

// IncorrectBitextExample ports org.languagetool.rules.bitext.IncorrectBitextExample.
type IncorrectBitextExample struct {
	Example     bitext.StringPair
	Corrections []string
}

func NewIncorrectBitextExample(example bitext.StringPair) IncorrectBitextExample {
	return IncorrectBitextExample{Example: example}
}

func NewIncorrectBitextExampleWithCorrections(example bitext.StringPair, corrections []string) IncorrectBitextExample {
	return IncorrectBitextExample{
		Example:     example,
		Corrections: append([]string(nil), corrections...),
	}
}

func (e IncorrectBitextExample) GetExample() bitext.StringPair { return e.Example }
func (e IncorrectBitextExample) GetCorrections() []string      { return e.Corrections }

func (e IncorrectBitextExample) String() string {
	return fmt.Sprintf("%s/ %s %v", e.Example.GetSource(), e.Example.GetTarget(), e.Corrections)
}
