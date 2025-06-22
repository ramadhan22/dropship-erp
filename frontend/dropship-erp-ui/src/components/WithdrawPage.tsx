import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  CircularProgress,
  MenuItem,
  Select,
  TextField,
} from "@mui/material";
import { listAllStores, withdrawShopeeBalance } from "../api";
import type { Store } from "../types";

export default function WithdrawPage() {
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [amount, setAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    listAllStores().then((r) => setStores(r));
  }, []);

  const handleSubmit = async () => {
    setLoading(true);
    setMsg(null);
    try {
      await withdrawShopeeBalance(store, Number(amount));
      setMsg({ type: "success", text: "Withdraw successful" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 8, maxWidth: 300 }}>
      <h2>Withdraw Shopee Balance</h2>
      <Select
        value={store}
        onChange={(e) => setStore(e.target.value as string)}
        displayEmpty
      >
        <MenuItem value="" disabled>
          Select Store
        </MenuItem>
        {stores.map((s) => (
          <MenuItem key={s.store_id} value={s.nama_toko}>
            {s.nama_toko}
          </MenuItem>
        ))}
      </Select>
      <TextField
        label="Amount"
        type="number"
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
      />
      <Button variant="contained" onClick={handleSubmit} disabled={loading || !store || !amount}>
        Withdraw
      </Button>
      {loading && <CircularProgress size={24} />}
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
    </div>
  );
}
