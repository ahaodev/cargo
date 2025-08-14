# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an embedded Linux bus ticket validation system running on ARM-based hardware (i.MX6 UltraLite). The system validates tickets through multiple input methods: IC cards, QR codes, and ID cards, providing audio feedback and display updates.

## Target Platform

- **Hardware**: Freescale i.MX6 UltraLite (ARMv7) embedded Linux system
- **OS**: Linux imx6ul7d 4.9.17 (ARM-based bus hardware)
- **Connection**: SSH access via `ssh -o HostKeyAlgorithms=+ssh-rsa root@192.168.8.30`

## Build System

### Cross-compilation Build
```bash
./build.sh  # Cross-compiles for ARM architecture (Linux only)
```

**IMPORTANT**: This project cannot be compiled on Windows due to:
- ARM cross-compilation toolchain (`arm-linux-gnueabihf-gcc`) not available on Windows
- CGO linking against ARM-compiled `libpos.so` library
- Unix-specific build scripts and dependencies

The build script:
- Sets `GOARCH=arm GOARM=7` for ARMv7 target
- Uses `arm-linux-gnueabihf-gcc` cross-compiler
- Enables CGO for C library integration
- Outputs binary as `runner`

### Development Requirements
- **Linux development environment** (required for compilation)
- ARM cross-compilation toolchain installed
- Access to target ARM device for deployment and testing

### CI/CD Pipeline (Drone)
Project uses Drone CI for automated ARM compilation:
```yaml
# .drone.yml triggers on dev/master branches
- Uses Ubuntu 16.04 with mounted ARM GCC toolchain
- Cross-compiles using arm-linux-gnueabihf-gcc  
- Auto-deploys to target device (192.168.8.30)
- Reboots target device after deployment
```

### Development Commands
```bash
# Build for target ARM platform (Linux only)
./build.sh

# SSH to target device for deployment/testing
ssh -o HostKeyAlgorithms=+ssh-rsa root@192.168.8.30
```

## Architecture

### Core Components

**Main Application Flow** (main.go:29-108):
- Initializes Sentry error tracking
- Loads configuration from `config.yml`
- Retrieves remote configuration from server
- Sets up NTP time sync
- Starts parallel goroutines for device interfaces
- Processes messages from hardware devices

**CGO Integration** (clib/):
- `clib.go` - Go wrapper for C library functions
- `clib.h` - C header definitions for hardware SDK
- `clib.c` - C implementation interfacing with hardware
- Links against pre-compiled SDK: `libpos.so` in `sdk/lib/`

**Hardware SDK** (sdk/):
- Pre-compiled C libraries for device communication
- Headers in `sdk/inc/`: Felica, LED, OS, and tools interfaces
- Static library `libpos.so` provides hardware abstraction

**Message Processing**:
- Channel-based message passing between C hardware layer and Go application
- Three message types: `IC_CARD`, `QRCODE`, `ID_CARD`
- Each triggers API validation and audio/display feedback

### Key Directories

- `api/` - Server communication and ticket validation
- `clib/` - CGO bridge to C hardware libraries  
- `config/` - Configuration management and versioning
- `internal/` - Core business logic (timers, updates, passed counts)
- `pkg/` - Utilities (logging, networking, time sync)
- `screen/` - LCD display management
- `speaker/` - Audio feedback system
- `sdk/` - Pre-compiled ARM hardware SDK

## Configuration

### config.yml Structure
- `deviceType` - Unique device identifier (e.g., 1467826845670965269)
- `serverUrl` - Operations server endpoint for API calls and config retrieval
- `enableIDCard` - Boolean flag to enable/disable ID card reader functionality (0/1)
- `enableICCard` - Boolean flag to enable/disable IC card reader functionality (0/1)

### Configuration Loading
- Remote config fetched from server on startup using `serverUrl` for equipment-specific settings
- Hardware modules initialized based on enable flags in config.yml
- Audio prompts defined for various validation scenarios (A-Z response codes)

## Development Workflow

1. **Platform Requirement**: Must use Linux development environment for compilation
2. **Cross-compilation**: Always use `./build.sh` to build ARM binaries
3. **Testing**: Deploy `runner` binary to target ARM device via SCP
4. **Debugging**: Use SSH to access target device logs  
5. **Hardware Dependencies**: C SDK libraries are pre-compiled and cannot be modified

### Windows Development Limitations
- Code editing and analysis possible on Windows
- Compilation must be done on Linux (WSL, Docker, or native Linux)  
- Cannot run `go build` or `go run` directly on Windows due to CGO ARM dependencies
- **Recommended**: Use Drone CI pipeline for compilation and deployment

## CGO Integration Notes

- CGO links ARM-compiled `libpos.so` library
- C functions called from Go handle direct hardware communication
- Hardware device initialization happens in C layer
- Message passing from C to Go uses channels for thread safety
- Critical hardware operations (IC card reading, display updates) are mutex-protected