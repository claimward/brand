<p align="center">
  <img src="logo/claimward-lockup.svg" alt="claimward" height="56">
</p>

# Brand assets

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
src/       outline.py + Inter-*.woff2            regenerate the outlined wordmark
```

## Usage notes

- **Org avatar**: upload `avatar/avatar-512.png` via the org Settings → Profile page (not committed as the avatar — this is just the archive).
- **macOS menu bar**: use `macos/claimward-trayTemplate*.png`, load it and set
  `image.isTemplate = true` (or keep the `Template` suffix so AppKit does it).
  Don't tint it yourself — macOS handles light/dark/selected.
- **Single ink**: `logo/claimward-mark-mono.svg` inherits `currentColor`.

## Regenerate the outlined wordmark

```sh
python3 -m venv .venv && .venv/bin/pip install "fonttools[woff]" brotli
.venv/bin/python src/outline.py     # rewrites logo/claimward-lockup.svg
```

Render PNGs with `rsvg-convert`, e.g.:

```sh
rsvg-convert -w 512 avatar/claimward-badge.svg -o avatar/avatar-512.png
```

---

Licensed BSD-3-Clause like the rest of claimward. Inter is under the SIL Open Font License.
