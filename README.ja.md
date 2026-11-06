<div align="center">

<img src="assets/logo.svg" width="150" alt="jango-bulk ロゴ">

# jango-bulk

**Jango AI のアカウントを一括作成し、セッションを取得して、プール全体を 1 つの OpenAI 互換エンドポイントとして 9Router に接続します。**

[![Go 1.23](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License MIT](https://img.shields.io/badge/License-MIT-3DA639?logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-5B6B7B)](#インストール)
[![Tests](https://img.shields.io/badge/tests-e2e%20passing-2EA043)](#テスト)
[![9Router](https://img.shields.io/badge/9Router-ready-7C5CFF)](https://9router.com)
[![Telegram](https://img.shields.io/badge/community-Telegram-26A5E4?logo=telegram&logoColor=white)](https://t.me/inspectoraiofc)

[English](README.md) · [Bahasa Indonesia](README.id.md) · [简体中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

---

`jango-bulk` は依存のない単一の Go バイナリです。使い捨てメールボックスを作成し、
Jango AI に登録し、確認リンクをたどり、セッション Cookie を保存したうえで、
すべてのアカウントをローカルの OpenAI 互換ブリッジ経由で 9Router に公開します。
ラウンドロビンでの負荷分散と、アカウントごとの会話メモリを備えます。

```text
mail.tm 受信箱 ──▶ 登録 ──▶ 確認 ──▶ Cookie + トークン ──▶ accounts.json
                                                              │
                                  9Router ◀── /v1 ブリッジ ◀───┘
```

## 特徴

- **ブラウザ不要、Selenium 不要。** HTTP とクッキージャーだけで、20 アカウント
  を数秒で作成。
- **再起動してもセッション維持。** 登録時に 30 日有効の
  `__Secure-jango_rt` を取得し、15 分のアクセストークンは自動更新。
- **全アカウントを 1 エンドポイントで。** `/v1/chat/completions` を話せる
  クライアントなら何でも利用可能。
- **接続も自動化。** 9Router の dashboard API でノード、モデル、別名、API
  キーを作成し、実際の返答で検証します。

## 利用可能なモデル

Jango AI の上流モデルは 1 つだけです。ブリッジはプレフィックス付きで登録し、
アカウントごとの固定 ID も公開します。

| モデル ID | 説明 | 上流 |
| --- | --- | --- |
| `jango/deepseek-v4-1-flash-uncensored` | 正常な全アカウントに負荷分散 | `deepseek-v4-1-flash-uncensored` |
| `jango/acc-<id>/deepseek-v4-1-flash-uncensored` | 1 アカウントに固定（状態保持チャット） | 同上 |

返答ごとに、応答したアカウントの**クレジットを 5 消費**します。ブリッジは残高を
追跡し、枯渇したアカウントを避けます。

## インストール

Go 1.23 以上が必要です。

```bash
git clone https://codeberg.org/0xgetz/jango-bulk.git
cd jango-bulk
make build          # -> bin/jango-bulk
./bin/jango-bulk --help
```

## クイックスタート

```bash
# 1. 5 アカウント作成（受信箱、登録、メール確認、セッション取得）
jango-bulk signup -n 5 -c 4

# 2. 結果を確認
jango-bulk list
jango-bulk check

# 3. プールを OpenAI 互換エンドポイントとして起動
jango-bulk serve --addr 127.0.0.1:20581

# 4. 稼働中の 9Router に登録し、返答を検証
jango-bulk connect --router http://127.0.0.1:20128

# ...まとめて実行する場合
jango-bulk run -n 5 --router http://127.0.0.1:20128
```

## コマンド

| コマンド | 内容 |
| --- | --- |
| `signup` | N アカウントを並列作成：受信箱 → 登録 → 確認 → 取得 |
| `list` | クレジット、確認状況、Cookie 健全性の一覧 |
| `check` | 全トークンを更新し残高を取得 |
| `export` | セッションを `json` / `csv` / `bridge` で出力 |
| `serve` | OpenAI 互換ブリッジを起動 |
| `connect` | プールを 9Router に登録し返答を検証 |
| `run` | `signup` → `export` → バックグラウンド `serve` → `connect` |
| `verify-models` | 9Router のモデル一覧 |
| `version` | バージョンと上流モデルを表示 |

構成図: [STRUCTURE.md](STRUCTURE.md)

## テスト

mail.tm、Jango API、9Router のプロセス内モックを使い、パイプライン全体を
オフラインで実行します。

```bash
go test ./... -count=1
```

## セキュリティ

- アカウントファイルは `0600` で保存され、どこにも送信されません。
- ブリッジは既定で `127.0.0.1` のみ待ち受けます。公開する場合は
  `--api-key` とトンネルを使ってください。
- Cookie とトークンは資格情報です。`data/` はパスワードストア扱いで。

## コミュニティ

質問・バグ・アイデア: [t.me/inspectoraiofc](https://t.me/inspectoraiofc)

## ライセンス

MIT — [LICENSE](LICENSE) を参照。
