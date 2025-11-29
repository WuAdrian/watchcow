# Bilibili Redirect Example

使用 WatchCow 创建一个 fnOS 快捷方式应用，点击后直接跳转到 B 站。

## 原理

使用 busybox 内置的 httpd + CGI 实现 302 重定向，镜像极小 (~1MB)。

## 快速开始

```bash
cd examples/bilibili

# 给脚本执行权限
chmod +x redirect.sh

# 启动容器
docker compose up -d
```

访问: http://localhost:3000/cgi-bin/index.cgi → 自动跳转到 https://www.bilibili.com/

## 文件说明

| 文件 | 说明 |
|------|------|
| `redirect.sh` | CGI 脚本，返回 302 重定向响应 |
| `compose.yaml` | Docker Compose 配置 + WatchCow labels |

## 自定义跳转目标

编辑 `redirect.sh` 中的 `Location` 行：

```sh
echo "Location: https://your-target-url.com/"
```

## 验证

```bash
# 测试重定向
curl -I http://localhost:3000/cgi-bin/index.cgi

# 预期输出:
# HTTP/1.1 302 Found
# Location: https://www.bilibili.com/
```

## WatchCow Labels

| Label | 值 | 说明 |
|-------|-----|------|
| `watchcow.enable` | `true` | 启用发现 |
| `watchcow.appname` | `watchcow.bilibili` | 应用标识 |
| `watchcow.display_name` | `哔哩哔哩` | 显示名称 |
| `watchcow.service_port` | `3000` | 服务端口 |
| `watchcow.path` | `/cgi-bin/index.cgi` | CGI 路径 |
| `watchcow.ui_type` | `url` | 新标签页打开 |
