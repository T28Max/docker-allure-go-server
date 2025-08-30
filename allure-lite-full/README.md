# Allure-Lite (Go + React + Nginx + Redis + SQLite)

A lightweight, self-hosted to Allure Server built on **free Allure Report**:

- Fancy **React** dashboard (Vite + Tailwind)
- **Upload** Allure artifact (`artifact.tar.zst` or zip with `allure-results/`)
- **Generate** full Allure static site with `allure generate`
- **List / Open / Delete** reports via Web UI
- **Nginx** static hosting for reports + reverse proxy for API
- **SQLite** metadata, **Redis** job queue
- Optional S3/MinIO & Bitbucket hooks can be added later

## Quick Start
```bash
cp .env.example .env
docker compose up --build
# open http://localhost:8080
```

## Notes
- Web UI uploads are great for **manual** reports (100s of MB). For huge CI artifacts (GBs), use an external uploader + API endpoints for multipart/S3 (can be added later).
- Reports live under a Docker volume and are served at `/reports/<project>/<id>/` and `/reports/<project>/latest/`.
