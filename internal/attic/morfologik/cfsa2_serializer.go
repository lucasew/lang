package morfologik

import (
	"bytes"
	"fmt"
	"sort"
)

// CFSA2Serializer ports morfologik.fsa.builders.CFSA2Serializer (2.2.0).
// Serializes any in-memory FSA (typically ConstantArc from FSABuilder) to CFSA2 bytes.
// LanguageTool MultiSpeller uses: new CFSA2Serializer().serialize(fsa, baos) without numbers.

const (
	cfsa2LabelIndexSize = 31 // CFSA2.LABEL_INDEX_SIZE
	cfsa2NoState        = -1

	// FSAFlags bits used by CFSA2Serializer (without NUMBERS)
	fsaFlagFlexible = 1 << 0
	fsaFlagStopbit  = 1 << 1
	fsaFlagNextbit  = 1 << 2
	// NUMBERS is cfsa2FlagNumbers (1<<8)
)

// CFSA2Serializer is the Go twin of morfologik.fsa.builders.CFSA2Serializer.
type CFSA2Serializer struct {
	withNumbers    bool
	offsets        map[int]int
	numbers        map[int]int
	labelsIndex    []byte
	labelsInvIndex [256]int
	scratch        [5]byte
}

// NewCFSA2Serializer creates a serializer (Java default: no numbers).
func NewCFSA2Serializer() *CFSA2Serializer {
	return &CFSA2Serializer{
		offsets: make(map[int]int),
	}
}

// WithNumbers enables right-language counts (Java withNumbers()).
func (s *CFSA2Serializer) WithNumbers() *CFSA2Serializer {
	s.withNumbers = true
	return s
}

// Serialize ports CFSA2Serializer.serialize(fsa, os) → bytes (Java returns OutputStream).
func (s *CFSA2Serializer) Serialize(fsa *FSA) ([]byte, error) {
	if s == nil {
		s = NewCFSA2Serializer()
	}
	if fsa == nil {
		return nil, fmt.Errorf("nil FSA")
	}
	s.offsets = make(map[int]int)
	s.computeLabelsIndex(fsa)
	if s.withNumbers {
		s.numbers = rightLanguageForAllStates(fsa)
	}
	linearized, err := s.linearize(fsa)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	// FSAHeader.write(os, CFSA2.VERSION)
	buf.WriteByte(fsaMagic0)
	buf.WriteByte(fsaMagic1)
	buf.WriteByte(fsaMagic2)
	buf.WriteByte(fsaMagic3)
	buf.WriteByte(versionCFSA2)

	// flags: FLEXIBLE | STOPBIT | NEXTBIT [+ NUMBERS]
	var sflags uint16 = fsaFlagFlexible | fsaFlagStopbit | fsaFlagNextbit
	if s.withNumbers {
		sflags |= cfsa2FlagNumbers
	}
	buf.WriteByte(byte(sflags >> 8))
	buf.WriteByte(byte(sflags))

	buf.WriteByte(byte(len(s.labelsIndex)))
	buf.Write(s.labelsIndex)

	size, err := s.emitNodes(fsa, &buf, linearized)
	if err != nil {
		return nil, err
	}
	if size != 0 {
		return nil, fmt.Errorf("CFSA2Serializer: size changed in final pass (%d)", size)
	}
	return buf.Bytes(), nil
}

// SerializeFSA is a convenience for NewCFSA2Serializer().Serialize(fsa).
func SerializeFSA(fsa *FSA) ([]byte, error) {
	return NewCFSA2Serializer().Serialize(fsa)
}

func (s *CFSA2Serializer) computeLabelsIndex(fsa *FSA) {
	countByValue := make([]int, 256)
	visitAllStates(fsa, func(state int) bool {
		for arc := fsa.firstArc(state); arc != 0; arc = fsa.nextArc(arc) {
			countByValue[fsa.arcLabel(arc)]++
		}
		return true
	})

	type lc struct{ label, count int }
	var pairs []lc
	for label := 0; label < 256; label++ {
		if countByValue[label] > 0 {
			pairs = append(pairs, lc{label, countByValue[label]})
		}
	}
	// Order by descending frequency, then increasing label (Java TreeSet comparator)
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].label < pairs[j].label
	})

	n := 1 + min(len(pairs), cfsa2LabelIndexSize)
	s.labelsIndex = make([]byte, n)
	s.labelsInvIndex = [256]int{}
	// Fill from end: highest-priority labels get highest index slots (Java loop)
	for i := n - 1; i > 0 && len(pairs) > 0; i-- {
		p := pairs[0]
		pairs = pairs[1:]
		s.labelsInvIndex[p.label] = i
		s.labelsIndex[i] = byte(p.label)
	}
}

func (s *CFSA2Serializer) linearize(fsa *FSA) ([]int, error) {
	inlinkCount := computeInlinkCount(fsa)
	var linearized []int

	maxStates := 0x7fffffff // Integer.MAX_VALUE
	minInlinkCount := 2
	states := computeFirstStates(inlinkCount, maxStates, minInlinkCount)

	// Initial linearize with empty fixed prefix
	serializedSize := s.linearizeAndCalculateOffsets(fsa, nil, &linearized)

	// Probe cuts of high-inlink states (Java: cut 25,50,...,150)
	cutAt := 0
	for cut := min(25, len(states)); cut <= min(150, len(states)); cut += 25 {
		sub := states[:cut]
		newSize := s.linearizeAndCalculateOffsets(fsa, sub, &linearized)
		if newSize >= serializedSize {
			break
		}
		cutAt = cut
		serializedSize = newSize
	}
	_ = s.linearizeAndCalculateOffsets(fsa, states[:cutAt], &linearized)
	return linearized, nil
}

func (s *CFSA2Serializer) linearizeAndCalculateOffsets(fsa *FSA, fixed []int, linearized *[]int) int {
	visited := map[int]bool{}
	*linearized = (*linearized)[:0]
	var stack []int // IntStack: push/pop LIFO

	for _, st := range fixed {
		s.linearizeState(fsa, &stack, linearized, visited, st)
	}

	// DFS from root
	stack = append(stack, fsa.RootNode())
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[node] {
			continue
		}
		s.linearizeState(fsa, &stack, linearized, visited, node)
	}

	// Initialize offsets to Integer.MAX_VALUE (Java int, not Go int64 max — v-int is 5 bytes).
	const maxOffset = 0x7fffffff // java.lang.Integer.MAX_VALUE
	s.offsets = make(map[int]int, len(*linearized))
	for _, st := range *linearized {
		s.offsets[st] = maxOffset
	}

	j := 0
	for {
		i, _ := s.emitNodes(fsa, nil, *linearized)
		if i <= 0 {
			break
		}
		j = i
	}
	return j
}

func (s *CFSA2Serializer) linearizeState(fsa *FSA, stack *[]int, linearized *[]int, visited map[int]bool, node int) {
	*linearized = append(*linearized, node)
	visited[node] = true
	for arc := fsa.firstArc(node); arc != 0; arc = fsa.nextArc(arc) {
		if !fsa.isArcTerminal(arc) {
			target := fsa.endNode(arc)
			if !visited[target] {
				*stack = append(*stack, target)
			}
		}
	}
}

// computeFirstStates ports CFSA2Serializer.computeFirstStates.
// Returns states so index 0 has highest inlink count among selected (Java PQ extract).
func computeFirstStates(inlinkCount map[int]int, maxStates, minInlinkCount int) []int {
	// Java PriorityQueue min-heap of IntIntHolder(inlink, state); keep top maxStates.
	type pair struct{ a, b int } // a=inlink, b=state
	var heap []pair
	less := func(i, j int) bool {
		if heap[i].a != heap[j].a {
			return heap[i].a < heap[j].a
		}
		return heap[i].b < heap[j].b
	}
	for state, cnt := range inlinkCount {
		if cnt <= minInlinkCount {
			continue
		}
		// add if room or better than current min
		if len(heap) < maxStates {
			heap = append(heap, pair{cnt, state})
			sort.Slice(heap, less)
		} else if len(heap) > 0 {
			// comparator.compare(scratch, peek) > 0 → scratch better than min
			min := heap[0]
			if cnt > min.a || (cnt == min.a && state > min.b) {
				heap[0] = pair{cnt, state}
				sort.Slice(heap, less)
			}
		}
		if len(heap) > maxStates {
			heap = heap[1:]
		}
	}
	// Java: remove min into states[--position] → [0]=max inlink
	sort.Slice(heap, less)
	states := make([]int, len(heap))
	for i, p := range heap {
		states[len(heap)-1-i] = p.b
	}
	return states
}

func computeInlinkCount(fsa *FSA) map[int]int {
	inlink := map[int]int{}
	visited := map[int]bool{}
	var stack []int
	stack = append(stack, fsa.RootNode())
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[node] {
			continue
		}
		visited[node] = true
		for arc := fsa.firstArc(node); arc != 0; arc = fsa.nextArc(arc) {
			if !fsa.isArcTerminal(arc) {
				target := fsa.endNode(arc)
				inlink[target]++
				if !visited[target] {
					stack = append(stack, target)
				}
			}
		}
	}
	return inlink
}

// emitNodes ports CFSA2Serializer.emitNodes. os nil → size-calculation pass.
// Returns total size if offsets changed, else 0.
func (s *CFSA2Serializer) emitNodes(fsa *FSA, os *bytes.Buffer, linearized []int) (int, error) {
	offset := 0

	// Epsilon state
	offset += s.emitNodeData(os, 0)
	root := fsa.RootNode()
	if root != 0 {
		offset += s.emitArc(os, cfsa2BitLastArc, '^', s.offsets[root])
	} else {
		offset += s.emitArc(os, cfsa2BitLastArc, '^', 0)
	}

	offsetsChanged := false
	max := len(linearized)
	for idx, state := range linearized {
		nextState := cfsa2NoState
		if idx+1 < max {
			nextState = linearized[idx+1]
		}
		if os == nil {
			if s.offsets[state] != offset {
				offsetsChanged = true
			}
			s.offsets[state] = offset
		} else {
			if s.offsets[state] != offset {
				return 0, fmt.Errorf("offset mismatch state %d: want %d got %d", state, s.offsets[state], offset)
			}
		}
		num := 0
		if s.withNumbers {
			num = s.numbers[state]
		}
		offset += s.emitNodeData(os, num)
		n, err := s.emitNodeArcs(fsa, os, state, nextState)
		if err != nil {
			return 0, err
		}
		offset += n
	}
	if offsetsChanged {
		return offset, nil
	}
	return 0, nil
}

func (s *CFSA2Serializer) emitNodeArcs(fsa *FSA, os *bytes.Buffer, state, nextState int) (int, error) {
	offset := 0
	for arc := fsa.firstArc(state); arc != 0; arc = fsa.nextArc(arc) {
		targetOffset := 0
		target := 0
		if fsa.isArcTerminal(arc) {
			target = 0
			targetOffset = 0
		} else {
			target = fsa.endNode(arc)
			targetOffset = s.offsets[target]
		}
		flags := 0
		if fsa.isArcFinal(arc) {
			flags |= cfsa2BitFinalArc
		}
		if fsa.nextArc(arc) == 0 {
			flags |= cfsa2BitLastArc
		}
		if targetOffset != 0 && target == nextState {
			flags |= cfsa2BitTargetNext
			targetOffset = 0
		}
		offset += s.emitArc(os, flags, fsa.arcLabel(arc), targetOffset)
	}
	return offset, nil
}

func (s *CFSA2Serializer) emitArc(os *bytes.Buffer, flags int, label byte, targetOffset int) int {
	length := 0
	labelIndex := s.labelsInvIndex[label]
	if labelIndex > 0 {
		if os != nil {
			os.WriteByte(byte(flags | labelIndex))
		}
		length++
	} else {
		if os != nil {
			os.WriteByte(byte(flags))
			os.WriteByte(label)
		}
		length += 2
	}
	if flags&cfsa2BitTargetNext == 0 {
		n := writeVInt(s.scratch[:], 0, targetOffset)
		if os != nil {
			os.Write(s.scratch[:n])
		}
		length += n
	}
	return length
}

func (s *CFSA2Serializer) emitNodeData(os *bytes.Buffer, number int) int {
	if !s.withNumbers {
		return 0
	}
	n := writeVInt(s.scratch[:], 0, number)
	if os != nil {
		os.Write(s.scratch[:n])
	}
	return n
}

// writeVInt ports CFSA2Serializer.writeVInt.
func writeVInt(array []byte, offset, value int) int {
	start := offset
	for value > 0x7F {
		array[offset] = byte(0x80 | (value & 0x7F))
		offset++
		value >>= 7
	}
	array[offset] = byte(value)
	offset++
	return offset - start
}

// visitAllStates ports FSA.visitAllStates → visitInPostOrder from root.
func visitAllStates(fsa *FSA, accept func(state int) bool) {
	if fsa == nil {
		return
	}
	visited := map[int]bool{}
	var walk func(node int) bool
	walk = func(node int) bool {
		if visited[node] {
			return true
		}
		visited[node] = true
		for arc := fsa.firstArc(node); arc != 0; arc = fsa.nextArc(arc) {
			if !fsa.isArcTerminal(arc) {
				if !walk(fsa.endNode(arc)) {
					return false
				}
			}
		}
		return accept(node)
	}
	root := fsa.RootNode()
	if root != 0 {
		walk(root)
	}
}

// rightLanguageForAllStates ports FSAUtils.rightLanguageForAllStates (numbers mode).
func rightLanguageForAllStates(fsa *FSA) map[int]int {
	// Count sequences from each node (right language size).
	out := map[int]int{}
	var count func(node int) int
	count = func(node int) int {
		if c, ok := out[node]; ok {
			return c
		}
		n := 0
		for arc := fsa.firstArc(node); arc != 0; arc = fsa.nextArc(arc) {
			if fsa.isArcFinal(arc) {
				n++
			}
			if !fsa.isArcTerminal(arc) {
				n += count(fsa.endNode(arc))
			}
		}
		out[node] = n
		return n
	}
	if fsa.RootNode() != 0 {
		count(fsa.RootNode())
	}
	// Also count any other nodes already discovered via post-order visit
	visitAllStates(fsa, func(state int) bool {
		count(state)
		return true
	})
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
