package diff

import (
	"testing"
)

// buildTestFrame creates a frame filled with a solid color
func buildTestFrame(w, h int, fill byte) *TestFrame {
	pixels := make([]byte, w*h*4)
	for i := range pixels {
		pixels[i] = fill
	}
	return &TestFrame{width: w, height: h, pixels: pixels, stride: w * 4}
}

type TestFrame struct {
	width, height, stride int
	pixels                []byte
}

func (f *TestFrame) Width() int        { return f.width }
func (f *TestFrame) Height() int       { return f.height }
func (f *TestFrame) Pixels() []byte    { return f.pixels }
func (f *TestFrame) Stride() int       { return f.stride }
func (f *TestFrame) PixelAt(x, y int) (r, g, b, a byte) {
	off := y*f.stride + x*4
	return f.pixels[off], f.pixels[off+1], f.pixels[off+2], f.pixels[off+3]
}

func TestFirstFrameAllDirty(t *testing.T) {
	d := NewDetector(16)
	frame := buildTestFrame(64, 64, 0xFF)
	rects := d.Detect(frame)

	// First frame: everything should be dirty
	if len(rects) == 0 {
		t.Fatal("first frame should produce dirty rects")
	}
	// Should have merged into a single rect covering the whole frame
	r := rects[0]
	if r.X != 0 || r.Y != 0 || r.W != 64 || r.H != 64 {
		t.Fatalf("expected full-frame rect, got (%d,%d,%d,%d)", r.X, r.Y, r.W, r.H)
	}
}

func TestIdenticalFrameNoDirty(t *testing.T) {
	d := NewDetector(16)
	frame := buildTestFrame(64, 64, 0xFF)
	d.Detect(frame) // first frame

	rects := d.Detect(frame) // same frame again
	if len(rects) != 0 {
		t.Fatalf("expected no dirty rects for identical frame, got %d", len(rects))
	}
}

func TestPartialChange(t *testing.T) {
	d := NewDetector(16)
	frame1 := buildTestFrame(64, 64, 0x00)
	d.Detect(frame1)

	// Modify one block: change top-left 16x16 to 0xFF
	frame2 := buildTestFrame(64, 64, 0x00)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			off := y*frame2.stride + x*4
			for c := 0; c < 4; c++ {
				frame2.pixels[off+c] = 0xFF
			}
		}
	}
	rects := d.Detect(frame2)
	if len(rects) == 0 {
		t.Fatal("expected at least one dirty rect")
	}
	// Dirty rect should cover around (0,0) to (16,16)
	found := false
	for _, r := range rects {
		if r.X < 16 && r.Y < 16 && r.X+r.W >= 16 && r.Y+r.H >= 16 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no rect covering the changed block, rects: %+v", rects)
	}
}

func TestBlockSizeNotDivisible(t *testing.T) {
	// 50x50 frame with 16x16 blocks — edge blocks should be handled
	d := NewDetector(16)
	frame := buildTestFrame(50, 50, 0xFF)
	rects := d.Detect(frame)
	if len(rects) == 0 {
		t.Fatal("non-divisible frame should still produce results")
	}
}
