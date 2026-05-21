package diff

import (
	"github.com/cespare/xxhash/v2"
)

const DefaultBlockSize = 16

type FrameParser interface {
	Width() int
	Height() int
	Pixels() []byte
	Stride() int
}

type DirtyRect struct {
	X, Y, W, H int
	Pixels     []byte
}

type Detector struct {
	blockSize   int
	cols        int
	rows        int
	prevHashes  []uint64
	initialized bool
}

func NewDetector(blockSize int) *Detector {
	return &Detector{blockSize: blockSize}
}

func (d *Detector) Detect(frame FrameParser) []DirtyRect {
	fw, fh := frame.Width(), frame.Height()
	stride := frame.Stride()
	pixels := frame.Pixels()

	cols := (fw + d.blockSize - 1) / d.blockSize
	rows := (fh + d.blockSize - 1) / d.blockSize

	if !d.initialized || cols != d.cols || rows != d.rows {
		d.cols = cols
		d.rows = rows
		d.prevHashes = make([]uint64, cols*rows)
		d.initialized = true
	}

	// Build a boolean grid of dirty blocks
	dirtyGrid := make([]bool, cols*rows)
	hashes := make([]uint64, cols*rows)
	dirtyCount := 0

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			h := d.blockHash(pixels, stride, fw, fh, col, row)
			hashes[idx] = h
			if h != d.prevHashes[idx] {
				dirtyGrid[idx] = true
				dirtyCount++
			}
		}
	}

	d.prevHashes = hashes

	if dirtyCount == 0 {
		return nil
	}

	// Extract dirty rects by scanning the grid
	rects := d.extractRects(dirtyGrid, cols, rows, frame)
	return rects
}

func (d *Detector) blockHash(pixels []byte, stride, fw, fh, col, row int) uint64 {
	x0 := col * d.blockSize
	y0 := row * d.blockSize
	x1 := x0 + d.blockSize
	y1 := y0 + d.blockSize
	if x1 > fw {
		x1 = fw
	}
	if y1 > fh {
		y1 = fh
	}

	dg := xxhash.New()
	// Sample every 4th pixel for speed
	for y := y0; y < y1; y += 4 {
		for x := x0; x < x1; x += 4 {
			off := y*stride + x*4
			if off+4 <= len(pixels) {
				dg.Write(pixels[off : off+4])
			}
		}
	}
	return dg.Sum64()
}

func (d *Detector) extractRects(grid []bool, cols, rows int, frame FrameParser) []DirtyRect {
	var rects []DirtyRect
	visited := make([]bool, cols*rows)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if !grid[idx] || visited[idx] {
				continue
			}
			// Find horizontal run
			endCol := col
			for endCol+1 < cols && grid[row*cols+endCol+1] && !visited[row*cols+endCol+1] {
				endCol++
			}
			// Try to extend vertically
			endRow := row
			for endRow+1 < rows {
				canExtend := true
				for c := col; c <= endCol; c++ {
					nidx := (endRow+1)*cols + c
					if !grid[nidx] || visited[nidx] {
						canExtend = false
						break
					}
				}
				if !canExtend {
					break
				}
				endRow++
			}
			// Mark visited
			for r := row; r <= endRow; r++ {
				for c := col; c <= endCol; c++ {
					visited[r*cols+c] = true
				}
			}
			// Compute pixel coordinates
			x := col * d.blockSize
			y := row * d.blockSize
			w := (endCol+1)*d.blockSize - x
			h := (endRow+1)*d.blockSize - y
			if x+w > frame.Width() {
				w = frame.Width() - x
			}
			if y+h > frame.Height() {
				h = frame.Height() - y
			}
			// Extract pixels
			pix := extractPixels(frame, x, y, w, h)
			rects = append(rects, DirtyRect{X: x, Y: y, W: w, H: h, Pixels: pix})
		}
	}
	return rects
}

func extractPixels(f FrameParser, x, y, w, h int) []byte {
	stride := f.Stride()
	src := f.Pixels()
	dst := make([]byte, w*h*4)
	for row := 0; row < h; row++ {
		srcOff := (y+row)*stride + x*4
		dstOff := row * w * 4
		copy(dst[dstOff:dstOff+w*4], src[srcOff:srcOff+w*4])
	}
	return dst
}
