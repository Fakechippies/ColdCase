# syntax=docker/dockerfile:1
# ColdCase — all-in-one forensics container image.
# Provides every tool that ColdCase wraps, so the Go binary can fall back
# to this image when tools are not installed on the host.

# ─── Stage 1: system packages ──────────────────────────────────────────────────
FROM ubuntu:26.04 AS system

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Core runtimes
    python3 python3-pip python3-venv \
    perl \
    ruby ruby-dev \
    git curl wget ca-certificates \
    # Network forensics
    tshark tcpdump ngrep tcpflow tcpreplay tcpstat argus-client argus-server \
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
    unzip \
    safecopy \
    mono-complete \
    && rm -rf /var/lib/apt/lists/*

# Install pcapfix (build from source — not in apt)
RUN apt-get update && apt-get install -y --no-install-recommends build-essential libpcap-dev \
    && git clone --depth=1 https://github.com/Rup0rt/pcapfix.git /tmp/pcapfix \
    && cd /tmp/pcapfix && make && mv pcapfix /usr/local/bin/ \
    && rm -rf /tmp/pcapfix \
    && apt-get purge -y build-essential libpcap-dev && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

# ─── Stage 2: Python tools ─────────────────────────────────────────────────────
FROM system AS python-tools

# Use a venv to avoid system pip conflicts on Ubuntu 24.04
RUN python3 -m venv /opt/coldcase-venv
ENV PATH="/opt/coldcase-venv/bin:$PATH"

# One package per layer: a single wrong name must not take the whole
# toolset down (this previously swallowed every install into one RUN).
RUN pip install --no-cache-dir plaso 2>/dev/null || true
RUN pip install --no-cache-dir timeliner 2>/dev/null || true
RUN pip install --no-cache-dir flare-floss 2>/dev/null || true
RUN pip install --no-cache-dir flare-capa 2>/dev/null || true
RUN pip install --no-cache-dir regrippy 2>/dev/null || true
RUN pip install --no-cache-dir analyzeMFT 2>/dev/null || true
RUN pip install --no-cache-dir vt-py 2>/dev/null || true
RUN pip install --no-cache-dir python-registry 2>/dev/null || true

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

# Hayabusa binary (Rust, zip release; extracted binary carries the version
# in its filename and expects ./rules + ./config as siblings)
RUN curl -sSL -o /tmp/hayabusa.zip \
    "https://github.com/Yamato-Security/hayabusa/releases/download/v4.5.0/hayabusa-4.5.0-lin-x64-gnu.zip" \
    && unzip -o -q /tmp/hayabusa.zip -d /opt/hayabusa \
    && printf '#!/bin/sh\ncd /opt/hayabusa && exec ./hayabusa-4.5.0-lin-x64-gnu "$@"\n' > /usr/local/bin/hayabusa \
    && chmod +x /usr/local/bin/hayabusa \
    && rm -f /tmp/hayabusa.zip || true

# Chainsaw binary (Rust, static; tarball nests it under a top-level dir)
RUN curl -sSL \
    "https://github.com/WithSecureLabs/chainsaw/releases/latest/download/chainsaw_x86_64-unknown-linux-gnu.tar.gz" \
    | tar -xz -C /usr/local/bin --wildcards '*/chainsaw' --strip-components=1 2>/dev/null || true

# evtx_dump binary (Rust; the pip 'evtx' wheel ships no CLI)
RUN curl -sSL -o /usr/local/bin/evtx_dump \
    "https://github.com/omerbenamram/evtx/releases/download/v0.15.1/evtx_dump-v0.15.1-x86_64-unknown-linux-gnu" \
    && chmod +x /usr/local/bin/evtx_dump || true

# evtx_dump binary is installed above; the pip 'evtx' wheel ships no CLI.

# wavsteg
RUN pip install --no-cache-dir wavsteg 2>/dev/null || true

# ssdeep Python bindings (already have ssdeep binary above)RUN pip install --no-cache-dir ssdeep 2>/dev/null || true

# Set up a data mount point for evidence files.
RUN mkdir -p /data
WORKDIR /data

LABEL org.opencontainers.image.title="ColdCase"
LABEL org.opencontainers.image.description="All-in-one digital forensics container"
LABEL org.opencontainers.image.source="https://github.com/Fakechippies/ColdCase"

CMD ["/bin/bash"]
