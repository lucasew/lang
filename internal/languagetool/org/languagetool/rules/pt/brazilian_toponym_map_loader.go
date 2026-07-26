package pt

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
)

// Official LT resource path under the resource dir (Java toponymFilepath).
// Embedded bytes live at data/brazilian_municipalities/*.tsv and must match
// inspiration/.../resource/pt/brazilian_municipalities/*.tsv.
const brazilianToponymFilepath = "pt/brazilian_municipalities"

//go:embed data/brazilian_municipalities/*.tsv
var brazilianMunicipalitiesFS embed.FS

// brazilianToponymStates ports BrazilianToponymMapLoader.states (exact Java order).
var brazilianToponymStates = []string{
	"AC", // Acre
	"AL", // Alagoas
	"AP", // Amapá
	"AM", // Amazonas
	"BA", // Bahia
	"CE", // Ceará
	"DF", // Distrito Federal
	"ES", // Espírito Santo
	"GO", // Goiás
	"MA", // Maranhão
	"MT", // Mato Grosso
	"MS", // Mato Grosso do Sul
	"MG", // Minas Gerais
	"PA", // Pará
	"PB", // Paraíba
	"PR", // Paraná
	"PE", // Pernambuco
	"PI", // Piauí
	"RJ", // Rio de Janeiro
	"RN", // Rio Grande do Norte
	"RS", // Rio Grande do Sul
	"RO", // Rondônia
	"RR", // Roraima
	"SC", // Santa Catarina
	"SP", // São Paulo
	"SE", // Sergipe
	"TO", // Tocantins
}

// BrazilianToponymMapLoader ports org.languagetool.rules.pt.BrazilianToponymMapLoader.
type BrazilianToponymMapLoader struct {
	// Java: private final String toponymFilepath = "pt/brazilian_municipalities";
	toponymFilepath string
	// Java: private final List<String> states = Arrays.asList(...);
	states []string
}

// NewBrazilianToponymMapLoader ports package-private BrazilianToponymMapLoader().
func NewBrazilianToponymMapLoader() *BrazilianToponymMapLoader {
	return &BrazilianToponymMapLoader{
		toponymFilepath: brazilianToponymFilepath,
		states:          brazilianToponymStates,
	}
}

// getToponymsFromState ports private getToponymsFromState(String state).
// Java:
//
//	List<String> toponyms = getDataBroker().getFromResourceDirAsLines(toponymFilepath + "/" + state + ".tsv");
//	return toponyms.stream().map(t -> t.replace('-', ' ').toLowerCase()).collect(Collectors.toList());
//
// getFromResourceDirAsLines keeps line content as-is (no trim); missing resource → RuntimeException.
func (l *BrazilianToponymMapLoader) getToponymsFromState(state string) []string {
	// Embedded twin of resource path toponymFilepath + "/" + state + ".tsv"
	path := "data/brazilian_municipalities/" + state + ".tsv"
	f, err := brazilianMunicipalitiesFS.Open(path)
	if err != nil {
		// Java DefaultResourceDataBroker.getFromResourceDirAsLines throws RuntimeException if missing.
		panic(fmt.Sprintf("RuntimeException: cannot load %s/%s.tsv: %v", l.toponymFilepath, state, err))
	}
	defer f.Close()

	var toponyms []string
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		// BufferedReader.readLine / Go Scanner: strip terminators only.
		// Keep a possible leftover CR from CRLF resources (same as other ports).
		line := strings.TrimSuffix(sc.Text(), "\r")
		// Java: toponym.replace('-', ' ').toLowerCase() — no trim, no empty-skip.
		norm := strings.ToLower(strings.ReplaceAll(line, "-", " "))
		toponyms = append(toponyms, norm)
	}
	if err := sc.Err(); err != nil {
		panic(fmt.Sprintf("RuntimeException: reading %s/%s.tsv: %v", l.toponymFilepath, state, err))
	}
	return toponyms
}

// BuildMap ports public Map<String, List<String>> buildMap().
func (l *BrazilianToponymMapLoader) BuildMap() map[string][]string {
	if l == nil {
		l = NewBrazilianToponymMapLoader()
	}
	if len(l.states) == 0 {
		l.states = brazilianToponymStates
	}
	if l.toponymFilepath == "" {
		l.toponymFilepath = brazilianToponymFilepath
	}
	m := make(map[string][]string, len(l.states))
	for _, state := range l.states {
		m[state] = l.getToponymsFromState(state)
	}
	return m
}
