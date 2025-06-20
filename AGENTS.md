# AGENTS Instructions

These guidelines apply to all directories in this repository. They are intended to help future code generation tools work with a consistent workflow.

## Testing
- Run `go test ./...` from `backend` when Go code or migrations are modified.
- Run `npm test` in `frontend/dropship-erp-ui` for changes to the React app.

If a command cannot be executed in the environment, mention this in the pull request notes.

## Formatting
- Format Go code with `gofmt -w` before committing.
- Run `npm run lint` in `frontend/dropship-erp-ui` to check TypeScript/JavaScript style.

## Documentation
- Update `README.md` when adding new commands, dependencies or features.
- Keep this `AGENTS.md` up to date with any workflow changes.

## UI Patterns
- Use the `SortableTable` component for displaying tabular data.
- Paginate long lists with the `usePagination` hook or `Pagination` component.
- Provide filter controls on pages that list data so users can refine results.

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
- **ReconcileService** – records `reconciled_transactions` and creates matching
  journal entries and lines.
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

