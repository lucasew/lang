package morfologik

import "testing"

// Twin of morfologik-speller/.../HMatrixTest.java

const maxWordLength = 120 // Speller.MAX_WORD_LENGTH

// Port of HMatrixTest.stressTestInit
func TestHMatrix_StressTestInit(t *testing.T) {
	for i := 0; i < 10; i++ {
		h := NewHMatrix(i, maxWordLength)
		if got := h.Get(1, 1); got != 0 {
			t.Fatalf("distance=%d H(1,1)=%d want 0", i, got)
		}
	}
}

func TestHMatrix_BandBasics(t *testing.T) {
	// distance 1: H(0,0)=0, H(1,0)=1 (along edge init)
	h := NewHMatrix(1, maxWordLength)
	if h.Get(0, 0) != 0 {
		t.Fatalf("H(0,0)=%d", h.Get(0, 0))
	}
	// After set/get round-trip
	h.Set(1, 1, 5)
	if h.Get(1, 1) != 5 {
		t.Fatalf("after set H(1,1)=%d", h.Get(1, 1))
	}
	h.Reset()
	if h.Get(1, 1) != 0 {
		t.Fatalf("after reset H(1,1)=%d", h.Get(1, 1))
	}
}

func TestHMatrix_EditDistanceField(t *testing.T) {
	h := NewHMatrix(3, 50)
	if h.EditDistance() != 3 {
		t.Fatalf("editDistance=%d", h.EditDistance())
	}
}
