import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  LinearProgress,
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
  updateShopeeStatuses,
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
import { formatCurrency, formatDateTime } from "../utils/format";

function formatLabel(label: string): string {
  return label.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function formatValue(key: string, val: any): string {
  if (val == null) return "";
  const k = key.toLowerCase();
  if (typeof val === "string" && /^\d{4}-\d{2}-\d{2}T/.test(val)) {
    return formatDateTime(val);
  }
  if (typeof val === "number") {
    if (k.includes("time") || k.includes("date")) {
      return formatDateTime(val * 1000);
    }
    if (
      k.includes("amount") ||
      k.includes("fee") ||
      k.includes("price") ||
      k.includes("total") ||
      k.includes("tax") ||
      k.includes("voucher") ||
      k.includes("discount") ||
      k.includes("cost")
    ) {
      return formatCurrency(val);
    }
  }
  if (
    typeof val === "string" &&
    /^\d+$/.test(val) &&
    (k.includes("time") || k.includes("date"))
  ) {
    return formatDateTime(Number(val) * 1000);
  }
  if (
    typeof val === "string" &&
    /^\d+(\.\d+)?$/.test(val) &&
    (k.includes("amount") ||
      k.includes("fee") ||
      k.includes("price") ||
      k.includes("total") ||
      k.includes("tax") ||
      k.includes("voucher") ||
      k.includes("discount") ||
      k.includes("cost"))
  ) {
    return formatCurrency(Number(val));
  }
  return String(val);
}

function renderValue(key: string, value: any): JSX.Element {
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
              <td>{renderValue(k, v)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }
  return <>{formatValue(key, value)}</>;
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
  const [progress, setProgress] = useState<{ done: number; total: number } | null>(
    null,
  );
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
    try {
      const all: ReconcileCandidate[] = [];
      const pageSize = 1000;
      let page = 1;
      const skipLoading = { headers: { "X-Skip-Loading": "1" } };
      while (true) {
        const res = await listCandidates(
          shop,
          order,
          from,
          to,
          page,
          pageSize,
          skipLoading,
        );
        all.push(...res.data.data);
        if (all.length >= res.data.total) break;
        page += 1;
      }

      for (let i = 0; i < all.length; i += 50) {
        try {
          await updateShopeeStatuses(
            all.slice(i, i + 50).map((r) => r.kode_invoice_channel),
            skipLoading,
          );
        } catch {
          // ignore batch errors
        }
      }

      setProgress({ done: 0, total: all.length });
      const concurrency =
        (typeof navigator !== "undefined" && navigator.hardwareConcurrency) || 4;
      let idx = 0;
      const workers = Array.from({ length: concurrency }, async () => {
        for (;;) {
          const cur = idx++;
          if (cur >= all.length) break;
          try {
            await reconcileCheck(all[cur].kode_pesanan, skipLoading);
          } catch {
            // ignore individual errors
          }
          setProgress((p) =>
            p ? { ...p, done: Math.min(p.done + 1, p.total) } : p,
          );
        }
      });
      await Promise.all(workers);

      setProgress(null);

      setMsg({ type: "success", text: "Completed" });
      reload();
    } catch (e: any) {
      setMsg({ type: "error", text: e.message });
    }
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
      {progress && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "0.5rem",
            marginTop: "0.5rem",
          }}
        >
          <LinearProgress
            variant="determinate"
            value={(progress.done / progress.total) * 100}
            sx={{ flexGrow: 1 }}
          />
          <span>
            Reconciling {progress.done}/{progress.total} (
            {Math.round((progress.done / progress.total) * 100)}%)
          </span>
        </div>
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
                    <td>{renderValue(key, value)}</td>
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
