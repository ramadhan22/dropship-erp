// File: src/components/MetricsPage.tsx

import {
  Alert,
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { computeMetrics, fetchMetrics, listAllStores } from "../api";
import type { Store } from "../types";
import type { Metric } from "../types";
import usePagination from "../usePagination";

export default function MetricsPage() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [period, setPeriod] = useState(new Date().toISOString().slice(0, 7));
  const [metric, setMetric] = useState<Metric | null>(null);
  const { paginated, controls } = usePagination(
    metric ? Object.entries(metric) : [],
  );
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  const handleCompute = async () => {
    try {
      await computeMetrics(shop, period);
      setMsg({ type: "success", text: "Metrics computed!" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleFetch = async () => {
    try {
      const res = await fetchMetrics(shop, period);
      setMetric(res.data);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Metrics (P&L &amp; Cash)</h2>

      <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem" }}>
        <select
          aria-label="Shop"
          value={shop}
          onChange={(e) => setShop(e.target.value)}
        >
          <option value="">Select Store</option>
          {stores.map((s) => (
            <option key={s.store_id} value={s.nama_toko}>
              {s.nama_toko}
            </option>
          ))}
        </select>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="Period (YYYY-MM)"
            views={["year", "month"]}
            openTo="month"
            format="yyyy-MM"
            value={new Date(period)}
            onChange={(date) => {
              if (!date) return;
              setPeriod(date.toISOString().slice(0, 7));
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <Button variant="contained" onClick={handleCompute}>
          Compute
        </Button>
        <Button variant="outlined" onClick={handleFetch}>
          Fetch
        </Button>
      </div>

      {msg && (
        <Alert severity={msg.type} sx={{ marginBottom: "1rem" }}>
          {msg.text}
        </Alert>
      )}

      {metric && (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Metric</TableCell>
              <TableCell align="right">Value</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {paginated.map(([key, val]) => (
              <TableRow key={key}>
                <TableCell>{key}</TableCell>
                <TableCell align="right">
                  {typeof val === "number"
                    ? val.toLocaleString("id-ID", {
                        style: "currency",
                        currency: "IDR",
                      })
                    : (val as any)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {controls}
    </div>
  );
}
