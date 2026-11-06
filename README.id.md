<div align="center">

<img src="assets/logo.svg" width="150" alt="logo jango-bulk">

# jango-bulk

**Buat banyak akun Jango AI sekaligus, simpan sesinya, dan sajikan seluruh pool ke 9Router sebagai satu endpoint kompatibel OpenAI.**

[![Go 1.23](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License MIT](https://img.shields.io/badge/License-MIT-3DA639?logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-5B6B7B)](#instalasi)
[![Tests](https://img.shields.io/badge/tests-e2e%20passing-2EA043)](#pengujian)
[![9Router](https://img.shields.io/badge/9Router-ready-7C5CFF)](https://9router.com)
[![Telegram](https://img.shields.io/badge/komunitas-Telegram-26A5E4?logo=telegram&logoColor=white)](https://t.me/inspectoraiofc)

[English](README.md) · [Bahasa Indonesia](README.id.md) · [简体中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

---

`jango-bulk` adalah satu binary Go tanpa dependensi eksternal. Program ini
membuat inbox sekali pakai, mendaftar ke Jango AI, mengikuti tautan verifikasi,
menyimpan cookie sesi, lalu mengekspos setiap akun ke 9Router lewat bridge
OpenAI-compatible lokal dengan load balancing round-robin dan memori
percakapan per akun.

```text
inbox mail.tm ──▶ daftar ──▶ verifikasi ──▶ cookie + token ──▶ accounts.json
                                                                  │
                                    9Router ◀── bridge /v1 ◀──────┘
```

## Kenapa dibuat

- **Tanpa browser, tanpa Selenium.** Hanya HTTP + cookie jar, jadi 20 akun
  selesai dalam hitungan detik.
- **Sesi tahan restart.** Cookie `__Secure-jango_rt` (30 hari) ditangkap saat
  daftar; access token 15 menit diperbarui otomatis.
- **Satu endpoint untuk semua akun.** Bridge-nya kompatibel OpenAI, jadi klien
  apa pun yang mendukung `/v1/chat/completions` bisa dipakai.
- **Wiring otomatis.** Node, model, alias, dan API key dibuat di 9Router lewat
  dashboard API, lalu diuji dengan satu balasan nyata.

## Model yang tersedia

Jango AI hanya mengekspos satu model upstream. Bridge mendaftarkannya dengan
prefix dan juga menyediakan satu id per akun:

| Model id | Keterangan | Upstream |
| --- | --- | --- |
| `jango/deepseek-v4-1-flash-uncensored` | Load balanced ke semua akun sehat | `deepseek-v4-1-flash-uncensored` |
| `jango/acc-<id>/deepseek-v4-1-flash-uncensored` | Terkunci ke satu akun (chat stateful) | sama |

Setiap balasan memakai **5 kredit** pada akun yang menjawab; bridge memantau
saldo dan menghindari akun yang kosong.

## Instalasi

Butuh Go 1.23+.

```bash
git clone https://codeberg.org/0xgetz/jango-bulk.git
cd jango-bulk
make build          # -> bin/jango-bulk
./bin/jango-bulk --help
```

## Mulai cepat

```bash
# 1. buat 5 akun (inbox, daftar, verifikasi email, tangkap sesi)
jango-bulk signup -n 5 -c 4

# 2. periksa hasilnya
jango-bulk list
jango-bulk check

# 3. jalankan pool sebagai endpoint kompatibel OpenAI
jango-bulk serve --addr 127.0.0.1:20581

# 4. daftarkan ke 9Router yang sedang berjalan dan uji balasannya
jango-bulk connect --router http://127.0.0.1:20128

# ...atau semuanya sekaligus
jango-bulk run -n 5 --router http://127.0.0.1:20128
```

## Perintah

| Perintah | Fungsi |
| --- | --- |
| `signup` | buat N akun paralel: inbox → daftar → verifikasi → tangkap sesi |
| `list` | tabel akun dengan kredit, status verifikasi, dan kesehatan cookie |
| `check` | segarkan semua token dan baca saldo kredit |
| `export` | ekspor sesi sebagai `json`, `csv`, atau `bridge` |
| `serve` | jalankan bridge OpenAI-compatible |
| `connect` | daftarkan pool ke 9Router dan uji satu balasan |
| `run` | `signup` → `export` → `serve` (background) → `connect` |
| `verify-models` | tampilkan model yang tersedia di 9Router |
| `version` | cetak versi dan model upstream |

Penjelasan arsitektur: [STRUCTURE.md](STRUCTURE.md).

## Pengujian

Seluruh pipeline berjalan offline memakai tiruan in-process untuk mail.tm, API
Jango, dan 9Router:

```bash
go test ./... -count=1
```

## Keamanan

- File akun ditulis dengan mode `0600`; tidak ada data yang diunggah.
- Bridge default hanya mendengarkan `127.0.0.1` — gunakan `--api-key` dan
  tunnel jika ingin mengeksposnya.
- Cookie dan token adalah kredensial: perlakukan folder `data/` seperti
  penyimpanan kata sandi.

## Komunitas

Pertanyaan, bug, dan ide: [t.me/inspectoraiofc](https://t.me/inspectoraiofc)

## Lisensi

MIT — lihat [LICENSE](LICENSE).
