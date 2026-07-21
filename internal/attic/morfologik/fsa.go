// Package morfologik implements a subset of morfologik-fsa / morfologik-stemming
// used by LanguageTool (CFSA2 and FSA5 automata + dictionary lookup).
package morfologik

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	fsaMagic0 = '\\'
	fsaMagic1 = 'f'
	fsaMagic2 = 's'
	fsaMagic3 = 'a'

	// Versions (Java morfologik.fsa.FSAHeader)
	versionFSA5         byte = 0x05
	versionCFSA2        byte = 0xc6
	versionConstantArc  byte = 0x00 // in-memory ConstantArcSizeFSA (FSABuilder), not file format

	// CFSA2 flags
	cfsa2BitTargetNext = 1 << 7
	cfsa2BitLastArc    = 1 << 6
	cfsa2BitFinalArc   = 1 << 5
	cfsa2LabelIndexBits = 5
	cfsa2LabelIndexMask = (1 << cfsa2LabelIndexBits) - 1
	cfsa2FlagNumbers   = 1 << 8

	// FSA5 flags (Java FSA5)
	fsa5BitFinalArc    = 1 << 0
	fsa5BitLastArc     = 1 << 1
	fsa5BitTargetNext  = 1 << 2
	fsa5AddressOffset  = 1
)

// FSA is a morfologik automaton (CFSA2, FSA5, or in-memory ConstantArcSizeFSA).
type FSA struct {
	version byte
	arcs    []byte

	// CFSA2
	labelMapping []byte
	hasNumbers   bool

	// FSA5
	nodeDataLength int
	gtl            int // goto length in bytes
	filler         byte
	annotation     byte

	// ConstantArcSizeFSA (FSABuilder)
	constantArc bool
	constantEps int // epsilon state offset (usually 0)
}

// OpenFSA loads a morfologik FSA from path.
func OpenFSA(path string) (*FSA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadFSA(f)
}

// ReadFSA reads a CFSA2 or FSA5 automaton from r.
func ReadFSA(r io.Reader) (*FSA, error) {
	var magic [4]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if magic[0] != fsaMagic0 || magic[1] != fsaMagic1 || magic[2] != fsaMagic2 || magic[3] != fsaMagic3 {
		return nil, fmt.Errorf("invalid FSA magic")
	}
	var ver [1]byte
	if _, err := io.ReadFull(r, ver[:]); err != nil {
		return nil, err
	}
	switch ver[0] {
	case versionCFSA2:
		return readCFSA2(r)
	case versionFSA5:
		return readFSA5(r)
	default:
		return nil, fmt.Errorf("unsupported FSA version 0x%02x (want CFSA2 0xc6 or FSA5 0x05)", ver[0])
	}
}

func readCFSA2(r io.Reader) (*FSA, error) {
	var flags uint16
	if err := binary.Read(r, binary.BigEndian, &flags); err != nil {
		return nil, err
	}
	hasNumbers := flags&cfsa2FlagNumbers != 0

	var labelSize [1]byte
	if _, err := io.ReadFull(r, labelSize[:]); err != nil {
		return nil, err
	}
	labelMapping := make([]byte, labelSize[0])
	if _, err := io.ReadFull(r, labelMapping); err != nil {
		return nil, err
	}
	arcs, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &FSA{
		version:      versionCFSA2,
		arcs:         arcs,
		labelMapping: labelMapping,
		hasNumbers:   hasNumbers,
	}, nil
}

func readFSA5(r io.Reader) (*FSA, error) {
	// Java FSA5: filler, annotation, hgtl
	var hdr [3]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	filler, annotation, hgtl := hdr[0], hdr[1], hdr[2]
	nodeDataLength := int((hgtl >> 4) & 0x0f)
	gtl := int(hgtl & 0x0f)
	if gtl < 1 {
		return nil, fmt.Errorf("FSA5 invalid goto length %d", gtl)
	}
	arcs, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &FSA{
		version:        versionFSA5,
		arcs:           arcs,
		nodeDataLength: nodeDataLength,
		gtl:            gtl,
		filler:         filler,
		annotation:     annotation,
	}, nil
}

func (f *FSA) RootNode() int {
	if f.constantArc || f.version == versionConstantArc {
		// ConstantArcSizeFSA: getEndNode(getFirstArc(epsilon))
		return f.destinationNodeOffset(f.firstArc(f.constantEps))
	}
	if f.version == versionFSA5 {
		// Skip dummy node marking terminating state, then follow epsilon.
		epsilonNode := f.skipArc(f.firstArc(0))
		return f.destinationNodeOffset(f.firstArc(epsilonNode))
	}
	return f.destinationNodeOffset(f.firstArc(0))
}

func (f *FSA) firstArc(node int) int {
	if f.constantArc || f.version == versionConstantArc {
		// Java ConstantArcSizeFSA: return node (arcs packed from node offset)
		if node == 0 && !f.constantArc {
			// terminal sentinel
		}
		return node
	}
	if f.version == versionFSA5 {
		return f.nodeDataLength + node
	}
	if f.hasNumbers {
		return f.skipVInt(node)
	}
	return node
}

func (f *FSA) nextArc(arc int) int {
	if f.isArcLast(arc) {
		return 0
	}
	if f.constantArc || f.version == versionConstantArc {
		return arc + casArcSize
	}
	return f.skipArc(arc)
}

func (f *FSA) getArc(node int, label byte) int {
	for arc := f.firstArc(node); arc != 0; arc = f.nextArc(arc) {
		if f.arcLabel(arc) == label {
			return arc
		}
	}
	return 0
}

func (f *FSA) endNode(arc int) int {
	return f.destinationNodeOffset(arc)
}

func (f *FSA) arcLabel(arc int) byte {
	if f.constantArc || f.version == versionConstantArc {
		return f.arcs[arc+casLabelOffset]
	}
	if f.version == versionFSA5 {
		return f.arcs[arc]
	}
	index := int(f.arcs[arc] & cfsa2LabelIndexMask)
	if index > 0 {
		return f.labelMapping[index]
	}
	return f.arcs[arc+1]
}

func (f *FSA) isArcFinal(arc int) bool {
	if f.constantArc || f.version == versionConstantArc {
		return f.arcs[arc+casFlagsOffset]&casBitArcFinal != 0
	}
	if f.version == versionFSA5 {
		return f.arcs[arc+fsa5AddressOffset]&fsa5BitFinalArc != 0
	}
	return f.arcs[arc]&cfsa2BitFinalArc != 0
}

func (f *FSA) isArcTerminal(arc int) bool {
	return f.destinationNodeOffset(arc) == 0
}

func (f *FSA) isArcLast(arc int) bool {
	if f.constantArc || f.version == versionConstantArc {
		return f.arcs[arc+casFlagsOffset]&casBitArcLast != 0
	}
	if f.version == versionFSA5 {
		return f.arcs[arc+fsa5AddressOffset]&fsa5BitLastArc != 0
	}
	return f.arcs[arc]&cfsa2BitLastArc != 0
}

func (f *FSA) isNextSet(arc int) bool {
	if f.constantArc || f.version == versionConstantArc {
		return false // constant-arc always stores full address
	}
	if f.version == versionFSA5 {
		return f.arcs[arc+fsa5AddressOffset]&fsa5BitTargetNext != 0
	}
	return f.arcs[arc]&cfsa2BitTargetNext != 0
}

func (f *FSA) destinationNodeOffset(arc int) int {
	if f.constantArc || f.version == versionConstantArc {
		a := arc + casAddressOffset
		return int(f.arcs[a])<<24 |
			int(f.arcs[a+1]&0xff)<<16 |
			int(f.arcs[a+2]&0xff)<<8 |
			int(f.arcs[a+3]&0xff)
	}
	if f.version == versionFSA5 {
		if f.isNextSet(arc) {
			return f.skipArc(arc)
		}
		// decodeFromBytes >>> 3
		return decodeFromBytes(f.arcs, arc+fsa5AddressOffset, f.gtl) >> 3
	}
	if f.isNextSet(arc) {
		for !f.isArcLast(arc) {
			arc = f.nextArc(arc)
		}
		return f.skipArc(arc)
	}
	off := arc + 1
	if f.arcs[arc]&cfsa2LabelIndexMask == 0 {
		off = arc + 2
	}
	return readVInt(f.arcs, off)
}

func (f *FSA) skipArc(offset int) int {
	if f.version == versionFSA5 {
		if f.isNextSet(offset) {
			return offset + 1 + 1 // label + flags
		}
		return offset + 1 + f.gtl // label + flags/address
	}
	flag := f.arcs[offset]
	offset++
	if flag&cfsa2LabelIndexMask == 0 {
		offset++
	}
	if flag&cfsa2BitTargetNext == 0 {
		offset = f.skipVInt(offset)
	}
	return offset
}

func (f *FSA) skipVInt(offset int) int {
	for f.arcs[offset]&0x80 != 0 {
		offset++
	}
	return offset + 1
}

// decodeFromBytes ports FSA5.decodeFromBytes (little-endian packed, n bytes).
func decodeFromBytes(arcs []byte, start, n int) int {
	r := 0
	for i := n; i > 0; i-- {
		r = r<<8 | int(arcs[start+i-1]&0xff)
	}
	return r
}

func readVInt(array []byte, offset int) int {
	b := array[offset]
	value := int(b & 0x7F)
	for shift := 7; b&0x80 != 0; shift += 7 {
		offset++
		b = array[offset]
		value |= int(b&0x7F) << shift
	}
	return value
}

// Match kinds from FSATraversal.
const (
	ExactMatch         = 0
	NoMatch            = -1
	AutomatonHasPrefix = -3
	SequenceIsAPrefix  = -4
)

// Match walks the automaton for sequence starting at node.
func (f *FSA) Match(sequence []byte, node int) (kind, index, outNode int) {
	if node == 0 {
		return NoMatch, 0, 0
	}
	for i := 0; i < len(sequence); i++ {
		arc := f.getArc(node, sequence[i])
		if arc == 0 {
			if i > 0 {
				return AutomatonHasPrefix, i, node
			}
			return NoMatch, i, node
		}
		if i+1 == len(sequence) && f.isArcFinal(arc) {
			return ExactMatch, i, node
		}
		if f.isArcTerminal(arc) {
			return AutomatonHasPrefix, i + 1, node
		}
		node = f.endNode(arc)
	}
	return SequenceIsAPrefix, 0, node
}

// CollectFinalSequences returns all byte sequences from node to final arcs (right language).
func (f *FSA) CollectFinalSequences(node int) [][]byte {
	var out [][]byte
	var buf []byte
	var walk func(n int)
	walk = func(n int) {
		for arc := f.firstArc(n); arc != 0; arc = f.nextArc(arc) {
			buf = append(buf, f.arcLabel(arc))
			if f.isArcFinal(arc) {
				cp := make([]byte, len(buf))
				copy(cp, buf)
				out = append(out, cp)
			}
			if !f.isArcTerminal(arc) {
				walk(f.endNode(arc))
			}
			buf = buf[:len(buf)-1]
		}
	}
	walk(node)
	return out
}
