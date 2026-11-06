<div align="center">

<img src="assets/logo.svg" width="150" alt="logo de jango-bulk">

# jango-bulk

**Crea cuentas de Jango AI en lote, captura sus sesiones y sirve todo el pool a 9Router como un único endpoint compatible con OpenAI.**

[![Go 1.23](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License MIT](https://img.shields.io/badge/License-MIT-3DA639?logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-5B6B7B)](#instalación)
[![Tests](https://img.shields.io/badge/tests-e2e%20passing-2EA043)](#pruebas)
[![9Router](https://img.shields.io/badge/9Router-ready-7C5CFF)](https://9router.com)
[![Telegram](https://img.shields.io/badge/comunidad-Telegram-26A5E4?logo=telegram&logoColor=white)](https://t.me/inspectoraiofc)

[English](README.md) · [Bahasa Indonesia](README.id.md) · [简体中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

---

`jango-bulk` es un único binario Go sin dependencias. Crea buzones desechables,
se registra en Jango AI, sigue el enlace de verificación, guarda la cookie de
sesión y expone cada cuenta a 9Router a través de un puente local compatible con
OpenAI, con balanceo round-robin y memoria de conversación por cuenta.

```text
buzón mail.tm ──▶ registro ──▶ verificación ──▶ cookie + token ──▶ accounts.json
                                                                     │
                                    9Router ◀── puente /v1 ◀─────────┘
```

## Por qué existe

- **Sin navegador ni Selenium.** Solo HTTP y tarros de cookies: 20 cuentas en
  segundos.
- **Las sesiones sobreviven al reinicio.** La cookie `__Secure-jango_rt` (30
  días) se captura al registrarse; el token de 15 minutos se renueva solo.
- **Un endpoint para todo el pool.** El puente habla `/v1/chat/completions`, así
  que funciona con cualquier cliente OpenAI.
- **Conexión automatizada.** Nodos, modelos, alias y claves se crean en 9Router
  vía su API, con una respuesta real como prueba final.

## Modelos disponibles

Jango AI expone un único modelo upstream. El puente lo registra con prefijo y
además publica un id fijo por cuenta:

| Id de modelo | Descripción | Upstream |
| --- | --- | --- |
| `jango/deepseek-v4-1-flash-uncensored` | Balanceado entre todas las cuentas sanas | `deepseek-v4-1-flash-uncensored` |
| `jango/acc-<id>/deepseek-v4-1-flash-uncensored` | Fijado a una cuenta (chat con estado) | igual |

Cada respuesta consume **5 créditos** de la cuenta que respondió; el puente
sigue los saldos y evita las cuentas vacías.

## Instalación

Requiere Go 1.23+.

```bash
git clone https://codeberg.org/0xgetz/jango-bulk.git
cd jango-bulk
make build          # -> bin/jango-bulk
./bin/jango-bulk --help
```

## Inicio rápido

```bash
# 1. crear 5 cuentas (buzón, registro, verificación, captura de sesión)
jango-bulk signup -n 5 -c 4

# 2. revisar lo capturado
jango-bulk list
jango-bulk check

# 3. servir el pool como endpoint compatible con OpenAI
jango-bulk serve --addr 127.0.0.1:20581

# 4. registrarlo en un 9Router activo y verificar una respuesta
jango-bulk connect --router http://127.0.0.1:20128

# ...o todo de una vez
jango-bulk run -n 5 --router http://127.0.0.1:20128
```

## Comandos

| Comando | Qué hace |
| --- | --- |
| `signup` | crea N cuentas en paralelo: buzón → registro → verificación → captura |
| `list` | tabla con créditos, verificación y salud de la cookie |
| `check` | renueva cada token y lee su saldo de créditos |
| `export` | sesiones en `json`, `csv` o `bridge` |
| `serve` | ejecuta el puente compatible con OpenAI |
| `connect` | registra el pool en 9Router y verifica una respuesta |
| `run` | `signup` → `export` → `serve` en segundo plano → `connect` |
| `verify-models` | lista los modelos que expone 9Router |
| `version` | muestra la versión y el modelo upstream |

Arquitectura: [STRUCTURE.md](STRUCTURE.md)

## Pruebas

Todo el pipeline corre offline contra dobles en proceso de mail.tm, la API de
Jango y un 9Router simulado:

```bash
go test ./... -count=1
```

## Seguridad

- Los archivos de cuentas se escriben con permisos `0600`; nada se sube.
- El puente escucha solo en `127.0.0.1` por defecto; usa `--api-key` y un túnel
  si lo expones.
- Las cookies y los tokens son credenciales: trata `data/` como un gestor de
  contraseñas.

## Comunidad

Dudas, errores e ideas: [t.me/inspectoraiofc](https://t.me/inspectoraiofc)

## Licencia

MIT — ver [LICENSE](LICENSE).
