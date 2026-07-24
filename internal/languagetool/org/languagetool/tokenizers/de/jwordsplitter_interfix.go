package de

// germanInterfixDisambiguator ports de.danielnaber.jwordsplitter.GermanInterfixDisambiguator.
type germanInterfixDisambiguator struct {
	dictionary map[string]struct{}
}

// wordsRequiringInterfixS ports GermanInterfixDisambiguator.wordsRequiringInterfixS.
var wordsRequiringInterfixS = map[string]struct{}{
	"Verhalten": {},
}

func newGermanInterfixDisambiguator(dictionary map[string]struct{}) *germanInterfixDisambiguator {
	return &germanInterfixDisambiguator{dictionary: dictionary}
}

func (d *germanInterfixDisambiguator) disambiguate(parts []string) []string {
	newParts := append([]string(nil), parts...)
	lastPartIdx := len(parts) - 1
	if len(parts) > 1 {
		lastPart := parts[lastPartIdx]
		if lastPart == "samt" || lastPart == "samts" || lastPart == "samtes" {
			// Verkehr+s+amt = Verkehrs+amt
			newParts[lastPartIdx-1] = parts[lastPartIdx-1] + "s"
			// Java: lastPart.replaceFirst("^s", "")
			if stringsHasPrefixS(lastPart) {
				newParts[lastPartIdx] = lastPart[1:]
			} else {
				newParts[lastPartIdx] = lastPart
			}
			return newParts
		}
	}
	for i := len(newParts) - 1; i >= 2; i-- {
		part := newParts[i]
		prevPart := newParts[i-1]
		prevPrevPart := newParts[i-2]
		if prevPart == "s" {
			partIsWord := d.isWord("s" + part)
			_, requires := wordsRequiringInterfixS[prevPrevPart]
			if !partIsWord || requires {
				// Schönheit+s+tempel = Schönheits-tempel
				newParts[i-2] = prevPrevPart + "s"
				// remove interfix at i-1
				newParts = append(newParts[:i-1], newParts[i:]...)
			}
		}
	}
	return newParts
}

func (d *germanInterfixDisambiguator) isWord(word string) bool {
	if d == nil || d.dictionary == nil {
		return false
	}
	_, ok := d.dictionary[javaStringToLower(word)]
	return ok
}

func stringsHasPrefixS(s string) bool {
	return len(s) > 0 && s[0] == 's'
}
