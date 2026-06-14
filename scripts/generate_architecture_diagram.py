#!/usr/bin/env python3
"""
Generate a vertical IEEE-style architecture diagram for ColdCase.
"""

from __future__ import annotations

import os

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt
from matplotlib.patches import FancyArrowPatch, FancyBboxPatch, Rectangle


SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_DIR = os.path.dirname(SCRIPT_DIR)
OUTPUT_DIR = os.path.join(PROJECT_DIR, "images")
os.makedirs(OUTPUT_DIR, exist_ok=True)


plt.rcParams.update(
    {
        "figure.facecolor": "white",
        "axes.facecolor": "white",
        "savefig.facecolor": "white",
        "savefig.dpi": 300,
        "font.family": "DejaVu Sans",
        "font.size": 9,
        "pdf.fonttype": 42,
        "ps.fonttype": 42,
    }
)


def add_box(ax, x, y, w, h, title, *, fill="0.92", rounded=True, size=10, lw=1.0):
    if rounded:
        patch = FancyBboxPatch(
            (x, y),
            w,
            h,
            boxstyle="round,pad=0.02,rounding_size=0.04",
            linewidth=lw,
            edgecolor="black",
            facecolor=fill,
            zorder=2,
        )
    else:
        patch = Rectangle((x, y), w, h, linewidth=lw, edgecolor="black", facecolor=fill, zorder=2)
    ax.add_patch(patch)
    ax.text(x + w / 2, y + h / 2, title, ha="center", va="center", fontsize=size, fontweight="bold", zorder=3)
    return patch


def add_arrow(ax, x1, y1, x2, y2, *, lw=1.1, dashed=False):
    patch = FancyArrowPatch(
        (x1, y1),
        (x2, y2),
        arrowstyle="-|>",
        mutation_scale=12,
        linewidth=lw,
        linestyle="--" if dashed else "-",
        color="black",
        zorder=1,
    )
    ax.add_patch(patch)


def add_band_label(ax, y, label):
    ax.text(0.7, y, label, ha="left", va="bottom", fontsize=8, fontweight="bold")


def build_figure():
    fig, ax = plt.subplots(figsize=(8.5, 13.5))
    ax.set_xlim(0, 8.5)
    ax.set_ylim(0, 13.5)
    ax.axis("off")

    ax.text(0.7, 13.0, "ColdCase System Architecture", ha="left", va="bottom", fontsize=14, fontweight="bold")

    add_band_label(ax, 12.2, "Operator Layer")
    add_box(ax, 1.0, 11.2, 2.7, 0.8, "Investigator / Analyst", fill="0.96")
    add_box(ax, 4.7, 11.2, 2.7, 0.8, "Host Environment", fill="0.90")
    add_arrow(ax, 3.7, 11.6, 4.7, 11.6)

    add_band_label(ax, 10.4, "CLI Composition Layer")
    add_box(ax, 1.0, 9.2, 2.2, 0.9, "Root CLI", fill="0.88")
    add_box(ax, 3.35, 9.2, 1.8, 0.9, "Init Registration", fill="0.93")
    add_box(ax, 5.3, 9.2, 2.1, 0.9, "Command Surface", fill="0.97")
    add_arrow(ax, 2.1, 11.2, 2.1, 10.1)
    add_arrow(ax, 4.25, 11.2, 4.25, 10.1)
    add_arrow(ax, 6.35, 11.2, 6.35, 10.1)

    add_band_label(ax, 8.3, "Adapter Layer")
    add_box(ax, 1.2, 7.1, 6.1, 1.0, "Category Adapters", fill="0.86")
    add_arrow(ax, 2.1, 9.2, 2.6, 8.1)
    add_arrow(ax, 4.25, 9.2, 4.25, 8.1)
    add_arrow(ax, 6.35, 9.2, 5.9, 8.1)

    add_box(ax, 1.2, 5.7, 2.9, 0.9, "Tool Contract", fill="0.94")
    add_box(ax, 4.4, 5.7, 2.9, 0.9, "Run Options", fill="0.98")
    add_arrow(ax, 3.2, 7.1, 2.65, 6.6)
    add_arrow(ax, 5.3, 7.1, 5.85, 6.6)

    add_band_label(ax, 4.8, "Execution Layer")
    add_box(ax, 1.0, 3.2, 2.2, 1.0, "Hybrid Runner", fill="0.74")
    add_box(ax, 3.55, 3.2, 1.7, 1.0, "Path Remapping", fill="0.90")
    add_box(ax, 5.6, 3.2, 1.9, 1.0, "Session Control", fill="0.82")
    add_arrow(ax, 2.65, 5.7, 2.1, 4.2)
    add_arrow(ax, 5.85, 5.7, 4.4, 4.2)
    add_arrow(ax, 4.25, 5.7, 6.55, 4.2)

    add_band_label(ax, 2.0, "External Tools / Storage")
    add_box(ax, 0.8, 0.9, 1.7, 0.8, "Native Binaries", fill="0.92", rounded=False, size=9)
    add_box(ax, 2.8, 0.9, 2.0, 0.8, "Container Runtime", fill="0.85", rounded=False, size=9)
    add_box(ax, 5.1, 0.9, 1.4, 0.8, "Bundled Suites", fill="0.95", rounded=False, size=9)
    add_box(ax, 6.8, 0.9, 1.0, 0.8, "Sessions", fill="0.89", rounded=False, size=9)

    add_arrow(ax, 1.8, 3.2, 1.65, 1.7)
    add_arrow(ax, 2.3, 3.2, 3.8, 1.7)
    add_arrow(ax, 4.4, 3.2, 3.8, 1.7)
    add_arrow(ax, 6.55, 3.2, 7.3, 1.7)
    add_arrow(ax, 5.6, 3.2, 5.8, 1.7, dashed=True)
    add_arrow(ax, 5.1, 1.3, 4.8, 1.3, dashed=True)

    return fig


def main():
    fig = build_figure()
    svg_path = os.path.join(OUTPUT_DIR, "coldcase_architecture_ieee.svg")
    png_path = os.path.join(OUTPUT_DIR, "coldcase_architecture_ieee.png")
    pdf_path = os.path.join(OUTPUT_DIR, "coldcase_architecture_ieee.pdf")
    fig.savefig(svg_path, bbox_inches="tight")
    fig.savefig(png_path, bbox_inches="tight")
    fig.savefig(pdf_path, bbox_inches="tight")
    plt.close(fig)
    print(f"Saved {svg_path}")
    print(f"Saved {png_path}")
    print(f"Saved {pdf_path}")


if __name__ == "__main__":
    main()
