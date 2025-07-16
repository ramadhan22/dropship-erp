# Returned Order Journal Entry API

## Overview

This feature implements automatic journal entry creation for returned orders in escrow settlements. It supports both full and partial returns with proper double-entry bookkeeping.

## API Endpoint

### Process Returned Order

**POST** `/reconcile/return`

Manually process a returned order and create appropriate journal entries.

#### Request Body

```json
{
  "invoice": "string",           // Required: Order invoice number
  "is_partial_return": boolean,  // Optional: Whether this is a partial return (default: false)
  "return_amount": number        // Required for partial returns: Amount being returned
}
```

#### Response

**Success (200 OK):**
```json
{
  "message": "Return processed successfully",
  "invoice": "ORDER123",
  "return_type": "full|partial",
  "amount": 50000.0
}
```

**Error Responses:**

- **409 Conflict**: Return already processed
```json
{
  "error": "Return journal already exists for this invoice"
}
```

- **400 Bad Request**: Invalid parameters
```json
{
  "error": "Return amount must be greater than 0 for partial returns"
}
```

## Usage Examples

### Full Return

```bash
curl -X POST http://localhost:8080/reconcile/return \
  -H "Content-Type: application/json" \
  -d '{
    "invoice": "ORDER123",
    "is_partial_return": false
  }'
```

### Partial Return

```bash
curl -X POST http://localhost:8080/reconcile/return \
  -H "Content-Type: application/json" \
  -d '{
    "invoice": "ORDER456", 
    "is_partial_return": true,
    "return_amount": 50000.0
  }'
```

## Automatic Processing

The system also automatically detects and processes returned orders when:

1. **Status Updates**: When `UpdateShopeeStatus` is called and order status contains "return"
2. **Batch Processing**: During "Reconcile All" operations, returned orders are processed automatically
3. **Bulk Status Updates**: When `UpdateShopeeStatuses` processes multiple orders

## Journal Entry Structure

### Full Return Example

For a returned order with original price 100,000 and fees:
- Commission: 5,000
- Service: 2,000  
- Voucher: 1,000

**Journal Entry Created:**
```
Dr. Pending Account (Store)     100,000
    Cr. Commission Fee                    5,000
    Cr. Service Fee                       2,000
    Cr. Voucher                           1,000
    Cr. Shipping Discount                 1,500
    Cr. Affiliate Fee                     3,000
    Dr. Refund Account              100,000
```

### Partial Return (50% return)

Same order but 50% returned (50,000):

**Journal Entry Created:**
```
Dr. Pending Account (Store)      50,000
    Cr. Commission Fee                    2,500
    Cr. Service Fee                       1,000
    Cr. Voucher                             500
    Cr. Shipping Discount                   750
    Cr. Affiliate Fee                     1,500
    Dr. Refund Account               50,000
```

## Account Mapping

| Account | ID | Purpose |
|---------|----|---------| 
| Refund | 52009 | Records customer refunds |
| Commission Fee | 52006 | Shopee commission charges |
| Service Fee | 52004 | Service fee charges |
| Voucher | 55001 | Voucher/discount amounts |
| Shipping Discount | 55006 | Shipping discounts |
| Affiliate Fee | 55002 | Affiliate commission |
| Pending Account | Store-specific | Pending receivables by store |

## Integration with Reconcile Dashboard

The returned order functionality integrates seamlessly with:

1. **Manual Processing**: Use the API endpoint for one-off return processing
2. **Batch Operations**: Automatically included in "Reconcile All" operations  
3. **Status Monitoring**: Returns are detected during regular status updates
4. **Audit Trail**: All return journal entries are properly logged with source tracking

## Error Handling

- **Duplicate Prevention**: System checks for existing return journals
- **Balance Validation**: All journal entries maintain debit/credit balance
- **Status Tracking**: Order status is updated to reflect return processing
- **Escrow Validation**: Return amounts are validated against escrow details