import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  Chip,
  CircularProgress,
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import SortableTable from './SortableTable';
import type { Column } from './SortableTable';
import { formatCurrency } from '../utils/format';
import {
  getAdsPerformance,
  getAdsPerformanceSummary,
  refreshAdsData,
  triggerFullHistorySync,
  type AdsPerformance,
  type AdsPerformanceSummary,
  type AdsPerformanceFilter,
} from '../api/adsPerformance';
import { listAllStores } from '../api';
import type { Store } from '../types';

// Summary metrics card component
const MetricCard: React.FC<{
  title: string;
  value: string | number;
  subtitle?: string;
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'error';
}> = ({ title, value, subtitle, color = 'primary' }) => (
  <Card>
    <CardContent>
      <Typography color="textSecondary" gutterBottom variant="body2">
        {title}
      </Typography>
      <Typography variant="h4" component="div" color={color}>
        {value}
      </Typography>
      {subtitle && (
        <Typography color="textSecondary" variant="body2">
          {subtitle}
        </Typography>
      )}
    </CardContent>
  </Card>
);

export default function AdsPerformanceDashboard() {
  const [ads, setAds] = useState<AdsPerformance[]>([]);
  const [summary, setSummary] = useState<AdsPerformanceSummary | null>(null);
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Filter state
  const [filter, setFilter] = useState<AdsPerformanceFilter>({
    date_from: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0], // 7 days ago
    date_to: new Date().toISOString().split('T')[0], // today
    limit: 50,
    offset: 0,
  });

  // Chart data state
  const [chartData, setChartData] = useState<any[]>([]);
  const [chartMetric, setChartMetric] = useState<'roas' | 'clicks' | 'sales'>('roas');

  // Load stores on component mount
  useEffect(() => {
    const loadStores = async () => {
      try {
        const storesData = await listAllStores();
        setStores(storesData);
      } catch (err) {
        console.error('Failed to load stores:', err);
      }
    };
    loadStores();
  }, []);

  // Load data when filter changes
  useEffect(() => {
    loadData();
  }, [filter]);

  const loadData = async () => {
    setLoading(true);
    setError(null);

    try {
      const [adsData, summaryData] = await Promise.all([
        getAdsPerformance(filter),
        getAdsPerformanceSummary({
          store_id: filter.store_id,
          campaign_status: filter.campaign_status,
          campaign_type: filter.campaign_type,
          date_from: filter.date_from,
          date_to: filter.date_to,
        }),
      ]);

      setAds(adsData.ads);
      setSummary(summaryData);

      // Prepare chart data - group by date (extract date from performance_hour)
      const chartMap = new Map<string, any>();
      adsData.ads.forEach((ad) => {
        const dateKey = ad.performance_hour.split('T')[0]; // Extract date part from ISO datetime
        if (!chartMap.has(dateKey)) {
          chartMap.set(dateKey, {
            date: dateKey,
            roas: 0,
            clicks: 0,
            sales: 0,
            costs: 0,
            count: 0,
          });
        }
        const entry = chartMap.get(dateKey)!;
        entry.roas += ad.roas;
        entry.clicks += ad.total_clicks;
        entry.sales += ad.sales_from_ads;
        entry.costs += ad.ad_costs;
        entry.count += 1;
      });

      // Calculate averages for ROAS
      const chartDataArray = Array.from(chartMap.values()).map((entry) => ({
        ...entry,
        roas: entry.count > 0 ? entry.roas / entry.count : 0,
      }));

      setChartData(chartDataArray.sort((a, b) => a.date.localeCompare(b.date)));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleRefreshData = async () => {
    if (!filter.date_from || !filter.date_to) {
      setError('Please select date range');
      return;
    }

    setRefreshing(true);
    setError(null);
    setSuccess(null);

    try {
      await refreshAdsData({
        date_from: filter.date_from,
        date_to: filter.date_to,
        store_id: filter.store_id,
      });

      setSuccess('Ads data refreshed successfully');
      // Reload data after refresh
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to refresh data');
    } finally {
      setRefreshing(false);
    }
  };

  const handleSyncData = async () => {
    if (!filter.store_id) {
      setError('Please select a store to sync');
      return;
    }

    setSyncing(true);
    setError(null);
    setSuccess(null);

    try {
      const result = await triggerFullHistorySync(filter.store_id);
      setSuccess(`Sync job created successfully (Job ID: ${result.job_id}). This will run in the background.`);
      // Optionally reload data after sync starts
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start sync');
    } finally {
      setSyncing(false);
    }
  };

  // Table columns
  const columns: Column<AdsPerformance>[] = [
    {
      label: 'Campaign',
      key: 'campaign_name',
      render: (value, row) => (
        <Box>
          <Typography variant="body2" fontWeight="bold">
            {value || row.campaign_id}
          </Typography>
          <Typography variant="caption" color="textSecondary">
            {row.campaign_type}
          </Typography>
        </Box>
      ),
    },
    {
      label: 'Status',
      key: 'campaign_status',
      render: (value) => (
        <Chip
          label={value}
          size="small"
          color={
            value === 'running' ? 'success' :
            value === 'ended' ? 'default' :
            value === 'paused' ? 'warning' : 'primary'
          }
        />
      ),
    },
    {
      label: 'Performance Hour',
      key: 'performance_hour',
      render: (value) => new Date(value).toLocaleString(),
    },
    {
      label: 'Impressions',
      key: 'ads_viewed',
      align: 'right',
      render: (value) => value.toLocaleString(),
    },
    {
      label: 'Clicks',
      key: 'total_clicks',
      align: 'right',
      render: (value) => value.toLocaleString(),
    },
    {
      label: 'CTR',
      key: 'click_rate',
      align: 'right',
      render: (value) => `${(value * 100).toFixed(2)}%`,
    },
    {
      label: 'Orders',
      key: 'orders_count',
      align: 'right',
      render: (value) => value.toLocaleString(),
    },
    {
      label: 'Sales',
      key: 'sales_from_ads',
      align: 'right',
      render: (value) => formatCurrency(value),
    },
    {
      label: 'Ad Costs',
      key: 'ad_costs',
      align: 'right',
      render: (value) => formatCurrency(value),
    },
    {
      label: 'ROAS',
      key: 'roas',
      align: 'right',
      render: (value) => value.toFixed(2),
    },
    {
      label: 'Daily Budget',
      key: 'daily_budget',
      align: 'right',
      render: (value) => formatCurrency(value),
    },
  ];

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          Ads Performance Dashboard
        </Typography>

        {/* Filters */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Grid container spacing={2} alignItems="center">
            <Grid item xs={12} sm={6} md={3}>
              <FormControl fullWidth size="small">
                <InputLabel>Store</InputLabel>
                <Select
                  value={filter.store_id || ''}
                  label="Store"
                  onChange={(e) =>
                    setFilter({ ...filter, store_id: e.target.value ? Number(e.target.value) : undefined })
                  }
                >
                  <MenuItem value="">All Stores</MenuItem>
                  {stores.map((store) => (
                    <MenuItem key={store.store_id} value={store.store_id}>
                      {store.nama_toko}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
              <FormControl fullWidth size="small">
                <InputLabel>Status</InputLabel>
                <Select
                  value={filter.campaign_status || ''}
                  label="Status"
                  onChange={(e) =>
                    setFilter({ ...filter, campaign_status: e.target.value || undefined })
                  }
                >
                  <MenuItem value="">All Status</MenuItem>
                  <MenuItem value="running">Running</MenuItem>
                  <MenuItem value="paused">Paused</MenuItem>
                  <MenuItem value="ended">Ended</MenuItem>
                  <MenuItem value="scheduled">Scheduled</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
              <DatePicker
                label="From Date"
                value={filter.date_from ? new Date(filter.date_from) : null}
                onChange={(date) =>
                  setFilter({ ...filter, date_from: date?.toISOString().split('T')[0] })
                }
                slotProps={{ textField: { size: 'small', fullWidth: true } }}
              />
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
              <DatePicker
                label="To Date"
                value={filter.date_to ? new Date(filter.date_to) : null}
                onChange={(date) =>
                  setFilter({ ...filter, date_to: date?.toISOString().split('T')[0] })
                }
                slotProps={{ textField: { size: 'small', fullWidth: true } }}
              />
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
              <Button
                variant="contained"
                fullWidth
                onClick={handleRefreshData}
                disabled={refreshing}
                startIcon={refreshing ? <CircularProgress size={16} /> : undefined}
              >
                {refreshing ? 'Refreshing...' : 'Refresh Data'}
              </Button>
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
              <Button
                variant="outlined"
                fullWidth
                onClick={handleSyncData}
                disabled={syncing}
                startIcon={syncing ? <CircularProgress size={16} /> : undefined}
              >
                {syncing ? 'Syncing...' : 'Sync All Data'}
              </Button>
            </Grid>
          </Grid>
        </Paper>

        {/* Alerts */}
        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}
        {success && (
          <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess(null)}>
            {success}
          </Alert>
        )}

        {loading ? (
          <Box display="flex" justifyContent="center" p={4}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {/* Summary Metrics */}
            {summary && (
              <Grid container spacing={3} sx={{ mb: 3 }}>
                <Grid item xs={12} sm={6} md={3}>
                  <MetricCard
                    title="Total Impressions"
                    value={summary.total_ads_viewed.toLocaleString()}
                    subtitle="Ads Viewed"
                  />
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <MetricCard
                    title="Total Clicks"
                    value={summary.total_clicks.toLocaleString()}
                    subtitle={`${((summary.total_clicks / summary.total_ads_viewed) * 100 || 0).toFixed(2)}% CTR`}
                    color="secondary"
                  />
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <MetricCard
                    title="Total Sales"
                    value={formatCurrency(summary.total_sales_from_ads)}
                    subtitle={`${summary.total_orders.toLocaleString()} orders`}
                    color="success"
                  />
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <MetricCard
                    title="Average ROAS"
                    value={summary.average_roas.toFixed(2)}
                    subtitle={`${formatCurrency(summary.total_ad_costs)} spent`}
                    color={summary.average_roas >= 2 ? 'success' : summary.average_roas >= 1 ? 'warning' : 'error'}
                  />
                </Grid>
              </Grid>
            )}

            {/* Performance Chart */}
            <Paper sx={{ p: 3, mb: 3 }}>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h6">Performance Trends</Typography>
                <FormControl size="small">
                  <InputLabel>Metric</InputLabel>
                  <Select
                    value={chartMetric}
                    label="Metric"
                    onChange={(e) => setChartMetric(e.target.value as typeof chartMetric)}
                  >
                    <MenuItem value="roas">ROAS</MenuItem>
                    <MenuItem value="clicks">Clicks</MenuItem>
                    <MenuItem value="sales">Sales</MenuItem>
                  </Select>
                </FormControl>
              </Box>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" />
                  <YAxis />
                  <Tooltip
                    formatter={(value, name) => {
                      if (name === 'roas') return [Number(value).toFixed(2), 'ROAS'];
                      if (name === 'sales') return [formatCurrency(Number(value)), 'Sales'];
                      return [Number(value).toLocaleString(), 'Clicks'];
                    }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey={chartMetric}
                    stroke={chartMetric === 'roas' ? '#8884d8' : chartMetric === 'sales' ? '#82ca9d' : '#ffc658'}
                    strokeWidth={2}
                  />
                </LineChart>
              </ResponsiveContainer>
            </Paper>

            {/* Ads List Table */}
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Campaign Performance
              </Typography>
              <SortableTable
                columns={columns}
                data={ads}
                defaultSort={{ key: "performance_hour", direction: "desc" }}
              />
            </Paper>
          </>
        )}
      </Box>
    </LocalizationProvider>
  );
}