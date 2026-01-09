package fpkgen

import (
	"testing"
	"testing/quick"
)

// **Feature: icon-format-support, Property 1: Format Detection Correctness**
// **Validates: Requirements 4.1, 4.3**
//
// For any valid image data of a supported format (PNG, JPEG, WebP, BMP, ICO),
// the detectFormat function SHALL correctly identify the format based on magic bytes,
// regardless of file extension or naming.

// TestDetectFormat_PNG tests PNG format detection
func TestDetectFormat_PNG(t *testing.T) {
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	if got := detectFormat(pngData); got != FormatPNG {
		t.Errorf("detectFormat(PNG data) = %v, want %v", got, FormatPNG)
	}
}

// TestDetectFormat_JPEG tests JPEG format detection
func TestDetectFormat_JPEG(t *testing.T) {
	// JPEG magic bytes: FF D8 FF
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	if got := detectFormat(jpegData); got != FormatJPEG {
		t.Errorf("detectFormat(JPEG data) = %v, want %v", got, FormatJPEG)
	}
}

// TestDetectFormat_WebP tests WebP format detection
func TestDetectFormat_WebP(t *testing.T) {
	// WebP magic bytes: RIFF....WEBP
	webpData := []byte{
		0x52, 0x49, 0x46, 0x46, // RIFF
		0x00, 0x00, 0x00, 0x00, // file size (placeholder)
		0x57, 0x45, 0x42, 0x50, // WEBP
	}
	if got := detectFormat(webpData); got != FormatWebP {
		t.Errorf("detectFormat(WebP data) = %v, want %v", got, FormatWebP)
	}
}

// TestDetectFormat_BMP tests BMP format detection
func TestDetectFormat_BMP(t *testing.T) {
	// BMP magic bytes: 42 4D (BM)
	bmpData := []byte{0x42, 0x4D, 0x00, 0x00, 0x00, 0x00}
	if got := detectFormat(bmpData); got != FormatBMP {
		t.Errorf("detectFormat(BMP data) = %v, want %v", got, FormatBMP)
	}
}

// TestDetectFormat_ICO tests ICO format detection
func TestDetectFormat_ICO(t *testing.T) {
	// ICO magic bytes: 00 00 01 00
	icoData := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00}
	if got := detectFormat(icoData); got != FormatICO {
		t.Errorf("detectFormat(ICO data) = %v, want %v", got, FormatICO)
	}
}

// TestDetectFormat_Unknown tests unknown format detection
func TestDetectFormat_Unknown(t *testing.T) {
	// Random data that doesn't match any format
	unknownData := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}
	if got := detectFormat(unknownData); got != FormatUnknown {
		t.Errorf("detectFormat(unknown data) = %v, want %v", got, FormatUnknown)
	}
}

// TestDetectFormat_TooShort tests handling of data too short to detect
func TestDetectFormat_TooShort(t *testing.T) {
	shortData := []byte{0x89} // Only 1 byte (minimum is 2)
	if got := detectFormat(shortData); got != FormatUnknown {
		t.Errorf("detectFormat(short data) = %v, want %v", got, FormatUnknown)
	}
}

// TestDetectFormat_Empty tests handling of empty data
func TestDetectFormat_Empty(t *testing.T) {
	if got := detectFormat([]byte{}); got != FormatUnknown {
		t.Errorf("detectFormat(empty) = %v, want %v", got, FormatUnknown)
	}
}

// Property test: PNG format detection is consistent
// For any byte slice with PNG magic bytes prefix, detectFormat returns FormatPNG
func TestProperty_PNGDetection(t *testing.T) {
	f := func(suffix []byte) bool {
		// Construct data with PNG magic bytes followed by arbitrary suffix
		pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		data := append(pngMagic, suffix...)
		return detectFormat(data) == FormatPNG
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: JPEG format detection is consistent
// For any byte slice with JPEG magic bytes prefix, detectFormat returns FormatJPEG
func TestProperty_JPEGDetection(t *testing.T) {
	f := func(suffix []byte) bool {
		// Construct data with JPEG magic bytes followed by arbitrary suffix
		jpegMagic := []byte{0xFF, 0xD8, 0xFF}
		data := append(jpegMagic, suffix...)
		return detectFormat(data) == FormatJPEG
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: BMP format detection is consistent
// For any byte slice with BMP magic bytes prefix, detectFormat returns FormatBMP
func TestProperty_BMPDetection(t *testing.T) {
	f := func(suffix []byte) bool {
		// Construct data with BMP magic bytes followed by arbitrary suffix
		bmpMagic := []byte{0x42, 0x4D}
		data := append(bmpMagic, suffix...)
		return detectFormat(data) == FormatBMP
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: ICO format detection is consistent
// For any byte slice with ICO magic bytes prefix, detectFormat returns FormatICO
func TestProperty_ICODetection(t *testing.T) {
	f := func(suffix []byte) bool {
		// Construct data with ICO magic bytes followed by arbitrary suffix
		icoMagic := []byte{0x00, 0x00, 0x01, 0x00}
		data := append(icoMagic, suffix...)
		return detectFormat(data) == FormatICO
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: WebP format detection is consistent
// For any byte slice with WebP magic bytes, detectFormat returns FormatWebP
func TestProperty_WebPDetection(t *testing.T) {
	f := func(fileSize uint32, suffix []byte) bool {
		// Construct data with WebP magic bytes (RIFF....WEBP) followed by arbitrary suffix
		webpHeader := []byte{
			0x52, 0x49, 0x46, 0x46, // RIFF
			byte(fileSize), byte(fileSize >> 8), byte(fileSize >> 16), byte(fileSize >> 24), // file size
			0x57, 0x45, 0x42, 0x50, // WEBP
		}
		data := append(webpHeader, suffix...)
		return detectFormat(data) == FormatWebP
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: Format detection is deterministic
// For any byte slice, calling detectFormat twice returns the same result
func TestProperty_Deterministic(t *testing.T) {
	f := func(data []byte) bool {
		result1 := detectFormat(data)
		result2 := detectFormat(data)
		return result1 == result2
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// TestImageFormat_String tests the String method of ImageFormat
func TestImageFormat_String(t *testing.T) {
	tests := []struct {
		format   ImageFormat
		expected string
	}{
		{FormatUnknown, "Unknown"},
		{FormatPNG, "PNG"},
		{FormatJPEG, "JPEG"},
		{FormatWebP, "WebP"},
		{FormatBMP, "BMP"},
		{FormatICO, "ICO"},
		{ImageFormat(100), "Unknown"}, // Out of range
	}

	for _, tt := range tests {
		if got := tt.format.String(); got != tt.expected {
			t.Errorf("ImageFormat(%d).String() = %q, want %q", tt.format, got, tt.expected)
		}
	}
}
