import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import {
  listCandidates,
  reconcileCheck,
  fetchShopeeDetail,
  cancelPurchase,
  updateShopeeStatus,
} from "../api/reconcile";
import { listAllStores } from "../api";
import type { ReconcileCandidate, Store, ShopeeOrderDetail } from "../types";
import usePagination from "../usePagination";
import JsonTabs from "./JsonTabs";

function formatLabel(label: string): string {
  return label
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [order, setOrder] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<ReconcileCandidate[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const [detail, setDetail] = useState<ShopeeOrderDetail | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailInvoice, setDetailInvoice] = useState("");
  const navigate = useNavigate();
  const { paginated, controls } = usePagination(data);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  const fetchData = () => {
    listCandidates(shop, order).then((r) => setData(r.data));
  };

  useEffect(() => {
    fetchData();
  }, [shop, order]);

  const handleReconcile = async (kode: string) => {
    try {
      const res = await reconcileCheck(kode);
      setMsg({ type: "success", text: res.data.message });
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.message || e.message });
    }
  };

  const handleCheckStatus = async (inv: string) => {
    try {
      const res = await fetchShopeeDetail(inv);
      setDetail(res.data);
      setDetailInvoice(inv);
      setDetailOpen(true);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleUpdateStatus = async () => {
    try {
      await updateShopeeStatus(detailInvoice);
      setMsg({ type: "success", text: "Updated" });
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleCancel = async (kode: string) => {
    try {
      await cancelPurchase(kode);
      setMsg({ type: "success", text: "Canceled" });
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
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
          onClick={() =>
            navigate(`/dropship?order=${row.kode_invoice_channel}`)
          }
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
    {
      label: "Cancel",
      render: (_, row) => (
        <Button size="small" onClick={() => handleCancel(row.kode_pesanan)}>
          Cancel
        </Button>
      ),
    },
    {
      label: "Check Status",
      render: (_, row) => (
        <Button
          size="small"
          onClick={() => handleCheckStatus(row.kode_invoice_channel)}
        >
          Check Status
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
      <input
        aria-label="Search Invoice"
        placeholder="Kode Invoice"
        value={order}
        onChange={(e) => setOrder(e.target.value)}
        style={{ height: "2rem", marginRight: "0.5rem" }}
      />
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
      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)}>
        <DialogTitle>Order Detail</DialogTitle>
        <DialogContent>
          {detail && (
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <tbody>
                {Object.entries(detail).map(([key, value]) => (
                  <tr key={key}>
                    <td
                      style={{
                        fontWeight: "bold",
                        verticalAlign: "top",
                        paddingRight: "0.5rem",
                      }}
                    >
                      {formatLabel(key)}
                    </td>
                    <td>
                      {Array.isArray(value) ? (
                        <JsonTabs items={value} />
                      ) : typeof value === "object" && value !== null ? (
                        <pre style={{ margin: 0, whiteSpace: "pre-wrap" }}>
                          {JSON.stringify(value, null, 2)}
                        </pre>
                      ) : (
                        String(value)
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleUpdateStatus}>Update Status Shopee</Button>
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
