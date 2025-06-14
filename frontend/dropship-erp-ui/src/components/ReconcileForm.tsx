import { Alert, Button, TextField } from "@mui/material";
import { useEffect, useState } from "react";
import { reconcile, listAllStores } from "../api";
import type { Store } from "../types";

export default function ReconcileForm() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [purchaseId, setPurchaseId] = useState("");
  const [orderId, setOrderId] = useState("");
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  const handleSubmit = async () => {
    try {
      await reconcile(purchaseId, orderId, shop);
      setMsg({ type: "success", text: "Reconciliation successful!" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Reconcile Purchase & Order</h2>
      <select
        aria-label="Shop"
        value={shop}
        onChange={(e) => setShop(e.target.value)}
        style={{ display: "block", width: "100%", marginBottom: "0.5rem" }}
      >
        <option value="">Select Store</option>
        {stores.map((s) => (
          <option key={s.store_id} value={s.nama_toko}>
            {s.nama_toko}
          </option>
        ))}
      </select>
      <TextField
        label="Purchase ID"
        fullWidth
        value={purchaseId}
        onChange={(e) => setPurchaseId(e.target.value)}
        sx={{ mb: 2 }}
      />
      <TextField
        label="Order ID"
        fullWidth
        value={orderId}
        onChange={(e) => setOrderId(e.target.value)}
        sx={{ mb: 2 }}
      />
      <Button variant="contained" onClick={handleSubmit}>
        Reconcile
      </Button>
      {msg && (
        <Alert severity={msg.type} sx={{ mt: 2 }}>
          {msg.text}
        </Alert>
      )}
    </div>
  );
}
