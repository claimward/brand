package main

import (
	"fmt"
	"math"

	"github.com/go-opentype/opentype"
)

const (
	fontSize        = 30.0
	letterSpacingPx = -0.5 // matches the design comp
	teal            = "#0D9488"
	dark            = "#134E4A"
	textX           = 74.0
	baseline        = 43.0
)

// word outlines text in the font, laid out from x=0 at upmSize pixels to the
// em: its path data in font units, its advance in pixels, and the font-unit
// to pixel scale.
func word(fontData []byte, text string, upmSize float64) (d string, width, s float64, err error) {
	f, err := opentype.Parse(fontData)
	if err != nil {
		return "", 0, 0, err
	}
	gs, err := newGlyphSet(f)
	if err != nil {
		return "", 0, 0, err
	}
	s = upmSize / float64(f.UnitsPerEm())
	lsUnits := letterSpacingPx / s
	p := &svgPathPen{glyphs: gs}
	x := 0.0
	for _, r := range text {
		gid, ok := f.GlyphIndex(r)
		if !ok {
			return "", 0, 0, fmt.Errorf("no glyph for %q", r)
		}
		// Every glyph goes through a TransformPen, the first one too: that
		// is what makes its x coordinates floats ("618.0") at x=0.0.
		tp := &transformPen{out: p, t: transform{ival(1), ival(0), ival(0), ival(1), fval(x), ival(0)}}
		if err := gs.draw(uint16(gid), tp, 0); err != nil {
			return "", 0, 0, err
		}
		// x += advance + ls_units: the sum is formed first, then added.
		x += float64(f.GlyphAdvance(gid)) + lsUnits
	}
	return p.d(), x * s, s, nil
}

// lockup renders logo/claimward-lockup.svg: the key-and-shield mark, then
// "claim" in Inter Medium and "ward" in Inter Bold, outlined.
func lockup(medium, bold []byte) ([]byte, error) {
	claimD, claimW, s, err := word(medium, "claim", fontSize)
	if err != nil {
		return nil, fmt.Errorf("claim: %w", err)
	}
	wardD, wardW, _, err := word(bold, "ward", fontSize)
	if err != nil {
		return nil, fmt.Errorf("ward: %w", err)
	}
	// Python's round() on a float: half to even.
	vbW := int(math.RoundToEven(textX + claimW + wardW + 6))

	mark := fmt.Sprintf(`  <g fill="none" stroke="%s" stroke-width="4" stroke-linecap="round" stroke-linejoin="round">
    <path d="M14 15 H50 V33 Q50 46 32 55 Q14 46 14 33 Z"/>
    <path d="M32 19 L25.9 22.5 L25.9 29.5 L32 33 L38.1 29.5 L38.1 22.5 Z"/>
    <path d="M32 33 V46"/><path d="M32 42 H37"/><path d="M32 46 H38"/>
  </g>`, teal)

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d 64" role="img" aria-label="claimward logo">
%s
  <g transform="translate(%.2f,%s) scale(%.6f,-%.6f)" fill="%s"><path d="%s"/></g>
  <g transform="translate(%.2f,%s) scale(%.6f,-%.6f)" fill="%s"><path d="%s"/></g>
</svg>
`, vbW, mark,
		textX, pyFloatRepr(baseline), s, s, teal, claimD,
		textX+claimW, pyFloatRepr(baseline), s, s, dark, wardD)
	return []byte(svg), nil
}
