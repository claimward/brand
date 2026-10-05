package main

import (
	"strconv"
	"strings"
)

// num is a coordinate as the original Python generator held it: either an int
// or a float. The distinction is visible in the committed SVG, which spells an
// int "-23" and a float "618.0", so it has to survive every step that the
// coordinate goes through, exactly as Python's arithmetic carries it:
// int op int stays int, anything with a float in it becomes a float, and
// 0.5*(a+b) is always a float.
//
// Every value here is far inside the range where float64 holds integers
// exactly, so v is enough to carry an int too.
type num struct {
	v       float64
	isFloat bool
}

func ival(v int) num        { return num{v: float64(v)} }
func fval(v float64) num    { return num{v: v, isFloat: true} }
func (a num) add(b num) num { return num{a.v + b.v, a.isFloat || b.isFloat} }

// mul rounds its product explicitly: without the conversion Go may fuse it
// with the add that follows into one FMA (arm64 does), which rounds once
// where Python rounds twice.
func (a num) mul(b num) num { return num{float64(a.v * b.v), a.isFloat || b.isFloat} }

// half is Python's 0.5 * x: a float whatever x was.
func (a num) half() num { return fval(0.5 * a.v) }

// eq is Python's ==, which compares an int and a float by value.
func (a num) eq(b num) bool { return a.v == b.v }

// String is Python's str() of the value: str(int) for an int, repr(float)
// for a float.
func (a num) String() string {
	if !a.isFloat {
		return strconv.FormatInt(int64(a.v), 10)
	}
	return pyFloatRepr(a.v)
}

// pyFloatRepr is Python 3's repr(float): the shortest digit string that reads
// back as the same float64 (which strconv also produces), laid out the way
// Python lays it out. Python writes positional notation when the decimal
// exponent is in [-4, 16), always with a fractional part ("618.0"), and
// scientific notation otherwise, with at least two exponent digits ("1e+16",
// "1.5e-05"). Go's own 'g' format switches to scientific at a different
// exponent, which is why this is spelled out.
func pyFloatRepr(f float64) string {
	e := strconv.FormatFloat(f, 'e', -1, 64) // e.g. "-1.5158666666666666e+03"
	i := strings.IndexByte(e, 'e')
	exp, err := strconv.Atoi(e[i+1:])
	if err != nil { // NaN and Inf have no exponent
		switch {
		case f != f:
			return "nan"
		case f > 0:
			return "inf"
		default:
			return "-inf"
		}
	}
	if exp < -4 || exp >= 16 {
		return e // strconv already writes "e+16" and "e-05" as Python does
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.ContainsAny(s, ".") {
		s += ".0"
	}
	return s
}
