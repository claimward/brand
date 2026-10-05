// Command outline draws the claimward wordmark from Inter as real SVG paths
// and writes logo/claimward-lockup.svg, so that the logo renders without the
// font: "claim" in Inter Medium (#0D9488) and "ward" in Inter Bold (#134E4A),
// 30 px, letter spacing -0.5 px, beside the key-and-shield mark.
//
// Run it from the root of the repository:
//
//	go run ./cmd/outline
//
// It reads src/Inter-Medium.woff2 and src/Inter-Bold.woff2, the same files the
// docs site serves for its headings. The output is byte-identical to the one
// the former fontTools script (src/outline.py) produced; CI regenerates it and
// fails when the committed file differs.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	src := flag.String("src", "src", "directory holding Inter-Medium.woff2 and Inter-Bold.woff2")
	out := flag.String("o", filepath.Join("logo", "claimward-lockup.svg"), "SVG to write")
	flag.Parse()
	if err := run(*src, *out); err != nil {
		fmt.Fprintln(os.Stderr, "outline:", err)
		os.Exit(1)
	}
}

func run(src, out string) error {
	medium, err := os.ReadFile(filepath.Join(src, "Inter-Medium.woff2"))
	if err != nil {
		return err
	}
	bold, err := os.ReadFile(filepath.Join(src, "Inter-Bold.woff2"))
	if err != nil {
		return err
	}
	svg, err := lockup(medium, bold)
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, svg, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d bytes)\n", out, len(svg))
	return nil
}
