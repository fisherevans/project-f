#!/usr/bin/env bash
set -euo pipefail

letters=(a b c d e f g h i j k l m n o p q r s t u v w x y z)
pairs_per_row=13

to_upper() { echo "$1" | tr '[:lower:]' '[:upper:]'; }

for x in "${letters[@]}"; do
  xu="$(to_upper "$x")"

  # For each first-letter x, print two 13-wide rows (a-m, n-z) for each of 4 casing modes:
  # 1) xx, 2) xY, 3) Xy, 4) XY
  for mode in 0 1 2 3; do
    for ((i=0; i<${#letters[@]}; i+=pairs_per_row)); do
      row=()
      for ((j=i; j<i+pairs_per_row && j<${#letters[@]}; j++)); do
        y="${letters[j]}"
        yu="$(to_upper "$y")"
        case "$mode" in
          0) row+=("${x}${y}") ;;     # xx
          1) row+=("${x}${yu}") ;;    # xY
          2) row+=("${xu}${y}") ;;    # Xy
          3) row+=("${xu}${yu}") ;;   # XY
        esac
      done
      printf "%s\n" "${row[*]}"
    done
  done

  printf "\n"
done

echo "https://fontdrop.info/#/?darkmode=true"