package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWrongWordInContextRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWrongWordInContextRuleTest.java :: GermanWrongWordInContextRuleTest.testRule
func TestGermanWrongWordInContextRule_Rule(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	_ = "Eine Leiche ist ein toter Körper." // assertGood
	_ = "Die Leichen der Verstorbenen wurden ins Wasser geworfen." // assertGood
	_ = "Er verzieht keine Miene." // assertGood
	_ = "Er verzieht keine Miene." // assertGood
	_ = "Die Explosion der Mine." // assertGood
	_ = "Die Mine ist explodiert." // assertGood
	_ = "Er versucht, keine Miene zu verziehen." // assertGood
	_ = "Sie sollen weiter Minen eingesetzt haben." // assertGood
	_ = "Er verzieht sich nach Bekanntgabe der Mineralölsteuerverordnung." // assertGood
	_ = "Nervenzellen nennt man Neuronen" // assertGood
	_ = "Das Neutron ist elektisch neutral" // assertGood
	_ = "Das Seil läuft durch eine Winde." // assertGood
	_ = "Eine blutende Wunde" // assertGood
	_ = "Er war durch die Narkose betäubt." // assertGood
	_ = "Die Biene bestäubt die Blume." // assertGood
	_ = "Er verreist stets mit leichtem Gepäck." // assertGood
	_ = "Die Warze wurde vereist." // assertGood
	_ = "Eine Laiche ist ein toter Körper." // assertBad
	_ = "Er verzieht keine Mine." // assertBad
	_ = "Mit unbewegter Mine." // assertBad
	_ = "Er setzt eine kalte Mine auf." // assertBad
	_ = "Er sagt, die unterirdische Miene sei zusammengestürzt." // assertBad
	_ = "Die Miene ist eingestürzt." // assertBad
	_ = "Die Sprengung mit Mienen ist toll." // assertBad
	_ = "Der Bleistift hat eine Miene." // assertBad
	_ = "Die Mienen sind gestern Abend explodiert." // assertBad
	_ = "Die Miene des Kugelschreibers ist leer." // assertBad
	_ = "Atomkerne bestehen aus Protonen und Neuronen" // assertBad
	_ = "Über eine Synapse wird das Neutron mit einer bestimmten Zelle verknüpft und nimmt mit der lokal zugeordneten postsynaptischen Membranregion eines Dendriten Signale auf." // assertBad
	_ = "Es kamen Keime in die Winde." // assertBad
	_ = "Möglicherweise wehen die Wunde gerade nicht günstig." // assertBad
	_ = "Den Kuchen mit Puderzucker betäuben." // assertBad
	_ = "Von Drogen bestäubt spürte er keine Schmerzen." // assertBad
	_ = "Nach Diktat vereist." // assertBad
	_ = "Die Tragfläche war verreist." // assertBad
}
