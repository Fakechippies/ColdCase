#!/usr/bin/env bash
# =============================================================================
# ColdCase – MANUAL Installation Benchmark
# Measures the real elapsed time to install each forensics tool individually
# using apt-get / pip3 as a normal user would do it without ColdCase.
#
# Usage:  bash benchmark_manual_install.sh [--dry-run]
#   --dry-run  Print the commands that would run without executing them.
# =============================================================================
set -euo pipefail

DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

WORK_DIR="$(mktemp -d /tmp/coldcase_manual_bench_XXXXXX)"
LOG_FILE="$WORK_DIR/install_log.txt"
RESULTS_CSV="$WORK_DIR/results.csv"

trap 'echo; echo "Temp dir left at: $WORK_DIR"' EXIT

cat_or_echo() {
  if $DRY_RUN; then
    echo "[DRY-RUN] $*"
  else
    eval "$*"
  fi
}

time_cmd() {
  local label="$1"; shift
  local start end elapsed
  start=$(date +%s%N)
  if $DRY_RUN; then
    echo "[DRY-RUN] $*"
    sleep 0.01   # simulate tiny delay so CSV isn't all zeros
  else
    eval "$*" >> "$LOG_FILE" 2>&1 || true
  fi
  end=$(date +%s%N)
  elapsed=$(( (end - start) / 1000000 ))   # milliseconds
  echo "$label,$elapsed" >> "$RESULTS_CSV"
  printf "  %-45s %6d ms\n" "$label" "$elapsed"
}

# ─────────────────────────────────────────────────────────────────────────────
# Header
# ─────────────────────────────────────────────────────────────────────────────
echo "============================================================"
echo "  ColdCase  ·  Manual Installation Benchmark"
echo "  Working dir : $WORK_DIR"
echo "  Dry-run     : $DRY_RUN"
echo "  Started at  : $(date)"
echo "============================================================"
echo ""
echo "tool,elapsed_ms" > "$RESULTS_CSV"

GRAND_START=$(date +%s%N)

# ─────────────────────────────────────────────────────────────────────────────
# 1. System update (required before installs)
# ─────────────────────────────────────────────────────────────────────────────
echo "[ Phase 1 ] System update"
time_cmd "apt-get update" \
  "sudo apt-get update -y -qq"

# ─────────────────────────────────────────────────────────────────────────────
# 2. Network Forensics tools
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 2 ] Network Forensics"
time_cmd "tshark"        "sudo apt-get install -y -qq tshark"
time_cmd "tcpdump"       "sudo apt-get install -y -qq tcpdump"
time_cmd "zeek"          "sudo apt-get install -y -qq zeek || true"
time_cmd "ngrep"         "sudo apt-get install -y -qq ngrep"
time_cmd "tcpflow"       "sudo apt-get install -y -qq tcpflow"
time_cmd "tcpreplay"     "sudo apt-get install -y -qq tcpreplay"
time_cmd "tcpstat"       "sudo apt-get install -y -qq tcpstat"
time_cmd "argus"         "sudo apt-get install -y -qq argus-client"
time_cmd "p0f"           "sudo apt-get install -y -qq p0f"

# ─────────────────────────────────────────────────────────────────────────────
# 3. File Carving & Recovery
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 3 ] File Carving & Recovery"
time_cmd "foremost"      "sudo apt-get install -y -qq foremost"
time_cmd "scalpel"       "sudo apt-get install -y -qq scalpel"
time_cmd "testdisk"      "sudo apt-get install -y -qq testdisk"
time_cmd "ddrescue"      "sudo apt-get install -y -qq gddrescue"
time_cmd "safecopy"      "sudo apt-get install -y -qq safecopy"
time_cmd "bulk_extractor" "sudo apt-get install -y -qq bulk-extractor"

# ─────────────────────────────────────────────────────────────────────────────
# 4. Malware & Pattern Matching
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 4 ] Malware & Pattern Matching"
time_cmd "yara"          "sudo apt-get install -y -qq yara"
time_cmd "strings"       "sudo apt-get install -y -qq binutils"
time_cmd "floss (pip)"   "pip3 install -q flare-floss || true"
time_cmd "capa (pip)"    "pip3 install -q flare-capa || true"

# ─────────────────────────────────────────────────────────────────────────────
# 5. Hashing & Verification
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 5 ] Hashing & Verification"
time_cmd "md5deep/hashdeep" "sudo apt-get install -y -qq md5deep"
time_cmd "ssdeep"        "sudo apt-get install -y -qq ssdeep"

# ─────────────────────────────────────────────────────────────────────────────
# 6. Steganography & Media
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 6 ] Steganography & Media"
time_cmd "steghide"      "sudo apt-get install -y -qq steghide"
time_cmd "mediainfo"     "sudo apt-get install -y -qq mediainfo"
time_cmd "zsteg (gem)"   "sudo gem install zsteg -q || true"
time_cmd "stegdetect"    "sudo apt-get install -y -qq stegdetect || true"

# ─────────────────────────────────────────────────────────────────────────────
# 7. Sleuth Kit
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 7 ] Sleuth Kit"
time_cmd "sleuthkit"     "sudo apt-get install -y -qq sleuthkit"

# ─────────────────────────────────────────────────────────────────────────────
# 8. System Utilities
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 8 ] System Utilities"
time_cmd "xxd"           "sudo apt-get install -y -qq xxd"
time_cmd "objdump/nm"    "sudo apt-get install -y -qq binutils"
time_cmd "exiftool"      "sudo apt-get install -y -qq exiftool"
time_cmd "binwalk"       "sudo apt-get install -y -qq binwalk"

# ─────────────────────────────────────────────────────────────────────────────
# 9. Windows Artifacts
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 9 ] Windows Artifacts"
time_cmd "regripper"     "sudo apt-get install -y -qq regripper || true"
time_cmd "regrippy (pip)" "pip3 install -q regrippy || true"
time_cmd "analyzeMFT (pip)" "pip3 install -q analyzeMFT || true"
time_cmd "sleuthkit (ntfs)" "sudo apt-get install -y -qq sleuthkit"

# ─────────────────────────────────────────────────────────────────────────────
# 10. Timeline & Log Analysis
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 10 ] Timeline & Log Analysis"
time_cmd "plaso (pip)"   "pip3 install -q plaso || true"
time_cmd "chainsaw (cargo)" "cargo install chainsaw --quiet 2>/dev/null || true"

# ─────────────────────────────────────────────────────────────────────────────
# 11. Mobile Forensics
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 11 ] Mobile Forensics"
time_cmd "adb"           "sudo apt-get install -y -qq adb"
time_cmd "libimobiledevice" "sudo apt-get install -y -qq libimobiledevice-utils"

# ─────────────────────────────────────────────────────────────────────────────
# 12. Didier Stevens Suite
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 12 ] Didier Stevens Suite (Python scripts)"
DS_DIR="$WORK_DIR/DidierStevensSuite"
time_cmd "DidierStevens clone" \
  "git clone -q https://github.com/DidierStevens/DidierStevensSuite.git '$DS_DIR'"
time_cmd "python-pdfminer (pip)" "pip3 install -q pdfminer.six"
time_cmd "python-olefile (pip)"  "pip3 install -q olefile"

# ─────────────────────────────────────────────────────────────────────────────
# 13. Volatility3
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 13 ] Volatility3 (git + pip)"
VOL_DIR="$WORK_DIR/volatility3"
time_cmd "volatility3 clone" \
  "git clone -q https://github.com/volatilityfoundation/volatility3.git '$VOL_DIR'"
time_cmd "volatility3 pip install" \
  "pip3 install -q -r '$VOL_DIR/requirements.txt' || true"

# ─────────────────────────────────────────────────────────────────────────────
# Grand Total
# ─────────────────────────────────────────────────────────────────────────────
GRAND_END=$(date +%s%N)
GRAND_MS=$(( (GRAND_END - GRAND_START) / 1000000 ))
GRAND_SEC=$(echo "scale=1; $GRAND_MS / 1000" | bc)

echo ""
echo "============================================================"
echo "  TOTAL manual installation time: ${GRAND_SEC}s  (${GRAND_MS} ms)"
echo "  CSV results written to : $RESULTS_CSV"
echo "  Full log written to    : $LOG_FILE"
echo "============================================================"

# Write summary row for plotting
echo "TOTAL,$GRAND_MS" >> "$RESULTS_CSV"
echo "$GRAND_MS" > "$WORK_DIR/manual_total_ms.txt"
echo "Done. Results: $RESULTS_CSV"
