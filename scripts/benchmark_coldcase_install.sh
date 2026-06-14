#!/usr/bin/env bash
# =============================================================================
# ColdCase – ColdCase Installation Benchmark
# Measures the time to get ALL 118+ forensics tools via ColdCase:
#   1. Clone the repo
#   2. Build the binary  (go build)
#   3. Either:  `coldcase install`          (native install mode)
#         OR:  `coldcase install --container` (Docker hybrid mode, DEFAULT)
#
# Usage:  bash benchmark_coldcase_install.sh [--native | --container]
#   --native      Time: clone + build + native host install
#   --container   Time: clone + build + Docker image pull  (default)
# =============================================================================
set -euo pipefail

MODE="${1:---container}"
WORK_DIR="$(mktemp -d /tmp/coldcase_bench_XXXXXX)"
LOG_FILE="$WORK_DIR/install_log.txt"
RESULTS_CSV="$WORK_DIR/results.csv"
REPO_URL="https://github.com/Fakechippies/ColdCase.git"
REPO_DIR="$WORK_DIR/ColdCase"

trap 'echo; echo "Temp dir left at: $WORK_DIR"' EXIT

time_phase() {
  local label="$1"; shift
  local start end elapsed
  start=$(date +%s%N)
  eval "$*" >> "$LOG_FILE" 2>&1
  end=$(date +%s%N)
  elapsed=$(( (end - start) / 1000000 ))
  echo "$label,$elapsed" >> "$RESULTS_CSV"
  printf "  %-40s %6d ms\n" "$label" "$elapsed"
}

# ─────────────────────────────────────────────────────────────────────────────
echo "============================================================"
echo "  ColdCase  ·  ColdCase Installation Benchmark"
echo "  Mode       : $MODE"
echo "  Working dir: $WORK_DIR"
echo "  Started at : $(date)"
echo "============================================================"
echo ""
echo "phase,elapsed_ms" > "$RESULTS_CSV"

GRAND_START=$(date +%s%N)

# ─────────────────────────────────────────────────────────────────────────────
# Phase 1: Clone ColdCase repository
# ─────────────────────────────────────────────────────────────────────────────
echo "[ Phase 1 ] Clone ColdCase repository"
time_phase "git clone ColdCase" \
  "git clone --depth 1 '$REPO_URL' '$REPO_DIR'"

# ─────────────────────────────────────────────────────────────────────────────
# Phase 2: Build the ColdCase binary
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "[ Phase 2 ] Build ColdCase binary (go build)"
time_phase "go build" \
  "cd '$REPO_DIR' && go build -o bin/coldcase ./cmd/coldcase"

BINARY="$REPO_DIR/bin/coldcase"

# ─────────────────────────────────────────────────────────────────────────────
# Phase 3: Install tools
# ─────────────────────────────────────────────────────────────────────────────
echo ""
if [[ "$MODE" == "--native" ]]; then
  echo "[ Phase 3 ] Native host install  (coldcase install)"
  time_phase "coldcase install (native)" \
    "cd '$REPO_DIR' && sudo '$BINARY' install"
else
  echo "[ Phase 3 ] Docker hybrid install (coldcase install --container)"
  # Sub-time the docker pull vs the actual container setup
  time_phase "docker pull forensics image" \
    "docker pull ghcr.io/fakechippies/coldcase:latest || \
     (cd '$REPO_DIR' && '$BINARY' container build)"
  time_phase "coldcase container verify" \
    "cd '$REPO_DIR' && '$BINARY' container status"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Grand Total
# ─────────────────────────────────────────────────────────────────────────────
GRAND_END=$(date +%s%N)
GRAND_MS=$(( (GRAND_END - GRAND_START) / 1000000 ))
GRAND_SEC=$(echo "scale=1; $GRAND_MS / 1000" | bc)

echo ""
echo "============================================================"
echo "  TOTAL ColdCase installation time : ${GRAND_SEC}s  (${GRAND_MS} ms)"
echo "  CSV results written to : $RESULTS_CSV"
echo "  Full log written to    : $LOG_FILE"
echo "============================================================"

echo "TOTAL,$GRAND_MS" >> "$RESULTS_CSV"
echo "$GRAND_MS" > "$WORK_DIR/coldcase_total_ms.txt"
echo "Done. Results: $RESULTS_CSV"
