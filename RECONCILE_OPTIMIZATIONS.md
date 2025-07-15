# Reconciliation Performance Optimizations

This document describes the performance optimizations implemented to address the "Reconcile All" process slowness and failed transaction status issues.

## Issues Addressed

### 1. Performance Bottlenecks in "Reconcile All" Process

**Problem:** The reconcile all process was slow due to:
- N+1 database queries when fetching DropshipPurchases individually
- Sequential processing without bulk operations
- Individual status checking for each transaction

**Solution:** 
- Added `GetDropshipPurchasesByInvoices()` bulk method to fetch multiple purchases in a single query
- Optimized `UpdateShopeeStatuses()` to use bulk fetching instead of individual calls
- Enhanced `ProcessReconcileBatch()` to bulk fetch purchases upfront and use in-memory lookups

### 2. Failed Transaction Status "shopee settled order not found"

**Problem:** Transactions showed "failed" status with "shopee settled order not found" message even when they were completed and escrow journals were posted.

**Root Cause:** The `CheckAndMarkComplete()` method was checking for entries in the `shopee_settled` table, but completed orders processed through escrow settlement don't always create entries there - they create journal entries with `source_type = "shopee_escrow"`.

**Solution:** 
- Updated `CheckAndMarkComplete()` to check multiple sources for completion status:
  1. First check if purchase is already marked as "Pesanan selesai"
  2. Check for escrow settlement journal entries (primary indicator)
  3. Fallback to checking `shopee_settled` table
- Added `ExistsBySourceTypeAndID()` method to efficiently check for journal entries

## Performance Improvements

### Database Query Optimization

**Before:**
```go
// N+1 query problem - one query per invoice
for _, inv := range invoices {
    dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, inv)
    // ... process individual purchase
}
```

**After:**
```go
// Single bulk query for all invoices
purchases, err := s.bulkGetDropshipPurchasesByInvoices(ctx, invoices)
// Create lookup map for O(1) access
purchaseMap := make(map[string]*models.DropshipPurchase)
for _, dp := range purchases {
    purchaseMap[dp.KodeInvoiceChannel] = dp
}
```

### Batch Processing Optimization

**Before:**
- Individual database calls for each transaction
- Sequential processing with repeated queries

**After:**  
- Bulk status updates
- Bulk purchase fetching
- In-memory lookups during processing
- Performance timing logs

## Code Changes Summary

### New Repository Methods

1. **`DropshipRepo.GetDropshipPurchasesByInvoices()`** - Bulk fetch purchases by invoice codes
2. **`JournalRepo.ExistsBySourceTypeAndID()`** - Check for journal entries by source type and ID

### Enhanced Service Methods

1. **`ReconcileService.UpdateShopeeStatuses()`** - Now uses bulk operations with performance logging
2. **`ReconcileService.CheckAndMarkComplete()`** - Improved logic to check multiple completion indicators  
3. **`ReconcileService.ProcessReconcileBatch()`** - Optimized with bulk operations and timing logs
4. **`ReconcileService.CreateReconcileBatches()`** - Returns detailed batch information

### Frontend Improvements

1. **Enhanced "Reconcile All" feedback** - Shows number of batches and transactions created
2. **Better error messages** - More specific and actionable error descriptions

## Performance Monitoring

The optimizations include comprehensive logging to monitor performance:

```
ProcessReconcileBatch 123: starting batch processing
ProcessReconcileBatch 123: status update completed in 2.1s  
ProcessReconcileBatch 123: bulk fetch completed in 150ms for 50 purchases
ProcessReconcileBatch 123 completed in 3.2s: 48/50 successful
```

## Testing

- Added unit tests to verify bulk operations work correctly
- Added fallback tests when bulk methods are not available
- All existing tests continue to pass

## Expected Performance Impact

- **Database queries reduced** from O(N) to O(1) for batch operations
- **Response time improved** for "Reconcile All" button (should show "batch created" message faster)
- **Status accuracy improved** - fewer false "failed" statuses due to better completion detection
- **Better observability** - timing logs help identify remaining bottlenecks

## Usage

The optimizations are transparent to users. The "Reconcile All" process now:

1. Responds faster with detailed batch creation information
2. Processes batches more efficiently in the background
3. Shows more accurate status results
4. Provides better error messages when issues occur

No configuration changes or database migrations are required.