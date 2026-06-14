#!/usr/bin/env python3
"""
ColdCase – Research Diagram Generator  (IEEE B&W edition)
==========================================================
Generates publication-quality black-and-white matplotlib figures.
Designed for IEEE double-column paper format – prints cleanly in
monochrome via hatching + grayscale fills (no colour dependency).

Figures produced (images/):
  1. install_time_comparison.png
  2. tool_categories.png
  3. forensic_lifecycle.png
  4. hybrid_engine_logic.png
  5. session_states.png

Run:  python3 scripts/generate_graphs.py
"""

import os
import numpy as np
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
from matplotlib.patches import FancyBboxPatch, Circle

# ── Output directory ──────────────────────────────────────────────────────────
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJ_DIR   = os.path.dirname(SCRIPT_DIR)
OUT_DIR    = os.path.join(PROJ_DIR, "images")
os.makedirs(OUT_DIR, exist_ok=True)

# ── IEEE B&W theme ─────────────────────────────────────────────────────────────
GRAYS   = ["0.00","0.12","0.22","0.32","0.42","0.52","0.62","0.72",
           "0.80","0.86","0.90","0.94","0.18","0.28"]
HATCHES = ["", "///", "...", "xxx", "+++", "\\\\\\", "|||", "---",
           "ooo", "///", "...", "xxx", "+++", "\\\\\\"]
BW_FILLS = list(zip(GRAYS, HATCHES))   # (face_gray, hatch) pairs

plt.rcParams.update({
    "figure.facecolor":  "white",
    "axes.facecolor":    "white",
    "axes.edgecolor":    "black",
    "axes.labelcolor":   "black",
    "xtick.color":       "black",
    "ytick.color":       "black",
    "text.color":        "black",
    "font.family":       "DejaVu Sans",
    "font.size":         10,
    "axes.titlesize":    12,
    "axes.titleweight":  "bold",
    "figure.dpi":        150,
    "savefig.dpi":       300,      # 300 dpi – IEEE minimum
    "savefig.bbox":      "tight",
    "savefig.facecolor": "white",
    "hatch.linewidth":   0.8,
})

def save(fig, name):
    path = os.path.join(OUT_DIR, name)
    fig.savefig(path, facecolor="white")
    plt.close(fig)
    print(f"  Saved → {path}")


# ══════════════════════════════════════════════════════════════════════════════
# 1.  Installation Time Comparison  (4-bar, single column)
# ══════════════════════════════════════════════════════════════════════════════
def fig_install_comparison():
    manual_total   = 405.0   # sum of all apt/pip/cargo phases
    cc_src         = 8.3 + 0.716   # git clone + go build
    cc_docker      = 92.0          # docker pull forensics image
    cc_total       = cc_src + cc_docker

    bars = [
        ("Manual\nInstall",          manual_total, "0.20", "///"),
        ("ColdCase\nSource Setup",   cc_src,        "0.60", ""),
        ("ColdCase\nDocker Pull",    cc_docker,     "0.50", "..."),
        ("ColdCase\nTotal",          cc_total,      "0.35", ""),
    ]

    fig, ax = plt.subplots(figsize=(6.5, 4.5))

    x_pos = np.arange(len(bars))
    for i, (label, val, gray, hatch) in enumerate(bars):
        ax.bar(i, val, color=gray, edgecolor="black", linewidth=1.0,
               width=0.55, hatch=hatch, zorder=3)
        ax.text(i, val + 4,
                f"{val:.0f}s\n({val/60:.1f} min)",
                ha="center", va="bottom", fontsize=9, fontweight="bold")

    # speedup bracket
    speedup = manual_total / cc_total
    ax.annotate(
        f"{speedup:.1f}× faster",
        xy=(3, cc_total + 10),
        xytext=(1.5, (manual_total + cc_total) / 2),
        fontsize=10, fontweight="bold", ha="center",
        arrowprops=dict(arrowstyle="-|>", color="black", lw=1.4,
                        connectionstyle="arc3,rad=0.3"),
    )

    ax.set_xticks(x_pos)
    ax.set_xticklabels([b[0] for b in bars], fontsize=9)
    ax.set_ylabel("Time (seconds)", fontsize=10)
    ax.set_title("Installation Time Benchmark  (118+ Forensic Tools)",
                 fontsize=11, fontweight="bold", pad=10)
    ax.set_ylim(0, manual_total * 1.30)
    ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda v, _: f"{v:.0f}s"))
    ax.spines[["top", "right"]].set_visible(False)
    ax.grid(axis="y", color="0.80", linestyle="--", linewidth=0.7, zorder=0)

    save(fig, "install_time_comparison.png")


# ══════════════════════════════════════════════════════════════════════════════
# 2.  Tool Categories Donut
# ══════════════════════════════════════════════════════════════════════════════
def fig_tool_categories():
    categories = {
        "Volatility3 (Memory)":      31,
        "DidierStevens Suite":        15,
        "Network Forensics":          11,
        "Timeline & Log Analysis":     9,
        "Windows Artifacts":           7,
        "File Carving & Recovery":     7,
        "System Utilities":            6,
        "Malware & Patterns":          5,
        "Steganography & Media":       5,
        "Sleuth Kit":                  5,
        "Cryptography & Sessions":     5,
        "Mobile Forensics":            4,
        "Hashing & Verification":      4,
        "Plaso (Timeline)":            4,
    }
    labels = list(categories.keys())
    sizes  = list(categories.values())
    total  = sum(sizes)

    gray_seq = [str(round(0.05 + i * (0.85 / len(sizes)), 2)) for i in range(len(sizes))]

    fig, ax = plt.subplots(figsize=(11, 7))
    wedges, _ = ax.pie(
        sizes,
        colors=gray_seq,
        wedgeprops=dict(width=0.52, edgecolor="white", linewidth=1.2),
        startangle=90,
        hatch=[HATCHES[i % len(HATCHES)] for i in range(len(sizes))],
    )
    for w in wedges:
        w.set_linewidth(0.6)

    # Centre label
    ax.text(0, 0.10, str(total), ha="center", va="center",
            fontsize=44, fontweight="bold", color="black")
    ax.text(0, -0.20, "total tools", ha="center", va="center",
            fontsize=12, color="0.30")

    # Legend
    patches = [
        mpatches.Patch(facecolor=gray_seq[i],
                       edgecolor="black", linewidth=0.7,
                       hatch=HATCHES[i % len(HATCHES)],
                       label=f"{labels[i]}  ({sizes[i]})")
        for i in range(len(labels))
    ]
    ax.legend(handles=patches, loc="center left",
              bbox_to_anchor=(1.02, 0.5), frameon=True,
              framealpha=1.0, edgecolor="black",
              fontsize=9)
    ax.set_title("ColdCase Tool Categories  (118+ integrated tools)",
                 fontsize=12, fontweight="bold", pad=16)
    save(fig, "tool_categories.png")


# ══════════════════════════════════════════════════════════════════════════════
# 3.  Forensic Lifecycle  (vertical flow)
# ══════════════════════════════════════════════════════════════════════════════
def fig_forensic_lifecycle():
    fig, ax = plt.subplots(figsize=(8, 12))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 14)
    ax.axis("off")
    ax.set_facecolor("white")

    steps = [
        ("1. Evidence Acquisition",
         "dd, ddrescue, safecopy — disk images & memory dumps",
         "0.92", 13.0),
        ("2. Session Initialisation",
         "coldcase session start — Ed25519-signed audit log",
         "0.82", 11.0),
        ("3. Integrity Verification",
         "hashdeep, md5deep, ssdeep — baseline hash manifest",
         "0.70", 9.0),
        ("4. Analysis & Examination",
         "Volatility3, Sleuth Kit, YARA, tshark, plaso",
         "0.58", 7.0),
        ("5. Artefact Recovery",
         "foremost, scalpel, photorec, bulk_extractor, binwalk",
         "0.45", 5.0),
        ("6. Chain-of-Custody Lock",
         "coldcase session lock / seal — immutable audit trail",
         "0.32", 3.0),
        ("7. Report Export",
         "coldcase session export — signed HTML / JSON report",
         "0.18", 1.0),
    ]

    BOX_W, BOX_H = 7.0, 0.85
    X_CTR = 5.0

    for title, detail, gray, y in steps:
        patch = FancyBboxPatch(
            (X_CTR - BOX_W/2, y - BOX_H/2), BOX_W, BOX_H,
            boxstyle="round,pad=0.10", linewidth=1.4,
            edgecolor="black", facecolor=gray, zorder=3)
        ax.add_patch(patch)

        text_col = "white" if float(gray) < 0.55 else "black"
        ax.text(X_CTR, y + 0.17, title,
                ha="center", va="center", fontsize=11, fontweight="bold",
                color=text_col, zorder=4)
        ax.text(X_CTR, y - 0.21, detail,
                ha="center", va="center", fontsize=8.5,
                color=text_col if float(gray) < 0.55 else "0.30",
                zorder=4)

        if y > 1.0:
            ax.annotate("", xy=(X_CTR, y - 2.0 + BOX_H/2 + 0.06),
                        xytext=(X_CTR, y - BOX_H/2 - 0.06),
                        arrowprops=dict(arrowstyle="-|>", color="black",
                                        lw=1.6, mutation_scale=16), zorder=2)

    ax.set_title("ColdCase Forensic Investigation Lifecycle",
                 fontsize=13, fontweight="bold", pad=12, y=0.98)
    save(fig, "forensic_lifecycle.png")


# ══════════════════════════════════════════════════════════════════════════════
# 4.  Hybrid Execution Engine  (vertical decision flow)
# ══════════════════════════════════════════════════════════════════════════════
def fig_hybrid_engine():
    fig, ax = plt.subplots(figsize=(7, 13))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 16)
    ax.axis("off")
    ax.set_facecolor("white")

    # node styles: (face_gray, text_color, line_width)
    STYLE = {
        "entry":   ("0.92", "black", 1.2),
        "process": ("0.82", "black", 1.2),
        "decide":  ("0.55", "white", 1.6),
        "native":  ("0.18", "white", 1.6),
        "docker":  ("0.45", "white", 1.4),
        "stream":  ("0.72", "black", 1.2),
        "session": ("0.62", "black", 1.2),
        "sign":    ("0.35", "white", 1.4),
        "exit":    ("0.92", "black", 1.0),
    }

    def rect(x, y, w, h, text, sty):
        fg, tc, lw = STYLE[sty]
        p = FancyBboxPatch((x-w/2, y-h/2), w, h,
                           boxstyle="round,pad=0.08",
                           facecolor=fg, edgecolor="black",
                           linewidth=lw, zorder=3)
        ax.add_patch(p)
        ax.text(x, y, text, ha="center", va="center",
                fontsize=9.5, fontweight="bold", color=tc,
                multialignment="center", zorder=4)

    def diamond(x, y, w, h, text):
        fg, tc, lw = STYLE["decide"]
        pts = np.array([[x, y+h/2],[x+w/2, y],[x, y-h/2],[x-w/2, y]])
        p = plt.Polygon(pts, closed=True, facecolor=fg,
                        edgecolor="black", linewidth=lw, zorder=3)
        ax.add_patch(p)
        ax.text(x, y, text, ha="center", va="center",
                fontsize=9.5, fontweight="bold", color=tc,
                multialignment="center", zorder=4)

    def arr(x1, y1, x2, y2, label=""):
        ax.annotate("", xy=(x2, y2), xytext=(x1, y1),
                    arrowprops=dict(arrowstyle="-|>", color="black",
                                    lw=1.4, mutation_scale=14), zorder=2)
        if label:
            mx, my = (x1+x2)/2, (y1+y2)/2
            ax.text(mx+0.18, my, label, fontsize=9, fontweight="bold")

    # ── Nodes ────────────────────────────────────────────────────────────────
    rect(5, 15.2, 5.5, 0.80, "Investigator runs:\ncoldcase <tool> [args]", "entry")
    rect(5, 13.4, 4.2, 0.80, "Parse arguments\nDetect host file paths",   "process")
    diamond(5, 11.5, 4.0, 1.30, "Binary on\nPATH?")
    rect(2, 9.2,  3.8, 0.80, "Run natively\n(zero overhead)",             "native")
    rect(8, 9.2,  3.8, 0.80, "Remap paths\nbuild -v mounts",              "docker")
    rect(8, 7.4,  3.8, 0.80, "docker run --rm\nforensics-img",            "docker")
    rect(5, 5.5,  4.0, 0.80, "Stream stdout/stderr\nto terminal",         "stream")
    rect(5, 3.7,  4.5, 0.80, "Session active?\nLog command + hashes",     "session")
    rect(5, 1.9,  4.0, 0.80, "Ed25519 sign\naudit entry",                 "sign")
    rect(5, 0.5,  4.5, 0.60, "Return exit code to shell",                 "exit")

    # ── Arrows ───────────────────────────────────────────────────────────────
    arr(5, 14.80, 5, 13.80)
    arr(5, 13.00, 5, 12.15)
    arr(3.00, 11.5, 2, 9.60, "YES")
    arr(7.00, 11.5, 8, 9.60, "NO")
    arr(2, 8.80, 5, 5.90)
    arr(8, 8.80, 8, 7.80)
    arr(8, 7.00, 5, 5.90)
    arr(5, 5.10, 5, 4.10)
    arr(5, 3.30, 5, 2.30)
    arr(5, 1.50, 5, 0.80)

    ax.set_title("ColdCase Hybrid Execution Engine",
                 fontsize=13, fontweight="bold", pad=10, y=0.985)
    save(fig, "hybrid_engine_logic.png")


# ══════════════════════════════════════════════════════════════════════════════
# 5.  Session State Machine
# ══════════════════════════════════════════════════════════════════════════════
def fig_session_states():
    fig, ax = plt.subplots(figsize=(11, 6.5))
    ax.set_xlim(0, 12)
    ax.set_ylim(0, 8)
    ax.axis("off")
    ax.set_facecolor("white")

    # (x, y, face_gray, text_col, double_circle)
    state_cfg = {
        "START":    (1.5,  4.0, "0.88", "black", False),
        "UNLOCKED": (4.5,  6.2, "0.75", "black", False),
        "LOCKED":   (7.5,  6.2, "0.45", "white", False),
        "SEALED":   (10.5, 4.0, "0.15", "white", True),
        "EXPORTED": (6.0,  1.5, "0.60", "black", False),
    }
    state_sub = {
        "UNLOCKED": "Tools log freely",
        "LOCKED":   "No new entries",
        "SEALED":   "Immutable forever",
    }

    R = 0.95
    for name, (x, y, fg, tc, dbl) in state_cfg.items():
        if dbl:
            ax.add_patch(Circle((x, y), R+0.13, facecolor="none",
                                edgecolor="black", linewidth=2.2, zorder=3))
        ax.add_patch(Circle((x, y), R, facecolor=fg,
                            edgecolor="black", linewidth=1.8, zorder=4))
        dy = 0.14 if name in state_sub else 0
        ax.text(x, y + dy, name, ha="center", va="center",
                fontsize=10, fontweight="bold", color=tc, zorder=5)
        if name in state_sub:
            ax.text(x, y - 0.24, state_sub[name], ha="center", va="center",
                    fontsize=7.5, color=tc, style="italic", zorder=5)

    def tr(xs, ys, xt, yt, label, rad=0.0, lx=0.0, ly=0.25):
        ax.annotate("", xy=(xt, yt), xytext=(xs, ys),
                    arrowprops=dict(arrowstyle="-|>", color="black", lw=1.5,
                                    mutation_scale=15,
                                    connectionstyle=f"arc3,rad={rad}"), zorder=2)
        mx = (xs+xt)/2 + lx
        my = (ys+yt)/2 + ly
        ax.text(mx, my, label, fontsize=8.5, ha="center",
                style="italic")

    tr(1.5+R, 4.0,   4.5-R, 6.2, "session start --sign", rad= 0.18, lx=0, ly=0.3)
    tr(4.5+R, 6.2,   7.5-R, 6.2, "session lock",         rad= 0.0,  lx=0, ly=0.3)
    tr(7.5+R, 6.2,  10.5-R, 4.0, "session seal",         rad=-0.18, lx=0.2, ly=0.2)
    tr(4.5,   6.2-R, 6.0,   1.5+R,"session export",      rad= 0.22, lx=-0.8, ly=0)
    tr(7.5,   6.2-R, 6.0,   1.5+R,"session export",      rad=-0.22, lx= 0.8, ly=0)

    # self-loop verify on UNLOCKED
    ax.annotate("", xy=(4.0, 6.2+R), xytext=(5.0, 6.2+R),
                arrowprops=dict(arrowstyle="-|>", color="black", lw=1.3,
                                connectionstyle="arc3,rad=-1.3"), zorder=2)
    ax.text(4.5, 7.55, "session verify", fontsize=8.5, ha="center", style="italic")

    # legend
    legend_entries = [
        ("UNLOCKED", "0.75", "Tools can run / log commands"),
        ("LOCKED",   "0.45", "Read-only; no new log entries"),
        ("SEALED",   "0.15", "Permanent; irrevocable"),
        ("EXPORTED", "0.60", "Signed HTML / JSON report"),
    ]
    patches = [
        mpatches.Patch(facecolor=g, edgecolor="black", linewidth=0.7,
                       label=f"{n} – {d}")
        for n, g, d in legend_entries
    ]
    ax.legend(handles=patches, loc="lower right",
              bbox_to_anchor=(1.0, 0.01), frameon=True,
              framealpha=1.0, edgecolor="black", fontsize=8.5)

    ax.set_title("ColdCase Forensic Session State Machine",
                 fontsize=13, fontweight="bold", pad=10, y=0.97)
    save(fig, "session_states.png")


# ══════════════════════════════════════════════════════════════════════════════
# Main
# ══════════════════════════════════════════════════════════════════════════════
if __name__ == "__main__":
    print("ColdCase – generating IEEE B&W diagrams …\n")

    print("[1/5] Installation time comparison")
    fig_install_comparison()

    print("[2/5] Tool categories donut")
    fig_tool_categories()

    print("[3/5] Forensic lifecycle")
    fig_forensic_lifecycle()

    print("[4/5] Hybrid engine logic")
    fig_hybrid_engine()

    print("[5/5] Session state machine")
    fig_session_states()

    print(f"\nAll figures saved to → {OUT_DIR}/")
