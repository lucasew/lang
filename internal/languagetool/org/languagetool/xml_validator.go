package languagetool

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// XMLValidator ports org.languagetool.XMLValidator well-formedness checks.
// Full DTD/XSD validation is deferred (uses encoding/xml parse only).
type XMLValidator struct{}

func NewXMLValidator() *XMLValidator { return &XMLValidator{} }

// ValidateXMLString checks that xmlStr is well-formed XML.
// dtdFile and docType are accepted for API parity and ignored for now.
func (v *XMLValidator) ValidateXMLString(xmlStr, dtdFile, docType string) error {
	dec := xml.NewDecoder(strings.NewReader(xmlStr))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("XML validation failed: %w", err)
		}
	}
}

// ValidateWellFormed is an alias without DTD args.
func (v *XMLValidator) ValidateWellFormed(xmlStr string) error {
	return v.ValidateXMLString(xmlStr, "", "")
}
