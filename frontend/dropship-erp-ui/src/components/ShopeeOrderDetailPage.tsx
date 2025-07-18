import { useEffect, useState, useMemo, useCallback } from "react";
import type { JSX } from "react";
import {
  Button,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
  Box,
  Pagination,
} from "@mui/material";
import VirtualizedTable from "./VirtualizedTable";
import type { Column } from "./SortableTable";
import JsonTabs from "./JsonTabs";
import { formatCurrency, formatDateTime } from "../utils/format";
import {
  getOrderDetail,
  listAllStores,
} from "../api";
import type {
  ShopeeOrderDetailRow,
  ShopeeOrderItemRow,
  ShopeeOrderPackageRow,
  Store,
} from "../types";
import { useShopeeOrderDetails } from "../hooks/useReconcileData";
import { useDebouncedInput } from "../hooks/useDebounce";

function formatLabel(label: string): string {
  return label.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function formatValue(val: any): string {
  if (typeof val === "string" && /^\d{4}-\d{2}-\d{2}T/.test(val)) {
    return formatDateTime(val);
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


export default function ShopeeOrderDetailPage() {
  const [stores, setStores] = useState<Store[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  
  // Use debounced inputs for better performance
  const storeInput = useDebouncedInput('');
  const orderInput = useDebouncedInput('');

  // Memoize filters to prevent unnecessary re-renders
  const filters = useMemo(() => ({
    store: storeInput.debouncedValue,
    order: orderInput.debouncedValue,
    page,
    pageSize,
  }), [storeInput.debouncedValue, orderInput.debouncedValue, page, pageSize]);

  // Use optimized React Query hook
  const { data, error, isLoading, refetch } = useShopeeOrderDetails(filters);

  const [detail, setDetail] = useState<{
    detail: ShopeeOrderDetailRow;
    items: ShopeeOrderItemRow[];
    packages: ShopeeOrderPackageRow[];
  } | null>(null);
  const [open, setOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState<string | null>(null);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

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

  const openDetail = useCallback(async (sn: string) => {
    if (detailLoading) return;
    
    setDetail(null);
    setOpen(true);
    setDetailLoading(sn);
    
    try {
      const res = await getOrderDetail(sn);
      setDetail({
        detail: res.data.detail,
        items: res.data.items ?? [],
        packages: res.data.packages ?? [],
      });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setDetailLoading(null);
    }
  }, [detailLoading]);

  const columns: Column<ShopeeOrderDetailRow>[] = useMemo(() => [
    { label: "Order SN", key: "order_sn" },
    { label: "Store", key: "nama_toko" },
    { label: "Status", key: "order_status" },
    {
      label: "Detail",
      render: (_, row) => (
        <Button 
          size="small" 
          onClick={() => openDetail(row.order_sn)}
          disabled={detailLoading === row.order_sn}
        >
          {detailLoading === row.order_sn ? (
            <CircularProgress size={16} />
          ) : (
            'View'
          )}
        </Button>
      ),
    },
  ], [openDetail, detailLoading]);

  const itemColumns: Column<ShopeeOrderItemRow>[] = useMemo(() => [
    { label: "Item Name", key: "item_name" },
    { label: "Model SKU", key: "model_sku" },
    { label: "Qty", key: "model_quantity_purchased", align: "right" },
    {
      label: "Orig Price",
      key: "model_original_price",
      align: "right",
      render: (v) => formatCurrency(v as number),
    },
    {
      label: "Disc Price",
      key: "model_discounted_price",
      align: "right",
      render: (v) => formatCurrency(v as number),
    },
    {
      label: "Total Orig",
      align: "right",
      render: (_, row) =>
        formatCurrency(
          (row.model_original_price ?? 0) *
            (row.model_quantity_purchased ?? 0),
        ),
    },
    {
      label: "Total Disc",
      align: "right",
      render: (_, row) =>
        formatCurrency(
          (row.model_discounted_price ?? 0) *
            (row.model_quantity_purchased ?? 0),
        ),
    },
  ], []);

  const packageColumns: Column<ShopeeOrderPackageRow>[] = useMemo(() => [
    { label: "Package #", key: "package_number" },
    { label: "Status", key: "logistics_status" },
    { label: "Carrier", key: "shipping_carrier" },
  ], []);

  // Calculations for totals
  const totalOrig = useMemo(() =>
    detail?.items.reduce(
      (sum, it) =>
        sum + (it.model_original_price ?? 0) * (it.model_quantity_purchased ?? 0),
      0,
    ) ?? 0, [detail?.items]);
  
  const totalDisc = useMemo(() =>
    detail?.items.reduce(
      (sum, it) =>
        sum +
        (it.model_discounted_price ?? 0) * (it.model_quantity_purchased ?? 0),
      0,
    ) ?? 0, [detail?.items]);

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
        {storeInput.value !== storeInput.debouncedValue && ' (filtering...)'}
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
  ), [data?.total, isLoading, storeInput.value, storeInput.debouncedValue, pageSize, page, totalPages]);

  return (
    <div>
      <h2>Shopee Order Details</h2>
      
      {/* Filters Section */}
      <Box mb={2} display="flex" gap={1} alignItems="center">
        <select
          aria-label="Store"
          value={storeInput.value}
          onChange={(e) => storeInput.setValue(e.target.value)}
          style={{ marginRight: "0.5rem" }}
          disabled={isLoading}
        >
          <option value="">All Stores</option>
          {stores.map((s) => (
            <option key={s.store_id} value={s.nama_toko}>
              {s.nama_toko}
            </option>
          ))}
        </select>
        <input
          placeholder="Order SN"
          value={orderInput.value}
          onChange={(e) => orderInput.setValue(e.target.value)}
          disabled={isLoading}
        />
        <Button 
          onClick={() => refetch()} 
          disabled={isLoading}
          variant="outlined"
        >
          {isLoading ? <CircularProgress size={20} /> : 'Refresh'}
        </Button>
      </Box>

      {/* Error Messages */}
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error.message || 'An error occurred'}
        </Alert>
      )}
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}

      {/* Data Table - Now virtualized for better performance */}
      <VirtualizedTable 
        columns={columns} 
        data={data?.data || []} 
        loading={isLoading}
        height={500}
        emptyMessage="No order details found"
      />
      
      {/* Pagination Controls */}
      {paginationControls}

      {/* Detail Modal */}
      <Dialog
        open={open}
        onClose={() => {
          setOpen(false);
          setDetail(null);
        }}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Order Detail</DialogTitle>
        <DialogContent>
          {detailLoading ? (
            <Box display="flex" justifyContent="center" alignItems="center" p={4}>
              <CircularProgress />
            </Box>
          ) : detail ? (
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <tbody>
                {Object.entries(detail.detail).map(([k, v]) => (
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
                {detail.items.length > 0 && (
                  <>
                    <tr>
                      <td colSpan={2} style={{ fontWeight: "bold", paddingTop: "1rem" }}>
                        Items
                      </td>
                    </tr>
                    <tr>
                      <td colSpan={2}>
                        <VirtualizedTable 
                          columns={itemColumns} 
                          data={detail.items} 
                          height={300}
                          itemHeight={53}
                        />
                      </td>
                    </tr>
                    <tr>
                      <td colSpan={2} style={{ textAlign: "right", paddingTop: "0.5rem" }}>
                        <div>Total Original Price: {formatCurrency(totalOrig)}</div>
                        <div>Total Discounted Price: {formatCurrency(totalDisc)}</div>
                      </td>
                    </tr>
                  </>
                )}
                {detail.packages.length > 0 && (
                  <>
                    <tr>
                      <td colSpan={2} style={{ fontWeight: "bold", paddingTop: "1rem" }}>
                        Packages
                      </td>
                    </tr>
                    <tr>
                      <td colSpan={2}>
                        <VirtualizedTable 
                          columns={packageColumns} 
                          data={detail.packages} 
                          height={200}
                          itemHeight={53}
                        />
                      </td>
                    </tr>
                  </>
                )}
              </tbody>
            </table>
          ) : (
            <Box display="flex" justifyContent="center" alignItems="center" p={4}>
              Failed to load order details
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              setOpen(false);
              setDetail(null);
            }}
          >
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
