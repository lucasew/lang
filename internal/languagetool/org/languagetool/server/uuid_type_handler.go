package server

import (
	"encoding/binary"
	"fmt"
)

// UUIDTypeHandler ports org.languagetool.server.UUIDTypeHandler binary codec.
// Represents UUID as [16]byte in big-endian most/least significant longs (Java layout).

// UUIDBits holds Java-style UUID most/least significant bits.
type UUIDBits struct {
	MostSignificant  uint64
	LeastSignificant uint64
}

func UUIDBitsToBytes(u UUIDBits) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[0:8], u.MostSignificant)
	binary.BigEndian.PutUint64(b[8:16], u.LeastSignificant)
	return b
}

func BytesToUUIDBits(b []byte) (UUIDBits, error) {
	if b == nil {
		return UUIDBits{}, nil
	}
	if len(b) != 16 {
		return UUIDBits{}, fmt.Errorf("UUID bytes must be length 16, got %d", len(b))
	}
	return UUIDBits{
		MostSignificant:  binary.BigEndian.Uint64(b[0:8]),
		LeastSignificant: binary.BigEndian.Uint64(b[8:16]),
	}, nil
}

// UUIDString formats bits as standard 8-4-4-4-12 hex (version/variant bits as-is).
func (u UUIDBits) String() string {
	b := UUIDBitsToBytes(u)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ParseUUIDString parses "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" into bits.
func ParseUUIDString(s string) (UUIDBits, error) {
	// strip hyphens
	hex := make([]byte, 0, 32)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' {
			continue
		}
		hex = append(hex, c)
	}
	if len(hex) != 32 {
		return UUIDBits{}, fmt.Errorf("invalid UUID string")
	}
	raw := make([]byte, 16)
	for i := 0; i < 16; i++ {
		var v byte
		_, err := fmt.Sscanf(string(hex[i*2:i*2+2]), "%02x", &v)
		if err != nil {
			return UUIDBits{}, err
		}
		raw[i] = v
	}
	return BytesToUUIDBits(raw)
}
