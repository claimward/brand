package main

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/go-opentype/opentype"
)

// glyphSet reads TrueType outlines point by point, on- and off-curve flags
// included. go-opentype decodes the WOFF2 container (brotli, and the glyf,
// loca and hmtx transforms) and hands back the reconstructed tables; what it
// does not export is the raw point list, and its GlyphOutline synthesises the
// implied on-curve points before any translation, which is one rounding away
// from what the committed SVG holds. So the glyf records are read here.
type glyphSet struct {
	font *opentype.Font
	glyf []byte
	loca []uint32 // numGlyphs+1 offsets into glyf
	lsb  []int16  // per glyph, from hmtx
}

func be16(b []byte) uint16 { return binary.BigEndian.Uint16(b) }

func newGlyphSet(f *opentype.Font) (*glyphSet, error) {
	need := func(tag string) ([]byte, error) {
		b, ok := f.Table(tag)
		if !ok {
			return nil, fmt.Errorf("font has no %q table", tag)
		}
		return b, nil
	}
	var tabs [5][]byte
	for i, tag := range []string{"head", "loca", "glyf", "hhea", "hmtx"} {
		b, err := need(tag)
		if err != nil {
			return nil, err
		}
		tabs[i] = b
	}
	head, loca, glyf, hhea, hmtx := tabs[0], tabs[1], tabs[2], tabs[3], tabs[4]
	if len(head) < 54 || len(hhea) < 36 {
		return nil, errors.New("head or hhea table is truncated")
	}
	n := f.NumGlyphs()
	g := &glyphSet{font: f, glyf: glyf, loca: make([]uint32, n+1), lsb: make([]int16, n)}

	long := be16(head[50:]) != 0 // indexToLocFormat
	for i := 0; i <= n; i++ {
		switch {
		case long && len(loca) >= 4*i+4:
			g.loca[i] = binary.BigEndian.Uint32(loca[4*i:])
		case !long && len(loca) >= 2*i+2:
			g.loca[i] = 2 * uint32(be16(loca[2*i:]))
		default:
			return nil, errors.New("loca table is truncated")
		}
	}

	nh := int(be16(hhea[34:])) // numberOfHMetrics
	if nh == 0 || nh > n || len(hmtx) < 4*nh+2*(n-nh) {
		return nil, errors.New("hmtx table is truncated")
	}
	for i := 0; i < n; i++ {
		if i < nh {
			g.lsb[i] = int16(be16(hmtx[4*i+2:]))
		} else {
			g.lsb[i] = int16(be16(hmtx[4*nh+2*(i-nh):]))
		}
	}
	return g, nil
}

// TrueType glyph flags.
const (
	flagOnCurve  = 0x01
	flagXShort   = 0x02
	flagYShort   = 0x04
	flagRepeat   = 0x08
	flagXSame    = 0x10
	flagYSame    = 0x20
	flagCubic    = 0x80 // the glyf v1 extension; Inter has none
	argsAreWords = 0x0001
	argsAreXY    = 0x0002
	hasScale     = 0x0008
	moreComps    = 0x0020
	hasXYScale   = 0x0040
	hasTwoByTwo  = 0x0080
)

// draw is fontTools' TTGlyph.draw followed by glyf Glyph.draw. depth is the
// component nesting depth: the hmtx left side bearing shifts a glyph only at
// the top level, as fontTools' glyph set does.
func (g *glyphSet) draw(gid uint16, p pen, depth int) error {
	if int(gid) >= len(g.lsb) {
		return fmt.Errorf("glyph %d out of range", gid)
	}
	start, end := g.loca[gid], g.loca[gid+1]
	if start > end || int(end) > len(g.glyf) {
		return fmt.Errorf("glyph %d: bad loca entry", gid)
	}
	b := g.glyf[start:end]
	if len(b) == 0 {
		return nil // an empty glyph, a space
	}
	if len(b) < 10 {
		return fmt.Errorf("glyph %d: truncated header", gid)
	}
	nContours := int16(be16(b))
	xMin := int(int16(be16(b[2:])))
	if nContours < 0 {
		return g.drawComposite(gid, b[10:], p)
	}
	offset := 0
	if depth == 0 {
		offset = int(g.lsb[gid]) - xMin
	}
	return drawSimple(gid, b[10:], int(nContours), offset, p)
}

func drawSimple(gid uint16, b []byte, nContours, offset int, p pen) error {
	bad := fmt.Errorf("glyph %d: truncated outline", gid)
	if len(b) < 2*nContours+2 {
		return bad
	}
	endPts := make([]int, nContours)
	for i := range endPts {
		endPts[i] = int(be16(b[2*i:]))
	}
	b = b[2*nContours:]
	nPoints := 0
	if nContours > 0 {
		nPoints = endPts[nContours-1] + 1
	}
	insLen := int(be16(b))
	b = b[2:]
	if len(b) < insLen {
		return bad
	}
	b = b[insLen:]

	flags := make([]byte, 0, nPoints)
	for len(flags) < nPoints {
		if len(b) == 0 {
			return bad
		}
		f := b[0]
		b = b[1:]
		flags = append(flags, f)
		if f&flagRepeat != 0 {
			if len(b) == 0 {
				return bad
			}
			for r := int(b[0]); r > 0 && len(flags) < nPoints; r-- {
				flags = append(flags, f)
			}
			b = b[1:]
		}
	}
	coords := func(short, same byte) ([]int, error) {
		out := make([]int, nPoints)
		v := 0
		for i, f := range flags {
			switch {
			case f&short != 0:
				if len(b) < 1 {
					return nil, bad
				}
				d := int(b[0])
				b = b[1:]
				if f&same == 0 {
					d = -d
				}
				v += d
			case f&same == 0:
				if len(b) < 2 {
					return nil, bad
				}
				v += int(int16(be16(b)))
				b = b[2:]
			}
			out[i] = v
		}
		return out, nil
	}
	xs, err := coords(flagXShort, flagXSame)
	if err != nil {
		return err
	}
	ys, err := coords(flagYShort, flagYSame)
	if err != nil {
		return err
	}

	first := 0
	for _, last := range endPts {
		end := last + 1
		if end <= first || end > nPoints {
			return fmt.Errorf("glyph %d: contour end points out of order", gid)
		}
		var contour []point
		var on []bool
		for i := first; i < end; i++ {
			if flags[i]&flagCubic != 0 {
				return fmt.Errorf("glyph %d: cubic glyf outlines are not supported", gid)
			}
			contour = append(contour, point{ival(xs[i] + offset), ival(ys[i])})
			on = append(on, flags[i]&flagOnCurve != 0)
		}
		first = end
		drawContour(contour, on, p)
	}
	return nil
}

// drawContour is the quadratic half of fontTools' glyf Glyph.draw contour
// loop: rotate so that the contour ends on its first on-curve point, move
// there, and emit each run of off-curve points with the on-curve point that
// ends it. A closing lineTo back to the start is left to closePath.
func drawContour(contour []point, on []bool, p pen) {
	firstOn := -1
	for i, o := range on {
		if o {
			firstOn = i
			break
		}
	}
	if firstOn < 0 {
		p.qCurveTo(contour, true)
		p.closePath()
		return
	}
	k := firstOn + 1
	contour = append(append([]point(nil), contour[k:]...), contour[:k]...)
	on = append(append([]bool(nil), on[k:]...), on[:k]...)
	p.moveTo(contour[len(contour)-1])
	for len(contour) > 0 {
		next := 0
		for !on[next] {
			next++
		}
		next++
		if next == 1 {
			if len(contour) > 1 {
				p.lineTo(contour[0])
			}
		} else {
			p.qCurveTo(contour[:next], false)
		}
		contour, on = contour[next:], on[next:]
	}
	p.closePath()
}

// drawComposite hands each component to the pen with its transformation, as
// glyf Glyph.draw does; the pen decomposes. Only the offset-only components
// the wordmark uses are supported (Inter's "i" is dotlessi + dotaccent):
// a scaled one would need F2Dot14 arithmetic matched to fontTools' too.
func (g *glyphSet) drawComposite(gid uint16, b []byte, p pen) error {
	for {
		if len(b) < 4 {
			return fmt.Errorf("glyph %d: truncated component", gid)
		}
		flags, cgid := be16(b), be16(b[2:])
		b = b[4:]
		if flags&argsAreXY == 0 {
			return fmt.Errorf("glyph %d: point-matched components are not supported", gid)
		}
		if flags&(hasScale|hasXYScale|hasTwoByTwo) != 0 {
			return fmt.Errorf("glyph %d: scaled components are not supported", gid)
		}
		var dx, dy int
		if flags&argsAreWords != 0 {
			if len(b) < 4 {
				return fmt.Errorf("glyph %d: truncated component", gid)
			}
			dx, dy = int(int16(be16(b))), int(int16(be16(b[2:])))
			b = b[4:]
		} else {
			if len(b) < 2 {
				return fmt.Errorf("glyph %d: truncated component", gid)
			}
			dx, dy = int(int8(b[0])), int(int8(b[1]))
			b = b[2:]
		}
		t := transform{ival(1), ival(0), ival(0), ival(1), ival(dx), ival(dy)}
		if err := p.addComponent(cgid, t); err != nil {
			return err
		}
		if flags&moreComps == 0 {
			return nil
		}
	}
}
