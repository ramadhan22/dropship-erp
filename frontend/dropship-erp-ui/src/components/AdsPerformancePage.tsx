import {
  Alert,
  Button,
  Card,
  CardContent,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  Typography,
  Box,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { Refresh, TrendingUp, TrendingDown, BarChart } from "@mui/icons-material";
import SortableTable from "./SortableTable";
import { listAllStores } from "../api";
import type { Store } from "../types";

// Types for ads performance data
interface AdsCampaign {
  campaign_id: number;
  store_id: number;
  campaign_name: string;
  campaign_type?: string;
  campaign_status: string;
  placement_type?: string;
  daily_budget?: number;
  total_budget?: number;
  target_roas?: number;
  start_date?: string;
  end_date?: string;
  created_at: string;
  updated_at: string;
  latest_metrics?: {
    total_clicks: number;
    total_orders: number;
    sales_from_ads: number;
    ad_costs: number;
    roas: number;
    click_percentage: number;
    conversion_rate: number;
  };
}

interface AdsPerformanceSummary {
  total_campaigns: number;
  active_campaigns: number;
  total_ads_viewed: number;
  total_clicks: number;
  overall_click_percent: number;
  total_orders: number;
  total_products_sold: number;
  total_sales_from_ads: number;
  total_ad_costs: number;
  overall_roas: number;
  overall_conversion_rate: number;
  date_range: string;
  store_filter?: number;
}

export default function AdsPerformancePage() {
  const [stores, setStores] = useState<Store[]>([]);
  const [selectedStore, setSelectedStore] = useState<number | "">("");
  const [campaigns, setCampaigns] = useState<AdsCampaign[]>([]);
  const [summary, setSummary] = useState<AdsPerformanceSummary | null>(null);
  const [startDate, setStartDate] = useState<Date>(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)); // 30 days ago
  const [endDate, setEndDate] = useState<Date>(new Date());
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [loading, setLoading] = useState(false);
  const [fetchDialogOpen, setFetchDialogOpen] = useState(false);
  const [accessToken, setAccessToken] = useState("");
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Fetch stores on component mount
  useEffect(() => {
    listAllStores().then(setStores);
  }, []);

  // Fetch data when filters change
  useEffect(() => {
    fetchData();
  }, [selectedStore, startDate, endDate, statusFilter]);

  const fetchData = async () => {
    setLoading(true);
    try {
      await Promise.all([fetchCampaigns(), fetchSummary()]);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setLoading(false);
    }
  };

  const fetchCampaigns = async () => {
    const params = new URLSearchParams();
    if (selectedStore) params.set("store_id", selectedStore.toString());
    if (statusFilter !== "all") params.set("status", statusFilter);
    params.set("limit", "100");

    const response = await fetch(`/api/ads/campaigns?${params}`);
    if (!response.ok) throw new Error("Failed to fetch campaigns");
    const data = await response.json();
    setCampaigns(data.campaigns || []);
  };

  const fetchSummary = async () => {
    const params = new URLSearchParams();
    if (selectedStore) params.set("store_id", selectedStore.toString());
    params.set("start_date", startDate.toISOString().split("T")[0]);
    params.set("end_date", endDate.toISOString().split("T")[0]);

    const response = await fetch(`/api/ads/summary?${params}`);
    if (!response.ok) throw new Error("Failed to fetch summary");
    const data = await response.json();
    setSummary(data);
  };

  const handleFetchFromShopee = async () => {
    if (!selectedStore || !accessToken) {
      setMsg({ type: "error", text: "Please select a store and enter access token" });
      return;
    }

    setLoading(true);
    try {
      // Fetch campaigns
      const campaignResponse = await fetch("/api/ads/campaigns/fetch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          store_id: selectedStore,
          access_token: accessToken,
        }),
      });

      if (!campaignResponse.ok) throw new Error("Failed to fetch campaigns from Shopee");

      // Fetch performance data for each campaign
      for (const campaign of campaigns) {
        await fetch("/api/ads/performance/fetch", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            store_id: selectedStore,
            campaign_id: campaign.campaign_id,
            start_date: startDate.toISOString().split("T")[0],
            end_date: endDate.toISOString().split("T")[0],
            access_token: accessToken,
          }),
        });
      }

      setMsg({ type: "success", text: "Data fetched from Shopee successfully" });
      setFetchDialogOpen(false);
      setAccessToken("");
      fetchData(); // Refresh data
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    return `${(value * 100).toFixed(2)}%`;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "ongoing":
        return "success";
      case "paused":
        return "warning";
      case "ended":
        return "default";
      default:
        return "secondary";
    }
  };

  // Campaign table columns
  const campaignColumns: Array<{
    label: string;
    key?: keyof AdsCampaign;
    render?: (value: any, campaign: AdsCampaign) => React.ReactNode;
  }> = [
    { 
      label: "Campaign Name",
      render: (_, campaign: AdsCampaign) => (
        <Box>
          <Typography variant="body2" fontWeight="bold">
            {campaign.campaign_name}
          </Typography>
          <Typography variant="caption" color="textSecondary">
            ID: {campaign.campaign_id}
          </Typography>
        </Box>
      )
    },
    {
      label: "Status",
      render: (_, campaign: AdsCampaign) => (
        <Chip
          label={campaign.campaign_status}
          color={getStatusColor(campaign.campaign_status) as any}
          size="small"
        />
      ),
    },
    {
      label: "Type",
      render: (_, campaign: AdsCampaign) => campaign.campaign_type || "-",
    },
    {
      label: "Daily Budget",
      render: (_, campaign: AdsCampaign) =>
        campaign.daily_budget ? formatCurrency(campaign.daily_budget) : "-",
    },
    {
      label: "Target ROAS",
      render: (_, campaign: AdsCampaign) =>
        campaign.target_roas ? `${campaign.target_roas.toFixed(2)}x` : "-",
    },
    {
      label: "Performance",
      render: (_, campaign: AdsCampaign) => {
        const metrics = campaign.latest_metrics;
        if (!metrics) return "-";
        
        return (
          <Box>
            <Typography variant="caption" display="block">
              Clicks: {metrics.total_clicks.toLocaleString()}
            </Typography>
            <Typography variant="caption" display="block">
              Orders: {metrics.total_orders.toLocaleString()}
            </Typography>
            <Typography variant="caption" display="block">
              ROAS: {metrics.roas.toFixed(2)}x
            </Typography>
          </Box>
        );
      },
    },
    {
      label: "Spend / Revenue",
      render: (_, campaign: AdsCampaign) => {
        const metrics = campaign.latest_metrics;
        if (!metrics) return "-";
        
        return (
          <Box>
            <Typography variant="caption" display="block" color="error">
              Spend: {formatCurrency(metrics.ad_costs)}
            </Typography>
            <Typography variant="caption" display="block" color="success.main">
              Revenue: {formatCurrency(metrics.sales_from_ads)}
            </Typography>
          </Box>
        );
      },
    },
  ];

  return (
    <div>
      <Typography variant="h4" gutterBottom>
        Ads Performance Dashboard
      </Typography>

      {/* Filters */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Grid container spacing={2} alignItems="center">
            <Grid item xs={12} sm={6} md={2}>
              <FormControl fullWidth size="small">
                <InputLabel>Store</InputLabel>
                <Select
                  value={selectedStore}
                  onChange={(e) => setSelectedStore(e.target.value as number)}
                  label="Store"
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
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  label="Status"
                >
                  <MenuItem value="all">All Status</MenuItem>
                  <MenuItem value="ongoing">Ongoing</MenuItem>
                  <MenuItem value="paused">Paused</MenuItem>
                  <MenuItem value="ended">Ended</MenuItem>
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12} sm={6} md={2}>
              <LocalizationProvider dateAdapter={AdapterDateFns}>
                <DatePicker
                  label="Start Date"
                  value={startDate}
                  onChange={(date) => date && setStartDate(date)}
                  slotProps={{ textField: { size: "small", fullWidth: true } }}
                />
              </LocalizationProvider>
            </Grid>

            <Grid item xs={12} sm={6} md={2}>
              <LocalizationProvider dateAdapter={AdapterDateFns}>
                <DatePicker
                  label="End Date"
                  value={endDate}
                  onChange={(date) => date && setEndDate(date)}
                  slotProps={{ textField: { size: "small", fullWidth: true } }}
                />
              </LocalizationProvider>
            </Grid>

            <Grid item xs={12} sm={6} md={2}>
              <Button
                variant="outlined"
                startIcon={<Refresh />}
                onClick={fetchData}
                disabled={loading}
                fullWidth
              >
                Refresh
              </Button>
            </Grid>

            <Grid item xs={12} sm={6} md={2}>
              <Button
                variant="contained"
                onClick={() => setFetchDialogOpen(true)}
                disabled={loading || !selectedStore}
                fullWidth
              >
                Fetch from Shopee
              </Button>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Error/Success Messages */}
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }} onClose={() => setMsg(null)}>
          {msg.text}
        </Alert>
      )}

      {/* Summary Cards */}
      {summary && (
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom variant="caption">
                      Total Campaigns
                    </Typography>
                    <Typography variant="h5">
                      {summary.total_campaigns}
                    </Typography>
                    <Typography variant="caption" color="success.main">
                      {summary.active_campaigns} active
                    </Typography>
                  </Box>
                  <BarChart color="primary" />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom variant="caption">
                      Total Clicks
                    </Typography>
                    <Typography variant="h5">
                      {summary.total_clicks.toLocaleString()}
                    </Typography>
                    <Typography variant="caption">
                      CTR: {formatPercentage(summary.overall_click_percent)}
                    </Typography>
                  </Box>
                  <TrendingUp color="success" />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom variant="caption">
                      Total Orders
                    </Typography>
                    <Typography variant="h5">
                      {summary.total_orders.toLocaleString()}
                    </Typography>
                    <Typography variant="caption">
                      Conv Rate: {formatPercentage(summary.overall_conversion_rate)}
                    </Typography>
                  </Box>
                  <TrendingUp color="info" />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom variant="caption">
                      ROAS
                    </Typography>
                    <Typography variant="h5">
                      {summary.overall_roas.toFixed(2)}x
                    </Typography>
                    <Typography variant="caption">
                      {formatCurrency(summary.total_sales_from_ads)} revenue
                    </Typography>
                  </Box>
                  {summary.overall_roas >= 1 ? (
                    <TrendingUp color="success" />
                  ) : (
                    <TrendingDown color="error" />
                  )}
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {/* Campaign List */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Campaigns ({campaigns.length})
          </Typography>
          <SortableTable
            data={campaigns}
            columns={campaignColumns}
          />
        </CardContent>
      </Card>

      {/* Fetch Dialog */}
      <Dialog open={fetchDialogOpen} onClose={() => setFetchDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Fetch Data from Shopee</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="textSecondary" gutterBottom>
            Enter your Shopee access token to fetch the latest ads performance data.
          </Typography>
          <TextField
            fullWidth
            label="Access Token"
            value={accessToken}
            onChange={(e) => setAccessToken(e.target.value)}
            margin="normal"
            type="password"
            placeholder="Enter your Shopee access token"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setFetchDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleFetchFromShopee}
            variant="contained"
            disabled={!accessToken || loading}
          >
            {loading ? "Fetching..." : "Fetch Data"}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}