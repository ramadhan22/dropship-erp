// File: src/components/MetricsPage.tsx

import {
  Alert,
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
} from "@mui/material";
import { useState } from "react";
import { computeMetrics, fetchMetrics } from "../api";
import type { Metric } from "../types";

export default function MetricsPage() {
  const [shop, setShop] = useState("");
  const [period, setPeriod] = useState("");
  const [metric, setMetric] = useState<Metric | null>(null);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

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
        <TextField
          label="Shop"
          value={shop}
          onChange={(e) => setShop(e.target.value)}
        />
        <TextField
          label="Period (YYYY-MM)"
          value={period}
          onChange={(e) => setPeriod(e.target.value)}
        />
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
              <TableCell>Value</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {Object.entries(metric).map(([key, val]) => (
              <TableRow key={key}>
                <TableCell>{key}</TableCell>
                <TableCell>{val}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}
