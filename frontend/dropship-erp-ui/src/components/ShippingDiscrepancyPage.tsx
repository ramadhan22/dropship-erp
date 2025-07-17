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
  TextField,
  Typography,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import {
  listShippingDiscrepancies,
  getShippingDiscrepancyStats,
} from "../api/shippingDiscrepancies";
import type { ShippingDiscrepancy } from "../types";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { getCurrentMonthRange } from "../utils/date";
import usePagination from "../usePagination";

export default function ShippingDiscrepancyPage() {
  const [list, setList] = useState<ShippingDiscrepancy[]>([]);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [storeName, setStoreName] = useState("");
  const [discrepancyType, setDiscrepancyType] = useState("");
  const [stats, setStats] = useState<Record<string, number>>({});
  const { paginated, controls, setPage } = usePagination(list);

  const fetchData = async () => {
    try {
      const res = await listShippingDiscrepancies({
        store_name: storeName || undefined,
        type: discrepancyType || undefined,
        page_size: 100,
      });
      setList(res.data.data);
      setMsg(null);
      setPage(1);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const fetchStats = async () => {
    try {
      const res = await getShippingDiscrepancyStats({
        start_date: from?.toISOString().split("T")[0],
        end_date: to?.toISOString().split("T")[0],
      });
      setStats(res.data.stats);
    } catch (e: any) {
      console.error("Failed to fetch stats:", e);
    }
  };

  useEffect(() => {
    fetchData();
  }, [storeName, discrepancyType]);

  useEffect(() => {
    fetchStats();
  }, [from, to]);

  const columns: Column<ShippingDiscrepancy>[] = [
    {
      id: "invoice_number",
      label: "Invoice Number",
      align: "left",
      render: (row) => row.invoice_number,
    },
    {
      id: "order_id",
      label: "Order ID",
      align: "left",
      render: (row) => row.order_id || "-",
    },
    {
      id: "discrepancy_type",
      label: "Type",
      align: "left",
      render: (row) => (
        <span
          style={{
            color: row.discrepancy_type === "selisih_ongkir" ? "#1976d2" : "#ed6c02",
          }}
        >
          {row.discrepancy_type === "selisih_ongkir" ? "Selisih Ongkir" : "Reverse Shipping Fee"}
        </span>
      ),
    },
    {
      id: "discrepancy_amount",
      label: "Amount",
      align: "right",
      render: (row) => (
        <span
          style={{
            color: row.discrepancy_amount >= 0 ? "#2e7d32" : "#d32f2f",
          }}
        >
          Rp {row.discrepancy_amount.toLocaleString("id-ID")}
        </span>
      ),
    },
    {
      id: "actual_shipping_fee",
      label: "Actual Shipping",
      align: "right",
      render: (row) =>
        row.actual_shipping_fee ? `Rp ${row.actual_shipping_fee.toLocaleString("id-ID")}` : "-",
    },
    {
      id: "buyer_paid_shipping_fee",
      label: "Buyer Paid",
      align: "right",
      render: (row) =>
        row.buyer_paid_shipping_fee ? `Rp ${row.buyer_paid_shipping_fee.toLocaleString("id-ID")}` : "-",
    },
    {
      id: "store_name",
      label: "Store",
      align: "left",
      render: (row) => row.store_name || "-",
    },
    {
      id: "order_date",
      label: "Order Date",
      align: "left",
      render: (row) =>
        row.order_date ? new Date(row.order_date).toLocaleDateString("id-ID") : "-",
    },
    {
      id: "created_at",
      label: "Recorded At",
      align: "left",
      render: (row) => new Date(row.created_at).toLocaleDateString("id-ID"),
    },
  ];

  const totalSelisihOngkir = stats.selisih_ongkir || 0;
  const totalReverseShipping = stats.reverse_shipping_fee || 0;

  return (
    <div style={{ padding: 20 }}>
      <Typography variant="h4" gutterBottom>
        Shipping Discrepancies
      </Typography>

      {msg && (
        <Alert severity={msg.type} style={{ marginBottom: 16 }}>
          {msg.text}
        </Alert>
      )}

      {/* Stats Cards */}
      <Grid container spacing={2} style={{ marginBottom: 20 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" color="primary">
                Selisih Ongkir
              </Typography>
              <Typography variant="h4">
                {totalSelisihOngkir}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Transactions with shipping cost differences
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" color="secondary">
                Reverse Shipping Fee
              </Typography>
              <Typography variant="h4">
                {totalReverseShipping}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Orders with reverse shipping fees
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Filters */}
      <div style={{ marginBottom: 20 }}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={3}>
            <TextField
              fullWidth
              label="Store Name"
              value={storeName}
              onChange={(e) => setStoreName(e.target.value)}
              variant="outlined"
            />
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth variant="outlined">
              <InputLabel>Discrepancy Type</InputLabel>
              <Select
                value={discrepancyType}
                onChange={(e) => setDiscrepancyType(e.target.value)}
                label="Discrepancy Type"
              >
                <MenuItem value="">All Types</MenuItem>
                <MenuItem value="selisih_ongkir">Selisih Ongkir</MenuItem>
                <MenuItem value="reverse_shipping_fee">Reverse Shipping Fee</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns}>
              <DatePicker
                label="Stats From"
                value={from}
                onChange={setFrom}
                slotProps={{
                  textField: { fullWidth: true, variant: "outlined" },
                }}
              />
            </LocalizationProvider>
          </Grid>
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns}>
              <DatePicker
                label="Stats To"
                value={to}
                onChange={setTo}
                slotProps={{
                  textField: { fullWidth: true, variant: "outlined" },
                }}
              />
            </LocalizationProvider>
          </Grid>
        </Grid>
      </div>

      <div style={{ marginBottom: 20 }}>
        <Button variant="contained" onClick={fetchData} style={{ marginRight: 10 }}>
          Refresh Data
        </Button>
        <Button variant="outlined" onClick={fetchStats}>
          Refresh Stats
        </Button>
      </div>

      {list.length > 0 ? (
        <SortableTable
          data={paginated}
          columns={columns}
          pagination={controls}
          emptyMessage="No shipping discrepancies found."
        />
      ) : (
        <Alert severity="info">No shipping discrepancies found for the current filters.</Alert>
      )}
    </div>
  );
}