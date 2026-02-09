#!/usr/bin/env python3
import argparse
import csv
from fontTools.ttLib import TTFont
from fontTools.ttLib.tables import _k_e_r_n


def px_to_units(px: float, units_per_px: float) -> int:
    # Convert pixel adjustment to font units using a calibrated units-per-pixel.
    # Positive increases spacing, negative tightens.
    return int(round(px * units_per_px))


def resolve_glyph_name(font: TTFont, token: str) -> str:
    """Resolve a CSV token to a glyph name.

    Accepts either:
      - a single Unicode character (e.g. "A", "'"), or
      - a glyph name (e.g. "quotesingle").

    Returns a glyph name present in the font.
    """
    tok = (token or "").strip()
    if not tok:
        raise ValueError("Empty glyph token")

    # If it's exactly one character, map via cmap to a glyph name.
    if len(tok) == 1:
        cmap = font.getBestCmap() or {}
        cp = ord(tok)
        gname = cmap.get(cp)
        if gname is None:
            raise ValueError(f"No cmap entry for character: {tok!r}")

        # Some fonts/cmaps may reference a glyph name that doesn't actually exist.
        # Validate it, and fall back to common names for known punctuation.
        rev = font.getReverseGlyphMap()
        if gname in rev:
            return gname

        candidates: list[str] = []
        if cp == 0x0027:  # apostrophe '
            candidates = [
                "quotesingle",
                "apostrophe",
                "uni0027",
                "u0027",
                "quoteright",
                "quotedblright",
            ]
        elif cp == 0x2019:  # right single quotation mark
            candidates = [
                "quoteright",
                "uni2019",
                "u2019",
                "quotesingle",
                "apostrophe",
            ]
        elif cp == 0x2018:  # left single quotation mark
            candidates = [
                "quoteleft",
                "uni2018",
                "u2018",
                "quotesingle",
            ]

        for cand in candidates:
            if cand in rev:
                return cand

        raise ValueError(
            f"cmap mapped {tok!r} (U+{cp:04X}) to missing glyph {gname!r}; tried {candidates!r}"
        )

    # Otherwise assume it's already a glyph name.
    if tok not in font.getReverseGlyphMap():
        raise ValueError(f"Unknown glyph name: {tok!r}")
    return tok

def build_kern_table(pairs_units: dict[tuple[str, str], int]) -> _k_e_r_n.table__k_e_r_n:
    kt = _k_e_r_n.table__k_e_r_n()
    kt.version = 0

    sub = _k_e_r_n.KernTable_format_0()
    sub.version = 0
    sub.coverage = 0x0001  # horizontal kerning
    sub.kernTable = pairs_units

    kt.kernTables = [sub]
    return kt

def main() -> None:
    ap = argparse.ArgumentParser(description="Idempotently patch TTF kerning from CSV.")
    ap.add_argument("--in", dest="in_path", required=True, help="Input TTF path")
    ap.add_argument("--out", dest="out_path", required=True, help="Output TTF path")
    ap.add_argument("--csv", dest="csv_path", required=True, help="Kerning CSV: left,right,adjust_px")
    ap.add_argument("--pxsize", type=int, required=True, help="Target pixel size used to interpret adjust_px")
    ap.add_argument("--drop-gpos", action="store_true", help="Remove GPOS table (optional)")
    ap.add_argument("--calibrate-char", default="0", help="Char used to calibrate units-per-pixel (default: 0)")
    ap.add_argument(
        "--calibrate-advance-px",
        type=float,
        default=0.0,
        help="Measured advance in pixels for calibrate-char (e.g., 7 for '0' if 6px ink + 1px gap). If > 0, overrides pxsize math.",
    )
    args = ap.parse_args()

    font = TTFont(args.in_path)
    upm = font["head"].unitsPerEm

    # Determine how many font units correspond to 1 rendered pixel.
    # Preferred: calibrate from a known glyph advance width and a measured advance in pixels.
    if args.calibrate_advance_px > 0:
        if len(args.calibrate_char) != 1:
            raise SystemExit("--calibrate-char must be exactly 1 character")

        cmap = font.getBestCmap()
        gname = cmap.get(ord(args.calibrate_char))
        if gname is None:
            raise SystemExit(f"No cmap entry for calibrate char: {args.calibrate_char!r}")

        adv_units, _lsb_units = font["hmtx"][gname]
        units_per_px = adv_units / float(args.calibrate_advance_px)
    else:
        # Fallback: interpret pixels using UPM/pxsize (often imperfect for pixel fonts).
        units_per_px = upm / float(args.pxsize)

    # Optional: ensure we're idempotent and not leaving other kerning behind.
    if "kern" in font:
        del font["kern"]
    if args.drop_gpos and "GPOS" in font:
        del font["GPOS"]

    units_pairs: dict[tuple[str, str], int] = {}

    with open(args.csv_path, "r", newline="") as f:
        r = csv.DictReader(f)
        required = {"left", "right", "adjust_px"}
        if set(r.fieldnames or []) != required:
            raise SystemExit("CSV must have headers: left,right,adjust_px")

        for row in r:
            try:
                left = resolve_glyph_name(font, row["left"])
                right = resolve_glyph_name(font, row["right"])
                px = float(row["adjust_px"])
            except Exception as e:
                raise SystemExit(f"Bad CSV row {row!r}: {e}")
            units = px_to_units(px, units_per_px)
            # Skip 0 adjustments (keeps table clean)
            if units != 0:
                units_pairs[(left, right)] = units

    font["kern"] = build_kern_table(units_pairs)
    font.save(args.out_path)

    # Helpful stdout (ASCII only)
    calib = (
        f"calib={args.calibrate_char!r}@{args.calibrate_advance_px}px" if args.calibrate_advance_px > 0 else "calib=none"
    )
    print(
        f"Patched kern: {len(units_pairs)} pairs, UPM={upm}, pxsize={args.pxsize}, units_per_px={units_per_px:.6f}, {calib}"
    )
    print(f"Wrote: {args.out_path}")

if __name__ == "__main__":
    main()