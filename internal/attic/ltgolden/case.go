package ltgolden

// Kind classifies ground-truth cases by LT source.
type Kind string

const (
	KindGrammarExample  Kind = "grammar_example"
	KindDisambigExample Kind = "disambig_example"
	KindJavaUnit        Kind = "java_unit"
)

// Case is one LT ground-truth expectation.
type Case struct {
	Kind        Kind
	Lang        string // short code / family, e.g. en, de
	RuleID      string
	RuleDefault string // on/off/temp_off from XML (empty if N/A)
	Text        string
	Incorrect   bool   // expect a hit for RuleID
	Correction  string // LT correction attribute (gold suggestion text, may be multi)
	HasMarker   bool
	MarkerFrom  int
	MarkerTo    int
	SourceFile  string
	ExampleType string // XML type attr or Java method name
	// Java-specific
	JavaClass  string
	JavaMethod string
}
