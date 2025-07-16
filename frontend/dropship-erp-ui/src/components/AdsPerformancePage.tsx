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
} from "@mui/material";
import { useEffect, useState } from "react";
import { 
  fetchAdsCampaigns, 
  fetchAdsPerformanceSummary, 
  syncHistoricalAdsPerformance 
} from "../api/adsPerformance";

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
  const [campaigns, setCampaigns] = useState<AdsCampaign[]>([]);
  const [summary, setSummary] = useState<AdsPerformanceSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [syncDialogOpen, setSyncDialogOpen] = useState(false);
  const [syncForm, setSyncForm] = useState({
    store_id: "",
  });
  const [stores, setStores] = useState<{ store_id: number; nama_toko: string }[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Fetch stores and initial data on component mount
  useEffect(() => {
    fetchStores();
    fetchData();
  }, []);

  const fetchStores = async () => {
    try {
      const response = await fetch("/api/stores/all");
      if (!response.ok) throw new Error("Failed to fetch stores");
      const data = await response.json();
      setStores(data || []);
    } catch (error) {
      console.error("Error fetching stores:", error);
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      await Promise.all([fetchCampaignsData(), fetchSummaryData()]);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setLoading(false);
    }
  };

  const fetchCampaignsData = async () => {
    try {
      const data = await fetchAdsCampaigns({ limit: 100 });
      setCampaigns(data.campaigns || []);
    } catch (error) {
      console.error("Error fetching campaigns:", error);
    }
  };

  const fetchSummaryData = async () => {
    try {
      const endDate = new Date();
      const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
      
      const data = await fetchAdsPerformanceSummary({
        start_date: startDate.toISOString().split("T")[0],
        end_date: endDate.toISOString().split("T")[0],
      });
      setSummary(data);
    } catch (error) {
      console.error("Error fetching summary:", error);
    }
  };

  const handleSyncDialogOpen = () => {
    setSyncDialogOpen(true);
  };

  const handleSyncDialogClose = () => {
    setSyncDialogOpen(false);
    setSyncForm({ store_id: "" });
  };

  const handleSyncSubmit = async () => {
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

  return (
    <div>
      <Typography variant="h4" gutterBottom>
        Ads Performance Dashboard
      </Typography>

      <Box sx={{ mb: 3 }}>
        <Button
          variant="contained"
          onClick={fetchData}
          disabled={loading}
          sx={{ mr: 2 }}
        >
          {loading ? "Loading..." : "Refresh Data"}
        </Button>
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
            This will sync all historical ads performance data hourly. The process runs in background and may take several minutes.
            Shopee credentials are automatically retrieved from the selected store configuration.
          </Typography>
          
          <FormControl fullWidth margin="normal">
            <InputLabel>Store</InputLabel>
            <Select
              value={syncForm.store_id}
              onChange={(e) => setSyncForm({ ...syncForm, store_id: e.target.value as string })}
              label="Store"
            >
              {stores.map((store) => (
                <MenuItem key={store.store_id} value={store.store_id.toString()}>
                  {store.nama_toko}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleSyncDialogClose}>Cancel</Button>
          <Button 
            onClick={handleSyncSubmit} 
            variant="contained"
            disabled={syncing || !syncForm.store_id}
          >
            {syncing ? "Starting..." : "Start Sync"}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}