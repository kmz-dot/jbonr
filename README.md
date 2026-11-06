<div align="center">

<img src="assets/logo.svg" width="150" alt="jango-bulk logo">

# jango-bulk

**Bulk-create Jango AI accounts, capture their sessions, and serve the whole pool to 9Router as one OpenAI-compatible endpoint.**

[![Go 1.23](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License MIT](https://img.shields.io/badge/License-MIT-3DA639?logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-5B6B7B)](#install)
[![Tests](https://img.shields.io/badge/tests-e2e%20passing-2EA043)](#tests)
[![9Router](https://img.shields.io/badge/9Router-ready-7C5CFF)](https://9router.com)
[![Telegram](https://img.shields.io/badge/community-Telegram-26A5E4?logo=telegram&logoColor=white)](https://t.me/inspectoraiofc)

[English](README.md) · [Bahasa Indonesia](README.id.md) · [简体中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

---

`jango-bulk` is a single, dependency-free Go binary. It creates disposable
mailboxes, signs up on Jango AI, follows the verification link, stores the
session cookie, and then exposes every account to 9Router through a local
OpenAI-compatible bridge with round-robin load balancing and per-account
conversation memory.

```text
mail.tm inbox ──▶ signup ──▶ verification ──▶ cookie + token ──▶ accounts.json
                                                                  │
                                    9Router ◀── /v1 bridge ◀──────┘
```

## Why it exists

- **No browser, no Selenium.** Pure HTTP + cookie jars, so 20 accounts take
  seconds, not hours.
- **Sessions survive restarts.** The 30-day `__Secure-jango_rt` cookie is
  captured on signup; a 15-minute access token is refreshed automatically.
- **One endpoint for the whole pool.** The bridge is OpenAI-compatible, so any
  client that speaks `/v1/chat/completions` works — 9Router included.
- **Wiring is automated.** Nodes, models, aliases and API keys are created in
  9Router through its dashboard API, with an end-to-end reply check.

## Available models

Jango AI exposes a single upstream model. The bridge registers it under a
prefix and also publishes one pinned id per account:

| Model id | Description | Upstream |
| --- | --- | --- |
| `jango/deepseek-v4-1-flash-uncensored` | Load-balanced across every healthy account | `deepseek-v4-1-flash-uncensored` |
| `jango/acc-<id>/deepseek-v4-1-flash-uncensored` | Pinned to one account (stateful chat) | same |

Each reply costs **5 credits** on the account that served it; the bridge keeps
tracking balances and routes away from empty accounts.

## Install

Requires Go 1.23+.

```bash
git clone https://codeberg.org/0xgetz/jango-bulk.git
cd jango-bulk
make build          # -> bin/jango-bulk
./bin/jango-bulk --help
```

Or straight from source:

```bash
go run ./cmd/jango-bulk signup -n 5
```

## Quick start

```bash
# 1. create 5 accounts (inboxes, signup, email verification, session capture)
jango-bulk signup -n 5 -c 4

# 2. inspect what was captured
jango-bulk list
jango-bulk check

# 3. serve the pool as an OpenAI-compatible endpoint
jango-bulk serve --addr 127.0.0.1:20581

# 4. register it in a running 9Router (nodes, model, key) and verify a reply
jango-bulk connect --router http://127.0.0.1:20128

# ...or do all of it in one shot
jango-bulk run -n 5 --router http://127.0.0.1:20128
```

Then use the returned router key with any OpenAI client:

```bash
curl -s http://127.0.0.1:20128/v1/chat/completions \
  -H "Authorization: Bearer $ROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"jango/deepseek-v4-1-flash-uncensored",
       "messages":[{"role":"user","content":"hi"}]}'
```

## Commands

| Command | What it does |
| --- | --- |
| `signup` | create N accounts in parallel: inbox → signup → verify → capture |
| `list` | table of accounts with credits, verification and cookie health |
| `check` | refresh every token and read its credit balance |
| `export` | sessions as `json`, `csv` or `bridge` (hand-off file) |
| `serve` | run the OpenAI-compatible bridge (`/v1/models`, `/v1/chat/completions`) |
| `connect` | register the pool in 9Router and verify a reply end to end |
| `run` | `signup` → `export` → background `serve` → `connect` in one command |
| `verify-models` | list the models 9Router exposes |
| `version` | print the build and upstream model |

Flags that matter:

```text
--store PATH        account store (default data/accounts.json)
-n, --count N       accounts to create
-c, --concurrency N parallel workers (default 4)
--name PREFIX       display name prefix
--email-prefix P    mailbox prefix
--mail-api URL      mail.tm base URL
--jango-api URL     Jango API base URL
--verify-timeout D  how long to wait for the mail (default 180s)
--no-verify         skip the mail step
--addr HOST:PORT    bridge listen address
--router URL        9Router base (default http://127.0.0.1:20128)
--cli-token TOKEN   x-9r-cli-token (default: derived from ~/.9router)
--key-name NAME     router API key name (default jango-bulk)
```

## How it works

1. **Inbox** — a fresh address on mail.tm.
2. **Signup** — `POST /auth/web/login` with `Origin: https://app.jangoai.chat`;
   the response sets the 30-day `__Secure-jango_rt` cookie.
3. **Verification** — the action link from the mail is followed on its own
   origin, which completes verification and returns a short-lived token.
4. **Capture** — email, password, cookie, token and credit balance go to
   `accounts.json` (`0600`, atomic write) with a JSONL audit trail.
5. **Bridge** — each account gets its own cookie jar. On a request the bridge
   refreshes the 15-minute token when needed, creates one chat per
   conversation key, and streams the answer back as JSON or SSE.
6. **9Router** — a provider node pointing at the bridge, a model alias, and a
   key are created through the dashboard API; `connect` then asks for one real
   reply before it reports success.

Details and diagrams: [STRUCTURE.md](STRUCTURE.md).

## Tests

The whole pipeline runs offline against in-process fakes of mail.tm, the Jango
API and a 9Router instance:

```bash
go test ./... -count=1
```

`internal/e2e` signs up three accounts, captures the sessions, starts the
bridge, wires a fake 9Router, and asserts both a non-streaming and a streaming
reply arrive through the router.

## Security

- Account files are written `0600`; nothing is uploaded anywhere.
- The bridge binds `127.0.0.1` by default — use `--api-key` and a tunnel if you
  expose it.
- Cookies and tokens are credentials: treat `data/` like a password store.

## Community

Questions, bugs and ideas: [t.me/inspectoraiofc](https://t.me/inspectoraiofc)

## License

MIT — see [LICENSE](LICENSE).
