package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWrongWordInContextRuleTest.java
//
// Java setup:
//   Language german = Languages.getLanguageForShortCode("de-DE");
//   lt = new JLanguageTool(german);
//   rule = new GermanWrongWordInContextRule(null, german);
//   assertGood/Bad: rule.match(lt.getAnalyzedSentence(sentence)).length
//
// Go: NewGermanWrongWordInContextRule + de-DE JLanguageTool.GetAnalyzedSentence.
// Only active (non-commented) Java cases are ported; Lid/Lied, malen/mahlen, Saite/Seite stay skipped.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestGermanWrongWordInContextRule_Rule(t *testing.T) {
	// Java: Languages.getLanguageForShortCode("de-DE") + new JLanguageTool(german)
	lt := languagetool.NewJLanguageTool("de-DE")
	rule := NewGermanWrongWordInContextRule(nil)

	assertGood := func(sentence string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", sentence)
		require.Equal(t, 0, len(rule.Match(sent)), "good: %q", sentence)
	}
	assertBad := func(sentence string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", sentence)
		require.Equal(t, 1, len(rule.Match(sent)), "bad: %q", sentence)
	}
	suggest := func(sentence, want string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(sentence)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", sentence)
		ms := rule.Match(sent)
		require.Len(t, ms, 1, "suggest: %q", sentence)
		reps := ms[0].GetSuggestedReplacements()
		require.NotEmpty(t, reps, "suggest: %q", sentence)
		require.Equal(t, want, reps[0], "suggest: %q", sentence)
	}

	// Laiche/Leiche
	assertBad("Eine Laiche ist ein toter Körper.")
	assertGood("Eine Leiche ist ein toter Körper.")
	assertGood("Die Leichen der Verstorbenen wurden ins Wasser geworfen.")

	// Lid/Lied — Java commented out (too many FPs)
	// malen/mahlen — Java commented out

	// Mine/Miene
	assertGood("Er verzieht keine Miene.")
	assertGood("Er verzieht keine Miene.") // Java duplicates this assert
	assertGood("Die Explosion der Mine.")
	assertGood("Die Mine ist explodiert.")
	assertGood("Er versucht, keine Miene zu verziehen.")
	assertGood("Sie sollen weiter Minen eingesetzt haben.")
	assertGood("Er verzieht sich nach Bekanntgabe der Mineralölsteuerverordnung.")
	assertBad("Er verzieht keine Mine.")
	assertBad("Mit unbewegter Mine.")
	assertBad("Er setzt eine kalte Mine auf.")
	assertBad("Er sagt, die unterirdische Miene sei zusammengestürzt.")
	assertBad("Die Miene ist eingestürzt.")
	assertBad("Die Sprengung mit Mienen ist toll.")
	assertBad("Der Bleistift hat eine Miene.")
	assertBad("Die Mienen sind gestern Abend explodiert.")
	assertBad("Die Miene des Kugelschreibers ist leer.")
	suggest("Er hat das mit den Mienen weggesprengt.", "Minen")
	suggest("Er versucht, keine Mine zu verziehen.", "Miene")

	// Saite/Seite — Java commented out

	// Neutron/Neuron
	assertGood("Nervenzellen nennt man Neuronen")
	assertGood("Das Neutron ist elektisch neutral")
	assertBad("Atomkerne bestehen aus Protonen und Neuronen")
	assertBad("Über eine Synapse wird das Neutron mit einer bestimmten Zelle verknüpft und nimmt mit der lokal zugeordneten postsynaptischen Membranregion eines Dendriten Signale auf.")
	suggest("Protonen und Neuronen sind Bausteine des Atomkerns", "Neutronen")
	suggest("Das Axon des Neutrons ...", "Neurons")

	// Wunde/Winde
	assertGood("Das Seil läuft durch eine Winde.")
	assertGood("Eine blutende Wunde")
	assertBad("Es kamen Keime in die Winde.")
	assertBad("Möglicherweise wehen die Wunde gerade nicht günstig.")

	// betäuben/bestäuben
	assertGood("Er war durch die Narkose betäubt.")
	assertGood("Die Biene bestäubt die Blume.")
	assertBad("Den Kuchen mit Puderzucker betäuben.")
	assertBad("Von Drogen bestäubt spürte er keine Schmerzen.")

	// ver(r)eisen
	assertGood("Er verreist stets mit leichtem Gepäck.")
	assertGood("Die Warze wurde vereist.")
	assertBad("Nach Diktat vereist.")
	assertBad("Die Tragfläche war verreist.")
}

// Java GermanWrongWordInContextRule: id, category, messages, example Miene→Mine.
func TestGermanWrongWordInContextRule_Metadata(t *testing.T) {
	rule := NewGermanWrongWordInContextRule(nil)
	require.Equal(t, "GERMAN_WRONG_WORD_IN_CONTEXT", rule.GetID())
	require.Equal(t, "Mögliche Wortverwechslungen: $match", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.CategoryConfusedWords, rule.GetCategory().GetID())
	require.Equal(t, "Leicht zu verwechselnde Wörter", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	require.Equal(t, "Mögliche Wortverwechslung: Meinten Sie <suggestion>$SUGGESTION</suggestion> anstatt '$WRONGWORD'?", rule.MessageString)
	require.Equal(t, "Mögliche Wortverwechslung", rule.ShortMessageString)
	require.Equal(t, "Mögliche Wortverwechslung: Meinten Sie <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) anstatt '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?", rule.LongMessageString)

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, []string{"Mine"}, inc[0].GetCorrections())
	cor := rule.GetCorrectExamples()
	require.Len(t, cor, 1)
}
