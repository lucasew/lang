package pt

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Java states list order/count (BrazilianToponymMapLoader.states).
var wantBrazilianToponymStates = []string{
	"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
	"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
	"RS", "RO", "RR", "SC", "SP", "SE", "TO",
}

func TestBrazilianToponymMapLoader_StatesOrderAndCount(t *testing.T) {
	require.Equal(t, 27, len(brazilianToponymStates))
	require.Equal(t, wantBrazilianToponymStates, brazilianToponymStates)
	require.Equal(t, "pt/brazilian_municipalities", brazilianToponymFilepath)

	l := NewBrazilianToponymMapLoader()
	require.Equal(t, wantBrazilianToponymStates, l.states)
	require.Equal(t, brazilianToponymFilepath, l.toponymFilepath)
}

func TestBrazilianToponymMapLoader_BuildMap(t *testing.T) {
	m := NewBrazilianToponymMapLoader().BuildMap()
	require.Equal(t, 27, len(m))
	for _, st := range wantBrazilianToponymStates {
		list, ok := m[st]
		require.Truef(t, ok, "missing state key %s", st)
		require.NotEmptyf(t, list, "state %s should have municipalities", st)
	}

	// Sample SP municipalities from official file after Java normalize.
	sp := m["SP"]
	require.Contains(t, sp, "são paulo")
	require.Contains(t, sp, "adamantina")
	// Hyphen → space + lower (Arco-Íris, Embu-Guaçu in SP.tsv).
	require.Contains(t, sp, "arco íris")
	require.Contains(t, sp, "embu guaçu")
	// DF capital.
	require.Contains(t, m["DF"], "brasília")
}

func TestBrazilianToponymMapLoader_HyphenNormalization(t *testing.T) {
	// Java stream map only: replace('-',' ').toLowerCase() — no trim.
	l := NewBrazilianToponymMapLoader()
	// SP file contains "Arco-Íris"
	list := l.getToponymsFromState("SP")
	require.Contains(t, list, "arco íris")
	// Ensure raw hyphen form is not kept.
	require.NotContains(t, list, "arco-íris")
	require.NotContains(t, list, "Arco-Íris")
}

func TestBrazilianToponymMapLoader_NoSilentSkipMissingState(t *testing.T) {
	l := NewBrazilianToponymMapLoader()
	require.Panics(t, func() {
		_ = l.getToponymsFromState("XX")
	})
}

func TestBrazilianToponymMapLoader_ResourceIdentity(t *testing.T) {
	// Embedded data must be byte-identical to official Java resources.
	inspRoot := findInspirationMunicipalities(t)
	for _, st := range wantBrazilianToponymStates {
		name := st + ".tsv"
		insp, err := os.ReadFile(filepath.Join(inspRoot, name))
		require.NoError(t, err, name)
		emb, err := brazilianMunicipalitiesFS.ReadFile("data/brazilian_municipalities/" + name)
		require.NoError(t, err, "embed "+name)
		require.Truef(t, bytes.Equal(insp, emb),
			"%s: embedded data differs from inspiration resource (len insp=%d emb=%d)",
			name, len(insp), len(emb))
	}
}

func TestBrazilianToponymMapLoader_LineCountMatchesFile(t *testing.T) {
	// getFromResourceDirAsLines keeps every line; normalize does not drop empties.
	// Official files have no empty lines; count must match raw line count from embed.
	l := NewBrazilianToponymMapLoader()
	for _, st := range wantBrazilianToponymStates {
		raw, err := brazilianMunicipalitiesFS.ReadFile("data/brazilian_municipalities/" + st + ".tsv")
		require.NoError(t, err)
		// Count lines as BufferedReader.readLine would (split on \n; last non-empty-terminated line counts).
		wantLines := countReadLineLines(raw)
		got := l.getToponymsFromState(st)
		require.Equalf(t, wantLines, len(got), "state %s line count", st)
	}
}

func countReadLineLines(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	n := 0
	start := 0
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			n++
			start = i + 1
		}
	}
	if start < len(b) {
		n++ // last line without trailing newline
	}
	return n
}

func findInspirationMunicipalities(t *testing.T) string {
	t.Helper()
	const rel = "inspiration/languagetool/languagetool-language-modules/pt/src/main/resources/org/languagetool/resource/pt/brazilian_municipalities"
	dir, err := os.Getwd()
	require.NoError(t, err)
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("inspiration brazilian_municipalities not found from cwd")
	return ""
}
