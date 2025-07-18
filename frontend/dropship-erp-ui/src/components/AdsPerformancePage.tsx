import {
  Alert,
  Button,
  Card,
  CardContent,
  Typography,
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
} from "@mui/material";
import { useEffect, useState, useCallback } from "react";
import { 
  fetchAdsCampaigns, 
  fetchAdsPerformanceSummary, 
  syncHistoricalAdsPerformance 
} from "../api/adsPerformance";
import { listAllStoresDirect } from "../api";

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
  bidding_method?: string;
  campaign_budget?: number;
  item_id_list?: string; // JSON array as string
  enhanced_cpc?: boolean;
  created_at: string;
  updated_at: string;
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

interface FilterState {
  store_id: string;
  status: string;
  date_range: string;
  start_date: string;
  end_date: string;
}

// Define available status options
const STATUS_OPTIONS = [
  { value: "", label: "All" },
  { value: "Terjadwal", label: "Terjadwal" },
  { value: "Berjalan", label: "Berjalan" },
  { value: "Nonaktif", label: "Nonaktif" },
  { value: "Berakhir", label: "Berakhir" },
  { value: "Dihapus", label: "Dihapus" },
];

// Define available date range presets
const DATE_RANGE_OPTIONS = [
  { value: "", label: "Custom" },
  { value: "today", label: "Today" },
  { value: "current_week", label: "Current Week" },
  { value: "current_month", label: "Current Month" },
  { value: "current_year", label: "Current Year" },
];

export default function AdsPerformancePage() {
  const [campaigns, setCampaigns] = useState<AdsCampaign[]>([]);
  const [summary, setSummary] = useState<AdsPerformanceSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [syncDialogOpen, setSyncDialogOpen] = useState(false);
  const [syncForm, setSyncForm] = useState({
    store_id: "",
  });
  const [stores, setStores] = useState<{ store_id: number; nama_toko: string }[]>([]);
  const [storesLoading, setStoresLoading] = useState(false);
  const [storesError, setStoresError] = useState<string | null>(null);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Filter state
  const [filters, setFilters] = useState<FilterState>({
    store_id: "",
    status: "",
    date_range: "current_month", // Default to current month
    start_date: "",
    end_date: "",
  });

  const fetchStores = useCallback(async () => {
    setStoresLoading(true);
    setStoresError(null);
    try {
      const response = await listAllStoresDirect();
      const data = response.data;
      setStores(data || []);
      
      if (!data || data.length === 0) {
        setStoresError("No stores found. Please create stores first.");
      }
    } catch (error) {
      console.error("Error fetching stores:", error);
      setStoresError(error instanceof Error ? error.message : "Failed to fetch stores");
    } finally {
      setStoresLoading(false);
    }
  }, []);

  const fetchCampaignsData = useCallback(async () => {
    try {
      const params: any = { limit: 100 };
      
      if (filters.store_id) {
        params.store_id = parseInt(filters.store_id);
      }
      
      if (filters.status) {
        params.status = filters.status;
      }
      
      if (filters.date_range) {
        params.date_range = filters.date_range;
      } else {
        // Use custom date range if no preset selected
        if (filters.start_date) {
          params.start_date = filters.start_date;
        }
        if (filters.end_date) {
          params.end_date = filters.end_date;
        }
      }
      
      const data = await fetchAdsCampaigns(params);
      setCampaigns(data.campaigns || []);
    } catch (error) {
      console.error("Error fetching campaigns:", error);
    }
  }, [filters]);

  const fetchSummaryData = useCallback(async () => {
    try {
      const params: any = {};
      
      if (filters.store_id) {
        params.store_id = parseInt(filters.store_id);
      }
      
      if (filters.date_range) {
        params.date_range = filters.date_range;
      } else {
        // Use custom date range if no preset selected
        if (filters.start_date) {
          params.start_date = filters.start_date;
        }
        if (filters.end_date) {
          params.end_date = filters.end_date;
        }
      }
      
      const data = await fetchAdsPerformanceSummary(params);
      setSummary(data);
    } catch (error) {
      console.error("Error fetching summary:", error);
    }
  }, [filters]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      await Promise.all([fetchCampaignsData(), fetchSummaryData()]);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setLoading(false);
    }
  }, [fetchCampaignsData, fetchSummaryData]);

  // Fetch stores and initial data on component mount
  useEffect(() => {
    fetchStores();
    fetchData();
  }, [fetchStores, fetchData]);

  const handleSyncDialogOpen = () => {
    setSyncDialogOpen(true);
  };

  const handleSyncDialogClose = () => {
    setSyncDialogOpen(false);
    setSyncForm({ store_id: "" });
  };

  const handleSyncSubmit = async () => {
    if (stores.length === 0) {
      setMsg({ type: "error", text: "No stores available. Please create stores first." });
      return;
    }
    
    if (!syncForm.store_id) {
      setMsg({ type: "error", text: "Please select a store" });
      return;
    }

    setSyncing(true);
    try {
      const result = await syncHistoricalAdsPerformance({
        store_id: parseInt(syncForm.store_id),
      });
      
      setMsg({ 
        type: "success", 
        text: `Historical sync started successfully. Batch ID: ${result.batch_id}` 
      });
      handleSyncDialogClose();
    } catch (error: any) {
      setMsg({ 
        type: "error", 
        text: error.message || "Failed to start historical sync" 
      });
    } finally {
      setSyncing(false);
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
      case "Berjalan":
        return "success";
      case "Nonaktif":
        return "warning";
      case "Berakhir":
        return "default";
      case "Terjadwal":
        return "info";
      case "Dihapus":
        return "error";
      default:
        return "secondary";
    }
  };

  const handleFilterChange = (field: keyof FilterState, value: string) => {
    setFilters(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const resetFilters = () => {
    setFilters({
      store_id: "",
      status: "",
      date_range: "current_month",
      start_date: "",
      end_date: "",
    });
  };

  return (
    <div>
      <Typography variant="h4" gutterBottom>
        Ads Performance Dashboard
      </Typography>

      {/* Filter Section */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Filters
          </Typography>
          <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", mb: 2 }}>
            <FormControl sx={{ minWidth: 200 }}>
              <InputLabel>Store</InputLabel>
              <Select
                value={filters.store_id}
                onChange={(e) => handleFilterChange("store_id", e.target.value)}
                label="Store"
                disabled={storesLoading}
              >
                <MenuItem value="">All Stores</MenuItem>
                {stores.map((store) => (
                  <MenuItem key={store.store_id} value={store.store_id.toString()}>
                    {store.nama_toko}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl sx={{ minWidth: 200 }}>
              <InputLabel>Status</InputLabel>
              <Select
                value={filters.status}
                onChange={(e) => handleFilterChange("status", e.target.value)}
                label="Status"
              >
                {STATUS_OPTIONS.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl sx={{ minWidth: 200 }}>
              <InputLabel>Date Range</InputLabel>
              <Select
                value={filters.date_range}
                onChange={(e) => handleFilterChange("date_range", e.target.value)}
                label="Date Range"
              >
                {DATE_RANGE_OPTIONS.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {/* Custom Date Range Fields (shown only when date_range is empty) */}
            {!filters.date_range && (
              <>
                <TextField
                  label="Start Date"
                  type="date"
                  value={filters.start_date}
                  onChange={(e) => handleFilterChange("start_date", e.target.value)}
                  InputLabelProps={{
                    shrink: true,
                  }}
                  sx={{ minWidth: 200 }}
                />
                <TextField
                  label="End Date"
                  type="date"
                  value={filters.end_date}
                  onChange={(e) => handleFilterChange("end_date", e.target.value)}
                  InputLabelProps={{
                    shrink: true,
                  }}
                  sx={{ minWidth: 200 }}
                />
              </>
            )}
          </Box>

          <Box sx={{ display: "flex", gap: 1 }}>
            <Button
              variant="contained"
              onClick={fetchData}
              disabled={loading}
            >
              {loading ? "Loading..." : "Apply Filters"}
            </Button>
            <Button
              variant="outlined"
              onClick={resetFilters}
              disabled={loading}
            >
              Reset
            </Button>
          </Box>
        </CardContent>
      </Card>

      <Box sx={{ mb: 3 }}>
        <Button
          variant="outlined"
          onClick={handleSyncDialogOpen}
          disabled={syncing}
          sx={{ mr: 2 }}
        >
          {syncing ? "Syncing..." : "Sync Historical Data"}
        </Button>
      </Box>

      {/* Error/Success Messages */}
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }} onClose={() => setMsg(null)}>
          {msg.text}
        </Alert>
      )}

      {/* Summary Cards */}
      {summary && (
        <Box sx={{ display: "flex", gap: 2, mb: 4, flexWrap: "wrap" }}>
          <Card sx={{ minWidth: 200 }}>
            <CardContent>
              <Typography color="textSecondary" gutterBottom variant="caption">
                Total Campaigns
              </Typography>
              <Typography variant="h5">
                {summary.total_campaigns}
              </Typography>
              <Typography variant="caption" color="success.main">
                {summary.active_campaigns} active
              </Typography>
            </CardContent>
          </Card>

          <Card sx={{ minWidth: 200 }}>
            <CardContent>
              <Typography color="textSecondary" gutterBottom variant="caption">
                Total Clicks
              </Typography>
              <Typography variant="h5">
                {summary.total_clicks.toLocaleString()}
              </Typography>
              <Typography variant="caption">
                CTR: {formatPercentage(summary.overall_click_percent)}
              </Typography>
            </CardContent>
          </Card>

          <Card sx={{ minWidth: 200 }}>
            <CardContent>
              <Typography color="textSecondary" gutterBottom variant="caption">
                Total Orders
              </Typography>
              <Typography variant="h5">
                {summary.total_orders.toLocaleString()}
              </Typography>
              <Typography variant="caption">
                Conv Rate: {formatPercentage(summary.overall_conversion_rate)}
              </Typography>
            </CardContent>
          </Card>

          <Card sx={{ minWidth: 200 }}>
            <CardContent>
              <Typography color="textSecondary" gutterBottom variant="caption">
                ROAS
              </Typography>
              <Typography variant="h5">
                {summary.overall_roas.toFixed(2)}x
              </Typography>
              <Typography variant="caption">
                {formatCurrency(summary.total_sales_from_ads)} revenue
              </Typography>
            </CardContent>
          </Card>
        </Box>
      )}

      {/* Campaign List */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Campaigns ({campaigns.length})
          </Typography>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Campaign Name</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Type</TableCell>
                  <TableCell>Bidding Method</TableCell>
                  <TableCell>Placement</TableCell>
                  <TableCell>Daily Budget</TableCell>
                  <TableCell>Target ROAS</TableCell>
                  <TableCell>Start Date</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {campaigns.map((campaign) => (
                  <TableRow key={campaign.campaign_id}>
                    <TableCell>
                      <Box>
                        <Typography variant="body2" fontWeight="bold">
                          {campaign.campaign_name}
                        </Typography>
                        <Typography variant="caption" color="textSecondary">
                          ID: {campaign.campaign_id}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={campaign.campaign_status}
                        color={getStatusColor(campaign.campaign_status) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>{campaign.campaign_type || "-"}</TableCell>
                    <TableCell>
                      {campaign.bidding_method ? (
                        <Chip 
                          label={campaign.bidding_method} 
                          color={campaign.bidding_method === "auto" ? "success" : "default"}
                          size="small"
                          variant="outlined"
                        />
                      ) : "-"}
                    </TableCell>
                    <TableCell>{campaign.placement_type || "-"}</TableCell>
                    <TableCell>
                      {campaign.daily_budget ? formatCurrency(campaign.daily_budget) : "-"}
                    </TableCell>
                    <TableCell>
                      {campaign.target_roas ? `${campaign.target_roas.toFixed(2)}x` : "-"}
                    </TableCell>
                    <TableCell>
                      {campaign.start_date ? new Date(campaign.start_date).toLocaleDateString() : "-"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
          
          {campaigns.length === 0 && !loading && (
            <Box sx={{ textAlign: "center", py: 4 }}>
              <Typography color="textSecondary">
                No campaigns found. Try fetching data from Shopee API.
              </Typography>
            </Box>
          )}
        </CardContent>
      </Card>

      {/* Historical Sync Dialog */}
      <Dialog open={syncDialogOpen} onClose={handleSyncDialogClose} maxWidth="sm" fullWidth>
        <DialogTitle>Sync Historical Ads Performance</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
            This will sync all historical ads performance data hourly, including detailed campaign settings 
            (bidding method, placement, product IDs, keywords, etc.). The process runs in background and may take several minutes.
            Shopee credentials are automatically retrieved from the selected store configuration.
          </Typography>
          
          <FormControl fullWidth margin="normal" error={!!storesError}>
            <InputLabel>Store</InputLabel>
            <Select
              value={syncForm.store_id}
              onChange={(e) => setSyncForm({ ...syncForm, store_id: e.target.value as string })}
              label="Store"
              disabled={storesLoading || stores.length === 0}
            >
              {storesLoading ? (
                <MenuItem disabled>Loading stores...</MenuItem>
              ) : storesError ? (
                <MenuItem disabled>{storesError}</MenuItem>
              ) : stores.length === 0 ? (
                <MenuItem disabled>No stores available</MenuItem>
              ) : (
                stores.map((store) => (
                  <MenuItem key={store.store_id} value={store.store_id.toString()}>
                    {store.nama_toko}
                  </MenuItem>
                ))
              )}
            </Select>
            {storesError && (
              <Typography variant="caption" color="error" sx={{ mt: 1 }}>
                {storesError}
                <Button 
                  size="small" 
                  onClick={fetchStores} 
                  sx={{ ml: 1 }}
                  disabled={storesLoading}
                >
                  Retry
                </Button>
              </Typography>
            )}
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleSyncDialogClose}>Cancel</Button>
          <Button 
            onClick={handleSyncSubmit} 
            variant="contained"
            disabled={syncing || !syncForm.store_id || stores.length === 0 || storesLoading}
          >
            {syncing ? "Starting..." : "Start Sync"}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}