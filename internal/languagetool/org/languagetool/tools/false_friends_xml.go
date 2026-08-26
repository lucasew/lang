package tools

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ValidateFalseFriendsXML checks that false-friends.xml is well-formed XML
// with a root <rules> element (structure surface of ValidateFalseFriendsXmlTest).
func ValidateFalseFriendsXML(r io.Reader) error {
	if r == nil {
		return fmt.Errorf("nil reader")
	}
	dec := xml.NewDecoder(r)
	dec.Strict = false // allow DOCTYPE / processing instructions loosely
	var root struct {
		XMLName xml.Name `xml:"rules"`
	}
	if err := dec.Decode(&root); err != nil {
		return fmt.Errorf("false-friends.xml is not well-formed: %w", err)
	}
	if !strings.EqualFold(root.XMLName.Local, "rules") {
		return fmt.Errorf("expected root element <rules>, got <%s>", root.XMLName.Local)
	}
	return nil
}
