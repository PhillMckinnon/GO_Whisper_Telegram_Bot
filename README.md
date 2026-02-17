# GO Whisper Telegram Bot
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-latest-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/built%20with-Docker-2496ED?logo=docker)](https://www.docker.com/)
[![Python](https://img.shields.io/badge/Python-3.10+-3776AB?logo=python)](https://www.python.org/)
[![Telegram Bot](https://img.shields.io/badge/Telegram-Bot-26A5E4?logo=telegram)](https://core.telegram.org/bots)

This project is a Dockerized Telegram bot written in Go, designed to handle speech **transcription** and **synthesis** via a Python backend. The backend code is directly included in this repository, and is based on components from two of my earlier projects.

---

The Python backend in this repo reuses and adapts functionality from:

- [Flutter_Python_Speech_App](https://github.com/PhillMckinnon/Flutter_Python_Speech_App)
- [TG_WebApp_Flutter_WCTTS](https://github.com/PhillMckinnon/TG_WebApp_Flutter_WCTTS)

These repositories provided the foundation.

---

- Telegram bot built with Go (`go-telegram-bot-api`)
- Two main actions triggered via buttons:
  - **Transcribe**: Upload a small audio or video file and receive a transcript
  - **Synthesize**: Upload a voice sample, enter text, choose output language, receive speech audio
- Python backend handles:
  - File validation
  - Whisper transcription
  - Language detection and voice cloning
- Dockerized and runnable via `docker-compose`

---

## Environment Setup

### `frontend/bot/.env`

Set your Telegram bot token:

```env
TELEGRAM_BOT_TOKEN=your_token_here
````

---

### `backend/.env`

Configure backend settings:

```env
PORT=5000
MAX_FILE_MB=20
MAX_FILE_DURATION_SEC=360
CORS_ORIGIN=http://localhost:8080
```

---

## Running the Bot

### Prerequisites

* Docker
* Docker Compose

### Start the system:

```bash
docker-compose up --build
```
---
