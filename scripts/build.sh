#!/bin/bash
set -e

# Configuration
VERSION="1.0.0"
BINARY="alternate-assets"
OUTPUT_DIR="./dist"
NPM_BIN_DIR="./npm/bin"
NPM_DIR="./npm"

# Create output directory if it doesn't exist
mkdir -p $OUTPUT_DIR
mkdir -p $NPM_BIN_DIR

# Copy the main README to npm package directory
echo "Copying README.md to npm package..."
cp README.md $NPM_DIR/

# Build for different platforms
echo "Building $BINARY v$VERSION..."

# List of platforms to build for
PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64" 
  "linux/arm64"
  "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  
  OUTPUT_NAME=$BINARY
  if [ $GOOS = "windows" ]; then
    OUTPUT_NAME+='.exe'
  fi

  echo "Building for $GOOS/$GOARCH..."
  
  GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT_DIR/$OUTPUT_NAME"
  
  # Create archive
  pushd "$OUTPUT_DIR" > /dev/null
  TAR_NAME="${BINARY}_${VERSION}_${GOOS}_${GOARCH}.tar.gz"
  tar -czf "$TAR_NAME" "$OUTPUT_NAME"
  
  # Copy binary to npm/bin directory with platform suffix
  NPM_BINARY="${BINARY}-${GOOS}-${GOARCH}"
  if [ $GOOS = "windows" ]; then
    NPM_BINARY+='.exe'
  fi
  cp "$OUTPUT_NAME" "../$NPM_BIN_DIR/$NPM_BINARY"
  
  rm "$OUTPUT_NAME"
  echo "Created $TAR_NAME and copied to npm/bin as $NPM_BINARY"
  popd > /dev/null
done

echo "Build complete!"
echo "README.md copied to npm package."