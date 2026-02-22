# syntax=docker/dockerfile:1
# ColdCase — all-in-one forensics container image.
# Provides every tool that ColdCase wraps, so the Go binary can fall back
# to this image when tools are not installed on the host.

# ─── Stage 1: system packages ──────────────────────────────────────────────────
FROM ubuntu:24.04 AS system

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Core runtimes
    python3 python3-pip python3-venv \
    perl \
    ruby ruby-dev \
    git curl wget ca-certificates \
    # Network forensics
    tshark tcpdump ngrep tcpflow tcpreplay tcpstat argus \
    # File carving & recovery
    foremost scalpel testdisk \
    # Hashing
    ssdeep md5deep \
    # Windows / filesystem tools
    sleuthkit ntfs-3g \
    # Metadata & binary analysis
    exiftool binwalk \
    mediainfo \
    binutils \
    xxd \
    # Steganography
    steghide \
    # Malware
    yara \
    # Mobile
    adb \
    libimobiledevice-utils \
    # Misc
    p0f \
    file \
    safecopy \
    mono-complete \
    && rm -rf /var/lib/apt/lists/*

# Install pcapfix (build from source — not in apt)
RUN apt-get update && apt-get install -y --no-install-recommends g++ libpcap-dev \
    && git clone --depth=1 https://github.com/Rup0rt/pcapfix.git /tmp/pcapfix \
    && cd /tmp/pcapfix && make && mv pcapfix /usr/local/bin/ \
    && rm -rf /tmp/pcapfix \
    && apt-get purge -y g++ libpcap-dev && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

# ─── Stage 2: Python tools ─────────────────────────────────────────────────────
FROM system AS python-tools

# Use a venv to avoid system pip conflicts on Ubuntu 24.04
RUN python3 -m venv /opt/coldcase-venv
ENV PATH="/opt/coldcase-venv/bin:$PATH"

RUN pip install --no-cache-dir \
    # Timeline analysis
    plaso \
    # Windows event log timeline
    hayabusa-python || true \
    # Timeliner (mactime rewrite)
    timeliner \
    # FLARE FLOSS — string deobfuscation
    flare-floss \
    # Mandiant CAPA
    flare-capa \
    # Mobile forensics
    alexibrignoni-aleapp || true \
    # Windows registry
    regrippy \
    analyzeMFT \
    # VT CLI (Python wrapper)
    vt-py \
    indxparse \
    python-registry \
    && true   # don't fail on optional packages

# Install zsteg (Ruby gem)
RUN gem install zsteg --no-document 2>/dev/null || true

# Install bulk_extractor if not in apt
RUN apt-get update && apt-get install -y --no-install-recommends bulk-extractor 2>/dev/null \
    || (git clone --depth=1 https://github.com/simsong/bulk_extractor.git /tmp/be \
    && cd /tmp/be && apt-get install -y --no-install-recommends \
    autoconf automake libtool libssl-dev libewf-dev libafflib-dev \
    && ./bootstrap.sh && ./configure --quiet && make -j"$(nproc)" \
    && mv src/bulk_extractor /usr/local/bin/ && rm -rf /tmp/be) \
    ; rm -rf /var/lib/apt/lists/*

# ddrescue
RUN apt-get update && apt-get install -y --no-install-recommends gddrescue \
    && rm -rf /var/lib/apt/lists/*

# ─── Stage 3: Volatility3 + DidierStevensSuite (from repo) ────────────────────
FROM python-tools AS final

WORKDIR /coldcase

# Copy bundled Python tool suites from the host build context.
COPY volatility3/ ./volatility3/
COPY DidierStevensSuite/ ./DidierStevensSuite/

ENV PATH="/opt/coldcase-venv/bin:$PATH"
RUN pip install --no-cache-dir -e "./volatility3[full]" 2>/dev/null || true

# Hayabusa binary (Rust, released as a static binary)
RUN ARCH="$(uname -m)" \
    && if [ "$ARCH" = "x86_64" ]; then \
    curl -sSL \
    "https://github.com/Yamato-Security/hayabusa/releases/latest/download/hayabusa-linux-x86_64.tar.gz" \
    | tar -xz -C /usr/local/bin --wildcards '*/hayabusa' --strip-components=1 2>/dev/null || true; \
    fi

# Chainsaw binary (Rust, static)
RUN curl -sSL \
    "https://github.com/WithSecureLabs/chainsaw/releases/latest/download/chainsaw_x86_64-unknown-linux-gnu.tar.gz" \
    | tar -xz -C /usr/local/bin chainsaw 2>/dev/null || true

# evtx_dump (Python)
RUN pip install --no-cache-dir python-evtx 2>/dev/null || true

# wavsteg
RUN pip install --no-cache-dir wavsteg 2>/dev/null || true

# ssdeep Python bindings (already have ssdeep binary above)
RUN pip install --no-cache-dir ssdeep 2>/dev/null || true

# Set up a data mount point for evidence files.
RUN mkdir -p /data
WORKDIR /data

LABEL org.opencontainers.image.title="ColdCase"
LABEL org.opencontainers.image.description="All-in-one digital forensics container"
LABEL org.opencontainers.image.source="https://github.com/Fakechippies/ColdCase"

CMD ["/bin/bash"]
