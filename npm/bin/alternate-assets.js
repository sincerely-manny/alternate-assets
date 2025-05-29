#!/usr/bin/env node

const { execFileSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

// Get platform and architecture
const platform = os.platform();
const arch = os.arch();

// Map Node.js platform and arch to Go platform and arch
function getPlatformName() {
  const platforms = {
    'win32': 'windows',
    'darwin': 'darwin',
    'linux': 'linux',
    'freebsd': 'freebsd',
  };
  return platforms[platform] || platform;
}

function getArchName() {
  const architectures = {
    'x64': 'amd64',
    'arm64': 'arm64',
    'x32': '386',
    'ia32': '386',
  };
  return architectures[arch] || arch;
}

const goPlatform = getPlatformName();
const goArch = getArchName();
const extension = platform === 'win32' ? '.exe' : '';

// Define binary name based on platform and architecture
const platformSpecificBinary = `alternate-assets-${goPlatform}-${goArch}${extension}`;
const genericBinary = `alternate-assets${extension}`;

// Possible locations of the binary
const binLocations = [
  // 1. Platform-specific binary bundled with the npm package
  path.join(__dirname, platformSpecificBinary),
  // 2. Generic binary in the same directory
  path.join(__dirname, genericBinary),
  // 3. Local development - relative to this script
  path.join(__dirname, '..', '..', genericBinary),
  // 4. System path (not checking explicitly, will be handled by execFileSync)
  genericBinary
];

// Find the binary
let binaryPath;
for (const location of binLocations) {
  try {
    fs.accessSync(location, fs.constants.X_OK);
    binaryPath = location;
    break;
  } catch (err) {
    // Binary not found at this location, continue searching
  }
}

// If binary not found, exit
if (!binaryPath) {
  console.error('Error: Could not find alternate-assets binary. Please ensure it is installed correctly.');
  process.exit(1);
}

try {
  // Pass all command-line arguments to the binary and execute it
  execFileSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
} catch (error) {
  // The binary will handle its own error messages
  process.exit(error.status || 1);
}