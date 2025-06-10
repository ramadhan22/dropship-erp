import { Alert, Button, TextField } from "@mui/material";
import { useState } from "react";
import { reconcile } from "../api";

export default function ReconcileForm() {
  const [shop, setShop] = useState("");
  const [purchaseId, setPurchaseId] = useState("");
  const [orderId, setOrderId] = useState("");
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

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
      <TextField
        label="Shop"
        fullWidth
        value={shop}
        onChange={(e) => setShop(e.target.value)}
        sx={{ mb: 2 }}
      />
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
