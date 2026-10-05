package main

import "strings"

// The pens below reproduce, operation for operation, the fontTools pipeline
// the wordmark was first generated with (fontTools 4.x: TTGlyph.draw into a
// TransformPen into an SVGPathPen). Matching it is not about taste: the
// committed logo/claimward-lockup.svg is the judge, and its path data spells
// out every float the way that pipeline computed and printed it. Where Go
// would naturally do something else (compute an implied point before rather
// than after the translation, say), the result can differ in the last bit,
// and repr() prints that bit.

type point [2]num

// transform is an affine matrix in fontTools' order: xx, xy, yx, yy, dx, dy.
type transform [6]num

var identity = transform{ival(1), ival(0), ival(0), ival(1), ival(0), ival(0)}

// apply is fontTools' Transform.transformPoint, with Python's left-to-right
// evaluation: (xx*x + yx*y) + dx.
func (t transform) apply(p point) point {
	x, y := p[0], p[1]
	return point{
		t[0].mul(x).add(t[2].mul(y)).add(t[4]),
		t[1].mul(x).add(t[3].mul(y)).add(t[5]),
	}
}

// compose is fontTools' Transform.transform(other): other applied first,
// then t.
func (t transform) compose(o transform) transform {
	xx1, xy1, yx1, yy1, dx1, dy1 := o[0], o[1], o[2], o[3], o[4], o[5]
	xx2, xy2, yx2, yy2, dx2, dy2 := t[0], t[1], t[2], t[3], t[4], t[5]
	return transform{
		xx1.mul(xx2).add(xy1.mul(yx2)),
		xx1.mul(xy2).add(xy1.mul(yy2)),
		yx1.mul(xx2).add(yy1.mul(yx2)),
		yx1.mul(xy2).add(yy1.mul(yy2)),
		xx2.mul(dx1).add(yx2.mul(dy1)).add(dx2),
		xy2.mul(dx1).add(yy2.mul(dy1)).add(dy2),
	}
}

func (t transform) eq(o transform) bool {
	for i := range t {
		if !t[i].eq(o[i]) {
			return false
		}
	}
	return true
}

// pen is the fontTools segment-pen protocol. In qCurveTo, noOnCurve=true stands
// for Python's trailing None: a contour with no on-curve point at all.
type pen interface {
	moveTo(p point)
	lineTo(p point)
	qCurveTo(pts []point, noOnCurve bool)
	closePath()
	addComponent(gid uint16, t transform) error
}

// transformPen is fontTools' TransformPen.
type transformPen struct {
	out pen
	t   transform
}

func (p *transformPen) moveTo(pt point) { p.out.moveTo(p.t.apply(pt)) }
func (p *transformPen) lineTo(pt point) { p.out.lineTo(p.t.apply(pt)) }
func (p *transformPen) closePath()      { p.out.closePath() }

func (p *transformPen) qCurveTo(pts []point, noOnCurve bool) {
	tp := make([]point, len(pts))
	for i, q := range pts {
		tp[i] = p.t.apply(q)
	}
	p.out.qCurveTo(tp, noOnCurve)
}

func (p *transformPen) addComponent(gid uint16, t transform) error {
	return p.out.addComponent(gid, p.t.compose(t))
}

// svgPathPen is fontTools' SVGPathPen, including what it inherits from
// BasePen (implied on-curve points) and DecomposingPen (components).
//
// Its quirks are kept, because the committed SVG has them: a lineTo straight
// after a moveTo is written without an "L" (" 31.0 1118"), and since that
// does not update the last command, every further plain lineTo is written the
// same way until an H, V, Q or Z intervenes.
type svgPathPen struct {
	glyphs   *glyphSet
	depth    int
	commands []string
	lastCmd  byte // 0 for None
	last     *point
}

func ntos(p point) string { return p[0].String() + " " + p[1].String() }

func (p *svgPathPen) moveTo(pt point) {
	if p.lastCmd == 'M' {
		p.commands = p.commands[:len(p.commands)-1]
	}
	p.commands = append(p.commands, "M"+ntos(pt))
	p.lastCmd = 'M'
	p.last = &pt
}

func (p *svgPathPen) lineTo(pt point) {
	var cmd byte
	var pts string
	switch {
	case p.last != nil && pt[0].eq(p.last[0]) && pt[1].eq(p.last[1]):
		return
	case p.last != nil && pt[0].eq(p.last[0]):
		cmd, pts = 'V', pt[1].String()
	case p.last != nil && pt[1].eq(p.last[1]):
		cmd, pts = 'H', pt[0].String()
	case p.lastCmd == 'M':
		pts = " " + ntos(pt)
	default:
		cmd, pts = 'L', ntos(pt)
	}
	t := ""
	if cmd != 0 {
		t = string(cmd)
		p.lastCmd = cmd
	}
	p.commands = append(p.commands, t+pts)
	p.last = &pt
}

func (p *svgPathPen) qCurveToOne(pt1, pt2 point) {
	p.commands = append(p.commands, "Q"+ntos(pt1)+" "+ntos(pt2))
	p.lastCmd = 'Q'
	p.last = &pt2
}

// qCurveTo is BasePen.qCurveTo with decomposeQuadraticSegment: the implied
// point between two off-curve points is 0.5*(x+nx), taken on the points as
// they arrive here, that is after any transformation.
func (p *svgPathPen) qCurveTo(pts []point, noOnCurve bool) {
	n := len(pts) // control points; with noOnCurve there is no final point
	if noOnCurve {
		last, first := pts[len(pts)-1], pts[0]
		start := point{last[0].add(first[0]).half(), last[1].add(first[1]).half()}
		p.moveTo(start)
		pts = append(append([]point(nil), pts...), start)
	} else {
		n--
	}
	if n == 0 {
		p.lineTo(pts[0])
		return
	}
	for i := 0; i < n-1; i++ {
		a, b := pts[i], pts[i+1]
		p.qCurveToOne(a, point{a[0].add(b[0]).half(), a[1].add(b[1]).half()})
	}
	p.qCurveToOne(pts[len(pts)-2], pts[len(pts)-1])
}

func (p *svgPathPen) closePath() {
	p.commands = append(p.commands, "Z")
	p.lastCmd = 'Z'
	p.last = nil
}

// addComponent is DecomposingPen.addComponent: draw the base glyph onto this
// pen, through a TransformPen unless the transformation is the identity.
func (p *svgPathPen) addComponent(gid uint16, t transform) error {
	var out pen = p
	if !t.eq(identity) {
		out = &transformPen{out: p, t: t}
	}
	p.depth++
	defer func() { p.depth-- }()
	return p.glyphs.draw(gid, out, p.depth)
}

func (p *svgPathPen) d() string { return strings.Join(p.commands, "") }
