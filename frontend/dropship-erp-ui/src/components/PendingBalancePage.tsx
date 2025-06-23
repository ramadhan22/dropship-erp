import { useEffect, useState } from "react";
import { Button, CircularProgress, Alert, FormControl, InputLabel, Select, MenuItem } from "@mui/material";
import { listAllStores, fetchPendingBalance } from "../api";
import type { Store } from "../types";

export default function PendingBalancePage() {
  const [store, setStore] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [balance, setBalance] = useState<number | null>(null);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  useEffect(() => {
    handleFetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [store]);

  async function handleFetch() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetchPendingBalance(store);
      setBalance(res.data.pending_balance);
    } catch (e: any) {
      setError(e.response?.data?.error || e.message);
      setBalance(null);
    } finally {
      setLoading(false);
    }
  }

  const money = (v: number) =>
    new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR" }).format(v);

  return (
    <div>
      <h2>Pending Balance</h2>
      <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
        <FormControl size="small" sx={{ minWidth: 160 }}>
          <InputLabel id="store-label">Store</InputLabel>
          <Select
            labelId="store-label"
            value={store}
            label="Store"
            onChange={(e) => setStore(e.target.value)}
          >
            <MenuItem value="">All Stores</MenuItem>
            {stores.map((s) => (
              <MenuItem key={s.store_id} value={s.nama_toko}>
                {s.nama_toko}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <Button variant="contained" onClick={handleFetch} disabled={loading}>
          Fetch
        </Button>
        {loading && <CircularProgress size={24} />}
      </div>
      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error}
        </Alert>
      )}
      {balance !== null && !loading && (
        <div style={{ marginTop: "1rem", fontSize: "1.2rem" }}>{money(balance)}</div>
      )}
    </div>
  );
}
