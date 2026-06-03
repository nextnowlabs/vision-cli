package imgconv_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/nextnowlabs/vision-cli/internal/image"
)

func newRGBAImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	return img
}

func TestToRGBPNG_RGBA(t *testing.T) {
	src := newRGBAImage()
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}

	result, err := imgconv.ToRGBPNG(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	decoded, _, err := image.Decode(bytes.NewReader(result))
	if err != nil {
		t.Fatal(err)
	}

	rgba, ok := decoded.(*image.RGBA)
	if !ok {
		t.Fatalf("expected *image.RGBA, got %T", decoded)
	}

	bounds := rgba.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Fatalf("expected 2x2, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	i := rgba.PixOffset(0, 0)
	r0, g0, b0 := rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2]
	t.Logf("pixel (0,0): R=%d G=%d B=%d", r0, g0, b0)
	if !(r0 == 254 && g0 == 127 && b0 == 127) {
		t.Errorf("pixel (0,0): RGBA{255,0,0,128} on white, got R=%d G=%d B=%d, want (254,127,127)", r0, g0, b0)
	}

	j := rgba.PixOffset(1, 0)
	r1, g1, b1 := rgba.Pix[j], rgba.Pix[j+1], rgba.Pix[j+2]
	t.Logf("pixel (1,0): R=%d G=%d B=%d", r1, g1, b1)
	if !(r1 == 0 && g1 == 255 && b1 == 0) {
		t.Errorf("pixel (1,0): RGBA{0,255,0,255} on white, got R=%d G=%d B=%d, want (0,255,0)", r1, g1, b1)
	}
}

func TestToRGBPNG_JPEG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatal(err)
	}

	result, err := imgconv.ToRGBPNG(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	decoded, _, err := image.Decode(bytes.NewReader(result))
	if err != nil {
		t.Fatal(err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Fatalf("expected 2x2, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestToRGBPNG_RGBPNG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	src.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}

	result, err := imgconv.ToRGBPNG(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	decoded, _, err := image.Decode(bytes.NewReader(result))
	if err != nil {
		t.Fatal(err)
	}

	rgba, ok := decoded.(*image.RGBA)
	if !ok {
		t.Fatalf("expected *image.RGBA, got %T", decoded)
	}

	bounds := rgba.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Fatalf("expected 2x2, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	i := rgba.PixOffset(0, 0)
	r0, g0, b0 := rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2]
	if !(r0 == 255 && g0 == 0 && b0 == 0) {
		t.Errorf("pixel (0,0): RGBA{255,0,0,255} on white, got R=%d G=%d B=%d, want (255,0,0)", r0, g0, b0)
	}

	j := rgba.PixOffset(1, 0)
	r1, g1, b1 := rgba.Pix[j], rgba.Pix[j+1], rgba.Pix[j+2]
	if !(r1 == 0 && g1 == 255 && b1 == 0) {
		t.Errorf("pixel (1,0): RGBA{0,255,0,255} on white, got R=%d G=%d B=%d, want (0,255,0)", r1, g1, b1)
	}
}

func TestToRGBPNG_BadInput(t *testing.T) {
	_, err := imgconv.ToRGBPNG([]byte("not an image"))
	if err == nil {
		t.Fatal("expected error for bad input")
	}
}
