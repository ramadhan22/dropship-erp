# Dropship ERP

Dropship ERP is a small ERP-style web application for managing dropshipping transactions. It consists of a Go backend and a React/TypeScript frontend.

## Backend

The backend lives under [`backend`](backend) and exposes a REST API using [Gin](https://github.com/gin-gonic/gin). It connects to PostgreSQL via [`sqlx`](https://github.com/jmoiron/sqlx) and uses embedded SQL migrations (powered by `golang-migrate`).

Main features:

- Import dropship purchases from CSV files.
- Import settled Shopee orders from XLSX files.
- Reconcile purchases with orders which creates journal entries and lines.
- Calculate revenue/COGS/fees/net profit metrics and cache them.
- Retrieve balance sheet data grouped by Assets, Liabilities and Equity.

Configuration is read from `backend/config.yaml` (values can be overridden with environment variables). On startup the application runs database migrations automatically.

To start the backend:

```bash
cd backend
go run ./cmd/api
```

Run tests with:

```bash
go test ./...
```

## Frontend

The UI resides in [`frontend/dropship-erp-ui`](frontend/dropship-erp-ui). It is a Vite powered React + TypeScript project using Material UI, React Query and Recharts for graphing. The home page now shows a sales summary chart that can be filtered by channel, store and date.

To develop the frontend:

```bash
cd frontend/dropship-erp-ui
npm install
npm run dev
```

## Repository Layout

- `backend/` – Go services, handlers and PostgreSQL migrations
- `frontend/` – React web application
- `.gitignore` – ignores node modules, Go build artifacts and environment files

## License

This repository was provided for demonstration purposes and does not include a specific license.
