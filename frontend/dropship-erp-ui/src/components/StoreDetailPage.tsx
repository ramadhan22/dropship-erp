import { useEffect, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { Alert, Button, TextField } from "@mui/material";
import { getStore, updateStore } from "../api";
import type { Store } from "../types";

export default function StoreDetailPage() {
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const [store, setStore] = useState<Store | null>(null);
  const [code, setCode] = useState("");
  const [shopId, setShopId] = useState("");
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    if (!id) return;
    getStore(Number(id)).then((res) => {
      const st = res.data;
      setStore(st);
      setCode(searchParams.get("code") ?? st.code_id ?? "");
      setShopId(searchParams.get("shop_id") ?? st.shop_id ?? "");
    });
  }, [id, searchParams]);

  const handleSave = async () => {
    if (!store) return;
    try {
      await updateStore(store.store_id, {
        nama_toko: store.nama_toko,
        jenis_channel_id: store.jenis_channel_id,
        code_id: code,
        shop_id: shopId,
      });
      setMsg({ type: "success", text: "Store updated" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  if (!store) return <div>Loading...</div>;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 8, maxWidth: 300 }}>
      <h2>Store Detail</h2>
      <TextField label="Store" value={store.nama_toko} disabled size="small" />
      <TextField
        label="Code ID"
        value={code}
        onChange={(e) => setCode(e.target.value)}
        size="small"
      />
      <TextField
        label="Shop ID"
        value={shopId}
        onChange={(e) => setShopId(e.target.value)}
        size="small"
      />
      <Button variant="contained" onClick={handleSave}>
        Save
      </Button>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
    </div>
  );
}
