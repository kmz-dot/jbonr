# Architecture

`jango-bulk` is a single Go binary that turns mail.tm inboxes into Jango AI
accounts, stores their session cookies, and serves the whole pool to 9Router as
one OpenAI-compatible endpoint.

```mermaid
flowchart LR
    subgraph CLI["jango-bulk (single binary)"]
        C1[signup] --> CR[creator]
        C2[check] --> JG
        C3[serve] --> BR[bridge :20581]
        C4[connect] --> WR[wire]
    end

    CR -- "create inbox" --> MT[(mail.tm API)]
    CR -- "POST /auth/web/login" --> JA[(api.jangoai.chat)]
    MT -- "verification mail" --> CR
    CR --> ST[(accounts.json<br/>0600, atomic)]

    ST --> WR
    ST --> BR

    BR -- "cookies -> 15 min token" --> JA
    BR -- "POST /chats/{id}/messages/stream (SSE)" --> JA

    WR -- "nodes / models / key" --> RT[(9Router :20128)]
    RT -- "/v1/chat/completions" --> BR
    HD[OpenAI clients] --> RT
```

## Request path

```mermaid
sequenceDiagram
    participant App as OpenAI client
    participant R9 as 9Router
    participant B as bridge
    participant J as Jango API

    App->>R9: POST /v1/chat/completions (model jango/…)
    R9->>B: forward with router key
    B->>B: pool.Pick() — healthiest account, most credits
    B->>J: POST /auth/web/refresh (cookie __Secure-jango_rt)
    J-->>B: accessToken (15 min, cached)
    B->>J: POST /chats (once per conversation key)
    J-->>B: chat id
    B->>J: POST /chats/{id}/messages/stream
    J-->>B: SSE deltas
    B-->>R9: JSON or SSE
    R9-->>App: response
```

## Packages

| Path | Responsibility |
| --- | --- |
| `cmd/jango-bulk` | CLI entry point, flag parsing, command dispatch |
| `internal/mailtm` | disposable inboxes and mail polling (mail.tm REST) |
| `internal/jango` | typed Jango client: login, refresh, verify, chats, SSE, per-account cookie jar |
| `internal/creator` | signup pipeline (inbox → signup → verify → session capture), parallel workers |
| `internal/store` | JSON account store + JSONL audit log, `0600`, atomic replace |
| `internal/bridge` | OpenAI-compatible server, account pool, load balancing, conversation mapping |
| `internal/router` | 9Router dashboard client (node, connection, model alias, API key, verify) |
| `internal/wire` | export sessions and register them in 9Router |
| `internal/mock` | in-process fakes for mail.tm, Jango and 9Router |
| `internal/e2e` | end-to-end test: signup → capture → bridge → router → reply |

## Data files

```
data/
├── accounts.json           # accounts, cookies, tokens (0600)
├── accounts.json.audit     # append-only JSONL of every state change
└── bridge-sessions.json    # optional hand-off file for `serve`
```

Nothing leaves the machine except requests to mail.tm, Jango and the local
9Router instance.
