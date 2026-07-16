package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java :: VerbAgreementRuleTest.testSuggestionSorting
func TestVerbAgreementRule_SuggestionSorting(t *testing.T) {
	// contains assertThat
}

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java :: VerbAgreementRuleTest.testPositions
func TestVerbAgreementRule_Positions(t *testing.T) {
	// contains assertThat
}

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java :: VerbAgreementRuleTest.testWrongVerb
func TestVerbAgreementRule_WrongVerb(t *testing.T) {
	_ = "*runterguck* das ist aber tief" // assertGood
	_ = "Weder Peter noch ich wollen das." // assertGood
	_ = "Du bist in dem Moment angekommen, als ich gegangen bin." // assertGood
	_ = "Kümmere du dich mal nicht darum!" // assertGood
	_ = "Ich weiß, was ich tun werde, falls etwas geschehen sollte." // assertGood
	_ = "...die dreißig Jahre jünger als ich ist." // assertGood
	_ = "Ein Mann wie ich braucht einen Hut." // assertGood
	_ = "Egal, was er sagen wird, ich habe meine Entscheidung getroffen." // assertGood
	_ = "Du Beharrst darauf, dein Wörterbuch hätte recht, hast aber von den Feinheiten des Japanischen keine Ahnung!" // assertGood
	_ = "Bin gleich wieder da." // assertGood
	_ = "Wobei ich äußerst vorsichtig bin." // assertGood
	_ = "Es ist klar, dass ich äußerst vorsichtig mit den Informationen umgehe" // assertGood
	_ = "Es ist klar, dass ich äußerst vorsichtig bin." // assertGood
	_ = "Wobei er äußerst selten darüber spricht." // assertGood
	_ = "Wobei er äußerst selten über seine erste Frau spricht." // assertGood
	_ = "Das Wort „schreibst“ ist schön." // assertGood
	_ = "Die Jagd nach bin Laden." // assertGood
	_ = "Die Unterlagen solltet ihr gründlich durcharbeiten." // assertGood
	_ = "Er reagierte äußerst negativ." // assertGood
	_ = "Max und ich sollten das machen." // assertGood
	_ = "Osama bin Laden stammt aus Saudi-Arabien." // assertGood
	_ = "Solltet ihr das machen?" // assertGood
	_ = "Dann beende du den Auftrag und bring sie ihrem Vater." // assertGood
	_ = "- Wirst du ausflippen?" // assertGood
	_ = "Ein Geschenk, das er einst von Aphrodite erhalten hatte." // assertGood
	_ = "Wenn ich sterben sollte, wer würde sich dann um die Katze kümmern?" // assertGood
	_ = "Wenn er sterben sollte, wer würde sich dann um die Katze kümmern?" // assertGood
	_ = "Wenn sie sterben sollte, wer würde sich dann um die Katze kümmern?" // assertGood
	_ = "Wenn es sterben sollte, wer würde sich dann um die Katze kümmern?" // assertGood
	_ = "Wenn ihr sterben solltet, wer würde sich dann um die Katze kümmern?" // assertGood
	_ = "Wenn wir sterben sollten, wer würde sich dann um die Katze kümmern?" // assertGood
	_ = "Dafür erhielten er sowie der Hofgoldschmied Theodor Heiden einen Preis." // assertGood
	_ = "Probst wurde deshalb in den Medien gefeiert." // assertGood
	_ = "/usr/bin/firefox" // assertGood
	_ = "Das sind Leute, die viel mehr als ich wissen." // assertGood
	_ = "Das ist mir nicht klar, kannst ja mal beim Kunden nachfragen." // assertGood
	_ = "So tes\u00ADtest Du das mit dem soft hyphen." // assertGood
	_ = "Viele Brunnen in Italiens Hauptstadt sind bereits abgeschaltet." // assertGood
	_ = "„Werde ich tun!“" // assertGood
	_ = "Könntest dir mal eine Scheibe davon abschneiden!" // assertGood
	_ = "Müsstest dir das mal genauer anschauen." // assertGood
	_ = "Kannst ein neues Release machen." // assertGood
	_ = "Sie fragte: „Muss ich aussagen?“" // assertGood
	_ = "„Können wir bitte das Thema wechseln, denn ich möchte ungern darüber reden?“" // assertGood
	_ = "Er sagt: „Willst du behaupten, dass mein Sohn euch liebt?“" // assertGood
	_ = "Kannst mich gerne anrufen." // assertGood
	_ = "Kannst ihn gerne anrufen." // assertGood
	_ = "Kannst sie gerne anrufen." // assertGood
	_ = "Aber wie ich sehe, benötigt ihr Nachschub." // assertGood
	_ = "Wie ich sehe, benötigt ihr Nachschub." // assertGood
	_ = "Einer wie du kennt doch bestimmt viele Studenten." // assertGood
	_ = "Für Sie mache ich eine Ausnahme." // assertGood
	_ = "Ohne sie hätte ich das nicht geschafft." // assertGood
	_ = "Ohne Sie hätte ich das nicht geschafft." // assertGood
	_ = "Ich hoffe du auch." // assertGood
	_ = "Ich hoffe ihr auch." // assertGood
	_ = "Wird hoffen du auch." // assertGood
	_ = "Hab einen schönen Tag!" // assertGood
	_ = "Tom traue ich mehr als Maria." // assertGood
	_ = "Tom kenne ich nicht besonders gut, dafür aber seine Frau." // assertGood
	_ = "Tom habe ich heute noch nicht gesehen." // assertGood
	_ = "Tom bezahle ich gut." // assertGood
	_ = "Tom werde ich nicht noch mal um Hilfe bitten." // assertGood
	_ = "Tom konnte ich überzeugen, nicht aber Maria." // assertGood
	_ = "Mach du mal!" // assertGood
	_ = "Das bekomme ich nicht hin." // assertGood
	_ = "Dies betreffe insbesondere Nietzsches Aussagen zu Kant und der Evolutionslehre." // assertGood
	_ = "❌Du fühlst Dich unsicher?" // assertGood
	_ = "Bringst nicht einmal so etwas Einfaches zustande!" // assertGood
	_ = "Bekommst sogar eine Sicherheitszulage" // assertGood
	_ = "Dallun sagte nur, dass er gleich kommen wird und legte wieder auf." // assertGood
	_ = "Tinne, Elvis und auch ich werden gerne wiederkommen!" // assertGood
	_ = "Du bist Lehrer und weißt diese Dinge nicht?" // assertGood
	_ = "Die Frage lautet: Bist du bereit zu helfen?" // assertGood
	_ = "Ich will nicht so wie er enden." // assertGood
	_ = "Das heißt, wir geben einander oft nach als gute Freunde, ob wir gleich nicht einer Meinung sind." // assertGood
	_ = "Wir seh'n uns in Berlin." // assertGood
	_ = "Bist du bereit, darüber zu sprechen?" // assertGood
	_ = "Bist du schnell eingeschlafen?" // assertGood
	_ = "Im Gegenzug bin ich bereit, beim Türkischlernen zu helfen." // assertGood
	_ = "Das habe ich lange gesucht." // assertGood
	_ = "Dann solltest du schnell eine Nummer der sexy Omas wählen." // assertGood
	_ = "Vielleicht würdest du bereit sein, ehrenamtlich zu helfen." // assertGood
	_ = "Werde nicht alt, egal wie lange du lebst." // assertGood
	_ = "Du bist hingefallen und hast dir das Bein gebrochen." // assertGood
	_ = "Mögest du lange leben!" // assertGood
	_ = "Planst du lange hier zu bleiben?" // assertGood
	_ = "Du bist zwischen 11 und 12 Jahren alt und spielst gern Fußball bzw. möchtest damit anfangen?" // assertGood
	_ = "Ein großer Hadithwissenschaftler, Scheich Şemseddin Mehmed bin Muhammed-ül Cezri, kam in der Zeit von Mirza Uluğ Bey nach Semerkant." // assertGood
	_ = "Die Prüfbescheinigung bekommst du gleich nach der bestanden Prüfung vom Prüfer." // assertGood
	_ = "Du bist sehr schön und brauchst überhaupt gar keine Schminke zu verwenden." // assertGood
	_ = "Ist das so schnell, wie du gehen kannst?" // assertGood
	_ = "Egal wie lange du versuchst, die Leute davon zu überzeugen" // assertGood
	_ = "Du bist verheiratet und hast zwei Kinder." // assertGood
	_ = "Du bist aus Berlin und wohnst in Bonn." // assertGood
	_ = "Sie befestigen die Regalbretter vermittelst dreier Schrauben." // assertGood
	_ = "Meine Familie & ich haben uns ein neues Auto gekauft." // assertGood
	_ = "Der Bescheid lasse im übrigen die Abwägungen vermissen, wie die Betriebsprüfung zu den Sachverhaltsbeurteilungen gelange, die den von ihr bekämpften Bescheiden zugrundegelegt worden seien." // assertGood
	_ = "Die Bildung des Samens erfolgte laut Alkmaion im Gehirn, von wo aus er durch die Adern in den Hoden gelange." // assertGood
	_ = "Michael Redmond (geb. 1963, USA)." // assertGood
	_ = "Würd mich sehr freuen drüber." // assertGood
	_ = "Es würd' ein jeder Doktor sein, wenn's Wissen einging wie der Wein." // assertGood
	_ = "Bald merkte er, dass er dank seines Talents nichts mehr in der österreichischen Jazzszene lernen konnte." // assertGood
	_ = "»Alles, was wir dank dieses Projektes sehen werden, wird für uns neu sein«, so der renommierte Bienenforscher." // assertGood
	_ = "Und da wir äußerst Laissez-faire sind, kann man das auch machen." // assertGood
	_ = "Duzen, jemanden mit Du anreden, eine Sitte, die bei allen alten Völkern üblich war." // assertGood
	_ = "Schreibtischtäter wie Du sind doch eher selten." // assertGood
	_ = "Nee, geh du!" // assertGood
	_ = "Als Borcarbid weißt es eine hohe Härte auf." // assertBad
	_ = "Das greift auf Vorläuferinstitutionen bist auf die Zeit von 1234 zurück." // assertBad
	_ = "Die Eisenbahn dienst überwiegend dem Güterverkehr." // assertBad
	_ = "Die Unterlagen solltest ihr gründlich durcharbeiten." // assertBad
	_ = "Peter bin nett." // assertBad
	_ = "Weiter befindest sich im Osten die Gemeinde Dorf." // assertBad
	_ = "Ich geht jetzt nach Hause, weil ich schon zu spät bin." // assertBad
	_ = "„Du muss gehen.“" // assertBad
	_ = "Du weiß es doch." // assertBad
	_ = "Sie sagte zu mir: „Du muss gehen.“" // assertBad
	_ = "„Ich müsst alles machen.“" // assertBad
	_ = "„Ich könnt mich sowieso nicht verstehen.“" // assertBad
	_ = "Er sagte düster: Ich brauchen mich nicht böse angucken." // assertBad
	_ = "David sagte düster: Ich brauchen mich nicht böse angucken." // assertBad
	_ = "Ich setzet mich auf den weichen Teppich und kreuzte die Unterschenkel wie ein Japaner." // assertBad
	_ = "Ich brauchen einen Karren mit zwei Ochsen." // assertBad
	_ = "Ich haben meinen Ohrring fallen lassen." // assertBad
	_ = "Ich stehen Ihnen gerne für Rückfragen zur Verfügung." // assertBad
}

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java :: VerbAgreementRuleTest.testWrongVerbSubject
func TestVerbAgreementRule_WrongVerbSubject(t *testing.T) {
	_ = "Auch morgen lebe ich." // assertGood
	_ = "Auch morgen leben wir noch." // assertGood
	_ = "Auch morgen lebst du." // assertGood
	_ = "Auch morgen lebt er." // assertGood
	_ = "Auch wenn du leben möchtest." // assertGood
	_ = "auf der er sieben Jahre blieb." // assertGood
	_ = "Das absolute Ich ist nicht mit dem individuellen Geist zu verwechseln." // assertGood
	_ = "Das Ich ist keine Einbildung" // assertGood
	_ = "Das lyrische Ich ist verzweifelt." // assertGood
	_ = "Den Park, von dem er äußerst genaue Karten zeichnete." // assertGood
	_ = "Der auffälligste Ring ist der erster Ring, obwohl er verglichen mit den anderen Ringen sehr schwach erscheint." // assertGood
	_ = "Der Fehler, falls er bestehen sollte, ist schwerwiegend." // assertGood
	_ = "Der Vorfall, bei dem er einen Teil seines Vermögens verloren hat, ist lange vorbei." // assertGood
	_ = "Diese Lösung wurde in der 64'er beschrieben, kam jedoch nie." // assertGood
	_ = "Die Theorie, mit der ich arbeiten konnte." // assertGood
	_ = "Du bist nett." // assertGood
	_ = "Du kannst heute leider nicht kommen." // assertGood
	_ = "Du lebst." // assertGood
	_ = "Du wünschst dir so viel." // assertGood
	_ = "Er geht zu ihr." // assertGood
	_ = "Er ist nett." // assertGood
	_ = "Er kann heute leider nicht kommen." // assertGood
	_ = "Er lebt." // assertGood
	_ = "Er wisse nicht, ob er lachen oder weinen solle." // assertGood
	_ = "Er und du leben." // assertGood
	_ = "Er und ich leben." // assertGood
	_ = "Falls er bestehen sollte, gehen sie weg." // assertGood
	_ = "Heere, des Gottes der Schlachtreihen Israels, den du verhöhnt hast." // assertGood
	_ = "Ich bin" // assertGood
	_ = "Ich bin Frankreich!" // assertGood
	_ = "Ich bin froh, dass ich arbeiten kann." // assertGood
	_ = "Ich bin nett." // assertGood
	_ = "‚ich bin tot‘" // assertGood
	_ = "Ich kann heute leider nicht kommen." // assertGood
	_ = "Ich lebe." // assertGood
	_ = "Lebst du?" // assertGood
	_ = "Morgen kommen du und ich." // assertGood
	_ = "Morgen kommen er, den ich sehr mag, und ich." // assertGood
	_ = "Morgen kommen er und ich." // assertGood
	_ = "Morgen kommen ich und sie." // assertGood
	_ = "Morgen kommen wir und sie." // assertGood
	_ = "nachdem er erfahren hatte" // assertGood
	_ = "Nett bin ich." // assertGood
	_ = "Nett bist du." // assertGood
	_ = "Nett ist er." // assertGood
	_ = "Nett sind wir." // assertGood
	_ = "Niemand ahnte, dass er gewinnen könne." // assertGood
	_ = "Sie lebt und wir leben." // assertGood
	_ = "Sie und er leben." // assertGood
	_ = "Sind ich und Peter nicht nette Kinder?" // assertGood
	_ = "Sodass ich sagen möchte, dass unsere schönen Erinnerungen gut sind." // assertGood
	_ = "Wann ich meinen letzten Film drehen werde, ist unbekannt." // assertGood
	_ = "Was ich tun muss." // assertGood
	_ = "Welche Aufgaben er dabei tatsächlich übernehmen könnte" // assertGood
	_ = "wie er beschaffen war" // assertGood
	_ = "Wir gelangen zu dir." // assertGood
	_ = "Wir können heute leider nicht kommen." // assertGood
	_ = "Wir leben noch." // assertGood
	_ = "Wir sind nett." // assertGood
	_ = "Wobei wir benutzt haben, dass der Satz gilt." // assertGood
	_ = "Wünschst du dir mehr Zeit?" // assertGood
	_ = "Wyrjtjbst du?" // assertGood
	_ = "Wenn ich du wäre, würde ich das nicht machen." // assertGood
	_ = "Er sagte: „Darf ich bitten, mir zu folgen?“" // assertGood
	_ = "Ja sind ab morgen dabei." // assertGood
	_ = "Oh bin überfragt." // assertGood
	_ = "Angenommen, du wärst ich." // assertGood
	_ = "Ich denke, dass das Haus, in das er gehen will, heute Morgen gestrichen worden ist." // assertGood
	_ = "Ich hab mein Leben, leb du deines!" // assertGood
	_ = "Da freut er sich, wenn er schlafen geht und was findet." // assertGood
	_ = "John nimmt weiter an einem Abendkurs über Journalismus teil." // assertGood
	_ = "Viele nahmen an der Aktion teil und am Ende des rAAd-Events war die Tafel zwar bunt, aber leider überwogen die roten Kärtchen sehr deutlich." // assertGood
	_ = "Musst also nichts machen." // assertGood
	_ = "Eine Situation, wo der Stadtrat gleich mal zum Du übergeht." // assertGood
	_ = "Machen wir, sobald wir frische und neue Akkus haben." // assertGood
	_ = "Darfst nicht so reden, Franz!" // assertGood
	_ = "Finde du den Jungen." // assertGood
	_ = "Finde Du den Jungen." // assertGood
	_ = "Kümmerst dich ja gar nicht um sie." // assertGood
	_ = "Könntest was erfinden, wie dein Papa." // assertGood
	_ = "Siehst aus wie ein Wachhund." // assertGood
	_ = "Solltest es mal in seinem Büro versuchen." // assertGood
	_ = "Stehst einfach nicht zu mir." // assertGood
	_ = "Stellst für deinen Dad etwas zu Essen bereit." // assertGood
	_ = "Springst weit, oder?" // assertGood
	_ = "Wirst groß, was?" // assertGood
	_ = "Auch morgen leben du." // assertBad
	_ = "Du weiß noch, dass du das gestern gesagt hast." // assertBad
	_ = "Auch morgen leben du" // assertBad
	_ = "Auch morgen leben er." // assertBad
	_ = "Auch morgen leben ich." // assertBad
	_ = "Auch morgen lebte wir noch." // assertBad
	_ = "Du bin nett." // assertBad
	_ = "Du können heute leider nicht kommen." // assertBad
	_ = "Du können heute leider nicht kommen." // assertBad
	_ = "Du leben." // assertBad
	_ = "Du wünscht dir so viel." // assertBad
	_ = "Er bin nett." // assertBad
	_ = "Er gelangst zu ihr." // assertBad
	_ = "Er können heute leider nicht kommen." // assertBad
	_ = "Er lebst." // assertBad
	_ = "Ich bist nett." // assertBad
	_ = "Ich kannst heute leider nicht kommen." // assertBad
	_ = "Ich leben." // assertBad
	_ = "Ich leben." // assertBad
	_ = "Lebe du?" // assertBad
	_ = "Lebe du?" // assertBad
	_ = "Leben du?" // assertBad
	_ = "Nett bist ich nicht." // assertBad
	_ = "Nett bist ich nicht." // assertBad
	_ = "Nett sind du." // assertBad
	_ = "Nett sind er." // assertBad
	_ = "Nett sind er." // assertBad
	_ = "Nett warst wir." // assertBad
	_ = "Wir bin nett." // assertBad
	_ = "Wir gelangst zu ihr." // assertBad
	_ = "Wir könnt heute leider nicht kommen." // assertBad
	_ = "Wünscht du dir mehr Zeit?" // assertBad
	_ = "Wir lebst noch." // assertBad
	_ = "Wir lebst noch." // assertBad
	_ = "Er sagte düster: „Ich brauchen mich nicht schuldig fühlen.“" // assertBad
	_ = "Er sagte: „Ich brauchen mich nicht schuldig fühlen.“" // assertBad
}
