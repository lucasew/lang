package languagetool

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// RuleMeta is a fallback when LocalMatch has no CategoryID/IssueType.
// It maps well-known Java rule ID families to Categories / ITS types used in LT
// (e.g. Morfologik → TYPOS/misspelling). Unknown IDs stay uncategorized — no invent.
// Prefer setting CategoryID/IssueType on LocalMatch from the rule itself.
func RuleMeta(ruleID string) (categoryID, categoryName, issueType, short string) {
	id := strings.ToUpper(ruleID)
	switch {
	// Multitoken proper-noun spelling (before generic SPELL → TYPOS).
	// Java grammar.xml category id=MULTITOKEN_SPELLING type=misspelling
	// (DE name Rechtschreibfehler; EN Orthographic errors; short DE: Möglicher Fehler).
	case strings.Contains(id, "MULTITOKEN_SPELLING"):
		if strings.HasPrefix(id, "DE_") {
			return "MULTITOKEN_SPELLING", "Rechtschreibfehler", "misspelling", "Möglicher Fehler"
		}
		return "MULTITOKEN_SPELLING", "Orthographic errors", "misspelling", "Spelling mistake"
	// DE speller IDs before generic SPELL (MessagesBundle_de desc_spelling_short / category_typo).
	case id == "GERMAN_SPELLER_RULE" || id == "AUSTRIAN_GERMAN_SPELLER_RULE" ||
		id == "SWISS_GERMAN_SPELLER_RULE" || strings.HasPrefix(id, "MORFOLOGIK_RULE_DE"):
		// Java GermanSpellerRule / MorfologikGermanyGermanSpellerRule: Categories.TYPOS
		return "TYPOS", "Mögliche Tippfehler", "misspelling", "Rechtschreibfehler"
	// OLD_SPELLING_RULE contains "SPELL" — must precede generic SPELL Contains path.
	case id == "OLD_SPELLING_RULE":
		// Java OldSpellingRule: Categories.TYPOS
		return "TYPOS", "Mögliche Tippfehler", "misspelling", "Alte Rechtschreibung"
	case strings.Contains(id, "MORFOLOGIK") || strings.Contains(id, "HUNSPELL") ||
		(strings.Contains(id, "SPELL") && !strings.Contains(id, "IGNORE_SPELLING")):
		// Java SpellingCheckRule / MorfologikSpellerRule: Categories.TYPOS, ITS misspelling
		return "TYPOS", "Possible Typo", "misspelling", "Spelling mistake"
	// Double punctuation: Java Categories.PUNCTUATION + Typographical (not TYPOGRAPHY).
	case id == "DE_DOUBLE_PUNCTUATION":
		// Java GermanDoublePunctuationRule; MessagesBundle_de category_punctuation / desc_double_punct
		return "PUNCTUATION", "Zeichensetzung", "typographical", "Doppelte Satzzeichen"
	case id == "DOUBLE_PUNCTUATION":
		return "PUNCTUATION", "Punctuation", "typographical", "Double punctuation"
	case id == "DE_SENTENCE_WHITESPACE":
		// Java de.SentenceWhitespaceRule: Categories.MISC; MessagesBundle_de category_misc
		return "MISC", "Sonstiges", "whitespace", "Leerzeichen zwischen Sätzen"
	// Paragraph whitespace rules are STYLE (Java), not TYPOGRAPHY — check before generic WHITESPACE.
	case id == "WHITESPACE_PARAGRAPH":
		// Java WhiteSpaceBeforeParagraphEnd; MessagesBundle_de category_style / end_desc
		return "STYLE", "Stil", "style", "Leerzeichen am Absatzende"
	case id == "WHITESPACE_PARAGRAPH_BEGIN":
		// Java WhiteSpaceAtBeginOfParagraph; MessagesBundle_de category_style / begin_desc
		return "STYLE", "Stil", "style", "Leerzeichen am Anfang des Absatzes"
	// Java CommaWhitespaceRule ID is COMMA_PARENTHESIS_WHITESPACE (no "WHITESPACE" substring).
	case id == "COMMA_PARENTHESIS_WHITESPACE" || id == "COMMA_WHITESPACE" ||
		id == "SENTENCE_WHITESPACE" || strings.Contains(id, "WHITESPACE"):
		// Java MultipleWhitespace / CommaWhitespace: TYPOGRAPHY + Whitespace
		// DE uses Typografie for shared layout IDs registered under de packs
		if id == "COMMA_PARENTHESIS_WHITESPACE" || id == "WHITESPACE_RULE" || id == "WHITESPACE_PUNCTUATION" {
			return "TYPOGRAPHY", "Typografie", "whitespace", "Typografie"
		}
		return "TYPOGRAPHY", "Typography", "whitespace", "Typography"
	case id == "EMPTY_LINE":
		// Java EmptyLineRule: Categories.STYLE + ITSIssueType.Style
		// Short from MessagesBundle_de empty_line_rule_desc
		return "STYLE", "Stil", "style", "Leere Zeile"
	case id == "PUNCTUATION_PARAGRAPH_END":
		// Java PunctuationMarkAtParagraphEnd: Categories.PUNCTUATION + Grammar
		// MessagesBundle_de category_punctuation
		return "PUNCTUATION", "Zeichensetzung", "grammar", "Absatzende"
	// Word-repeat families — order matters (more specific first).
	case id == "GERMAN_WORD_REPEAT_BEGINNING_RULE" ||
		(strings.Contains(id, "WORD_REPEAT_BEGINNING") && strings.HasPrefix(id, "GERMAN_")):
		// Java GermanWordRepeatBeginningRule: REPETITIONS_STYLE; MessagesBundle_de category_repetitions
		return "REPETITIONS_STYLE", "Wiederholungen", "style", "Wortwiederholung am Satzanfang"
	case strings.Contains(id, "WORD_REPEAT_BEGINNING"):
		// Java WordRepeatBeginningRule: REPETITIONS_STYLE
		return "REPETITIONS_STYLE", "Repetitions (Style)", "style", "Word repetition at beginning"
	case id == "GERMAN_WORD_REPEAT_RULE":
		// Java GermanWordRepeatRule: Categories.REDUNDANCY; MessagesBundle_de category_redundancy / desc_repetition_short
		return "REDUNDANCY", "Redundanz", "duplication", "Wortwiederholung"
	case strings.Contains(id, "WORD_REPEAT"):
		// Java WordRepeatRule: Categories.MISC
		return "MISC", "Miscellaneous", "duplication", "Word repetition"
	case id == "EN_A_VS_AN" || strings.Contains(id, "A_VS_AN"):
		// Java AvsAnRule: Categories.MISC + ITSIssueType.Misspelling
		return "MISC", "Miscellaneous", "misspelling", "Wrong article"
	// German grammar rules (Java AgreementRule / VerbAgreement / SubjectVerbAgreement)
	case id == "DE_AGREEMENT" || id == "DE_AGREEMENT2" ||
		id == "DE_SUBJECT_VERB_AGREEMENT" || id == "DE_VERBAGREEMENT" ||
		id == "MISSING_VERB":
		// Java AgreementRule / VerbAgreement / MissingVerbRule: Categories.GRAMMAR
		// Short: DE "Kongruenz" (from getDescription family), category name MessagesBundle_de
		return "GRAMMAR", "Grammatik", "grammar", "Kongruenz"
	case id == "COMMA_IN_FRONT_RELATIVE_CLAUSE" || id == "COMMA_BEHIND_RELATIVE_CLAUSE":
		// Java MissingCommaRelativeClauseRule: custom HILFESTELLUNG_KOMMASETZUNG
		return "HILFESTELLUNG_KOMMASETZUNG", "Hilfestellung Kommasetzung", "grammar", "Fehlendes Komma"
	case id == "DE_CASE" || id == "DE_DU_UPPER_LOWER" || id == "DE_UPPER_CASE_NGRAM":
		// Java CaseRule / DuUpperLowerCaseRule / UpperCaseNgramRule: Categories.CASING
		// MessagesBundle_de category_case
		return "CASING", "Groß-/Kleinschreibung", "typographical", "Großschreibung"
	case id == "DE_COMPOUNDS" || id == "DE_CH_COMPOUNDS" || id == "DE_DASH" ||
		id == "COMPOUND_INFINITIV_RULE":
		// Java GermanCompoundRule / DashRule / CompoundInfinitivRule: Categories.COMPOUNDING
		// MessagesBundle_de category_compounding
		return "COMPOUNDING", "Getrennt- und Zusammenschreibung", "misspelling", "Komposita"
	case id == "DE_CONFUSION_RULE" || strings.HasPrefix(id, "DE_CONFUSION_RULE_"):
		// Java GermanConfusionProbabilityRule: Categories.TYPOS, ITS NonConformance
		return "TYPOS", "Mögliche Tippfehler", "non-conformance", "Verwechslung"
	case strings.HasPrefix(id, "CONFUSION_RULE_"):
		// Non-DE confusion probability IDs
		return "TYPOS", "Possible Typo", "non-conformance", "Confusion"
	case id == "DE_WIEDER_VS_WIDER" || id == "DE_SIMILAR_NAMES" ||
		id == "DE_PROHIBITED_COMPOUNDS" || strings.HasPrefix(id, "DE_PROHIBITED_COMPOUNDS_"):
		// Java WiederVsWider / SimilarName / ProhibitedCompound: Categories.TYPOS
		return "TYPOS", "Mögliche Tippfehler", "misspelling", "Rechtschreibfehler"
	case id == "DE_SIMPLE_REPLACE":
		// Java AbstractSimpleReplaceRule: Categories.MISC; MessagesBundle_de category_misc
		return "MISC", "Sonstiges", "misspelling", "Falsches Wort"
	case id == "DE_WORD_COHERENCY":
		// Java AbstractWordCoherencyRule: Categories.MISC
		return "MISC", "Sonstiges", "misspelling", "Uneinheitliche Schreibweise"
	case id == "DE_COMPOUND_COHERENCY" || id == "REDUNDANT_MODAL_VERB" ||
		id == "STYLE_REPEATED_WORD_RULE_DE":
		// Java CompoundCoherencyRule / RedundantModal / GermanStyleRepeatedWord: STYLE
		// MessagesBundle_de category_style
		return "STYLE", "Stil", "style", "Stil"
	// AbstractStatisticStyleRule / creative-writing DE ids: CREATIVE_WRITING + Style
	case id == "FILLER_WORDS_DE" || id == "PASSIVE_SENTENCE_DE" ||
		id == "SENTENCE_WITH_MAN_DE" || id == "SENTENCE_WITH_MODAL_VERB_DE" ||
		id == "NON_SIGNIFICANT_VERB_DE" || id == "SENTENCE_BEGINNING_WITH_CONJUNCTION_DE" ||
		id == "UNNECESSARY_PHRASES_DE" ||
		id == "STYLE_REPEATED_SHORT_SENTENCES" || id == "STYLE_REPEATED_SENTENCE_BEGINNING" ||
		id == "READABILITY_RULE_SIMPLE_DE" || id == "READABILITY_RULE_DIFFICULT_DE" ||
		// Java StyleTooOftenUsed{Noun,Verb,Adjective}Rule
		id == "TOO_OFTEN_USED_NOUN_DE" || id == "TOO_OFTEN_USED_VERB_DE" ||
		id == "TOO_OFTEN_USED_ADJECTIVE_DE":
		// MessagesBundle_de category_creative_writing
		return "CREATIVE_WRITING", "Stiltipps für kreatives Schreiben", "style", "Stil"
	case id == "EINHEITEN_METRISCH":
		// Java AbstractUnitConversionRule: Categories.STYLE
		return "STYLE", "Stil", "style", "Maßeinheiten"
	case id == "GERMAN_WRONG_WORD_IN_CONTEXT" || strings.HasPrefix(id, "GERMAN_WRONG_WORD_IN_CONTEXT_"):
		// Java WrongWordInContextRule: CategoryIds.CONFUSED_WORDS + Misspelling
		// Name: DE getCategoryString / MessagesBundle_de category_confused_words
		return "CONFUSED_WORDS", "Oft verwechselte Wörter", "misspelling", "Wortverwechslung"
	case id == "DE_REPEATEDWORDS" || strings.HasPrefix(id, "DE_REPEATEDWORDS_"):
		// Java AbstractRepeatedWordsRule: Categories.REPETITIONS_STYLE + Style
		// MessagesBundle_de category_repetitions; specificRuleId DE_REPEATEDWORDS_{lemma}
		return "REPETITIONS_STYLE", "Wiederholungen", "style", "Wortwiederholung"
	case id == "DE_PHRASE_REPETITION":
		// Java de grammar.xml rulegroup DE_PHRASE_REPETITION under category GRAMMAR
		return "GRAMMAR", "Grammatik", "grammar", "Wortgruppenwiederholung"
	case id == "GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE":
		// Java ParagraphRepeatBeginningRule: Categories.STYLE + Style
		return "STYLE", "Stil", "style", "Absatzanfang"
	case id == "AI_DE_MERGED_MATCH" ||
		// Java remote AI families (German.getPriorityForId AI_DE_GGEC / HYDRA / KOMMA).
		// Fallback when remote payload left category empty — not invent for unknown IDs.
		strings.HasPrefix(id, "AI_DE_GGEC") || strings.HasPrefix(id, "AI_DE_HYDRA") ||
		strings.HasPrefix(id, "AI_DE_KOMMA"):
		// Java German.mergeMatches short "Potenzieller Fehler"; GGEC/HYDRA default grammar.
		return "GRAMMAR", "Grammatik", "grammar", "Potenzieller Fehler"
	case strings.HasPrefix(id, "AI_FR_MERGED_MATCH") ||
		// Java French.mergeMatches short "Erreur potentielle"; AI_FR_GGEC / HYDRA families.
		strings.HasPrefix(id, "AI_FR_GGEC") || strings.HasPrefix(id, "AI_FR_HYDRA"):
		return "GRAMMAR", "Grammaire", "grammar", "Erreur potentielle"
	case strings.Contains(id, "CASING") && strings.HasPrefix(id, "AI_ES_"):
		// Java Spanish.filterRuleMatches casing rewrite → Categories.CASING
		return "CASING", "Mayúsculas y minúsculas", "typographical", "Mayúsculas y minúsculas"
	case strings.HasPrefix(id, "AI_ES_GGEC") || strings.HasPrefix(id, "AI_ES_"):
		// Fallback for Spanish remote AI (not invent for unknown non-AI ids).
		return "GRAMMAR", "Gramática", "grammar", ""
	case strings.HasPrefix(id, "AI_PT_GGEC") || strings.HasPrefix(id, "AI_PT_HYDRA") || strings.HasPrefix(id, "AI_PT_"):
		// Fallback for Portuguese remote AI (Portuguese.getPriorityForId families).
		return "GRAMMAR", "Gramática", "grammar", ""
	case strings.HasPrefix(id, "AI_NL_HYDRA") || strings.HasPrefix(id, "AI_NL_"):
		// Fallback for Dutch remote AI (Dutch.getPriorityForId families).
		return "GRAMMAR", "Grammatica", "grammar", ""
	case id == "DE_UNPAIRED_QUOTES" || id == "UNPAIRED_BRACKETS":
		// MessagesBundle_de category_typography
		return "TYPOGRAPHY", "Typografie", "typographical", "Unpaarige Zeichen"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "TYPOGRAPHY", "Typography", "typographical", "Unpaired symbol"
	case id == "UPPERCASE_SENTENCE_START" || strings.Contains(id, "UPPERCASE_SENTENCE_START"):
		// Shared layout; DE MessagesBundle category_case when used under de
		return "CASING", "Groß-/Kleinschreibung", "typographical", "Großschreibung"
	// Paragraph before sentence: TOO_LONG_PARAGRAPH contains TOO_LONG but is LongParagraphRule.
	case id == "TOO_LONG_PARAGRAPH" || strings.Contains(id, "LONG_PARAGRAPH"):
		// Java LongParagraphRule: Categories.STYLE + Style
		return "STYLE", "Stil", "style", "Langer Absatz"
	case id == "TOO_LONG_SENTENCE_DE" || (strings.Contains(id, "LONG_SENTENCE") && strings.HasSuffix(id, "_DE")):
		return "STYLE", "Stil", "style", "Langer Satz"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		// Java LongSentenceRule (non-DE)
		return "STYLE", "Style", "style", "Long sentence"
	case strings.Contains(id, "FALSE_FRIEND"):
		// Java FalseFriendRule: Categories.FALSE_FRIENDS; MessagesBundle_de category_false_friend
		return "FALSEFRIENDS", "Falsche Freunde", "misspelling", "False friend"
	default:
		if ruleID == "" {
			return "MISC", "Miscellaneous", "uncategorized", ""
		}
		// Unknown rule: uncategorized — do not invent grammar/style from ID shape.
		return "MISC", "Miscellaneous", "uncategorized", ""
	}
}

// RuleDescription returns a short description for known Java rule families.
// Prefer rule.GetDescription() when available; this is CLI/API fallback only.
func RuleDescription(ruleID string) string {
	id := strings.ToUpper(ruleID)
	switch {
	case id == "EN_A_VS_AN" || strings.Contains(id, "A_VS_AN"):
		return "Use of 'a' versus 'an'"
	case id == "GERMAN_WORD_REPEAT_BEGINNING_RULE" ||
		(strings.Contains(id, "WORD_REPEAT_BEGINNING") && strings.HasPrefix(id, "GERMAN_")):
		// Java MessagesBundle_de desc_repetition_beginning
		return "Aufeinanderfolgende Sätze beginnen mit dem gleichen Wort"
	case strings.Contains(id, "WORD_REPEAT_BEGINNING"):
		// Java WordRepeatBeginningRule: messages desc_repetition_beginning (EN fallback)
		return "Successive sentences beginning with the same word"
	case id == "GERMAN_WORD_REPEAT_RULE":
		// Java MessagesBundle_de desc_repetition
		return "Wortwiederholung (z. B. 'als als')"
	case strings.Contains(id, "WORD_REPEAT"):
		return "Word repetition"
	// DE rule descriptions: Java getDescription strings (not invent).
	case id == "DE_AGREEMENT":
		return "Kongruenz von Nominalphrasen (unvollständig!), z.B. 'mein kleiner (kleines) Haus'"
	case id == "DE_AGREEMENT2":
		return "Kongruenz von Adjektiv und Nomen (unvollständig!), z.B. 'kleiner (kleines) Haus'"
	case id == "DE_SUBJECT_VERB_AGREEMENT":
		return "Kongruenz von Subjekt und Prädikat (unvollständig)"
	case id == "DE_VERBAGREEMENT":
		return "Kongruenz von Subjekt und Prädikat (nur 1. u. 2. Person oder m. Personalpronomen), z.B. 'Er bist (ist)'"
	case id == "MISSING_VERB":
		return "Satz ohne Verb"
	case id == "DE_CASE":
		return "Großschreibung von Nomen und substantivierten Verben"
	case id == "DE_UPPER_CASE_NGRAM":
		return "Prüft Wörter, ob sie fälschlich groß- oder fälschlich kleingeschrieben sind"
	case id == "DE_DU_UPPER_LOWER":
		return "Einheitliche Verwendung von Du/du, Dir/dir etc."
	case id == "DE_COMPOUNDS" || id == "DE_CH_COMPOUNDS":
		return "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'"
	case id == "DE_DASH":
		return "Keine Leerzeichen in Bindestrich-Komposita (wie z.B. in 'Diäten- Erhöhung')"
	case id == "COMPOUND_INFINITIV_RULE":
		return "Erweiterter Infinitiv mit zu (Zusammenschreibung)"
	case id == "DE_PROHIBITED_COMPOUNDS" || strings.HasPrefix(id, "DE_PROHIBITED_COMPOUNDS_"):
		return "Markiert wahrscheinlich falsche Komposita wie 'Lehrzeile', wenn 'Leerzeile' häufiger vorkommt."
	case id == "OLD_SPELLING_RULE":
		return "Findet Schreibweisen, die nur in der alten Rechtschreibung gültig waren"
	case id == "DE_CONFUSION_RULE" || strings.HasPrefix(id, "DE_CONFUSION_RULE_") ||
		strings.HasPrefix(id, "CONFUSION_RULE_"):
		// Java MessagesBundle_de statistics_rule_description (without term placeholders)
		return "Mögliche Verwechselungen zwischen Wörtern erkennen"
	case id == "DE_WIEDER_VS_WIDER":
		// Java WiederVsWiderRule.getDescription
		return "Möglicher Tippfehler 'spiegeln ... wieder (wider)'"
	case id == "DE_SIMILAR_NAMES":
		return "Mögliche Tippfehler in Namen finden"
	case id == "DE_SIMPLE_REPLACE":
		return "Prüft auf bestimmte falsche Wörter/Phrasen: $match"
	case id == "DE_WORD_COHERENCY":
		return "Einheitliche Schreibweise für Wörter mit mehr als einer korrekten Schreibweise"
	case id == "DE_COMPOUND_COHERENCY":
		return "Einheitliche Schreibweise bei Komposita (mit oder ohne Bindestrich)"
	case id == "GERMAN_WRONG_WORD_IN_CONTEXT" || strings.HasPrefix(id, "GERMAN_WRONG_WORD_IN_CONTEXT_"):
		// Java GermanWrongWordInContextRule.getDescription (without $match expansion)
		return "Mögliche Wortverwechslungen: $match"
	case id == "DE_REPEATEDWORDS" || strings.HasPrefix(id, "DE_REPEATEDWORDS_"):
		return "Synonyme für wiederholte Wörter."
	case id == "DE_PHRASE_REPETITION":
		// Java de grammar.xml rulegroup name=
		return "Wiederholung von Wortgruppen (z.B. 'auf der auf der Straße')"
	case id == "GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE":
		// Java MessagesBundle_de repetition_paragraph_beginning_desc
		return "Gleicher Anfang von aufeinanderfolgenden Absätzen"
	case id == "AI_DE_MERGED_MATCH" ||
		strings.HasPrefix(id, "AI_DE_GGEC") || strings.HasPrefix(id, "AI_DE_HYDRA") ||
		strings.HasPrefix(id, "AI_DE_KOMMA"):
		// Java German.mergeMatches message; remote AI families use same fallback.
		return "Hier scheint es einen Fehler zu geben."
	case strings.HasPrefix(id, "AI_FR_MERGED_MATCH") ||
		strings.HasPrefix(id, "AI_FR_GGEC") || strings.HasPrefix(id, "AI_FR_HYDRA"):
		// Java French.mergeMatches message.
		return "Il pourrait y avoir un problème ici."
	case id == "COMMA_IN_FRONT_RELATIVE_CLAUSE":
		return "Fehlendes Komma vor Relativsatz"
	case id == "COMMA_BEHIND_RELATIVE_CLAUSE":
		return "Fehlendes Komma nach Relativsatz"
	case id == "REDUNDANT_MODAL_VERB":
		return "Redundantes Modal- oder Hilfsverb"
	case id == "STYLE_REPEATED_WORD_RULE_DE":
		return "Wiederholte Worte in aufeinanderfolgenden Sätzen"
	case id == "FILLER_WORDS_DE":
		return "Statistische Stilanalyse: Füllwörter"
	case id == "PASSIVE_SENTENCE_DE":
		return "Statistische Stilanalyse: Passivsätze"
	case id == "NON_SIGNIFICANT_VERB_DE":
		return "Statistische Stilanalyse: Verben mit wenig Aussagekraft"
	case id == "SENTENCE_WITH_MAN_DE":
		return "Statistische Stilanalyse: Sätze mit indirekter Leseransprache 'man'"
	case id == "SENTENCE_WITH_MODAL_VERB_DE":
		return "Statistische Stilanalyse: Sätze mit Modalverben"
	case id == "SENTENCE_BEGINNING_WITH_CONJUNCTION_DE":
		return "Statistische Stilanalyse: Sätze beginnend mit Konjunktion"
	case id == "UNNECESSARY_PHRASES_DE":
		return "Statistische Stilanalyse: Potenzielle Phrasen"
	case id == "STYLE_REPEATED_SHORT_SENTENCES":
		// Java StyleRepeatedVeryShortSentences.getDescription
		return "Stakkato-Sätze"
	case id == "STYLE_REPEATED_SENTENCE_BEGINNING":
		// Java StyleRepeatedSentenceBeginning.getDescription
		return "Subjekt als wiederholter Satzanfang"
	case id == "TOO_OFTEN_USED_NOUN_DE":
		return "Statistische Stilanalyse: Zu häufig genutztes Substantiv"
	case id == "TOO_OFTEN_USED_VERB_DE":
		return "Statistische Stilanalyse: Zu häufig genutztes Verb"
	case id == "TOO_OFTEN_USED_ADJECTIVE_DE":
		return "Statistische Stilanalyse: Zu häufig genutztes Adjektiv"
	case id == "READABILITY_RULE_SIMPLE_DE":
		return "Lesbarkeit: Zu einfacher Text"
	case id == "READABILITY_RULE_DIFFICULT_DE":
		return "Lesbarkeit: Zu schwieriger Text"
	case id == "EINHEITEN_METRISCH":
		return "Schlägt vor oder überprüft Angaben des metrischen Äquivalentes bei bestimmten Maßeinheiten."
	case id == "DE_DOUBLE_PUNCTUATION":
		// Java MessagesBundle_de desc_double_punct
		return "Zwei aufeinanderfolgende Kommas oder Punkte"
	case id == "DOUBLE_PUNCTUATION":
		// Java MessagesBundle_en desc_double_punct
		return "Use of two consecutive dots or commas"
	case id == "TOO_LONG_SENTENCE_DE" || (strings.Contains(id, "LONG_SENTENCE") && strings.HasSuffix(id, "_DE")):
		// Java de LongSentenceRule: "Findet lange Sätze" (not the MessagesBundle long_sentence_rule_desc template)
		return "Findet lange Sätze"
	case id == "GERMAN_SPELLER_RULE" || id == "AUSTRIAN_GERMAN_SPELLER_RULE" ||
		id == "SWISS_GERMAN_SPELLER_RULE" ||
		// Java MorfologikGermanyGermanSpellerRule (and DE variants): MessagesBundle_de desc_spelling
		strings.HasPrefix(id, "MORFOLOGIK_RULE_DE"):
		// Java MessagesBundle_de desc_spelling
		return "Möglicher Rechtschreibfehler"
	// Java de grammar.xml DE_MULTITOKEN_SPELLING_{TWO,THREE,FOUR} rulegroup names
	case strings.HasPrefix(id, "DE_MULTITOKEN_SPELLING"):
		return "Rechtschreibfehler in Eigennamen"
	case strings.Contains(id, "MULTITOKEN_SPELLING"):
		// EN/other: grammar.xml "Spelling mistakes in proper nouns (…)"
		return "Spelling mistakes in proper nouns"
	case strings.Contains(id, "MORFOLOGIK") || strings.Contains(id, "HUNSPELL") ||
		(strings.Contains(id, "SPELL") && !strings.Contains(id, "IGNORE_SPELLING")):
		return "Possible spelling mistake"
	case id == "DE_SENTENCE_WHITESPACE":
		// Java SentenceWhitespaceRule (de): Fehlendes Leerzeichen zwischen Sätzen…
		return "Fehlendes Leerzeichen zwischen Sätzen oder nach Ordnungszahlen"
	case id == "WHITESPACE_PARAGRAPH":
		// Java MessagesBundle_de whitespace_before_parapgraph_end_desc
		return "Leerzeichen am Absatzende"
	case id == "WHITESPACE_PARAGRAPH_BEGIN":
		// Java MessagesBundle_de whitespace_at_begin_parapgraph_desc
		return "Leerzeichen am Anfang des Absatzes"
	case id == "COMMA_PARENTHESIS_WHITESPACE":
		// Java MessagesBundle_de desc_comma_whitespace (GermanCommaWhitespaceRule uses DE bundle)
		return "Leerzeichen vor/hinter Kommas und Klammern"
	case id == "WHITESPACE_PUNCTUATION":
		// Java MessagesBundle_de desc_whitespace_before_punctuation
		// (German layout pack + shared WhitespaceBeforePunctuationRule)
		return "Leerzeichen vor Doppelpunkt, Semikolon oder Prozentzeichen."
	case id == "WHITESPACE_RULE":
		// Java MultipleWhitespaceRule: MessagesBundle_de desc_whitespacerepetition
		return "Wiederholung von Leerzeichen"
	case id == "SENTENCE_WHITESPACE" || id == "COMMA_WHITESPACE" || strings.Contains(id, "WHITESPACE"):
		return "Whitespace"
	case id == "EMPTY_LINE":
		// Java MessagesBundle_de empty_line_rule_desc
		return "Leere Zeile"
	case id == "PUNCTUATION_PARAGRAPH_END":
		// Java MessagesBundle_de punctuation_mark_paragraph_end_desc
		return "Kein Satzzeichen am Ende des Absatzes"
	case id == "DE_UNPAIRED_QUOTES":
		// Java MessagesBundle_de desc_unpaired_quotes
		return "Unpaarige Anführungszeichen"
	case id == "UNPAIRED_BRACKETS" || (strings.Contains(id, "UNPAIRED") && strings.HasPrefix(id, "GERMAN_")):
		// Java GermanUnpairedBracketsRule keeps ID UNPAIRED_BRACKETS; MessagesBundle_de
		return "Unpaarige Anführungszeichen und Klammern"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "Unpaired brackets"
	case id == "UPPERCASE_SENTENCE_START" || strings.Contains(id, "UPPERCASE_SENTENCE_START"):
		// Java MessagesBundle_de desc_uppercase_sentence
		return "Großschreibung am Satzanfang"
	// TOO_LONG_PARAGRAPH before generic TOO_LONG (LongParagraphRule, not LongSentenceRule)
	case id == "TOO_LONG_PARAGRAPH" || strings.Contains(id, "LONG_PARAGRAPH"):
		// Java MessagesBundle_de long_paragraph_rule_desc without {0} substitution
		return "Lesbarkeit: Absatz mit mehr als {0} Wörtern"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		return "Long sentence"
	case strings.Contains(id, "FALSE_FRIEND"):
		return "False friend"
	case ruleID == "":
		return ""
	default:
		return ruleID
	}
}

// SeverityFromIssueType maps ITS issue types to SARIF 2.1 result levels (SPEC §2.2).
func SeverityFromIssueType(issueType string) string {
	switch strings.ToLower(tools.JavaStringTrim(issueType)) {
	case "misspelling", "grammar", "non-conformance":
		// Java confusion / non-conformance treated as error-level (with misspelling/grammar).
		return "error"
	case "style", "register", "locale-violation", "locale-specific-content":
		return "note"
	case "":
		return "warning"
	default:
		// whitespace, typographical, duplication, uncategorized, …
		return "warning"
	}
}

// RuleURL returns the LanguageTool community rule page for a rule ID.
// lang defaults via RuleLangHint then "en".
func RuleURL(ruleID, lang string) string {
	if ruleID == "" {
		return ""
	}
	if lang == "" {
		lang = RuleLangHint(ruleID)
	}
	if lang == "" {
		lang = "en"
	}
	if i := strings.IndexByte(lang, '-'); i > 0 {
		lang = lang[:i]
	}
	return "https://community.languagetool.org/rule/show/" + ruleID + "?lang=" + lang
}

// RuleLangHint infers a language code from a rule ID prefix (e.g. DE_… → de).
// Only known LT language codes; empty if unknown.
// Also maps well-known DE rule ID shapes that are not 2–3 letter codes:
// GERMAN_*/SWISS_*/AUSTRIAN_*, *_DE suffix (Java de statistic/style IDs),
// MORFOLOGIK_RULE_{lang}[_variant] (Java Morfologik*SpellerRule IDs),
// and a small set of DE-only Java rule IDs without those markers — not invent.
func RuleLangHint(ruleID string) string {
	up := strings.ToUpper(tools.JavaStringTrim(ruleID))
	// DE-specific long prefixes used by German* rules (not lang-code_RULE).
	if strings.HasPrefix(up, "GERMAN_") || strings.HasPrefix(up, "SWISS_") ||
		strings.HasPrefix(up, "AUSTRIAN_") {
		return "de"
	}
	// Java AI_DE_* remote / GGEC / Hydra rule IDs (German AI rule families).
	if strings.HasPrefix(up, "AI_DE_") {
		return "de"
	}
	// Java AI_FR_* remote / GGEC / Hydra / MERGED rule IDs (French AI rule families).
	if strings.HasPrefix(up, "AI_FR_") {
		return "fr"
	}
	// Java AI_ES_* remote / GGEC rule IDs (Spanish AI rule families).
	if strings.HasPrefix(up, "AI_ES_") {
		return "es"
	}
	// Java AI_PT_* / AI_NL_* remote rule families.
	if strings.HasPrefix(up, "AI_PT_") {
		return "pt"
	}
	if strings.HasPrefix(up, "AI_NL_") {
		return "nl"
	}
	// Java English AI / Hydra families (English.getPriorityForId AI_HYDRA / AI_EN / AI_SPELLING).
	if strings.HasPrefix(up, "AI_EN_") || strings.HasPrefix(up, "AI_HYDRA_") ||
		strings.HasPrefix(up, "AI_SPELLING_") {
		return "en"
	}
	// Java Morfologik*SpellerRule: MORFOLOGIK_RULE_DE_DE, MORFOLOGIK_RULE_EN_US, …
	// (first '_' is after MORFOLOGIK — not a 2–3 letter lang prefix).
	if strings.HasPrefix(up, "MORFOLOGIK_RULE_") {
		if p := morfologikRuleLang(up); p != "" {
			return p
		}
	}
	// Java DE statistic / style rule IDs end with _DE (FILLER_WORDS_DE, …).
	if strings.HasSuffix(up, "_DE") {
		return "de"
	}
	// DE-only Java rule IDs without DE_ prefix / _DE suffix.
	switch up {
	case "MISSING_VERB", "OLD_SPELLING_RULE", "COMPOUND_INFINITIV_RULE",
		"EINHEITEN_METRISCH", "REDUNDANT_MODAL_VERB",
		"COMMA_IN_FRONT_RELATIVE_CLAUSE", "COMMA_BEHIND_RELATIVE_CLAUSE",
		"STYLE_REPEATED_SHORT_SENTENCES", "STYLE_REPEATED_SENTENCE_BEGINNING",
		"UNPAIRED_BRACKETS": // GermanUnpairedBracketsRule keeps legacy ID without DE_
		return "de"
	}
	i := strings.IndexByte(up, '_')
	if i < 2 || i > 3 {
		return ""
	}
	return knownLangCode(strings.ToLower(up[:i]))
}

// morfologikRuleLang extracts the language short code from
// MORFOLOGIK_RULE_{lang} or MORFOLOGIK_RULE_{lang}_{variant} (Java speller IDs).
func morfologikRuleLang(up string) string {
	const pfx = "MORFOLOGIK_RULE_"
	if !strings.HasPrefix(up, pfx) {
		return ""
	}
	rest := up[len(pfx):]
	if rest == "" {
		return ""
	}
	code := rest
	if i := strings.IndexByte(rest, '_'); i > 0 {
		code = rest[:i]
	}
	return knownLangCode(strings.ToLower(code))
}

// knownLangCode returns lang if it is a known LT short code used in rule IDs.
func knownLangCode(p string) string {
	switch p {
	case "en", "de", "fr", "es", "pt", "it", "nl", "pl", "ru", "uk", "sv", "da",
		"ca", "gl", "sk", "ro", "el", "ar", "fa", "ga", "br", "eo", "sl", "sr",
		"be", "is", "ja", "km", "lt", "ml", "ta", "tl", "zh", "ast", "crh":
		return p
	default:
		return ""
	}
}
