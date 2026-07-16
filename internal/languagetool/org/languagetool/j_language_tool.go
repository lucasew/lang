package languagetool

// Constants and enums from org.languagetool.JLanguageTool.

const (
	SentenceStartTagName = "SENT_START"
	SentenceEndTagName   = "SENT_END"
	ParagraphEndTagName  = "PARA_END"

	PatternFile                 = "grammar.xml"
	StyleFile                   = "style.xml"
	CustomPatternFile           = "grammar_custom.xml"
	FalseFriendFile             = "false-friends.xml"
	MessageBundleName = "org.languagetool.MessagesBundle"
	DictionaryFilenameExtension = ".dict"
)

// Mode ports JLanguageTool.Mode.
type Mode string

const (
	ModeAll             Mode = "ALL"
	ModeTextLevelOnly   Mode = "TEXTLEVEL_ONLY"
	ModeAllButTextLevel Mode = "ALL_BUT_TEXTLEVEL_ONLY"
)

// ParagraphHandling ports JLanguageTool.ParagraphHandling.
type ParagraphHandling string

const (
	ParagraphNormal      ParagraphHandling = "NORMAL"
	ParagraphOnlyPara    ParagraphHandling = "ONLYPARA"
	ParagraphOnlyNonPara ParagraphHandling = "ONLYNONPARA"
)

// CheckCancelledCallback ports JLanguageTool.CheckCancelledCallback.
type CheckCancelledCallback func() bool

// JLanguageTool is a minimal façade for pure-Go check orchestration (growing).
// Full Java parity is not attempted here.
type JLanguageTool struct {
	LanguageCode string
	Mode         Mode
	Level        Level
	// Rules registered for sentence-level matching (any rule with Match method deferred).
	// Matchers implement Match(sentence) ([]*rules.RuleMatch, error) via RuleMatcher-like surface.
	sentenceMatchers []func(sentence *AnalyzedSentence) error
}

func NewJLanguageTool(languageCode string) *JLanguageTool {
	return &JLanguageTool{
		LanguageCode: languageCode,
		Mode:         ModeAll,
		Level:        LevelDefault,
	}
}

func (lt *JLanguageTool) GetLanguageCode() string { return lt.LanguageCode }
func (lt *JLanguageTool) GetMode() Mode           { return lt.Mode }
func (lt *JLanguageTool) SetMode(m Mode)          { lt.Mode = m }
func (lt *JLanguageTool) GetLevel() Level         { return lt.Level }
func (lt *JLanguageTool) SetLevel(l Level)        { lt.Level = l }
