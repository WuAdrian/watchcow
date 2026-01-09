//go:build ignore

// This file generates test data files for the fpkgen package tests.
// Run with: go run generate_testdata.go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"

	"golang.org/x/image/bmp"
)

func main() {
	// Create a simple 32x32 test image
	img := createTestImage(32, 32, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Generate PNG
	if err := savePNG("test.png", img); err != nil {
		fmt.Printf("Failed to create test.png: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created test.png")

	// Generate JPEG
	if err := saveJPEG("test.jpg", img); err != nil {
		fmt.Printf("Failed to create test.jpg: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created test.jpg")

	// Generate BMP
	if err := saveBMP("test.bmp", img); err != nil {
		fmt.Printf("Failed to create test.bmp: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created test.bmp")

	// Generate WebP using cwebp if available, otherwise create a minimal valid WebP
	if err := saveWebP("test.webp", "test.png"); err != nil {
		fmt.Printf("Note: Could not create test.webp via cwebp: %v\n", err)
		// Create a minimal valid WebP file manually
		if err := createMinimalWebP("test.webp"); err != nil {
			fmt.Printf("Failed to create test.webp: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("Created test.webp")

	// Generate ICO with single image
	if err := saveICO("test.ico", []image.Image{img}); err != nil {
		fmt.Printf("Failed to create test.ico: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created test.ico")

	// Generate ICO with multiple resolutions
	img16 := createTestImage(16, 16, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img48 := createTestImage(48, 48, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img64 := createTestImage(64, 64, color.RGBA{R: 255, G: 255, B: 0, A: 255})
	if err := saveICO("test_multi.ico", []image.Image{img16, img, img48, img64}); err != nil {
		fmt.Printf("Failed to create test_multi.ico: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created test_multi.ico")

	// Generate invalid binary file
	if err := os.WriteFile("invalid.bin", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 0644); err != nil {
		fmt.Printf("Failed to create invalid.bin: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created invalid.bin")

	fmt.Println("\nAll test data files created successfully!")
}

func createTestImage(width, height int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func savePNG(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func saveJPEG(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
}

func saveBMP(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return bmp.Encode(f, img)
}

func saveWebP(filename string, pngSource string) error {
	// Try to use cwebp command-line tool if available
	cmd := exec.Command("cwebp", pngSource, "-o", filename)
	return cmd.Run()
}

func createMinimalWebP(filename string) error {
	// Create a minimal valid WebP file (lossy format)
	// This is a 1x1 pixel WebP image
	// WebP format: RIFF header + WEBP + VP8 chunk
	webpData := []byte{
		// RIFF header
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x1a, 0x00, 0x00, 0x00, // File size - 8 (26 bytes)
		0x57, 0x45, 0x42, 0x50, // "WEBP"
		// VP8 chunk header
		0x56, 0x50, 0x38, 0x20, // "VP8 "
		0x0e, 0x00, 0x00, 0x00, // Chunk size (14 bytes)
		// VP8 bitstream (minimal 1x1 image)
		0x30, 0x01, 0x00, 0x9d, 0x01, 0x2a, 0x01, 0x00,
		0x01, 0x00, 0x00, 0x34, 0x25, 0x9f,
	}
	return os.WriteFile(filename, webpData, 0644)
}

func saveICO(filename string, images []image.Image) error {
	// Encode each image as PNG
	pngDataList := make([][]byte, len(images))
	for i, img := range images {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return err
		}
		pngDataList[i] = buf.Bytes()
	}

	// Calculate total size
	headerSize := 6
	directorySize := 16 * len(images)
	totalImageSize := 0
	for _, data := range pngDataList {
		totalImageSize += len(data)
	}

	// Create ICO file
	icoData := make([]byte, headerSize+directorySize+totalImageSize)

	// Write header
	binary.LittleEndian.PutUint16(icoData[0:2], 0)                    // Reserved
	binary.LittleEndian.PutUint16(icoData[2:4], 1)                    // Type (1 = ICO)
	binary.LittleEndian.PutUint16(icoData[4:6], uint16(len(images))) // Count

	// Write directory entries and image data
	offset := uint32(headerSize + directorySize)
	for i, img := range images {
		entryOffset := headerSize + i*16
		pngData := pngDataList[i]
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// Width and Height (0 means 256)
		w := uint8(width)
		h := uint8(height)
		if width == 256 {
			w = 0
		}
		if height == 256 {
			h = 0
		}

		icoData[entryOffset] = w
		icoData[entryOffset+1] = h
		icoData[entryOffset+2] = 0 // ColorCount
		icoData[entryOffset+3] = 0 // Reserved
		binary.LittleEndian.PutUint16(icoData[entryOffset+4:entryOffset+6], 1)                     // Planes
		binary.LittleEndian.PutUint16(icoData[entryOffset+6:entryOffset+8], 32)                    // BitCount
		binary.LittleEndian.PutUint32(icoData[entryOffset+8:entryOffset+12], uint32(len(pngData))) // Size
		binary.LittleEndian.PutUint32(icoData[entryOffset+12:entryOffset+16], offset)              // Offset

		// Copy PNG data
		copy(icoData[offset:], pngData)
		offset += uint32(len(pngData))
	}

	return os.WriteFile(filename, icoData, 0644)
}
