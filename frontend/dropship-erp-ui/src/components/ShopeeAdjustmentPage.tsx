import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import {
  listShopeeAdjustments,
  updateShopeeAdjustment,
  deleteShopeeAdjustment,
} from "../api/shopeeAdjustments";
import type { ShopeeAdjustment } from "../types";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { getCurrentMonthRange } from "../utils/date";
import usePagination from "../usePagination";

export default function ShopeeAdjustmentPage() {
  const [list, setList] = useState<ShopeeAdjustment[]>([]);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const { paginated, controls, setPage } = usePagination(list);
  const [editOpen, setEditOpen] = useState(false);
  const [selected, setSelected] = useState<ShopeeAdjustment | null>(null);
  const [editReason, setEditReason] = useState("");
  const [editAmount, setEditAmount] = useState("0");
  const [editType, setEditType] = useState("");

  const fetchData = async () => {
    try {
      const res = await listShopeeAdjustments({ from, to });
      setList(res.data);
      setMsg(null);
      setPage(1);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [from, to]);

  const handleSave = async () => {
    if (!selected) return;
    try {
      await updateShopeeAdjustment(selected.id, {
        ...selected,
        alasan_penyesuaian: editReason,
        biaya_penyesuaian: Number(editAmount),
        tipe_penyesuaian: editType,
      });
      setEditOpen(false);
      setSelected(null);
      fetchData();
      setMsg({ type: "success", text: "updated" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const columns: Column<ShopeeAdjustment>[] = [
    { label: "Store", key: "nama_toko" },
    { label: "Date", key: "tanggal_penyesuaian" },
    { label: "Type", key: "tipe_penyesuaian" },
    { label: "Reason", key: "alasan_penyesuaian" },
    {
      label: "Amount",
      key: "biaya_penyesuaian",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    { label: "Order", key: "no_pesanan" },
    {
      label: "",
      render: (_, row) => (
        <>
          <Button
            size="small"
            onClick={() => {
              setSelected(row);
              setEditReason(row.alasan_penyesuaian);
              setEditAmount(String(row.biaya_penyesuaian));
              setEditType(row.tipe_penyesuaian);
              setEditOpen(true);
            }}
          >
            Edit
          </Button>
          <Button
            size="small"
            color="error"
            onClick={async () => {
              if (!confirm("Delete adjustment?")) return;
              try {
                await deleteShopeeAdjustment(row.id);
                fetchData();
              } catch (e: any) {
                setMsg({ type: "error", text: e.response?.data?.error || e.message });
              }
            }}
          >
            Delete
          </Button>
        </>
      ),
    },
  ];

  return (
    <div>
      <h2>Shopee Adjustments</h2>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="From"
            format="yyyy-MM-dd"
            value={new Date(from)}
            onChange={(d) => {
              if (!d) return;
              setFrom(d.toISOString().split("T")[0]);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="To"
            format="yyyy-MM-dd"
            value={new Date(to)}
            onChange={(d) => {
              if (!d) return;
              setTo(d.toISOString().split("T")[0]);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}
      <SortableTable columns={columns} data={paginated} />
      {controls}
      <Dialog open={editOpen} onClose={() => setEditOpen(false)}>
        <DialogTitle>Edit Adjustment</DialogTitle>
        <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <TextField
            label="Type"
            value={editType}
            onChange={(e) => setEditType(e.target.value)}
            size="small"
          />
          <TextField
            label="Reason"
            value={editReason}
            onChange={(e) => setEditReason(e.target.value)}
            size="small"
          />
          <TextField
            label="Amount"
            type="number"
            value={editAmount}
            onChange={(e) => setEditAmount(e.target.value)}
            size="small"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleSave}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
