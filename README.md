![](images/gopher-image.png)
# ColdCase: Integrated Digital Forensics Tool

A comprehensive CLI tool that integrates various digital forensics utilities into a single, unified interface. Similar to netexec but specifically designed for digital forensics investigations and malware analysis.

## Showcase

![](images/final.gif)

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

### External Tool Integration (9 tools)
- **ExifTool**: Extract comprehensive metadata from files
- **Binwalk**: Firmware analysis and extraction
- **Sleuth Kit**: `fls`, `fsstat`, `istat`, `jls`, `tsk_loaddb` - Complete filesystem analysis suite

### Built-in Utilities
- **list**: Display all available tools
- **check**: Verify which external tools are installed

## Installation

### Prerequisites
```bash
# Install required external tools (Ubuntu/Debian)
sudo apt update
sudo apt install -y python3 exiftool binwalk sleuthkit

# DidierStevensSuite is included in the repository
```

### Build
```bash
git clone https://github.com/Fakechippies/ColdCase
cd ColdCase
go build -o bin/coldcase .
```

## Usage

### Basic Commands
```bash
# List all available tools
bin/coldcase list

# Check which tools are installed
bin/coldcase check

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

## Tool Categories

### Document Analysis
- PDF documents: `pdf-parser`, `pdfid`, `1768`
- Office documents: `oledump`
- Email files: `emldump`

### Executable Analysis  
- PE files: `pecheck`
- Embedded scripts: `extractscripts`
- Cobalt Strike: `cs-parse-traffic`

### Data Extraction
- Encoded data: `base64dump`
- File carving: `find-file-in-file`
- Byte manipulation: `cut-bytes`

### Metadata & Analysis
- File metadata: `exif`, `jpegdump`
- File hashing: `hash`
- Statistics: `byte-stats`
- Filesystem: `fls`, `fsstat`, `istat`, `jls`, `tsk_loaddb`

## Architecture

The tool is built in Go with a modular architecture:
- Each tool is implemented as a separate interface
- Commands are dynamically registered using Cobra CLI framework
- External tools are called as subprocesses
- Built-in error handling and tool availability checking

## Contributing

1. Add new tools by implementing the `Tool` interface
2. Create commands in separate files for organization
3. Update `utils.go` to include tools in the `list` command
4. Ensure proper error handling and help text

## License

This project integrates various open-source forensics tools. Please check individual tool licenses for specific requirements.
