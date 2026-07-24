package languagetool

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
)

// XMLValidator ports org.languagetool.XMLValidator well-formedness checks.
// Full DTD/XSD validation is partial (encoding/xml + optional entity resolve);
// Java uses SAX/Schema factory with entity expansion limits.
type XMLValidator struct {
	// Broker optional for ValidateWithDtd classpath-style loads.
	Broker broker.ResourceDataBroker
	// SchemaLookup returns schema content by classpath path (optional).
	SchemaLookup func(path string) (io.ReadCloser, error)
}

func NewXMLValidator() *XMLValidator { return &XMLValidator{} }

// ValidateXMLString ports validateXMLString — well-formedness (+ DTD path accepted).
// dtdFile and docType are accepted for API parity; when dtdFile non-empty, DOCTYPE
// injection path is mirrored (strip existing DOCTYPE, require XML decl).
func (v *XMLValidator) ValidateXMLString(xmlStr, dtdFile, docType string) error {
	if dtdFile != "" {
		return v.validateInternalDTD(xmlStr, dtdFile, docType)
	}
	return v.validateWellFormed(xmlStr)
}

// ValidateWellFormed is an alias without DTD args.
func (v *XMLValidator) ValidateWellFormed(xmlStr string) error {
	return v.validateWellFormed(xmlStr)
}

func (v *XMLValidator) validateWellFormed(xmlStr string) error {
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

var doctypeRE = regexp.MustCompile(`(?s)<!DOCTYPE.+?>`)

// validateInternalDTD ports private validateInternal(xml, dtdPath, docType) up to parse.
// Full SAX validating parse with DTD is not available in encoding/xml; after DOCTYPE
// rewrite we require well-formed XML (DTD existence checked when SchemaLookup/Broker set).
func (v *XMLValidator) validateInternalDTD(xmlStr, dtdPath, docType string) error {
	cleanXML := doctypeRE.ReplaceAllString(xmlStr, "")
	const decl = `<?xml version="1.0"`
	const endDecl = `?>`
	pos := strings.Index(cleanXML, decl)
	endPos := strings.Index(cleanXML, endDecl)
	if pos == -1 {
		snippet := cleanXML
		if len(snippet) > 100 {
			snippet = snippet[:100]
		}
		return fmt.Errorf("No XML declaration found in '%s...'", snippet)
	}
	// Check DTD path resolvable when possible
	if v != nil && v.SchemaLookup != nil {
		rc, err := v.SchemaLookup(dtdPath)
		if err != nil || rc == nil {
			return fmt.Errorf("DTD not found in classpath: %s", dtdPath)
		}
		_ = rc.Close()
	}
	// Inject DOCTYPE like Java (dtd URL string not required for well-formed check)
	dtd := fmt.Sprintf(`<!DOCTYPE %s PUBLIC "-//W3C//DTD Rules 0.1//EN" "%s">`, docType, dtdPath)
	newXML := cleanXML[:endPos+len(endDecl)] + "\r\n" + dtd + cleanXML[endPos+len(endDecl):]
	// encoding/xml may reject DOCTYPE; strip again for well-formed token pass
	return v.validateWellFormed(doctypeRE.ReplaceAllString(newXML, ""))
}

// ValidateWithDtd ports validateWithDtd(filename, dtdPath, docType) via Broker.
func (v *XMLValidator) ValidateWithDtd(filename, dtdPath, docType string) error {
	if v == nil || v.Broker == nil {
		return fmt.Errorf("Not found in classpath: %s", filename)
	}
	rc, err := v.Broker.GetAsStream(filename)
	if err != nil || rc == nil {
		return fmt.Errorf("Not found in classpath: %s", filename)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("Cannot load or parse '%s': %w", filename, err)
	}
	if err := v.ValidateXMLString(string(data), dtdPath, docType); err != nil {
		return fmt.Errorf("Cannot load or parse '%s': %w", filename, err)
	}
	return nil
}

// XMLErrorHandler ports XMLValidator.ErrorHandler — throw on warning/error.
type XMLErrorHandler struct {
	// OnMessage optional logger (Java System.err.println).
	OnMessage func(string)
}

// Warning ports warning(SAXParseException) — logs and returns error.
func (h *XMLErrorHandler) Warning(msg string, line, column int) error {
	full := fmt.Sprintf("%s Problem found at line %d, column %d.", msg, line, column)
	if h != nil && h.OnMessage != nil {
		h.OnMessage(full)
	} else {
		fmt.Fprintln(os.Stderr, full)
	}
	return fmt.Errorf("%s", full)
}

// Error ports error(SAXParseException).
func (h *XMLErrorHandler) Error(msg string, line, column int) error {
	return h.Warning(msg, line, column)
}

// LSRuleEntityResolver ports XMLValidator.LSRuleEntityResolver for .ent resources.
type LSRuleEntityResolver struct {
	Resolver *RuleEntityResolver
}

// ResolveResource ports resolveResource — returns stream for .ent systemIds.
func (r *LSRuleEntityResolver) ResolveResource(systemID string) (io.ReadCloser, error) {
	if systemID == "" || !strings.HasSuffix(systemID, ".ent") {
		return nil, nil
	}
	if r == nil || r.Resolver == nil {
		return nil, nil
	}
	return r.Resolver.ResolveEntity(systemID)
}

// EntityAsInput ports XMLValidator.EntityAsInput (LSInput twin simplified).
type EntityAsInput struct {
	PublicID     string
	SystemID     string
	InputStream  io.ReadCloser
	ruleResolver *RuleEntityResolver
}

// NewEntityAsInput ports EntityAsInput(publicId, systemId).
func NewEntityAsInput(publicID, systemID string, resolver *RuleEntityResolver) (*EntityAsInput, error) {
	e := &EntityAsInput{PublicID: publicID, SystemID: systemID, ruleResolver: resolver}
	if err := e.setInputStream(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *EntityAsInput) setInputStream() error {
	if e.SystemID != "" && strings.HasSuffix(e.SystemID, ".ent") && e.ruleResolver != nil {
		rc, err := e.ruleResolver.ResolveEntity(e.SystemID)
		if err != nil {
			return err
		}
		e.InputStream = rc
	}
	return nil
}

func (e *EntityAsInput) GetByteStream() io.ReadCloser {
	if e == nil {
		return nil
	}
	return e.InputStream
}

func (e *EntityAsInput) GetSystemId() string {
	if e == nil {
		return ""
	}
	return e.SystemID
}

func (e *EntityAsInput) GetPublicId() string {
	if e == nil {
		return ""
	}
	return e.PublicID
}
