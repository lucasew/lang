package morfologik

import (
	"bytes"
	"encoding/hex"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCFSA2Serializer_RoundTripMembership(t *testing.T) {
	words := []string{"receive", "recipe", "the", "cat", "software", "house"}
	built := BuildFSAFromWords(words)
	raw, err := SerializeFSA(built)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 8 || raw[4] != versionCFSA2 {
		t.Fatalf("not CFSA2 header: %x", raw[:min(8, len(raw))])
	}
	fsa, err := ReadFSA(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range words {
		kind, _, _ := fsa.Match([]byte(w), fsa.RootNode())
		if kind != ExactMatch {
			t.Errorf("%q after CFSA2 round-trip kind=%d want ExactMatch", w, kind)
		}
	}
	// non-words
	for _, w := range []string{"xyz", "receiv"} {
		kind, _, _ := fsa.Match([]byte(w), fsa.RootNode())
		if kind == ExactMatch {
			t.Errorf("%q should not be ExactMatch", w)
		}
	}
}

func TestCFSA2Serializer_MatchesJavaOracle(t *testing.T) {
	// Java morfologik 2.2.0 CFSA2Serializer.serialize(FSABuilder.build(words))
	oracleDir := "/tmp/mf-oracle"
	cp := filepath.Join(oracleDir, "morfologik-fsa-2.2.0.jar") + ":" +
		filepath.Join(oracleDir, "morfologik-fsa-builders-2.2.0.jar") + ":" +
		filepath.Join(oracleDir, "hppc-0.8.2.jar")
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("no java")
	}
	cases := [][]string{
		{"receive", "recipe", "the", "cat"},
		{"a", "ab", "abc", "b", "ba"},
		{"software", "house", "the"},
	}
	for _, words := range cases {
		args := append([]string{"-cp", oracleDir + ":" + cp, "CFSA2Oracle"}, words...)
		out, err := exec.Command("java", args...).CombinedOutput()
		if err != nil {
			t.Skipf("java oracle unavailable: %v %s", err, out)
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) < 1 {
			t.Fatal("empty oracle")
		}
		javaHex := strings.TrimSpace(lines[0])
		javaBytes, err := hex.DecodeString(javaHex)
		if err != nil {
			t.Fatal(err)
		}
		built := BuildFSAFromWords(words)
		goBytes, err := SerializeFSA(built)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(goBytes, javaBytes) {
			t.Errorf("words=%v\n go=%x\njava=%x\n goLen=%d javaLen=%d",
				words, goBytes, javaBytes, len(goBytes), len(javaBytes))
		}
	}
}

func TestNewDictionaryFromWords_UsesCFSA2(t *testing.T) {
	d := NewDictionaryFromWords([]string{"receive", "recipe", "the", "cat"}, nil)
	if d == nil || d.FSA == nil {
		t.Fatal("nil dict")
	}
	if d.FSA.version != versionCFSA2 {
		t.Fatalf("want CFSA2 version, got 0x%02x", d.FSA.version)
	}
	if !d.Contains("receive") || d.Contains("xyz") {
		t.Fatal("membership")
	}
	sp := NewSpeller(d, 1)
	cds := sp.FindReplacementCandidatesFull("recieve", false)
	found := false
	for _, c := range cds {
		if c.Word == "receive" {
			found = true
		}
	}
	if !found {
		t.Fatalf("suggest via CFSA2 dict: %+v", cds)
	}
	// sticky separators still applies (ExactMatch plain words)
	if !sp.IsInDictionary("receive") {
		t.Fatal("in dict")
	}
	if sp.ContainsSeparators() {
		t.Fatal("plain CFSA2 words ExactMatch clears separators")
	}
}
