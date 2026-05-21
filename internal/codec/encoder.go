package codec

import (
	"bytes"
	"image"
	"image/png"
)

func EncodeRGBA(pixels []byte, w, h int) ([]byte, error) {
	img := &image.RGBA{
		Pix:    pixels,
		Stride: w * 4,
		Rect:   image.Rect(0, 0, w, h),
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
