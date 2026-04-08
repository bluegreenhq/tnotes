#!/bin/bash
set -euo pipefail

rm -rf /tmp/tnotes-demo
go build -o /tmp/tnotes-demo-bin .

DEMO_COLS=100
DEMO_ROWS=24

asciinema rec \
  --overwrite \
  --headless \
  --window-size "${DEMO_COLS}x${DEMO_ROWS}" \
  --command "DEMO_COLS=${DEMO_COLS} DEMO_ROWS=${DEMO_ROWS} expect demo.exp" \
  demo.cast

# --theme: 背景色のみ171717に変更、他はasciinemaデフォルトの値
agg --font-size 16 --line-height 1.25 \
  --theme "171717,cccccc,000000,dd3c69,4ebf22,ddaf3c,26b0d7,b954e1,54e1b9,d9d9d9,4d4d4d,dd3c69,4ebf22,ddaf3c,26b0d7,b954e1,54e1b9,ffffff" \
  demo.cast demo.gif
