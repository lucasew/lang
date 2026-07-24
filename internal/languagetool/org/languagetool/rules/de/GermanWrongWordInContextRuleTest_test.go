package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWrongWordInContextRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanWrongWordInContextRule_Rule(t *testing.T) {
	rule := NewGermanWrongWordInContextRule(nil)
	assertGood := func(sentence string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(sentence))), "good: %q", sentence)
	}
	assertBad := func(sentence string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain(sentence))), "bad: %q", sentence)
	}

	// Laiche/Leiche
	assertBad("Eine Laiche ist ein toter Körper.")
	assertGood("Eine Leiche ist ein toter Körper.")
	assertGood("Die Leichen der Verstorbenen wurden ins Wasser geworfen.")

	// Mine/Miene
	assertGood("Er verzieht keine Miene.")
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
	require.Equal(t, "Minen", rule.Match(languagetool.AnalyzePlain("Er hat das mit den Mienen weggesprengt."))[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Miene", rule.Match(languagetool.AnalyzePlain("Er versucht, keine Mine zu verziehen."))[0].GetSuggestedReplacements()[0])

	// Neutron/Neuron
	assertGood("Nervenzellen nennt man Neuronen")
	assertGood("Das Neutron ist elektisch neutral")
	assertBad("Atomkerne bestehen aus Protonen und Neuronen")
	assertBad("Über eine Synapse wird das Neutron mit einer bestimmten Zelle verknüpft und nimmt mit der lokal zugeordneten postsynaptischen Membranregion eines Dendriten Signale auf.")
	require.Equal(t, "Neutronen", rule.Match(languagetool.AnalyzePlain("Protonen und Neuronen sind Bausteine des Atomkerns"))[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Neurons", rule.Match(languagetool.AnalyzePlain("Das Axon des Neutrons ..."))[0].GetSuggestedReplacements()[0])

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
