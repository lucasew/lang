package tokenizers

import (
	"bufio"
	"io"
	"strings"
)

// SrxDocument is a minimal stand-in for loomchild SrxDocument (rule language codes).
type SrxDocument struct {
	// LanguageCodes listed in the SRX (best-effort from <languagerule> tags).
	LanguageCodes []string
	// Raw is the original SRX payload (optional).
	Raw string
}

// SrxTools ports org.languagetool.tokenizers.SrxTools (without loomchild dependency).
// CreateSrxDocument loads XML text; TokenizeWithSrx falls back to SRXSentenceTokenizer.
type SrxTools struct{}

// CreateSrxDocumentFromReader parses a lightweight SRX document from r.
func CreateSrxDocumentFromReader(r io.Reader) (*SrxDocument, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return CreateSrxDocumentFromString(string(data)), nil
}

// CreateSrxDocumentFromString builds an SrxDocument by scanning language rule codes.
func CreateSrxDocumentFromString(xml string) *SrxDocument {
	doc := &SrxDocument{Raw: xml}
	// crude attribute scan: languagerule languagerulename="en"
	sc := bufio.NewScanner(strings.NewReader(xml))
	seen := map[string]struct{}{}
	for sc.Scan() {
		line := sc.Text()
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "languagerule") {
			continue
		}
		// find languagerulename="..."
		const key = `languagerulename="`
		i := strings.Index(lower, key)
		if i < 0 {
			const key2 = `languagerulename='`
			i = strings.Index(lower, key2)
			if i < 0 {
				continue
			}
			rest := line[i+len(key2):]
			j := strings.IndexByte(rest, '\'')
			if j > 0 {
				code := rest[:j]
				if _, ok := seen[code]; !ok {
					seen[code] = struct{}{}
					doc.LanguageCodes = append(doc.LanguageCodes, code)
				}
			}
			continue
		}
		// use original line for case of code
		idx := strings.Index(strings.ToLower(line), key)
		rest := line[idx+len(key):]
		j := strings.IndexByte(rest, '"')
		if j > 0 {
			code := rest[:j]
			if _, ok := seen[code]; !ok {
				seen[code] = struct{}{}
				doc.LanguageCodes = append(doc.LanguageCodes, code)
			}
		}
	}
	return doc
}

// TokenizeWithSrx segments text using SRXSentenceTokenizer for languageCode.
// The document is accepted for API parity; full rule evaluation is deferred.
func TokenizeWithSrx(text string, doc *SrxDocument, languageCode string) []string {
	_ = doc
	tok := NewSRXSentenceTokenizer(languageCode)
	return tok.Tokenize(text)
}
