package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// prohibitedPair is a confusable compound fragment pair (all-lowercase).
type prohibitedPair struct {
	part1, desc1, part2, desc2 string
}

// Preferred bad→good fragments without n-gram LM (covers Java test compounds).
var prohibitedPreferred = map[string]string{
	"lehr": "leer",
	"uhrb": "urb",   // Uhrberliner
	"uhre": "ure",   // Uhreinwohner
	"mita": "mieta", // Mitauto
}

var lowercaseProhibitedPairs = []prohibitedPair{
	{"knoten", "Verschlingung von Fäden", "konten", "Plural von 'Konto'"},
	{"schaf", "Tier", "schaft", "'-schaft' (Element zur Wortbildung, z. B. 'Freundschaft')"},
	{"schafen", "Dativ Plural von 'Schaf'", "schaften", "'-schaften' (Element zur Wortbildung, z. B. 'Freundschaften')"},
	{"alpen", "Hochgebirge in Mittel- und Südeuropa", "alben", "Plural von 'Album'"},
	{"pillen", "Tabletten", "pullen", "Plural von 'Pulle' (Flasche)"},
	{"tauben", "Vogelart", "trauben", "Obstsorte"},
	{"panel", "ausgewählte Personengruppe", "paneel", "Platte für Wand- und Deckenverkleidungen"},
	{"nabe", "Mittelteil eines Rades", "narbe", "verheilende Wunde"},
	{"first", "höchste Kante an einem geneigten Dach", "frist", "spätester Zeitpunkt"},
	{"kisten", "Behälter", "kosten", "Ausgaben"},
	{"koma", "Zustand tiefer Bewusstlosigkeit", "komma", "Satzzeichen"},
	{"korn", "Getreide sowie dessen Frucht", "kron", "Vorsilbe z.B. in 'Kronkorken'"},
	{"bauten", "Form von 'Bau' (Bauwerk, Haus, ...)", "beuten", "Form von 'Beute'"},
	{"file", "engl. 'Datei'", "filet", "ein Stück Fleisch oder Fisch"},
	{"zecke", "blutsaugender Parasit", "zwecke", "Dativ von 'Zweck' (Ziel)"},
	{"frucht", "Teil einer Pflanze; Obst", "furcht", "Angst"},
	{"rate", "Verhältnis zwischen zwei Größen", "ratte", "Nagetier"},
	{"posten", "Arbeitsplatz, Wachposten", "posen", "Pose: betonte Körperhaltung"},
	{"himmel", "Bereich über der Erde", "hummel", "Insekt"},
	{"server", "Computer", "servier", "zu 'servieren'"},
	{"ziege", "Tier", "ziegel", "Ziegelstein"},
	{"robe", "Kleidungsstück", "probe", "Test, Kontrolle"},
	{"mode", "Kleidung", "monde", "Begleiter eines Planeten"},
	{"eigen", "'selbst', z.B. 'Eigenzitat'", "eingen", "Möglicher Tippfehler"},
	{"stümpfe", "Rest eines Körpergliedes", "strümpfe", "Bekleidungsstück für den Fuß"},
	{"gelände", "Gebiet", "geländer", "Konstruktion zum Festhalten entlang von Treppen"},
	{"tropen", "feuchtwarme Gebiete am Äquator", "tropfen", "kleine Menge Flüssigkeit"},
	{"enge", "Mangel an Platz", "menge", "Anzahl an Einheiten"},
	{"ritt", "Reiten", "tritt", "Aufsetzen eines Fußes"},
	{"beine", "Körperteil", "biene", "Insekt"},
	{"rebe", "Weinrebe", "reibe", "Küchenreibe"},
	{"ass", "Spielkarte", "pass", "Reisepass; Übergang durch ein Gebirge"},
	{"türmer", "Turmwächter", "türme", "Plural von 'Turm' (Bauwerk)"},
	{"soge", "ziehende Strömungen", "sorge", "bedrückendes Gefühl"},
	{"panne", "technischer Defekt", "spanne", "Zeitraum"},
	{"elfer", "Elfmeter", "helfer", "Person, die hilft"},
	{"bau", "Bauwerk, Baustelle", "baum", "Pflanze"},
	{"gase", "Plural von 'Gas' (Aggregatzustand)", "gasse", "kleine Straße"},
	{"ekel", "Abscheu", "enkel", "Kind eines eigenen Kindes"},
	{"reis", "Nahrungsmittel", "reise", "Ausflug/Fahrt"},
	{"speichel", "Körperflüssigkeit", "speicher", "Lager, Depot, Ablage"},
	{"hüte", "Kopfbedeckungen", "häute", "Plural von 'Haut'"},
	{"bach", "kleiner Fluss", "bauch", "Teil des menschlichen Körpers"},
	{"lage", "Position", "alge", "im Wasser lebende Organismen"},
	{"schenke", "Gastwirtschaft (auch: Schänke)", "schenkel", "Ober- und Unterschenkel"},
	{"rune", "Schriftzeichen der Germanen", "runde", "Rundstrecke"},
	{"mai", "Monat nach April", "mail", "E-Mail"},
	{"pump", "'auf Pump': umgangssprachlich für 'auf Kredit'", "pumpe", "Gerät zur Beförderung von Flüssigkeiten"},
	{"mitte", "zentral", "mittel", "Methode, um etwas zu erreichen"},
	{"fein", "feinkörnig, genau, gut", "feind", "Gegner"},
	{"traum", "Erleben während des Schlafes", "trauma", "Verletzung"},
	{"name", "Bezeichnung (z.B. 'Vorname')", "nahme", "zu 'nehmen' (z.B. 'Teilnahme')"},
	{"bart", "Haarbewuchs im Gesicht", "dart", "Wurfpfeil"},
	{"hart", "fest", "dart", "Wurfpfeil"},
	{"speiche", "Verbindung zwischen Nabe und Felge beim Rad", "speicher", "Lagerraum"},
	{"speichen", "Verbindung zwischen Nabe und Felge beim Rad", "speicher", "Lagerraum"},
	{"kart", "Gokart (Fahrzeug)", "karte", "Fahrkarte, Postkarte, Landkarte, ..."},
	{"karts", "Kart = Gokart (Fahrzeug)", "karte", "Fahrkarte, Postkarte, Landkarte, ..."},
	{"kurz", "Gegenteil von 'lang'", "kur", "medizinische Vorsorge und Rehabilitation"},
	{"kiefer", "knöcherner Teil des Schädels", "kiefern", "Kieferngewächse (Baum)"},
	{"gel", "dickflüssige Masse", "geld", "Zahlungsmittel"},
	{"flucht", "Entkommen, Fliehen", "frucht", "Ummantelung des Samens einer Pflanze"},
	{"kamp", "Flurname für ein Stück Land", "kampf", "Auseinandersetzung"},
	{"obst", "Frucht", "ost", "Himmelsrichtung"},
	{"beeren", "Früchte", "bären", "Raubtiere"},
	{"laus", "Insekt", "lauf", "Bewegungsart"},
	{"läuse", "Insekt", "läufe", "Bewegungsart"},
	{"läusen", "Insekt", "läufen", "Bewegungsart"},
	{"ruck", "plötzliche Bewegung", "druck", "Belastung"},
	{"brüste", "Plural von Brust", "bürste", "Gerät mit Borsten, z.B. zum Reinigen"},
	{"attraktion", "Sehenswürdigkeit", "akttaktion", "vermutlicher Tippfehler"},
	{"nah", "zu 'nah' (wenig entfernt)", "näh", "zu 'nähen' (mit einem Faden verbinden)"},
	{"turn", "zu 'turnen'", "turm", "hohes Bauwerk"},
	{"mit", "Präposition", "miet", "zu 'Miete' (Überlassung gegen Bezahlung)"},
	{"bart", "Behaarung im Gesicht", "brat", "zu 'braten', z.B. 'Bratkartoffel'"},
	{"uhr", "Instrument zur Zeitmessung", "ur", "ursprünglich"},
	{"abschluss", "Ende", "abschuss", "Vorgang des Abschießens, z.B. mit einer Waffe"},
	{"brache", "verlassenes Grundstück", "branche", "Wirtschaftszweig"},
	{"wieder", "erneut, wiederholt, nochmal (Wiederholung, Wiedervorlage, ...)", "wider", "gegen, entgegen (Widerwille, Widerstand, Widerspruch, ...)"},
	{"leer", "ohne Inhalt", "lehr", "bezogen auf Ausbildung und Wissen"},
	{"gewerbe", "wirtschaftliche Tätigkeit", "gewebe", "gewebter Stoff; Verbund ähnlicher Zellen"},
	{"schuh", "Fußbekleidung", "schul", "auf die Schule bezogen"},
	{"klima", "langfristige Wetterzustände", "lima", "Hauptstadt von Peru"},
	{"modell", "vereinfachtes Abbild der Wirklichkeit", "model", "Fotomodell"},
	{"treppen", "Folge von Stufen (Mehrzahl)", "truppen", "Armee oder Teil einer Armee (Mehrzahl)"},
	{"häufigkeit", "Anzahl von Ereignissen", "häutigkeit", "z.B. in Dunkelhäutigkeit"},
	{"hin", "in Richtung", "hirn", "Gehirn, Denkapparat"},
	{"verklärung", "Beschönigung, Darstellung in einem besseren Licht", "erklärung", "Darstellung, Erläuterung"},
	{"spitze", "spitzes Ende eines Gegenstandes", "spritze", "medizinisches Instrument zur Injektion"},
	{"punk", "Jugendkultur", "punkt", "Satzzeichen"},
	{"reis", "Nahrungsmittel", "eis", "gefrorenes Wasser"},
	{"balkan", "Region in Südosteuropa", "balkon", "Plattform, die aus einem Gebäude herausragt"},
	{"haft", "Freiheitsentzug", "schaft", "-schaft (Element zur Wortbildung)"},
	{"stande", "zu 'Stand'", "stange", "länglicher Gegenstand"},
}

// ProhibitedCompoundRule is a surface stand-in for
// org.languagetool.rules.de.ProhibitedCompoundRule without n-gram LM / speller.
type ProhibitedCompoundRule struct {
	Messages map[string]string
	Prefer   map[string]string
}

func NewProhibitedCompoundRule(messages map[string]string) *ProhibitedCompoundRule {
	return &ProhibitedCompoundRule{Messages: messages, Prefer: prohibitedPreferred}
}

func (r *ProhibitedCompoundRule) GetID() string { return "DE_PROHIBITED_COMPOUNDS" }

func removeHyphensAndAdaptCase(word string) string {
	if !strings.Contains(word, "-") {
		return ""
	}
	parts := strings.Split(word, "-")
	for _, p := range parts {
		if utf8.RuneCountInString(p) <= 1 {
			return ""
		}
	}
	var b strings.Builder
	for i, p := range parts {
		if i == 0 {
			b.WriteString(p)
			continue
		}
		rs := []rune(p)
		if len(rs) == 0 {
			continue
		}
		rs[0] = unicode.ToLower(rs[0])
		b.WriteString(string(rs))
	}
	return b.String()
}

func (r *ProhibitedCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	prev := ""
	seen := map[[2]int]bool{}
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		word := tok.GetToken()
		if prev == "Herr" || prev == "Frau" || prev == "Herrn" {
			prev = word
			continue
		}
		candidates := []string{word}
		if j := removeHyphensAndAdaptCase(word); j != "" {
			candidates = append(candidates, j)
		}
		for _, part := range strings.Split(word, "-") {
			candidates = append(candidates, part)
		}
		for _, part := range candidates {
			if m := r.matchPart(sentence, tok, part); m != nil {
				key := [2]int{m.GetFromPos(), m.GetToPos()}
				if !seen[key] {
					seen[key] = true
					matches = append(matches, m)
				}
				break
			}
		}
		prev = word
	}
	return matches
}

func (r *ProhibitedCompoundRule) matchPart(sentence *languagetool.AnalyzedSentence, tok *languagetool.AnalyzedTokenReadings, wordPart string) *rules.RuleMatch {
	if utf8.RuneCountInString(wordPart) <= 6 {
		return nil
	}
	prefer := r.Prefer
	if prefer == nil {
		prefer = prohibitedPreferred
	}
	lc := strings.ToLower(wordPart)
	// longest bad key first
	type kv struct{ bad, good string }
	var keys []kv
	for bad, good := range prefer {
		keys = append(keys, kv{bad, good})
	}
	// sort by len bad desc
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if len(keys[j].bad) > len(keys[i].bad) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	for _, kg := range keys {
		bad, good := kg.bad, kg.good
		if !strings.Contains(lc, bad) {
			continue
		}
		idx := strings.Index(lc, bad)
		// if this is already the good form overlapping, skip
		if idx >= 0 && idx+len(good) <= len(lc) && lc[idx:idx+len(good)] == good {
			continue
		}
		variant := wordPart[:idx] + good + wordPart[idx+len(bad):]
		// fix case of replacement segment: keep lower for mid-word
		if tools.StartsWithUppercase(wordPart) && idx == 0 {
			variant = tools.UppercaseFirstChar(variant)
		}
		msg := "Möglicher Tippfehler: " + bad + "/" + good
		for _, p := range lowercaseProhibitedPairs {
			if (p.part1 == bad || p.part2 == bad) && (p.part1 == good || p.part2 == good ||
				strings.HasPrefix(bad, p.part1) || strings.HasPrefix(bad, p.part2)) {
				msg = "Möglicher Tippfehler. " + tools.UppercaseFirstChar(p.part1) + ": " + p.desc1 + ", " +
					tools.UppercaseFirstChar(p.part2) + ": " + p.desc2
				break
			}
		}
		// map short keys to pair descs
		if bad == "lehr" || good == "leer" {
			msg = "Möglicher Tippfehler. Leer: ohne Inhalt, Lehr: bezogen auf Ausbildung und Wissen"
		}
		rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
		rm.ShortMessage = "vermutlich falsches Kompositum"
		rm.SetSuggestedReplacement(variant)
		return rm
	}
	return nil
}
