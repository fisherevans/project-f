#!/usr/bin/env bash
set -euo pipefail

# batch_patch_kern.sh
# Repeatable kerning patch step for one or more fonts.
# Each font has its own kerning CSV.
#
# Requires: python3, patch_kern.py

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PATCH_PY="${ROOT_DIR}/patch_kern.py"

# Interpret adjust_px in each font's CSV at this pixel size.
PX_SIZE="${PX_SIZE:-16}"

# Set DROP_GPOS=1 if you want to remove GPOS (usually leave off).
DROP_GPOS="${DROP_GPOS:-0}"

# Add fonts here. Format: "input.ttf|output.ttf|kerning.csv|calibrate_char|calibrate_advance_px"
FONTS=(
  "${ROOT_DIR}/ff57i.ttf|${ROOT_DIR}/ff57i.patched.ttf|${ROOT_DIR}/ff57i.kerning.csv|0|7"
)

common_args=(
  "${PATCH_PY}"
  "--pxsize" "${PX_SIZE}"
)

if [[ "${DROP_GPOS}" == "1" ]]; then
  common_args+=("--drop-gpos")
fi

for entry in "${FONTS[@]}"; do
  IFS='|' read -r in_ttf out_ttf kern_csv calib_char calib_advance_px <<< "${entry}"

  if [[ ! -f "${in_ttf}" ]]; then
    echo "ERROR: input font not found: ${in_ttf}" >&2
    exit 1
  fi
  if [[ ! -f "${kern_csv}" ]]; then
    echo "ERROR: kerning CSV not found: ${kern_csv}" >&2
    exit 1
  fi

  if [[ -z "${calib_char:-}" ]]; then
    calib_char="0"
  fi
  if [[ -z "${calib_advance_px:-}" ]]; then
    calib_advance_px="0"
  fi

  # Basic validation
  if [[ "${#calib_char}" -ne 1 ]]; then
    echo "ERROR: calibrate_char must be exactly 1 character (got: ${calib_char})" >&2
    exit 1
  fi
  if [[ ! "${calib_advance_px}" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
    echo "ERROR: calibrate_advance_px must be a number (got: ${calib_advance_px})" >&2
    exit 1
  fi

  echo "Patching: ${in_ttf} -> ${out_ttf}"
  echo "  CSV: ${kern_csv}"
  echo "  Calib: ${calib_char}@${calib_advance_px}px"
  python3 "${common_args[@]}" --in "${in_ttf}" --out "${out_ttf}" --csv "${kern_csv}" --calibrate-char "${calib_char}" --calibrate-advance-px "${calib_advance_px}"
done

echo "Done."