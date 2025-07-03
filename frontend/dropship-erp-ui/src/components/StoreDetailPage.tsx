import { useEffect, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { Alert, TextField } from "@mui/material";
import { getStore, updateStore } from "../api";
import type { Store } from "../types";

export default function StoreDetailPage() {
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const [store, setStore] = useState<Store | null>(null);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    if (!id) return;
    (async () => {
      try {
        let st = (await getStore(Number(id))).data;
        const code = searchParams.get("code");
        const shopId = searchParams.get("shop_id");
        if (code && shopId) {
          await updateStore(st.store_id, {
            nama_toko: st.nama_toko,
            jenis_channel_id: st.jenis_channel_id,
            code_id: code,
            shop_id: shopId,
          });
          st = (await getStore(st.store_id)).data;
          setMsg({ type: "success", text: "Store authorized" });
        } else if (st.code_id && st.shop_id) {
          setMsg({ type: "success", text: "Store authorized" });
        }
        setStore(st);
      } catch (e: any) {
        setMsg({ type: "error", text: e.response?.data?.error || e.message });
      }
    })();
  }, [id, searchParams]);

  if (!store) return <div>Loading...</div>;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 8, maxWidth: 300 }}>
      <h2>Store Detail</h2>
      <TextField label="Store" value={store.nama_toko} disabled size="small" />
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
    </div>
  );
}
