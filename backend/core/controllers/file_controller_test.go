package controllers

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"
)

func TestPrepareUploadPayloadCompressesLargeJPEG(t *testing.T) {
	original := buildLargeTestJPEG(t, 2600, 2100)
	srcFile := writeUploadTempFile(t, original)
	defer srcFile.Close()

	finalData, mimeType, err := prepareUploadPayload(srcFile)
	if err != nil {
		t.Fatalf("prepareUploadPayload returned error: %v", err)
	}

	if mimeType != "image/jpeg" {
		t.Fatalf("expected jpeg mime type, got %q", mimeType)
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(finalData))
	if err != nil {
		t.Fatalf("decode compressed payload: %v", err)
	}

	if format != "jpeg" {
		t.Fatalf("expected jpeg output, got %q", format)
	}

	if config.Width > maxUploadImageDimension || config.Height > maxUploadImageDimension {
		t.Fatalf("expected compressed image within %d px, got %dx%d", maxUploadImageDimension, config.Width, config.Height)
	}

	if len(finalData) >= len(original) {
		t.Fatalf("expected compressed payload smaller than original, original=%d compressed=%d", len(original), len(finalData))
	}
}

func buildLargeTestJPEG(t *testing.T, width, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8((x*7 + y*3) % 256),
				G: uint8((x*5 + y*11) % 256),
				B: uint8((x*13 + y*17) % 256),
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatalf("encode source jpeg: %v", err)
	}

	return buf.Bytes()
}

func writeUploadTempFile(t *testing.T, payload []byte) *os.File {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "upload-*.jpg")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	if _, err := file.Write(payload); err != nil {
		t.Fatalf("write temp upload: %v", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatalf("rewind temp upload: %v", err)
	}

	return file
}
