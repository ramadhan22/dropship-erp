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
} from "@mui/material";
import { useEffect, useState } from "react";

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
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Fetch stores on component mount
  useEffect(() => {
    fetchData();
  }, []);

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
    try {
      const response = await fetch("/api/ads/campaigns?limit=100");
      if (!response.ok) throw new Error("Failed to fetch campaigns");
      const data = await response.json();
      setCampaigns(data.campaigns || []);
    } catch (error) {
      console.error("Error fetching campaigns:", error);
    }
  };

  const fetchSummary = async () => {
    try {
      const endDate = new Date();
      const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
      
      const params = new URLSearchParams();
      params.set("start_date", startDate.toISOString().split("T")[0]);
      params.set("end_date", endDate.toISOString().split("T")[0]);

      const response = await fetch(`/api/ads/summary?${params}`);
      if (!response.ok) throw new Error("Failed to fetch summary");
      const data = await response.json();
      setSummary(data);
    } catch (error) {
      console.error("Error fetching summary:", error);
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
    </div>
  );
}