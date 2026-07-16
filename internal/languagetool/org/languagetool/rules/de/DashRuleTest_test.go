package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/DashRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/DashRuleTest.java :: DashRuleTest.testRule
func TestDashRule_Rule(t *testing.T) {
	_ = "Die große Diäten-Erhöhung kam dann doch." // assertGood
	_ = "Die große Diätenerhöhung kam dann doch." // assertGood
	_ = "Die große Diäten-Erhöhungs-Manie kam dann doch." // assertGood
	_ = "Die große Diäten- und Gehaltserhöhung kam dann doch." // assertGood
	_ = "Die große Diäten- sowie Gehaltserhöhung kam dann doch." // assertGood
	_ = "Die große Diäten- oder Gehaltserhöhung kam dann doch." // assertGood
	_ = "Erst so - Karl-Heinz dann blah." // assertGood
	_ = "Erst so -- Karl-Heinz aber..." // assertGood
	_ = "Nord- und Südkorea" // assertGood
	_ = "NORD- UND SÜDKOREA" // assertGood
	_ = "NORD- BZW. SÜDKOREA" // assertGood
	_ = "Die große Diäten- Erhöhung kam dann doch." // assertBad
	_ = "Die große Diäten-  Erhöhung kam dann doch." // assertBad
	_ = "Die große Diäten-Erhöhungs- Manie kam dann doch." // assertBad
	_ = "Die große Diäten- Erhöhungs-Manie kam dann doch." // assertBad
	_ = "MAZEDONIEN- SKOPJE Str." // assertBad
}
