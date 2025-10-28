<div align="center">

# idfmgr
#### Simplifying ESP32 project management

<img alt="idfmgr logo" src="https://raw.githubusercontent.com/Dwarf1er/idfmgr/master/assets/idfmgr-logo.png" height="250" />

![License](https://img.shields.io/github/license/Dwarf1er/idfmgr?style=for-the-badge)
![Issues](https://img.shields.io/github/issues/Dwarf1er/idfmgr?style=for-the-badge)
![PRs](https://img.shields.io/github/issues-pr/Dwarf1er/idfmgr?style=for-the-badge)
![Contributors](https://img.shields.io/github/contributors/Dwarf1er/idfmgr?style=for-the-badge)
![Stars](https://img.shields.io/github/stars/Dwarf1er/idfmgr?style=for-the-badge)

</div>

## Project Description

**idfmgr** is a command-line wrapper around Espressif's idf.py tool that simplifies ESP32 development workflows. It handles the complexity of managing multiple ESP-IDF versions, provides project templates, and makes it easy to switch between GCC and Clang toolchains—all while leveraging the power of the official idf.py tool underneath.

### Key Features

- **Version Management**: Install, list, and remove multiple ESP-IDF versions
- **Project Templates**: Create projects with base or Arduino templates
- **Dual Toolchain Support**: Build with GCC or Clang (separate build directories)
- **Per-Project Versioning**: Track ESP-IDF version with `.espidf-version` files
- **Integrated Workflow**: Build and flash with automatic environment setup
- **Multi-Target Support**: ESP32, ESP32-S2, ESP32-S3, ESP32-C3, ESP32-C6, ESP32-H2

# Table of Contents
- [idfmgr](#idfmgr)
			- [Simplifying ESP32 project management](#simplifying-esp32-project-management)
	- [Project Description](#project-description)
		- [Key Features](#key-features)
- [Table of Contents](#table-of-contents)
	- [Installation](#installation)
		- [From Source](#from-source)
		- [Prerequisites](#prerequisites)
	- [Quick Start](#quick-start)
	- [Commands](#commands)
		- [Version Management](#version-management)
			- [`list`](#list)
			- [`install <version>`](#install-version)
			- [`installed`](#installed)
			- [`remove [version...]`](#remove-version)
		- [Project Management](#project-management)
			- [`create <project-name>`](#create-project-name)
		- [Building and Flashing](#building-and-flashing)
			- [`build`](#build)
			- [`flash`](#flash)
	- [Templates](#templates)
		- [Base Template](#base-template)
		- [Arduino Template (`--arduino`)](#arduino-template---arduino)
	- [Configuration](#configuration)
		- [Environment Variables](#environment-variables)
		- [Per-Project Configuration](#per-project-configuration)
	- [Tips \& Tricks](#tips--tricks)
		- [Dual Toolchain Workflow](#dual-toolchain-workflow)
		- [Global Verbose Mode](#global-verbose-mode)
		- [Quick Arduino Development](#quick-arduino-development)
	- [ESP IDF Project Structure](#esp-idf-project-structure)
	- [License](#license)

## Installation

### Using go install (Recommended)

If you have Go installed (1.21+), this is the easiest method:
```bash
go install github.com/Dwarf1er/idfmgr@latest
```

Make sure `$GOPATH/bin` (typically `~/go/bin`) is in your `PATH`.

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/Dwarf1er/idfmgr/releases), replace version as needed in the commands:

**Linux:**
```bash
wget https://github.com/Dwarf1er/idfmgr/releases/download/vX.X.X/idfmgr-linux-x64-vX.X.X
chmod +x idfmgr-linux-x64-vX.X.X
sudo mv idfmgr-linux-x64-vX.X.X /usr/local/bin/idfmgr
```

**macOS:**
```bash
# Intel Mac
wget https://github.com/Dwarf1er/idfmgr/releases/download/vX.X.X/idfmgr-osx-x64-vX.X.X
chmod +x idfmgr-osx-x64-vX.X.X
sudo mv idfmgr-osx-x64-vX.X.X /usr/local/bin/idfmgr

# Apple Silicon Mac
wget https://github.com/Dwarf1er/idfmgr/releases/download/vX.X.X/idfmgr-osx-arm64-vX.X.X
chmod +x idfmgr-osx-arm64-vX.X.X
sudo mv idfmgr-osx-arm64-vX.X.X /usr/local/bin/idfmgr
```

**Windows:**
Download the `.exe` file and add it to your PATH.

### Build from Source
```bash
git clone https://github.com/Dwarf1er/idfmgr.git
cd idfmgr
go build -o idfmgr
sudo mv idfmgr /usr/local/bin/
# or move to a directory in your PATH on Windows
```

### Prerequisites

`idfmgr` will check for these automatically, but you'll need:
- Git
- Python 3
- CMake
- Ninja build system
- wget

## Quick Start
```bash
# 1. Install an ESP-IDF version
idfmgr install v5.1.2

# 2. Create a new project
idfmgr create my-project

# 3. Build with GCC (default)
cd my-project
idfmgr build

# 4. Flash to device
idfmgr flash --monitor
```

## Commands

### Version Management

#### `list`
List available ESP-IDF versions from GitHub releases
```bash
idfmgr list
```

#### `install <version>`
Install a specific ESP-IDF version
```bash
# Install specific version
idfmgr install v5.1.2

# Install latest version
idfmgr install latest

# Skip prerequisite checks
idfmgr install v5.1.2 --skip-prereqs

# Skip Clang toolchain installation
idfmgr install v5.1.2 --skip-clang
```

#### `installed`
List currently installed ESP-IDF versions
```bash
idfmgr installed
```

#### `remove [version...]`
Remove installed ESP-IDF versions
```bash
# Remove specific version
idfmgr remove v5.1.2

# Remove multiple versions
idfmgr remove v4.4.6 v5.0.0

# Remove all versions
idfmgr remove all

# Skip confirmation prompt
idfmgr remove v5.1.2 --force
```

### Project Management

#### `create <project-name>`
Create a new ESP-IDF project
```bash
# Basic project
idfmgr create my-project

# Arduino-based project
idfmgr create my-arduino-project --arduino

# Specific ESP-IDF version
idfmgr create my-project --version v5.1.2

# Specific target chip
idfmgr create my-project --target esp32s3

# Combined options
idfmgr create my-project --arduino --target esp32s3 --version v5.1.2
```

**What gets created:**
- `.espidf-version` - Tracks ESP-IDF version for this project
- `.clangd` - LSP configuration for IDE support
- `sdkconfig.defaults` - Sensible default configurations
- `.gitignore` - Ignores build artifacts and generated files
- Git repository with initial commit

#### `info`

Show project and environment information
```bash
idfmgr info
```

**Output includes:**
- Current ESP-IDF version
- Installation path
- Build status (GCC/Clang)
- Manual activation instructions
- Usage examples

### Building and Flashing

#### `build`
Build the current project
```bash
# Build with GCC (default) - output: build/
idfmgr build

# Build with Clang - output: build-clang/
idfmgr build --clang
```

#### `flash`
Flash the built project to device
```bash
# Flash GCC build
idfmgr flash

# Flash Clang build
idfmgr flash --clang

# Flash and open serial monitor
idfmgr flash --monitor
idfmgr flash -m

# Specify serial port
idfmgr flash --port /dev/ttyUSB0
idfmgr flash -p /dev/ttyUSB0

# Combined options
idfmgr flash --clang --monitor --port /dev/ttyUSB1
```

#### `exec [idf.py args...]`

Execute any idf.py command with proper environment setup
```bash
# Open menuconfig
idfmgr exec menuconfig

# Monitor serial output
idfmgr exec monitor

# Monitor with specific port
idfmgr exec -p /dev/ttyUSB0 monitor

# Erase flash
idfmgr exec erase-flash

# Full clean
idfmgr exec fullclean

# App-only flash (faster)
idfmgr exec app-flash
```

This command is perfect for accessing idf.py features not wrapped by idfmgr, while still benefiting from automatic environment management.

## Templates

### Base Template

Standard ESP-IDF project with:
- Renamed `main.c` for consistency
- Pre-configured `.clangd` for LSP support
- Sensible defaults in `sdkconfig.defaults`
- Git repository initialized

### Arduino Template (`--arduino`)

All base template features plus:
- Arduino-ESP32 as git submodule
- Pre-configured CMake for Arduino support
- `main.cpp` with Arduino-style `setup()` and `loop()`
- Serial communication ready

---

## Configuration

### Environment Variables

- `ESP_BASE` - Installation directory for ESP-IDF versions (default: `~/.esp`)
```bash
export ESP_BASE=/custom/path/to/esp
```

### Per-Project Configuration

Each project contains a `.espidf-version` file:
```
v5.1.2
```

This ensures consistent ESP-IDF version across builds and between developers.

---

## Tips & Tricks

### Dual Toolchain Workflow

Build and flash with both toolchains:
```bash
idfmgr build          # GCC -> build/
idfmgr build --clang  # Clang -> build-clang/

idfmgr flash          # Flash GCC build
idfmgr flash --clang  # Flash Clang build
```

### Quick Arduino Development
```bash
idfmgr create blink --arduino --target esp32s3
cd blink
idfmgr build --clang
idfmgr flash --clang --monitor
```

### Using idf.py Directly

When you need features not wrapped by idfmgr, use `exec`:
```bash
# Instead of manually sourcing export.sh and running idf.py
idfmgr exec menuconfig

# Instead of:
# . ~/esp/esp-idf-v5.1.2/export.sh
# idf.py menuconfig
```

Or check the manual activation path:
```bash
idfmgr info
# Shows: . ~/esp/esp-idf-v5.1.2/export.sh
```

## ESP IDF Project Structure
```
my-project/
├── .espidf-version          # ESP-IDF version tracking
├── .clangd                  # LSP configuration
├── .gitignore               # Build artifacts ignored
├── sdkconfig.defaults       # Default configurations
├── CMakeLists.txt           # Root CMake configuration
├── main/
│   ├── CMakeLists.txt       # Main component configuration
│   └── main.c (or main.cpp) # Application entry point
└── components/              # Custom components
    └── arduino/             # (Arduino template only)
```

## License

This software is licensed under the [MIT license](LICENSE)