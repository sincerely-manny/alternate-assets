# Alternate Assets

A CLI utility to slightly alter image assets to change their checksums while keeping visual changes imperceptible.

## Installation

### Option 1: Use with npx (no installation required)

```bash
npx alternate-assets <path> [flags]
```

### Option 2: Install from npm globally

```bash
npm install -g alternate-assets
```

### Option 3: Install from source (Go)

```bash
go install github.com/sincerely-manny/alternate-assets@latest
```

### Option 4: Clone and build locally

```bash
git clone https://github.com/your-username/alternate-assets.git
cd alternate-assets
go build
```

## Usage

```bash
# If installed via Go or built locally
alternate-assets <path> [flags]

# With npx (no installation)
npx alternate-assets <path> [flags]
```

### Arguments

- `<path>`: Path to the image file or directory to process

### Flags

- `-r, --recursive int`: Process directories recursively up to specified depth (default: 0)
- `-v, --verbose`: Display detailed information about the operations
- `-h, --help`: Help for alternate-assets

## Supported Image Formats

- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- WebP (.webp)

## Examples

Process a single image:
```bash
alternate-assets path/to/image.jpg
# or with npx (no installation required)
npx alternate-assets path/to/image.jpg
```

Process all images in a directory (non-recursively):
```bash
alternate-assets path/to/directory
# or with npx (no installation required)
npx alternate-assets path/to/directory
```

Process images recursively with a maximum depth of 3:
```bash
alternate-assets path/to/directory -r 3
# or with npx (no installation required)
npx alternate-assets path/to/directory -r 3
```

Show detailed information:
```bash
alternate-assets path/to/image.png -v
# or with npx (no installation required)
npx alternate-assets path/to/image.png -v
```

## How It Works

The tool creates imperceptible changes to images by:

1. Applying a very subtle brightness adjustment (0.1%) to the image
2. For JPEGs: Re-encoding with slightly different quality settings
3. For PNGs: Using best compression level when saving
4. For other formats: Using the imaging library's default encoding

These changes are enough to alter the file's checksum without introducing visible artifacts or quality degradation.

## Building and Publishing

### Building for Distribution

This project includes scripts for building cross-platform binaries:

```bash
# Build for all platforms (macOS, Linux, Windows)
npm run build:go

# Build and package for npm
npm run build:all
```

The build script creates binaries for multiple platforms and architectures:
- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)

### npm Package Structure

The npm package includes:

- A JavaScript wrapper that detects the user's platform and architecture
- Pre-compiled binaries for all supported platforms
- No post-install scripts or runtime dependencies

This approach ensures a smooth user experience with no compilation steps required during installation.

### Publishing to npm

After building all binaries:

```bash
npm run publish:npm
```

This will publish the npm package with pre-compiled binaries for all supported platforms.

## License

MIT