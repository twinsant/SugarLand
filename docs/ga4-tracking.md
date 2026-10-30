# SugarLand 的 GA4 埋点

## 测量 ID

`G-YR2FR9QRWL`（GA4 媒体资源 `twinsant websites` 下独立数据流「SugarLand」）

## 为什么不用改 GDScript

线上 `game.twinsant.com` 是 Godot 的 Web 导出，直接托管 `dist/web/` 的静态文件。埋点全部注入在导出产物的 `index.html` 里，不碰游戏代码：

- `page_view`：gtag 的 `config` 自动发
- `game_start`：捕获阶段监听 `pointerdown` / `touchstart` / `keydown`，第一次交互时发一次

`game_start` 用 capture 阶段监听，因为 Godot 的 canvas 会 `preventDefault`，挂在冒泡阶段时点击可能触不到。

## 重新导出之后必须做一次

```sh
bash tools/patch-web-export.sh dist/web
gzip -9 -c dist/web/index.html > dist/web/index.html.gz
```

脚本是幂等的：已经注入过就直接跳过。

## 部署位置

服务器 `t` 上由 nginx 直接托管：

```
/root/projects/sl/index.html        当前首页（已注入）
/root/projects/sl/index.html.gz     nginx 开了 gzip_static，两个文件必须同步
```

上传时保持属主 `501:games`、权限 `644`，否则 nginx 读不到。每次替换前先备份：

```sh
ssh t "cp /root/projects/sl/index.html /root/projects/sl/index.html.bak.$(date +%Y%m%d%H%M%S)"
scp dist/web/index.html t:/root/projects/sl/index.html
scp dist/web/index.html.gz t:/root/projects/sl/index.html.gz
ssh t "chown 501:games /root/projects/sl/index.html*; chmod 644 /root/projects/sl/index.html*"
```

## 验证

```sh
curl -s --compressed https://game.twinsant.com/ | grep -c G-YR2FR9QRWL   # 应为 1 或 2
```

浏览器打开页面点一下画布，GA4 实时报表应出现 `page_view` 与 `game_start`。
