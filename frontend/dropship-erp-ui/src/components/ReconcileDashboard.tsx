import { useEffect, useState } from "react";
import { Alert, Button } from "@mui/material";
import { useNavigate } from "react-router-dom";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listCandidates, reconcileCheck } from "../api/reconcile";
import { listAllStores } from "../api";
import type { ReconcileCandidate, Store } from "../types";
import usePagination from "../usePagination";

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<ReconcileCandidate[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const navigate = useNavigate();
  const { paginated, controls } = usePagination(data);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  const fetchData = () => {
    listCandidates(shop).then((r) => setData(r.data));
  };

  useEffect(() => {
    fetchData();
  }, [shop]);

  const handleReconcile = async (kode: string) => {
    try {
      const res = await reconcileCheck(kode);
      setMsg({ type: "success", text: res.data.message });
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.message || e.message });
    }
  };

  const handleReconcileAll = async () => {
    for (const row of data) {
      try {
        await reconcileCheck(row.kode_pesanan);
      } catch {
        // ignore individual errors and continue
      }
    }
    setMsg({ type: "success", text: "Completed" });
    fetchData();
  };

  const columns: Column<ReconcileCandidate>[] = [
    { label: "Kode Pesanan", key: "kode_pesanan" },
    { label: "Kode Invoice Channel", key: "kode_invoice_channel" },
    { label: "Status", key: "status_pesanan_terakhir" },
    { label: "No Pesanan Shopee", key: "no_pesanan" },
    {
      label: "Dropship",
      render: (_, row) => (
        <Button
          size="small"
          onClick={() => navigate(`/dropship?order=${row.kode_invoice_channel}`)}
        >
          View
        </Button>
      ),
    },
    {
      label: "Action",
      render: (_, row) => (
        <Button size="small" onClick={() => handleReconcile(row.kode_pesanan)}>
          Reconcile
        </Button>
      ),
    },
  ];

  return (
    <div>
      <h2>Reconcile Dashboard</h2>
      <select
        aria-label="Shop"
        value={shop}
        onChange={(e) => setShop(e.target.value)}
        style={{ marginRight: "0.5rem" }}
      >
        <option value="">Select Store</option>
        {stores.map((s) => (
          <option key={s.store_id} value={s.nama_toko}>
            {s.nama_toko}
          </option>
        ))}
      </select>
      <Button onClick={fetchData}>Refresh</Button>
      <Button onClick={handleReconcileAll} sx={{ ml: 1 }}>
        Reconcile All
      </Button>
      {msg && (
        <Alert severity={msg.type} sx={{ mt: 2 }}>
          {msg.text}
        </Alert>
      )}
      <SortableTable columns={columns} data={paginated} />
      {controls}
    </div>
  );
}
