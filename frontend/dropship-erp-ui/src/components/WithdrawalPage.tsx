import { useEffect, useState } from "react";
import {
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
} from "@mui/material";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listAllStores } from "../api";
import {
  createWithdrawal,
  listWithdrawals,
  importWithdrawals,
} from "../api/withdrawals";
import type { Store, Withdrawal } from "../types";
import usePagination from "../usePagination";

export default function WithdrawalPage() {
  const [list, setList] = useState<Withdrawal[]>([]);
  const { paginated, controls } = usePagination(list);
  const [stores, setStores] = useState<Store[]>([]);
  const [open, setOpen] = useState(false);
  const [store, setStore] = useState("");
  const [date, setDate] = useState(new Date().toISOString().split("T")[0]);
  const [amount, setAmount] = useState("0");
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(
    null,
  );

  const fetchData = () => listWithdrawals().then((r) => setList(r.data));

  useEffect(() => {
    fetchData();
    listAllStores().then((r) => setStores(r));
  }, []);

  const columns: Column<Withdrawal>[] = [
    {
      label: "Date",
      key: "date",
      render: (v) => new Date(v).toLocaleDateString(),
    },
    { label: "Store", key: "store" },
    {
      label: "Amount",
      key: "amount",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" }),
    },
  ];

  const handleSave = async () => {
    try {
      await createWithdrawal({
        store,
        date: new Date(date).toISOString().split("T")[0],
        amount: Number(amount),
      });
      setOpen(false);
      setStore("");
      setAmount("0");
      fetchData();
      setMsg({ type: "success", text: "saved" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleImport = async () => {
    if (!file) return;
    try {
      await importWithdrawals(file);
      setFile(null);
      fetchData();
      setMsg({ type: "success", text: "imported" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.message });
    }
  };

  return (
    <div>
      <h2>Withdrawals</h2>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <input
          type="file"
          accept=".xlsx"
          onChange={(e) => setFile(e.target.files ? e.target.files[0] : null)}
        />
        <Button variant="contained" onClick={handleImport} disabled={!file}>
          Import
        </Button>
        <Button
          variant="contained"
          onClick={() => setOpen(true)}
          sx={{ ml: "auto" }}
        >
          Add Withdrawal
        </Button>
      </div>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <SortableTable columns={columns} data={paginated} />
      {controls}
      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Add Withdrawal</DialogTitle>
        <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <TextField
            label="Store"
            value={store}
            onChange={(e) => setStore(e.target.value)}
            select
            SelectProps={{ native: true }}
          >
            <option value=""></option>
            {stores.map((s) => (
              <option key={s.store_id} value={s.nama_toko}>
                {s.nama_toko}
              </option>
            ))}
          </TextField>
          <TextField
            label="Date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            InputLabelProps={{ shrink: true }}
          />
          <TextField
            label="Amount"
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleSave} disabled={!store || !amount}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
