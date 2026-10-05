package main

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// The judge: the logo as committed, which the fontTools script produced
// before this tool replaced it. CI also runs the tool and diffs the tree; this
// test says which byte went wrong when that diff would only say "changed".
func TestItRegeneratesTheCommittedLockupByteForByte(t *testing.T) {
	root := filepath.Join("..", "..")
	want, err := os.ReadFile(filepath.Join(root, "logo", "claimward-lockup.svg"))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "lockup.svg")
	if err := run(filepath.Join(root, "src"), out); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(got, want) {
		return
	}
	i := 0
	for i < len(got) && i < len(want) && got[i] == want[i] {
		i++
	}
	ctx := func(b []byte) string {
		lo, hi := max(i-40, 0), min(i+40, len(b))
		return string(b[lo:hi])
	}
	t.Fatalf("lockup differs from the committed file at byte %d (got %d bytes, want %d)\n got: …%s…\nwant: …%s…",
		i, len(got), len(want), ctx(got), ctx(want))
}

// Each expectation is what CPython 3 prints for repr() of the same float.
func TestFloatsAreSpelledAsPythonSpellsThem(t *testing.T) {
	for _, c := range []struct {
		f    float64
		want string
	}{
		{618, "618.0"},
		{1515.8666666666666, "1515.8666666666666"},
		{2090.733333333333, "2090.733333333333"},
		{0.30000000000000004, "0.30000000000000004"}, // 0.1+0.2 at run time
		{math.Copysign(0, -1), "-0.0"},
		{1234567890123456, "1234567890123456.0"}, // exponent 15: still positional
		{1e16, "1e+16"},                          // exponent 16: scientific
		{1.23e22, "1.23e+22"},
		{0.0001, "0.0001"},                                 // exponent -4: still positional
		{3.3333333333333335e-05, "3.3333333333333335e-05"}, // exponent -5: scientific
		{1.5e-5, "1.5e-05"},
	} {
		if got := pyFloatRepr(c.f); got != c.want {
			t.Errorf("pyFloatRepr(%v) = %q, Python says %q", c.f, got, c.want)
		}
	}
	if got := ival(-23).String(); got != "-23" {
		t.Errorf("an int is spelled %q, Python says -23", got)
	}
}

// Expectations printed by fontTools 4.66's SVGPathPen for the same calls.
func TestThePathPenWritesWhatFontToolsWrites(t *testing.T) {
	pt := func(x, y int) point { return point{ival(x), ival(y)} }

	p := &svgPathPen{}
	p.moveTo(pt(0, 0))
	p.lineTo(pt(1, 1)) // after M: no "L"
	p.lineTo(pt(2, 3)) // and still none: the last command is still M
	p.lineTo(pt(2, 5)) // vertical
	p.lineTo(pt(4, 6)) // now a plain L
	p.lineTo(pt(4, 6)) // a duplicate point is dropped
	p.lineTo(pt(7, 6)) // horizontal
	p.qCurveTo([]point{pt(3, 3), pt(7, 5), pt(11, 4)}, false)
	p.closePath()
	if got, want := p.d(), "M0 0 1 1 2 3V5L4 6H7Q3 3 5.0 4.0Q7 5 11 4Z"; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}

	p = &svgPathPen{}
	p.qCurveTo([]point{pt(0, 0), pt(10, 0), pt(10, 10)}, true) // no on-curve point at all
	p.closePath()
	if got, want := p.d(), "M5.0 5.0Q0 0 5.0 0.0Q10 0 10.0 5.0Q10 10 5.0 5.0Z"; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}

	// Through a TransformPen with a float offset, x becomes a float and y
	// stays an int, and the implied point is taken after the translation.
	p = &svgPathPen{}
	tp := &transformPen{out: p, t: transform{ival(1), ival(0), ival(0), ival(1), fval(0.5), ival(0)}}
	tp.moveTo(pt(1, 2))
	tp.qCurveTo([]point{pt(2, 2), pt(3, 3), pt(4, 2)}, false)
	tp.closePath()
	if got, want := p.d(), "M1.5 2Q2.5 2 3.0 2.5Q3.5 3 4.5 2Z"; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// Python rounds a product before adding to it. Go may fuse the two into one
// FMA, which rounds once; on arm64 it does, and 0.1*10 - 1 then comes out as
// 5.551115123125783e-17 instead of Python's 0.0.
// The operands live in a variable so the compiler cannot fold the expression
// at compile time, where it would round each step and hide the fusion.
var fmaOperands = []float64{0.1, 10, -1}

func TestAProductIsRoundedBeforeItIsAdded(t *testing.T) {
	tenth, ten, minusOne := fval(fmaOperands[0]), fval(fmaOperands[1]), fval(fmaOperands[2])
	if got := tenth.mul(ten).add(minusOne); got.v != 0 {
		t.Errorf("0.1*10 + -1 = %v, Python says 0.0", got.v)
	}
}
