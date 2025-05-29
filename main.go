package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

	// Open the image file
	src, err := imaging.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Make a subtle change to the image
	// For PNG and JPG, we'll apply the slightest brightness adjustment
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
		saveErr = jpeg.Encode(outputFile, processed, &jpeg.Options{Quality: 99})
	case ".png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
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