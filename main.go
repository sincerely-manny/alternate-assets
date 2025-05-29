package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/spf13/cobra"
)

// Version information
const (
	Version = "1.0.0"
)

// Supported image extensions
var supportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// estimateJpegQuality attempts to estimate the quality of a JPEG image
func estimateJpegQuality(filePath string) int {
	// This is a simplistic approach - in reality, estimating JPEG quality accurately 
	// is challenging without access to the original encoding parameters
	
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return 75 // Return default quality on error
	}
	defer file.Close()
	
	// Read file info to get size
	fileInfo, err := file.Stat()
	if err != nil {
		return 75 // Return default quality on error
	}
	
	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil || format != "jpeg" {
		return 75 // Return default quality on error
	}
	
	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixelCount := width * height
	
	// Calculate bytes per pixel
	bytesPerPixel := float64(fileInfo.Size()) / float64(pixelCount)
	
	// Heuristic mapping of bytes-per-pixel to quality
	// This is approximate and will vary based on image content
	switch {
	case bytesPerPixel < 0.5:
		return 60 // Low quality
	case bytesPerPixel < 0.75:
		return 70
	case bytesPerPixel < 1.0:
		return 80
	case bytesPerPixel < 1.5:
		return 90
	default:
		return 95 // High quality
	}
}

// estimatePngCompressionLevel estimates PNG compression level
func estimatePngCompressionLevel(filePath string) png.CompressionLevel {
	// For PNG, exact compression level detection is difficult
	// We'll return a fixed level for now, but in the future this could be improved
	
	// Simple heuristic based on file size
	file, err := os.Open(filePath)
	if err != nil {
		return png.DefaultCompression
	}
	defer file.Close()
	
	fileInfo, err := file.Stat()
	if err != nil {
		return png.DefaultCompression
	}
	
	// If the file is already highly compressed, use less compression
	// to ensure the checksum changes
	if fileInfo.Size() < 10*1024 { // Less than 10KB
		return png.NoCompression
	} else if fileInfo.Size() < 100*1024 { // Less than 100KB
		return png.BestSpeed
	} else if fileInfo.Size() < 1024*1024 { // Less than 1MB
		return png.DefaultCompression
	} else {
		return png.BestCompression
	}
}

var (
	recursiveDepth int
	verbose        bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "alternate-assets [path]",
		Short: "Slightly alters image assets to change their checksums while keeping visual changes imperceptible",
		Version: Version,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			stat, err := os.Stat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if stat.IsDir() {
				err = processDirectory(path, 0, recursiveDepth)
			} else {
				err = processFile(path)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().IntVarP(&recursiveDepth, "recursive", "r", 0, "Process directories recursively up to specified depth")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Display detailed information about the operations")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Calculate file checksum
func calculateFileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Process a single image file
func processFile(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	
	if !supportedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	originalChecksum, err := calculateFileChecksum(path)
	if err != nil {
		return fmt.Errorf("failed to calculate original checksum: %w", err)
	}

	// Estimate quality and other properties based on file format
	jpegQuality := 75 // Default quality if we can't determine
	var compressionLevel png.CompressionLevel = png.DefaultCompression

	if ext == ".jpg" || ext == ".jpeg" {
		// For JPEG, try to estimate quality
		jpegQuality = estimateJpegQuality(path)
	} else if ext == ".png" {
		// For PNG, try to detect current compression level
		compressionLevel = estimatePngCompressionLevel(path)
	}

	// Open the image file for processing
	src, err := imaging.Open(path)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Make a subtle change to the image (brightness adjustment by 0.1%)
	processed := imaging.AdjustBrightness(src, 0.1)

	// Save the modified image
	outputFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	var saveErr error
	switch ext {
	case ".jpg", ".jpeg":
		// Apply a slight change to quality (±1) to ensure checksum changes 
		// while maintaining visual similarity
		adjustedQuality := jpegQuality
		if jpegQuality > 90 {
			adjustedQuality = jpegQuality - 1 // Decrease slightly for high quality
		} else {
			adjustedQuality = jpegQuality + 1 // Increase slightly for lower quality
		}
		saveErr = jpeg.Encode(outputFile, processed, &jpeg.Options{Quality: adjustedQuality})
		if verbose {
			fmt.Printf("JPEG quality adjustment: %d → %d\n", jpegQuality, adjustedQuality)
		}
	case ".png":
		// For PNG, we'll use the estimated compression but make a small adjustment
		encoder := png.Encoder{CompressionLevel: compressionLevel}
		saveErr = encoder.Encode(outputFile, processed)
	default:
		// For other formats, use the imaging library's Save function
		saveErr = imaging.Save(processed, path)
	}

	if saveErr != nil {
		return fmt.Errorf("failed to save image: %w", saveErr)
	}

	// Calculate the new checksum
	newChecksum, err := calculateFileChecksum(path)
	if err != nil {
		return fmt.Errorf("failed to calculate new checksum: %w", err)
	}

	// Print results
	if verbose {
		fmt.Printf("Processed: %s\n", path)
		fmt.Printf("Original checksum: %s\n", originalChecksum)
		fmt.Printf("New checksum: %s\n", newChecksum)
		fmt.Printf("Checksum changed: %t\n\n", originalChecksum != newChecksum)
	} else {
		changeSymbol := "✗"
		if originalChecksum != newChecksum {
			changeSymbol = "✓"
		}
		fmt.Printf("Processed: %s %s\n", path, changeSymbol)
	}

	return nil
}

// Process a directory
func processDirectory(path string, currentDepth, maxDepth int) error {
	if currentDepth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if entry.IsDir() && currentDepth < maxDepth {
			err := processDirectory(entryPath, currentDepth+1, maxDepth)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error processing directory %s: %v\n", entryPath, err)
			}
		} else if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if supportedExtensions[ext] {
				err := processFile(entryPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", entryPath, err)
				}
			}
		}
	}

	return nil
}