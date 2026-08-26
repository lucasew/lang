package tools

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// AsStream ports JLanguageTool.getDataBroker().getAsStream for Tools.getStream.
// Wire from the data broker (or a test map). Nil → GetStream fails closed.
var AsStream func(path string) (io.ReadCloser, error)

// GetStream ports Tools.getStream(path).
// Error message matches Java: Could not load file from classpath: '…'
func GetStream(path string) (io.ReadCloser, error) {
	if AsStream == nil {
		return nil, fmt.Errorf("Could not load file from classpath: '%s'", path)
	}
	is, err := AsStream(path)
	if err != nil {
		return nil, err
	}
	if is == nil {
		return nil, fmt.Errorf("Could not load file from classpath: '%s'", path)
	}
	return is, nil
}

// GetStreamWithHash ports Tools.getStream(path, requiredHash).
// requiredHash is lowercase hex SHA-256 (Java HexFormat.of().formatHex).
func GetStreamWithHash(path, requiredHash string) (io.ReadCloser, error) {
	is, err := GetStream(path)
	if err != nil {
		return nil, err
	}
	defer is.Close()
	data, err := io.ReadAll(is)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	computed := hex.EncodeToString(sum[:])
	if computed != requiredHash {
		return nil, fmt.Errorf("Checksum mismatch for the file '%s': expected %s, got %s", path, requiredHash, computed)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

// UseDataBroker wires AsStream to a ResourceDataBroker.GetAsStream-compatible opener.
// Ports the Tools.getStream → JLanguageTool.getDataBroker().getAsStream path.
func UseDataBroker(getAsStream func(path string) (io.ReadCloser, error)) {
	AsStream = getAsStream
}
