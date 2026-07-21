package morfologik

import "sort"

// FSABuilder ports morfologik.fsa.builders.FSABuilder (constant-arc automaton).
// Input sequences must be added in lexicographic (unsigned byte) order.

const (
	// ConstantArcSizeFSA layout
	casTargetAddressSize = 4
	casFlagsSize         = 1
	casLabelSize         = 1
	casArcSize           = casFlagsSize + casLabelSize + casTargetAddressSize // 6
	casFlagsOffset       = 0
	casLabelOffset       = casFlagsSize
	casAddressOffset     = casLabelOffset + casLabelSize

	casBitArcFinal = 1 << 1
	casBitArcLast  = 1 << 0

	casTerminalState = 0
	casMaxLabels     = 256
	casBufferGrowth  = 256 * 1024 // smaller than Java 5MB for plain-text builders
)

// FSABuilder builds a ConstantArcSizeFSA (Java FSABuilder).
type FSABuilder struct {
	serialized     []byte
	size           int
	bufferGrowth   int
	activePath     []int
	activePathLen  int
	nextArcOffset  []int
	root           int
	epsilon        int
	hashSet        []int
	hashSize       int
	previous       []byte
	previousLength int
	reallocs       int
}

// NewFSABuilder creates a builder (Java FSABuilder()).
func NewFSABuilder() *FSABuilder {
	b := &FSABuilder{
		bufferGrowth: casBufferGrowth,
		hashSet:      make([]int, 2),
	}
	// Allocate epsilon state.
	b.epsilon = b.allocateState(1)
	b.serialized[b.epsilon+casFlagsOffset] |= casBitArcLast
	// Allocate root with initial empty set of output arcs.
	b.expandActivePath(1)
	b.root = b.activePath[0]
	return b
}

// Add appends a sequence; must be lexicographically ≥ previous (unsigned bytes).
// len may be 0 only for complete()'s final flush.
func (b *FSABuilder) Add(sequence []byte) {
	if b.serialized == nil {
		panic("automaton already built")
	}
	start, length := 0, len(sequence)
	if b.previous != nil && length > 0 {
		if compareBytes(b.previous[:b.previousLength], sequence) > 0 {
			panic("input must be sorted")
		}
	}
	b.setPrevious(sequence)

	commonPrefix := b.commonPrefix(sequence, start, length)
	b.expandActivePath(length)

	// Freeze states after the common prefix.
	for i := b.activePathLen - 1; i > commonPrefix; i-- {
		frozenState := b.freezeState(i)
		b.setArcTarget(b.nextArcOffset[i-1]-casArcSize, frozenState)
		b.nextArcOffset[i] = b.activePath[i]
	}

	// Create arcs to new suffix states.
	j := start + commonPrefix
	for i := commonPrefix + 1; i <= length; i++ {
		p := b.nextArcOffset[i-1]
		flags := byte(0)
		if i == length {
			flags = casBitArcFinal
		}
		b.serialized[p+casFlagsOffset] = flags
		b.serialized[p+casLabelOffset] = sequence[j]
		j++
		target := casTerminalState
		if i != length {
			target = b.activePath[i]
		}
		b.setArcTarget(p, target)
		b.nextArcOffset[i-1] = p + casArcSize
	}
	b.activePathLen = length
}

// Complete finalizes and returns an FSA in constant-arc format.
func (b *FSABuilder) Complete() *FSA {
	b.Add(nil) // empty sequence flush (Java add(new byte[0],0,0))

	if b.nextArcOffset[0]-b.activePath[0] == 0 {
		b.setArcTarget(b.epsilon, casTerminalState)
	} else {
		b.root = b.freezeState(0)
		b.setArcTarget(b.epsilon, b.root)
	}

	data := make([]byte, b.size)
	copy(data, b.serialized[:b.size])
	fsa := &FSA{
		version:       versionConstantArc,
		arcs:          data,
		constantEps:   b.epsilon,
		constantArc:   true,
	}
	b.serialized = nil
	b.hashSet = nil
	return fsa
}

// BuildFSAFromSortedBytes builds from sorted unique sequences (Java FSABuilder.build).
func BuildFSAFromSortedBytes(sequences [][]byte) *FSA {
	b := NewFSABuilder()
	for _, seq := range sequences {
		b.Add(seq)
	}
	return b.Complete()
}

// BuildFSAFromWords sorts UTF-8 words (unsigned byte order) and builds an FSA.
func BuildFSAFromWords(words []string) *FSA {
	// dedupe + sort
	seen := map[string]struct{}{}
	var uniq []string
	for _, w := range words {
		if w == "" {
			continue
		}
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		uniq = append(uniq, w)
	}
	// unsigned byte order = Go string compare for UTF-8? Java uses unsigned byte compare.
	// For pure ASCII same as strings; for UTF-8 multi-byte, C locale is byte-wise.
	sortWordsUnsigned(uniq)
	seqs := make([][]byte, len(uniq))
	for i, w := range uniq {
		seqs[i] = []byte(w)
	}
	return BuildFSAFromSortedBytes(seqs)
}

func sortWordsUnsigned(words []string) {
	sort.Slice(words, func(i, j int) bool {
		return compareBytes([]byte(words[i]), []byte(words[j])) < 0
	})
}

func compareBytes(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		c1, c2 := a[i], b[i]
		if c1 != c2 {
			return int(c1) - int(c2) // unsigned already for byte in Go when converted to int 0-255
		}
	}
	return len(a) - len(b)
}

func (b *FSABuilder) setPrevious(sequence []byte) {
	if cap(b.previous) < len(sequence) {
		b.previous = make([]byte, len(sequence))
	}
	b.previous = b.previous[:len(sequence)]
	copy(b.previous, sequence)
	b.previousLength = len(sequence)
}

func (b *FSABuilder) commonPrefix(sequence []byte, start, length int) int {
	max := length
	if b.activePathLen < max {
		max = b.activePathLen
	}
	i := 0
	for i < max {
		lastArc := b.nextArcOffset[i] - casArcSize
		if sequence[start+i] != b.getArcLabel(lastArc) {
			break
		}
		i++
	}
	return i
}

func (b *FSABuilder) freezeState(activePathIndex int) int {
	start := b.activePath[activePathIndex]
	end := b.nextArcOffset[activePathIndex]
	length := end - start

	// last arc flag
	b.serialized[end-casArcSize+casFlagsOffset] |= casBitArcLast

	bucketMask := len(b.hashSet) - 1
	slot := b.hash(start, length) & bucketMask
	for i := 0; ; i++ {
		state := b.hashSet[slot]
		if state == 0 {
			state = b.serialize(activePathIndex)
			b.hashSet[slot] = state
			b.hashSize++
			if b.hashSize > len(b.hashSet)/2 {
				b.expandAndRehash()
			}
			return state
		} else if b.equivalent(state, start, length) {
			return state
		}
		slot = (slot + i + 1) & bucketMask
	}
}

func (b *FSABuilder) expandAndRehash() {
	newHash := make([]int, len(b.hashSet)*2)
	bucketMask := len(newHash) - 1
	for _, state := range b.hashSet {
		if state > 0 {
			slot := b.hash(state, b.stateLength(state)) & bucketMask
			for i := 0; newHash[slot] > 0; i++ {
				slot = (slot + i + 1) & bucketMask
			}
			newHash[slot] = state
		}
	}
	b.hashSet = newHash
}

func (b *FSABuilder) stateLength(state int) int {
	arc := state
	for !b.isArcLast(arc) {
		arc += casArcSize
	}
	return arc - state + casArcSize
}

func (b *FSABuilder) equivalent(start1, start2, length int) bool {
	if start1+length > b.size || start2+length > b.size {
		return false
	}
	for length > 0 {
		if b.serialized[start1] != b.serialized[start2] {
			return false
		}
		start1++
		start2++
		length--
	}
	return true
}

func (b *FSABuilder) serialize(activePathIndex int) int {
	b.expandBuffers()
	newState := b.size
	start := b.activePath[activePathIndex]
	length := b.nextArcOffset[activePathIndex] - start
	copy(b.serialized[newState:newState+length], b.serialized[start:start+length])
	b.size += length
	return newState
}

func (b *FSABuilder) hash(start, byteCount int) int {
	h := 0
	for arcs := byteCount / casArcSize; arcs > 0; arcs-- {
		h = 17*h + int(b.getArcLabel(start))
		h = 17*h + b.getArcTarget(start)
		if b.isArcFinal(start) {
			h += 17
		}
		start += casArcSize
	}
	return h
}

func (b *FSABuilder) expandActivePath(size int) {
	if len(b.activePath) < size {
		p := len(b.activePath)
		na := make([]int, size)
		nn := make([]int, size)
		copy(na, b.activePath)
		copy(nn, b.nextArcOffset)
		b.activePath = na
		b.nextArcOffset = nn
		for i := p; i < size; i++ {
			st := b.allocateState(casMaxLabels)
			b.activePath[i] = st
			b.nextArcOffset[i] = st
		}
	}
}

func (b *FSABuilder) expandBuffers() {
	need := b.size + casArcSize*casMaxLabels
	if len(b.serialized) < need {
		n := len(b.serialized) + b.bufferGrowth
		if n < need {
			n = need
		}
		ns := make([]byte, n)
		copy(ns, b.serialized)
		b.serialized = ns
		b.reallocs++
	}
}

func (b *FSABuilder) allocateState(labels int) int {
	b.expandBuffers()
	state := b.size
	b.size += labels * casArcSize
	return state
}

func (b *FSABuilder) setArcTarget(arc, state int) {
	// Java writes big-endian into 4 bytes ending at ADDRESS_OFFSET+4
	pos := arc + casAddressOffset + casTargetAddressSize
	for i := 0; i < casTargetAddressSize; i++ {
		pos--
		b.serialized[pos] = byte(state)
		state >>= 8
	}
}

func (b *FSABuilder) getArcTarget(arc int) int {
	arc += casAddressOffset
	return int(b.serialized[arc])<<24 |
		int(b.serialized[arc+1]&0xff)<<16 |
		int(b.serialized[arc+2]&0xff)<<8 |
		int(b.serialized[arc+3]&0xff)
}

func (b *FSABuilder) getArcLabel(arc int) byte {
	return b.serialized[arc+casLabelOffset]
}

func (b *FSABuilder) isArcFinal(arc int) bool {
	return b.serialized[arc+casFlagsOffset]&casBitArcFinal != 0
}

func (b *FSABuilder) isArcLast(arc int) bool {
	return b.serialized[arc+casFlagsOffset]&casBitArcLast != 0
}
