![](images/gopher-image.png)

# ColdCase: Integrated Digital Forensics Tool

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://choosealicense.com/licenses/mit/)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-informational)](https://github.com/Fakechippies/ColdCase)
[![Python](https://img.shields.io/badge/Python-3.8+-blue?style=flat&logo=python)](https://www.python.org/)
[![Package Managers](https://img.shields.io/badge/Package%20Managers-apt%20%7C%20dnf%20%7C%20pacman%20%7C%20brew%20%7C%20choco%20%7C%20winget-success)](https://github.com/Fakechippies/ColdCase)

A comprehensive CLI tool that integrates various digital forensics utilities into a single, unified interface. Similar to netexec but specifically designed for digital forensics investigations and malware analysis.

## Showcase

![](images/final.gif)

##  Features Overview

[![DidierStevens Suite](https://img.shields.io/badge/DidierStevens%20Suite-16%20tools-red?style=flat)](#didierstevens-suite-integration-16-tools)
[![Volatility3](https://img.shields.io/badge/Volatility3-22%20tools-purple?style=flat)](#volatility3-memory-forensics-22-tools)
[![Sleuth Kit](https://img.shields.io/badge/Sleuth%20Kit-5%20tools-orange?style=flat)](#external-tool-integration-9-tools)
[![External Tools](https://img.shields.io/badge/External%20Tools-4%20tools-green?style=flat)](#external-tool-integration-9-tools)
[![Total Tools](https://img.shields.io/badge/Total%20Tools-47%20integrated-blue?style=flat)](#features)

## Features

### DidierStevens Suite Integration (16 tools)
- **PDF Analysis**: `1768`, `pdf-parser`, `pdfid` - Comprehensive PDF malware analysis
- **Office Documents**: `oledump` - Analyze OLE files and Office documents
- **PE Analysis**: `pecheck` - Detailed PE file information and analysis
- **Data Extraction**: `base64dump`, `emldump`, `cut-bytes`, `find-file-in-file`, `extractscripts` - Extract various data types from files
- **Image Analysis**: `jpegdump` - JPEG structure and metadata analysis
- **Utilities**: `hash`, `byte-stats` - File hashing and byte distribution analysis
- **Cobalt Strike**: `cs-parse-traffic` - Parse and analyze Cobalt Strike traffic
- **System Analysis**: `amsiscan` - Windows AMSI bypass detection

### Volatility3 Memory Forensics (22 tools)
- **Core**: `vol` (main framework), `volshell` (interactive shell), `info` (image info)
- **Windows Analysis**: `windows.pslist`, `windows.pstree`, `windows.dlllist`, `windows.handles`, `windows.cmdline`, `windows.envars`, `windows.filescan`, `windows.modules`, `windows.driverscan`, `windows.callbacks`, `windows.services`, `windows.registry`, `windows.hashdump`
- **Linux Analysis**: `linux.pslist`, `linux.pstree`, `linux.bash`, `linux.proc_maps`
- **macOS Analysis**: `mac.pslist`, `mac.pstree`

### External Tool Integration (9 tools)
- **ExifTool**: Extract comprehensive metadata from files
- **Binwalk**: Firmware analysis and extraction
- **Sleuth Kit**: `fls`, `fsstat`, `istat`, `jls`, `tsk_loaddb` - Complete filesystem analysis suite

### Built-in Utilities
- **list**: Display all available tools
- **check**: Verify which external tools are installed
- **install**: Install all project dependencies (with optional uv support)
- **deps**: Manage project dependencies (install, check, update)

## Installation

### Prerequisites

#### Option 1: Automatic Installation (Recommended)
```bash
# Install all dependencies automatically for any supported platform
bin/coldcase install

# With uv for faster Python dependency management
bin/coldcase install --uv
```

#### Option 2: Manual Installation
**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit
```

**Linux (Fedora/RHEL/CentOS):**
```bash
sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit
# or for older systems:
sudo yum install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit
```

**Linux (Arch):**
```bash
sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit
```

**macOS:**
```bash
brew install python exiftool binwalk sleuthkit
```

**Windows:**
```bash
# Using Chocolatey:
choco install -y python exiftool

# Using Winget:
winget install -e --id Python.Python.3
winget install -e --id ExifTool.ExifTool
```

**Install Python dependencies:**
```bash
cd volatility3 && python3 -m pip install -e .[full]

# DidierStevensSuite is included in the repository
```

### Build

[![Go Build](https://img.shields.io/badge/Go-Build-success?style=flat&logo=go)](#installation)
[![Dependency Management](https://img.shields.io/badge/Dependencies-Automatic-green?style=flat)](#automatic-installation-recommended)
[![Quick Start](https://img.shields.io/badge/Quick-Start-blue?style=flat)](#usage)

```bash
git clone https://github.com/Fakechippies/ColdCase
cd ColdCase

# Install dependencies (automatic)
go build -o bin/coldcase .
./bin/coldcase install

# Or with uv for faster Python dependency management
./bin/coldcase install --uv
```

## Usage

### Basic Commands
```bash
# List all available tools
bin/coldcase list

# Check which tools are installed
bin/coldcase check

# Show platform-specific setup information
bin/coldcase platform

# Install dependencies
bin/coldcase install

# Manage dependencies
bin/coldcase deps check
bin/coldcase deps update

# Get help for specific tool
bin/coldcase <tool> --help
```

### Examples

#### PDF Analysis
```bash
# Analyze PDF for malicious content
bin/coldcase pdf-parser suspicious.pdf

# Quick PDF scan
bin/coldcase pdfid document.pdf

# Specialized 1768 PDF analysis
bin/coldcase 1768 malware.pdf
```

#### Office Document Analysis
```bash
# Analyze OLE files (Word, Excel, etc.)
bin/coldcase oledump document.doc

# Extract embedded scripts
bin/coldcase extractscripts file.exe
```

#### PE File Analysis
```bash
# Display PE file information
bin/coldcase pecheck malware.exe
```

#### Data Extraction
```bash
# Extract base64 strings
bin/coldcase base64dump encoded_file.bin

# Find embedded files
bin/coldcase find-file-in-file container.bin

# Extract specific byte ranges
bin/coldcase cut-bytes -o output.bin -s 100 -l 50 file.bin
```

#### Metadata Analysis
```bash
# Extract metadata with ExifTool
bin/coldcase exif image.jpg

# Calculate file hashes
bin/coldcase hash -a sha256 malware.exe

# Filesystem analysis
bin/coldcase fls -r /dev/sdX1
```

#### Memory Analysis with Volatility3
```bash
# Display information about a memory image
bin/coldcase info -f memory.dmp

# List running processes
bin/coldcase windows.pslist -f memory.dmp

# Show process tree
bin/coldcase windows.pstree -f memory.dmp

# Interactive memory analysis
bin/coldcase volshell -f memory.dmp

# Linux memory analysis
bin/coldcase linux.pslist -f linux_memory.dmp

# Registry analysis
bin/coldcase windows.registry -f memory.dmp
```

##  Tool Categories

###  Document Analysis
[![PDF](https://img.shields.io/badge/PDF-Tools-red?style=flat)](#document-analysis)
[![Office](https://img.shields.io/badge/Office-Tools-orange?style=flat)](#document-analysis)
- PDF documents: `pdf-parser`, `pdfid`, `1768`
- Office documents: `oledump`
- Email files: `emldump`

###  Executable Analysis  
[![PE](https://img.shields.io/badge/PE-Analysis-yellow?style=flat)](#executable-analysis)
- PE files: `pecheck`
- Embedded scripts: `extractscripts`
- Cobalt Strike: `cs-parse-traffic`

###  Data Extraction
[![Data](https://img.shields.io/badge/Data-Extraction-blue?style=flat)](#data-extraction)
- Encoded data: `base64dump`
- File carving: `find-file-in-file`
- Byte manipulation: `cut-bytes`

###  Metadata & Analysis
[![Metadata](https://img.shields.io/badge/Metadata-Analysis-green?style=flat)](#metadata--analysis)
- File metadata: `exif`, `jpegdump`
- File hashing: `hash`
- Statistics: `byte-stats`
- Filesystem: `fls`, `fsstat`, `istat`, `jls`, `tsk_loaddb`

### 🧠 Memory Forensics
[![Windows](https://img.shields.io/badge/Windows-Memory-0078D4?style=flat&logo=windows)](#memory-forensics)
[![Linux](https://img.shields.io/badge/Linux-Memory-FCC624?style=flat&logo=linux)](#memory-forensics)
[![macOS](https://img.shields.io/badge/macOS-Memory-000000?style=flat&logo=apple)](#memory-forensics)
- Windows analysis: `windows.pslist`, `windows.pstree`, `windows.dlllist`, `windows.handles`, `windows.cmdline`, `windows.envars`, `windows.filescan`, `windows.modules`, `windows.driverscan`, `windows.callbacks`, `windows.services`, `windows.registry`, `windows.hashdump`
- Linux analysis: `linux.pslist`, `linux.pstree`, `linux.bash`, `linux.proc_maps`
- macOS analysis: `mac.pslist`, `mac.pstree`
- Core tools: `vol`, `volshell`, `info`

## Dependency Management

ColdCase includes built-in dependency management with support for both pip and uv:

### Automatic Installation
```bash
# Install all dependencies using system package manager and pip
./bin/coldcase install

# Install with uv for faster Python dependency management
./bin/coldcase install --uv

# Skip system packages (only install Python dependencies)
./bin/coldcase install --skip-system
```

### Manual Dependency Management
```bash
# Check for missing dependencies
./bin/coldcase deps check

# Install/update Python dependencies only
./bin/coldcase deps install

# Update Python dependencies
./bin/coldcase deps update
```

### Dependencies Included

[![Python](https://img.shields.io/badge/Python-3.8+-blue?style=flat&logo=python)](#prerequisites)
[![uv](https://img.shields.io/badge/uv-Optional-fast%20dependency%20management-green?style=flat&logo=python)](#automatic-installation-recommended)
[![Volatility3](https://img.shields.io/badge/Volatility3-Memory%20Forensics-purple?style=flat)](#volatility3-memory-forensics-22-tools)

- **System Tools**: python3, exiftool, binwalk, sleuthkit
- **Python Packages**: volatility3 with full feature set  
- **Optional**: uv for faster Python package management

### Supported Platforms

[![Linux](https://img.shields.io/badge/Linux-FCC624?style=flat&logo=linux&logoColor=black)](#option-1-automatic-installation-recommended)
[![Ubuntu](https://img.shields.io/badge/Ubuntu-E95420?style=flat&logo=ubuntu&logoColor=white)](#option-1-automatic-installation-recommended)
[![Fedora](https://img.shields.io/badge/Fedora-51A2DA?style=flat&logo=fedora&logoColor=white)](#option-1-automatic-installation-recommended)
[![Arch](https://img.shields.io/badge/Arch-1793D1?style=flat&logo=arch-linux&logoColor=white)](#option-1-automatic-installation-recommended)
[![macOS](https://img.shields.io/badge/macOS-000000?style=flat&logo=apple&logoColor=white)](#option-1-automatic-installation-recommended)
[![Windows](https://img.shields.io/badge/Windows-0078D4?style=flat&logo=windows&logoColor=white)](#option-1-automatic-installation-recommended)

**Package Managers:**
- **Linux**: Ubuntu/Debian (apt), Fedora/RHEL/CentOS (dnf/yum), Arch (pacman)
- **macOS**: Homebrew  
- **Windows**: Chocolatey, Winget

### Platform-Specific Notes
- **Linux**: Uses distribution package managers with correct package names
- **macOS**: Requires Homebrew to be installed first
- **Windows**: Chocolatey or Winget must be installed first
- **All platforms**: Python 3.8+ is required for volatility3

## Architecture

The tool is built in Go with a modular architecture:
- Each tool is implemented as a separate interface
- Commands are dynamically registered using Cobra CLI framework
- External tools are called as subprocesses
- Built-in dependency management and tool availability checking
- Support for uv for accelerated Python package management

## Contributing

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat&logo=github)](http://makeapullrequest.com)
[![Issues](https://img.shields.io/github/issues/Fakechippies/ColdCase?style=flat&logo=github)](https://github.com/Fakechippies/ColdCase/issues)

1. Add new tools by implementing the `Tool` interface
2. Create commands in separate files for organization
3. Update `utils.go` to include tools in the `list` command
4. Ensure proper error handling and help text

##  Project Stats

[![GitHub stars](https://img.shields.io/github/stars/Fakechippies/ColdCase?style=flat&logo=github)](https://github.com/Fakechippies/ColdCase/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/Fakechippies/ColdCase?style=flat&logo=github)](https://github.com/Fakechippies/ColdCase/network)
[![GitHub issues](https://img.shields.io/github/issues/Fakechippies/ColdCase?style=flat&logo=github)](https://github.com/Fakechippies/ColdCase/issues)
[![GitHub license](https://img.shields.io/github/license/Fakechippies/ColdCase?style=flat&logo=github)](https://github.com/Fakechippies/ColdCase/blob/main/LICENSE)

## License

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://choosealicense.com/licenses/mit/)

This project integrates various open-source forensics tools. Please check individual tool licenses for specific requirements.
