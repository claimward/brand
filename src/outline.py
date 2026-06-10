#!/usr/bin/env python3
"""Outline the claimward wordmark from Inter into real SVG paths.

Run from anywhere; reads Inter-*.woff2 next to this script and rewrites
../logo/claimward-lockup.svg.

    python3 -m venv .venv && .venv/bin/pip install "fonttools[woff]" brotli
    .venv/bin/python src/outline.py
"""
import os
from fontTools.ttLib import TTFont
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(HERE, "..", "logo", "claimward-lockup.svg")

FONT_SIZE = 30.0
LETTER_SPACING_PX = -0.5      # matches the design comp
TEAL = "#0D9488"
DARK = "#134E4A"

def word_path(woff2, text, upm_size):
    f = TTFont(os.path.join(HERE, woff2))
    upm = f["head"].unitsPerEm
    cmap = f.getBestCmap()
    gs = f.getGlyphSet()
    hmtx = f["hmtx"]
    s = upm_size / upm
    ls_units = LETTER_SPACING_PX / s
    pen = SVGPathPen(gs)
    x = 0.0
    for ch in text:
        gname = cmap[ord(ch)]
        gs[gname].draw(TransformPen(pen, (1, 0, 0, 1, x, 0)))
        x += hmtx[gname][0] + ls_units
    return pen.getCommands(), x * s, s

claim_d, claim_w, s = word_path("Inter-Medium.woff2", "claim", FONT_SIZE)
ward_d,  ward_w,  _ = word_path("Inter-Bold.woff2",   "ward",  FONT_SIZE)

TEXT_X, BASELINE = 74.0, 43.0
vb_w = round(TEXT_X + claim_w + ward_w + 6)

mark = ('  <g fill="none" stroke="%s" stroke-width="4" stroke-linecap="round" stroke-linejoin="round">\n'
        '    <path d="M14 15 H50 V33 Q50 46 32 55 Q14 46 14 33 Z"/>\n'
        '    <path d="M32 19 L25.9 22.5 L25.9 29.5 L32 33 L38.1 29.5 L38.1 22.5 Z"/>\n'
        '    <path d="M32 33 V46"/><path d="M32 42 H37"/><path d="M32 46 H38"/>\n'
        '  </g>') % TEAL

svg = f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {vb_w} 64" role="img" aria-label="claimward logo">
{mark}
  <g transform="translate({TEXT_X:.2f},{BASELINE}) scale({s:.6f},-{s:.6f})" fill="{TEAL}"><path d="{claim_d}"/></g>
  <g transform="translate({TEXT_X + claim_w:.2f},{BASELINE}) scale({s:.6f},-{s:.6f})" fill="{DARK}"><path d="{ward_d}"/></g>
</svg>
'''
open(OUT, "w").write(svg)
print(f"wrote {os.path.relpath(OUT)}  (viewBox {vb_w}x64)")
