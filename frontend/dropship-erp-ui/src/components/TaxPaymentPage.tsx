import { Alert, Button, CircularProgress, MenuItem, Select } from "@mui/material";
import { useState } from "react";
import { fetchTaxPayment, payTax } from "../api/tax";
import { listAllStores } from "../api";
import { formatCurrency } from "../utils/format";
import type { Store, TaxPayment } from "../types";

export default function TaxPaymentPage() {
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [type, setType] = useState("monthly");
  const [period, setPeriod] = useState("");
  const [data, setData] = useState<TaxPayment | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useState(() => {
    listAllStores().then((s) => setStores(s));
  });

  const handleFetch = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetchTaxPayment(store, type, period);
      setData(res.data);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  const handlePay = async () => {
    if (!data) return;
    setLoading(true);
    setError(null);
    try {
      await payTax(data);
      setData({ ...data, is_paid: true });
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };


  return (
    <div>
      <h2>Tax Payment</h2>
      <div style={{ display: "flex", gap: "0.5rem" }}>
        <Select
          value={store}
          onChange={(e) => setStore(e.target.value)}
          displayEmpty
        >
          <MenuItem value="">Select Store</MenuItem>
          {stores.map((s) => (
            <MenuItem key={s.store_id} value={s.nama_toko}>
              {s.nama_toko}
            </MenuItem>
          ))}
        </Select>
        <Select value={type} onChange={(e) => setType(e.target.value)}>
          <MenuItem value="monthly">Monthly</MenuItem>
          <MenuItem value="yearly">Yearly</MenuItem>
        </Select>
        <input
          aria-label="Period"
          value={period}
          onChange={(e) => setPeriod(e.target.value)}
          placeholder={type === "monthly" ? "2025-06" : "2025"}
        />
        <Button variant="contained" onClick={handleFetch} disabled={loading}>
          Fetch
        </Button>
      </div>
      {loading && <CircularProgress size={24} />}
      {error && <Alert severity="error">{error}</Alert>}
      {data && (
        <div style={{ marginTop: "1rem" }}>
          <p>Revenue: {formatCurrency(data.revenue)}</p>
          <p>Tax Amount: {formatCurrency(data.tax_amount)}</p>
          {data.is_paid ? (
            <Alert severity="success">Paid on {data.paid_at}</Alert>
          ) : (
            <Button variant="contained" onClick={handlePay} disabled={loading}>
              Bayar Pajak
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
