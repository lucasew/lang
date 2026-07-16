// Package morfologik implements a subset of morfologik-fsa / morfologik-stemming
// used by LanguageTool (CFSA2 automata + dictionary lookup).
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

	// CFSA2 version
	versionCFSA2 byte = 0xc6

	bitTargetNext = 1 << 7
	bitLastArc    = 1 << 6
	bitFinalArc   = 1 << 5

	labelIndexBits = 5
	labelIndexMask = (1 << labelIndexBits) - 1

	flagNumbers = 1 << 8
)

// FSA is a CFSA2 automaton.
type FSA struct {
	arcs         []byte
	labelMapping []byte
	hasNumbers   bool
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

// ReadFSA reads a CFSA2 automaton from r.
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
	if ver[0] != versionCFSA2 {
		return nil, fmt.Errorf("unsupported FSA version 0x%02x (want CFSA2 0xc6)", ver[0])
	}
	var flags uint16
	if err := binary.Read(r, binary.BigEndian, &flags); err != nil {
		return nil, err
	}
	// Only NUMBERS flag (and known Daciuk bits) expected; accept any known set.
	hasNumbers := flags&flagNumbers != 0

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
	return &FSA{arcs: arcs, labelMapping: labelMapping, hasNumbers: hasNumbers}, nil
}

func (f *FSA) RootNode() int {
	return f.destinationNodeOffset(f.firstArc(0))
}

func (f *FSA) firstArc(node int) int {
	if f.hasNumbers {
		return f.skipVInt(node)
	}
	return node
}

func (f *FSA) nextArc(arc int) int {
	if f.isArcLast(arc) {
		return 0
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
	index := int(f.arcs[arc] & labelIndexMask)
	if index > 0 {
		return f.labelMapping[index]
	}
	return f.arcs[arc+1]
}

func (f *FSA) isArcFinal(arc int) bool {
	return f.arcs[arc]&bitFinalArc != 0
}

func (f *FSA) isArcTerminal(arc int) bool {
	return f.destinationNodeOffset(arc) == 0
}

func (f *FSA) isArcLast(arc int) bool {
	return f.arcs[arc]&bitLastArc != 0
}

func (f *FSA) isNextSet(arc int) bool {
	return f.arcs[arc]&bitTargetNext != 0
}

func (f *FSA) destinationNodeOffset(arc int) int {
	if f.isNextSet(arc) {
		for !f.isArcLast(arc) {
			arc = f.nextArc(arc)
		}
		return f.skipArc(arc)
	}
	off := arc + 1
	if f.arcs[arc]&labelIndexMask == 0 {
		off = arc + 2
	}
	return readVInt(f.arcs, off)
}

func (f *FSA) skipArc(offset int) int {
	flag := f.arcs[offset]
	offset++
	if flag&labelIndexMask == 0 {
		offset++
	}
	if flag&bitTargetNext == 0 {
		offset = f.skipVInt(offset)
	}
	return offset
}

func (f *FSA) skipVInt(offset int) int {
	// Java signed-byte continuation: high bit set means more bytes.
	for f.arcs[offset]&0x80 != 0 {
		offset++
	}
	return offset + 1
}

func readVInt(array []byte, offset int) int {
	b := array[offset]
	value := int(b & 0x7F)
	// Java uses signed byte b < 0 for continuation (== high bit set).
	for shift := 7; b&0x80 != 0; shift += 7 {
		offset++
		b = array[offset]
		value |= int(b&0x7F) << shift
	}
	return value
}

// Match kinds from FSATraversal.
const (
	ExactMatch          = 0
	NoMatch             = -1
	AutomatonHasPrefix  = -3
	SequenceIsAPrefix   = -4
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
