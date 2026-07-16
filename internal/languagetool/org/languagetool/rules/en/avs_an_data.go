package en

import (
	"bufio"
	"embed"
	"strings"
	"sync"
)

//go:embed data/det_a.txt data/det_an.txt
var avsAnFS embed.FS

var (
	avsAnOnce     sync.Once
	wordsRequireA map[string]bool
	wordsRequireAn map[string]bool
)

func loadAvsAnData() {
	avsAnOnce.Do(func() {
		wordsRequireA = loadDetWords("data/det_a.txt")
		wordsRequireAn = loadDetWords("data/det_an.txt")
	})
}

func loadDetWords(path string) map[string]bool {
	f, err := avsAnFS.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	set := make(map[string]bool)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if line[0] == '*' {
			set[line[1:]] = true
		} else {
			set[strings.ToLower(line)] = true
		}
	}
	return set
}

func getWordsRequiringA() map[string]bool {
	loadAvsAnData()
	return wordsRequireA
}

func getWordsRequiringAn() map[string]bool {
	loadAvsAnData()
	return wordsRequireAn
}
