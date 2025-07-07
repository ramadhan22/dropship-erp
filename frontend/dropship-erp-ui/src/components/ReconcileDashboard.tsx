import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useNavigate } from "react-router-dom";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import {
  listCandidates,
  reconcileCheck,
  cancelPurchase,
  updateShopeeStatus,
  fetchEscrowDetail,
  fetchShopeeDetail,
} from "../api/reconcile";
import { listAllStores } from "../api";
import type {
  ReconcileCandidate,
  Store,
  ShopeeOrderDetail,
  ShopeeEscrowDetail,
} from "../types";
import { getCurrentMonthRange } from "../utils/date";
import useServerPagination from "../useServerPagination";
import JsonTabs from "./JsonTabs";

function formatLabel(label: string): string {
  return label.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function formatValue(val: any): string {
  if (typeof val === "number" && val > 1e10) {
    return new Date(val * 1000).toLocaleString();
  }
  return String(val);
}

function renderValue(value: any): JSX.Element {
  if (Array.isArray(value)) {
    return <JsonTabs items={value} />;
  }
  if (typeof value === "object" && value !== null) {
    return (
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <tbody>
          {Object.entries(value).map(([k, v]) => (
            <tr key={k}>
              <td
                style={{
                  fontWeight: "bold",
                  verticalAlign: "top",
                  paddingRight: "0.5rem",
                }}
              >
                {formatLabel(k)}
              </td>
              <td>{renderValue(v)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }
  return <>{formatValue(value)}</>;
}

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [order, setOrder] = useState("");
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [stores, setStores] = useState<Store[]>([]);
  const { data, controls, reload } = useServerPagination((params) =>
    listCandidates(shop, order, from, to, params.page, params.pageSize).then(
      (r) => r.data,
    ),
  );
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const [detail, setDetail] = useState<
    ShopeeOrderDetail | ShopeeEscrowDetail | null
  >(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailInvoice, setDetailInvoice] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  useEffect(() => {
    reload();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [shop, order, from, to]);

  const handleReconcile = async (kode: string) => {
    try {
      const res = await reconcileCheck(kode);
      setMsg({ type: "success", text: res.data.message });
      reload();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.message || e.message });
    }
  };

  const handleCheckStatus = async (inv: string, status: string) => {
    try {
      const apiCall = status.toLowerCase() === "completed" ? fetchEscrowDetail : fetchShopeeDetail;
      const res = await apiCall(inv);
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
      reload();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleCancel = async (kode: string) => {
    try {
      await cancelPurchase(kode);
      setMsg({ type: "success", text: "Canceled" });
      reload();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleReconcileAll = async () => {
    await Promise.all(
      data.map(async (row) => {
        try {
          await updateShopeeStatus(row.kode_invoice_channel);
          await reconcileCheck(row.kode_pesanan);
        } catch {
          // ignore individual errors
        }
      }),
    );
    setMsg({ type: "success", text: "Completed" });
    reload();
  };

  const columns: Column<ReconcileCandidate>[] = [
    { label: "Kode Pesanan", key: "kode_pesanan" },
    { label: "Kode Invoice Channel", key: "kode_invoice_channel" },
    { label: "Status", key: "status_pesanan_terakhir" },
    { label: "No Pesanan Shopee", key: "no_pesanan" },
    {
      label: "Shopee Order Status",
      key: "shopee_order_status",
    },
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
          onClick={() => handleCheckStatus(row.kode_invoice_channel, row.shopee_order_status)}
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
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DatePicker
          label="From"
          format="yyyy-MM-dd"
          value={new Date(from)}
          onChange={(date) => {
            if (!date) return;
            setFrom(date.toISOString().split("T")[0]);
          }}
          slotProps={{ textField: { size: "small" } }}
        />
      </LocalizationProvider>
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DatePicker
          label="To"
          format="yyyy-MM-dd"
          value={new Date(to)}
          onChange={(date) => {
            if (!date) return;
            setTo(date.toISOString().split("T")[0]);
          }}
          slotProps={{ textField: { size: "small" } }}
        />
      </LocalizationProvider>
      <Button onClick={() => reload()}>Refresh</Button>
      <Button onClick={handleReconcileAll} sx={{ ml: 1 }}>
        Reconcile All
      </Button>
      {msg && (
        <Alert severity={msg.type} sx={{ mt: 2 }}>
          {msg.text}
        </Alert>
      )}
      <SortableTable columns={columns} data={data} />
      {controls}
      <Dialog
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        maxWidth="md"
        fullWidth
      >
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
                    <td>{renderValue(value)}</td>
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
