package morfologik

// HMatrix ports morfologik.speller.HMatrix — diagonal-band edit-distance matrix
// stored as a flat vector (Oflazer). Used by Speller.findRepl / ed / cuted.
type HMatrix struct {
	p            []int
	rowLength    int
	columnHeight int
	editDistance int
}

// NewHMatrix ports HMatrix(distance, maxLength).
// distance = max edit distance; maxLength = max word length (Java Speller.MAX_WORD_LENGTH=120).
func NewHMatrix(distance, maxLength int) *HMatrix {
	if distance < 0 {
		distance = 0
	}
	if maxLength < 0 {
		maxLength = 0
	}
	h := &HMatrix{
		rowLength:    maxLength + 2,
		columnHeight: 2*distance + 3,
		editDistance: distance,
	}
	h.p = make([]int, h.rowLength*h.columnHeight)
	h.init()
	return h
}

// init ports private HMatrix.init.
func (h *HMatrix) init() {
	if h == nil || len(h.p) == 0 {
		return
	}
	size := len(h.p)
	// Initialize edges of the diagonal band to distance + 1 (i.e. distance too big)
	for i := 0; i < h.rowLength-h.editDistance-1; i++ {
		h.p[i] = h.editDistance + 1 // H(distance + j, j) = distance + 1
		h.p[size-i-1] = h.editDistance + 1 // H(i, distance + i) = distance + 1
	}
	// Initialize items H(i,j) with at least one index equal to zero to |i - j|
	for j := 0; j < h.editDistance+2; j++ {
		h.p[j*h.rowLength] = h.editDistance + 1 - j // H(i=0..distance+1,0)=i
		h.p[(j+h.editDistance+1)*h.rowLength+j] = j // H(0,j=0..distance+1)=j
	}
}

// Reset ports HMatrix.reset — zero then re-init (Java Speller recreates issues).
func (h *HMatrix) Reset() {
	if h == nil {
		return
	}
	for i := range h.p {
		h.p[i] = 0
	}
	h.init()
}

// Get ports HMatrix.get(i, j) — item H[i][j] in the simulated diagonal band.
func (h *HMatrix) Get(i, j int) int {
	if h == nil {
		return 0
	}
	return h.p[(j-i+h.editDistance+1)*h.rowLength+j]
}

// Set ports HMatrix.set(i, j, val). Indices must be in-band (no bounds check, same as Java).
func (h *HMatrix) Set(i, j, val int) {
	if h == nil {
		return
	}
	h.p[(j-i+h.editDistance+1)*h.rowLength+j] = val
}

// EditDistance returns the max edit distance this matrix was built for.
func (h *HMatrix) EditDistance() int {
	if h == nil {
		return 0
	}
	return h.editDistance
}
