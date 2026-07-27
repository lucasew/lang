package ga

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/people.txt
var peopleFS embed.FS

// DhaNoBeirtPeoplePath is the official rules-dir path Java loads
// (JLanguageTool.getDataBroker().getFromRulesDirAsStream("/ga/people.txt")).
const DhaNoBeirtPeoplePath = "/ga/people.txt"

var (
	daoineOnce sync.Once
	daoineSet  map[string]bool
)

// numberReplacements ports DhaNoBeirtData static numbers map (exact 9 entries).
// Keys are lowercase; Match uses exact-token get (not case-fold) like Java HashMap.get.
var numberReplacements = map[string]string{
	"dhá":      "beirt",
	"trí":      "triúr",
	"ceathair": "ceathrar",
	"cúig":     "cúigear",
	"sé":       "seisear",
	"seacht":   "seachtar",
	"ocht":     "ochtar",
	"naoi":     "naonúr",
	"deich":    "deichniúr",
}

// DhaNoBeirtData ports org.languagetool.rules.ga.DhaNoBeirtData (package-private leaf).
// Loads humans from official /ga/people.txt; number map is the static block twin.
type DhaNoBeirtData struct{}

// getDaoine ports DhaNoBeirtData.getDaoine() static (used by DhaNoBeirtRule).
func getDaoine() map[string]bool {
	return loadDaoine()
}

// getNumberReplacements ports DhaNoBeirtData.getNumberReplacements() static.
// Returns the live map (Java returns the static HashMap reference).
func getNumberReplacements() map[string]string {
	return numberReplacements
}

// GetDaoine ports DhaNoBeirtData.getDaoine for type-method twin access.
func (DhaNoBeirtData) GetDaoine() map[string]bool {
	return getDaoine()
}

// GetNumberReplacements ports DhaNoBeirtData.getNumberReplacements for type-method twin.
func (DhaNoBeirtData) GetNumberReplacements() map[string]string {
	return getNumberReplacements()
}

// loadDaoine ports DhaNoBeirtData.loadWords("/ga/people.txt"):
// trim; skip empty/#; '*' keeps case of rest; else lowercase.
func loadDaoine() map[string]bool {
	daoineOnce.Do(func() {
		f, err := peopleFS.Open("data/people.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m := map[string]bool{}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := tools.JavaStringTrim(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			if line[0] == '*' {
				// Java: set.add(line.substring(1)) — keep case
				m[line[1:]] = true
			} else {
				// Java: set.add(line.toLowerCase())
				m[strings.ToLower(line)] = true
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		daoineSet = m
	})
	return daoineSet
}
