# 🤖 AI Assistant Context Guide (README_LLM)

This document is optimized for LLMs (Large Language Models) to understand the project structure, logic flow, and constraints quickly.

## 🏗️ Technical Stack
- **Backend**: Go 1.25+, Fiber Web Framework.
- **Frontend**: Vue 3 (Script Setup), Vite, Ant Design Vue 4.x, Tailwind CSS.
- **Database**: SQLite (via `modernc.org/sqlite` - CGO-free).
- **Processing**: System `cwebp` binary for image conversion.
- **DevOps**: Docker (Multi-stage), GitHub Actions (GHCR).

## 📂 Key Architecture
- `/main.go`: Application entry, middleware setup, and routing.
- `/handlers/`: Logical controllers (Image processing, Auth, Admins).
- `/middleware/`: JWT-based `UserAuth` and custom header `AdminAuth`.
- `/utils/webp.go`: Shells out to `cwebp` for uuid-named conversion.
- `/web/src/services/api.js`: Centralized Axios interface supporting FormData.

## 🔐 Authentication Logic
- **Users**: standard JWT via `Authorization: Bearer <token>`.
- **Admin**: simplistic but effective `X-Admin-Token` header for protected routes, validated against session-base logic in `/handlers/admin.go`.

## ⚠️ Known Constraints & Quirks
- **WebP Conversion**: Requires `libwebp-tools` installed in the environment (provided in Dockerfile).
- **CGO**: Disabled (`CGO_ENABLED=0`) to ensure Alpine compatibility.
- **Upload Limits**: Server `BodyLimit` is 50MB. Nginx `client_max_body_size` must match.
- **SPA Routing**: Go handles API routes; all other paths fall back to `./web/dist/index.html`.

## 🛠️ Common Modification Tasks
- **Adding API**: Update `main.go` route group -> Create handler in `/handlers` -> Add method in `api.js`.
- **UI Tweaks**: Components are in `/web/src/views`. Use Ant Design Vue props first before custom CSS.
