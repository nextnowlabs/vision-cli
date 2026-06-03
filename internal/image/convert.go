package imgconv

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
)

func ToRGBPNG(data []byte) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := src.Bounds()
	rgb := image.NewRGBA(bounds)
	draw.Draw(rgb, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(rgb, bounds, src, bounds.Min, draw.Over)

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgb); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
