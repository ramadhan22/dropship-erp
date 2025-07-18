# Dropship ERP

Dropship ERP is a full featured web application for managing dropshipping and
online marketplace transactions.  The project consists of a Go backend and a
React/TypeScript frontend.

## API Rate Limiting

The application implements robust rate limiting for Shopee API calls to comply with their usage policies:

- **100 requests per minute limit**: All Shopee API calls are rate-limited to a maximum of 100 requests per minute across all endpoints
- **Minute-based reset**: When the rate limit is reached, the system waits until the next minute before allowing additional requests
- **Comprehensive coverage**: Rate limiting applies to all Shopee API endpoints including:
  - Order details and batch fetching
  - Escrow detail requests
  - Wallet transaction lists
  - Return list requests
  - Access token refresh and authentication
  - Order status queries
- **Graceful handling**: When rate limits are exceeded, the system automatically waits for the next minute rather than failing requests
- **Context-aware**: Rate limiting respects context cancellation for proper request timeout handling

This ensures compliance with Shopee's API usage policies while maintaining system reliability.

The backend and frontend only require Go and Node.js. Parsing ad invoice PDFs
depends on the `pdftotext` utility from the `poppler-utils` package. Install this
system package so invoice imports work correctly.

## Backend

The backend lives under [`backend`](backend) and exposes a REST API using [Gin](https://github.com/gin-gonic/gin). It connects to PostgreSQL via [`sqlx`](https://github.com/jmoiron/sqlx) and uses embedded SQL migrations (powered by `golang-migrate`).

Main features:

- Import dropship purchases from CSV files with optional channel filter.
- Import settled Shopee orders from XLSX files.
- Import Shopee affiliate conversions from CSV. Journal entries are recorded only when the row's `status_terverifikasi` is `Sah`.
- Import Shopee adjustments from income reports with automatic journal postings.
  Existing entries for the same order/date/type are replaced and they are also
  captured when importing settled orders.
  Adjustments can be browsed on the **Shopee Adjustments** page.
 - Multi-file imports are queued in `batch_history` and processed asynchronously
   by a scheduler. The API response simply reports how many files were queued.
Adjustments may be edited or deleted; updating replaces the original
journal entry. Negative values are posted to a dedicated **Refund** account
instead of the Discount account. The **Shipping Fee Discrepancy** sheet of the
income report is also parsed so extra courier charges are imported as
adjustments. When posting Shopee escrow settlements the application now compares
the actual and estimated shipping fee and automatically records a *Shipping Fee
Discrepancy* adjustment when they differ.
- Reconcile purchases with marketplace orders which creates journal entries and
  lines.
- **Robust Error Handling**: Reconciliation processes now include comprehensive error handling that allows operations to continue even when individual transactions fail. Failed transactions are logged with detailed error information and can be retried later. This prevents entire batches from stopping due to single transaction errors.
- Check Shopee order details from the Reconcile dashboard using the store's saved access token.
- Browse stored Shopee order details on the **Order Details** page with a modal showing items and packages.
- Order detail lookups now save key fields in `shopee_order_details`,
  `shopee_order_items` and `shopee_order_packages` tables rather than raw JSON.
  All time values are converted to timestamps for easier analysis.
- Dropship CSV imports fetch Shopee order detail for each invoice to record
  pending sales amounts and save the raw detail. Order lookups are batched up
  to 50 invoices per request to reduce API calls. Transactions are skipped when
  the order detail cannot be retrieved. Order detail batches are now fetched
  concurrently for faster imports.
- Shopee order status is now fetched server-side when loading the Reconcile dashboard for faster rendering.
- Filter reconcile candidates by date range to limit results.
- Reconcile dashboard now supports filtering candidates by purchase status.
- Reconcile All now creates batch records grouped by store (50 invoices per batch) and returns immediately. The scheduler processes these batches asynchronously.
- Escrow details are fetched in batches of up to 50 orders when reconciling all, reducing API requests.
- Shopee reconciliation batches run in parallel for faster processing.
- Automatically compute revenue, COGS, fees and net profit metrics.
 - Sales Profit page shows discounts and links to all related journal entries.
   Shipping discounts (account `5.5.0.6`) and escrow journals are included when
   calculating profit. Adjustments including shipping fee discrepancies are also
   factored into profit calculations.
- View general ledger, balance sheet and profit and loss pages.
- Manage channels, accounts and expenses. Expenses can be edited with a selectable date and the previous journal is reversed automatically.
- Sales Summary dashboard now shows cancelled order count and total Biaya Mitra posted for those cancellations.
- Store detail pages automatically save Shopee `code` and `shop_id` values when provided in the callback URL.

### New Reconciliation API Endpoints

The following API endpoints support the enhanced error handling and reporting capabilities:

- **POST** `/api/reconcile/bulk` - Process multiple reconciliation pairs with error handling. Returns a detailed report of successful and failed transactions.
- **GET** `/api/reconcile/report?shop=<shop>&days=<days>` - Generate a comprehensive reconciliation report for the specified shop and time period.
- **POST** `/api/reconcile/retry` - Retry failed reconciliation transactions. Accepts `shop` and `max_retries` parameters.
- **GET** `/api/reconcile/summary?shop=<shop>&days=<days>` - Get a quick summary of failed reconciliations including failure categories.

These endpoints provide detailed error categorization, failure rate tracking, and automatic retry capabilities to improve operational resilience.
  The page is accessible via `/stores/:id` either directly or via the detail button on the Channel page.
- Pay UMKM final tax (0.5% of revenue) per store and period on the **Tax Payment** page. Journal entries are created automatically when paying.
- A dedicated `PPh Final UMKM` account (`5.4.1`) tracks these tax expenses.
- View Shopee wallet transactions by store on the **Wallet Transactions** page;
  pagination now uses a **More** button with filter dropdowns.
- View batch import history on the **Batch History** page with a button to see
  transaction-level results.

Configuration is read from `backend/config.yaml` and values can be overridden
with environment variables. On startup the application runs database migrations
automatically.
Application logs are written to the directory specified by `logging.dir` in the
config file. A new file named `YYYY-MM-DD.log` is created each day.
Batch operations that fetch Shopee data in parallel respect the `max_threads`
setting which limits how many goroutines run concurrently (defaults to `5`).

Shopee API calls require credentials including a long-lived `refresh_token`.
`ShopeeClient` refreshes the short-lived access token when it has expired rather
than on every request. To exchange the authorization `code` for tokens the
client issues a POST request to `/api/v2/auth/token/get` with JSON payload:

```json
{
  "shop_id": "<shop_id>",
  "code": "<authorization_code>"
}
```

Query parameters include `partner_id`, `sign` and `timestamp`.
Order detail requests use the `SHOPEE_PARTNER_ID`, `SHOPEE_PARTNER_KEY`,
`SHOPEE_SHOP_ID` and optional `SHOPEE_BASE_URL` environment variables for
signing API calls. When an endpoint requires an access token, generate the
signature as HMAC-SHA256 of `partner_id + api path + timestamp + access_token +
shop_id` using the partner key. `base_url_shopee` in `config.yaml` defines the
Partner API host used when generating authorization links. As of this version,
`ShopeeClient` no longer loads configuration inside `RefreshAccessToken`; all required values
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
Use mocks for Shopee API calls during unit tests to avoid network access.


## Frontend

The UI resides in [`frontend/dropship-erp-ui`](frontend/dropship-erp-ui). It is
a Vite powered React + TypeScript project using Material UI, Tailwind CSS,
React Query and Recharts for graphing. The application provides dashboards for
sales summaries, profit & loss, balance sheet and general ledger as well as
pages for reconciliation and data imports. A Tailwind based analytics dashboard
is available at `/dashboard`.


Common UI elements include sortable tables with built-in pagination and filter
controls. The `SortableTable` component and `usePagination` hook are reused
across pages to keep behavior consistent.
API requests trigger a global loading spinner by default; add an `X-Skip-Loading`
header to disable the overlay when a custom progress bar is shown.

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
- When database migrations involve JSON data from an external API, parse the
  fields and store them in regular columns instead of JSONB so the values can be
  queried directly.
- Whenever a new database table is introduced, add a corresponding page in the
  React app that displays the data with filtering, pagination, sorting and a
  detail modal so users can inspect rows easily.

## License

This repository was provided for demonstration purposes and does not include a specific license.
