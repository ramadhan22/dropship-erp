import { useEffect, useState } from "react";
import {
  Button,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import JsonTabs from "./JsonTabs";
import {
  listOrderDetails,
  getOrderDetail,
  listAllStores,
} from "../api";
import type {
  ShopeeOrderDetailRow,
  ShopeeOrderItemRow,
  ShopeeOrderPackageRow,
  Store,
} from "../types";
import useServerPagination from "../useServerPagination";

function formatLabel(label: string): string {
  return label.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function formatValue(val: any): string {
  if (typeof val === "string" && /^\d{4}-\d{2}-\d{2}T/.test(val)) {
    return new Date(val).toLocaleString();
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
  const [store, setStore] = useState("");
  const [order, setOrder] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const {
    data,
    controls,
    reload,
  } = useServerPagination((params) =>
    listOrderDetails({
      store,
      order,
      page: params.page,
      page_size: params.pageSize,
    }).then((r) => r.data),
  );
  const [detail, setDetail] = useState<{
    detail: ShopeeOrderDetailRow;
    items: ShopeeOrderItemRow[];
    packages: ShopeeOrderPackageRow[];
  } | null>(null);
  const [open, setOpen] = useState(false);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  useEffect(() => {
    reload();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [store, order]);

  const openDetail = async (sn: string) => {
    try {
      const res = await getOrderDetail(sn);
      setDetail(res.data);
      setOpen(true);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const columns: Column<ShopeeOrderDetailRow>[] = [
    { label: "Order SN", key: "order_sn" },
    { label: "Store", key: "nama_toko" },
    { label: "Status", key: "order_status" },
    {
      label: "Detail",
      render: (_, row) => (
        <Button size="small" onClick={() => openDetail(row.order_sn)}>
          View
        </Button>
      ),
    },
  ];

  const itemColumns: Column<ShopeeOrderItemRow>[] = [
    { label: "Item Name", key: "item_name" },
    { label: "Model SKU", key: "model_sku" },
    { label: "Qty", key: "model_quantity_purchased", align: "right" },
  ];

  const packageColumns: Column<ShopeeOrderPackageRow>[] = [
    { label: "Package #", key: "package_number" },
    { label: "Status", key: "logistics_status" },
    { label: "Carrier", key: "shipping_carrier" },
  ];

  return (
    <div>
      <h2>Shopee Order Details</h2>
      <div style={{ marginBottom: "0.5rem" }}>
        <select
          aria-label="Store"
          value={store}
          onChange={(e) => setStore(e.target.value)}
          style={{ marginRight: "0.5rem" }}
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
          value={order}
          onChange={(e) => setOrder(e.target.value)}
        />
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}
      <SortableTable columns={columns} data={data} />
      {controls}
      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>Order Detail</DialogTitle>
        <DialogContent>
          {detail && (
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
                        <SortableTable columns={itemColumns} data={detail.items} />
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
                        <SortableTable columns={packageColumns} data={detail.packages} />
                      </td>
                    </tr>
                  </>
                )}
              </tbody>
            </table>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
