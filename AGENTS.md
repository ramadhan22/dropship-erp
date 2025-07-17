# AGENTS Instructions

These guidelines apply to all directories in this repository. They are intended to help future code generation tools work with a consistent workflow.

## Testing
- Use mock servers for Shopee API interactions in unit tests.
- Run `go test ./...` from `backend` when Go code or migrations are modified.
- Run `npm test` in `frontend/dropship-erp-ui` for changes to the React app.

If a command cannot be executed in the environment, mention this in the pull request notes.

## Formatting
- Format Go code with `gofmt -w` before committing.
- Run `npm run lint` in `frontend/dropship-erp-ui` to check TypeScript/JavaScript style.
## Version Control
- Before finishing a task or committing changes, run `git pull origin master` to ensure your branch is up to date and minimize merge conflicts.

## Documentation
- Update `README.md` when adding new commands, dependencies or features.
- Keep this `AGENTS.md` up to date with any workflow changes.
- Prefer passing configuration into services rather than calling
  `config.MustLoadConfig` from within business logic. This keeps tests
  independent from external files.
- When adding a new page or API endpoint, adjust the backend CORS
  configuration so the page can be accessed without cross-origin errors.

## Logging
- Use `log.Printf` for informational messages and `logutil.Errorf` for errors.
- Log each outbound HTTP request in the backend before it is sent so requests can be traced.
- For Shopee API calls (e.g. `FetchShopeeOrderDetail`), ensure a log entry is written for every request so integrations can be audited.
- When an API requires an `access_token`, generate the `sign` value using
  `partner_id + api path + timestamp + access_token + shop_id` hashed with the
  `partner_key` via HMAC-SHA256.
- Before calling any Shopee API, confirm the access token has not expired and refresh it if necessary.
- Add start and completion logs for critical service and repository operations to aid debugging.
- Always include the error message in each error log across all services and repositories, especially when hitting Shopee API endpoints.
- Write application logs to `logs/YYYY-MM-DD.log` (configurable via `logging.dir`) with a new file created each day.

## Import Guidelines
- When creating an import feature, first delete any existing data for the
  same key so repeated imports overwrite previous rows.
- Remove any related journal entries when old data is purged to prevent
  duplicate postings.
- When generating journal entries (for example during dropship purchase
  imports), ensure the total debit amount equals the total credit amount so
  accounts like *Pendapatan Operasional* and pending receivables are not
  doubled.
- When storing data that originates as JSON (e.g. API responses), avoid JSONB
  columns.  Expand the fields into normal table columns using appropriate data
  types during the migration so values can be queried and indexed easily.
- When adding a new database table, also create a matching frontend page that
  lists the records using `SortableTable` with filtering, pagination and sorting.
  Provide a detail modal so users can inspect all fields.

## UI Patterns
- Use the `SortableTable` component for displaying tabular data.
- Paginate long lists with the `usePagination` hook or `Pagination` component.
- Provide filter controls on pages that list data so users can refine results.
- Display JSON or array values in modals with the `JsonTabs` component and
  format field labels by replacing underscores with spaces and capitalizing
  each word.

## Feature ↔ Table Mapping
The backend tables drive distinct pages and APIs in the frontend. When changing
models or introducing a new page, keep this mapping in sync.

- `accounts` – used on **AccountPage** for CRUD and referenced throughout the
  journal, balance sheet and expenses features.
- `journal_entries` & `journal_lines` – listed on **JournalPage** and summarised
  by **GLPage** and **PLPage**.
- `dropship_purchases` & `dropship_purchase_details` – imported via
  **DropshipImport** and reconciled with Shopee data; also feeds sales profit
  metrics.
- `shopee_settled_orders`/`shopee_settled` – data for **SalesProfitPage** and
  affiliate calculations.
- `reconciled_transactions` – managed on **ReconcileDashboard**.
- `cached_metrics` – displayed on **MetricsPage** and used for profit & loss as
  well as the balance sheet.
- `expenses` & `expense_lines` – recorded on **ExpensePage**.
- `ad_invoices` – uploaded on **AdInvoicePage**.
- `asset_accounts` – listed and adjusted on **KasAccountPage**.
- `jenis_channels` & `stores` – maintained on **ChannelPage** and used across
  filters.

  - `shopee_order_details`, `shopee_order_items` & `shopee_order_packages` – store
    Shopee order information for reconciliation and analysis. Displayed on
    **ShopeeOrderDetailPage**.

## Service ↔ Table Mapping
Understanding which backend service manipulates which tables helps when adding
features or debugging issues.

- **AccountService** – CRUD operations for `accounts`.
- **JournalService** – manages `journal_entries` and `journal_lines` directly.
- **DropshipService** – imports `dropship_purchases` & `dropship_purchase_details`
  and posts pending sales to the journal tables.
- **ShopeeService** – imports `shopee_settled_orders`, `shopee_settled` and
  `shopee_affiliate_sales`; updates dropship statuses and writes journal
  entries/lines.
  When reconciling settled orders, compare
  `shopee_settled.harga_asli_produk` with the sum of
  `dropship_purchase_details.total_harga_produk_channel` for the matching
  invoice.
- **ReconcileService** – records `reconciled_transactions` and creates matching
  journal entries and lines.
- DropshipService and ReconcileService – save Shopee order details in
  `shopee_order_details`, `shopee_order_items` and `shopee_order_packages`.
- **ExpenseService** – stores `expenses` and `expense_lines` while journaling the
  totals.
- **AdInvoiceService** – parses invoices into `ad_invoices` and adds journal
  entries for advertising fees.
- **AssetAccountService** – manages `asset_accounts` and adjusts balances via the
  journal.
- **ChannelService** – CRUD for `jenis_channels` and `stores` tables.
- **MetricService**, **PLService** and **ProfitLossReportService** – read
  `cached_metrics`, `journal_entries`, `dropship_purchases` and
  `shopee_settled_orders` to compute reports.
- **BalanceService** and **GLService** – read from the journal tables and
  `accounts` to build balance sheets and ledgers.

Update both this guide and `README.md` if you add or remove a table or page.

