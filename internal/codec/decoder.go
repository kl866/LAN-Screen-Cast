package codec

import (
	"bytes"
	"image"
	"image/png"
)

type DecodedImage struct {
	W, H   int
	Pixels []byte // RGBA
}

func DecodeToRGBA(data []byte) (*DecodedImage, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	rgba, ok := img.(*image.RGBA)
	if ok {
		return &DecodedImage{W: w, H: h, Pixels: cloneBytes(rgba.Pix)}, nil
	}

	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, img.At(x, y))
		}
	}
	return &DecodedImage{W: w, H: h, Pixels: dst.Pix}, nil
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
