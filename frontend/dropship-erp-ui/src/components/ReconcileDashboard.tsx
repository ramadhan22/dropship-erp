import { useEffect, useState, useMemo, useCallback } from "react";
import type { JSX } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  LinearProgress,
  CircularProgress,
  Box,
  Pagination,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useNavigate } from "react-router-dom";
import VirtualizedTable from "./VirtualizedTable";
import type { Column } from "./SortableTable";
import {
  reconcileCheck,
  cancelPurchase,
  updateShopeeStatus,
  createReconcileBatch,
  fetchEscrowDetail,
  fetchShopeeDetail,
  checkJobStatus,
} from "../api/reconcile";
import { listAllStores } from "../api";
import type {
  ReconcileCandidate,
  Store,
  ShopeeOrderDetail,
  ShopeeEscrowDetail,
} from "../types";
import { getCurrentMonthRange } from "../utils/date";
import { useReconcileCandidates, useReconcileMutations } from "../hooks/useReconcileData";
import { useDebouncedInput } from "../hooks/useDebounce";
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
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [stores, setStores] = useState<Store[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  
  // Use debounced inputs for better performance
  const shopInput = useDebouncedInput('');
  const orderInput = useDebouncedInput('');
  const statusInput = useDebouncedInput('');

  // Memoize filters to prevent unnecessary re-renders
  const filters = useMemo(() => ({
    shop: shopInput.debouncedValue,
    order: orderInput.debouncedValue,
    status: statusInput.debouncedValue,
    from,
    to,
    page,
    pageSize,
  }), [shopInput.debouncedValue, orderInput.debouncedValue, statusInput.debouncedValue, from, to, page, pageSize]);

  // Use optimized React Query hook
  const { data, error, isLoading, refetch } = useReconcileCandidates(filters);
  const { optimisticUpdateCandidate } = useReconcileMutations();

  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const [progress] = useState<{ done: number; total: number } | null>(
    null,
  );
  const [detail, setDetail] = useState<
    ShopeeOrderDetail | ShopeeEscrowDetail | null
  >(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailInvoice, setDetailInvoice] = useState("");
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [pendingJobs, setPendingJobs] = useState<Map<string, number>>(new Map());
  const navigate = useNavigate();

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  // Clear message after 5 seconds
  useEffect(() => {
    if (msg) {
      const timer = setTimeout(() => setMsg(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [msg]);

  const handleReconcile = useCallback(async (kode: string) => {
    if (actionLoading) return;
    
    setActionLoading(kode);
    try {
      const res = await reconcileCheck(kode);
      setMsg({ type: "success", text: res.data.message });
      
      // Optimistically update the candidate status
      const updatedCandidate = data?.data.find(c => c.kode_pesanan === kode);
      if (updatedCandidate) {
        optimisticUpdateCandidate({
          ...updatedCandidate,
          status_pesanan_terakhir: 'Pesanan selesai'
        });
      }
      
      // Refresh data
      refetch();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.message || e.message });
    } finally {
      setActionLoading(null);
    }
  }, [actionLoading, data?.data, optimisticUpdateCandidate, refetch]);

  const handleCheckStatus = useCallback(async (inv: string, status: string) => {
    if (actionLoading) return;
    
    setActionLoading(inv);
    try {
      const apiCall = status.toLowerCase() === "completed" ? fetchEscrowDetail : fetchShopeeDetail;
      const res = await apiCall(inv);
      
      // Check if response indicates background processing
      if (res.data && typeof res.data === 'object' && 'status' in res.data && res.data.status === 'processing') {
        const batchData = res.data as { status: string; batch_id: number; message: string };
        setPendingJobs(prev => new Map(prev.set(inv, batchData.batch_id)));
        setMsg({ type: "info", text: batchData.message });
        
        // Start polling for job completion
        pollForJobCompletion(inv, batchData.batch_id);
      } else {
        // We have immediate data
        setDetail(res.data as ShopeeOrderDetail | ShopeeEscrowDetail);
        setDetailInvoice(inv);
        setDetailOpen(true);
      }
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setActionLoading(null);
    }
  }, [actionLoading]);

  const pollForJobCompletion = useCallback(async (invoice: string, batchId: number) => {
    const maxAttempts = 60; // Poll for up to 1 minute
    let attempts = 0;
    
    const poll = async () => {
      if (attempts >= maxAttempts) {
        setPendingJobs(prev => {
          const newMap = new Map(prev);
          newMap.delete(invoice);
          return newMap;
        });
        setMsg({ type: "error", text: "Job timed out. Please try again later." });
        return;
      }
      
      try {
        const statusRes = await checkJobStatus(batchId);
        const jobStatus = statusRes.data.status;
        
        if (jobStatus === 'completed') {
          // Job completed, fetch the data again
          setPendingJobs(prev => {
            const newMap = new Map(prev);
            newMap.delete(invoice);
            return newMap;
          });
          
          // Fetch the completed data
          const apiCall = fetchShopeeDetail; // Use the correct API call
          const res = await apiCall(invoice);
          setDetail(res.data as ShopeeOrderDetail);
          setDetailInvoice(invoice);
          setDetailOpen(true);
          setMsg({ type: "success", text: "Order detail loaded successfully!" });
        } else if (jobStatus === 'failed') {
          setPendingJobs(prev => {
            const newMap = new Map(prev);
            newMap.delete(invoice);
            return newMap;
          });
          setMsg({ type: "error", text: "Failed to fetch order detail. Please try again." });
        } else {
          // Still processing, continue polling
          attempts++;
          setTimeout(poll, 1000); // Poll every second
        }
      } catch (error) {
        // Error checking status, stop polling
        setPendingJobs(prev => {
          const newMap = new Map(prev);
          newMap.delete(invoice);
          return newMap;
        });
        setMsg({ type: "error", text: "Error checking job status." });
      }
    };
    
    // Start polling after a short delay
    setTimeout(poll, 1000);
  }, []);

  const handleUpdateStatus = useCallback(async () => {
    if (actionLoading) return;
    
    setActionLoading(detailInvoice);
    try {
      await updateShopeeStatus(detailInvoice);
      setMsg({ type: "success", text: "Updated" });
      refetch();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setActionLoading(null);
    }
  }, [actionLoading, detailInvoice, refetch]);

  const handleCancel = useCallback(async (kode: string) => {
    if (actionLoading) return;
    
    setActionLoading(kode);
    try {
      await cancelPurchase(kode);
      setMsg({ type: "success", text: "Canceled" });
      refetch();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setActionLoading(null);
    }
  }, [actionLoading, refetch]);

  const handleReconcileAll = useCallback(async () => {
    if (actionLoading) return;
    
    setActionLoading('reconcile-all');
    try {
      const response = await createReconcileBatch(
        shopInput.debouncedValue, 
        orderInput.debouncedValue, 
        statusInput.debouncedValue, 
        from, 
        to
      );
      const responseData = response.data;
      let message = "Reconcile batches created successfully";
      if (responseData.batches_created && responseData.total_transactions) {
        message = `Created ${responseData.batches_created} batches for ${responseData.total_transactions} transactions. Processing will begin shortly.`;
      } else if (responseData.message) {
        message = responseData.message;
      }
      setMsg({ type: "success", text: message });
      refetch();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setActionLoading(null);
    }
  }, [actionLoading, shopInput.debouncedValue, orderInput.debouncedValue, statusInput.debouncedValue, from, to, refetch]);

  const columns: Column<ReconcileCandidate>[] = useMemo(() => [
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
        <Button 
          size="small" 
          onClick={() => handleReconcile(row.kode_pesanan)}
          disabled={actionLoading === row.kode_pesanan}
        >
          {actionLoading === row.kode_pesanan ? (
            <CircularProgress size={16} />
          ) : (
            'Reconcile'
          )}
        </Button>
      ),
    },
    {
      label: "Cancel",
      render: (_, row) => (
        <Button 
          size="small" 
          onClick={() => handleCancel(row.kode_pesanan)}
          disabled={actionLoading === row.kode_pesanan}
        >
          {actionLoading === row.kode_pesanan ? (
            <CircularProgress size={16} />
          ) : (
            'Cancel'
          )}
        </Button>
      ),
    },
    {
      label: "Check Status",
      render: (_, row) => {
        const isPending = pendingJobs.has(row.kode_invoice_channel);
        const isLoading = actionLoading === row.kode_invoice_channel;
        
        return (
          <Button
            size="small"
            onClick={() => handleCheckStatus(row.kode_invoice_channel, row.shopee_order_status)}
            disabled={isLoading || isPending}
            variant={isPending ? "outlined" : "text"}
            color={isPending ? "warning" : "primary"}
          >
            {isLoading ? (
              <CircularProgress size={16} />
            ) : isPending ? (
              <>
                <CircularProgress size={16} sx={{ mr: 1 }} />
                Processing...
              </>
            ) : (
              'Check Status'
            )}
          </Button>
        );
      },
    },
  ], [navigate, handleReconcile, handleCancel, handleCheckStatus, actionLoading, pendingJobs]);

  // Pagination controls
  const totalPages = Math.max(1, Math.ceil((data?.total || 0) / pageSize));
  const paginationControls = useMemo(() => (
    <div
      style={{
        marginTop: "1rem",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <div>
        Total: {data?.total || 0} 
        {isLoading && ' (loading...)'}
        {shopInput.value !== shopInput.debouncedValue && ' (filtering...)'}
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
        <select
          value={pageSize}
          onChange={(e) => {
            setPageSize(Number(e.target.value));
            setPage(1);
          }}
          disabled={isLoading}
        >
          {[10, 20, 50, 100].map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
        <Pagination
          page={page}
          count={totalPages}
          onChange={(_, val) => setPage(val)}
          disabled={isLoading}
          showFirstButton
          showLastButton
        />
      </div>
    </div>
  ), [data?.total, isLoading, shopInput.value, shopInput.debouncedValue, pageSize, page, totalPages]);

  return (
    <div>
      <h2>Reconcile Dashboard</h2>
      
      {/* Filters Section */}
      <Box mb={2} display="flex" gap={1} flexWrap="wrap" alignItems="center">
        <select
          aria-label="Shop"
          value={shopInput.value}
          onChange={(e) => shopInput.setValue(e.target.value)}
          style={{ marginRight: "0.5rem" }}
          disabled={isLoading}
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
          value={orderInput.value}
          onChange={(e) => orderInput.setValue(e.target.value)}
          style={{ height: "2rem", marginRight: "0.5rem" }}
          disabled={isLoading}
        />
        <input
          aria-label="Status"
          placeholder="Status"
          value={statusInput.value}
          onChange={(e) => statusInput.setValue(e.target.value)}
          style={{ height: "2rem", marginRight: "0.5rem" }}
          disabled={isLoading}
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
            disabled={isLoading}
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
            disabled={isLoading}
          />
        </LocalizationProvider>
        <Button 
          onClick={() => refetch()} 
          disabled={isLoading}
          variant="outlined"
        >
          {isLoading ? <CircularProgress size={20} /> : 'Refresh'}
        </Button>
        <Button 
          onClick={handleReconcileAll} 
          sx={{ ml: 1 }}
          disabled={actionLoading === 'reconcile-all' || isLoading}
          variant="contained"
        >
          {actionLoading === 'reconcile-all' ? (
            <CircularProgress size={20} />
          ) : (
            'Reconcile All'
          )}
        </Button>
      </Box>

      {/* Error and Success Messages */}
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}

      {/* Progress Dialog */}
      <Dialog open={progress !== null} fullWidth maxWidth="sm">
        <DialogTitle>Reconciling</DialogTitle>
        <DialogContent
          sx={{ display: "flex", alignItems: "center", gap: "0.5rem" }}
        >
          <LinearProgress
            variant="determinate"
            value={progress ? (progress.done / progress.total) * 100 : 0}
            sx={{ flexGrow: 1 }}
          />
          {progress && (
            <span>
              {progress.done}/{progress.total} (
              {Math.round((progress.done / progress.total) * 100)}%)
            </span>
          )}
        </DialogContent>
      </Dialog>

      {/* Data Table - Now virtualized for better performance */}
      <VirtualizedTable 
        columns={columns} 
        data={data?.data || []} 
        loading={isLoading}
        height={500}
        emptyMessage="No reconcile candidates found"
      />
      
      {/* Pagination Controls */}
      {paginationControls}

      {/* Detail Modal */}
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
          <Button 
            onClick={handleUpdateStatus}
            disabled={actionLoading === detailInvoice}
          >
            {actionLoading === detailInvoice ? (
              <CircularProgress size={20} />
            ) : (
              'Update Status Shopee'
            )}
          </Button>
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
