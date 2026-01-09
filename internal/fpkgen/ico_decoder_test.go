package fpkgen

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/quick"
)

// **Feature: icon-format-support, Property 2: ICO Highest Resolution Selection**
// **Validates: Requirements 1.4, 2.2**
//
// For any valid ICO file containing multiple images of different resolutions,
// the decodeICO function SHALL return the image with the largest pixel dimensions (width × height).

// generateICOWithPNGImages creates a valid ICO file with multiple PNG images of different sizes
func generateICOWithPNGImages(sizes []int) ([]byte, int) {
	if len(sizes) == 0 {
		return nil, 0
	}

	// Find the largest size
	maxSize := 0
	for _, size := range sizes {
		if size > maxSize {
			maxSize = size
		}
	}

	// Generate PNG data for each size
	pngDataList := make([][]byte, len(sizes))
	for i, size := range sizes {
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		// Fill with a color based on size for identification
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				img.SetRGBA(x, y, color.RGBA{
					R: uint8(size % 256),
					G: uint8((size / 2) % 256),
					B: uint8((size / 4) % 256),
					A: 255,
				})
			}
		}
		var buf bytes.Buffer
		png.Encode(&buf, img)
		pngDataList[i] = buf.Bytes()
	}

	// Calculate total size
	headerSize := 6
	directorySize := 16 * len(sizes)
	totalImageSize := 0
	for _, data := range pngDataList {
		totalImageSize += len(data)
	}

	// Create ICO file
	icoData := make([]byte, headerSize+directorySize+totalImageSize)

	// Write header
	binary.LittleEndian.PutUint16(icoData[0:2], 0)                  // Reserved
	binary.LittleEndian.PutUint16(icoData[2:4], 1)                  // Type (1 = ICO)
	binary.LittleEndian.PutUint16(icoData[4:6], uint16(len(sizes))) // Count

	// Write directory entries and image data
	offset := uint32(headerSize + directorySize)
	for i, size := range sizes {
		entryOffset := headerSize + i*16
		pngData := pngDataList[i]

		// Width and Height (0 means 256)
		width := uint8(size)
		height := uint8(size)
		if size == 256 {
			width = 0
			height = 0
		}

		icoData[entryOffset] = width
		icoData[entryOffset+1] = height
		icoData[entryOffset+2] = 0 // ColorCount
		icoData[entryOffset+3] = 0 // Reserved
		binary.LittleEndian.PutUint16(icoData[entryOffset+4:entryOffset+6], 1)                  // Planes
		binary.LittleEndian.PutUint16(icoData[entryOffset+6:entryOffset+8], 32)                 // BitCount
		binary.LittleEndian.PutUint32(icoData[entryOffset+8:entryOffset+12], uint32(len(pngData))) // Size
		binary.LittleEndian.PutUint32(icoData[entryOffset+12:entryOffset+16], offset)           // Offset

		// Copy PNG data
		copy(icoData[offset:], pngData)
		offset += uint32(len(pngData))
	}

	return icoData, maxSize
}


// TestProperty_ICOHighestResolutionSelection tests that decodeICO always returns the highest resolution image
// Property 2: For any valid ICO file containing multiple images of different resolutions,
// the decodeICO function SHALL return the image with the largest pixel dimensions (width × height).
func TestProperty_ICOHighestResolutionSelection(t *testing.T) {
	f := func(sizesInput []uint8) bool {
		// Filter to valid sizes (1-255, we'll also test 256 separately)
		var sizes []int
		for _, s := range sizesInput {
			if s > 0 && s <= 128 { // Limit to 128 to keep test fast
				sizes = append(sizes, int(s))
			}
		}

		// Need at least 2 different sizes to test the property
		if len(sizes) < 2 {
			return true // Skip trivial cases
		}

		// Ensure we have at least 2 unique sizes
		uniqueSizes := make(map[int]bool)
		for _, s := range sizes {
			uniqueSizes[s] = true
		}
		if len(uniqueSizes) < 2 {
			return true // Skip if all sizes are the same
		}

		// Limit to 5 images to keep test fast
		if len(sizes) > 5 {
			sizes = sizes[:5]
		}

		icoData, expectedMaxSize := generateICOWithPNGImages(sizes)
		if icoData == nil {
			return true // Skip invalid input
		}

		img, err := decodeICO(icoData)
		if err != nil {
			t.Logf("decodeICO failed: %v", err)
			return false
		}

		bounds := img.Bounds()
		actualWidth := bounds.Dx()
		actualHeight := bounds.Dy()

		// The returned image should have the maximum size
		if actualWidth != expectedMaxSize || actualHeight != expectedMaxSize {
			t.Logf("Expected size %dx%d, got %dx%d (sizes: %v)", 
				expectedMaxSize, expectedMaxSize, actualWidth, actualHeight, sizes)
			return false
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// TestDecodeICO_SingleImage tests decoding ICO with a single image
func TestDecodeICO_SingleImage(t *testing.T) {
	sizes := []int{64}
	icoData, _ := generateICOWithPNGImages(sizes)

	img, err := decodeICO(icoData)
	if err != nil {
		t.Fatalf("decodeICO failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("Expected 64x64, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestDecodeICO_MultipleImages tests decoding ICO with multiple images
func TestDecodeICO_MultipleImages(t *testing.T) {
	sizes := []int{16, 32, 48, 64, 128}
	icoData, _ := generateICOWithPNGImages(sizes)

	img, err := decodeICO(icoData)
	if err != nil {
		t.Fatalf("decodeICO failed: %v", err)
	}

	bounds := img.Bounds()
	// Should return the largest (128x128)
	if bounds.Dx() != 128 || bounds.Dy() != 128 {
		t.Errorf("Expected 128x128 (largest), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestDecodeICO_InvalidHeader tests error handling for invalid ICO header
func TestDecodeICO_InvalidHeader(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"too short", []byte{0x00, 0x00}},
		{"invalid reserved", []byte{0x01, 0x00, 0x01, 0x00, 0x01, 0x00}},
		{"invalid type", []byte{0x00, 0x00, 0x02, 0x00, 0x01, 0x00}}, // Type 2 is CUR, not ICO
		{"zero count", []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeICO(tt.data)
			if err == nil {
				t.Error("expected error for invalid ICO header")
			}
		})
	}
}

// TestDecodeICO_TruncatedDirectory tests error handling for truncated directory
func TestDecodeICO_TruncatedDirectory(t *testing.T) {
	// Header says 1 image, but no directory entry
	data := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10} // Only 7 bytes, need 22 for header + 1 entry
	_, err := decodeICO(data)
	if err == nil {
		t.Error("expected error for truncated directory")
	}
}

// TestIcoEntry_GetActualDimensions tests the dimension conversion methods
func TestIcoEntry_GetActualDimensions(t *testing.T) {
	tests := []struct {
		width, height       uint8
		expectedW, expectedH int
	}{
		{0, 0, 256, 256},     // 0 means 256
		{16, 16, 16, 16},
		{32, 32, 32, 32},
		{128, 128, 128, 128},
		{255, 255, 255, 255},
	}

	for _, tt := range tests {
		entry := &icoEntry{Width: tt.width, Height: tt.height}
		if got := entry.getActualWidth(); got != tt.expectedW {
			t.Errorf("getActualWidth() for %d = %d, want %d", tt.width, got, tt.expectedW)
		}
		if got := entry.getActualHeight(); got != tt.expectedH {
			t.Errorf("getActualHeight() for %d = %d, want %d", tt.height, got, tt.expectedH)
		}
	}
}

// TestIcoEntry_Resolution tests the resolution calculation
func TestIcoEntry_Resolution(t *testing.T) {
	tests := []struct {
		width, height uint8
		expected      int
	}{
		{0, 0, 256 * 256},     // 0 means 256
		{16, 16, 16 * 16},
		{32, 32, 32 * 32},
		{64, 128, 64 * 128},
	}

	for _, tt := range tests {
		entry := &icoEntry{Width: tt.width, Height: tt.height}
		if got := entry.resolution(); got != tt.expected {
			t.Errorf("resolution() for %dx%d = %d, want %d", tt.width, tt.height, got, tt.expected)
		}
	}
}
