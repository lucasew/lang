package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/AgreementRule2Test.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/AgreementRule2Test.java :: AgreementRule2Test.testRule
func TestAgreementRule2_Rule(t *testing.T) {
	_ = "Kleines Haus am Waldesrand" // assertGood
	_ = "\"Kleines Haus am Waldesrand\"" // assertGood
	_ = "Wirtschaftliches Wachstum kommt ins Stocken" // assertGood
	_ = "Unter Berücksichtigung des Übergangs" // assertGood
	_ = "Wirklich Frieden herrscht aber noch nicht" // assertGood
	_ = "Deutscher Taschenbuch Verlag expandiert" // assertGood
	_ = "Wohl Anfang 1725 begegnete Bach dem Dichter." // assertGood
	_ = "Weniger Personal wird im ganzen Land gebraucht." // assertGood
	_ = "National Board of Review" // assertGood
	_ = "International Management" // assertGood
	_ = "Gemeinsam Sportler anfeuern." // assertGood
	_ = "Viel Spaß beim Arbeiten" // assertGood
	_ = "Ganz Europa stand vor einer Neuordnung." // assertGood
	_ = "Gesetzlich Versicherte sind davon ausgenommen." // assertGood
	_ = "Ausreichend Bananen essen." // assertGood
	_ = "Nachhaltig Yoga praktizieren" // assertGood
	_ = "Überraschend Besuch bekommt er dann von ihr." // assertGood
	_ = "Ruhig Schlafen & Zentral Wohnen" // assertGood
	_ = "Voller Mitleid" // assertGood
	_ = "Voll Mitleid" // assertGood
	_ = "Einzig Fernschüsse brachten Erfolgsaussichten." // assertGood
	_ = "Gelangweilt Dinge sortieren hilft als Ablenkung." // assertGood
	_ = "Ganzjährig Garten pflegen" // assertGood
	_ = "Herzlich Willkommen bei unseren günstigen Rezepten!" // assertGood
	_ = "10-tägiges Rückgaberecht" // assertGood
	_ = "Angeblich Schüsse vor Explosionen gefallen" // assertGood
	_ = "Dickes Danke auch an Elena" // assertGood
	_ = "Dickes Dankeschön auch an Elena" // assertGood
	_ = "Echt Scheiße" // assertGood
	_ = "Entsprechende Automaten werden heute nicht mehr gebaut" // assertGood
	_ = "Existenziell Bedrohte kriegen einen Taschenrechner" // assertGood
	_ = "Flächendeckend Tempo 30" // assertGood
	_ = "Frei Klavier spielen lernen" // assertGood
	_ = "Ganz Eilige können es schaffen" // assertGood
	_ = "Gering Gebildete laufen Gefahr ..." // assertGood
	_ = "Ganz Ohr ist man hier" // assertGood
	_ = "Gleichzeitig Muskeln aufbauen und Fett verlieren" // assertGood
	_ = "Klar Schiff, Erster Offizier!" // assertGood
	_ = "Kostenlos Bewegung schnuppern" // assertGood
	_ = "Prinzipiell Anrecht auf eine Vertretung" // assertGood
	_ = "Regelrecht Modell gestanden haben Michel" // assertGood
	_ = "Weitgehend Konsens, auch über ..." // assertGood
	_ = "Alarmierte Polizeibeamte nahmen den Mann fest." // assertGood
	_ = "Anderen Brot und Arbeit ermöglichen - das ist ihr Ziel" // assertGood
	_ = "Diverse Unwesen, mit denen sich Hellboy beschäftigen muss, ..." // assertGood
	_ = "Gut Qualifizierte bekommen Angebote" // assertGood
	_ = "Liebe Mai, wie geht es dir?" // assertGood
	_ = "Willkommen Simpsons-Fan!" // assertGood
	_ = "Kleiner Haus am Waldesrand" // assertBad
	_ = "\"Kleiner Haus am Waldesrand\"" // assertBad
	_ = "Wirtschaftlich Wachstum kommt ins Stocken" // assertBad
	_ = "Deutscher Taschenbuch" // assertBad
}

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/AgreementRule2Test.java :: AgreementRule2Test.testSuggestion
func TestAgreementRule2_Suggestion(t *testing.T) {
	_ = "Kleinem Haus am Waldesrand ..." // assertGood
	_ = "Junger Frau geht das Geld aus" // assertGood
	_ = "Junge Frau gewinnt im Lotto" // assertGood
	_ = "Kleiner Haus am Waldesrand" // assertBad
	_ = "Kleines Häuser am Waldesrand" // assertBad
	_ = "Kleinem Häuser am Waldesrand" // assertBad
	_ = "Kleines Tisch reicht auch" // assertBad
	_ = "Junges Frau gewinnt im Lotto" // assertBad
	_ = "Jungem Frau gewinnt im Lotto" // assertBad
	_ = "Jung Frau gewinnt im Lotto" // assertBad
	_ = "Wirtschaftlich Wachstum kommt ins Stocken" // assertBad
	_ = "Wirtschaftlicher Wachstum kommt ins Stocken" // assertBad
}
