package fpkgen

import "bytes"

// ImageFormat represents supported image formats
type ImageFormat int

const (
	FormatUnknown ImageFormat = iota
	FormatPNG
	FormatJPEG
	FormatWebP
	FormatBMP
	FormatICO
)

// String returns the string representation of the image format
func (f ImageFormat) String() string {
	names := []string{"Unknown", "PNG", "JPEG", "WebP", "BMP", "ICO"}
	if int(f) < len(names) {
		return names[f]
	}
	return "Unknown"
}

// Magic bytes for format detection
var (
	magicPNG  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG signature
	magicJPEG = []byte{0xFF, 0xD8, 0xFF}                                // JPEG SOI marker
	magicBMP  = []byte{0x42, 0x4D}                                      // "BM"
	magicICO  = []byte{0x00, 0x00, 0x01, 0x00}                          // ICO header
	magicRIFF = []byte{0x52, 0x49, 0x46, 0x46}                          // "RIFF" for WebP
	magicWEBP = []byte{0x57, 0x45, 0x42, 0x50}                          // "WEBP" at offset 8
)

// detectFormat detects the image format by examining magic bytes in the data.
// It does not rely on file extension, only on the actual content.
func detectFormat(data []byte) ImageFormat {
	// Need at least 2 bytes for the shortest magic (BMP)
	if len(data) < 2 {
		return FormatUnknown
	}

	// Check PNG (8 bytes signature)
	if len(data) >= 8 && bytes.HasPrefix(data, magicPNG) {
		return FormatPNG
	}

	// Check JPEG (starts with FF D8 FF)
	if len(data) >= 3 && bytes.HasPrefix(data, magicJPEG) {
		return FormatJPEG
	}

	// Check WebP (RIFF....WEBP format)
	// WebP files start with "RIFF" followed by 4 bytes of file size, then "WEBP"
	if len(data) >= 12 && bytes.HasPrefix(data, magicRIFF) && bytes.Equal(data[8:12], magicWEBP) {
		return FormatWebP
	}

	// Check ICO (starts with 00 00 01 00)
	if len(data) >= 4 && bytes.HasPrefix(data, magicICO) {
		return FormatICO
	}

	// Check BMP (starts with "BM") - only 2 bytes needed
	if bytes.HasPrefix(data, magicBMP) {
		return FormatBMP
	}

	return FormatUnknown
}
