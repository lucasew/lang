package tokenizers

import (
	_ "embed"
	"sync"

	"github.com/lucasew/lang/internal/attic/srx"
)

//go:embed data/segment-simple.srx
var segmentSimpleSRX []byte

// SimpleSentenceTokenizer ports org.languagetool.tokenizers.SimpleSentenceTokenizer.
//
// Java:
//
//	public class SimpleSentenceTokenizer extends SRXSentenceTokenizer {
//	  public SimpleSentenceTokenizer() {
//	    super(new AnyLanguage(), "/org/languagetool/tokenizers/segment-simple.srx");
//	  }
//	  // AnyLanguage.getShortCode() == "xx"
//	}
//
// Behavior (paragraph flags, Tokenize) is inherited from SRXSentenceTokenizer.
// Resource is official segment-simple.srx (embedded; byte-identical to
// inspiration/.../resource/org/languagetool/tokenizers/segment-simple.srx).
type SimpleSentenceTokenizer struct {
	*SRXSentenceTokenizer
}

// NewSimpleSentenceTokenizer ports SimpleSentenceTokenizer().
func NewSimpleSentenceTokenizer() *SimpleSentenceTokenizer {
	// Java: super(new AnyLanguage(), "/org/languagetool/tokenizers/segment-simple.srx")
	// AnyLanguage.getShortCode() == "xx"
	inner := NewSRXSentenceTokenizerWithPath("xx", "/org/languagetool/tokenizers/segment-simple.srx")
	return &SimpleSentenceTokenizer{SRXSentenceTokenizer: inner}
}

// AsSentenceTokenizer returns this value as SentenceTokenizer.
// Java SimpleSentenceTokenizer is-a SentenceTokenizer via extends; Go embedding
// promotes SRXSentenceTokenizer methods so *SimpleSentenceTokenizer satisfies
// the interface directly.
func (t *SimpleSentenceTokenizer) AsSentenceTokenizer() SentenceTokenizer {
	return t
}

// Ensure SimpleSentenceTokenizer is a SentenceTokenizer (Java inheritance).
var _ SentenceTokenizer = (*SimpleSentenceTokenizer)(nil)

var (
	simpleDocOnce sync.Once
	simpleDoc     *srx.Document
	simpleDocErr  error
)

// segmentSimpleDocument loads the embedded official segment-simple.srx once
// (SrxTools.createSrxDocument for "/org/languagetool/tokenizers/segment-simple.srx").
// Java: JLanguageTool.getDataBroker().getFromResourceDirAsStream(path).
func segmentSimpleDocument() (*srx.Document, error) {
	simpleDocOnce.Do(func() {
		name, err := materializeEmbed("segment-simple", segmentSimpleSRX)
		if err != nil {
			simpleDocErr = err
			return
		}
		simpleDoc, simpleDocErr = srx.Load(name)
	})
	return simpleDoc, simpleDocErr
}
