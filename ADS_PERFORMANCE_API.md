# Ads Performance Dashboard API

## Overview
The Ads Performance Dashboard provides comprehensive analytics for advertising campaigns across multiple stores, with integration to Shopee API v2 for real-time data fetching.

## API Endpoints

### GET /api/ads-performance
Retrieve ads performance data with filtering and pagination.

**Query Parameters:**
- `store_id` (optional): Filter by specific store ID
- `campaign_status` (optional): Filter by campaign status (running, paused, ended, scheduled)
- `campaign_type` (optional): Filter by campaign type
- `date_from` (optional): Start date filter (YYYY-MM-DD format)
- `date_to` (optional): End date filter (YYYY-MM-DD format)
- `limit` (optional): Number of records to return (default: 50, max: 100)
- `offset` (optional): Number of records to skip (default: 0)

**Response:**
```json
{
  "ads": [
    {
      "id": 1,
      "store_id": 1,
      "campaign_id": "12345",
      "campaign_name": "Summer Sale Campaign",
      "campaign_type": "search",
      "campaign_status": "running",
      "date_from": "2024-01-01",
      "date_to": "2024-01-07",
      "ads_viewed": 10000,
      "total_clicks": 500,
      "orders_count": 25,
      "products_sold": 30,
      "sales_from_ads": 1500.00,
      "ad_costs": 300.00,
      "click_rate": 0.05,
      "roas": 5.00,
      "daily_budget": 100.00,
      "target_roas": 4.00,
      "performance_change_percentage": 12.5,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-07T23:59:59Z"
    }
  ],
  "limit": 50,
  "offset": 0,
  "count": 1
}
```

### GET /api/ads-performance/summary
Get aggregated metrics across all campaigns for the specified filters.

**Query Parameters:** Same as above (except limit and offset)

**Response:**
```json
{
  "total_ads_viewed": 50000,
  "total_clicks": 2500,
  "total_orders": 125,
  "total_products_sold": 150,
  "total_sales_from_ads": 7500.00,
  "total_ad_costs": 1500.00,
  "average_click_rate": 0.05,
  "average_roas": 5.00,
  "date_from": "2024-01-01",
  "date_to": "2024-01-07"
}
```

### POST /api/ads-performance/refresh
Fetch fresh ads performance data from Shopee API for the specified date range.

**Request Body:**
```json
{
  "date_from": "2024-01-01",
  "date_to": "2024-01-07",
  "store_id": 1  // optional - if not provided, refreshes all stores
}
```

**Response:**
```json
{
  "message": "Ads data refreshed successfully",
  "date_from": "2024-01-01",
  "date_to": "2024-01-07"
}
```

## Database Schema

### ads_performance table
- `id` - Primary key
- `store_id` - Foreign key to stores table
- `campaign_id` - Shopee campaign identifier
- `campaign_name` - Campaign display name
- `campaign_type` - Type of campaign (search, discovery, etc.)
- `campaign_status` - Current status (running, paused, ended, scheduled)
- `date_from`, `date_to` - Date range for metrics
- Core metrics: `ads_viewed`, `total_clicks`, `orders_count`, `products_sold`, `sales_from_ads`, `ad_costs`
- Calculated metrics: `click_rate`, `roas`
- Budget info: `daily_budget`, `target_roas`
- Performance tracking: `performance_change_percentage`
- Timestamps: `created_at`, `updated_at`

## Features

### Dashboard Components
1. **Summary Metrics Panel** - Key performance indicators with color-coded status
2. **Performance Charts** - Interactive trend visualization with metric selection
3. **Campaign Table** - Sortable, filterable list of all campaigns
4. **Advanced Filtering** - By store, status, date range, campaign type

### Data Integration
- Shopee API v2 integration with rate limiting and retry logic
- Automatic calculation of derived metrics (CTR, ROAS)
- Bulk data refresh with error handling
- Historical data storage for trend analysis

### UI/UX Features
- Material UI components for consistent design
- Responsive layout for mobile and desktop
- Real-time error handling and success notifications
- Date range selection with validation
- Export-ready data structure

## Usage Examples

### Fetch campaign data for a specific store
```
GET /api/ads-performance?store_id=1&date_from=2024-01-01&date_to=2024-01-07
```

### Get summary for running campaigns only
```
GET /api/ads-performance/summary?campaign_status=running
```

### Refresh data for all stores
```
POST /api/ads-performance/refresh
{
  "date_from": "2024-01-01",
  "date_to": "2024-01-07"
}
```