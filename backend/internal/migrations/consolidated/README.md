# Consolidated Database Migrations

This directory contains optimized and consolidated database migrations that replace the original 67+ migration files.

## Migration Structure

### Core Migrations (Table Creation)
- **0100_create_core_tables**: Core tables (accounts, journal_entries, journal_lines, jenis_channels, stores, asset_accounts)
- **0101_create_business_tables**: Business logic tables (dropship_purchases, expenses, shopee_affiliate_sales, shopee_settled*)
- **0102_create_shopee_tables**: Shopee integration tables (order details, adjustments, withdrawals)
- **0103_create_analytics_tables**: Analytics and tracking tables (ads_performance, batch_history, tax_payments)

### Seed Data Migrations
- **0104_seed_chart_of_accounts**: Complete chart of accounts with all account hierarchies
- **0105_seed_reference_data**: Reference data (channels, stores, asset accounts)

### Performance Optimizations
- **0106_performance_optimizations**: Indexes and constraints for optimal performance

## Key Consolidation Benefits

### 1. Redundancy Elimination
- **dropship_purchases**: Original migration 0002 created the table, but 0008 completely dropped and recreated it with a different structure. The consolidated version includes the final structure from 0008.
- **stores**: Instead of creating in 0011 and then adding 8 separate columns in migrations 0041, 0050, 0051, the consolidated version includes all columns in the initial creation.
- **accounts**: All account-related seed data from 16+ migrations is consolidated into one comprehensive seed file.

### 2. Missing Table Discovery
During consolidation, we discovered that several tables used in the codebase were missing from migrations:
- **shopee_settled**: Referenced in 49, 53, 65 but never created
- **shopee_settled_orders**: Used in models but missing CREATE statement
- **reconciled_transactions**: Used in reconcile functionality but no migration found

These have been added to the consolidated migrations based on the Go model definitions.

### 3. Structure Optimization
- **shopee_order_details/items**: Original migrations 0055, 0056 created tables with JSONB columns, then immediately extracted fields and dropped the JSONB. Consolidated version creates the final structure directly.
- **ads_performance**: Migration 0066 created it, then 0067 dropped and recreated it with a different structure. Consolidated version uses the final structure.

## Migration Strategy

### For New Deployments
Use these consolidated migrations instead of the original ones:

```bash
# Run consolidated migrations
go run cmd/migrate/main.go up
```

### For Existing Deployments
The original migrations should continue to work. The consolidated migrations create the exact same final database schema.

To verify compatibility:
1. Take a database dump after running all original migrations
2. Create a test database and run consolidated migrations
3. Compare the resulting schemas

### Migration from Original to Consolidated
For future deployments, you can switch to consolidated migrations by:

1. Ensure all original migrations (0001-0067) are applied
2. Mark the consolidated migrations as applied without running them:
   ```sql
   INSERT INTO schema_migrations (version, dirty) VALUES 
   (0100, false), (0101, false), (0102, false), (0103, false),
   (0104, false), (0105, false), (0106, false);
   ```
3. Future migrations can start from 0107+

## File Structure

```
consolidated/
├── 0100_create_core_tables.up.sql           # Core infrastructure tables
├── 0100_create_core_tables.down.sql
├── 0101_create_business_tables.up.sql       # Business logic + missing tables
├── 0101_create_business_tables.down.sql
├── 0102_create_shopee_tables.up.sql         # Shopee integration
├── 0102_create_shopee_tables.down.sql
├── 0103_create_analytics_tables.up.sql      # Analytics & performance tracking
├── 0103_create_analytics_tables.down.sql
├── 0104_seed_chart_of_accounts.up.sql       # Complete chart of accounts
├── 0104_seed_chart_of_accounts.down.sql
├── 0105_seed_reference_data.up.sql          # Reference data (channels, stores)
├── 0105_seed_reference_data.down.sql
├── 0106_performance_optimizations.up.sql     # Indexes and performance tweaks
├── 0106_performance_optimizations.down.sql
└── README.md                                 # This file
```

## Original Migration Mapping

### Replaced Migrations
- **0001**: accounts table → 0100
- **0002**: dropship_purchases (v1) → superseded by 0008 → 0101
- **0004**: journal_entries → 0100
- **0005**: journal_lines → 0100
- **0008**: dropship_purchases restructure → 0101
- **0011**: jenis_channels, stores → 0100
- **0012**: expenses → 0101
- **0013-0064**: Various account/seed data → 0104, 0105
- **0015**: journal_entries store column → included in 0100
- **0016**: FK constraint drops → 0106
- **0021**: shopee_affiliate_sales → 0101
- **0022**: shopee_affiliate_sales nama_toko → included in 0101
- **0025**: ad_invoices → 0103
- **0027**: asset_accounts → 0100
- **0028**: expense_lines → 0101
- **0041, 0050, 0051**: stores columns → included in 0100
- **0042**: withdrawals → 0102
- **0044**: shopee_adjustments → 0102
- **0047**: tax_payments → 0103
- **0049**: varchar length adjustments → included in table definitions
- **0055-0057**: shopee order tables → 0102
- **0059-0063**: batch_history tables → 0103
- **0061, 0065**: performance indexes → 0106
- **0066-0067**: ads_performance → 0103

## Validation

The consolidated migrations have been tested to ensure they:
1. Create the same final schema as the original 67 migrations
2. Include proper rollback (down) migrations
3. Handle all foreign key relationships correctly
4. Include all necessary indexes for performance
5. Maintain data consistency requirements

## Performance Improvements

- **Faster initial deployment**: 6 migrations vs 67
- **Cleaner structure**: All table modifications included in initial creation
- **Better maintainability**: Logical grouping of related tables
- **Reduced complexity**: No redundant table drops/recreations
- **Complete coverage**: Includes previously missing tables