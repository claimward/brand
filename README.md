<p align="center">
  <img src="logo/claimward-lockup.svg" alt="claimward" height="56">
</p>

# Brand assets

[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD--3--Clause-blue.svg)](https://opensource.org/license/bsd-3-clause)

Canonical home for the **claimward** visual identity. Other repos and the docs
site should link assets from here (raw URLs) rather than keeping their own copies.

The mark is a monoline **key + shield**. The key's teeth are the lock *wards* —
a nod to the name (OIDC *claims* + *ward* / guard) and to OIDC + hardware-key auth.

## Colours

| Role | Hex |
|------|-----|
| Brand (teal) | `#0D9488` |
| Wordmark "ward" (deep teal) | `#134E4A` |
| Knockout (badge / favicon / tray) | `#FFFFFF` |

## Typeface

Wordmark drawn from **Inter** (SIL OFL) — `claim` Medium + `ward` Bold. In the
SVGs the wordmark is **outlined** (real paths), so no font is needed to render it.

## Layout

```
logo/      claimward-mark.svg        icon, teal on transparent
           claimward-mark-mono.svg   single-colour (currentColor)
           claimward-lockup.svg      horizontal logo (mark + outlined wordmark)
avatar/    claimward-badge.svg       square org avatar (white mark on teal)
           avatar-512.png            ready-to-upload GitHub org avatar
favicon/   claimward-favicon.svg     simplified small-size mark
           favicon-16/32/512.png, apple-touch-180.png
macos/     claimward-tray-Template.svg   menu-bar TEMPLATE (black+alpha; macOS tints it)
           claimward-trayTemplate.png @2x @3x   (18 / 36 / 54 px)
           claimward-tray-white.svg, tray-white-18.png @2x   hard white, dark-only
src/       Inter-Medium.woff2, Inter-Bold.woff2 + OFL.txt   the wordmark's fonts (also served by the docs site)
cmd/outline/   Go tool that outlines the wordmark into claimward-lockup.svg (go.mod at the root)
```

## Usage notes

- **Org avatar**: upload `avatar/avatar-512.png` via the org Settings → Profile page (not committed as the avatar — this is just the archive).
- **macOS menu bar**: use `macos/claimward-trayTemplate*.png`, load it and set
  `image.isTemplate = true` (or keep the `Template` suffix so AppKit does it).
  Don't tint it yourself — macOS handles light/dark/selected.
- **Single ink**: `logo/claimward-mark-mono.svg` inherits `currentColor`.

## Regenerate the outlined wordmark

```sh
go run ./cmd/outline     # rewrites logo/claimward-lockup.svg
```

It draws `claim` from `src/Inter-Medium.woff2` and `ward` from
`src/Inter-Bold.woff2` at 30 px with −0.5 px letter spacing, in `#0D9488` and
`#134E4A`, beside the mark. It is pure Go (the WOFF2 files are decoded by
[go-opentype](https://github.com/go-opentype/opentype)), and needs only a Go
toolchain at the version `go.mod` names.

The output is byte-for-byte the file a fontTools script produced before it, and
CI holds it there: it runs the tool and fails when `logo/claimward-lockup.svg`
differs from what is committed. Change the font or a parameter, and commit the
regenerated logo with it.

Render PNGs with `rsvg-convert`, e.g.:

```sh
rsvg-convert -w 512 avatar/claimward-badge.svg -o avatar/avatar-512.png
```

---

Licensed BSD-3-Clause like the rest of claimward ([LICENSE](LICENSE)). The Inter
fonts in `src/` are © The Inter Project Authors, under the SIL Open Font License
1.1, whose text travels with them as the licence requires ([src/OFL.txt](src/OFL.txt),
from [rsms/inter](https://github.com/rsms/inter)).
