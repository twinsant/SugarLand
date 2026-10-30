#!/usr/bin/env bash
# 给 Godot Web 导出注入 GA4。每次重新导出后跑一次，幂等。
# 用法：bash tools/patch-web-export.sh [导出目录，默认 dist/web]
set -euo pipefail

MEASUREMENT_ID="${GA_MEASUREMENT_ID:-G-YR2FR9QRWL}"
TARGET_DIR="${1:-dist/web}"
INDEX="${TARGET_DIR}/index.html"

[[ -f "${INDEX}" ]] || { echo "找不到 ${INDEX}，先导出 Web 版" >&2; exit 1; }
if grep -q "${MEASUREMENT_ID}" "${INDEX}"; then
  echo "已注入 ${MEASUREMENT_ID}，跳过"
  exit 0
fi

SNIPPET=$(cat <<HTML
	<script async src="https://www.googletagmanager.com/gtag/js?id=${MEASUREMENT_ID}"></script>
	<script>
		window.dataLayer = window.dataLayer || [];
		function gtag() { dataLayer.push(arguments); }
		gtag('js', new Date());
		gtag('config', '${MEASUREMENT_ID}');
		(function () {
			var started = false;
			function start() { if (started) { return; } started = true; gtag('event', 'game_start'); }
			// capture 阶段监听：Godot 的 canvas 会 preventDefault 甚至吞掉冒泡，
			// 挂在 window 冒泡阶段时点击可能触不到。
			['pointerdown', 'touchstart', 'keydown'].forEach(function (type) {
				document.addEventListener(type, start, { passive: true, capture: true });
			});
		}());
	</script>
HTML
)

python3 - "${INDEX}" "${SNIPPET}" <<'PY'
import sys
path, snippet = sys.argv[1], sys.argv[2]
html = open(path, encoding='utf-8').read()
assert '</head>' in html, "index.html 里没有 </head>"
html = html.replace('</head>', snippet + '\n</head>', 1)
open(path, 'w', encoding='utf-8').write(html)
print("已注入 GA4:", path)
PY

echo "别忘了重新生成 gzip：gzip -9 -kf ${INDEX}（线上有 gzip_static）"
