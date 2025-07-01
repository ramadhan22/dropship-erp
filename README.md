# Dropship ERP

Dropship ERP is a full featured web application for managing dropshipping and
online marketplace transactions.  The project consists of a Go backend and a
React/TypeScript frontend.

## Dependencies

The backend and frontend only require Go and Node.js. Parsing ad invoice PDFs
depends on the `pdftotext` utility from the `poppler-utils` package. Install this
system package so invoice imports work correctly.

## Backend

The backend lives under [`backend`](backend) and exposes a REST API using [Gin](https://github.com/gin-gonic/gin). It connects to PostgreSQL via [`sqlx`](https://github.com/jmoiron/sqlx) and uses embedded SQL migrations (powered by `golang-migrate`).

Main features:

- Import dropship purchases from CSV files with optional channel filter.
- Import settled Shopee orders from XLSX files.
- Import Shopee affiliate conversions from CSV. Journal entries are recorded only when the row's `status_terverifikasi` is `Sah`.
- Reconcile purchases with marketplace orders which creates journal entries and
  lines.
- Automatically compute revenue, COGS, fees and net profit metrics.
- Sales Profit page shows discounts and links to all related journal entries.
- View general ledger, balance sheet and profit and loss pages.
- Manage channels, accounts and expenses. Expenses can be edited with a selectable date and the previous journal is reversed automatically.
- Store detail pages automatically save Shopee `code` and `shop_id` values when provided in the callback URL.
  The page is accessible via `/stores/:id` either directly or via the detail button on the Channel page.

Configuration is read from `backend/config.yaml` and values can be overridden
with environment variables. On startup the application runs database migrations
automatically.

Shopee API calls require credentials including a long-lived `refresh_token`.
`ShopeeClient` automatically refreshes the short-lived access token on each
request using this value.
Order detail requests use the `SHOPEE_PARTNER_ID`, `SHOPEE_PARTNER_KEY`,
`SHOPEE_SHOP_ID` and optional `SHOPEE_BASE_URL` environment variables for
signing API calls. `base_url_shopee` in `config.yaml` defines the Partner API
host used when generating authorization links. As of this version, `ShopeeClient`
no longer loads configuration inside `RefreshAccessToken`; all required values
are taken from the struct fields
initialized in `NewShopeeClient`.

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

Common UI elements include sortable tables with built-in pagination and filter
controls. The `SortableTable` component and `usePagination` hook are reused
across pages to keep behavior consistent.

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

## Data Mapping
The tables in the PostgreSQL database correspond to specific UI pages. Updating
models or pages should keep this list in sync.

- **accounts** – managed on `AccountPage` and referenced by journal, balance
  sheet and expense features.
- **journal_entries** and **journal_lines** – shown on `JournalPage` and
  summarised in `GLPage` and `PLPage`.
- **dropship_purchases** and **dropship_purchase_details** – uploaded through
  `DropshipImport` and reconciled with Shopee orders, forming sales metrics.
- **shopee_settled_orders`/`shopee_settled** – used on `SalesProfitPage` and for
  affiliate reports.
- **reconciled_transactions** – displayed in `ReconcileDashboard`.
- **cached_metrics** – provides data for `MetricsPage`, balance sheet and P&L.
- **expenses** and **expense_lines** – editable via `ExpensePage`.
- **ad_invoices** – imported from `AdInvoicePage`.
- **asset_accounts** – only accounts under code `1.1.1` appear on `KasAccountPage`.
 - **jenis_channels** and **stores** – maintained on `ChannelPage` and referenced
   across filters. `stores` now include optional `code_id` and `shop_id` for
   Shopee API authorization. `StoreDetailPage` allows saving these values from
   the OAuth callback.

### Service ↔ Table Mapping
Services in the backend interact with particular tables. Updating services or
tables should keep this mapping aligned.

- **AccountService** – operates on `accounts`.
 - **JournalService** – writes to `journal_entries` and `journal_lines`.
   Duplicate entries with the same `source_type` and `source_id` are removed
   before inserting new records.
- **DropshipService** – manages `dropship_purchases` and
  `dropship_purchase_details`, creating journal entries for pending sales.
- **ShopeeService** – imports `shopee_settled_orders`, `shopee_settled` and
  `shopee_affiliate_sales` while updating dropship purchases and journal lines.
- **ReconcileService** – logs `reconciled_transactions` and posts journals for
  matched orders.
- **ExpenseService** – saves `expenses` and `expense_lines` along with journal
  postings.
- **AdInvoiceService** – records `ad_invoices` and journals ad expenses.
- **AssetAccountService** – uses `asset_accounts` and journal entries to adjust
  balances.
- **ChannelService** – CRUD for `jenis_channels` and `stores`.
- **MetricService**, **PLService** and **ProfitLossReportService** – read
  `cached_metrics`, journal data, dropship purchases and Shopee orders to produce
  reports.
- **BalanceService** and **GLService** – read from journal tables and `accounts`
  for statements.

## Contribution Guidelines

Automated agents contributing to this repository should follow the
[AGENTS.md](AGENTS.md) instructions. In short:

- Format Go code with `gofmt -w` and run `go test ./...` inside `backend`.
- Run `npm run lint` and `npm test` inside `frontend/dropship-erp-ui`.
- Update this README or other documentation when new features or commands are
  introduced.
- Ensure the backend CORS configuration is kept in sync whenever new pages or
  API endpoints are added.

## License

This repository was provided for demonstration purposes and does not include a specific license.
