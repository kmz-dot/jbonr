<div align="center">

<img src="assets/logo.svg" width="150" alt="jango-bulk logo">

# jango-bulk

**批量创建 Jango AI 账号，捕获会话，并把整组账号作为一个 OpenAI 兼容端点接入 9Router。**

[![Go 1.23](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License MIT](https://img.shields.io/badge/License-MIT-3DA639?logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-5B6B7B)](#安装)
[![Tests](https://img.shields.io/badge/tests-e2e%20passing-2EA043)](#测试)
[![9Router](https://img.shields.io/badge/9Router-ready-7C5CFF)](https://9router.com)
[![Telegram](https://img.shields.io/badge/社区-Telegram-26A5E4?logo=telegram&logoColor=white)](https://t.me/inspectoraiofc)

[English](README.md) · [Bahasa Indonesia](README.id.md) · [简体中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

---

`jango-bulk` 是一个零依赖的 Go 单文件程序。它创建临时邮箱、注册 Jango AI、
跟进验证链接、保存会话 Cookie，然后把每个账号通过本地 OpenAI 兼容桥接暴露
给 9Router，支持轮询负载均衡和按账号隔离的对话上下文。

```text
mail.tm 收件箱 ──▶ 注册 ──▶ 验证 ──▶ Cookie + Token ──▶ accounts.json
                                                          │
                                  9Router ◀── /v1 桥接 ◀──┘
```

## 设计目标

- **不用浏览器、不用 Selenium。** 纯 HTTP + Cookie 罐，20 个账号几秒完成。
- **会话可跨重启。** 注册时捕获 30 天有效的 `__Secure-jango_rt`，15 分钟的
  访问令牌自动刷新。
- **一个端点服务全部账号。** 桥接兼容 OpenAI，任何支持
  `/v1/chat/completions` 的客户端都能用。
- **自动接线。** 通过 9Router 的 dashboard API 创建节点、模型、别名和
  API Key，并做一次真实回复验证。

## 可用模型

Jango AI 只提供一上游模型。桥接会在前缀下注册它，并为每个账号发布一个
固定 ID：

| 模型 ID | 说明 | 上游 |
| --- | --- | --- |
| `jango/deepseek-v4-1-flash-uncensored` | 在所有健康账号间负载均衡 | `deepseek-v4-1-flash-uncensored` |
| `jango/acc-<id>/deepseek-v4-1-flash-uncensored` | 固定到某个账号（有状态对话） | 同上 |

每次回复消耗该账号 **5 积分**，桥接会跟踪余额并避开积分耗尽的账号。

## 安装

需要 Go 1.23+。

```bash
git clone https://codeberg.org/0xgetz/jango-bulk.git
cd jango-bulk
make build          # -> bin/jango-bulk
./bin/jango-bulk --help
```

## 快速开始

```bash
# 1. 创建 5 个账号（收件箱、注册、邮箱验证、会话捕获）
jango-bulk signup -n 5 -c 4

# 2. 查看结果
jango-bulk list
jango-bulk check

# 3. 以 OpenAI 兼容端点运行账号池
jango-bulk serve --addr 127.0.0.1:20581

# 4. 注册到正在运行的 9Router 并验证一次回复
jango-bulk connect --router http://127.0.0.1:20128

# ...或者一步到位
jango-bulk run -n 5 --router http://127.0.0.1:20128
```

## 命令

| 命令 | 作用 |
| --- | --- |
| `signup` | 并行创建 N 个账号：收件箱 → 注册 → 验证 → 捕获 |
| `list` | 账号表：积分、验证状态、Cookie 健康度 |
| `check` | 刷新所有令牌并读取积分余额 |
| `export` | 导出会话为 `json`、`csv` 或 `bridge` |
| `serve` | 运行 OpenAI 兼容桥接 |
| `connect` | 把账号池注册到 9Router 并验证回复 |
| `run` | `signup` → `export` → 后台 `serve` → `connect` |
| `verify-models` | 列出 9Router 提供的模型 |
| `version` | 输出版本与上游模型 |

架构说明见 [STRUCTURE.md](STRUCTURE.md)。

## 测试

整条流水线在本地通过 mail.tm、Jango API 和 9Router 的进程内模拟离线运行：

```bash
go test ./... -count=1
```

## 安全

- 账号文件权限为 `0600`，数据不会上传到任何地方。
- 桥接默认只监听 `127.0.0.1`；如需暴露请使用 `--api-key` 加隧道。
- Cookie 和令牌就是凭据：请把 `data/` 当作密码库对待。

## 社区

问题、反馈与想法：[t.me/inspectoraiofc](https://t.me/inspectoraiofc)

## 许可证

MIT — 见 [LICENSE](LICENSE)。
