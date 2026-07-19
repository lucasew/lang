package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleMeta_KnownJavaFamilies(t *testing.T) {
	id, name, issue, short := RuleMeta("EN_A_VS_AN")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammar", name)
	require.Equal(t, "grammar", issue)
	require.NotEmpty(t, short)

	id, _, issue, _ = RuleMeta("MORFOLOGIK_RULE_EN_US")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "misspelling", issue)

	// Java grammar.xml category MULTITOKEN_SPELLING (before generic SPELL→TYPOS)
	// Java de grammar.xml DE_MULTITOKEN_SPELLING_{TWO,THREE,FOUR} share MULTITOKEN_SPELLING
	for _, rid := range []string{
		"DE_MULTITOKEN_SPELLING_TWO",
		"DE_MULTITOKEN_SPELLING_THREE",
		"DE_MULTITOKEN_SPELLING_FOUR",
	} {
		id, name, issue, short = RuleMeta(rid)
		require.Equal(t, "MULTITOKEN_SPELLING", id, rid)
		require.Equal(t, "Rechtschreibfehler", name, rid)
		require.Equal(t, "misspelling", issue, rid)
		require.Equal(t, "Möglicher Fehler", short, rid)
		require.Equal(t, "Rechtschreibfehler in Eigennamen", RuleDescription(rid), rid)
		require.Equal(t, "de", RuleLangHint(rid), rid)
	}
	id, name, issue, short = RuleMeta("EN_MULTITOKEN_SPELLING_FOUR")
	require.Equal(t, "MULTITOKEN_SPELLING", id)
	require.Equal(t, "Orthographic errors", name)
	require.Equal(t, "misspelling", issue)

	id, _, issue, _ = RuleMeta("WHITESPACE_RULE")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "whitespace", issue)

	// Java CommaWhitespaceRule / GermanCommaWhitespaceRule: ID COMMA_PARENTHESIS_WHITESPACE
	// (no "WHITESPACE" substring) — Categories.TYPOGRAPHY + ITS Whitespace
	id, name, issue, short = RuleMeta("COMMA_PARENTHESIS_WHITESPACE")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "Typografie", name)
	require.Equal(t, "whitespace", issue)
	require.Equal(t, "Typografie", short)
	require.Equal(t, "Leerzeichen vor/hinter Kommas und Klammern",
		RuleDescription("COMMA_PARENTHESIS_WHITESPACE"))
	// Java WhitespaceBeforePunctuationRule — MessagesBundle_de desc
	require.Equal(t, "Leerzeichen vor Doppelpunkt, Semikolon oder Prozentzeichen.",
		RuleDescription("WHITESPACE_PUNCTUATION"))
	// Java MultipleWhitespaceRule — MessagesBundle_de desc_whitespacerepetition
	require.Equal(t, "Wiederholung von Leerzeichen", RuleDescription("WHITESPACE_RULE"))
	// Java MessagesBundle_de empty_line_rule_desc / desc_uppercase_sentence
	require.Equal(t, "Leere Zeile", RuleDescription("EMPTY_LINE"))
	require.Equal(t, "Großschreibung am Satzanfang", RuleDescription("UPPERCASE_SENTENCE_START"))
	// Java GermanUnpairedBracketsRule ID UNPAIRED_BRACKETS; MessagesBundle_de
	require.Equal(t, "Unpaarige Anführungszeichen und Klammern",
		RuleDescription("UNPAIRED_BRACKETS"))
	require.Equal(t, "Unpaarige Anführungszeichen", RuleDescription("DE_UNPAIRED_QUOTES"))

	// Soft invent IDs must not get special grammar/style invent — uncategorized.
	id, _, issue, _ = RuleMeta("EN_SOFT_YOUR_YOU_RE")
	require.Equal(t, "MISC", id)
	require.Equal(t, "uncategorized", issue)

	id, name, issue, short = RuleMeta("EMPTY_LINE")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Leere Zeile", short)

	// Java GermanWordRepeatRule: REDUNDANCY (not base WordRepeatRule MISC)
	id, name, issue, short = RuleMeta("GERMAN_WORD_REPEAT_RULE")
	require.Equal(t, "REDUNDANCY", id)
	require.Equal(t, "Redundanz", name) // MessagesBundle_de category_redundancy
	require.Equal(t, "duplication", issue)
	require.Equal(t, "Wortwiederholung", short) // MessagesBundle_de desc_repetition_short

	// Java WordRepeatBeginningRule: REPETITIONS_STYLE
	id, name, issue, short = RuleMeta("GERMAN_WORD_REPEAT_BEGINNING_RULE")
	require.Equal(t, "REPETITIONS_STYLE", id)
	require.Equal(t, "Wiederholungen", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Wortwiederholung am Satzanfang", short)

	// Java DE speller soft short MessagesBundle_de desc_spelling_short
	// AustrianGermanSpellerRule / SwissGermanSpellerRule share DE speller soft (Java).
	for _, rid := range []string{
		"GERMAN_SPELLER_RULE", "AUSTRIAN_GERMAN_SPELLER_RULE", "SWISS_GERMAN_SPELLER_RULE",
		"MORFOLOGIK_RULE_DE_DE",
	} {
		id, name, issue, short = RuleMeta(rid)
		require.Equal(t, "TYPOS", id, rid)
		require.Equal(t, "Mögliche Tippfehler", name, rid)
		require.Equal(t, "misspelling", issue, rid)
		require.Equal(t, "Rechtschreibfehler", short, rid)
		require.Equal(t, "Möglicher Rechtschreibfehler", RuleDescription(rid), rid)
	}
	_, _, _, short = RuleMeta("MORFOLOGIK_RULE_EN_US")
	require.Equal(t, "Spelling mistake", short)

	// Base EN word-repeat stays MISC
	id, _, issue, _ = RuleMeta("WORD_REPEAT_RULE")
	require.Equal(t, "MISC", id)
	require.Equal(t, "duplication", issue)

	// DE agreement / case / compounds (Java Categories)
	id, name, issue, short = RuleMeta("DE_AGREEMENT")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammatik", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Kongruenz", short)
	// Java AgreementRule2 / SubjectVerbAgreement / VerbAgreement share GRAMMAR meta
	id, name, issue, short = RuleMeta("DE_AGREEMENT2")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Kongruenz", short)
	id, name, issue, short = RuleMeta("DE_SUBJECT_VERB_AGREEMENT")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Kongruenz", short)
	id, name, issue, short = RuleMeta("DE_VERBAGREEMENT")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Kongruenz", short)
	id, name, issue, short = RuleMeta("DE_CASE")
	require.Equal(t, "CASING", id)
	require.Equal(t, "Groß-/Kleinschreibung", name)
	require.Equal(t, "Großschreibung", short)
	id, name, issue, short = RuleMeta("DE_UPPER_CASE_NGRAM")
	require.Equal(t, "CASING", id)
	require.Equal(t, "Groß-/Kleinschreibung", name)
	require.Equal(t, "typographical", issue)
	require.Equal(t, "Großschreibung", short)
	id, name, issue, short = RuleMeta("DE_COMPOUNDS")
	require.Equal(t, "COMPOUNDING", id)
	require.Equal(t, "Getrennt- und Zusammenschreibung", name)
	require.Equal(t, "Komposita", short)
	id, _, issue, short = RuleMeta("DE_SENTENCE_WHITESPACE")
	require.Equal(t, "MISC", id)
	require.Equal(t, "Leerzeichen zwischen Sätzen", short)
	// Java shared layout IDs (CommaWhitespace / SentenceWhitespace / WhitespaceBeforePunctuation)
	// DE soft typography for DE-registered layout pack IDs; generic SENTENCE_WHITESPACE stays EN labels.
	id, name, issue, short = RuleMeta("WHITESPACE_PUNCTUATION")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "Typografie", name)
	require.Equal(t, "whitespace", issue)
	require.Equal(t, "Typografie", short)
	id, name, issue, short = RuleMeta("COMMA_WHITESPACE")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "Typography", name) // not in DE soft-id short-list
	require.Equal(t, "whitespace", issue)
	id, name, issue, short = RuleMeta("SENTENCE_WHITESPACE")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "Typography", name)
	require.Equal(t, "Whitespace", RuleDescription("SENTENCE_WHITESPACE"))
	// Java UppercaseSentenceStartRule: Categories.CASING (MessagesBundle_de category_case when under de)
	id, name, issue, short = RuleMeta("UPPERCASE_SENTENCE_START")
	require.Equal(t, "CASING", id)
	require.Equal(t, "Groß-/Kleinschreibung", name)
	require.Equal(t, "typographical", issue)
	require.Equal(t, "Großschreibung", short)
	// Java MissingCommaRelativeClauseRule behind-id + PassiveativeSentence + unpaired quotes
	id, name, issue, short = RuleMeta("COMMA_BEHIND_RELATIVE_CLAUSE")
	require.Equal(t, "HILFESTELLUNG_KOMMASETZUNG", id)
	require.Equal(t, "Hilfestellung Kommasetzung", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Fehlendes Komma", short)
	id, name, issue, short = RuleMeta("PASSIVE_SENTENCE_DE")
	require.Equal(t, "CREATIVE_WRITING", id)
	require.Equal(t, "Stiltipps für kreatives Schreiben", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Stil", short)
	id, name, issue, short = RuleMeta("DE_UNPAIRED_QUOTES")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "Typografie", name)
	require.Equal(t, "typographical", issue)
	require.Equal(t, "Unpaarige Zeichen", short)
	id, name, issue, short = RuleMeta("UNPAIRED_BRACKETS")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "Typografie", name)
	require.Equal(t, "Unpaarige Zeichen", short)

	// More DE Java categories (RuleMeta fallback only)
	id, name, issue, short = RuleMeta("DE_DASH")
	require.Equal(t, "COMPOUNDING", id)
	require.Equal(t, "Getrennt- und Zusammenschreibung", name)
	require.Equal(t, "Komposita", short)
	id, _, issue, _ = RuleMeta("MISSING_VERB")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "grammar", issue)
	id, name, issue, short = RuleMeta("OLD_SPELLING_RULE")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "Mögliche Tippfehler", name)
	require.Equal(t, "misspelling", issue)
	require.Equal(t, "Alte Rechtschreibung", short)
	id, name, issue, short = RuleMeta("DE_WIEDER_VS_WIDER")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "Mögliche Tippfehler", name)
	require.Equal(t, "Rechtschreibfehler", short)
	// Java SimilarNameRule / AbstractSimpleReplaceRule / RedundantModal / GermanStyleRepeatedWord
	id, name, issue, short = RuleMeta("DE_SIMILAR_NAMES")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "Mögliche Tippfehler", name)
	require.Equal(t, "misspelling", issue)
	require.Equal(t, "Rechtschreibfehler", short)
	id, name, issue, short = RuleMeta("DE_SIMPLE_REPLACE")
	require.Equal(t, "MISC", id)
	require.Equal(t, "Sonstiges", name)
	require.Equal(t, "misspelling", issue)
	require.Equal(t, "Falsches Wort", short)
	id, name, issue, short = RuleMeta("REDUNDANT_MODAL_VERB")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Stil", short)
	id, name, issue, short = RuleMeta("STYLE_REPEATED_WORD_RULE_DE")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Stil", short)
	id, name, issue, short = RuleMeta("DE_CH_COMPOUNDS")
	require.Equal(t, "COMPOUNDING", id)
	require.Equal(t, "Komposita", short)
	id, name, issue, short = RuleMeta("DE_COMPOUND_COHERENCY")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Stil", short)
	id, name, issue, short = RuleMeta("DE_WORD_COHERENCY")
	require.Equal(t, "MISC", id)
	require.Equal(t, "Sonstiges", name)
	require.Equal(t, "Uneinheitliche Schreibweise", short)
	id, name, issue, short = RuleMeta("DE_CONFUSION_RULE")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "Mögliche Tippfehler", name)
	require.Equal(t, "Verwechslung", short)

	// Double punctuation is PUNCTUATION (not TYPOGRAPHY)
	id, _, issue, short = RuleMeta("DE_DOUBLE_PUNCTUATION")
	require.Equal(t, "PUNCTUATION", id)
	require.Equal(t, "typographical", issue)
	require.Equal(t, "Doppelte Satzzeichen", short)
	id, _, issue, short = RuleMeta("DOUBLE_PUNCTUATION")
	require.Equal(t, "PUNCTUATION", id)
	require.Equal(t, "Double punctuation", short)

	// Java WhiteSpaceBeforeParagraphEnd / WhiteSpaceAtBeginOfParagraph: STYLE (not TYPOGRAPHY)
	id, _, issue, short = RuleMeta("WHITESPACE_PARAGRAPH")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "style", issue)
	require.Equal(t, "Leerzeichen am Absatzende", short)
	id, _, issue, short = RuleMeta("WHITESPACE_PARAGRAPH_BEGIN")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "style", issue)
	require.Equal(t, "Leerzeichen am Anfang des Absatzes", short)
	// Generic multiple-whitespace still TYPOGRAPHY
	id, _, issue, _ = RuleMeta("WHITESPACE_RULE")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "whitespace", issue)

	// Java LongParagraphRule ID TOO_LONG_PARAGRAPH must not hit LongSentence TOO_LONG path
	id, name, issue, short = RuleMeta("TOO_LONG_PARAGRAPH")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Langer Absatz", short)
	require.Equal(t, "Lesbarkeit: Absatz mit mehr als {0} Wörtern", RuleDescription("TOO_LONG_PARAGRAPH"))

	// Java PunctuationMarkAtParagraphEnd: PUNCTUATION + Grammar
	id, name, issue, short = RuleMeta("PUNCTUATION_PARAGRAPH_END")
	require.Equal(t, "PUNCTUATION", id)
	require.Equal(t, "Zeichensetzung", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Absatzende", short)

	// Java German.mergeMatches specificRuleId
	id, name, issue, short = RuleMeta("AI_DE_MERGED_MATCH")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammatik", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Potenzieller Fehler", short)
	require.Equal(t, "Hier scheint es einen Fehler zu geben.", RuleDescription("AI_DE_MERGED_MATCH"))
	// Java French.mergeMatches SoftRule fallback (AI_FR_MERGED_MATCH[_STYLE][_PICKY])
	id, name, issue, short = RuleMeta("AI_FR_MERGED_MATCH")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammaire", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Erreur potentielle", short)
	require.Equal(t, "Il pourrait y avoir un problème ici.", RuleDescription("AI_FR_MERGED_MATCH_STYLE_PICKY"))
	require.Equal(t, "fr", RuleLangHint("AI_FR_GGEC_X"))
	require.Equal(t, "fr", RuleLangHint("AI_FR_MERGED_MATCH"))
	// Java Spanish AI_ES SoftRule / LangHint
	id, name, issue, short = RuleMeta("AI_ES_GGEC_REPLACEMENT_CASING_X")
	require.Equal(t, "CASING", id)
	require.Equal(t, "Mayúsculas y minúsculas", name)
	require.Equal(t, "typographical", issue)
	require.Equal(t, "es", RuleLangHint("AI_ES_GGEC_X"))
	require.Equal(t, "pt", RuleLangHint("AI_PT_HYDRA_LEO_X"))
	require.Equal(t, "nl", RuleLangHint("AI_NL_HYDRA_LEO_X"))
	require.Equal(t, "en", RuleLangHint("AI_EN_LECTOR_X"))
	require.Equal(t, "en", RuleLangHint("AI_HYDRA_LEO_CP_X"))
	require.Equal(t, "en", RuleLangHint("AI_SPELLING_RULE_X"))
	id, name, issue, _ = RuleMeta("AI_PT_GGEC_REPLACEMENT_OTHER")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Gramática", name)
	require.Equal(t, "grammar", issue)
	id, name, issue, _ = RuleMeta("AI_NL_HYDRA_LEO_X")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammatica", name)
	require.Equal(t, "grammar", issue)
	id, name, issue, _ = RuleMeta("AI_ES_GGEC_REPLACEMENT_NOUN")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Gramática", name)
	require.Equal(t, "grammar", issue)
	// Java remote AI_DE_GGEC / HYDRA / KOMMA (German.getPriorityForId families)
	id, _, issue, short = RuleMeta("AI_DE_GGEC_REPLACEMENT_NOUN")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Potenzieller Fehler", short)
	id, _, issue, _ = RuleMeta("AI_DE_HYDRA_LEO_CP_X")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "grammar", issue)
	id, _, issue, _ = RuleMeta("AI_DE_KOMMA_FOO")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Hier scheint es einen Fehler zu geben.", RuleDescription("AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD"))

	// Statistic style / creative writing DE ids
	id, name, issue, short = RuleMeta("FILLER_WORDS_DE")
	require.Equal(t, "CREATIVE_WRITING", id)
	require.Equal(t, "Stiltipps für kreatives Schreiben", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Stil", short)
	id, _, issue, _ = RuleMeta("NON_SIGNIFICANT_VERB_DE")
	require.Equal(t, "CREATIVE_WRITING", id)
	// Java AbstractStatisticStyleRule / StyleRepeated* / Readability / LongSentence DE
	for _, rid := range []string{
		"SENTENCE_WITH_MAN_DE",
		"SENTENCE_WITH_MODAL_VERB_DE",
		"SENTENCE_BEGINNING_WITH_CONJUNCTION_DE",
		"UNNECESSARY_PHRASES_DE",
		"STYLE_REPEATED_SHORT_SENTENCES",
		"STYLE_REPEATED_SENTENCE_BEGINNING",
		"READABILITY_RULE_SIMPLE_DE",
		"READABILITY_RULE_DIFFICULT_DE",
	} {
		id, name, issue, short = RuleMeta(rid)
		require.Equal(t, "CREATIVE_WRITING", id, rid)
		require.Equal(t, "Stiltipps für kreatives Schreiben", name, rid)
		require.Equal(t, "style", issue, rid)
		require.Equal(t, "Stil", short, rid)
	}
	id, name, issue, short = RuleMeta("TOO_LONG_SENTENCE_DE")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Langer Satz", short)
	// Java ProhibitedCompoundRule: Categories.TYPOS; specificRuleId DE_PROHIBITED_COMPOUNDS_*
	id, name, issue, short = RuleMeta("DE_PROHIBITED_COMPOUNDS")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "Mögliche Tippfehler", name)
	require.Equal(t, "misspelling", issue)
	require.Equal(t, "Rechtschreibfehler", short)
	id, name, issue, short = RuleMeta("DE_PROHIBITED_COMPOUNDS_FOO_BAR")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "Rechtschreibfehler", short)
	// Java DE_PROHIBITED_COMPOUNDS_PREMIUM_* still TYPOS soft (priority is separate: -4)
	id, _, _, short = RuleMeta("DE_PROHIBITED_COMPOUNDS_PREMIUM_X")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "Rechtschreibfehler", short)
	require.Equal(t, "Markiert wahrscheinlich falsche Komposita wie 'Lehrzeile', wenn 'Leerzeile' häufiger vorkommt.",
		RuleDescription("DE_PROHIBITED_COMPOUNDS_FOO_BAR"))
	id, name, issue, short = RuleMeta("COMMA_IN_FRONT_RELATIVE_CLAUSE")
	require.Equal(t, "HILFESTELLUNG_KOMMASETZUNG", id)
	require.Equal(t, "Hilfestellung Kommasetzung", name)
	require.Equal(t, "Fehlendes Komma", short)
	id, name, issue, short = RuleMeta("COMPOUND_INFINITIV_RULE")
	require.Equal(t, "COMPOUNDING", id)
	require.Equal(t, "Getrennt- und Zusammenschreibung", name)
	require.Equal(t, "misspelling", issue)
	require.Equal(t, "Komposita", short)
	id, name, issue, short = RuleMeta("EINHEITEN_METRISCH")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "Maßeinheiten", short)
	id, name, issue, short = RuleMeta("DE_DU_UPPER_LOWER")
	require.Equal(t, "CASING", id)
	require.Equal(t, "Groß-/Kleinschreibung", name)
	require.Equal(t, "Großschreibung", short)

	// Java ConfusionProbabilityRule: TYPOS + NonConformance
	id, _, issue, _ = RuleMeta("DE_CONFUSION_RULE")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "non-conformance", issue)
	id, _, issue, _ = RuleMeta("DE_CONFUSION_RULE_seit_seid")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "non-conformance", issue)

	// Java GermanWrongWordInContextRule: CONFUSED_WORDS + Misspelling
	id, name, issue, short = RuleMeta("GERMAN_WRONG_WORD_IN_CONTEXT")
	require.Equal(t, "CONFUSED_WORDS", id)
	require.Equal(t, "Oft verwechselte Wörter", name) // MessagesBundle_de category_confused_words
	require.Equal(t, "misspelling", issue)
	require.Equal(t, "Wortverwechslung", short)
	// Specific rule IDs from WrongWordInContextRule (getId + matched + repl)
	id, _, issue, _ = RuleMeta("GERMAN_WRONG_WORD_IN_CONTEXT_MIENE_MINE")
	require.Equal(t, "CONFUSED_WORDS", id)
	require.Equal(t, "misspelling", issue)

	// Java AbstractRepeatedWordsRule: REPETITIONS_STYLE (DE_REPEATEDWORDS)
	id, name, issue, short = RuleMeta("DE_REPEATEDWORDS")
	require.Equal(t, "REPETITIONS_STYLE", id)
	require.Equal(t, "Wiederholungen", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Wortwiederholung", short)

	// Java ParagraphRepeatBeginningRule: STYLE
	id, name, issue, short = RuleMeta("GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Stil", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Absatzanfang", short)

	// Java StyleTooOftenUsed*Rule: CREATIVE_WRITING
	// Java StyleTooOftenUsed{Noun,Verb,Adjective}Rule: CREATIVE_WRITING + Style
	for _, rid := range []string{
		"TOO_OFTEN_USED_NOUN_DE",
		"TOO_OFTEN_USED_VERB_DE",
		"TOO_OFTEN_USED_ADJECTIVE_DE",
	} {
		id, name, issue, short = RuleMeta(rid)
		require.Equal(t, "CREATIVE_WRITING", id, rid)
		require.Equal(t, "Stiltipps für kreatives Schreiben", name, rid)
		require.Equal(t, "style", issue, rid)
		require.Equal(t, "Stil", short, rid)
		require.Equal(t, "de", RuleLangHint(rid), rid)
	}

	require.Equal(t, "error", SeverityFromIssueType("grammar"))
	require.Equal(t, "error", SeverityFromIssueType("misspelling"))
	require.Equal(t, "error", SeverityFromIssueType("non-conformance"))
	require.Equal(t, "note", SeverityFromIssueType("style"))
	require.Equal(t, "warning", SeverityFromIssueType("whitespace"))
}

func TestRuleDescription_Known(t *testing.T) {
	require.Equal(t, "Use of 'a' versus 'an'", RuleDescription("EN_A_VS_AN"))
	require.Equal(t, "Leere Zeile", RuleDescription("EMPTY_LINE"))
	// Soft invent: description is the raw id, not a fancy invent label.
	require.Equal(t, "EN_SOFT_YOUR_YOU_RE", RuleDescription("EN_SOFT_YOUR_YOU_RE"))
	require.Equal(t, "Mögliche Wortverwechslungen: $match", RuleDescription("GERMAN_WRONG_WORD_IN_CONTEXT"))
	require.Equal(t, "Synonyme für wiederholte Wörter.", RuleDescription("DE_REPEATEDWORDS"))
	require.Equal(t, "Statistische Stilanalyse: Zu häufig genutztes Substantiv",
		RuleDescription("TOO_OFTEN_USED_NOUN_DE"))
	// Java DE getDescription / MessagesBundle_de strings (faithful RuleDescription fallback)
	require.Equal(t, "Kongruenz von Nominalphrasen (unvollständig!), z.B. 'mein kleiner (kleines) Haus'",
		RuleDescription("DE_AGREEMENT"))
	require.Equal(t, "Kongruenz von Adjektiv und Nomen (unvollständig!), z.B. 'kleiner (kleines) Haus'",
		RuleDescription("DE_AGREEMENT2"))
	require.Equal(t, "Satz ohne Verb", RuleDescription("MISSING_VERB"))
	require.Equal(t, "Findet lange Sätze", RuleDescription("TOO_LONG_SENTENCE_DE"))
	require.Equal(t, "Lesbarkeit: Zu einfacher Text", RuleDescription("READABILITY_RULE_SIMPLE_DE"))
	require.Equal(t, "Lesbarkeit: Zu schwieriger Text", RuleDescription("READABILITY_RULE_DIFFICULT_DE"))
	require.Equal(t, "Fehlendes Komma vor Relativsatz", RuleDescription("COMMA_IN_FRONT_RELATIVE_CLAUSE"))
	require.Equal(t, "Markiert wahrscheinlich falsche Komposita wie 'Lehrzeile', wenn 'Leerzeile' häufiger vorkommt.",
		RuleDescription("DE_PROHIBITED_COMPOUNDS"))
	// Java CompoundInfinitivRule / OldSpellingRule / DashRule / GermanCompoundRule getDescription
	require.Equal(t, "Erweiterter Infinitiv mit zu (Zusammenschreibung)",
		RuleDescription("COMPOUND_INFINITIV_RULE"))
	require.Equal(t, "Findet Schreibweisen, die nur in der alten Rechtschreibung gültig waren",
		RuleDescription("OLD_SPELLING_RULE"))
	require.Equal(t, "Keine Leerzeichen in Bindestrich-Komposita (wie z.B. in 'Diäten- Erhöhung')",
		RuleDescription("DE_DASH"))
	require.Equal(t, "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'",
		RuleDescription("DE_COMPOUNDS"))
	require.Equal(t, "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'",
		RuleDescription("DE_CH_COMPOUNDS"))
	// Java CaseRule / DuUpperLower / UpperCaseNgram / SimilarName / SimpleReplace / coherency
	require.Equal(t, "Großschreibung von Nomen und substantivierten Verben", RuleDescription("DE_CASE"))
	require.Equal(t, "Einheitliche Verwendung von Du/du, Dir/dir etc.", RuleDescription("DE_DU_UPPER_LOWER"))
	require.Equal(t, "Prüft Wörter, ob sie fälschlich groß- oder fälschlich kleingeschrieben sind",
		RuleDescription("DE_UPPER_CASE_NGRAM"))
	require.Equal(t, "Mögliche Tippfehler in Namen finden", RuleDescription("DE_SIMILAR_NAMES"))
	require.Equal(t, "Prüft auf bestimmte falsche Wörter/Phrasen: $match", RuleDescription("DE_SIMPLE_REPLACE"))
	require.Equal(t, "Einheitliche Schreibweise für Wörter mit mehr als einer korrekten Schreibweise",
		RuleDescription("DE_WORD_COHERENCY"))
	require.Equal(t, "Einheitliche Schreibweise bei Komposita (mit oder ohne Bindestrich)",
		RuleDescription("DE_COMPOUND_COHERENCY"))
	require.Equal(t, "Kongruenz von Subjekt und Prädikat (unvollständig)",
		RuleDescription("DE_SUBJECT_VERB_AGREEMENT"))
	require.Equal(t, "Kongruenz von Subjekt und Prädikat (nur 1. u. 2. Person oder m. Personalpronomen), z.B. 'Er bist (ist)'",
		RuleDescription("DE_VERBAGREEMENT"))
	require.Equal(t, "Fehlendes Komma nach Relativsatz", RuleDescription("COMMA_BEHIND_RELATIVE_CLAUSE"))
	require.Equal(t, "Redundantes Modal- oder Hilfsverb", RuleDescription("REDUNDANT_MODAL_VERB"))
	require.Equal(t, "Wiederholte Worte in aufeinanderfolgenden Sätzen", RuleDescription("STYLE_REPEATED_WORD_RULE_DE"))
	require.Equal(t, "Statistische Stilanalyse: Füllwörter", RuleDescription("FILLER_WORDS_DE"))
	require.Equal(t, "Statistische Stilanalyse: Passivsätze", RuleDescription("PASSIVE_SENTENCE_DE"))
	require.Equal(t, "Statistische Stilanalyse: Zu häufig genutztes Verb", RuleDescription("TOO_OFTEN_USED_VERB_DE"))
	require.Equal(t, "Statistische Stilanalyse: Zu häufig genutztes Adjektiv",
		RuleDescription("TOO_OFTEN_USED_ADJECTIVE_DE"))
	require.Equal(t, "Schlägt vor oder überprüft Angaben des metrischen Äquivalentes bei bestimmten Maßeinheiten.",
		RuleDescription("EINHEITEN_METRISCH"))
	require.Equal(t, "Fehlendes Leerzeichen zwischen Sätzen oder nach Ordnungszahlen",
		RuleDescription("DE_SENTENCE_WHITESPACE"))
	// More statistic-style DE Java getDescription strings
	require.Equal(t, "Statistische Stilanalyse: Verben mit wenig Aussagekraft",
		RuleDescription("NON_SIGNIFICANT_VERB_DE"))
	require.Equal(t, "Statistische Stilanalyse: Sätze mit indirekter Leseransprache 'man'",
		RuleDescription("SENTENCE_WITH_MAN_DE"))
	require.Equal(t, "Statistische Stilanalyse: Sätze mit Modalverben",
		RuleDescription("SENTENCE_WITH_MODAL_VERB_DE"))
	require.Equal(t, "Statistische Stilanalyse: Sätze beginnend mit Konjunktion",
		RuleDescription("SENTENCE_BEGINNING_WITH_CONJUNCTION_DE"))
	require.Equal(t, "Statistische Stilanalyse: Potenzielle Phrasen",
		RuleDescription("UNNECESSARY_PHRASES_DE"))
	require.Equal(t, "Stakkato-Sätze", RuleDescription("STYLE_REPEATED_SHORT_SENTENCES"))
	require.Equal(t, "Subjekt als wiederholter Satzanfang", RuleDescription("STYLE_REPEATED_SENTENCE_BEGINNING"))
	require.Equal(t, "Zwei aufeinanderfolgende Kommas oder Punkte", RuleDescription("DE_DOUBLE_PUNCTUATION"))
	require.Equal(t, "Use of two consecutive dots or commas", RuleDescription("DOUBLE_PUNCTUATION"))
	require.Equal(t, "Möglicher Tippfehler 'spiegeln ... wieder (wider)'",
		RuleDescription("DE_WIEDER_VS_WIDER"))
	// MessagesBundle_de for DE word-repeat / speller / confusion / paragraph layout
	require.Equal(t, "Wortwiederholung (z. B. 'als als')", RuleDescription("GERMAN_WORD_REPEAT_RULE"))
	require.Equal(t, "Aufeinanderfolgende Sätze beginnen mit dem gleichen Wort",
		RuleDescription("GERMAN_WORD_REPEAT_BEGINNING_RULE"))
	require.Equal(t, "Möglicher Rechtschreibfehler", RuleDescription("GERMAN_SPELLER_RULE"))
	// Java MorfologikGermanyGermanSpellerRule uses MessagesBundle_de desc_spelling (not EN generic)
	require.Equal(t, "Möglicher Rechtschreibfehler", RuleDescription("MORFOLOGIK_RULE_DE_DE"))
	require.Equal(t, "Possible spelling mistake", RuleDescription("MORFOLOGIK_RULE_EN_US"))
	// Java de grammar.xml MULTITOKEN_SPELLING (not generic TYPOS RuleMeta)
	require.Equal(t, "Rechtschreibfehler in Eigennamen", RuleDescription("DE_MULTITOKEN_SPELLING_TWO"))
	require.Equal(t, "Rechtschreibfehler in Eigennamen", RuleDescription("DE_MULTITOKEN_SPELLING_THREE"))
	require.Equal(t, "Rechtschreibfehler in Eigennamen", RuleDescription("DE_MULTITOKEN_SPELLING_FOUR"))
	require.Equal(t, "Spelling mistakes in proper nouns", RuleDescription("EN_MULTITOKEN_SPELLING_THREE"))
	require.Equal(t, "Spelling mistakes in proper nouns", RuleDescription("EN_MULTITOKEN_SPELLING_FOUR"))
	require.Equal(t, "Mögliche Verwechselungen zwischen Wörtern erkennen", RuleDescription("DE_CONFUSION_RULE"))
	require.Equal(t, "Leerzeichen am Absatzende", RuleDescription("WHITESPACE_PARAGRAPH"))
	require.Equal(t, "Leerzeichen am Anfang des Absatzes", RuleDescription("WHITESPACE_PARAGRAPH_BEGIN"))
	require.Equal(t, "Kein Satzzeichen am Ende des Absatzes", RuleDescription("PUNCTUATION_PARAGRAPH_END"))
	require.Equal(t, "Lesbarkeit: Absatz mit mehr als {0} Wörtern", RuleDescription("TOO_LONG_PARAGRAPH"))
	require.Equal(t, "Gleicher Anfang von aufeinanderfolgenden Absätzen",
		RuleDescription("GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE"))
}

func TestRuleLangHint(t *testing.T) {
	require.Equal(t, "de", RuleLangHint("DE_AGREEMENT"))
	require.Equal(t, "fr", RuleLangHint("FR_AGREEMENT"))
	require.Equal(t, "", RuleLangHint("UNKNOWN"))
	// Java DE rules with GERMAN_/SWISS_/AUSTRIAN_ prefixes (not 2–3 letter codes)
	require.Equal(t, "de", RuleLangHint("GERMAN_WRONG_WORD_IN_CONTEXT"))
	require.Equal(t, "de", RuleLangHint("GERMAN_WORD_REPEAT_RULE"))
	require.Equal(t, "de", RuleLangHint("SWISS_GERMAN_SPELLER_RULE"))
	require.Equal(t, "de", RuleLangHint("AUSTRIAN_GERMAN_SPELLER_RULE"))
	// Java ProhibitedCompound specificRuleId + DE_ prefix; AI_DE_* remote families
	require.Equal(t, "de", RuleLangHint("DE_PROHIBITED_COMPOUNDS"))
	require.Equal(t, "de", RuleLangHint("DE_PROHIBITED_COMPOUNDS_FOO_BAR"))
	require.Equal(t, "de", RuleLangHint("DE_PROHIBITED_COMPOUNDS_PREMIUM_X"))
	require.Equal(t, "de", RuleLangHint("AI_DE_MERGED_MATCH"))
	require.Equal(t, "de", RuleLangHint("AI_DE_GGEC_REPLACEMENT_NOUN"))
	require.Equal(t, "de", RuleLangHint("COMPOUND_INFINITIV_RULE"))
	require.Equal(t, "de", RuleLangHint("OLD_SPELLING_RULE"))
	require.Equal(t, "de", RuleLangHint("EINHEITEN_METRISCH"))
	require.Equal(t, "de", RuleLangHint("REDUNDANT_MODAL_VERB"))
	require.Equal(t, "de", RuleLangHint("COMMA_BEHIND_RELATIVE_CLAUSE"))
	require.Equal(t, "de", RuleLangHint("STYLE_REPEATED_SENTENCE_BEGINNING"))
	require.Equal(t, "de", RuleLangHint("UNPAIRED_BRACKETS"))
	require.Equal(t, "de", RuleLangHint("MISSING_VERB"))
	// Java DE statistic/style IDs end with _DE
	require.Equal(t, "de", RuleLangHint("FILLER_WORDS_DE"))
	require.Equal(t, "de", RuleLangHint("TOO_LONG_SENTENCE_DE"))
	require.Equal(t, "de", RuleLangHint("READABILITY_RULE_SIMPLE_DE"))
	// DE-only Java IDs without DE_ / _DE markers
	require.Equal(t, "de", RuleLangHint("MISSING_VERB"))
	require.Equal(t, "de", RuleLangHint("COMPOUND_INFINITIV_RULE"))
	require.Equal(t, "de", RuleLangHint("COMMA_IN_FRONT_RELATIVE_CLAUSE"))
	require.Equal(t, "de", RuleLangHint("STYLE_REPEATED_SHORT_SENTENCES"))
	require.Equal(t, "de", RuleLangHint("UNPAIRED_BRACKETS"))
	// Java AI_DE_* remote/GGEC rule families
	require.Equal(t, "de", RuleLangHint("AI_DE_GGEC_REPLACEMENT_NOUN"))
	require.Equal(t, "de", RuleLangHint("AI_DE_MERGED_MATCH"))
	require.Equal(t, "de", RuleLangHint("AI_DE_HYDRA_LEO_CP_X"))
	// Java Morfologik*SpellerRule IDs: MORFOLOGIK_RULE_{lang}_{variant}
	require.Equal(t, "de", RuleLangHint("MORFOLOGIK_RULE_DE_DE"))
	require.Equal(t, "en", RuleLangHint("MORFOLOGIK_RULE_EN_US"))
	require.Equal(t, "en", RuleLangHint("MORFOLOGIK_RULE_EN_CA"))
	require.Equal(t, "nl", RuleLangHint("MORFOLOGIK_RULE_NL_NL"))
	require.Equal(t, "tl", RuleLangHint("MORFOLOGIK_RULE_TL"))
	require.Equal(t, "ast", RuleLangHint("MORFOLOGIK_RULE_AST"))
	require.Equal(t, "crh", RuleLangHint("MORFOLOGIK_RULE_CRH_UA"))
}

func TestRuleURL(t *testing.T) {
	require.Contains(t, RuleURL("EN_A_VS_AN", "en"), "lang=en")
	require.Contains(t, RuleURL("DE_AGREEMENT", ""), "lang=de")
	// Java MorfologikGermanyGermanSpellerRule → de community URL (not en default)
	require.Contains(t, RuleURL("MORFOLOGIK_RULE_DE_DE", ""), "lang=de")
	require.Contains(t, RuleURL("MORFOLOGIK_RULE_EN_US", ""), "lang=en")
	// DE-only IDs without DE_ prefix: RuleLangHint → de when lang empty
	require.Contains(t, RuleURL("COMPOUND_INFINITIV_RULE", ""), "lang=de")
	require.Contains(t, RuleURL("OLD_SPELLING_RULE", ""), "lang=de")
	require.Contains(t, RuleURL("AUSTRIAN_GERMAN_SPELLER_RULE", ""), "lang=de")
	require.Contains(t, RuleURL("SWISS_GERMAN_SPELLER_RULE", ""), "lang=de")
	require.Contains(t, RuleURL("DE_PROHIBITED_COMPOUNDS_FOO_BAR", ""), "lang=de")
	require.Contains(t, RuleURL("AI_DE_MERGED_MATCH", ""), "lang=de")
	require.Contains(t, RuleURL("REDUNDANT_MODAL_VERB", ""), "lang=de")
	require.Contains(t, RuleURL("COMMA_BEHIND_RELATIVE_CLAUSE", ""), "lang=de")
	require.Contains(t, RuleURL("STYLE_REPEATED_SENTENCE_BEGINNING", ""), "lang=de")
	require.Contains(t, RuleURL("UNPAIRED_BRACKETS", ""), "lang=de")
	require.Contains(t, RuleURL("EINHEITEN_METRISCH", ""), "lang=de")
	require.Contains(t, RuleURL("MISSING_VERB", ""), "lang=de")
	require.Contains(t, RuleURL("DE_MULTITOKEN_SPELLING_THREE", ""), "lang=de")
	require.Contains(t, RuleURL("TOO_OFTEN_USED_NOUN_DE", ""), "lang=de")
	require.Contains(t, RuleURL("FILLER_WORDS_DE", ""), "lang=de")
	require.Contains(t, RuleURL("DE_CH_COMPOUNDS", ""), "lang=de")
}

func TestSoftRule_DEPhraseRepetition(t *testing.T) {
	id, name, issue, short := RuleMeta("DE_PHRASE_REPETITION")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammatik", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Wortgruppenwiederholung", short)
	require.Equal(t, "Wiederholung von Wortgruppen (z.B. 'auf der auf der Straße')",
		RuleDescription("DE_PHRASE_REPETITION"))
	require.Equal(t, "de", RuleLangHint("DE_PHRASE_REPETITION"))
}

// Java AbstractRepeatedWordsRule setSpecificRuleId(ruleId + "_" + toId(lemma))
func TestSoftRule_DERepeatedWordsSpecificId(t *testing.T) {
	id, name, issue, short := RuleMeta("DE_REPEATEDWORDS_AUSSERDEM")
	require.Equal(t, "REPETITIONS_STYLE", id)
	require.Equal(t, "Wiederholungen", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Wortwiederholung", short)
	require.Equal(t, "Synonyme für wiederholte Wörter.",
		RuleDescription("DE_REPEATEDWORDS_AUSSERDEM"))
}
