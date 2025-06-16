# Dropship ERP

Dropship ERP is a full featured web application for managing dropshipping and
online marketplace transactions.  The project consists of a Go backend and a
React/TypeScript frontend.

## Dependencies

The backend and frontend only require Go and Node.js. PDF parsing for ads
invoices is handled with a pure Go library so no additional system packages are
needed.

## Backend

The backend lives under [`backend`](backend) and exposes a REST API using [Gin](https://github.com/gin-gonic/gin). It connects to PostgreSQL via [`sqlx`](https://github.com/jmoiron/sqlx) and uses embedded SQL migrations (powered by `golang-migrate`).

Main features:

- Import dropship purchases from CSV files.
- Import settled Shopee orders from XLSX files.
- Import Shopee affiliate conversions from CSV.
- Reconcile purchases with marketplace orders which creates journal entries and
  lines.
- Automatically compute revenue, COGS, fees and net profit metrics.
- View general ledger, balance sheet and profit and loss pages.
- Manage channels, accounts and expenses.

Configuration is read from `backend/config.yaml` and values can be overridden
with environment variables. On startup the application runs database migrations
automatically.

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

The UI resides in [`frontend/dropship-erp-ui`](frontend/dropship-erp-ui). It is
a Vite powered React + TypeScript project using Material UI, React Query and
Recharts for graphing. The application provides dashboards for sales summaries,
profit & loss, balance sheet and general ledger as well as pages for
reconciliation and data imports.

To develop the frontend:

```bash
cd frontend/dropship-erp-ui
npm install
npm run dev
npm test
```

## Repository Layout

- `backend/` – Go services, handlers and PostgreSQL migrations
- `frontend/` – React web application
- `.gitignore` – ignores node modules, Go build artifacts and environment files

## License

This repository was provided for demonstration purposes and does not include a specific license.
