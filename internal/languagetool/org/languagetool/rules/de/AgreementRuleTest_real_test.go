package de

// Twin of AgreementRuleTest.java — real GermanTagger analysis (Java JLanguageTool path).
// Match counts and suggestions must equal Java rule.match(lt.getAnalyzedSentence(s)).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func arSetup(t *testing.T) (*languagetool.JLanguageTool, *AgreementRule) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	if !WireGermanTagWord(lt) {
		t.Skip("GermanTagger resources unavailable")
	}
	_ = WireGermanDisambiguator(lt)
	rule := WireAgreementRule(nil)
	require.NotNil(t, rule)
	return lt, rule
}

// arMatch ports rule.match(lt.getAnalyzedSentence(s)).
func arMatch(lt *languagetool.JLanguageTool, rule *AgreementRule, s string) []*rules.RuleMatch {
	sent := lt.GetAnalyzedSentence(s)
	if sent == nil {
		sents := lt.Analyze(s)
		if len(sents) == 0 {
			return nil
		}
		return rule.Match(sents[0])
	}
	return rule.Match(sent)
}

func arAssertGood(t *testing.T, lt *languagetool.JLanguageTool, rule *AgreementRule, s string) {
	t.Helper()
	ms := arMatch(lt, rule, s)
	require.Equal(t, 0, len(ms), "Found unexpected match in sentence %q: %v", s, ms)
}

func arAssertBad(t *testing.T, lt *languagetool.JLanguageTool, rule *AgreementRule, s string, expected ...string) {
	t.Helper()
	ms := arMatch(lt, rule, s)
	require.Equal(t, 1, len(ms), "Did not find one match in sentence %q (got %d)", s, len(ms))
	if len(expected) == 0 {
		return
	}
	require.Equal(t, expected, ms[0].GetSuggestedReplacements(), "suggestions for %q", s)
}

// arAssertBad1 ports assertBad1 — expected suggestion in first 4.
func arAssertBad1(t *testing.T, lt *languagetool.JLanguageTool, rule *AgreementRule, s, expected string) {
	t.Helper()
	ms := arMatch(lt, rule, s)
	require.Equal(t, 1, len(ms), "Did not find one match in sentence %q", s)
	got := ms[0].GetSuggestedReplacements()
	limit := 4
	if len(got) < limit {
		limit = len(got)
	}
	found := false
	for _, g := range got[:limit] {
		if g == expected {
			found = true
			break
		}
	}
	require.True(t, found, "Expected suggestion %q in first 4 of %v for %q", expected, got, s)
}

// Twin of AgreementRuleTest.testCompoundMatch — real GermanTagger analysis.
func TestAgreementRule_Real_CompoundMatch(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Es gibt ein Sprichwort, dem zufolge der tägliche Genuss einer Mandel dem Gedächtnis förderlich sei.`,
		`War das Eifersucht?`,
		`Sie gehörte einst zu den besten Afrikas.`,
		`Dieses Bild stammt von einem lange Zeit unbekannten Maler.`,
		`Das Staatsoberhaupt ist der Verfassung zufolge der König.`,
		`Der Ende der 1960er Jahre umgestaltete Garten ist schön.`,
		`Der Ende der achtziger Jahre umgestaltete Garten hat unter anderem ungefähr 70 verschiedene Sorten von Rosen und Volieren für exotische Vögel.`,
		`Als Vorboten des Discounthandels sind die Ende der 50er Jahre in der Bundesrepublik Deutschland wiederauflebenden Erscheinungsformen des Beziehungs-, Betriebs- und Belegschaftshandels anzusehen.`,
		`Die Anfang des letzten Monats umgestaltete Veranda ist schön.`,
		`Der Mitte 2001 umgestaltete Garten ist schön.`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
	// bad cases (exact suggestions when Java lists them)
	arAssertBad(t, lt, rule, `Das ist die Original Mail`, `die Originalmail`, `die Original-Mail`)
	arAssertBad(t, lt, rule, `Das ist die neue Original Mail`, `die neue Originalmail`, `die neue Original-Mail`)
	arAssertBad(t, lt, rule, `Das ist die ganz neue Original Mail`, `die ganz neue Originalmail`, `die ganz neue Original-Mail`)
	arAssertBad(t, lt, rule, `Doch dieser kleine Magnesium Anteil ist entscheidend.`, `dieser kleine Magnesiumanteil`, `dieser kleine Magnesium-Anteil`)
	arAssertBad(t, lt, rule, `Doch dieser sehr kleine Magnesium Anteil ist entscheidend.`, `dieser sehr kleine Magnesiumanteil`, `dieser sehr kleine Magnesium-Anteil`)
	arAssertBad(t, lt, rule, `Die Standard Priorität ist 5.`, `Die Standardpriorität`, `Die Standard-Priorität`)
	arAssertBad(t, lt, rule, `Die derzeitige Standard Priorität ist 5.`, `Die derzeitige Standardpriorität`, `Die derzeitige Standard-Priorität`)
	arAssertBad(t, lt, rule, `Ein neuer LanguageTool Account`, `Ein neuer LanguageTool-Account`)
	arAssertBad(t, lt, rule, `Danke für deine Account Daten`, `deine Accountdaten`, `deine Account-Daten`)
	arAssertBad(t, lt, rule, `Mit seinem Konkurrent Alistair Müller`, `seinem Konkurrenten`)
	arAssertBad(t, lt, rule, `Wir gehen ins Fitness Studio`, `ins Fitnessstudio`, `ins Fitness-Studio`)
	arAssertBad(t, lt, rule, `Wir gehen durchs Fitness Studio`, `durchs Fitnessstudio`, `durchs Fitness-Studio`)
	arAssertBad(t, lt, rule, `Was für eine schöner Sonnenuntergang!`, `ein schöner Sonnenuntergang`)
	arAssertBad(t, lt, rule, `Es ist ein sehr interessantes kostenloses Slot Spiel.`, `ein sehr interessantes kostenloses Slotspiel`, `ein sehr interessantes kostenloses Slot-Spiel`)
	// Java: lt.check("Er folgt damit dem Tipp des Autoren Michael Müller.") size 1, not compound msg
	ms := arMatch(lt, rule, "Er folgt damit dem Tipp des Autoren Michael Müller.")
	require.Equal(t, 1, len(ms))
	require.False(t, strings.Contains(ms[0].GetMessage(), "zusammengesetztes Nomen"))
}

// Twin of AgreementRuleTest.testDetNounRule — real GermanTagger analysis.
func TestAgreementRule_Real_DetNounRule(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Bis zur Anfang Juni geplanten Eröffnung gebe es noch einiges zu tun.`,
		`Der fließend Französisch sprechende Präsident dankt stilvoll ab.`,
		`Inwiefern soll denn das romantische Hoffnungen begründen?`,
		`Spricht der fließend Französisch?`,
		`Spricht dieser fließend Französisch, muss er viel Geld verdienen.`,
		`Der letzte Woche beschlossene Etat ist unwirksam.`,
		`Die Einen sagen dies, die Anderen das.`,
		`So ist es in den USA.`,
		`Das ist der Tisch.`,
		`Das ist das Haus.`,
		`Das ist die Frau.`,
		`Das ist das Auto der Frau.`,
		`Das gehört dem Mann.`,
		`Das Auto des Mannes.`,
		`Das interessiert den Mann.`,
		`Das interessiert die Männer.`,
		`Das Auto von einem Mann.`,
		`Das Auto eines Mannes.`,
		`Des großen Mannes.`,
		`Und nach der Nummerierung kommt die Überschrift.`,
		`Sie wiesen dieselben Verzierungen auf.`,
		`Die erwähnte Konferenz ist am Samstag.`,
		`Sie erreichten 5 Prozent.`,
		`Sie erreichten mehrere Prozent Zustimmung.`,
		`Die Bestandteile, aus denen Schwefel besteht.`,
		`Ich tat für ihn, was kein anderer Autor für ihn tat.`,
		`Ich tat für ihn, was keine andere Autorin für ihn tat.`,
		`Ich tat für ihn, was kein anderes Kind für ihn tat.`,
		`Ich tat für ihn, was dieser andere Autor für ihn tat.`,
		`Ich tat für ihn, was diese andere Autorin für ihn tat.`,
		`Ich tat für ihn, was dieses andere Kind für ihn tat.`,
		`Ich tat für ihn, was jener andere Autor für ihn tat.`,
		`Ich tat für ihn, was jeder andere Autor für ihn tat.`,
		`Ich tat für ihn, was jede andere Autorin für ihn tat.`,
		`Ich tat für ihn, was jedes andere Kind für ihn tat.`,
		`Klebe ein Preisschild auf jedes einzelne Produkt.`,
		`Eine Stadt, in der zurzeit eine rege Bautätigkeit herrscht.`,
		`... wo es zu einer regen Bautätigkeit kam.`,
		`Mancher ausscheidende Politiker hinterlässt eine Lücke.`,
		`Kern einer jeden Tragödie ist es, ..`,
		`Das wenige Sekunden alte Baby schrie laut.`,
		`Meistens sind das Frauen, die damit besser umgehen können.`,
		`Er fragte, ob das Spaß macht.`,
		`Das viele Geld wird ihr helfen.`,
		`Er verspricht jedem hohe Gewinne.`,
		`Er versprach allen Renditen jenseits von 15 Prozent.`,
		`Sind das Eier aus Bodenhaltung?`,
		`Sie sind sehr gute Kameraden, auf die Verlass ist.`,
		`Dir macht doch irgendwas Sorgen.`,
		`Sie fragte, ob das wirklich Kunst sei.`,
		`Für ihn ist das Alltag.`,
		`Für die Religiösen ist das Blasphemie und führt zu Aufständen.`,
		`Das Orange ist schön.`,
		`Dieses rötliche Orange gefällt mir am besten.`,
		`Das ist ein super Tipp.`,
		`Er nahm allen Mut zusammen und ging los.`,
		`Sie kann einem Angst einjagen.`,
		`Damit sollten zum einen neue Energien gefördert werden, zum anderen der Sozialbereich.`,
		`Nichts ist mit dieser einen Nacht zu vergleichen.`,
		`dann muss Schule dem Rechnung tragen.`,
		`Das Dach von meinem Auto.`,
		`Das Dach von meinen Autos.`,
		`Da stellt sich die Frage: Ist das Science-Fiction oder moderne Mobilität?`,
		`Er hat einen Post veröffentlicht.`,
		`Eine lückenlose Aufklärung sämtlicher physiologischer Gehirnprozesse`,
		`Sie fragte verwirrt: „Ist das Zucker?“`,
		`Er versuchte sich vorzustellen, was sein Klient für ein Mensch sei.`,
		`Sie legen ein Teilstück jenes Weges zurück, den die Tausenden Juden 1945 auf sich nehmen mussten.`,
		`Aber das ignorierte Herr Grey bewusst.`,
		`Aber das ignorierte Herr Müller bewusst.`,
		`Ich werde mich zurücknehmen und mich frischen Ideen zuwenden.`,
		`Das, plus ein eigener Firmenwagen.`,
		`Dieses leise Summen stört nicht.`,
		`Die Tiroler Küche`,
		`Was ist denn das für ein ungewöhnlicher Name?`,
		`Besonders reizen mich Fahrräder.`,
		`Und nur, weil mich psychische Erkrankungen aus der Bahn werfen`,
		`Das kostet dich Zinsen.`,
		`Sie hatten keine Chance gegen das kleinere Preußen.`,
		`Den 2019er Wert hatten sie geschätzt.`,
		`Andere formale Systeme, deren Semantiken jeweils...`,
		`Gesetz zur Änderung des Kündigungsrechts und anderer arbeitsrechtlicher Vorschriften`,
		`Die dauerhafte Abgrenzung des später Niedersachsen genannten Gebietes von Westfalen begann im 12. Jahrhundert.`,
		`Lieber jemanden, der einem Tipps gibt.`,
		`Jainas ist sogar der Genuss jeglicher tierischer Nahrungsmittel strengstens untersagt.`,
		`Es sind jegliche tierische Nahrungsmittel untersagt.`,
		`Das reicht bis weit ins heutige Hessen.`,
		`Die Customer Journey.`,
		`Für dich gehört Radfahren zum perfekten Urlaub dazu?`,
		`:D:D Leute, bitte!`,
		`Es genügt, wenn ein Mann sein eigenes Geschäft versteht und sich nicht in das anderer Leute einmischt.`,
		`Ich habe das einige Male versucht.`,
		`Und keine Märchen erzählst, die dem anderen Hoffnungen machen können.`,
		`Um diese Körpergrößen zu erreichen, war das Wachstum der Vertreter der Gattung Dinornis offenbar gegenüber dem anderer Moa-Gattungen beschleunigt`,
		`Der Schädel entspricht in den Proportionen dem anderer Vulpes-Arten, besitzt aber sehr große Paukenhöhlen, ein typisches Merkmal von Wüstenbewohnern.`,
		`Deuterium lässt sich aufgrund des großen Massenunterschieds leichter anreichern als die Isotope der anderer Elemente wie z. B. Uran.`,
		`Unklar ist, ob er zwischen der Atemseele des Menschen und der anderer Lebewesen unterschied.`,
		`Die Liechtensteiner Grenze ist im Verhältnis zu der anderer Länder kurz, da Liechtenstein ein eher kleines Land ist.`,
		`Picassos Kunstwerke werden häufiger gestohlen als die anderer Künstler.`,
		`Schreibe einen Artikel über deine Erfahrungen im Ausland oder die anderer Leute in deinem Land.`,
		`Die Bevölkerungen Chinas und Indiens lassen die anderer Staaten als Zwerge erscheinen.`,
		`Der eine mag Obst, ein anderer Gemüse, wieder ein anderer mag Fisch; allen kann man es nicht recht machen.`,
		`Mittels eines Bootloaders und zugehöriger Software kann nach jedem Anstecken des Adapters eine andere Firmware-Varianten geladen werden`,
		`Wenn sie eine andere Größe benötigen, teilen uns ihre speziellen Wünsche mit und wir unterbreiten ihnen ein Angebot über Preis und Lieferung.`,
		`Dabei wird in einer Vakuumkammer eine einige Mikrometer dicke CVD-Diamantschicht auf den Substraten abgeschieden.`,
		`1916 versuchte Gilbert Newton Lewis, die chemische Bindung durch Wechselwirkung der Elektronen eines Atoms mit einem anderen Atomen zu erklären.`,
		`Vom einen Ende der Straße zum anderen.`,
		`Er war müde vom vielen Laufen.`,
		`Sind das echte Diamanten?`,
		`Es wurde eine Verordnung erlassen, der zufolge jeder Haushalt Energie einsparen muss.`,
		`Im Jahr 1922 verlieh ihm König George V. den erblichen Titel eines Baronet. `,
		`... der zu dieser Zeit aber ohnehin schon allen Einfluss verloren hatte.`,
		`Ein Geschenk, das Maßstäbe setzt`,
		`Einwohnerzahl stieg um das Zweieinhalbfache`,
		`Die Müllers aus Hamburg.`,
		`Es ist noch unklar, wann und für wen Impfungen vorgenommen werden könnten.`,
		`Macht dir das Hoffnung?`,
		`Mich fasziniert Macht.`,
		`Der solchen Einsätzen gegenüber kritische Hitler wurde nicht im Voraus informiert.`,
		`Gregor wählte die Gestalt des wenige Jahrzehnte zuvor verstorbenen Klostergründers.`,
		`Wir machen das Januar.`,
		`Wir teilen das Morgen mit.`,
		`Wir präsentierten das vorletzten Sonnabend.`,
		`Ich release das Vormittags.`,
		`Sie aktualisieren das Montags.`,
		`Kannst du das Mittags machen?`,
		`Können Sie das nächsten Monat erledigen?`,
		`Können Sie das auch nächsten Monat erledigen?`,
		`War das Absicht?`,
		`Alles Große und Edle ist einfacher Art.`,
		`Dieser vereint Sprachprüfung, Thesaurus und Umformuliertool in einem.`,
		`Das Dach meines Autos.`,
		`Das Dach meiner Autos.`,
		`Das Dach meines großen Autos.`,
		`Das Dach meiner großen Autos.`,
		`Dann schlug er so kräftig wie er konnte mit den Schwingen.`,
		`Also wenn wir Glück haben, ...`,
		`Wenn wir Pech haben, ...`,
		`Ledorn öffnete eines der an ihr vorhandenen Fächer.`,
		`Auf der einen Seite endlose Dünen`,
		`In seinem Maul hielt er einen blutigen Fleischklumpen.`,
		`Gleichzeitig dachte er intensiv an Nebelschwaden, aus denen Wolken ja bestanden.`,
		`Warum stellte der bloß immer wieder dieselben Fragen?`,
		`Bei der Hinreise.`,
		`Schließlich tauchten in einem Waldstück unter ihnen Schienen auf.`,
		`Das Wahlrecht, das Frauen damals zugesprochen bekamen.`,
		`Es war Karl, dessen Leiche Donnerstag gefunden wurde.`,
		`Erst recht ich Arbeiter.`,
		`Erst recht wir Arbeiter.`,
		`Erst recht wir fleißigen Arbeiter.`,
		`Dann lud er Freunde ein.`,
		`Dann lud sie Freunde ein.`,
		`Aller Kommunikation liegt dies zugrunde.`,
		`Pragmatisch wählt man solche Formeln als Axiome.`,
		`Der eine Polizist rief dem anderen zu...`,
		`Der eine große Polizist rief dem anderen zu...`,
		`Das eine Kind rief dem anderen zu...`,
		`Er wollte seine Interessen wahrnehmen.`,
		`Es birgt für mich ein zu hohes juristisches Risiko.`,
		`... wo Krieg den Unschuldigen Leid und Tod bringt.`,
		`Der Abschuss eines Papageien.`,
		`Die Beibehaltung des Art. 1 ist geplant.`,
		`Die Verschiebung des bisherigen Art. 1 ist geplant.`,
		`In diesem Fall hatte das Vorteile.`,
		`So hat das Konsequenzen.`,
		`Ein für viele wichtiges Anliegen.`,
		`Das weckte bei vielen ungute Erinnerungen.`,
		`Etwas, das einem Angst macht.`,
		`Einem geschenkten Gaul schaut man nicht ins Maul.`,
		`Das erfordert Können.`,
		`Ist das Kunst?`,
		`Ist das Kunst oder Abfall?`,
		`Die Zeitdauer, während der Wissen nützlich bleibt, wird kürzer.`,
		`Es sollte nicht viele solcher Bilder geben`,
		`In den 80er Jahren.`,
		`Hast du etwas das Carina machen kann?`,
		`Ein Artikel in den Ruhr Nachrichten.`,
		`Ich wollte nur allen Hallo sagen.`,
		`Ich habe deshalb allen Freund*innen Bescheid gegeben.`,
		`Ich habe deshalb allen Freund_innen Bescheid gegeben.`,
		`Ich habe deshalb allen Freund:innen Bescheid gegeben.`,
		`Das betrifft auch eure Werkstudent:innen-Zielgruppe.`,
		`Das betrifft auch eure Werkstudent:innenzielgruppe.`,
		`Das betrifft auch eure Jurist:innenausbildung.`,
		`Sein*e Mitarbeiter*in ist davon auch betroffen.`,
		`Jede*r Mitarbeiter*in ist davon betroffen.`,
		`Alle Professor*innen`,
		`Gleichzeitig wünscht sich Ihr frostresistenter Mitbewohner einige Grad weniger im eigenen Zimmer?`,
		`Ein Trainer, der zum einen Fußballspiele sehr gut lesen und analysieren kann`,
		`Eine Massengrenze, bis zu der Lithium nachgewiesen werden kann.`,
		`Bei uns im Krankenhaus betrifft das Operationssäle.`,
		`Macht dir das Freude?`,
		`Das macht jedem Angst.`,
		`Dann macht das Sinn.`,
		`Das sind beides Lichtschalter.`,
		`Spielst du vielleicht auf das Bordell neben unserm Hotel an?`,
		`Spielst du vielleicht auf das Bordell neben unsrem Hotel an?`,
		`Dieses ungeahnt prophetische Wort`,
		`Das bestätigte Regierungssprecher Steffen Hebestreit am Freitag`,
		`Es kann gut sein, dass bei sowas Probleme erkannt werden.`,
		`Das Recht, das Frauen eingeräumt wird.`,
		`Der Mann, in dem quadratische Fische schwammen.`,
		`Der Mann, durch den quadratische Fische schwammen.`,
		`Gutenberg, der quadratische Mann.`,
		`Die größte Stuttgarter Grünanlage ist der Friedhof.`,
		`Die meisten Lebensmittel enthalten das.`,
		`Die wärmsten Monate sind August und September, die kältesten Januar und Februar.`,
		`Das Münchener Fest.`,
		`Das Münchner Fest.`,
		`Die Planung des Münchener Festes.`,
		`Das Berliner Wetter.`,
		`Den Berliner Arbeitern ist das egal.`,
		`Das Haus des Berliner Arbeiters.`,
		`Es gehört dem Berliner Arbeiter.`,
		`Das Stuttgarter Auto.`,
		`Das Bielefelder Radio.`,
		`Das Gütersloher Radio.`,
		`Das wirklich Wichtige kommt jetzt erst.`,
		`Besonders wenn wir Wermut oder Absinth trinken.`,
		`Ich wünsche dir alles Gute.`,
		`Es ist nicht bekannt, mit welchem Alter Kinder diese Fähigkeit erlernen.`,
		`Dieser ist nun in den Ortungsbereich des einen Roboters gefahren.`,
		`Wenn dies großen Erfolg hat, werden wir es weiter fördern.`,
		`Alles Gute!`,
		`Das bedeutet nichts Gutes.`,
		`Die Ereignisse dieses einen Jahres waren sehr schlimm.`,
		`Er musste einen Hochwasser führenden Fluss nach dem anderen überqueren.`,
		`Darf ich Ihren Füller für ein paar Minuten ausleihen?`,
		`Bringen Sie diesen Gepäckaufkleber an Ihrem Gepäck an.`,
		`Extras, die den Wert Ihres Autos erhöhen.`,
		`Er hat einen 34-jährigen Sohn.`,
		`Die Polizei erwischte die Diebin, weil diese Ausweis und Visitenkarte hinterließ.`,
		`Dieses Versäumnis soll vertuscht worden sein - es wurde Anzeige erstattet.`,
		`Die Firmen - nicht nur die ausländischen, auch die katalanischen - treibt diese Frage um.`,
		`Stell dich dem Leben lächelnd!`,
		`Die Messe wird auf das vor der Stadt liegende Ausstellungsgelände verlegt.`,
		`Sie sind ein den Frieden liebendes Volk.`,
		`Zum Teil sind das Krebsvorstufen.`,
		`Er sagt, dass das Rache bedeutet.`,
		`Wenn das Kühe sind, bin ich ein Elefant.`,
		`Karl sagte, dass sie niemandem Bescheid gegeben habe.`,
		`Es blieb nur dieser eine Satz.`,
		`Oder ist das Mathematikern vorbehalten?`,
		`Wenn hier einer Fragen stellt, dann ich.`,
		`Wenn einer Katzen mag, dann meine Schwester.`,
		`Ergibt das Sinn?`,
		`Sie ist über die Maßen schön.`,
		`Ich vertraue ganz auf die Meinen.`,
		`Was nützt einem Gesundheit, wenn man sonst ein Idiot ist?`,
		`Auch das hatte sein Gutes.`,
		`Auch wenn es sein Gutes hatte, war es doch traurig.`,
		`Er wollte doch nur jemandem Gutes tun.`,
		`und das erst Jahrhunderte spätere Auftauchen der Legende`,
		`Texas und New Mexico, beides spanische Kolonien, sind...`,
		`Texas und New Mexico - beides spanische Kolonien - sind...`,
		`Texas und New Mexico – beides spanische Kolonien – sind...`,
		`Weitere Brunnen sind insbesondere der Wittelsbacher und der Vater-Rhein-Brunnen auf der Museumsinsel, beides Werke von Adolf von Hildebrand.`,
		`Für manche ist das Anlass genug, darüber nicht weiter zu diskutieren.`,
		`Vielleicht schreckt das Frauen ab.`,
		`Unser Hund vergräbt seine Knochen im Garten.`,
		`Ob das Mehrwert bringt?`,
		`Warum das Sinn macht?`,
		`Das hängt davon ab, ob die Deutsch sprechen`,
		`Die meisten Coaches wissen nichts.`,
		`Die Präsent AG.`,
		`In New York war er der Titelheld in Richard III. und spielte den Mark Anton in Julius Cäsar.`,
		`Vielen Dank fürs Bescheid geben.`,
		`Welche Display Ads?`,
		`Das letzte Mal war das Anfang der 90er Jahre des vergangenen Jahrhunderts`,
		`Der vom Rat der Justizminister gefasste Beschluss zur Aufnahme von Vertriebenen...`,
		`Der letzte Woche vom Rat der Justizminister gefasste Beschluss zur Aufnahme von Vertriebenen...`,
		`Was war sie nur für eine dumme Person!`,
		`Was war ich für ein Idiot!`,
		`Was für ein Idiot!`,
		`Was für eine blöde Kuh!`,
		`Was ist sie nur für eine blöde Kuh!`,
		`Wie viele Paar Stiefel brauche ich eigentlich?`,
		`Dieses versuchten Mathematiker 400 Jahre lang vergeblich zu beweisen.`,
		`Gemälde informieren uns über das Leben von den vergangenen Jahrhunderten…`,
		`Die Partei, die bei den vorangegangenen Wahlen noch seine Politik unterstützt hatte.`,
		`Bei Zunahme der aufgelösten Mineralstoffe, bei denen...`,
		`Je mehr Muskelspindeln in einem Muskel vorhanden sind, desto feiner können die mit diesem verbundenen Bewegungen abgestimmt werden.`,
		`Diese datentechnischen Operationen werden durch Computerprogramme ausgelöst, d. h. über entsprechende, in diesen enthaltene Befehle (als Teil eines implementierten Algorithmus') vorgegeben.`,
		`Aus diesen resultierten Konflikte wie der Bauernkrieg und der Pfälzische Erbfolgekrieg.`,
		`Die Staatshandlungen einer Mikronation und von dieser herausgegebene Ausweise, Urkunden und Dokumente gelten im Rechtsverkehr als unwirksam`,
		`Auf der Hohen See und auf den mit dieser verbundenen Gewässern gelten die internationalen Kollisionsverhütungsregeln.`,
		`Art. 11 Abs. 2 GGV setzt dem bestimmte Arten der außergemeinschaftlichen Zugänglichmachung gleich`,
		`Grundsätzlich sind die Heilungschancen von Männern mit Brustkrebs nicht schlechter als die betroffener Frauen.`,
		`In diesem Viertel bin ich aufgewachsen.`,
		`Im November wurde auf dem Gelände der Wettbewerb ausgetragen.`,
		`Er ist Eigentümer des gleichnamigen Schemas und stellt dieses interessierten Domänen zur Verfügung.`,
		`Dort finden sie viele Informationen rund um die Themen Schwangerschaft, Geburt, Stillen, Babys und Kinder.`,
		`Die Galerie zu den Bildern findet sich hier.`,
		`Ganz im Gegensatz zu den Blättern des Brombeerstrauches.`,
		`Er erzählte von den Leuten und den Dingen, die er auf seiner Reise gesehen hatte.`,
		`Diese Partnerschaft wurde 1989 nach dem Massaker auf dem Platz des Himmlischen Friedens eingefroren.`,
		`Die Feuergefahr hingegen war für für die Londoner Teil des Alltags.`,
		`Was ist, wenn ein Projekt bei den Berliner Type Awards mit einem Diplom ausgezeichnet wird?`,
		`Was ist mit dem Liechtensteiner Kulturleben los?`,
		`Das ist der Mann den Präsident Xi Jinping verurteilte.`,
		`Wie viele Kolleg/-innen haben sie?`,
		`Die Ideen der neuen Kolleg/-innen sind gut!`,
		`Das erlaubt Forschern, neue Versuche durchzuführen.`,
		`Dies ermöglicht Forschern, neue Versuche durchzuführen.`,
		`Je länger zugewartet wird, desto schwieriger dürfte es werden, die Jungtiere von den Elterntieren zu unterscheiden.`,
		`Er schrieb ein von 1237 bis 1358 reichendes Geschichtswerk, dessen Schwerpunkt auf den Ereignissen in der Lombardei liegt.`,
		`Private Veranstaltungen waren, darauf hat die Strandhaus Norderstedt GmbH im Rahmen rechtlicher Klärungen selbst bestanden, nicht Bestandteil dieser Verträge.`,
		`Die Klientel der Partei.`,
		`Wenn die Gott zugeschriebenen Eigenschaften stimmen, dann...`,
		`Dieses Grünkern genannte Getreide ist aber nicht backbar.`,
		`Außerdem unterstützt mich Herr Müller beim abheften`,
		`Außerdem unterstützt mich Frau Müller beim abheften`,
		`Ich gebe dir ein kleines Kaninchen.`,
		`Ich gebe dir das kleine Kaninchen.`,
		`Die Top 3 der Umfrage`,
		`Dein Vorschlag befindet sich unter meinen Top 5.`,
		`Unter diesen rief das großen Unmut hervor.`,
		`Bei mir löste das Panik aus.`,
		`Sie können das machen in dem sie die CAD.pdf öffnen.`,
		`Ich mache eine Ausbildung zur Junior Digital Marketing Managerin.`,
		`Dann wird das Konsequenzen haben.`,
		`Dann hat das Konsequenzen.`,
		`Sollte das Konsequenzen nach sich ziehen?`,
		`Der Echo Show von Amazon`,
		`Die BVG kommen immer zu spät.`,
		`In der Frühe der Nacht.`,
		`Der TV Steinfurt.`,
		`Ein ID 3 von Volkswagen.`,
		`Der ID.3 von Volkswagen.`,
		`Der ID3 von Volkswagen.`,
		`Das bedeutet Krieg!`,
		`Im Tun zu sein verhindert Prokrastination.`,
		`Das ist doch lächerlich, was ist denn das für eine Klinik?`,
		`Was ist denn das für ein Typ?`,
		`Hier geht's zur Customer Journey.`,
		`Das führt zur Verbesserung der gesamten Customer Journey.`,
		`Meint er das wirklich Ernst?`,
		`Meinen Sie das Ernst?`,
		`Die können sich in unserer Hall of Fame verewigen.`,
		`Die können sich in unserer neuen Hall of Fame verewigen.`,
		`Auch, wenn das weite Teile der Bevölkerung betrifft.`,
		`Hat das Einfluss auf Ihr Trinkverhalten?`,
		`Ihr wisst aber schon, dass das Blödsinn ist.`,
		`Aber mein Wissen über die Antike ist ausbaufähig.`,
		`Sie werden merken, dass das echte Nutzer sind.`,
		`Dieses neue Mac OS trug den Codenamen Rhapsody.`,
		`Das Mac OS is besser als Windows.`,
		`Damit steht das Porsche Museum wie kaum ein anderes Museum für Lebendigkeit und Abwechslung.`,
		`Weitere Krankenhäuser sind dass Eastern Shore Memorial Hospital, IWK Health Centre, Nova Scotia Hospital und das Queen Elizabeth II Health Sciences Centre.`,
		`Ich bin von Natur aus ein sehr neugieriger Mensch.`,
		`Ich bin auf der Suche nach einer Junior Developerin.`,
		`War das Eifersucht?`,
		`Waren das schwierige Entscheidungen?`,
		`Soll das Demokratie sein?`,
		`Hat das Spaß gemacht?`,
		`Soll das Sinn stiften?`,
		`Soll das Freude machen?`,
		`Die Trial ist ausgelaufen.`,
		`Ein geworbener Neukunde interagiert zusätzlich mit dem Unternehmen.`,
		`Ich weiß, dass jeder LanguageTool benutzen sollte.`,
		`1992 übernahm die damalige Ernst Klett Schulbuchverlag GmbH, Stuttgart, den reprivatisierten Verlag Haack Gotha`,
		`Überlegst du dir einen ID.3 zu leasen?`,
		`Der Deutsch Langhaar ist ein mittelgroßer Jagdhund`,
		`Eine Lösung die Spaß macht`,
		`Mir machte das Spaß.`,
		`Wir möchten nicht, dass irgendjemand Fragen stellt.`,
		`Die Multiple Sklerose hat 1000 Gesichter.`,
		`Na ja, einige nennen das Freundschaft plus, aber das machen wir besser nicht.`,
		`Vogue, eigentlich als B-Seite der letzten Like A Prayer-Auskopplung Keep It Together gedacht, wurde kurzfristig als eigenständige Single herausgebracht`,
		`..., die laufend Gewaltsituationen ausgeliefert sind`,
		`Dann folgte die Festnahme der dringend Tatverdächtigen.`,
		`Von der ersten Spielminute an machten die Münsteraner Druck und ...`,
		`Wenn diese Prognose bestätigt wird, wird empfohlen, dass Unternehmen die gefährliche Güter benötigen, die Transporte am Montag und Dienstag machen.`,
		`Ich habe meine Projektidee (die riesiges finanzielles Potenzial hat) an einen Unternehmenspräsidenten geschickt.`,
		`Als weitere Rechtsquelle gelten gelegentlich noch immer der Londoner Court of Appeal und das britische House of Lords.`,
		`Die Evangelische Kirche befindet sich in der Bad Sodener Altstadt direkt neben dem Quellenpark.`,
		`Der volle Windows 10 Treibersupport`,
		`Zugleich stärkt es die renommierte Berliner Biodiversitätsforschung.`,
		`Der Windows 10 Treibersupport`,
		`Kennt irgendwer Tipps wie Kopfhörer länger halten?`,
		`George Lucas 1999 über seine sechsteilige Star Wars Saga.`,
		`… und von denen mehrere Gegenstand staatsanwaltlicher Ermittlungen waren.`,
		`Natürlich ist das Quatsch!`,
		`Die Xi Jinping Ära ist …`,
		`Die letzte unter Windows 98 lauffähige Version ist 5.1.`,
		`Das veranlasste Bürgermeister Adam, selbst tätig zu werden, denn er wollte es nicht zulassen, dass in seiner Stadt Notleidende ohne Hilfe dastehen.`,
		`Die südlichste Düsseldorfer Rheinbrücke ist die Fleher Brücke, eine Schrägseilbrücke mit dem höchsten Brückenpylon in Deutschland und einer Vielzahl von fächerförmig angeordneten Seilen.`,
		`Ein zeitweise wahres Stakkato an einschlägigen Patenten, das Benz & Cie.`,
		`Wem Rugby nicht sehr geläufig ist, dem wird auch das Six Nations nicht viel sagen.`,
		`Eine Boeing 767 der Air China stürzt beim Landeanflug in ein Waldgebiet.`,
		`Wir sind immer offen für Mitarbeiter die Teil eines der traditionellsten Malerbetriebe auf dem Platz Zürich werden möchten.`,
		`Gelingt das mit Erregern rechtzeitig, könnte das Infektionen sogar oft verhindern.`,
		`In der aktuellen Niedrigzinsphase bedeutet das sehr geringe Zinsen, die aber deutlich ansteigen können.`,
		`Es gibt viele Stock Screener.`,
		`Wir gehen zur Learning Academy.`,
		`Es ist ein stiller Bank Run.`,
		`Whirlpool Badewanne der europäische Marke SPAtec Modell Infinity.`,
		`1944 eroberte diese weite Teile von Südosteuropa.`,
		`Auch die Monopolstellung des staatlichen All India Radio, das in 24 Sprachen sendet`,
		`Das schwedischen Entwicklerstudio MachineGames hat uns vor drei Jahren mit Wolfenstein: The New Order positiv überrascht.`,
		`In einem normalen Joint habe es etwa ein halbes Gramm Hanf.`,
		`den leidenschaftlichen Lobpreis der texanischen Gateway Church aus`,
		`die gegnerischen Shooting Guards`,
		`Bald läppert sich das zu richtigem Geld zusammen.`,
		`Die Weimarer Parks laden ja förmlich ein zu Fotos im öffentlichen Raum.`,
		`Es is schwierig für mich, diese zu Sätzen zu verbinden.`,
		`Es kam zum einen zu technischen Problemen, zum anderen wurde es unübersichtlich.`,
		`Das Spiel wird durch den zu neuer Größe gewachsenen Torwart dominiert.`,
		`Dort findet sich schlicht und einfach alles & das zu sagenhafter Hafenkulisse.`,
		`Man darf gespannt sein, wen Müller für diese Aufgabe gewinnt.`,
		`Das Vereinslokal in welchem Zusammenkünfte stattfinden.`,
		`Er lässt niemanden zu Wort kommen.`,
		`Es war eine alles in allem spannende Geschichte.`,
		`Eine mehrere hundert Meter lange Startbahn.`,
		`Wir müssen jetzt um ein vielfaches höhere Preise zahlen.`,
		`Und eine von manchem geforderte Übergewinnsteuer.`,
		`Sie hat niemandem wirkliches Leid zugefügt.`,
		`Die Organe eines gerade Verstorbenen`,
		`Da wusste keiner Bescheid bezüglich dieser Sache.`,
		`Es braucht keiner Bescheid wissen.`,
		`Das sind auch beides staatliche Organe.`,
		`Ein Haus für die weniger Glücklichen.`,
		`Wir können sowas Mittwoch machen.`,
		`Den schlechter Verdienenden geht es schlecht.`,
		`Mit der weit weniger bekannten Horrorkomödie begann ihre Karriere.`,
		`Die Adelmanns wohnen in Herford.`,
		`Die Idee des Werbekaufmanns kam gut an.`,
		`Solch harte Worte!`,
		`Ich habe es an unseren amerikanischen Commercial Lawyer geschickt.`,
		`Dieser eine Schritt hat gedauert.`,
		`Es besteht durchaus die Gefahr, dass die Telekom eine solch starke monopolistische Stellung auf dem Markt hat, dass sich kaum Wettbewerb entfalten kann.`,
		`Wenn ein Tiger einen Menschen tötet, ist das Grausamkeit.`,
		`Kombinieren Sie diese zu ganzen Bewegungsprogrammen`,
		`Erst später wurde Kritik hauptsächlich an den Plänen zu einem Patriot Act II laut.`,
		`Laut Charlie XCX selbst sind das Personen, die vielleicht eine ...`,
		`Solch frivolen Gedanken wollen wir gar nicht erst nachgehen.`,
		`Er erwartete solch aggressives Verhalten.`,
		`Eine solch schöne Frau.`,
		`Einer solch schönen Frau.`,
		`Ein solch schöner Tisch.`,
		`Ein solch schöner neuer Tisch.`,
		`Eine solch begnadete Fotografin mit dabei zu haben und Tipps für die Fotosession zu bekommen, wäre schon toll.`,
		`Wie können wir als globale Gemeinschaft solch brennende Themen wie Klimawandel und Rezession in entwickelten Märkten in Angriff nehmen?`,
		`Wir waren überrascht, von ihm solch beißende Bemerkungen über seinen besten Freund zu hören.`,
		`Worten, wohl gewählt, wohnt solch große Macht inne.`,
		`Warum gelingt es den Stuten die Hengste in solch großen Rennen zu schlagen.`,
		`Eine solch gute Beratung kann natürlich nicht kostenlos sein, daher lassen Sie mich bitte wissen welche Kosten für die PKV und BU-Beratung auf mich zukommen werden.`,
		`Ein solch großer Ausbruch außerhalb des Nahen Ostens ist eine neue Entwicklung, heißt es weiter.`,
		`Natürlich dürfen unsere Sporteinheiten nicht fehlen, vor allem nicht bei solch gutem Essen.`,
		`Wie konnten Sie solch harte Songs gegen Ihre Familie schreiben?`,
		`Zur Krönung auf dem schönsten Aussichtsberg in der Ferienregion Tirol West gelegen, bietet die Venet Gipfelhütte alle Annehmlichkeiten, die man sich auf solch hohem Niveau (2.212 m) wünschen kann.`,
		`Vor allem nicht einen, der sich ein solch hohen Ballbesitz organisierte und die Berliner nicht zur Entfaltung kommen ließ.`,
		`Du solltest im Winter keinen solch hohen Berg besteigen.`,
		`Man darf an dieser Stelle fragen, wie ein solch hoher Verlust zustande kommt, wo doch der Stuttgarter Weg aus Sparen bestand.`,
		`Kann der Patient eine solch lange Operation überstehen?`,
		`Er kennt sich aus mit solch monumentalen Projekten.`,
		`Eine solch schöne hübsche Frau.`,
		`Bisher hat Gül einen solch offenen Affront gegen die ErdoganRegierung vermieden.`,
		`Umso dankbarer bin ich für Brüder, die klare Kante in theologischer Hinsicht zeigen und nachvollziehbar die Bibel auch in solch schwierigen unpopulären Themen auslegen.`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
	// bad cases (exact suggestions when Java lists them)
	arAssertBad(t, lt, rule, `Denn die einzelnen sehen sich einer sehr verschieden starken Macht des...`)
	arAssertBad(t, lt, rule, `Es birgt für mich ein zu hohes juristische Risiko.`)
	arAssertBad(t, lt, rule, `Das betrifft auch eure Werkstudent:innen-Xihfrisfgds.`)
	arAssertBad(t, lt, rule, `Das betrifft auch eure Jurist:innenxyzdfsdf.`)
	arAssertBad(t, lt, rule, `Gutenberg, die Genie.`)
	arAssertBad(t, lt, rule, `Wahrlich ein äußerst kritische Jury.`, `eine äußerst kritische Jury`)
	arAssertBad(t, lt, rule, `Das ist ein enorm großer Auto.`, `ein enorm großes Auto`)
	arAssertBad(t, lt, rule, `Ein Buch mit einem ganz ähnlichem Titel.`)
	arAssertBad(t, lt, rule, `Meiner Chef raucht.`)
	arAssertBad(t, lt, rule, `Er hat eine 34-jährigen Sohn.`)
	arAssertBad(t, lt, rule, `Es sind die Tisch.`, `die Tische`, `der Tisch`, `den Tisch`, `dem Tisch`)
	arAssertBad(t, lt, rule, `Es sind das Tisch.`, `der Tisch`, `den Tisch`, `dem Tisch`)
	arAssertBad(t, lt, rule, `Es sind die Haus.`, `das Haus`, `dem Haus`, `die Häuser`)
	arAssertBad(t, lt, rule, `Es sind der Haus.`, `dem Haus`, `das Haus`, `der Häuser`)
	arAssertBad(t, lt, rule, `Es sind das Frau.`, `die Frau`, `der Frau`)
	arAssertBad(t, lt, rule, `Das Auto des Mann.`, `der Mann`, `den Mann`, `dem Mann`, `des Manns`, `des Mannes`)
	arAssertBad(t, lt, rule, `Das interessiert das Mann.`, `der Mann`, `den Mann`, `dem Mann`)
	arAssertBad(t, lt, rule, `Das interessiert die Mann.`, `der Mann`, `den Mann`, `dem Mann`, `die Männer`)
	arAssertBad(t, lt, rule, `Das Auto ein Mannes.`, `ein Mann`, `eines Mannes`)
	arAssertBad(t, lt, rule, `Das Auto einem Mannes.`, `eines Mannes`, `einem Mann`)
	arAssertBad(t, lt, rule, `Das Auto einer Mannes.`, `eines Mannes`)
	arAssertBad(t, lt, rule, `Das Auto einen Mannes.`, `eines Mannes`, `einen Mann`)
	arAssertBad(t, lt, rule, `Die Galerie zu den Bilder findet sich hier.`)
	arAssertBad(t, lt, rule, `Ganz im Gegensatz zu den Blätter des Brombeerstrauches.`)
	arAssertBad(t, lt, rule, `Die erwähnt Konferenz ist am Samstag.`)
	arAssertBad(t, lt, rule, `Die erwähntes Konferenz ist am Samstag.`)
	arAssertBad(t, lt, rule, `Die erwähnten Konferenz ist am Samstag.`)
	arAssertBad(t, lt, rule, `Die erwähnter Konferenz ist am Samstag.`)
	arAssertBad(t, lt, rule, `Die erwähntem Konferenz ist am Samstag.`)
	arAssertBad(t, lt, rule, `Die gemessen Werte werden in die länderspezifische Höhe über dem Meeresspiegel umgerechnet.`)
	arAssertBad(t, lt, rule, `Darüber hinaus haben wir das berechtigte Interessen, diese Daten zu verarbeiten.`)
	arAssertBad(t, lt, rule, `Eine Amnestie kann den Hingerichteten nicht das Leben und dem heimgesuchten Familien nicht das Glück zurückgeben.`)
	arAssertBad(t, lt, rule, `Z. B. therapeutisches Klonen, um aus den gewonnen Zellen in vitro Ersatzorgane für den Patienten zu erzeugen`)
	arAssertBad(t, lt, rule, `Die Partei, die bei den vorangegangen Wahlen noch seine Politik unterstützt hatte.`)
	arAssertBad(t, lt, rule, `Bei Zunahme der aufgelösten Mineralstoffen, bei denen...`)
	arAssertBad(t, lt, rule, `Durch die große Vielfalt der verschiedene Linien ist für jeden Anspruch die richtige Brille im Portfolio.`)
	arAssertBad(t, lt, rule, `In diesen Viertel bin ich aufgewachsen.`)
	arAssertBad(t, lt, rule, `Im November wurde auf den Gelände der Wettbewerb ausgetragen.`)
	arAssertBad(t, lt, rule, `Dort finden sie Testberichte und viele Informationen rund um das Themen Schwangerschaft, Geburt, Stillen, Babys und Kinder.`)
	arAssertBad(t, lt, rule, `Je länger zugewartet wird, desto schwieriger dürfte es werden, die Jungtiere von den Elterntiere zu unterscheiden.`)
	arAssertBad(t, lt, rule, `Er schrieb ein von 1237 bis 1358 reichendes Geschichtswerk, dessen Schwerpunkt auf den Ereignisse in der Lombardei liegt.`)
	arAssertBad(t, lt, rule, `Des großer Mannes.`)
	arAssertBad(t, lt, rule, `Er erzählte von den Leute und den Dingen, die er gesehen hatte.`)
	arAssertBad(t, lt, rule, `Diese Partnerschaft wurde 1989 nach den Massaker auf dem Platz des Himmlischen Friedens eingefroren.`)
	arAssertBad(t, lt, rule, `Das Dach von meine Auto.`, `meinem Auto`)
	arAssertBad(t, lt, rule, `Das Dach von meinen Auto.`, `meinem Auto`, `meinen Autos`)
	arAssertBad(t, lt, rule, `Das Dach mein Autos.`, `mein Auto`, `meine Autos`, `meines Autos`, `meinen Autos`, `meiner Autos`)
	arAssertBad(t, lt, rule, `Das Dach meinem Autos.`, `meinem Auto`, `meines Autos`, `meine Autos`, `meinen Autos`, `meiner Autos`)
	arAssertBad(t, lt, rule, `Das Dach meinem großen Autos.`)
	arAssertBad(t, lt, rule, `Das Dach mein großen Autos.`)
	arAssertBad(t, lt, rule, `Das Klientel der Partei.`, `Die Klientel`, `Der Klientel`)
	arAssertBad(t, lt, rule, `Der Haus ist groß`, `Dem Haus`, `Das Haus`, `Der Häuser`)
	arAssertBad(t, lt, rule, `Aber der Haus ist groß`, `dem Haus`, `das Haus`, `der Häuser`)
	arAssertBad(t, lt, rule, `Ich habe einen Feder gefunden.`, `eine Feder`, `einer Feder`)
	arAssertBad(t, lt, rule, `Der Zustand meiner Gehirns.`)
	arAssertBad(t, lt, rule, `Lebensmittel sind da, um den menschliche Körper zu ernähren.`)
	arAssertBad(t, lt, rule, `Geld ist da, um den menschliche Überleben sicherzustellen.`)
	arAssertBad(t, lt, rule, `Sie hatte das kleinen Kaninchen.`)
	arAssertBad(t, lt, rule, `Frau Müller hat das wichtigen Dokument gefunden.`)
	arAssertBad(t, lt, rule, `Ich gebe dir ein kleine Kaninchen.`)
	arAssertBad(t, lt, rule, `Ich gebe dir ein kleinen Kaninchen.`)
	arAssertBad(t, lt, rule, `Ich gebe dir ein kleinem Kaninchen.`)
	arAssertBad(t, lt, rule, `Ich gebe dir ein kleiner Kaninchen.`)
	arAssertBad(t, lt, rule, `Ich gebe dir das kleinen Kaninchen.`)
	arAssertBad(t, lt, rule, `Ich gebe dir das kleinem Kaninchen.`)
	arAssertBad(t, lt, rule, `Ich gebe dir das kleiner Kaninchen.`)
	arAssertBad(t, lt, rule, `Geprägt ist der Platz durch einen 142 Meter hoher Obelisken`, `einen 142 Meter hohen Obelisken`)
	arAssertBad(t, lt, rule, `Es birgt für mich ein überraschend hohes juristische Risiko.`, `ein überraschend hohes juristisches Risiko`)
	arAssertBad(t, lt, rule, `Es birgt für mich ein zu hohes juristische Risiko.`, `ein zu hohes juristisches Risiko`)
	arAssertBad(t, lt, rule, `Hier steht Ihre Text.`)
	arAssertBad(t, lt, rule, `Hier steht ihre Text.`)
	arAssertBad(t, lt, rule, `Antje Last, Inhaberin des Berliner Kult Hotels Auberge, freute sich ebenfalls über die Gastronomenfamilie aus Bayern.`)
	arAssertBad(t, lt, rule, `Das ist doch lächerlich, was ist denn das für ein Klinik?`)
	arAssertBad(t, lt, rule, `Ich weiß nicht mehr, was unser langweiligen Thema war.`)
	arAssertBad(t, lt, rule, `Er ging ins Küche.`)
	arAssertBad(t, lt, rule, `Er ging ans Luft.`)
	arAssertBad(t, lt, rule, `Eine Niereninsuffizienz führt zur Störungen des Wasserhaushalts.`)
	arAssertBad(t, lt, rule, `Er stieg durchs Fensters.`)
	arAssertBad(t, lt, rule, `Ich habe heute ein Krankenwagen gesehen.`)
	arAssertBad(t, lt, rule, `Funktioniert das Software auch mit Windows?`)
	arAssertBad(t, lt, rule, `So soll er etwa Texte des linken Literaturwissenschaftler Helmut Lethen mit besonderem Interesse gelesen haben.`)
	arAssertBad(t, lt, rule, `Auf dieser Website werden allerdings keine solche Daten weiterverxxx.`)
	arAssertBad(t, lt, rule, `Bei größeren Gruppen und/oder mehrere Tagen gibts einen nennenswerten Nachlass.`)
	arAssertBad(t, lt, rule, `Mit der weit weniger bekannte Horrorkomödie begann ihre Karriere.`, `der weit weniger bekannten Horrorkomödie`)
	arAssertBad(t, lt, rule, `Die Idee des Werbekaufmann kam gut an.`)
	arAssertBad(t, lt, rule, `Ich habe keine Zeit für solche kleinlichen Belangen.`)
	arAssertBad(t, lt, rule, `Einen Dämonen wird er nicht aufhalten.`)
	arAssertBad(t, lt, rule, `Das versetzte den Kronprinz in Schrecken.`)
	arAssertBad(t, lt, rule, `Solche kleinen Anbietern nutzen dann eines der drei großen Mobilfunknetze Deutschlands.`)
	arAssertBad(t, lt, rule, `..., das heißt solche natürliche Personen, welche unsere Leistungen in Anspruch nehmen, ...`)
	arAssertBad(t, lt, rule, `Der Erwerb solcher kultureller Güter ist natürlich stark an das ökonomische Kapital gebunden.`)
	arAssertBad(t, lt, rule, `Wir haben das Abo beendet und des Betrag erstattet.`)
}

// Twin of AgreementRuleTest.testZurReplacement — real GermanTagger analysis.
func TestAgreementRule_Real_ZurReplacement(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Hier geht's zur Sonne.`,
		`Hier geht's zum Schrank.`,
		`Niereninsuffizienz führt zu Störungen des Wasserhaushalts.`,
		`Das hat der fließend Englisch sprechende Mitarbeiter veranlasst.`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
	// bad cases (exact suggestions when Java lists them)
	arAssertBad(t, lt, rule, `Hier geht's zur Schrank.`, `zum Schrank`)
	arAssertBad(t, lt, rule, `Hier geht's zur Schränken.`, `zum Schränken`, `zu Schränken`)
	arAssertBad(t, lt, rule, `Hier geht's zur Männern.`, `zu Männern`)
	arAssertBad(t, lt, rule, `Hier geht's zur Portal.`, `zum Portal`)
	arAssertBad(t, lt, rule, `Hier geht's zur Portalen.`, `zu Portalen`)
	arAssertBad(t, lt, rule, `Sie gehen zur Frauen.`, `zu Frauen`, `zur Frau`)
	arAssertBad(t, lt, rule, `Niereninsuffizienz führt zur Störungen des Wasserhaushalts.`, `zu Störungen`, `zur Störung`)
	arAssertBad(t, lt, rule, `Das Motiv wird in der Klassik auch zur Darstellungen übernommen.`, `zu Darstellungen`, `zur Darstellung`)
}

// Twin of AgreementRuleTest.testVieleWenige — real GermanTagger analysis.
func TestAgreementRule_Real_VieleWenige(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Zusammenschluss mehrerer dörflicher Siedlungen an einer Furt`,
		`Für einige markante Szenen`,
		`Für einige markante Szenen baute Hitchcock ein Schloss.`,
		`Haben Sie viele glückliche Erfahrungen in Ihrer Kindheit gemacht?`,
		`Es gibt viele gute Sachen auf der Welt.`,
		`Viele englische Wörter haben lateinischen Ursprung`,
		`Ein Bericht über Fruchtsaft, einige ähnliche Erzeugnisse und Fruchtnektar`,
		`Der Typ, der seit einiger Zeit immer wieder hierher kommt.`,
		`Jede Schnittmenge abzählbar vieler offener Mengen`,
		`Es kam zur Fusion der genannten und noch einiger weiterer Unternehmen.`,
		`Zu dieser Fragestellung gibt es viele unterschiedliche Meinungen.`,
		`Wir zeigen die Gründe auf, wieso noch nicht jeder solche Anschlüsse hat.`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
}

// Twin of AgreementRuleTest.testRegression
func TestAgreementRule_Real_Regression(t *testing.T) {
	lt, rule := arSetup(t)
	// Java: lt.check("Und so.\r\nDie Bier.") size 1 (full check, DE_AGREEMENT fires)
	str := "Und so.\r\nDie Bier."
	// rule.match per sentence; second sentence has the error
	sents := lt.Analyze(str)
	var n int
	for _, sent := range sents {
		n += len(rule.Match(sent))
	}
	require.Equal(t, 1, n, "regression Die Bier")
}

// Twin of AgreementRuleTest.testDetAdjNounRule — real GermanTagger analysis.
func TestAgreementRule_Real_DetAdjNounRule(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Die Übernahme der früher selbständigen Gesellschaft`,
		`Das ist, weil man oft bei anderen schreckliches Essen vorgesetzt bekommt.`,
		`Das ist der riesige Tisch.`,
		`Der riesige Tisch ist groß.`,
		`Die Kanten der der riesigen Tische.`,
		`Den riesigen Tisch mag er.`,
		`Es mag den riesigen Tisch.`,
		`Die Kante des riesigen Tisches.`,
		`Dem riesigen Tisch fehlt was.`,
		`Die riesigen Tische sind groß.`,
		`Der riesigen Tische wegen.`,
		`An der roten Ampel.`,
		`Dann hat das natürlich Nachteile.`,
		`Ihre erste Nr. 1`,
		`Wir bedanken uns bei allen Teams.`,
		`Als Heinrich versuchte, seinen Kandidaten für den Mailänder Bischofssitz durchzusetzen, reagierte der Papst sofort.`,
		`Den neuen Finanzierungsweg wollen sie daher Hand in Hand mit dem Leser gehen.`,
		`Lieber den Spatz in der Hand...`,
		`Wir wollen sein ein einzig Volk von Brüdern`,
		`Eine Zeitreise durch die 68er Revolte`,
		`Ich besitze ein Modell aus der 300er Reihe.`,
		`Aber ansonsten ist das erste Sahne`,
		`...damit diese ausreichend Sauerstoff geben.`,
		`...als auch die jedem zukommende Freiheit.`,
		`...als auch die daraus jedem zukommende Freiheit.`,
		`Damit zeigen wir, wie bedeutungsreich manche deutsche Begriffe sein können.`,
		`Damit zeigen wir, wie bedeutungsreich manche deutschen Begriffe sein können.`,
		`2009 gab es im Rathaus daher Bestrebungen ein leichter handhabbares Logo einzuführen.`,
		`Das ist eine leichter handhabbare Situation.`,
		`Es gibt viele verschiedene Stock Screener.`,
		`Die Ware umräumen, um einer anderen genügend Platz zu schaffen.`,
		`Ich widerrufe den mit Ihnen geschlossenen Vertrag.`,
		`Er klagte auch gegen den ohne ihn verkündeten Sachbeschluss.`,
		`Dieser relativ gesehen starke Mann.`,
		`Diese relativ gesehen starke Frau.`,
		`Dieses relativ gesehen starke Auto.`,
		`Es kann gut sein, dass bei sowas echte Probleme erkannt werden.`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
	// bad cases (exact suggestions when Java lists them)
	arAssertBad(t, lt, rule, `Er hatte ein anstrengenden Tag`, `ein anstrengender Tag`, `ein anstrengendes Tag`, `einen anstrengenden Tag`, `einem anstrengenden Tag`)
	arAssertBad(t, lt, rule, `Es sind die riesigen Tisch.`)
	arAssertBad(t, lt, rule, `Als die riesigen Tischs kamen.`)
	arAssertBad(t, lt, rule, `Als die riesigen Tisches kamen.`)
	arAssertBad(t, lt, rule, `Der riesigen Tisch und so.`)
	arAssertBad(t, lt, rule, `An der roter Ampel.`)
	arAssertBad(t, lt, rule, `An der rote Ampel.`)
	arAssertBad(t, lt, rule, `An der rotes Ampel.`)
	arAssertBad(t, lt, rule, `An der rotem Ampel.`)
	arAssertBad(t, lt, rule, `Er hatte ihn aus dem 1,4 Meter tiefem Wasser gezogen.`, `dem 1,4 Meter tiefen Wasser`)
	arAssertBad(t, lt, rule, `Das ist ein sehr schönes Tisch.`, `ein sehr schöner Tisch`)
	arAssertBad(t, lt, rule, `Er hatte eine sehr schweren Infektion.`)
	arAssertBad(t, lt, rule, `Ein fast 5 Meter hohem Haus.`)
	arAssertBad(t, lt, rule, `Ein fünf Meter hohem Haus.`)
	arAssertBad(t, lt, rule, `Es wurden Karavellen eingesetzt, da diese für die flachen Gewässern geeignet waren.`)
	arAssertBad(t, lt, rule, `Wir bedanken uns bei allem Teams.`)
	arAssertBad(t, lt, rule, `Dabei geht es um das altbekannte Frage der Dynamiken der Eigenbildung..`)
	arAssertBad(t, lt, rule, `Den neue Finanzierungsweg wollen sie daher Hand in Hand mit dem Leser gehen.`)
	arAssertBad(t, lt, rule, `Den neuen Finanzierungsweg wollen sie daher Hand in Hand mit dem Lesern gehen.`)
	arAssertBad(t, lt, rule, `Ich widerrufe den mit Ihnen geschlossene Vertrag.`, `der mit Ihnen geschlossene Vertrag`, `den mit Ihnen geschlossenen Vertrag`)
	arAssertBad(t, lt, rule, `Er klagte auch gegen den ohne ihn verkündete Sachbeschluss.`, `den ohne ihn verkündeten Sachbeschluss`)
}

// Twin of AgreementRuleTest.testDetAdjAdjNounRule — real GermanTagger analysis.
func TestAgreementRule_Real_DetAdjAdjNounRule(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Das verlangt reifliche Überlegung.`,
		`Das bedeutet private Versicherungssummen ab 100€.`,
		`Das erfordert einigen Mut.`,
		`Die abnehmend aufwendige Gestaltung der Portale...`,
		`Die strahlend roten Blumen.`,
		`Der weiter vorhandene Widerstand konnte sich nicht durchsetzen.`,
		`Das jetzige gemeinsame Ergebnis...`,
		`Das früher übliche Abdecken mit elementarem Schwefel...`,
		`Das einzig wirklich Schöne...`,
		`Andere weniger bekannte Vorschläge waren „Konsistenter Empirismus“ oder...`,
		`Werden mehrere solcher physikalischen Elemente zu einer Einheit zusammengesetzt...`,
		`Aufgrund ihrer weniger guten Bonitätslage.`,
		`Mit ihren teilweise eigenwilligen Außenformen...`,
		`Die deutsche Kommasetzung bedarf einiger technischer Ausarbeitung.`,
		`Die deutsche Kommasetzung bedarf einiger guter technischer Ausarbeitung.`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
	// bad cases (exact suggestions when Java lists them)
	arAssertBad(t, lt, rule, `Das ist eine solides strategisches Fundament`, `ein solides strategisches Fundament`)
	arAssertBad(t, lt, rule, `Das ist eine solide strategisches Fundament`, `ein solides strategisches Fundament`)
	arAssertBad1(t, lt, rule, `Das ist eine solide strategische Fundament`, `ein solides strategisches Fundament`)
	arAssertBad(t, lt, rule, `Das ist ein solide strategisches Fundament`, `ein solides strategisches Fundament`)
	arAssertBad(t, lt, rule, `Das ist ein solides strategische Fundament`, `ein solides strategisches Fundament`)
	arAssertBad(t, lt, rule, `Das ist ein solides strategisches Fundamente`, `ein solides strategisches Fundament`)
	arAssertBad(t, lt, rule, `Das ist ein solides strategisches Fundaments`, `ein solides strategisches Fundament`)
	arAssertBad(t, lt, rule, `Die deutsche Kommasetzung bedarf einiger technisches Ausarbeitung.`)
	arAssertBad(t, lt, rule, `Die deutsche Kommasetzung bedarf einiger guter technische Ausarbeitung.`)
}

// Twin of AgreementRuleTest.testKonUntArtDefSub — real GermanTagger analysis.
func TestAgreementRule_Real_KonUntArtDefSub(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Wieso verstehst du nicht, dass das komplett verschiedene Dinge sind?`,
		`Ich frage mich sehr, ob die wirklich zusätzliche Gebühren abdrücken wollen`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
	// bad cases (exact suggestions when Java lists them)
	arAssertBad(t, lt, rule, `Dies wurde durchgeführt um das moderne Charakter zu betonen.`, `den modernen Charakter`)
	arAssertBad(t, lt, rule, `Nur bei Topfpflanzung ist eine regelmäßige Düngung wichtig, da die normalen Bodenbildungsprozessen nicht stattfinden.`, `die normalen Bodenbildungsprozesse`, `den normalen Bodenbildungsprozessen`)
	arAssertBad(t, lt, rule, `Die Höhe kommt oft darauf an, ob die richtigen Leuten gut mit einen können oder nicht.`)
}

// Twin of AgreementRuleTest.testBugFixes — real GermanTagger analysis.
func TestAgreementRule_Real_BugFixes(t *testing.T) {
	lt, rule := arSetup(t)
	goods := []string{
		`Das passiert nur, wenn der zu Pflegende bereit ist.`,
		`Peter, iss nicht meine`,
	}
	for _, s := range goods {
		arAssertGood(t, lt, rule, s)
	}
	// bad cases (exact suggestions when Java lists them)
	arAssertBad(t, lt, rule, `Denn die einzelnen sehen sich einer sehr verschieden starken Macht des...`, `einer sehr verschiedenen starken Macht`)
}
